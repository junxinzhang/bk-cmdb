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

/**
 * 权限管理工具类
 */
class PermissionManager {
  constructor() {
    this.permissions = []
    this.isLoggedIn = false
  }

  /**
   * 设置用户权限
   * @param {Array} permissions - 权限数组
   */
  setPermissions(permissions) {
    this.permissions = Array.isArray(permissions) ? permissions : []
  }

  /**
   * 设置登录状态
   * @param {boolean} status - 登录状态
   */
  setLoginStatus(status) {
    this.isLoggedIn = status
  }

  /**
   * 检查用户是否有指定权限
   * @param {string} permission - 权限标识
   * @returns {boolean}
   */
  hasPermission(permission) {
    if (!this.isLoggedIn) {
      return false
    }
    return this.permissions.includes(permission)
  }

  /**
   * 检查用户是否有任意一个权限
   * @param {Array} permissions - 权限标识数组
   * @returns {boolean}
   */
  hasAnyPermission(permissions) {
    if (!this.isLoggedIn) {
      return false
    }
    return permissions.some(permission => this.permissions.includes(permission))
  }

  /**
   * 检查用户是否有所有权限
   * @param {Array} permissions - 权限标识数组
   * @returns {boolean}
   */
  hasAllPermissions(permissions) {
    if (!this.isLoggedIn) {
      return false
    }
    return permissions.every(permission => this.permissions.includes(permission))
  }

  /**
   * 获取当前用户的所有权限
   * @returns {Array}
   */
  getAllPermissions() {
    return [...this.permissions]
  }

  /**
   * 清空权限（登出时调用）
   */
  clearPermissions() {
    this.permissions = []
    this.isLoggedIn = false
  }
}

// 创建全局权限管理实例
const permissionManager = new PermissionManager()

/**
 * 菜单权限映射配置
 * 将菜单ID映射到对应的权限标识
 */
export const MENU_PERMISSION_MAP = {
  // 一级菜单权限映射
  'menu-index': 'home',
  'menu-business': 'business', 
  'menu-business-set': 'business',
  'menu-resource': 'resource',
  'menu-model': 'model',
  'menu-analysis': 'operation',
  'menu-platform-management': 'admin'
}

/**
 * 从is_login接口获取用户权限并初始化权限管理器
 * @returns {Promise<Object>} 返回登录状态和权限信息
 */
export async function initUserPermissions() {
  try {
    const response = await fetch('/is_login', {
      method: 'GET',
      credentials: 'include'
    })
    
    if (!response.ok) {
      throw new Error(`HTTP error! status: ${response.status}`)
    }
    
    const data = await response.json()
    
    if (data.result) {
      // 用户已登录，设置权限
      permissionManager.setLoginStatus(true)
      permissionManager.setPermissions(data.permission || [])
      
      return {
        isLoggedIn: true,
        permissions: data.permission || [],
        userData: data
      }
    } else {
      // 用户未登录
      permissionManager.setLoginStatus(false)
      permissionManager.clearPermissions()
      
      return {
        isLoggedIn: false,
        permissions: [],
        userData: data
      }
    }
  } catch (error) {
    console.error('Failed to fetch user permissions:', error)
    permissionManager.clearPermissions()
    
    return {
      isLoggedIn: false,
      permissions: [],
      error: error.message
    }
  }
}

/**
 * 检查菜单是否应该显示
 * @param {string} menuId - 菜单ID
 * @returns {boolean}
 */
export function shouldShowMenu(menuId) {
  const requiredPermission = MENU_PERMISSION_MAP[menuId]
  
  // 如果没有配置权限要求，默认显示
  if (!requiredPermission) {
    return true
  }
  
  return permissionManager.hasPermission(requiredPermission)
}

/**
 * 检查权限的Vue指令
 * 使用方法: v-permission="'admin'" 或 v-permission="['admin', 'operation']"
 */
export const permissionDirective = {
  inserted(el, binding) {
    const { value } = binding
    
    if (!value) {
      return
    }
    
    let hasPermission = false
    
    if (typeof value === 'string') {
      hasPermission = permissionManager.hasPermission(value)
    } else if (Array.isArray(value)) {
      hasPermission = permissionManager.hasAnyPermission(value)
    }
    
    if (!hasPermission) {
      el.style.display = 'none'
    }
  },
  
  update(el, binding) {
    const { value } = binding
    
    if (!value) {
      return
    }
    
    let hasPermission = false
    
    if (typeof value === 'string') {
      hasPermission = permissionManager.hasPermission(value)
    } else if (Array.isArray(value)) {
      hasPermission = permissionManager.hasAnyPermission(value)
    }
    
    el.style.display = hasPermission ? '' : 'none'
  }
}

/**
 * 权限检查混入
 * 在Vue组件中使用：mixins: [permissionMixin]
 */
export const permissionMixin = {
  methods: {
    $hasPermission(permission) {
      return permissionManager.hasPermission(permission)
    },
    
    $hasAnyPermission(permissions) {
      return permissionManager.hasAnyPermission(permissions)
    },
    
    $hasAllPermissions(permissions) {
      return permissionManager.hasAllPermissions(permissions)
    },
    
    $getAllPermissions() {
      return permissionManager.getAllPermissions()
    },
    
    $shouldShowMenu(menuId) {
      return shouldShowMenu(menuId)
    }
  }
}

// 导出权限管理器实例
export default permissionManager