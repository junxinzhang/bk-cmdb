import rolePermissionApi from './api/role-permission'

const state = {
  roles: [],
  permissions: [
    {
      key: 'home',
      label: '首页',
      description: '访问系统首页，查看系统概览和统计数据',
      category: 'basic'
    },
    {
      key: 'business',
      label: '业务',
      description: '管理业务拓扑、服务实例、进程配置等业务相关功能',
      category: 'business'
    },
    {
      key: 'resource',
      label: '资源',
      description: '管理主机资源、云区域配置、资源池等基础资源',
      category: 'resource'
    },
    {
      key: 'model',
      label: '模型',
      description: '管理配置模型、对象属性、模型关联等数据模型',
      category: 'config'
    },
    {
      key: 'operation',
      label: '运营分析',
      description: '查看运营数据分析、审计日志、系统监控等运营信息',
      category: 'operation'
    },
    {
      key: 'admin',
      label: '平台管理',
      description: '系统配置、用户管理、权限设置、全局配置等管理功能',
      category: 'admin'
    }
  ],
  permissionMatrix: {
    admin: ['home', 'business', 'resource', 'model', 'operation', 'admin'],
    operator: ['home', 'business', 'resource']
  },
  roleUsers: {},
  loading: {
    roles: false,
    permissions: false,
    matrix: false,
    users: false
  }
}

const getters = {
  roles: state => state.roles,
  permissions: state => state.permissions,
  permissionMatrix: state => state.permissionMatrix,
  roleUsers: state => state.roleUsers,
  loading: state => state.loading,
  
  // 获取指定角色信息
  getRoleByKey: state => roleKey => {
    return state.roles.find(role => role.key === roleKey)
  },
  
  // 获取指定角色的权限列表
  getRolePermissions: state => roleKey => {
    return state.permissionMatrix[roleKey] || []
  },
  
  // 获取指定角色的用户数量
  getRoleUserCount: state => roleKey => {
    const users = state.roleUsers[roleKey]
    return users ? users.length : 0
  },
  
  // 检查角色是否拥有指定权限
  hasRolePermission: state => (roleKey, permissionKey) => {
    const rolePermissions = state.permissionMatrix[roleKey] || []
    return rolePermissions.includes(permissionKey)
  },
  
  // 根据分类获取权限
  getPermissionsByCategory: state => category => {
    return state.permissions.filter(permission => permission.category === category)
  },
  
  // 获取权限详情
  getPermissionByKey: state => permissionKey => {
    return state.permissions.find(permission => permission.key === permissionKey)
  },
  
  // 获取所有权限分类
  permissionCategories: state => {
    const categories = [...new Set(state.permissions.map(p => p.category))]
    return categories.map(category => ({
      key: category,
      permissions: state.permissions.filter(p => p.category === category)
    }))
  }
}

const mutations = {
  setRoles(state, roles) {
    // 确保 roles 是数组
    const roleList = Array.isArray(roles) ? roles : []
    state.roles = roleList.map(role => ({
      ...role,
      theme: (role.key === 'admin' || role.role_name === 'admin') ? 'danger' : 'info'
    }))
  },
  
  setPermissions(state, permissions) {
    // 确保 permissions 是数组
    state.permissions = Array.isArray(permissions) ? permissions : []
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
    
    // 更新角色信息中的权限，支持多种匹配方式
    const role = state.roles.find(r => 
      r.key === roleKey || 
      r.role_name === roleKey ||
      r.name === roleKey
    )
    if (role) {
      role.permissions = permissions
    }
  },
  
  addRole(state, role) {
    state.roles.push({
      ...role,
      theme: role.key === 'admin' ? 'danger' : 'info',
      userCount: 0
    })
  },
  
  updateRole(state, updatedRole) {
    const index = state.roles.findIndex(role => 
      role.key === updatedRole.key || 
      role.role_name === updatedRole.role_name ||
      role.key === updatedRole.role_name
    )
    if (index !== -1) {
      state.roles.splice(index, 1, {
        ...updatedRole,
        theme: (updatedRole.key === 'admin' || updatedRole.role_name === 'admin') ? 'danger' : 'info'
      })
    }
  },
  
  removeRole(state, roleKey) {
    state.roles = state.roles.filter(role => 
      role.key !== roleKey && 
      role.role_name !== roleKey
    )
    delete state.permissionMatrix[roleKey]
    delete state.roleUsers[roleKey]
  },
  
  setLoading(state, { type, loading }) {
    state.loading = {
      ...state.loading,
      [type]: loading
    }
  }
}

