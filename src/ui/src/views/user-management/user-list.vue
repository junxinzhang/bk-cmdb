<template>
  <div class="user-list">
    <div class="list-header">
      <div class="search-area">
        <bk-input
          v-model="searchKeyword"
          placeholder="请输入用户姓名或邮箱搜索"
          :clearable="true"
          :right-icon="'bk-icon icon-search'"
          style="width: 300px;"
          @enter="handleSearch"
          @clear="handleSearch">
        </bk-input>
        <bk-button @click="handleSearch" style="margin-left: 8px;">搜索</bk-button>
      </div>
      <div class="filter-area">
        <bk-select
          v-model="selectedRole"
          placeholder="选择角色"
          :clearable="true"
          style="width: 120px;"
          @change="handleSearch">
          <bk-option
            v-for="role in roleOptions"
            :key="role.value"
            :id="role.value"
            :name="role.label">
          </bk-option>
        </bk-select>
      </div>
    </div>

    <bk-table
      :data="userList"
      :pagination="pagination"
      v-bkloading="{ isLoading: loading }"
      @page-change="handlePageChange"
      @page-limit-change="handlePageLimitChange">
      
      <bk-table-column prop="email" label="邮箱" min-width="200">
        <template slot-scope="props">
          <span>{{ props.row.email }}</span>
        </template>
      </bk-table-column>
      
      <bk-table-column prop="name" label="姓名" min-width="120">
        <template slot-scope="props">
          <span>{{ props.row.name }}</span>
        </template>
      </bk-table-column>
      
      <bk-table-column prop="role" label="角色" min-width="100">
        <template slot-scope="props">
          <bk-tag :theme="getRoleTheme(props.row.role)">
            {{ getRoleLabel(props.row.role) }}
          </bk-tag>
        </template>
      </bk-table-column>
      
      <bk-table-column prop="status" label="状态" min-width="80">
        <template slot-scope="props">
          <bk-tag :theme="props.row.status === 'active' ? 'success' : 'danger'">
            {{ props.row.status === 'active' ? '启用' : '禁用' }}
          </bk-tag>
        </template>
      </bk-table-column>
      
      <bk-table-column prop="created_at" label="创建时间" min-width="160">
        <template slot-scope="props">
          <span>{{ formatDisplayTime(props.row.created_at) }}</span>
        </template>
      </bk-table-column>
      
      <bk-table-column prop="last_login" label="最后登录" min-width="160">
        <template slot-scope="props">
          <span>{{ props.row.last_login ? formatDisplayTime(props.row.last_login) : '从未登录' }}</span>
        </template>
      </bk-table-column>
      
      <bk-table-column label="操作" min-width="160" fixed="right">
        <template slot-scope="props">
          <bk-button
            text
            theme="primary"
            @click="handleEdit(props.row)">
            编辑
          </bk-button>
          <bk-button
            text
            :theme="props.row.status === 'active' ? 'warning' : 'success'"
            @click="handleToggleStatus(props.row)">
            {{ props.row.status === 'active' ? '禁用' : '启用' }}
          </bk-button>
          <bk-button
            text
            theme="danger"
            @click="handleDelete(props.row)">
            删除
          </bk-button>
        </template>
      </bk-table-column>
    </bk-table>
  </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'
import { formatTime } from '@/utils/tools'

export default {
  name: 'UserList',
  data() {
    return {
      searchKeyword: '',
      selectedRole: '',
      loading: false,
      roleOptions: [
        { value: 'admin', label: '管理员' },
        { value: 'operator', label: '操作员' }
      ]
    }
  },
  computed: {
    ...mapState('userManagement', [
      'userList',
      'pagination'
    ])
  },
  created() {
    this.fetchUsers()
  },
  methods: {
    ...mapActions('userManagement', [
      'getUserList',
      'toggleUserStatus'
    ]),

    async fetchUsers() {
      this.loading = true
      try {
        await this.getUserList({
          page: this.pagination.current,
          limit: this.pagination.limit,
          search: this.searchKeyword,
          role: this.selectedRole
        })
      } catch (error) {
        this.$bkMessage({
          theme: 'error',
          message: error.message || '获取用户列表失败'
        })
      } finally {
        this.loading = false
      }
    },

    handleSearch() {
      this.$store.commit('userManagement/updatePagination', {
        current: 1
      })
      this.fetchUsers()
    },

    handlePageChange(page) {
      this.$store.commit('userManagement/updatePagination', {
        current: page
      })
      this.fetchUsers()
    },

    handlePageLimitChange(limit) {
      this.$store.commit('userManagement/updatePagination', {
        current: 1,
        limit
      })
      this.fetchUsers()
    },

    handleEdit(user) {
      this.$emit('edit', user)
    },

    handleDelete(user) {
      this.$emit('delete', user)
    },

    async handleToggleStatus(user) {
      try {
        await this.toggleUserStatus({
          id: user.id,
          status: user.status === 'active' ? 'inactive' : 'active'
        })
        this.$bkMessage({
          theme: 'success',
          message: user.status === 'active' ? '用户已禁用' : '用户已启用'
        })
        this.fetchUsers()
      } catch (error) {
        this.$bkMessage({
          theme: 'error',
          message: error.message || '操作失败'
        })
      }
    },

    getRoleTheme(role) {
      return role === 'admin' ? 'danger' : 'info'
    },

    getRoleLabel(role) {
      const roleMap = {
        admin: '管理员',
        operator: '操作员'
      }
      return roleMap[role] || role
    },

    formatDisplayTime(timestamp) {
      return formatTime(timestamp, 'YYYY-MM-DD HH:mm:ss')
    }
  }
}
</script>

<style lang="scss" scoped>
.user-list {
  padding: 20px;
  
  .list-header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 16px;
    
    .search-area {
      display: flex;
      align-items: center;
    }
    
    .filter-area {
      display: flex;
      align-items: center;
      gap: 8px;
    }
  }
}
</style>