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
	"configcenter/src/common/http/rest"
	"configcenter/src/common/metadata"

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
		msg := fmt.Sprintf("OIDC callback error: %s, rid: %s", errorParam, rid)
		blog.Errorf(msg)
		s.renderOIDCErrorPage(c, msg)
		return
	}

	if code == "" {
		msg := fmt.Sprintf("OIDC authorization code is empty, rid: %s", rid)
		blog.Errorf(msg)
		s.renderOIDCErrorPage(c, msg)
		return
	}

	// 验证状态码
	session := sessions.Default(c)
	savedState, exists := session.Get("oidc_state").(string)
	if !exists || savedState != state {
		msg := fmt.Sprintf("OIDC state mismatch, saved: %s, received: %s, rid: %s", savedState, state, rid)
		blog.Errorf(msg)
		s.renderOIDCErrorPage(c, msg)
		return
	}

	// 交换授权码获取令牌
	token, err := s.exchangeCodeForToken(code, rid)
	if err != nil {
		msg := fmt.Sprintf("exchange code for token failed, code: %s, err: %s, rid: %s", code[:10]+"...", err.Error(), rid)
		blog.Errorf(msg)
		s.renderOIDCErrorPage(c, msg)
		return
	}

	blog.Infof("successfully exchanged code for token, token_type: %s, expires_in: %d, rid: %s",
		token.TokenType, token.ExpiresIn, rid)

	// 获取用户信息
	userInfo, err := s.fetchUserInfo(token.AccessToken, rid)
	if err != nil {
		blog.Errorf("fetch user info failed, err: %s, access_token: %s, rid: %s", err.Error(), token.AccessToken[:10]+"...", rid)
		s.renderOIDCErrorPage(c, "Failed to fetch user information from SSO provider")
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
		s.renderOIDCErrorPage(c, "Cannot determine username from SSO response")
		return
	}

	blog.Infof("OIDC user identified: %s, rid: %s", userName, rid)

	// 验证用户是否在系统用户列表中
	//if !s.validateUserExists(userName, rid) {
	//	blog.Warnf("OIDC user %s does not exist in system user list, rid: %s", userName, rid)
	//	c.HTML(200, "login.html", gin.H{
	//		"error": fmt.Sprintf("用户 %s 不存在于系统中，请联系管理员添加用户权限", userName),
	//	})
	//	return
	//}

	// 验证用户是否在 cc_user_management 表中存在且状态为活跃
	//  // 用户认证头 - 解决API认证问题
	//  'BK_User': getCurrentUser(),
	//  // 供应商ID头 - 蓝鲸平台必需
	//  'HTTP_BLUEKING_SUPPLIER_ID': '0'
	// c.Request.Header 中手动封装 BK_User 和 HTTP_BLUEKING_SUPPLIER_ID
	requestHeader := make(http.Header)
	for k, v := range c.Request.Header {
		requestHeader[k] = v
	}
	requestHeader.Set("BK_User", userName)
	requestHeader.Set("HTTP_BLUEKING_SUPPLIER_ID", common.BKDefaultOwnerID)

	kit := rest.NewKitFromHeader(requestHeader, s.Engine.CCErr)

	// 使用 ListUsers 方法进行邮箱模糊查询
	userListRequest := &metadata.UserListRequest{
		Search: userName, // 邮箱搜索，后端会进行大小写不敏感的匹配
		Limit:  10,       // 获取多个结果以便进行精确匹配
	}

	userListResult, err := s.Engine.CoreAPI.CoreService().UserManagement().ListUsers(kit.Ctx, requestHeader, userListRequest)
	if err != nil {
		blog.Errorf("failed to search user from cc_user_management, user: %s, err: %v, rid: %s", userName, err, rid)
		s.renderOIDCErrorPage(c, "该用户不存在，请联系管理员")
		return
	}

	// 检查是否找到用户，并进行精确匹配
	var user *metadata.User
	if userListResult.Total > 0 && len(userListResult.Items) > 0 {
		// 在返回的结果中查找精确匹配的邮箱
		for _, u := range userListResult.Items {
			if strings.EqualFold(u.Email, userName) { // 大小写不敏感比较
				user = &u
				break
			}
		}
	}

	if user == nil {
		blog.Warnf("OIDC user %s not found in cc_user_management, rid: %s", userName, rid)
		s.renderOIDCErrorPage(c, "该用户不存在，请联系管理员")
		return
	}

	// 检查用户状态
	if user.Status != metadata.UserStatusActive {
		blog.Warnf("OIDC user %s exists but status is not active: %s, rid: %s", userName, user.Status, rid)
		s.renderOIDCErrorPage(c, "该用户已经被禁用，请联系管理员")
		return
	}

	blog.Infof("OIDC user %s validated successfully in cc_user_management, status: %s, user_id: %s, current login_count: %d, rid: %s", 
		userName, user.Status, user.UserID, user.LoginCount, rid)

	// 更新用户登录记录
	now := time.Now()
	newLoginCount := user.LoginCount + 1

	blog.Infof("attempting to update login record for OIDC user %s, current login_count: %d, new login_count: %d, rid: %s", 
		userName, user.LoginCount, newLoginCount, rid)

	// 构建更新请求，只更新登录相关字段
	updateUserRequest := &metadata.UpdateUserRequest{
		LastLogin:  &now,
		LoginCount: &newLoginCount,
	}

	updateUserKit := rest.NewKitFromHeader(requestHeader, s.Engine.CCErr)
	blog.V(3).Infof("calling UpdateUser API with user_id: %s, rid: %s", user.UserID, rid)
	
	updatedUser, err := s.Engine.CoreAPI.CoreService().UserManagement().UpdateUser(updateUserKit.Ctx, requestHeader, user.UserID, updateUserRequest)
	if err != nil {
		// 如果更新用户记录失败，记录日志但不阻止登录流程
		blog.Errorf("failed to update user login record for OIDC user %s, user_id: %s, err: %v, rid: %s", userName, user.UserID, err, rid)
	} else {
		blog.Infof("successfully updated login record for OIDC user %s (user_id: %s), login_count: %d, last_login: %v, rid: %s", 
			userName, user.UserID, newLoginCount, now, rid)
		if updatedUser != nil {
			blog.V(3).Infof("updated user object: login_count=%d, last_login=%v, rid: %s", 
				updatedUser.LoginCount, updatedUser.LastLogin, rid)
		}
	}

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
		s.renderOIDCErrorPage(c, "会话保存失败，请重试")
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

