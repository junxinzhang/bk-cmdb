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
	"net/url"
	"strings"
	"time"

	"configcenter/src/common"
	cc "configcenter/src/common/backbone/configcenter"
	"configcenter/src/common/blog"
	httpheader "configcenter/src/common/http/header"
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
	oidcUsername, isOIDCUser := session.Get("oidc_username").(string)

	if isOIDCUser && oidcUsername != "" {
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
		c.JSON(200, gin.H{
			common.HTTPBKAPIErrorCode:    0,
			common.HTTPBKAPIErrorMessage: nil,
			"permission":                 nil,
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

// buildOIDCLogoutURL 构建OIDC退出登录URL
func (s *Service) buildOIDCLogoutURL(c *gin.Context) string {
	session := sessions.Default(c)

	// 获取OIDC配置中的退出URL
	logoutBaseURL := s.Config.OIDC.LogoutUrl
	if logoutBaseURL == "" {
		blog.Warnf("OIDC logout URL not configured, using default")
		logoutBaseURL = "https://sso.rfc-friso.com/684f89a8a50a4e31e35fc262/oidc/session/end"
	}

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
