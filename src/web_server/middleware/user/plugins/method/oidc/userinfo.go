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

// Package oidc OIDC login method for SSO authentication
package oidc

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"configcenter/src/common"
	cc "configcenter/src/common/backbone/configcenter"
	"configcenter/src/common/blog"
	"configcenter/src/common/errors"
	httpheader "configcenter/src/common/http/header"
	"configcenter/src/common/metadata"
	webCommon "configcenter/src/web_server/common"
	"configcenter/src/web_server/middleware/user/plugins/manager"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

const (
	// OIDC会话key常量
	OIDCSessionUsernameKey = "oidc_username"
	OIDCSessionChnameKey   = "oidc_chname"
	OIDCSessionEmailKey    = "oidc_email"
	OIDCSessionPhoneKey    = "oidc_phone"
	OIDCSessionRoleKey     = "oidc_role"
	OIDCSessionAvatarKey   = "oidc_avatar_url"
	OIDCSessionTokenKey    = "oidc_token"
	OIDCSessionExpireKey   = "oidc_expire"
)

func init() {
	plugin := &metadata.LoginPluginInfo{
		Name:       "OIDC SSO authentication",
		Version:    common.BKOIDCLoginPluginVersion,
		HandleFunc: &user{},
	}
	manager.RegisterPlugin(plugin)
}

type user struct{}

// LoginUser OIDC用户登录验证
func (m *user) LoginUser(c *gin.Context, config map[string]string, isMultiOwner bool) (*metadata.LoginUserInfo, bool) {
	rid := httpheader.GetRid(c.Request.Header)
	session := sessions.Default(c)

	// 处理OwnerID设置
	cookieOwnerID, err := c.Cookie(common.HTTPCookieSupplierAccount)
	if "" == cookieOwnerID || err != nil {
		c.SetCookie(common.HTTPCookieSupplierAccount, common.BKDefaultOwnerID, 0, "/", "", false, false)
		session.Set(common.WEBSessionOwnerUinKey, common.BKDefaultOwnerID)
		cookieOwnerID = common.BKDefaultOwnerID
	} else if cookieOwnerID != session.Get(common.WEBSessionOwnerUinKey) {
		session.Set(common.WEBSessionOwnerUinKey, cookieOwnerID)
	}

	// 获取用户名cookie
	cookieUser, err := c.Cookie(common.BKUser)
	if "" == cookieUser || nil != err {
		blog.Errorf("OIDC login user not found in cookie, rid: %s", rid)
		return nil, false
	}

	// 检查OIDC会话信息
	oidcUsername, usernameExists := session.Get(OIDCSessionUsernameKey).(string)
	if !usernameExists || oidcUsername == "" {
		blog.Errorf("OIDC username not found in session, rid: %s", rid)
		return nil, false
	}

	// 检查token有效性
	oidcToken, tokenExists := session.Get(OIDCSessionTokenKey).(string)
	if !tokenExists || oidcToken == "" {
		blog.Errorf("OIDC token not found in session, rid: %s", rid)
		return nil, false
	}

	// 检查过期时间
	expireTime, expireExists := session.Get(OIDCSessionExpireKey).(int64)
	if !expireExists || time.Now().Unix() > expireTime {
		blog.Errorf("OIDC session expired, rid: %s", rid)
		return nil, false
	}

	// 验证用户名一致性
	if oidcUsername != cookieUser {
		blog.Errorf("OIDC user mismatch: cookie=%s, session=%s, rid: %s", cookieUser, oidcUsername, rid)
		return nil, false
	}

	// 生成BkToken
	bkToken := generateBkToken(cookieUser, expireTime)

	// 验证BkToken一致性
	cookieBkToken, _ := c.Cookie(common.HTTPCookieBKToken)
	if cookieBkToken != bkToken {
		blog.Warnf("OIDC BkToken mismatch, regenerating, rid: %s", rid)
		c.SetCookie(common.HTTPCookieBKToken, bkToken, 24*60*60, "/", "", false, false)
	}

	// 获取OIDC会话数据
	oidcChname, _ := session.Get(OIDCSessionChnameKey).(string)
	oidcEmail, _ := session.Get(OIDCSessionEmailKey).(string)
	oidcPhone, _ := session.Get(OIDCSessionPhoneKey).(string)
	oidcRole, _ := session.Get(OIDCSessionRoleKey).(string)
	oidcAvatarUrl, _ := session.Get(OIDCSessionAvatarKey).(string)

	// 设置默认值
	if oidcChname == "" {
		oidcChname = cookieUser
	}
	if oidcRole == "" {
		oidcRole = "user"
	}

	// 构建用户信息
	userInfo := &metadata.LoginUserInfo{
		UserName:      cookieUser,
		ChName:        oidcChname,
		Phone:         oidcPhone,
		Email:         oidcEmail,
		Role:          oidcRole,
		BkToken:       bkToken,
		BkTicket:      "",
		OnwerUin:      cookieOwnerID,
		IsOwner:       cookieOwnerID == common.BKDefaultOwnerID,
		Language:      webCommon.GetLanguageByHTTPRequest(c),
		AvatarUrl:     oidcAvatarUrl,
		MultiSupplier: isMultiOwner,
	}

	// 保存完整的会话信息
	if err := session.Save(); err != nil {
		blog.Warnf("save OIDC session failed, err: %s, rid: %s", err.Error(), rid)
	}

	blog.Infof("OIDC user login successful: %s, rid: %s", cookieUser, rid)
	return userInfo, true
}

