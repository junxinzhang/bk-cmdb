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

package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"configcenter/src/apimachinery/apiserver"
	"configcenter/src/apimachinery/discovery"
	"configcenter/src/common"
	"configcenter/src/common/backbone"
	"configcenter/src/common/blog"
	httpheader "configcenter/src/common/http/header"
	"configcenter/src/common/http/httpclient"
	"configcenter/src/common/http/rest"
	"configcenter/src/common/metadata"
	"configcenter/src/common/resource/apigw"
	"configcenter/src/common/resource/esb"
	"configcenter/src/common/resource/jwt"
	"configcenter/src/storage/dal/redis"
	"configcenter/src/web_server/app/options"
	webCommon "configcenter/src/web_server/common"
	"configcenter/src/web_server/middleware/user"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// Engine TODO
var Engine *backbone.Engine

// CacheCli TODO
var CacheCli redis.Client

const (
	message          string = "message"
	inaccessibleCode int    = 1302403
)

// ValidLogin valid the user login status
func ValidLogin(config options.Config, disc discovery.DiscoveryInterface,
	apiCli apiserver.ApiServerClientInterface) gin.HandlerFunc {

	return func(c *gin.Context) {
		rid := httpheader.GetRid(c.Request.Header)
		pathArr := strings.Split(c.Request.URL.Path, "/")
		path1 := pathArr[1]

		// 删除 Accept-Encoding 避免返回值被压缩
		c.Request.Header.Del("Accept-Encoding")

		switch path1 {
		case "healthz", "metrics", "login", "static", "is_login", "oidc", "sso":
			c.Next()
			return
		case "api":
			// Check if this is user management API
			if strings.HasPrefix(c.Request.URL.Path, "/api/v3/usermgmt/") || strings.HasPrefix(c.Request.URL.Path, "/api/v3/role/") {
				// Allow user management API without authentication for now
				// Add necessary headers for core service communication
				httpheader.AddUser(c.Request.Header, "admin")
				c.Request.Header.Set("HTTP_BLUEKING_SUPPLIER_ID", "0")
				c.Request.Header.Set("BK_User", "admin")
				c.Next()
				return
			}
		}

		if isAuthed(c, config, apiCli) {
			handleAuthedReq(c, config, path1, disc, rid)
			return
		}

		if path1 == "api" {
			c.JSON(401, gin.H{
				"status": "log out",
			})
			c.Abort()
			return
		}

		user := user.NewUser(config, Engine, CacheCli, apiCli)
		url := user.GetLoginUrl(c)
		c.Redirect(302, url)
		c.Abort()
	}

}

func handleAuthedReq(c *gin.Context, config options.Config, path1 string, disc discovery.DiscoveryInterface,
	rid string) {

	// http request header add user
	session := sessions.Default(c)
	userName, _ := session.Get(common.WEBSessionUinKey).(string)
	ownerID, _ := session.Get(common.WEBSessionOwnerUinKey).(string)
	bkToken, _ := session.Get(common.HTTPCookieBKToken).(string)
	bkTicket, _ := session.Get(common.HTTPCookieBKTicket).(string)
	language := webCommon.GetLanguageByHTTPRequest(c)
	httpheader.AddUser(c.Request.Header, userName)
	httpheader.AddLanguage(c.Request.Header, language)
	httpheader.AddSupplierAccount(c.Request.Header, ownerID)
	httpheader.SetUserToken(c.Request.Header, bkToken)
	httpheader.SetUserTicket(c.Request.Header, bkTicket)

	if config.LoginVersion == common.BKBluekingLoginPluginVersion {
		resp, err := esb.EsbClient().LoginSrv().GetUser(c.Request.Context(), c.Request.Header)
		if err != nil {
			c.JSON(http.StatusInternalServerError,
				gin.H{"status": fmt.Sprintf("get user from bk-login failed, err: %v", err)})
			c.Abort()
			return
		}

		if resp.Code == inaccessibleCode {
			data := gin.H{
				message: resp.Message,
			}
			c.HTML(http.StatusOK, webCommon.InaccessibleHtml, data)
			c.Abort()
			return
		}
	}

	if path1 == "api" {
		// proxy request to api gateway for blueking deployment method
		if config.DeploymentMethod == common.BluekingDeployment {
			apigw.Client().Cmdb().Proxy(c.Request, c.Writer)
			return
		}

		// proxy request to api-server for independent deployment method
		header, err := jwt.GetHandler().Sign(c.Request.Header)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"status": fmt.Sprintf("sign jwt info failed, err: %v", err)})
			c.Abort()
			return
		}
		c.Request.Header = header

		servers, err := disc.ApiServer().GetServers()
		if err != nil || len(servers) == 0 {
			blog.Errorf("no api server can be used. err: %v, rid: %s", err, rid)
			c.JSON(503, gin.H{
				"status": "no api server can be used.",
			})
			c.Abort()
			return
		}
		url := servers[0]
		httpclient.ProxyHttp(c, url)
		return
	}
	c.Next()
}

