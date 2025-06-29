/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package service

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"configcenter/src/common"
	cc "configcenter/src/common/backbone/configcenter"
	"configcenter/src/common/blog"
	httpheader "configcenter/src/common/http/header"
	"configcenter/src/common/http/rest"
	"configcenter/src/common/metadata"
	"configcenter/src/web_server/middleware/user"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// LogOutUser log out user
func (s *Service) LogOutUser(c *gin.Context) {
	rid := httpheader.GetRid(c.Request.Header)
	session := sessions.Default(c)

	// 检查是否是OIDC用户 - 通过检查oidc_username来判断
	oidcUsername := session.Get("oidc_username").(string)

	if oidcUsername != "" {
		blog.Infof("OIDC user logout, clearing OIDC session, rid: %s", rid)

		// 构建OIDC退出URL (在清除session之前)
		logoutURL := s.buildOIDCLogoutURL(c)
		blog.Infof("Built OIDC logout URL: %s, rid: %s", logoutURL, rid)

		// 清除所有会话数据
		session.Clear()

		// 清除Cookie
		c.SetCookie(common.BKUser, "", -1, "/", "", false, false)
		c.SetCookie(common.HTTPCookieSupplierAccount, "", -1, "/", "", false, false)
		c.SetCookie(common.HTTPCookieBKToken, "", -1, "/", "", false, false)

		ret := metadata.LogoutResult{}
		ret.BaseResp.Result = true
		ret.Data.LogoutURL = logoutURL
		blog.Infof("Returning logout response with URL: %s, rid: %s", logoutURL, rid)
		c.JSON(200, ret)
		return
	}

	// 非OIDC用户的标准退出流程
	session.Clear()
	c.Request.URL.Path = ""
	userManger := user.NewUser(*s.Config, s.Engine, s.CacheCli, s.ApiCli)
	loginURL := userManger.GetLoginUrl(c)
	ret := metadata.LogoutResult{}
	ret.BaseResp.Result = true
	ret.Data.LogoutURL = loginURL
	c.JSON(200, ret)
	return
}

// IsLogin user is login
func (s *Service) IsLogin(c *gin.Context) {
	user := user.NewUser(*s.Config, s.Engine, s.CacheCli, s.ApiCli)
	isLogin := user.LoginUser(c)
	if isLogin {
		// 获取用户权限
		permissions := s.getUserPermissions(c)
		c.JSON(200, gin.H{
			common.HTTPBKAPIErrorCode:    0,
			common.HTTPBKAPIErrorMessage: nil,
			"permission":                 permissions,
			"result":                     true,
		})
		return
	}
	c.JSON(200, gin.H{
		common.HTTPBKAPIErrorCode:    0,
		common.HTTPBKAPIErrorMessage: "Unauthorized",
		"permission":                 nil,
		"result":                     false,
	})
	return
}

// Login html file
func (s *Service) Login(c *gin.Context) {
	c.HTML(200, "login.html", gin.H{})
}

// LoginUser log in user
func (s *Service) LoginUser(c *gin.Context) {
	rid := httpheader.GetRid(c.Request.Header)
	defErr := s.CCErr.CreateDefaultCCErrorIf(httpheader.GetLanguage(c.Request.Header))
	userName := c.PostForm("username")
	password := c.PostForm("password")
	if userName == "" || password == "" {
		c.HTML(200, "login.html", gin.H{
			"error": defErr.CCError(common.CCErrWebNeedFillinUsernamePasswd).Error(),
		})
	}
	userInfo, err := cc.String("webServer.session.userInfo")
	if err != nil {
		c.HTML(200, "login.html", gin.H{
			"error": defErr.CCError(common.CCErrWebNoUsernamePasswd).Error(),
		})
		return
	}
	userInfos := strings.Split(userInfo, ",")
	for _, userInfo := range userInfos {
		userWithPassword := strings.Split(userInfo, ":")
		if len(userWithPassword) != 2 {
			blog.Errorf("user info config %s invalid, rid: %s", userInfo, rid)
			c.HTML(200, "login.html", gin.H{
				"error": defErr.CCError(common.CCErrWebUserinfoFormatWrong).Error(),
			})
			return
		}
		if userWithPassword[0] == userName && userWithPassword[1] == password {
			c.SetCookie(common.BKUser, userName, 24*60*60, "/", "", false, false)
			session := sessions.Default(c)
			session.Set(userName, time.Now().Unix())
			if err := session.Save(); err != nil {
				blog.Warnf("save session failed, err: %s, rid: %s", err.Error(), rid)
			}
			userManger := user.NewUser(*s.Config, s.Engine, s.CacheCli, s.ApiCli)
			userManger.LoginUser(c)
			var redirectURL string
			if c.Query("c_url") != "" {
				redirectURL = c.Query("c_url")
			} else {
				redirectURL = s.Config.Site.DomainUrl
			}
			c.Redirect(302, redirectURL)
			return
		}
	}
	c.HTML(200, "login.html", gin.H{
		"error": defErr.CCError(common.CCErrWebUsernamePasswdWrong).Error(),
	})
	return
}

