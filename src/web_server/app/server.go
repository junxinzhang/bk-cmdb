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

package app

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"configcenter/src/apimachinery/apiserver"
	"configcenter/src/apimachinery/rest"
	"configcenter/src/common"
	"configcenter/src/common/backbone"
	cc "configcenter/src/common/backbone/configcenter"
	apigwcli "configcenter/src/common/resource/apigw"
	"configcenter/src/common/resource/esb"
	"configcenter/src/common/resource/jwt"
	"configcenter/src/common/types"
	"configcenter/src/storage/dal/redis"
	"configcenter/src/thirdparty/apigw"
	"configcenter/src/web_server/app/options"
	webcomm "configcenter/src/web_server/common"
	"configcenter/src/web_server/logics"
	websvc "configcenter/src/web_server/service"
)

// WebServer TODO
type WebServer struct {
	Config options.Config
}

// Run web-server
func Run(ctx context.Context, cancel context.CancelFunc, op *options.ServerOption) error {
	svrInfo, err := types.NewServerInfo(op.ServConf)
	if err != nil {
		return fmt.Errorf("wrap server info failed, err: %v", err)
	}

	webSvr := new(WebServer)
	webSvr.Config.DeploymentMethod = op.DeploymentMethod

	input := &backbone.BackboneParameter{
		ConfigUpdate: webSvr.onServerConfigUpdate,
		ConfigPath:   op.ServConf.ExConfig,
		SrvRegdiscv:  backbone.SrvRegdiscv{Regdiscv: op.ServConf.RegDiscover, TLSConfig: op.ServConf.GetTLSClientConf()},
		SrvInfo:      svrInfo,
	}
	if op.DeploymentMethod == common.BluekingDeployment {
		input.Disable = true
	}

	engine, err := backbone.NewBackbone(ctx, input)
	if err != nil {
		return fmt.Errorf("new backbone failed, err: %v", err)
	}

	configReady := false
	for sleepCnt := 0; sleepCnt < common.APPConfigWaitTime; sleepCnt++ {
		if "" == webSvr.Config.Site.DomainUrl {
			time.Sleep(time.Second)
		} else {
			configReady = true
			break
		}
	}
	if !configReady {
		return errors.New("configuration item not found")
	}

	service, err := initWebService(webSvr, engine)
	if err != nil {
		return err
	}

	if err := service.InitNotice(); err != nil {
		return err
	}

	err = backbone.StartServer(ctx, cancel, engine, service.WebService(), false)
	if err != nil {
		return err
	}

	select {
	case <-ctx.Done():
	}

	return nil
}

func initWebService(webSvr *WebServer, engine *backbone.Engine) (*websvc.Service, error) {
	service := new(websvc.Service)

	var err error
	webSvr.Config.Redis, err = engine.WithRedis()
	if err != nil {
		return nil, err
	}

	var redisErr error
	if webSvr.Config.Redis.MasterName == "" {
		// MasterName 为空，表示使用直连redis 。 使用Host,Port 做链接redis参数
		service.Session, redisErr = redis.NewRedisStore(10, "tcp", webSvr.Config.Redis.Address,
			webSvr.Config.Redis.Password, webSvr.Config.Redis.TLSConfig, []byte("secret"))
	} else {
		// MasterName 不为空，表示使用哨兵模式的redis。MasterName 是Master标记
		address := strings.Split(webSvr.Config.Redis.Address, ";")
		service.Session, redisErr = redis.NewRedisStoreWithSentinel(address, 10, webSvr.Config.Redis.MasterName, "tcp",
			webSvr.Config.Redis.Password, webSvr.Config.Redis.SentinelPassword, webSvr.Config.Redis.TLSConfig, []byte("secret"))
	}
	if redisErr != nil {
		return nil, fmt.Errorf("create new redis store failed, err: %v", redisErr)
	}

	cacheCli, err := redis.NewFromConfig(webSvr.Config.Redis)
	if err != nil {
		return nil, err
	}

	service.Engine = engine
	service.CacheCli = cacheCli
	service.Logics = &logics.Logics{Engine: engine}
	service.Config = &webSvr.Config

	// init esb client
	esb.InitEsbClient(nil)

	// init jwt handler
	if err = jwt.Init("webServer"); err != nil {
		return nil, fmt.Errorf("init jwt failed, err: %v", err)
	}

	// init api gateway client
	apigwClients := make([]apigw.ClientType, 0)
	if webSvr.Config.DeploymentMethod == common.BluekingDeployment {
		apigwClients = append(apigwClients, apigw.Cmdb)
	}
	if webSvr.Config.EnableNotification {
		apigwClients = append(apigwClients, apigw.Notice)
	}
	if len(apigwClients) > 0 {
		err = apigwcli.Init("apiGW", engine.Metric().Registry(), apigwClients)
		if err != nil {
			return nil, fmt.Errorf("init api gateway client error, err: %v", err)
		}
	}

	// init api client
	switch webSvr.Config.DeploymentMethod {
	case common.BluekingDeployment:
		cmdbCli := apigwcli.Client().Cmdb()
		headerWrapper := rest.HeaderWrapper(cmdbCli.SetApiGWAuthHeader)
		baseUrlWrapper := rest.BaseUrlWrapper(fmt.Sprintf("/api/%s/", webcomm.API_VERSION))
		service.ApiCli = apiserver.NewWrappedApiServerClientI(cmdbCli.Client(), baseUrlWrapper, headerWrapper)
	default:
		service.ApiCli = engine.CoreAPI.ApiServer()
	}
	service.Logics.ApiCli = service.ApiCli
	// Add CoreAPI alias for compatibility with user_management.go
	service.Logics.CoreAPI = engine.CoreAPI

	return service, nil
}

