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

package user

import (
	"encoding/json"

	"configcenter/src/apimachinery/apiserver"
	"configcenter/src/common"
	"configcenter/src/common/backbone"
	cc "configcenter/src/common/ba
	"configcenter/src/common/blog"
	"configcenter/src/common/errors"
	httpheader "configcenter/src/common/http/header"
	"configcenter/src/common/metadata"
	"configcenter/src/storage/dal/redis"
	"configcenter/src/web_server/app/options"
	"configcenter/src/web_server/middleware/user/plugins"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

type publicUser struct {
	config   options.Config
	engine   *backbone.Engine
	cacheCli redis.Client
	apiCli   apiserver.ApiServerClientInterface
}

// smartSelectPlugin 智能选择登录插件
func (m *publicUser) smartSelectPlugin() metadata.LoginUserPluginInerface {
	// 检查是否配置了OIDC并且用户已通过OIDC登录
	if m.isOIDCEnabled() {
		// 如果配置启用了OIDC，优先使用OIDC插件
		blog.Infof("OIDC is enabled, trying to use OIDC plugin")
		plugin := plugins.CurrentPlugin(common.BKOIDCLoginPluginVersion)
		if plugin != nil {
			blog.Infof("OIDC plugin found and selected")
			return plugin
		}
		blog.Warnf("OIDC plugin not found, falling back to default")
	}

	// 否则使用配置的默认插件
	blog.Infof("Using configured login version: %s", m.config.LoginVersion)
	plugin := plugins.CurrentPlugin(m.config.LoginVersion)
	if plugin != nil {
		blog.Infof("Default plugin found and selected: %s", m.config.LoginVersion)
		return plugin
	}

	// 最后回退到开源插件
	blog.Warnf("Default plugin not found, falling back to opensource plugin")
	plugin = plugins.CurrentPlugin(common.BKOpenSourceLoginPluginVersion)
	if plugin != nil {
		blog.Infof("Opensource plugin found and selected")
		return plugin
	}

	blog.Errorf("No login plugin found at all!")
	return nil
}

// isOIDCEnabled 检查OIDC是否启用
func (m *publicUser) isOIDCEnabled() bool {
	oidcClientId, err := cc.String("webServer.oidc.clientId")
	if err != nil || oidcClientId == "" {
		return false
	}

	oidcAuthUrl, err := cc.String("webServer.oidc.authUrl")
	if err != nil || oidcAuthUrl == "" {
		return false
	}

	return true
}

// LoginUser  user login
func (m *publicUser) LoginUser(c *gin.Context) bool {
	rid := httpheader.GetRid(c.Request.Header)

	isMultiOwner := false
	loginSuccess := false
	var userInfo *metadata.LoginUserInfo
	multipleOwner := m.config.Session.MultipleOwner
	if common.LoginSystemMultiSupplierTrue == multipleOwner {
		isMultiOwner = true
	}

	user := m.smartSelectPlugin()
	if user == nil {
		blog.Errorf("no valid login plugin found, rid: %s", rid)
		return false
	}
	userInfo, loginSuccess = user.LoginUser(c, m.config.ConfigMap, isMultiOwner)

	if !loginSuccess {
		blog.Infof("login user with plugin failed, rid: %s", rid)
		return false
	}
	if true == isMultiOwner || true == userInfo.MultiSupplier {
		ownerM := NewOwnerManager(userInfo.UserName, userInfo.OnwerUin, userInfo.Language)
		ownerM.CacheCli = m.cacheCli
		ownerM.Engine = m.engine
		ownerM.ApiCli = m.apiCli
		// 初始化失败，不影响登录
		_, err := ownerM.InitOwner()
		if nil != err {
			blog.ErrorJSON("init onwer resource pool failed, err:%s, user:%s, rid: %s", err, userInfo, rid)
		}
	}
	strOwnerUinList := []byte("")
	if 0 != len(userInfo.OwnerUinArr) {
		strOwnerUinList, _ = json.Marshal(userInfo.OwnerUinArr)
	}

	session := sessions.Default(c)

	session.Set(common.WEBSessionUinKey, userInfo.UserName)
	session.Set(common.WEBSessionChineseNameKey, userInfo.ChName)
	session.Set(common.WEBSessionPhoneKey, userInfo.Phone)
	session.Set(common.WEBSessionEmailKey, userInfo.Email)
	session.Set(common.WEBSessionRoleKey, userInfo.Role)
	session.Set(common.HTTPCookieBKToken, userInfo.BkToken)
	session.Set(common.HTTPCookieBKTicket, userInfo.BkTicket)
	session.Set(common.WEBSessionOwnerUinKey, userInfo.OnwerUin)
	session.Set(common.WEBSessionAvatarUrlKey, userInfo.AvatarUrl)
	session.Set(common.WEBSessionOwnerUinListeKey, string(strOwnerUinList))
	if userInfo.MultiSupplier {
		session.Set(common.WEBSessionMultiSupplierKey, common.LoginSystemMultiSupplierTrue)
	} else {
		session.Set(common.WEBSessionMultiSupplierKey, common.LoginSystemMultiSupplierFalse)
	}

	if err := session.Save(); err != nil {
		blog.Warnf("save session failed, err: %s, rid: %s", err.Error(), rid)
	}
	return true
}

// GetLoginUrl TODO
func (m *publicUser) GetLoginUrl(c *gin.Context) string {

	params := new(metadata.LogoutRequestParams)
	err := json.NewDecoder(c.Request.Body).Decode(params)
	if nil != err || (common.LogoutHTTPSchemeHTTP != params.HTTPScheme && common.LogoutHTTPSchemeHTTPS != params.HTTPScheme) {
		params.HTTPScheme, err = c.Cookie(common.LogoutHTTPSchemeCookieKey)
		if nil != err || (common.LogoutHTTPSchemeHTTP != params.HTTPScheme && common.LogoutHTTPSchemeHTTPS != params.HTTPScheme) {
			params.HTTPScheme = common.LogoutHTTPSchemeHTTP
		}
	}

	user := m.smartSelectPlugin()
	return user.GetLoginUrl(c, m.config.ConfigMap, params)

}

// GetUserList TODO
func (m *publicUser) GetUserList(c *gin.Context) ([]*metadata.LoginSystemUserInfo, *errors.RawErrorInfo) {
	user := m.smartSelectPlugin()
	return user.GetUserList(c, m.config.ConfigMap)
}