// saveBasicOIDCSession 保存基本的OIDC会话信息，主要用于logout时能正确清除SSO状态
func (s *Service) saveBasicOIDCSession(c *gin.Context, idToken, userName string) {
	session := sessions.Default(c)

	// 只保存logout时必需的最基本信息
	session.Set("oidc_id_token", idToken)
	session.Set("oidc_username", userName)

	// 保存会话
	if err := session.Save(); err != nil {
		blog.Errorf("failed to save basic OIDC session for user %s, err: %s", userName, err.Error())
	} else {
		blog.Infof("saved basic OIDC session for user %s to enable proper logout", userName)
	}
}

// renderOIDCErrorPage 渲染OIDC错误页面
func (s *Service) renderOIDCErrorPage(c *gin.Context, errorMessage string) {
	// 获取session以构造logout URL
	session := sessions.Default(c)
	idToken, _ := session.Get("oidc_id_token").(string)

	// 构造OIDC logout URL
	logoutURL := s.Config.OIDC.LogoutUrl
	if idToken != "" {
		logoutURL = fmt.Sprintf("%s?id_token_hint=%s&post_logout_redirect_uri=%s",
			s.Config.OIDC.LogoutUrl,
			url.QueryEscape(idToken),
			url.QueryEscape(s.Config.OIDC.RedirectUri))
	}

	c.Header("Content-Type", "text/html; charset=utf-8")
	c.String(200, `
<!DOCTYPE html>
<html lang="zh-CN">
<head>
    <meta charset="utf-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>蓝鲸配置平台 - 登录失败</title>
    <style>
        * {
            margin: 0;
            padding: 0;
            box-sizing: border-box;
        }
        
        body {
            font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', Roboto, 'Helvetica Neue', Arial, sans-serif;
            background: linear-gradient(135deg, #667eea 0%%, #764ba2 100%%);
            min-height: 100vh;
            display: flex;
            align-items: center;
            justify-content: center;
            color: #333;
        }
        
        .container {
            background: white;
            border-radius: 16px;
            box-shadow: 0 20px 60px rgba(0, 0, 0, 0.1);
            padding: 60px 40px;
            text-align: center;
            max-width: 480px;
            width: 90%%;
            position: relative;
            overflow: hidden;
        }
        
        .container::before {
            content: '';
            position: absolute;
            top: 0;
            left: 0;
            right: 0;
            height: 4px;
            background: linear-gradient(90deg, #ff6b6b, #feca57, #48cae4, #a8e6cf);
        }
        
        .icon {
            width: 80px;
            height: 80px;
            margin: 0 auto 30px;
            background: linear-gradient(135deg, #ff6b6b, #ff8e8e);
            border-radius: 50%%;
            display: flex;
            align-items: center;
            justify-content: center;
            color: white;
            font-size: 36px;
        }
        
        h1 {
            font-size: 28px;
            font-weight: 600;
            color: #2c3e50;
            margin-bottom: 16px;
        }
        
        .error-message {
            color: #e74c3c;
            font-size: 16px;
            line-height: 1.5;
            margin: 24px 0 40px;
            padding: 20px;
            background: #fdf2f2;
            border: 1px solid #fecaca;
            border-radius: 8px;
            border-left: 4px solid #e74c3c;
        }
        
        .buttons {
            display: flex;
            gap: 16px;
            justify-content: center;
            flex-wrap: wrap;
        }
        
        .btn {
            display: inline-flex;
            align-items: center;
            justify-content: center;
            padding: 12px 24px;
            border: none;
            border-radius: 8px;
            font-size: 14px;
            font-weight: 500;
            text-decoration: none;
            cursor: pointer;
            transition: all 0.3s ease;
            min-width: 140px;
            position: relative;
            overflow: hidden;
        }
        
        .btn-primary {
            background: linear-gradient(135deg, #667eea, #764ba2);
            color: white;
        }
        
        .btn-primary:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(102, 126, 234, 0.4);
        }
        
        .btn-success {
            background: linear-gradient(135deg, #56ab2f, #a8e6cf);
            color: white;
        }
        
        .btn-success:hover {
            transform: translateY(-2px);
            box-shadow: 0 8px 25px rgba(86, 171, 47, 0.4);
        }
        
        .btn:active {
            transform: translateY(0);
        }
        
        .footer {
            margin-top: 40px;
            color: #7f8c8d;
            font-size: 14px;
        }
        
        .footer a {
            color: #667eea;
            text-decoration: none;
        }
        
        .footer a:hover {
            text-decoration: underline;
        }
        
        @media (max-width: 640px) {
            .container {
                padding: 40px 20px;
                margin: 20px;
            }
            
            .buttons {
                flex-direction: column;
                align-items: center;
            }
            
            .btn {
                width: 100%%;
                max-width: 200px;
            }
        }
        
        /* 动画效果 */
        @keyframes fadeInUp {
            from {
                opacity: 0;
                transform: translateY(30px);
            }
            to {
                opacity: 1;
                transform: translateY(0);
            }
        }
        
        .container {
            animation: fadeInUp 0.6s ease-out;
        }
    </style>
</head>
<body>
    <div class="container">
        <div class="icon">⚠️</div>
        <h1>登录失败</h1>
        <div class="error-message">%s</div>
        <div class="buttons">
            <button class="btn btn-primary" onclick="handleLogout()">
                🔐 返回登录页面
            </button>
            <button class="btn btn-success" onclick="refreshPage()">
                🔄 刷新页面
            </button>
        </div>
        <div class="footer">
            如有疑问，请联系系统管理员<br>
            <a href="mailto:jason.zhang@frieslandcampina.com">jason.zhang@frieslandcampina.com</a>
        </div>
    </div>
    
    <script>
        function handleLogout() {
            // 添加加载状态
            const btn = event.target;
            const originalText = btn.innerHTML;
            btn.innerHTML = '⏳ 跳转中...';
            btn.disabled = true;
            
            // 直接跳转到OIDC logout URL，清理SSO状态
            setTimeout(() => {
                window.location.href = '%s';
            }, 500);
        }
        
        function refreshPage() {
            // 添加加载状态
            const btn = event.target;
            const originalText = btn.innerHTML;
            btn.innerHTML = '⏳ 刷新中...';
            btn.disabled = true;
            
            // 跳转到首页
            setTimeout(() => {
                window.location.href = '/';
            }, 500);
        }
        
        // 添加键盘快捷键支持
        document.addEventListener('keydown', function(e) {
            if (e.key === 'Enter') {
                handleLogout();
            } else if (e.key === 'F5' || (e.ctrlKey && e.key === 'r')) {
                e.preventDefault();
                refreshPage();
            }
        });
    </script>
</body>
</html>`, errorMessage, logoutURL)
}

func generateBkToken(username string, expireTime int64) string {
	// 使用用户名和过期时间生成一致的token
	data := fmt.Sprintf("oidc:%s:%d", username, expireTime)
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)[:32]
}