// GetLoginUrl 获取OIDC登录URL
func (m *user) GetLoginUrl(c *gin.Context, config map[string]string, input *metadata.LogoutRequestParams) string {
	var siteURL string
	var err error

	if common.LogoutHTTPSchemeHTTPS == input.HTTPScheme {
		siteURL, err = cc.String("webServer.site.httpsDomainUrl")
	} else {
		siteURL, err = cc.String("webServer.site.domainUrl")
	}
	if err != nil {
		siteURL = ""
	}
	siteURL = strings.TrimRight(siteURL, "/")

	// 返回SSO登录URL而不是普通登录URL
	return fmt.Sprintf("%s/sso/login?c_url=%s%s", siteURL, siteURL, c.Request.URL.String())
}

// GetUserList 获取OIDC用户列表 - 从配置中读取允许的用户列表
func (m *user) GetUserList(c *gin.Context, config map[string]string) ([]*metadata.LoginSystemUserInfo, *errors.RawErrorInfo) {
	rid := httpheader.GetRid(c.Request.Header)
	users := make([]*metadata.LoginSystemUserInfo, 0)

	// 从配置中读取用户信息
	userInfo, err := cc.String("webServer.session.userInfo")
	if err != nil {
		blog.Errorf("OIDC user list not found in config webServer.session.userInfo, rid: %s", rid)
		return nil, &errors.RawErrorInfo{
			ErrCode: common.CCErrWebNoUsernamePasswd,
		}
	}

	userInfos := strings.Split(userInfo, ",")
	for _, userInfo := range userInfos {
		userParts := strings.Split(userInfo, ":")
		if len(userParts) < 1 {
			continue
		}

		userName := strings.TrimSpace(userParts[0])
		if userName == "" {
			continue
		}

		user := &metadata.LoginSystemUserInfo{
			CnName: userName,
			EnName: userName,
		}
		users = append(users, user)
	}

	blog.Infof("OIDC user list retrieved: %d users, rid: %s", len(users), rid)
	return users, nil
}

// generateBkToken 生成一致的BkToken
func generateBkToken(username string, expireTime int64) string {
	// 使用用户名和过期时间生成一致的token
	data := fmt.Sprintf("oidc:%s:%d", username, expireTime)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)[:32]
}
