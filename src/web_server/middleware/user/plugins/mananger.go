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

// Package plugins defines all login method related plugins
package plugins

import (
	"configcenter/src/common"
	"configcenter/src/common/metadata"
	"configcenter/src/web_server/middleware/user/plugins/manager"
	// register plugins
	_ "configcenter/src/web_server/middleware/user/plugins/register"
)

// CurrentPlugin get current login plugin
func CurrentPlugin(version string) metadata.LoginUserPluginInerface {
	if "" == version {
		version = common.BKBluekingLoginPluginVersion
	}

	var defaultPlugin *metadata.LoginPluginInfo
	var opensourcePlugin *metadata.LoginPluginInfo

	for _, plugin := range manager.LoginPluginInfo {
		if plugin.Version == version {
			return plugin.HandleFunc
		}
		if common.BKBluekingLoginPluginVersion == plugin.Version {
			defaultPlugin = plugin
		}
		if common.BKOpenSourceLoginPluginVersion == plugin.Version {
			opensourcePlugin = plugin
		}
	}

	if nil != defaultPlugin {
		return defaultPlugin.HandleFunc
	}

	// 回退到开源插件
	if nil != opensourcePlugin {
		return opensourcePlugin.HandleFunc
	}

	return nil
}