func (w *WebServer) onServerConfigUpdate(previous, current cc.ProcessConfig) {
	domainUrl, _ := cc.String("webServer.site.domainUrl")
	w.Config.Site.DomainUrl = domainUrl + "/"
	w.Config.Site.HtmlRoot, _ = cc.String("webServer.site.htmlRoot")
	w.Config.Site.ResourcesPath, _ = cc.String("webServer.site.resourcesPath")
	w.Config.Site.BkLoginUrl, _ = cc.String("webServer.site.bkLoginUrl")
	w.Config.Site.AppCode, _ = cc.String("webServer.site.appCode")
	w.Config.Site.CheckUrl, _ = cc.String("webServer.site.checkUrl")

	authscheme, err := cc.String("webServer.site.authscheme")
	if err != nil {
		w.Config.Site.AuthScheme = "internal"
	} else {
		w.Config.Site.AuthScheme = authscheme
	}

	fullTextSearch, err := cc.String("es.fullTextSearch")
	if err != nil {
		w.Config.Site.FullTextSearch = "off"
	} else {
		w.Config.Site.FullTextSearch = fullTextSearch
	}

	w.Config.Site.AccountUrl, _ = cc.String("webServer.site.bkAccountUrl")
	w.Config.Site.BkHttpsLoginUrl, _ = cc.String("webServer.site.bkHttpsLoginUrl")
	w.Config.Site.HttpsDomainUrl, _ = cc.String("webServer.site.httpsDomainUrl")
	w.Config.Site.PaasDomainUrl, _ = cc.String("webServer.site.paasDomainUrl")
	w.Config.Site.BkDomain, _ = cc.String("webServer.site.bkDomain")
	w.Config.Site.HelpDocUrl, _ = cc.String("webServer.site.helpDocUrl")
	w.Config.Site.BkComponentApiUrl, _ = cc.String("webServer.site.bkComponentApiUrl")
	w.Config.Site.BkSharedResUrl, _ = cc.String("webServer.site.bkSharedResUrl")

	w.Config.Session.Name, _ = cc.String("webServer.session.name")
	w.Config.Session.MultipleOwner, _ = cc.String("webServer.session.multipleOwner")
	w.Config.Session.DefaultLanguage, _ = cc.String("webServer.session.defaultlanguage")
	w.Config.LoginVersion, _ = cc.String("webServer.login.version")
	if "" == w.Config.Session.DefaultLanguage {
		w.Config.Session.DefaultLanguage = "zh-cn"
	}

	w.Config.Version, _ = cc.String("webServer.api.version")
	w.Config.AgentAppUrl, _ = cc.String("webServer.app.agentAppUrl")
	w.Config.AuthCenter.AppCode, _ = cc.String("webServer.app.authAppCode")
	w.Config.AuthCenter.URL, _ = cc.String("webServer.app.authUrl")
	w.Config.LoginUrl = fmt.Sprintf(w.Config.Site.BkLoginUrl, w.Config.Site.AppCode, w.Config.Site.DomainUrl)
	if esbConfig, err := esb.ParseEsbConfig(); err == nil {
		esb.UpdateEsbConfig(*esbConfig)
	}
	w.Config.DisableOperationStatistic, _ = cc.Bool("operationServer.disableOperationStatistic")

	w.Config.EnableNotification, _ = cc.Bool("webServer.enableNotification")

	// OIDC 配置
	w.Config.OIDC.Issuer, _ = cc.String("webServer.oidc.issuer")
	w.Config.OIDC.ClientId, _ = cc.String("webServer.oidc.clientId")
	w.Config.OIDC.ClientSecret, _ = cc.String("webServer.oidc.clientSecret")
	w.Config.OIDC.RedirectUri, _ = cc.String("webServer.oidc.redirectUri")
	w.Config.OIDC.AuthUrl, _ = cc.String("webServer.oidc.authUrl")
	w.Config.OIDC.TokenUrl, _ = cc.String("webServer.oidc.tokenUrl")
	w.Config.OIDC.UserInfoUrl, _ = cc.String("webServer.oidc.userInfoUrl")
	w.Config.OIDC.LogoutUrl, _ = cc.String("webServer.oidc.logoutUrl")
	w.Config.OIDC.Scopes, _ = cc.String("webServer.oidc.scopes")
	w.Config.OIDC.AllowedUsers, _ = cc.String("webServer.oidc.allowedUsers")
}

// Stop the ccapi server
func (ccWeb *WebServer) Stop() error {
	return nil
}
