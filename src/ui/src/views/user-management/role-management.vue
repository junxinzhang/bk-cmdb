<template>
  <div class="role-management">
    <div class="role-header">
      <div class="header-left">
        <h3>角色管理</h3>
        <p class="role-desc">系统预设两种角色，每种角色具有不同的权限范围</p>
      </div>
      <div class="header-actions">
        <bk-button
          theme="primary"
          icon="plus"
          @click="createRole">
          创建角色
        </bk-button>
      </div>
    </div>

    <div class="role-list">
      <div
        v-for="role in roles"
        v-if="role"
        :key="role.key || role.role_name || Math.random()"
        class="role-card"
        :class="{ 'is-active': selectedRole === role.key }"
        @click="selectRole(role.key)">

        <div class="role-info">
          <div class="role-title">
            <bk-tag :theme="role.theme || 'info'" class="role-tag">{{ role.role_name || role.name || '未命名角色' }}</bk-tag>
          </div>
          <div class="role-description">{{ role.description }}</div>
          <div class="role-permissions">
            <span class="permission-label">权限范围：</span>
            <bk-tag
              v-for="permission in (role.permissions || [])"
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
            @click.stop="editRole(role)">
            编辑角色
          </bk-button>
          <bk-button
            text
            theme="primary"
            @click.stop="editRolePermissions(role)">
            编辑权限
          </bk-button>
          <bk-button
            v-if="!role.is_system"
            text
            theme="danger"
            @click.stop="deleteRole(role)">
            删除
          </bk-button>
        </div>
      </div>
    </div>

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
        <div class="permission-grid">
          <div
            v-for="permission in availablePermissions"
            :key="permission.key"
            class="permission-item">
            <bk-checkbox
              :value="editingPermissions.includes(permission.key)"
              :disabled="isPermissionLocked(permission.key)"
              @change="handlePermissionChange(permission.key, $event)">
              {{ permission.label }}
            </bk-checkbox>
            <div class="permission-desc">{{ permission.description }}</div>
          </div>
        </div>
      </div>
    </bk-dialog>

    <!-- 角色创建/编辑弹窗 -->
    <role-form
      :show.sync="showRoleForm"
      :role-data="currentEditRole"
      :is-edit="isEditMode"
      @success="handleRoleFormSuccess"
      @cancel="handleRoleFormCancel">
    </role-form>
  </div>
</template>

<script>
  import { mapState, mapActions } from 'vuex'
  import RoleForm from './role-form.vue'

  export default {
    name: 'RoleManagement',
    components: {
      RoleForm
    },
    data() {
      return {
        selectedRole: 'admin',
        showPermissionDialog: false,
        showRoleForm: false,
        currentRoleName: '',
        currentRoleKey: '',
        currentEditRole: null,
        isEditMode: false,
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
        'updateRolePermissions',
        'createRole',
        'updateRole',
        'deleteRole'
      ]),

      async fetchRoles() {
        try {
          console.log('Fetching roles...')
          const roles = await this.$store.dispatch('rolePermission/getRoles')
          console.log('Roles fetched:', roles)
          console.log('Store roles state:', this.$store.state.rolePermission.roles)
        } catch (error) {
          console.error('Failed to fetch roles:', error)
          this.$bkMessage({
            theme: 'error',
            message: error.message || '获取角色信息失败'
          })
        }
      },

      selectRole(roleKey) {
        this.selectedRole = roleKey
      },

      createRole() {
        this.currentEditRole = null
        this.isEditMode = false
        this.showRoleForm = true
      },

      editRole(role) {
        this.currentEditRole = role
        this.isEditMode = true
        this.showRoleForm = true
      },

      editRolePermissions(role) {
        this.currentRoleKey = role.key || role.role_name
        this.currentRoleName = role.role_name || role.name
        this.editingPermissions = [...(role?.permissions || [])]
        this.showPermissionDialog = true
      },

      async deleteRole(role) {
        console.log('deleteRole 方法被调用，角色信息:', role)
        this.$bkInfo({
          title: '确认删除',
          subTitle: `确定要删除角色"${role.role_name || role.name}"吗？删除后无法恢复。`,
          confirmText: '删除',
          cancelText: '取消',
          confirmFn: async () => {
            console.log('用户确认删除，开始执行删除操作')
            try {
              const roleKey = role.key || role.role_name
              console.log('即将删除的角色key:', roleKey)
              // 调用Vuex action删除角色
              const result = await this.$store.dispatch('rolePermission/deleteRole', roleKey)
              console.log('删除角色结果:', result)
              this.$bkMessage({
                theme: 'success',
                message: '角色删除成功'
              })
              await this.fetchRoles()
            } catch (error) {
              console.error('删除角色失败:', error)
              this.$bkMessage({
                theme: 'error',
                message: error.message || '角色删除失败'
              })
            }
          }
        })
      },

      async saveRolePermissions() {
        try {
          console.log('Saving role permissions:', {
            roleKey: this.currentRoleKey,
            permissions: this.editingPermissions
          })
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
          console.error('Permission update error:', error)
          this.$bkMessage({
            theme: 'error',
            message: error.message || '权限更新失败'
          })
        }
      },

      getRoleNameByKey(key) {
        const role = this.roles.find(r => (r.key || r.role_name) === key)
        return role?.role_name || role?.name || key
      },

      getPermissionLabel(permissionKey) {
        const permission = this.availablePermissions.find(p => p.key === permissionKey)
        return permission?.label || permissionKey
      },

      isPermissionLocked(permissionKey) {
        return this.currentRoleKey === 'admin' && permissionKey === 'admin'
      },

      handlePermissionChange(permissionKey, isChecked) {
        if (isChecked) {
          // 添加权限
          if (!this.editingPermissions.includes(permissionKey)) {
            this.editingPermissions.push(permissionKey)
          }
        } else {
          // 移除权限
          const index = this.editingPermissions.indexOf(permissionKey)
          if (index > -1) {
            this.editingPermissions.splice(index, 1)
          }
        }
      },

      handleRoleFormSuccess() {
        this.fetchRoles()
      },

      handleRoleFormCancel() {
        this.currentEditRole = null
        this.isEditMode = false
      }
    }
  }
</script>

<style lang="scss" scoped>
.role-management {
  padding: 20px;

  .role-header {
    display: flex;
    justify-content: space-between;
    align-items: flex-start;
    margin-bottom: 24px;

    .header-left {
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

    .header-actions {
      flex-shrink: 0;
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

    .permission-grid {
      display: grid;
      grid-template-columns: repeat(2, 1fr);
      gap: 16px;

      .permission-item {
        .permission-desc {
          font-size: 12px;
          color: #979ba5;
          margin-top: 4px;
          margin-left: 24px;
          line-height: 1.4;
        }
      }
    }
  }

}
</style>
