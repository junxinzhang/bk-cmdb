import $http from '@/api'

const state = {
  roles: [],
  permissions: [],
  permissionMatrix: {},
  roleUsers: {}
}

const getters = {
  roles: state => state.roles,
  permissions: state => state.permissions,
  permissionMatrix: state => state.permissionMatrix,
  roleUsers: state => state.roleUsers,

  // 获取指定角色的权限列表
  getRolePermissions: state => roleKey => state.permissionMatrix[roleKey] || [],

  // 获取指定角色的用户数量
  getRoleUserCount: state => roleKey => state.roleUsers[roleKey]?.length || 0,

  // 检查角色是否拥有指定权限
  hasRolePermission: state => (roleKey, permissionKey) => {
    const rolePermissions = state.permissionMatrix[roleKey] || []
    return rolePermissions.includes(permissionKey)
  },

  // 获取权限分类数据
  permissionCategories: () => [
    {
      key: 'basic',
      permissions: [
        {
          key: 'home',
          label: '首页',
          description: '访问系统首页和概览信息'
        }
      ]
    },
    {
      key: 'business',
      permissions: [
        {
          key: 'business',
          label: '业务',
          description: '管理和查看业务拓扑、服务实例等'
        }
      ]
    },
    {
      key: 'resource',
      permissions: [
        {
          key: 'resource',
          label: '资源',
          description: '管理主机、云区域等资源信息'
        }
      ]
    },
    {
      key: 'config',
      permissions: [
        {
          key: 'model',
          label: '模型',
          description: '管理配置模型、字段和关联关系'
        }
      ]
    },
    {
      key: 'operation',
      permissions: [
        {
          key: 'operation',
          label: '运营分析',
          description: '查看运营数据和分析报告'
        }
      ]
    },
    {
      key: 'admin',
      permissions: [
        {
          key: 'admin',
          label: '平台管理',
          description: '系统配置、用户管理等管理功能'
        }
      ]
    }
  ]
}

const mutations = {
  setRoles(state, roles) {
    state.roles = roles
  },
  setPermissions(state, permissions) {
    state.permissions = permissions
  },
  setPermissionMatrix(state, matrix) {
    state.permissionMatrix = matrix
  },
  setRoleUsers(state, { roleKey, users }) {
    state.roleUsers = {
      ...state.roleUsers,
      [roleKey]: users
    }
  },
  updateRolePermissions(state, { roleKey, permissions }) {
    state.permissionMatrix = {
      ...state.permissionMatrix,
      [roleKey]: permissions
    }
  },
  addRole(state, role) {
    state.roles.push(role)
  },
  updateRole(state, updatedRole) {
    const index = state.roles.findIndex(role => role.key === updatedRole.key)
    if (index !== -1) {
      state.roles.splice(index, 1, updatedRole)
    }
  },
  removeRole(state, roleKey) {
    state.roles = state.roles.filter((role) => {
      const currentKey = role.key || role.role_name
      return currentKey !== roleKey
    })
  }
}

