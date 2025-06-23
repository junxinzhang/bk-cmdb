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
  getRolePermissions: state => roleKey => {
    return state.permissionMatrix[roleKey] || []
  },
  
  // 获取指定角色的用户数量
  getRoleUserCount: state => roleKey => {
    return state.roleUsers[roleKey]?.length || 0
  },
  
  // 检查角色是否拥有指定权限
  hasRolePermission: state => (roleKey, permissionKey) => {
    const rolePermissions = state.permissionMatrix[roleKey] || []
    return rolePermissions.includes(permissionKey)
  }
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
    state.roles = state.roles.filter(role => role.key !== roleKey)
  }
}

const actions = {
  // 获取所有角色列表
  async getRoles({ commit }) {
    try {
      const response = await $http.get('/role/list')
      const roles = response.data || []
      
      // 为每个角色添加主题色和用户数量
      const enrichedRoles = roles.map(role => ({
        ...role,
        theme: role.key === 'admin' ? 'danger' : 'info',
        userCount: 0 // 初始化，后续会通过其他接口获取
      }))
      
      commit('setRoles', enrichedRoles)
      return enrichedRoles
    } catch (error) {
      console.error('获取角色列表失败:', error)
      throw error
    }
  },

  // 获取所有可用权限列表
  async getPermissions({ commit }) {
    try {
      const response = await $http.get('/permission/list')
      commit('setPermissions', response.data || [])
      return response.data
    } catch (error) {
      console.error('获取权限列表失败:', error)
      throw error
    }
  },

  // 获取权限矩阵
  async getPermissionMatrix({ commit }) {
    try {
      const response = await $http.get('/permission/matrix')
      commit('setPermissionMatrix', response.data || {})
      return response.data
    } catch (error) {
      console.error('获取权限矩阵失败:', error)
      throw error
    }
  },

  // 更新权限矩阵
  async updatePermissionMatrix({ commit }, matrix) {
    try {
      const response = await $http.put('/permission/matrix', {
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
      const response = await $http.get(`/role/${roleKey}/users`)
      const users = response.data || []
      commit('setRoleUsers', { roleKey, users })
      return users
    } catch (error) {
      console.error('获取角色用户列表失败:', error)
      throw error
    }
  },

  // 更新角色权限
  async updateRolePermissions({ commit }, { roleKey, permissions }) {
    try {
      const response = await $http.put(`/role/${roleKey}/permissions`, {
        permissions
      })
      commit('updateRolePermissions', { roleKey, permissions })
      return response.data
    } catch (error) {
      console.error('更新角色权限失败:', error)
      throw error
    }
  },

  // 创建自定义角色
  async createRole({ commit }, roleData) {
    try {
      const response = await $http.post('/role/create', {
        key: roleData.key,
        name: roleData.name,
        description: roleData.description,
        permissions: roleData.permissions || []
      })
      commit('addRole', response.data)
      return response.data
    } catch (error) {
      console.error('创建角色失败:', error)
      throw error
    }
  },

  // 更新角色信息
  async updateRole({ commit }, { roleKey, ...roleData }) {
    try {
      const response = await $http.put(`/role/${roleKey}`, {
        name: roleData.name,
        description: roleData.description
      })
      commit('updateRole', response.data)
      return response.data
    } catch (error) {
      console.error('更新角色失败:', error)
      throw error
    }
  },

  // 删除角色
  async deleteRole({ commit }, roleKey) {
    try {
      await $http.delete(`/role/${roleKey}`)
      commit('removeRole', roleKey)
      return true
    } catch (error) {
      console.error('删除角色失败:', error)
      throw error
    }
  },

  // 批量分配角色给用户
  async assignRoleToUsers({ commit }, { roleKey, userIds }) {
    try {
      const response = await $http.post(`/role/${roleKey}/assign`, {
        user_ids: userIds
      })
      return response.data
    } catch (error) {
      console.error('分配角色失败:', error)
      throw error
    }
  },

  // 从角色中移除用户
  async removeUsersFromRole({ commit }, { roleKey, userIds }) {
    try {
      const response = await $http.delete(`/role/${roleKey}/users`, {
        data: { user_ids: userIds }
      })
      return response.data
    } catch (error) {
      console.error('移除用户角色失败:', error)
      throw error
    }
  },

  // 获取角色统计信息
  async getRoleStatistics({ commit }) {
    try {
      const response = await $http.get('/role/statistics')
      return response.data
    } catch (error) {
      console.error('获取角色统计信息失败:', error)
      throw error
    }
  },

  // 验证权限配置
  async validatePermissionConfig({ commit }, { roleKey, permissions }) {
    try {
      const response = await $http.post('/permission/validate', {
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
  async getPermissionDependencies({ commit }) {
    try {
      const response = await $http.get('/permission/dependencies')
      return response.data
    } catch (error) {
      console.error('获取权限依赖关系失败:', error)
      throw error
    }
  },

  // 检查用户权限
  async checkUserPermission({ commit }, { userId, permission }) {
    try {
      const response = await $http.post('/permission/check', {
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