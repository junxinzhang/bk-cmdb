<template>
  <div class="user-list">
    <div class="list-header">
      <div class="search-area">
        <bk-input
          v-model="searchKeyword"
          :placeholder="$t('请输入用户姓名或邮箱搜索')"
          :clearable="true"
          :right-icon="'bk-icon icon-search'"
          style="width: 300px;"
          @enter="handleSearch"
          @clear="handleSearch">
        </bk-input>
        <bk-button @click="handleSearch" style="margin-left: 8px;">{{ $t('搜索') }}</bk-button>
      </div>
      <div class="filter-area">
        <bk-select
          v-model="selectedRole"
          :placeholder="$t('请选择角色')"
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

      <bk-table-column prop="email" :label="$t('邮箱')" min-width="200">
        <template slot-scope="props">
          <span>{{ props.row.email }}</span>
        </template>
      </bk-table-column>

      <bk-table-column prop="name" :label="$t('姓名')" min-width="120">
        <template slot-scope="props">
          <span>{{ props.row.name }}</span>
        </template>
      </bk-table-column>

      <bk-table-column prop="role" :label="$t('角色')" min-width="100">
        <template slot-scope="props">
          <bk-tag :theme="getRoleTheme(props.row.role)">
            {{ getRoleLabel(props.row.role) }}
          </bk-tag>
        </template>
      </bk-table-column>

      <bk-table-column prop="status" :label="$t('状态')" min-width="80">
        <template slot-scope="props">
          <bk-tag :theme="getStatusTheme(props.row.status)">
            {{ getStatusLabel(props.row.status) }}
          </bk-tag>
        </template>
      </bk-table-column>

      <bk-table-column prop="created_at" :label="$t('创建时间')" min-width="160">
        <template slot-scope="props">
          <span>{{ formatDisplayTime(props.row.created_at) }}</span>
        </template>
      </bk-table-column>

      <bk-table-column prop="last_login" :label="$t('最后登录')" min-width="160">
        <template slot-scope="props">
          <span>{{ props.row.last_login ? formatDisplayTime(props.row.last_login) : $t('从未登录') }}</span>
        </template>
      </bk-table-column>

      <bk-table-column :label="$t('操作')" min-width="160" fixed="right">
        <template slot-scope="props">
          <bk-button
            text
            theme="primary"
            @click="handleEdit(props.row)">
            {{ $t('编辑用户') }}
          </bk-button>
          <bk-button
            text
            :theme="getToggleButtonTheme(props.row.status)"
            :disabled="props.row.status === 'locked'"
            @click="handleToggleStatus(props.row)">
            {{ getToggleButtonLabel(props.row.status) }}
          </bk-button>
          <bk-button
            text
            theme="danger"
            @click="handleDelete(props.row)">
            {{ $t('删除用户') }}
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
        loading: false
      }
    },
    computed: {
      ...mapState('userManagement', [
        'userList',
        'pagination'
      ]),
      ...mapState('rolePermission', [
        'roles'
      ]),
      roleOptions() {
        return this.roles.map(role => ({
          value: role.key || role.role_name,
          label: role.name || role.role_name
        }))
      }
    },
    created() {
      this.fetchUsers()
      this.fetchRoles()
    },
    methods: {
      ...mapActions('userManagement', [
        'getUserList',
        'toggleUserStatus',
        'disableUser',
        'enableUser'
      ]),
      ...mapActions('rolePermission', [
        'getRoles'
      ]),

      async fetchUsers() {
        this.loading = true
        try {
          // 确保store中的searchFilters是最新的
          this.$store.commit('userManagement/updateSearchFilters', {
            keyword: this.searchKeyword,
            role: this.selectedRole
          })
          
          await this.getUserList({
            page: this.pagination.current,
            limit: this.pagination.limit,
            search: this.searchKeyword,
            role: this.selectedRole
          })
        } catch (error) {
          this.$bkMessage({
            theme: 'error',
            message: error.message || this.$t('获取用户列表失败')
          })
        } finally {
          this.loading = false
        }
      },

      async fetchRoles() {
        try {
          await this.getRoles()
        } catch (error) {
          console.error(this.$t('获取角色列表失败') + ':', error)
        }
      },

      handleSearch() {
        this.$store.commit('userManagement/updateSearchFilters', {
          keyword: this.searchKeyword,
          role: this.selectedRole
        })
        this.$store.commit('userManagement/updatePagination', {
          current: 1
        })
        this.fetchUsers()
      },

      handleRoleChange() {
        this.handleSearch()
      },

      handleClear() {
        this.searchKeyword = ''
        this.selectedRole = ''
        this.handleSearch()
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
        // 检查用户是否被锁定
        if (user.status === 'locked') {
          this.$bkMessage({
            theme: 'warning',
            message: this.$t('无法操作已锁定的用户')
          })
          return
        }

        try {
          const userId = user._id || user.id || user.user_id
          const isActive = user.status === 'active'

          if (isActive) {
            // 禁用用户
            await this.disableUser(userId)
            this.$bkMessage({
              theme: 'success',
              message: this.$t('用户已禁用')
            })
          } else {
            // 启用用户
            await this.enableUser(userId)
            this.$bkMessage({
              theme: 'success',
              message: this.$t('用户已启用')
            })
          }

          // 刷新用户列表以显示最新状态
          this.fetchUsers()
        } catch (error) {
          console.error('Toggle user status error:', error)
          this.$bkMessage({
            theme: 'error',
            message: error.message || this.$t('操作失败')
          })
        }
      },

      getRoleTheme(role) {
        return role === 'admin' ? 'danger' : 'info'
      },

      getRoleLabel(role) {
        const roleMap = {
          admin: this.$t('管理员'),
          operator: this.$t('操作员')
        }
        return roleMap[role] || role
      },

      getStatusTheme(status) {
        const statusThemeMap = {
          active: 'success',
          inactive: 'danger',
          locked: 'warning'
        }
        return statusThemeMap[status] || 'danger'
      },

      getStatusLabel(status) {
        const statusLabelMap = {
          active: this.$t('启用'),
          inactive: this.$t('禁用'),
          locked: this.$t('已锁定')
        }
        return statusLabelMap[status] || this.$t('未知')
      },

      getToggleButtonTheme(status) {
        if (status === 'active') return 'warning'
        if (status === 'inactive') return 'success'
        if (status === 'locked') return 'default'
        return 'default'
      },

      getToggleButtonLabel(status) {
        if (status === 'active') return this.$t('禁用')
        if (status === 'inactive') return this.$t('启用')
        if (status === 'locked') return this.$t('已锁定')
        return this.$t('操作')
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