// isAuthed check user is authed
func isAuthed(c *gin.Context, config options.Config, apiCli apiserver.ApiServerClientInterface) bool {
	rid := httpheader.GetRid(c.Request.Header)
	user := user.NewUser(config, Engine, CacheCli, apiCli)
	session := sessions.Default(c)

	// check bk_token
	ccToken := session.Get(common.HTTPCookieBKToken)
	if ccToken == nil {
		blog.Errorf("session key %s not found, rid: %s", common.HTTPCookieBKToken, rid)
		return user.LoginUser(c)
	}

	// check username
	userName, ok := session.Get(common.WEBSessionUinKey).(string)
	if !ok || "" == userName {
		return user.LoginUser(c)
	}

	// check owner_uin
	ownerID, ok := session.Get(common.WEBSessionOwnerUinKey).(string)
	if !ok || "" == ownerID {
		return user.LoginUser(c)
	}

	bkTokenName := common.HTTPCookieBKToken
	bkToken, err := c.Cookie(bkTokenName)
	blog.V(5).Infof("valid user login session token %s, cookie token %s, rid: %s", ccToken, bkToken, rid)
	if nil != err || bkToken != ccToken {
		return user.LoginUser(c)
	}

	// 检查用户状态 - 实时验证用户是否被禁用
	if !checkUserStatus(c, userName, apiCli, rid) {
		blog.Warnf("user %s status check failed, forcing logout, rid: %s", userName, rid)
		// 清除会话
		session.Clear()
		if err := session.Save(); err != nil {
			blog.Errorf("failed to clear session for disabled user %s, err: %s, rid: %s", userName, err.Error(), rid)
		}
		// 清除Cookie
		c.SetCookie(common.BKUser, "", -1, "/", "", false, false)
		c.SetCookie(common.HTTPCookieSupplierAccount, "", -1, "/", "", false, false)
		c.SetCookie(common.HTTPCookieBKToken, "", -1, "/", "", false, false)
		return user.LoginUser(c)
	}

	return true
}

// checkUserStatus 检查用户状态是否为active
func checkUserStatus(c *gin.Context, userName string, apiCli apiserver.ApiServerClientInterface, rid string) bool {
	// 构造请求头
	requestHeader := make(http.Header)
	for k, v := range c.Request.Header {
		requestHeader[k] = v
	}
	requestHeader.Set("BK_User", userName)
	requestHeader.Set("HTTP_BLUEKING_SUPPLIER_ID", common.BKDefaultOwnerID)

	kit := rest.NewKitFromHeader(requestHeader, Engine.CCErr)

	// 查询用户信息
	userListRequest := &metadata.UserListRequest{
		Search: userName, // 使用用户名搜索
		Limit:  10,
	}

	userListResult, err := Engine.CoreAPI.CoreService().UserManagement().ListUsers(kit.Ctx, requestHeader, userListRequest)
	if err != nil {
		blog.Errorf("failed to query user status for %s, err: %v, rid: %s", userName, err, rid)
		// 查询失败时，为了安全起见，认为用户无效
		return false
	}

	// 查找匹配的用户
	var userFound *metadata.User
	if userListResult.Total > 0 && len(userListResult.Items) > 0 {
		for _, u := range userListResult.Items {
			// 支持邮箱和用户名匹配
			if strings.EqualFold(u.Email, userName) || strings.EqualFold(u.UserID, userName) {
				userFound = &u
				break
			}
		}
	}

	if userFound == nil {
		blog.Warnf("user %s not found in user management system, rid: %s", userName, rid)
		return false
	}

	// 检查用户状态
	if userFound.Status != metadata.UserStatusActive {
		blog.Warnf("user %s status is %s, not active, rid: %s", userName, userFound.Status, rid)
		return false
	}

	blog.V(5).Infof("user %s status check passed, status: %s, rid: %s", userName, userFound.Status, rid)
	return true
}