// getUserPermissions 获取用户权限
func (s *Service) getUserPermissions(c *gin.Context) []string {
	rid := httpheader.GetRid(c.Request.Header)
	
	// 从cookie中获取用户名
	cookieUser, err := c.Cookie(common.BKUser)
	if err != nil || cookieUser == "" {
		blog.Warnf("Failed to get user from cookie, rid: %s", rid)
		return []string{}
	}
	
	// 尝试从用户管理服务获取用户权限
	userInfo, err := s.getUserFromDatabase(c, cookieUser, rid)
	if err != nil {
		blog.Warnf("Failed to get user from database, user: %s, err: %v, fallback to default permissions, rid: %s", 
			cookieUser, err, rid)
		return s.getDefaultPermissionsByRole(cookieUser)
	}
	
	if userInfo == nil {
		blog.Warnf("User not found in database, user: %s, fallback to default permissions, rid: %s", cookieUser, rid)
		return s.getDefaultPermissionsByRole(cookieUser)
	}
	
	// 如果用户权限为空，使用默认权限
	if len(userInfo.Permissions) == 0 {
		blog.Infof("User permissions is empty, user: %s, using default permissions, rid: %s", cookieUser, rid)
		return s.getDefaultPermissionsByRole(cookieUser)
	}
	
	// 返回用户的权限列表
	blog.V(4).Infof("Get user permissions from database, user: %s, permissions: %v, rid: %s", 
		cookieUser, userInfo.Permissions, rid)
	return userInfo.Permissions
}

// getUserFromDatabase 从cc_user_management表中获取用户信息
func (s *Service) getUserFromDatabase(c *gin.Context, username string, rid string) (*metadata.User, error) {
	// 构建请求头
	requestHeader := make(http.Header)
	for k, v := range c.Request.Header {
		requestHeader[k] = v
	}
	requestHeader.Set("BK_User", username)
	requestHeader.Set("HTTP_BLUEKING_SUPPLIER_ID", common.BKDefaultOwnerID)

	kit := rest.NewKitFromHeader(requestHeader, s.Engine.CCErr)

	// 尝试直接通过用户ID获取用户信息
	userInfo, err := s.Engine.CoreAPI.CoreService().UserManagement().GetUser(kit.Ctx, requestHeader, username)
	if err == nil && userInfo != nil {
		return userInfo, nil
	}

	// 如果直接获取失败，尝试通过列表查询
	userListRequest := &metadata.UserListRequest{
		Search: username, // 按用户名或邮箱搜索
		Limit:  10,       // 获取多个结果以便进行精确匹配
	}

	userListResult, err := s.Engine.CoreAPI.CoreService().UserManagement().ListUsers(kit.Ctx, requestHeader, userListRequest)
	if err != nil {
		blog.Errorf("failed to search user from cc_user_management, user: %s, err: %v, rid: %s", username, err, rid)
		return nil, err
	}

	// 检查是否找到用户，并进行精确匹配
	if userListResult != nil && userListResult.Total > 0 && len(userListResult.Items) > 0 {
		// 在返回的结果中查找精确匹配的用户名或邮箱
		for _, u := range userListResult.Items {
			if u.UserID == username || strings.EqualFold(u.Email, username) {
				return &u, nil
			}
		}
	}

	// 没有找到用户
	blog.V(4).Infof("user not found in cc_user_management, user: %s, rid: %s", username, rid)
	return nil, nil
}

// getDefaultPermissionsByRole 根据用户角色返回默认权限
func (s *Service) getDefaultPermissionsByRole(username string) []string {
	// 根据用户名判断权限
	// 管理员用户返回所有权限
	if username == "admin" || username == "Administrator" || username == "root" {
		return []string{"admin", "home", "business", "model", "resource", "operation"}
	}
	
	// 查看是否是运维用户 (常见的运维账号名模式)
	if strings.Contains(strings.ToLower(username), "ops") || strings.Contains(strings.ToLower(username), "operator") {
		return []string{"home", "business", "resource", "operation"}
	}
	
	// 查看是否是开发用户 (常见的开发账号名模式)
	if strings.Contains(strings.ToLower(username), "dev") || strings.Contains(strings.ToLower(username), "developer") {
		return []string{"home", "business", "model", "resource"}
	}
	
	// 默认用户返回基础权限（首页、业务、资源）
	return []string{"home", "business", "resource"}
}

// buildOIDCLogoutURL 构建OIDC退出登录URL
func (s *Service) buildOIDCLogoutURL(c *gin.Context) string {
	session := sessions.Default(c)

	// 获取OIDC配置中的退出URL
	logoutBaseURL := s.Config.OIDC.LogoutUrl

	// 获取用户的id_token
	idToken, exists := session.Get("oidc_id_token").(string)
	if !exists || idToken == "" {
		blog.Warnf("OIDC id_token not found in session, using logout without id_token_hint")
		// 没有id_token，只能简单退出
		blog.Infof("OIDC logout URL (no id_token): %s", logoutBaseURL)
		return logoutBaseURL
	}

	// 构建OIDC退出URL，使用id_token_hint
	logoutURL := fmt.Sprintf("%s?id_token_hint=%s&post_logout_redirect_uri=%s", logoutBaseURL, url.QueryEscape(idToken), url.QueryEscape(s.Config.OIDC.RedirectUri))

	blog.Infof("OIDC logout URL: %s", logoutURL)
	return logoutURL
}
