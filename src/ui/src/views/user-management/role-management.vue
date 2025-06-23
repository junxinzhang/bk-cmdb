<template>
  <div class="role-management">
    <div class="role-header">
      <h3>角色管理</h3>
      <p class="role-desc">系统预设两种角色，每种角色具有不同的权限范围</p>
    </div>

    <div class="role-list">
      <div
        v-for="role in roles"
        :key="role.key"
        class="role-card"
        :class="{ 'is-active': selectedRole === role.key }"
        @click="selectRole(role.key)">
        
        <div class="role-info">
          <div class="role-title">
            <bk-tag :theme="role.theme" class="role-tag">{{ role.name }}</bk-tag>
            <span class="user-count">{{ role.userCount }} 个用户</span>
          </div>
          <div class="role-description">{{ role.description }}</div>
          <div class="role-permissions">
            <span class="permission-label">权限范围：</span>
            <bk-tag
              v-for="permission in role.permissions"
              :key="permission"
              class="permission-tag">
              {{ getPermissionLabel(permission) }}
            </bk-tag>
          </div>
        </div>
        
        <div class="role-actions">
          <bk-button
            text
            theme="primary"
            @click.stop="viewRoleUsers(role.key)">
            查看用户
          </bk-button>
          <bk-button
            text
            theme="primary"
            @click.stop="editRolePermissions(role.key)">
            编辑权限
          </bk-button>
        </div>
      </div>
    </div>

    <!-- 角色用户列表弹窗 -->
    <bk-dialog
      v-model="showUserDialog"
      :title="`${currentRoleName} - 用户列表`"
      :width="800">
      <bk-table
        :data="roleUsers"
        v-bkloading="{ isLoading: loadingUsers }">
        <bk-table-column prop="email" label="邮箱" min-width="200"></bk-table-column>
        <bk-table-column prop="name" label="姓名" min-width="120"></bk-table-column>
        <bk-table-column prop="status" label="状态" min-width="80">
          <template slot-scope="props">
            <bk-tag :theme="props.row.status === 'active' ? 'success' : 'danger'">
              {{ props.row.status === 'active' ? '启用' : '禁用' }}
            </bk-tag>
          </template>
        </bk-table-column>
        <bk-table-column prop="last_login" label="最后登录" min-width="160">
          <template slot-scope="props">
            <span>{{ props.row.last_login ? formatDisplayTime(props.row.last_login) : '从未登录' }}</span>
          </template>
        </bk-table-column>
      </bk-table>
      <div slot="footer">
        <bk-button @click="showUserDialog = false">关闭</bk-button>
      </div>
    </bk-dialog>

    <!-- 权限编辑弹窗 -->
    <bk-dialog
      v-model="showPermissionDialog"
      :title="`编辑 ${currentRoleName} 权限`"
      :width="600"
      @confirm="saveRolePermissions">
      <div class="permission-editor">
        <div class="permission-tip">
          <bk-alert type="info">
            <template slot="title">
              管理员默认拥有所有权限，操作员可根据需要调整权限范围
            </template>
          </bk-alert>
        </div>
        <div class="permission-list">
          <div
            v-for="permission in availablePermissions"
            :key="permission.key"
            class="permission-item">
            <bk-checkbox
              v-model="editingPermissions"
              :value="permission.key"
              :disabled="isPermissionLocked(permission.key)">
              <span class="permission-name">{{ permission.label }}</span>
            </bk-checkbox>
            <div class="permission-desc">{{ permission.description }}</div>
          </div>
        </div>
      </div>
    </bk-dialog>
  </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'
import { formatTime } from '@/utils/tools'

