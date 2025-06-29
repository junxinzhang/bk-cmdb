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

import { checkMenuPermission } from '@/utils/login-helper'

/**
 * 权限控制指令
 * 用法：
 * v-permission="'MENU_INDEX'" - 检查菜单权限
 * v-permission:login - 检查登录状态
 * v-permission:menu="'MENU_PLATFORM_MANAGEMENT'" - 检查特定菜单权限
 */
export default {
  inserted(el, binding, vnode) {
    const { arg, value } = binding
    const store = vnode.context.$store
    
    let hasPermission = false
    
    if (arg === 'login') {
      // 检查登录状态
      hasPermission = store.getters.isLogin
    } else if (arg === 'menu' && value) {
      // 检查菜单权限
      hasPermission = checkMenuPermission(store, value)
    } else if (value) {
      // 默认检查菜单权限
      hasPermission = checkMenuPermission(store, value)
    } else {
      // 默认检查登录状态
      hasPermission = store.getters.isLogin
    }
    
    if (!hasPermission) {
      // 如果没有权限，隐藏元素
      el.style.display = 'none'
      // 或者移除元素（可选）
      // el.parentNode && el.parentNode.removeChild(el)
    }
  },
  
  update(el, binding, vnode) {
    const { arg, value } = binding
    const store = vnode.context.$store
    
    let hasPermission = false
    
    if (arg === 'login') {
      hasPermission = store.getters.isLogin
    } else if (arg === 'menu' && value) {
      hasPermission = checkMenuPermission(store, value)
    } else if (value) {
      hasPermission = checkMenuPermission(store, value)
    } else {
      hasPermission = store.getters.isLogin
    }
    
    if (hasPermission) {
      el.style.display = ''
    } else {
      el.style.display = 'none'
    }
  }
}