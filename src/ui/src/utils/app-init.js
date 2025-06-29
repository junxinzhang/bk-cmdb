/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2022 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

import { initUserPermissions, permissionDirective } from '@/utils/permission'

/**
 * 应用初始化函数
 * @param {Vue} app - Vue应用实例
 */
export async function initializeApp(app) {
  // 1. 注册权限指令
  app.directive('permission', permissionDirective)
  
  // 2. 初始化用户权限
  try {
    const permissionResult = await initUserPermissions()
    
    if (permissionResult.isLoggedIn) {
      console.log('用户权限初始化成功:', permissionResult.permissions)
    } else {
      console.log('用户未登录或权限获取失败')
    }
    
    return permissionResult
  } catch (error) {
    console.error('应用初始化失败:', error)
    return {
      isLoggedIn: false,
      permissions: [],
      error: error.message
    }
  }
}

/**
 * 权限状态更新函数
 * 用于在用户登录/登出时更新权限
 */
export async function refreshUserPermissions() {
  try {
    const permissionResult = await initUserPermissions()
    console.log('用户权限刷新成功:', permissionResult)
    return permissionResult
  } catch (error) {
    console.error('权限刷新失败:', error)
    return {
      isLoggedIn: false,
      permissions: [],
      error: error.message
    }
  }
}