export default {
  name: 'RoleManagement',
  data() {
    return {
      selectedRole: 'admin',
      showUserDialog: false,
      showPermissionDialog: false,
      loadingUsers: false,
      roleUsers: [],
      currentRoleName: '',
      currentRoleKey: '',
      editingPermissions: [],
      availablePermissions: [
        {
          key: 'home',
          label: '首页',
          description: '访问系统首页和概览信息'
        },
        {
          key: 'business',
          label: '业务',
          description: '管理和查看业务拓扑、服务实例等'
        },
        {
          key: 'resource',
          label: '资源',
          description: '管理主机、云区域等资源信息'
        },
        {
          key: 'model',
          label: '模型',
          description: '管理配置模型、字段和关联关系'
        },
        {
          key: 'operation',
          label: '运营分析',
          description: '查看运营数据和分析报告'
        },
        {
          key: 'admin',
          label: '平台管理',
          description: '系统配置、用户管理等管理功能'
        }
      ]
    }
  },
  computed: {
    ...mapState('rolePermission', [
      'roles'
    ])
  },
  created() {
    this.fetchRoles()
  },
  methods: {
    ...mapActions('rolePermission', [
      'getRoles',
      'getRoleUsers',
      'updateRolePermissions'
    ]),

    async fetchRoles() {
      try {
        await this.getRoles()
      } catch (error) {
        this.$bkMessage({
          theme: 'error',
          message: error.message || '获取角色信息失败'
        })
      }
    },

    selectRole(roleKey) {
      this.selectedRole = roleKey
    },

    async viewRoleUsers(roleKey) {
      this.loadingUsers = true
      this.currentRoleKey = roleKey
      this.currentRoleName = this.getRoleNameByKey(roleKey)
      this.showUserDialog = true
      
      try {
        this.roleUsers = await this.getRoleUsers(roleKey)
      } catch (error) {
        this.$bkMessage({
          theme: 'error',
          message: error.message || '获取用户列表失败'
        })
      } finally {
        this.loadingUsers = false
      }
    },

    editRolePermissions(roleKey) {
      this.currentRoleKey = roleKey
      this.currentRoleName = this.getRoleNameByKey(roleKey)
      const role = this.roles.find(r => r.key === roleKey)
      this.editingPermissions = [...(role?.permissions || [])]
      this.showPermissionDialog = true
    },

    async saveRolePermissions() {
      try {
        await this.updateRolePermissions({
          roleKey: this.currentRoleKey,
          permissions: this.editingPermissions
        })
        this.$bkMessage({
          theme: 'success',
          message: '权限更新成功'
        })
        this.showPermissionDialog = false
        this.fetchRoles()
      } catch (error) {
        this.$bkMessage({
          theme: 'error',
          message: error.message || '权限更新失败'
        })
      }
    },

    getRoleNameByKey(key) {
      const role = this.roles.find(r => r.key === key)
      return role?.name || key
    },

    getPermissionLabel(permissionKey) {
      const permission = this.availablePermissions.find(p => p.key === permissionKey)
      return permission?.label || permissionKey
    },

    isPermissionLocked(permissionKey) {
      return this.currentRoleKey === 'admin' && permissionKey === 'admin'
    },

    formatDisplayTime(timestamp) {
      return formatTime(timestamp, 'YYYY-MM-DD HH:mm:ss')
    }
  }
}
</script>

<style lang="scss" scoped>
.role-management {
  padding: 20px;
  
  .role-header {
    margin-bottom: 24px;
    
    h3 {
      margin: 0 0 8px 0;
      font-size: 16px;
      color: #313238;
    }
    
    .role-desc {
      margin: 0;
      color: #979ba5;
      font-size: 12px;
    }
  }
  
  .role-list {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(400px, 1fr));
    gap: 16px;
  }
  
  .role-card {
    background: #fff;
    border: 1px solid #dcdee5;
    border-radius: 2px;
    padding: 20px;
    cursor: pointer;
    transition: all 0.2s;
    
    &:hover {
      border-color: #3a84ff;
    }
    
    &.is-active {
      border-color: #3a84ff;
      box-shadow: 0 0 0 1px #3a84ff;
    }
    
    .role-info {
      .role-title {
        display: flex;
        align-items: center;
        justify-content: space-between;
        margin-bottom: 12px;
        
        .role-tag {
          font-size: 14px;
        }
        
        .user-count {
          color: #979ba5;
          font-size: 12px;
        }
      }
      
      .role-description {
        color: #63656e;
        font-size: 13px;
        line-height: 1.5;
        margin-bottom: 16px;
      }
      
      .role-permissions {
        .permission-label {
          color: #979ba5;
          font-size: 12px;
          margin-right: 8px;
        }
        
        .permission-tag {
          margin-right: 4px;
          margin-bottom: 4px;
        }
      }
    }
    
    .role-actions {
      margin-top: 20px;
      padding-top: 16px;
      border-top: 1px solid #f0f1f5;
      display: flex;
      gap: 16px;
    }
  }
  
  .permission-editor {
    .permission-tip {
      margin-bottom: 20px;
    }
    
    .permission-list {
      .permission-item {
        padding: 12px 0;
        border-bottom: 1px solid #f0f1f5;
        
        &:last-child {
          border-bottom: none;
        }
        
        .permission-name {
          font-weight: 500;
        }
        
        .permission-desc {
          color: #979ba5;
          font-size: 12px;
          margin-top: 4px;
          margin-left: 24px;
        }
      }
    }
  }
}
</style>