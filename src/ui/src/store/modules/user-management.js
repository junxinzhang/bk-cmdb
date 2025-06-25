import $http from '@/api'

const state = {
  userList: [],
  pagination: {
    current: 1,
    count: 0,
    limit: 20
  },
  currentUser: null,
  searchFilters: {
    keyword: '',
    role: '',
    status: ''
  },
  loading: {
    list: false,
    detail: false,
    create: false,
    update: false,
    delete: false
  }
}

const getters = {
  userList: state => state.userList,
  pagination: state => state.pagination,
  currentUser: state => state.currentUser,
  searchFilters: state => state.searchFilters,
  loading: state => state.loading,
  
  // 根据角色筛选用户
  getUsersByRole: state => role => {
    if (!role) return state.userList
    return state.userList.filter(user => user.role === role)
  },
  
  // 根据状态筛选用户
  getUsersByStatus: state => status => {
    if (!status) return state.userList
    return state.userList.filter(user => user.status === status)
  },
  
  // 获取活跃用户数量
  activeUserCount: state => {
    return state.userList.filter(user => user.status === 'active').length
  },
  
  // 获取管理员用户数量
  adminUserCount: state => {
    return state.userList.filter(user => user.role === 'admin').length
  },
  
  // 获取操作员用户数量
  operatorUserCount: state => {
    return state.userList.filter(user => user.role === 'operator').length
  }
}

const mutations = {
  setUserList(state, { list, pagination }) {
    state.userList = list || []
    if (pagination) {
      state.pagination = {
        ...state.pagination,
        ...pagination
      }
    }
  },
  
  setCurrentUser(state, user) {
    state.currentUser = user
  },
  
  updatePagination(state, pagination) {
    state.pagination = {
      ...state.pagination,
      ...pagination
    }
  },
  
  updateSearchFilters(state, filters) {
    state.searchFilters = {
      ...state.searchFilters,
      ...filters
    }
  },
  
  setLoading(state, { type, loading }) {
    state.loading = {
      ...state.loading,
      [type]: loading
    }
  },
  
  addUser(state, user) {
    state.userList.unshift(user)
    state.pagination.count += 1
  },
  
  updateUser(state, updatedUser) {
    const userId = updatedUser.user_id || updatedUser.id
    const index = state.userList.findIndex(user => (user.user_id || user.id) === userId)
    if (index !== -1) {
      state.userList.splice(index, 1, updatedUser)
    }
    if (state.currentUser && (state.currentUser.user_id || state.currentUser.id) === userId) {
      state.currentUser = updatedUser
    }
  },
  
  removeUser(state, userId) {
    const index = state.userList.findIndex(user => (user.user_id || user.id) === userId)
    if (index !== -1) {
      state.userList.splice(index, 1)
      state.pagination.count -= 1
    }
    if (state.currentUser && (state.currentUser.user_id || state.currentUser.id) === userId) {
      state.currentUser = null
    }
  },
  
  clearUserList(state) {
    state.userList = []
    state.pagination = {
      current: 1,
      count: 0,
      limit: 20
    }
  },
  
  resetSearchFilters(state) {
    state.searchFilters = {
      keyword: '',
      role: '',
      status: ''
    }
  }
}