const actions = {
  // 获取所有角色列表
  async getRoles({ commit }) {
    try {
      const response = await $http.get('role/list')
      console.log('API Response:', response)

      // 安全地处理API响应
      let roles = []
      if (response && response.data) {
        roles = Array.isArray(response.data) ? response.data : []
      } else if (Array.isArray(response)) {
        roles = response
      }

      console.log('Raw roles from API:', roles)

      // 为每个角色添加主题色和用户数量，统一数据格式
      const enrichedRoles = roles.map(role => ({
        ...role,
        key: role.role_name || role.key, // 使用role_name作为key
        name: role.role_name || role.name, // 确保有name字段
        role_name: role.role_name, // 保持原始字段
        description: role.description || '', // 角色描述
        theme: (role.role_name === 'admin' || role.role_name === '管理员') ? 'danger' : 'info',
        permissions: role.permissions || [], // 角色权限列表
        is_system: role.is_system || false // 是否系统角色
      }))

      console.log('Enriched roles:', enrichedRoles)
      commit('setRoles', enrichedRoles)
      return enrichedRoles
    } catch (error) {
      console.error('获取角色列表失败:', error)
      // 返回空数组而不是抛出错误，避免界面崩溃
      commit('setRoles', [])
      return []
    }
  },

  // 获取所有可用权限列表
  async getPermissions({ commit }) {
    try {
      const response = await $http.get('role/permission-matrix')
      const matrix = response.data || {}
      commit('setPermissions', matrix.permissions || [])
      return matrix.permissions
    } catch (error) {
      console.error('获取权限列表失败:', error)
      throw error
    }
  },

  // 获取权限矩阵
  async getPermissionMatrix({ commit }) {
    try {
      const response = await $http.get('role/permission-matrix')
      const matrix = response.data || {}
      commit('setPermissionMatrix', matrix.matrix || {})
      return matrix.matrix
    } catch (error) {
      console.error('获取权限矩阵失败:', error)
      throw error
    }
  },

  // 更新权限矩阵
  async updatePermissionMatrix({ commit }, matrix) {
    try {
      const response = await $http.put('role/permission-matrix', {
        matrix
      })
      commit('setPermissionMatrix', matrix)
      return response.data
    } catch (error) {
      console.error('更新权限矩阵失败:', error)
      throw error
    }
  },

  // 获取指定角色的用户列表
  async getRoleUsers({ commit }, roleKey) {
    try {
      console.log('Fetching users for role:', roleKey)
      // 修改为查询 cc_user_management 表，根据 role 字段过滤
      const response = await $http.get('user/list', {
        params: {
          role: roleKey,
          page: 1,
          limit: 1000  // 设置一个较大的限制，获取所有用户
        }
      })
      console.log('Role users response:', response)

      // 安全地处理API响应
      let users = []
      if (response && response.data) {
        // 处理分页响应格式
        if (response.data.list && Array.isArray(response.data.list)) {
          users = response.data.list
        } else if (Array.isArray(response.data)) {
          users = response.data
        }
      } else if (Array.isArray(response)) {
        users = response
      }

      console.log('Processed users:', users)

      // 确保用户数据格式统一
      const normalizedUsers = users.map(user => ({
        ...user,
        email: user.email || user.username || '',
        name: user.name || user.display_name || user.username || '',
        status: user.status || (user.is_active !== false ? 'active' : 'inactive'),
        last_login: user.last_login || user.lastLoginTime || null,
        role: user.role || roleKey  // 确保包含角色信息
      }))

      commit('setRoleUsers', { roleKey, users: normalizedUsers })
      return normalizedUsers
    } catch (error) {
      console.error('获取角色用户列表失败:', error)
      // 返回空数组而不是抛出错误，避免界面崩溃
      commit('setRoleUsers', { roleKey, users: [] })
      return []
    }
  },

  // 更新角色权限
  async updateRolePermissions({ commit }, { roleKey, permissions }) {
    try {
      console.log('Updating role permissions for:', roleKey, 'with permissions:', permissions)
      const response = await $http.put(`role/${roleKey}`, {
        permissions
      })
      console.log('Update permissions response:', response)
      commit('updateRolePermissions', { roleKey, permissions })
      return response.data || response
    } catch (error) {
      console.error('更新角色权限失败:', error)
      throw error
    }
  },

  // 创建自定义角色
  async createRole({ commit }, roleData) {
    try {
      const response = await $http.post('role/create', {
        role_name: roleData.roleName,
        permissions: roleData.permissions || [],
        description: roleData.description || ''
      })

      // 统一数据格式，添加安全检查
      const role = response.data || response || {}
      const normalizedRole = {
        ...role,
        key: role.role_name || role.key || '',
        name: role.role_name || role.name || '',
        role_name: role.role_name || '',
        theme: (role.role_name === 'admin' || role.role_name === '管理员') ? 'danger' : 'info',
        userCount: 0
      }

      commit('addRole', normalizedRole)
      return normalizedRole
    } catch (error) {
      console.error('创建角色失败:', error)
      throw error
    }
  },

  // 更新角色信息
  async updateRole({ commit }, { roleKey, ...roleData }) {
    try {
      const response = await $http.put(`role/${roleKey}`, {
        permissions: roleData.permissions,
        description: roleData.description
      })

      // 统一数据格式，添加安全检查
      const role = response.data || response || {}
      const normalizedRole = {
        ...role,
        key: role.role_name || role.key || '',
        name: role.role_name || role.name || '',
        role_name: role.role_name || '',
        theme: (role.role_name === 'admin' || role.role_name === '管理员') ? 'danger' : 'info'
      }

      commit('updateRole', normalizedRole)
      return normalizedRole
    } catch (error) {
      console.error('更新角色失败:', error)
      throw error
    }
  },

  // 删除角色
  async deleteRole({ commit }, roleKey) {
    try {
      console.log('Deleting role with key:', roleKey)
      const response = await $http.delete(`role/${roleKey}`)
      console.log('Delete role response:', response)
      commit('removeRole', roleKey)
      return response.data || true
    } catch (error) {
      console.error('删除角色失败:', error)
      throw error
    }
  },

  // 批量分配角色给用户
  async assignRoleToUsers(_, { roleKey, userIds }) {
    try {
      const response = await $http.post(`role/${roleKey}/assign`, {
        user_ids: userIds
      })
      return response.data
    } catch (error) {
      console.error('分配角色失败:', error)
      throw error
    }
  },

  // 从角色中移除用户
  async removeUsersFromRole(_, { roleKey, userIds }) {
    try {
      const response = await $http.delete(`role/${roleKey}/users`, {
        data: { user_ids: userIds }
      })
      return response.data
    } catch (error) {
      console.error('移除用户角色失败:', error)
      throw error
    }
  },

  // 获取角色统计信息
  async getRoleStatistics() {
    try {
      const response = await $http.get('role/statistics')
      return response.data
    } catch (error) {
      console.error('获取角色统计信息失败:', error)
      throw error
    }
  },

  // 验证权限配置
  async validatePermissionConfig(_, { roleKey, permissions }) {
    try {
      const response = await $http.post('permission/validate', {
        role_key: roleKey,
        permissions
      })
      return response.data
    } catch (error) {
      console.error('验证权限配置失败:', error)
      throw error
    }
  },

  // 获取权限依赖关系
  async getPermissionDependencies() {
    try {
      const response = await $http.get('permission/dependencies')
      return response.data
    } catch (error) {
      console.error('获取权限依赖关系失败:', error)
      throw error
    }
  },

  // 检查用户权限
  async checkUserPermission(_, { userId, permission }) {
    try {
      const response = await $http.post('permission/check', {
        user_id: userId,
        permission
      })
      return response.data
    } catch (error) {
      console.error('检查用户权限失败:', error)
      throw error
    }
  }
}

export default {
  namespaced: true,
  state,
  getters,
  mutations,
  actions
}
