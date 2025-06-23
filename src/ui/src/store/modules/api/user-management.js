import $http from '@/api'

const state = {
  users: [],
  userDetail: null
}

const getters = {
  users: state => state.users,
  userDetail: state => state.userDetail
}

const mutations = {
  setUsers(state, users) {
    state.users = users
  },
  setUserDetail(state, user) {
    state.userDetail = user
  },
  addUser(state, user) {
    state.users.unshift(user)
  },
  updateUser(state, updatedUser) {
    const index = state.users.findIndex(user => user.id === updatedUser.id)
    if (index !== -1) {
      state.users.splice(index, 1, updatedUser)
    }
  },
  removeUser(state, userId) {
    const index = state.users.findIndex(user => user.id === userId)
    if (index !== -1) {
      state.users.splice(index, 1)
    }
  }
}

const actions = {
  // 获取用户列表
  async getUserList({ commit }, params = {}) {
    try {
      const response = await $http.get('/user/list', {
        params: {
          page: params.page || 1,
          limit: params.limit || 20,
          search: params.search || '',
          role: params.role || '',
          status: params.status || ''
        }
      })
      commit('setUsers', response.data || [])
      return response
    } catch (error) {
      console.error('获取用户列表失败:', error)
      throw error
    }
  },

  // 获取用户详情
  async getUserDetail({ commit }, userId) {
    try {
      const response = await $http.get(`/user/${userId}`)
      commit('setUserDetail', response.data)
      return response.data
    } catch (error) {
      console.error('获取用户详情失败:', error)
      throw error
    }
  },

  // 创建用户
  async createUser({ commit }, userData) {
    try {
      const response = await $http.post('/user/create', {
        email: userData.email,
        name: userData.name,
        role: userData.role,
        permissions: userData.permissions || [],
        status: userData.status || 'active'
      })
      commit('addUser', response.data)
      return response.data
    } catch (error) {
      console.error('创建用户失败:', error)
      throw error
    }
  },

  // 更新用户
  async updateUser({ commit }, { id, ...userData }) {
    try {
      const response = await $http.put(`/user/${id}`, {
        name: userData.name,
        role: userData.role,
        permissions: userData.permissions || [],
        status: userData.status
      })
      commit('updateUser', response.data)
      return response.data
    } catch (error) {
      console.error('更新用户失败:', error)
      throw error
    }
  },

  // 删除用户
  async deleteUser({ commit }, userId) {
    try {
      await $http.delete(`/user/${userId}`)
      commit('removeUser', userId)
      return true
    } catch (error) {
      console.error('删除用户失败:', error)
      throw error
    }
  },

  // 切换用户状态
  async toggleUserStatus({ commit }, { id, status }) {
    try {
      const response = await $http.patch(`/user/${id}/status`, {
        status
      })
      commit('updateUser', response.data)
      return response.data
    } catch (error) {
      console.error('切换用户状态失败:', error)
      throw error
    }
  },

  // 批量删除用户
  async batchDeleteUsers({ commit }, userIds) {
    try {
      await $http.delete('/user/batch', {
        data: { user_ids: userIds }
      })
      userIds.forEach(id => {
        commit('removeUser', id)
      })
      return true
    } catch (error) {
      console.error('批量删除用户失败:', error)
      throw error
    }
  },

  // 重置用户密码
  async resetUserPassword({ commit }, userId) {
    try {
      const response = await $http.post(`/user/${userId}/reset-password`)
      return response.data
    } catch (error) {
      console.error('重置用户密码失败:', error)
      throw error
    }
  },

  // 导出用户列表
  async exportUsers({ commit }, params = {}) {
    try {
      const response = await $http.get('/user/export', {
        params: {
          role: params.role || '',
          status: params.status || '',
          format: params.format || 'excel'
        },
        responseType: 'blob'
      })
      return response
    } catch (error) {
      console.error('导出用户列表失败:', error)
      throw error
    }
  },

  // 导入用户
  async importUsers({ commit }, formData) {
    try {
      const response = await $http.post('/user/import', formData, {
        headers: {
          'Content-Type': 'multipart/form-data'
        }
      })
      return response.data
    } catch (error) {
      console.error('导入用户失败:', error)
      throw error
    }
  },

  // 获取用户统计信息
  async getUserStatistics({ commit }) {
    try {
      const response = await $http.get('/user/statistics')
      return response.data
    } catch (error) {
      console.error('获取用户统计信息失败:', error)
      throw error
    }
  },

  // 验证邮箱是否可用
  async validateEmail({ commit }, email) {
    try {
      const response = await $http.post('/user/validate-email', {
        email
      })
      return response.data
    } catch (error) {
      console.error('验证邮箱失败:', error)
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