const actions = {
  // 获取用户列表
  async getUserList({ commit, state }, params = {}) {
    commit('setLoading', { type: 'list', loading: true })
    try {
      const requestParams = {
        page: params.page || state.pagination.current,
        limit: params.limit || state.pagination.limit,
        search: params.search || state.searchFilters.keyword,
        role: params.role || state.searchFilters.role,
        status: params.status || state.searchFilters.status
      }
      
      const response = await $http.get('usermgmt/list', {
        params: requestParams
      })
      
      // response 就是完整的API响应
      const apiData = response
      
      commit('setUserList', {
        list: apiData.items || [],
        pagination: {
          current: apiData.page || requestParams.page,
          count: apiData.total || apiData.count || 0,
          limit: apiData.limit || requestParams.limit
        }
      })
      
      return response
    } finally {
      commit('setLoading', { type: 'list', loading: false })
    }
  },
  
  // 获取用户详情
  async getUserDetail({ commit }, userId) {
    commit('setLoading', { type: 'detail', loading: true })
    try {
      const user = await this.dispatch('userManagement/api/getUserDetail', userId)
      commit('setCurrentUser', user)
      return user
    } finally {
      commit('setLoading', { type: 'detail', loading: false })
    }
  },
  
  // 创建用户
  async createUser({ commit }, userData) {
    commit('setLoading', { type: 'create', loading: true })
    try {
      const user = await this.dispatch('userManagement/api/createUser', userData)
      commit('addUser', user)
      return user
    } finally {
      commit('setLoading', { type: 'create', loading: false })
    }
  },
  
  // 更新用户
  async updateUser({ commit }, { id, ...userData }) {
    commit('setLoading', { type: 'update', loading: true })
    try {
      const user = await this.dispatch('userManagement/api/updateUser', { id, ...userData })
      commit('updateUser', user)
      return user
    } finally {
      commit('setLoading', { type: 'update', loading: false })
    }
  },
  
  // 删除用户
  async deleteUser({ commit }, userId) {
    commit('setLoading', { type: 'delete', loading: true })
    try {
      await this.dispatch('userManagement/api/deleteUser', userId)
      commit('removeUser', userId)
      return true
    } finally {
      commit('setLoading', { type: 'delete', loading: false })
    }
  },
  
  // 切换用户状态
  async toggleUserStatus({ commit }, { user_id, status }) {
    try {
      const user = await this.dispatch('userManagement/api/toggleUserStatus', { user_id, status })
      commit('updateUser', user)
      return user
    } catch (error) {
      throw error
    }
  },
  
  // 批量删除用户
  async batchDeleteUsers({ commit }, userIds) {
    commit('setLoading', { type: 'delete', loading: true })
    try {
      await this.dispatch('userManagement/api/batchDeleteUsers', userIds)
      userIds.forEach(id => {
        commit('removeUser', id)
      })
      return true
    } finally {
      commit('setLoading', { type: 'delete', loading: false })
    }
  },
  
  // 重置用户密码
  async resetUserPassword({ commit }, userId) {
    try {
      return await this.dispatch('userManagement/api/resetUserPassword', userId)
    } catch (error) {
      throw error
    }
  },
  
  // 导出用户列表
  async exportUsers({ commit }, params = {}) {
    try {
      return await this.dispatch('userManagement/api/exportUsers', params)
    } catch (error) {
      throw error
    }
  },
  
  // 导入用户
  async importUsers({ commit }, formData) {
    try {
      return await this.dispatch('userManagement/api/importUsers', formData)
    } catch (error) {
      throw error
    }
  },
  
  // 获取用户统计信息
  async getUserStatistics({ commit }) {
    try {
      return await this.dispatch('userManagement/api/getUserStatistics')
    } catch (error) {
      throw error
    }
  },
  
  // 验证邮箱是否可用
  async validateEmail({ commit }, email) {
    try {
      return await this.dispatch('userManagement/api/validateEmail', email)
    } catch (error) {
      throw error
    }
  },
  
  // 搜索用户
  async searchUsers({ commit, dispatch }, { keyword, filters = {} }) {
    commit('updateSearchFilters', { keyword, ...filters })
    commit('updatePagination', { current: 1 })
    return await dispatch('getUserList')
  },
  
  // 重置搜索
  async resetSearch({ commit, dispatch }) {
    commit('resetSearchFilters')
    commit('updatePagination', { current: 1 })
    return await dispatch('getUserList')
  },
  
  // 刷新用户列表
  async refreshUserList({ dispatch }) {
    return await dispatch('getUserList')
  }
}

export default {
  namespaced: true,
  state,
  getters,
  mutations,
  actions
}