const actions = {
  // 获取所有角色列表
  async getRoles({ commit }) {
    commit('setLoading', { type: 'roles', loading: true })
    try {
      const roles = await rolePermissionApi.actions.getRoles({ commit })
      commit('setRoles', roles)
      return roles
    } finally {
      commit('setLoading', { type: 'roles', loading: false })
    }
  },
  
  // 获取所有权限列表
  async getPermissions({ commit }) {
    commit('setLoading', { type: 'permissions', loading: true })
    try {
      const permissions = await rolePermissionApi.actions.getPermissions({ commit })
      commit('setPermissions', permissions)
      return permissions
    } finally {
      commit('setLoading', { type: 'permissions', loading: false })
    }
  },
  
  // 获取权限矩阵
  async getPermissionMatrix({ commit }) {
    commit('setLoading', { type: 'matrix', loading: true })
    try {
      const matrix = await rolePermissionApi.actions.getPermissionMatrix({ commit })
      commit('setPermissionMatrix', matrix)
      return matrix
    } finally {
      commit('setLoading', { type: 'matrix', loading: false })
    }
  },
  
  // 更新权限矩阵
  async updatePermissionMatrix({ commit }, matrix) {
    commit('setLoading', { type: 'matrix', loading: true })
    try {
      await rolePermissionApi.actions.updatePermissionMatrix({ commit }, matrix)
      commit('setPermissionMatrix', matrix)
      return true
    } finally {
      commit('setLoading', { type: 'matrix', loading: false })
    }
  },
  
  // 获取指定角色的用户列表
  async getRoleUsers({ commit }, roleKey) {
    commit('setLoading', { type: 'users', loading: true })
    try {
      const users = await rolePermissionApi.actions.getRoleUsers({ commit }, roleKey)
      commit('setRoleUsers', { roleKey, users })
      return users
    } finally {
      commit('setLoading', { type: 'users', loading: false })
    }
  },
  
  // 更新角色权限
  async updateRolePermissions({ commit }, { roleKey, permissions }) {
    try {
      await rolePermissionApi.actions.updateRolePermissions({ commit }, { roleKey, permissions })
      commit('updateRolePermissions', { roleKey, permissions })
      return true
    } catch (error) {
      throw error
    }
  },
  
  // 创建自定义角色
  async createRole({ commit }, roleData) {
    try {
      const role = await rolePermissionApi.actions.createRole({ commit }, roleData)
      commit('addRole', role)
      return role
    } catch (error) {
      throw error
    }
  },
  
  // 更新角色信息
  async updateRole({ commit }, { roleKey, ...roleData }) {
    try {
      const role = await rolePermissionApi.actions.updateRole({ commit }, { roleKey, ...roleData })
      commit('updateRole', role)
      return role
    } catch (error) {
      throw error
    }
  },
  
  // 删除角色
  async deleteRole({ commit }, roleKey) {
    try {
      await rolePermissionApi.actions.deleteRole({ commit }, roleKey)
      commit('removeRole', roleKey)
      return true
    } catch (error) {
      throw error
    }
  },
  
  // 批量分配角色给用户
  async assignRoleToUsers({ commit }, { roleKey, userIds }) {
    try {
      return await rolePermissionApi.actions.assignRoleToUsers(null, { roleKey, userIds })
    } catch (error) {
      throw error
    }
  },
  
  // 从角色中移除用户
  async removeUsersFromRole({ commit }, { roleKey, userIds }) {
    try {
      return await rolePermissionApi.actions.removeUsersFromRole(null, { roleKey, userIds })
    } catch (error) {
      throw error
    }
  },
  
  // 获取角色统计信息
  async getRoleStatistics({ commit }) {
    try {
      return await rolePermissionApi.actions.getRoleStatistics()
    } catch (error) {
      throw error
    }
  },
  
  // 验证权限配置
  async validatePermissionConfig({ commit }, { roleKey, permissions }) {
    try {
      return await rolePermissionApi.actions.validatePermissionConfig(null, { roleKey, permissions })
    } catch (error) {
      throw error
    }
  },
  
  // 获取权限依赖关系
  async getPermissionDependencies({ commit }) {
    try {
      return await rolePermissionApi.actions.getPermissionDependencies()
    } catch (error) {
      throw error
    }
  },
  
  // 检查用户权限
  async checkUserPermission({ commit }, { userId, permission }) {
    try {
      return await rolePermissionApi.actions.checkUserPermission(null, { userId, permission })
    } catch (error) {
      throw error
    }
  },
  
  // 初始化权限数据
  async initPermissionData({ dispatch }) {
    try {
      await Promise.all([
        dispatch('getRoles'),
        dispatch('getPermissions'),
        dispatch('getPermissionMatrix')
      ])
      return true
    } catch (error) {
      throw error
    }
  },
  
  // 刷新角色用户数量
  async refreshRoleUserCounts({ commit, state }) {
    try {
      const promises = state.roles.map(role => 
        rolePermissionApi.actions.getRoleUsers({ commit }, role.key)
          .then(users => ({ roleKey: role.key, users }))
      )
      
      const results = await Promise.all(promises)
      results.forEach(({ roleKey, users }) => {
        commit('setRoleUsers', { roleKey, users })
      })
      
      return true
    } catch (error) {
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