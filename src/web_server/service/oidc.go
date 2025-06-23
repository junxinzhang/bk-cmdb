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
	"crypto/md5"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"configcenter/src/common"
	cc "configcenter/src/common/backbone/configcenter"
	"configcenter/src/common/blog"
	httpheader "configcenter/src/common/http/header"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
)

// OIDCUser OIDC用户信息结构
type OIDCUser struct {
	Sub               string `json:"sub"`
	Name              string `json:"name"`
	PreferredUsername string `json:"preferred_username"`
	Email             string `json:"email"`
	EmailVerified     bool   `json:"email_verified"`
}

// TokenResponse OIDC Token响应结构
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int    `json:"expires_in"`
	IDToken      string `json:"id_token"`
}

// generateRandomState 生成随机状态码
func generateRandomState() string {
	b := make([]byte, 32)
	rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// SSOLogin 新的SSO登录入口，用于用户主动选择SSO登录
func (s *Service) SSOLogin(c *gin.Context) {
	rid := httpheader.GetRid(c.Request.Header)

	// 检查OIDC是否启用
	if !s.IsOIDCEnabled() {
		blog.Errorf("SSO login requested but OIDC is not enabled, rid: %s", rid)
		c.HTML(200, "login.html", gin.H{
			"error":        "SSO登录功能未启用，请使用传统登录方式",
			"oidc_enabled": false,
		})
		return
	}

	// 调用OIDC登录流程
	s.OIDCLogin(c)
}

// OIDCLogin 处理OIDC登录请求
func (s *Service) OIDCLogin(c *gin.Context) {
	rid := httpheader.GetRid(c.Request.Header)

	// 检查OIDC配置是否完整
	if s.Config.OIDC.ClientId == "" || s.Config.OIDC.AuthUrl == "" {
		blog.Errorf("OIDC configuration is incomplete, rid: %s", rid)
		c.HTML(200, "login.html", gin.H{
			"error": "OIDC configuration is incomplete",
		})
		return
	}

	// 生成状态码并保存到会话
	state := generateRandomState()
	session := sessions.Default(c)
	session.Set("oidc_state", state)
	if err := session.Save(); err != nil {
		blog.Warnf("save oidc state to session failed, err: %s, rid: %s", err.Error(), rid)
	}

	// 构建授权URL
	authURL := fmt.Sprintf("%s?response_type=code&client_id=%s&redirect_uri=%s&scope=%s&state=%s",
		s.Config.OIDC.AuthUrl,
		url.QueryEscape(s.Config.OIDC.ClientId),
		url.QueryEscape(s.Config.OIDC.RedirectUri),
		url.QueryEscape(s.Config.OIDC.Scopes),
		url.QueryEscape(state),
	)

	blog.Infof("redirecting to OIDC provider: %s, rid: %s", authURL, rid)
	c.Redirect(302, authURL)
}

// OIDCCallback 处理OIDC回调
func (s *Service) OIDCCallback(c *gin.Context) {
	rid := httpheader.GetRid(c.Request.Header)

	// 获取授权码和状态
	code := c.Query("code")
	state := c.Query("state")
	errorParam := c.Query("error")

	if errorParam != "" {
		blog.Errorf("OIDC callback error: %s, rid: %s", errorParam, rid)
		c.HTML(200, "login.html", gin.H{
			"error": fmt.Sprintf("Authentication failed: %s", errorParam),
		})
		return
	}

	if code == "" {
		blog.Errorf("OIDC authorization code is empty, rid: %s", rid)
		c.HTML(200, "login.html", gin.H{
			"error": "Authorization code is missing",
		})
		return
	}

	// 验证状态码
	session := sessions.Default(c)
	savedState, exists := session.Get("oidc_state").(string)
	if !exists || savedState != state {
		blog.Errorf("OIDC state mismatch, saved: %s, received: %s, rid: %s", savedState, state, rid)
		c.HTML(200, "login.html", gin.H{
			"error": "Invalid state parameter",
		})
		return
	}

	// 交换授权码获取令牌
	token, err := s.exchangeCodeForToken(code, rid)
	if err != nil {
		blog.Errorf("exchange code for token failed, code: %s, err: %s, rid: %s", code[:10]+"...", err.Error(), rid)
		c.HTML(200, "login.html", gin.H{
			"error": "Failed to exchange authorization code for token. Please check SSO configuration.",
		})
		return
	}

	blog.Infof("successfully exchanged code for token, token_type: %s, expires_in: %d, rid: %s",
		token.TokenType, token.ExpiresIn, rid)

	// 获取用户信息
	userInfo, err := s.fetchUserInfo(token.AccessToken, rid)
	if err != nil {
		blog.Errorf("fetch user info failed, err: %s, access_token: %s, rid: %s", err.Error(), token.AccessToken[:10]+"...", rid)
		c.HTML(200, "login.html", gin.H{
			"error": "Failed to fetch user information from SSO provider",
		})
		return
	}

	blog.Infof("fetched user info: email=%s, preferred_username=%s, name=%s, rid: %s",
		userInfo.Email, userInfo.PreferredUsername, userInfo.Name, rid)

	// 确定用户名 - 优先使用邮箱，转换为小写
	userName := ""
	if userInfo.Email != "" {
		userName = strings.ToLower(userInfo.Email)
	} else if userInfo.PreferredUsername != "" {
		userName = strings.ToLower(userInfo.PreferredUsername)
	} else if userInfo.Name != "" {
		userName = strings.ToLower(userInfo.Name)
	}

	if userName == "" {
		blog.Errorf("cannot determine username from OIDC user info, rid: %s", rid)
		c.HTML(200, "login.html", gin.H{
			"error": "Cannot determine username from SSO response",
		})
		return
	}

	blog.Infof("OIDC user identified: %s, rid: %s", userName, rid)

	// 验证用户是否在系统用户列表中
	if !s.validateUserExists(userName, rid) {
		blog.Warnf("OIDC user %s does not exist in system user list, rid: %s", userName, rid)
		c.HTML(200, "login.html", gin.H{
			"error": fmt.Sprintf("用户 %s 不存在于系统中，请联系管理员添加用户权限", userName),
		})
		return
	}

	blog.Infof("OIDC user %s validated successfully, rid: %s", userName, rid)

	// 生成一致的BkToken
	expireTime := time.Now().Unix() + 24*60*60 // 24小时后过期
	bkToken := generateBkToken(userName, expireTime)

	// 设置Cookie
	c.SetCookie(common.BKUser, userName, 24*60*60, "/", "", false, false)
	c.SetCookie(common.HTTPCookieSupplierAccount, common.BKDefaultOwnerID, 24*60*60, "/", "", false, false)
	c.SetCookie(common.HTTPCookieBKToken, bkToken, 24*60*60, "/", "", false, false)

	// 设置完整的OIDC会话信息
	session = sessions.Default(c)

	// 设置OIDC特定的会话数据 - 使用字符串类型避免gob序列化问题
	session.Set("oidc_username", userName)
	session.Set("oidc_chname", userInfo.Name)
	session.Set("oidc_email", userInfo.Email)
	session.Set("oidc_phone", "")
	session.Set("oidc_role", "user")
	session.Set("oidc_avatar_url", "")
	session.Set("oidc_token", token.AccessToken)
	session.Set("oidc_id_token", token.IDToken)
	session.Set("oidc_expire", expireTime)

	// 设置标准会话数据
	session.Set(common.WEBSessionUinKey, userName)
	session.Set(common.WEBSessionChineseNameKey, userInfo.Name)
	session.Set(common.WEBSessionEmailKey, userInfo.Email)
	session.Set(common.WEBSessionPhoneKey, "")
	session.Set(common.WEBSessionRoleKey, "user")
	session.Set(common.HTTPCookieBKToken, bkToken)
	session.Set(common.WEBSessionOwnerUinKey, common.BKDefaultOwnerID)
	session.Set(common.WEBSessionAvatarUrlKey, "")
	session.Set(common.WEBSessionMultiSupplierKey, common.LoginSystemMultiSupplierFalse)

	// 设置登录时间戳
	session.Set(userName, time.Now().Unix())

	if err := session.Save(); err != nil {
		blog.Errorf("save OIDC session failed, err: %s, rid: %s", err.Error(), rid)
		c.HTML(200, "login.html", gin.H{
			"error": "会话保存失败，请重试",
		})
		return
	}

	blog.Infof("OIDC user session established successfully: %s, rid: %s", userName, rid)

	// 重定向到目标URL
	redirectURL := s.Config.Site.DomainUrl
	if c.Query("c_url") != "" {
		redirectURL = c.Query("c_url")
	}

	blog.Infof("redirecting OIDC user to: %s, rid: %s", redirectURL, rid)
	c.Redirect(302, redirectURL)
}

// exchangeCodeForToken 交换授权码获取访问令牌
func (s *Service) exchangeCodeForToken(code, rid string) (*TokenResponse, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", s.Config.OIDC.RedirectUri)
	data.Set("client_id", s.Config.OIDC.ClientId)
	data.Set("client_secret", s.Config.OIDC.ClientSecret)

	req, err := http.NewRequest("POST", s.Config.OIDC.TokenUrl, strings.NewReader(data.Encode()))
	if err != nil {
		return nil, fmt.Errorf("create token request failed: %v", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("token request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read token response failed: %v", err)
	}

	if resp.StatusCode != http.StatusOK {
		blog.Errorf("token endpoint returned error: %d %s, body: %s, rid: %s",
			resp.StatusCode, resp.Status, string(body), rid)
		return nil, fmt.Errorf("token endpoint error: %d %s", resp.StatusCode, resp.Status)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("parse token response failed: %v", err)
	}

	return &tokenResp, nil
}

// fetchUserInfo 获取用户信息
func (s *Service) fetchUserInfo(accessToken, rid string) (*OIDCUser, error) {
	blog.Infof("fetching user info from: %s, rid: %s", s.Config.OIDC.UserInfoUrl, rid)

	req, err := http.NewRequest("GET", s.Config.OIDC.UserInfoUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("create userinfo request failed: %v", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("userinfo request failed: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read userinfo response failed: %v", err)
	}

	blog.Infof("userinfo response status: %d, body length: %d, rid: %s", resp.StatusCode, len(body), rid)

	if resp.StatusCode != http.StatusOK {
		blog.Errorf("userinfo endpoint returned error: %d %s, body: %s, rid: %s",
			resp.StatusCode, resp.Status, string(body), rid)
		return nil, fmt.Errorf("userinfo endpoint error: %d %s, response: %s", resp.StatusCode, resp.Status, string(body))
	}

	blog.Infof("received userinfo response: %s, rid: %s", string(body), rid)

	var userInfo OIDCUser
	if err := json.Unmarshal(body, &userInfo); err != nil {
		blog.Errorf("parse userinfo response failed, body: %s, err: %v, rid: %s", string(body), err, rid)
		return nil, fmt.Errorf("parse userinfo response failed: %v", err)
	}

	return &userInfo, nil
}

// IsOIDCEnabled 检查是否启用了OIDC
func (s *Service) IsOIDCEnabled() bool {
	return s.Config.OIDC.ClientId != "" &&
		s.Config.OIDC.AuthUrl != "" &&
		s.Config.OIDC.TokenUrl != "" &&
		s.Config.OIDC.UserInfoUrl != ""
}

// validateUserExists 验证用户是否在系统用户列表中
func (s *Service) validateUserExists(userName, rid string) bool {
	// 首先尝试从配置中获取用户列表
	userInfo, err := cc.String("webServer.session.userInfo")
	if err != nil {
		blog.Warnf("failed to get userInfo config, will allow OIDC user by default, err: %s, rid: %s", err.Error(), rid)
		// 如果配置不存在，允许OIDC用户登录（适用于企业版部署）
		return true
	}

	if userInfo == "" {
		blog.Warnf("userInfo config is empty, will allow OIDC user by default, rid: %s", rid)
		return true
	}

	userInfos := strings.Split(userInfo, ",")
	for _, userInfoItem := range userInfos {
		userInfoItem = strings.TrimSpace(userInfoItem)
		if userInfoItem == "" {
			continue
		}

		// 支持两种格式：
		// 1. username:password (开源版格式)
		// 2. username (仅用户名格式，适用于OIDC)
		userParts := strings.Split(userInfoItem, ":")
		configUserName := strings.TrimSpace(userParts[0])

		if configUserName == "" {
			continue
		}

		// 支持大小写不敏感的用户名/邮箱匹配
		if strings.ToLower(configUserName) == strings.ToLower(userName) {
			blog.Infof("user %s found in system user list, rid: %s", userName, rid)
			return true
		}
	}

	// 检查是否配置了OIDC允许的用户列表
	oidcUsers, err := cc.String("webServer.oidc.allowedUsers")
	if err == nil && oidcUsers != "" {
		blog.Infof("checking OIDC allowed users list, rid: %s", rid)
		allowedUsers := strings.Split(oidcUsers, ",")
		for _, allowedUser := range allowedUsers {
			allowedUser = strings.TrimSpace(allowedUser)
			if strings.ToLower(allowedUser) == strings.ToLower(userName) {
				blog.Infof("user %s found in OIDC allowed users list, rid: %s", userName, rid)
				return true
			}
		}
	}

	blog.Warnf("user %s not found in system user list, rid: %s", userName, rid)
	return false
}

// generateBkToken 生成一致的BkToken
func generateBkToken(username string, expireTime int64) string {
	// 使用用户名和过期时间生成一致的token
	data := fmt.Sprintf("oidc:%s:%d", username, expireTime)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)[:32]
}
