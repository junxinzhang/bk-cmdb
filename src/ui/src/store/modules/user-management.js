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
    if (!updatedUser) {
      console.error('updateUser mutation called with undefined user')
      return
    }
    const userId = updatedUser._id || updatedUser.id || updatedUser.user_id
    if (!userId) {
      console.error('updateUser mutation: user has no _id, id or user_id', updatedUser)
      return
    }
    
    const index = state.userList.findIndex(user => (user._id || user.id || user.user_id) === userId)
    if (index !== -1) {
      // 如果 updatedUser 只包含部分数据（比如只有状态），则合并到现有用户数据中
      if (Object.keys(updatedUser).length <= 3 && updatedUser.status) {
        // 部分更新，只更新特定字段
        state.userList[index] = {
          ...state.userList[index],
          ...updatedUser,
          updated_at: new Date().toISOString()
        }
      } else {
        // 完整更新
        state.userList.splice(index, 1, updatedUser)
      }
    }
    
    if (state.currentUser && (state.currentUser._id || state.currentUser.id || state.currentUser.user_id) === userId) {
      if (Object.keys(updatedUser).length <= 3 && updatedUser.status) {
        // 部分更新当前用户
        state.currentUser = {
          ...state.currentUser,
          ...updatedUser,
          updated_at: new Date().toISOString()
        }
      } else {
        state.currentUser = updatedUser
      }
    }
  },
  
  removeUser(state, userId) {
    const index = state.userList.findIndex(user => (user._id || user.id || user.user_id) === userId)
    if (index !== -1) {
      state.userList.splice(index, 1)
      state.pagination.count -= 1
    }
    if (state.currentUser && (state.currentUser._id || state.currentUser.id || state.currentUser.user_id) === userId) {
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
      
      // response.data 是API响应的数据部分
      const apiData = response.data || response
      
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
      const response = await $http.post('usermgmt/create', userData)
      // Handle CMDB API response format
      const user = response.data?.data || response.data || response
      commit('addUser', user)
      return user
    } finally {
      commit('setLoading', { type: 'create', loading: false })
    }
  },
  
  // 更新用户
  async updateUser({ commit, dispatch }, { id, ...userData }) {
    commit('setLoading', { type: 'update', loading: true })
    try {
      const response = await $http.put(`usermgmt/${id}`, userData)
      console.log('Update user response:', response)
      
      // Handle CMDB API response format - try multiple possible response structures
      let user = null
      if (response.data && typeof response.data === 'object') {
        // Standard CMDB format: { result: true, code: 0, data: {...} }
        if (response.data.data) {
          user = response.data.data
        }
        // Direct data format: { _id: "...", name: "...", ... }
        else if (response.data._id || response.data.user_id || response.data.email) {
          user = response.data
        }
      }
      // Fallback to response itself
      if (!user && (response._id || response.user_id)) {
        user = response
      }
      
      console.log('Extracted user data:', user)
      
      if (user && (user._id || user.user_id || user.email)) {
        commit('updateUser', user)
        return user
      } else {
        // If no user data returned, try to refresh the user list to get updated data
        console.warn('No user data in update response, refreshing user list')
        await dispatch('getUserList')
        return { user_id: id, ...userData }
      }
    } catch (error) {
      console.error('Update user error:', error)
      throw error
    } finally {
      commit('setLoading', { type: 'update', loading: false })
    }
  },

  // 编辑用户
  async editUser({ commit, dispatch }, { id, ...userData }) {
    commit('setLoading', { type: 'update', loading: true })
    try {
      const response = await $http.put(`usermgmt/${id}`, userData)
      
      // Handle CMDB API response format - try multiple possible response structures
      let user = null
      if (response.data && typeof response.data === 'object') {
        // Standard CMDB format: { result: true, code: 0, data: {...} }
        if (response.data.data) {
          user = response.data.data
        }
        // Direct data format: { _id: "...", name: "...", ... }
        else if (response.data._id || response.data.user_id || response.data.email) {
          user = response.data
        }
      }
      // Fallback to response itself
      if (!user && (response._id || response.user_id)) {
        user = response
      }
      
      if (user && (user._id || user.user_id || user.email)) {
        commit('updateUser', user)
        return user
      } else {
        // If no user data returned, try to refresh the user list to get updated data
        console.warn('No user data in edit response, refreshing user list')
        await dispatch('getUserList')
        return { user_id: id, ...userData }
      }
    } finally {
      commit('setLoading', { type: 'update', loading: false })
    }
  },
  
  // 删除用户
  async deleteUser({ commit }, userId) {
    commit('setLoading', { type: 'delete', loading: true })
    try {
      await $http.delete(`usermgmt/${userId}`)
      commit('removeUser', userId)
      return true
    } finally {
      commit('setLoading', { type: 'delete', loading: false })
    }
  },

  // 禁用用户
  async disableUser({ commit }, userId) {
    commit('setLoading', { type: 'update', loading: true })
    try {
      const response = await $http.put(`usermgmt/${userId}/disable`)
      console.log('Disable user response:', response)
      
      // Handle CMDB API response format
      let user = null
      if (response.data && typeof response.data === 'object') {
        if (response.data.data) {
          user = response.data.data
        } else if (response.data._id || response.data.user_id || response.data.email) {
          user = response.data
        }
      }
      if (!user && (response._id || response.user_id)) {
        user = response
      }
      
      if (user && (user._id || user.user_id || user.email)) {
        commit('updateUser', user)
        return user
      } else {
        console.warn('No user data in disable response, the user may have been disabled successfully')
        // 禁用成功但没有返回用户数据，手动更新状态
        const updatedUser = { 
          _id: userId, 
          user_id: userId, 
          status: 'inactive' 
        }
        commit('updateUser', updatedUser)
        return updatedUser
      }
    } catch (error) {
      console.error('Disable user error:', error)
      throw error
    } finally {
      commit('setLoading', { type: 'update', loading: false })
    }
  },

  // 启用用户
  async enableUser({ commit }, userId) {
    commit('setLoading', { type: 'update', loading: true })
    try {
      const response = await $http.put(`usermgmt/${userId}/enable`)
      console.log('Enable user response:', response)
      
      // Handle CMDB API response format
      let user = null
      if (response.data && typeof response.data === 'object') {
        if (response.data.data) {
          user = response.data.data
        } else if (response.data._id || response.data.user_id || response.data.email) {
          user = response.data
        }
      }
      if (!user && (response._id || response.user_id)) {
        user = response
      }
      
      if (user && (user._id || user.user_id || user.email)) {
        commit('updateUser', user)
        return user
      } else {
        console.warn('No user data in enable response, the user may have been enabled successfully')
        // 启用成功但没有返回用户数据，手动更新状态
        const updatedUser = { 
          _id: userId, 
          user_id: userId, 
          status: 'active' 
        }
        commit('updateUser', updatedUser)
        return updatedUser
      }
    } catch (error) {
      console.error('Enable user error:', error)
      throw error
    } finally {
      commit('setLoading', { type: 'update', loading: false })
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
  async resetUserPassword({ }, userId) {
    try {
      return await this.dispatch('userManagement/api/resetUserPassword', userId)
    } catch (error) {
      throw error
    }
  },
  
  // 导出用户列表
  async exportUsers({ }, params = {}) {
    try {
      return await this.dispatch('userManagement/api/exportUsers', params)
    } catch (error) {
      throw error
    }
  },
  
  // 导入用户
  async importUsers({ }, formData) {
    try {
      return await this.dispatch('userManagement/api/importUsers', formData)
    } catch (error) {
      throw error
    }
  },
  
  // 获取用户统计信息
  async getUserStatistics({ }) {
    try {
      return await this.dispatch('userManagement/api/getUserStatistics')
    } catch (error) {
      throw error
    }
  },
  
  // 验证邮箱是否可用
  async validateEmail({ }, email) {
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