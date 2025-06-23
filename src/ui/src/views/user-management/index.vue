<template>
  <div class="user-management">
    <div class="header">
      <h2>用户管理</h2>
      <bk-button
        theme="primary"
        icon="plus"
        @click="handleCreateUser">
        新增用户
      </bk-button>
    </div>
    
    <div class="content">
      <bk-tab
        :active.sync="activeTab"
        type="unborder-card">
        <bk-tab-panel name="users" label="用户列表">
          <user-list 
            ref="userList"
            @edit="handleEditUser"
            @delete="handleDeleteUser" />
        </bk-tab-panel>
        <bk-tab-panel name="roles" label="角色管理">
          <role-management />
        </bk-tab-panel>
        <bk-tab-panel name="permissions" label="权限设置">
          <permission-settings />
        </bk-tab-panel>
      </bk-tab>
    </div>

    <!-- 用户编辑弹窗 -->
    <bk-dialog
      v-model="showUserDialog"
      :title="isEditMode ? '编辑用户' : '新增用户'"
      :width="600"
      :mask-close="false"
      @confirm="handleConfirmUser"
      @cancel="handleCancelUser">
      <user-form
        ref="userForm"
        :user-data="currentUser"
        :is-edit="isEditMode" />
    </bk-dialog>
  </div>
</template>

<script>
import UserList from './user-list.vue'
import UserForm from './user-form.vue'
import RoleManagement from './role-management.vue'
import PermissionSettings from './permission-settings.vue'
import { mapActions } from 'vuex'

export default {
  name: 'UserManagement',
  components: {
    UserList,
    UserForm,
    RoleManagement,
    PermissionSettings
  },
  data() {
    return {
      activeTab: 'users',
      showUserDialog: false,
      isEditMode: false,
      currentUser: null
    }
  },
  methods: {
    ...mapActions('userManagement', [
      'createUser',
      'updateUser',
      'deleteUser'
    ]),

    handleCreateUser() {
      this.isEditMode = false
      this.currentUser = null
      this.showUserDialog = true
    },

    handleEditUser(user) {
      this.isEditMode = true
      this.currentUser = { ...user }
      this.showUserDialog = true
    },

    async handleDeleteUser(user) {
      try {
        await this.$bkInfo({
          title: '确认删除',
          subTitle: `确定要删除用户 ${user.name} 吗？`,
          confirmFn: async () => {
            await this.deleteUser(user.id)
            this.$bkMessage({
              theme: 'success',
              message: '删除成功'
            })
            this.$refs.userList.fetchUsers()
          }
        })
      } catch (error) {
        this.$bkMessage({
          theme: 'error',
          message: error.message || '删除失败'
        })
      }
    },

    async handleConfirmUser() {
      try {
        const formData = await this.$refs.userForm.validate()
        if (this.isEditMode) {
          await this.updateUser({ id: this.currentUser.id, ...formData })
          this.$bkMessage({
            theme: 'success',
            message: '更新成功'
          })
        } else {
          await this.createUser(formData)
          this.$bkMessage({
            theme: 'success',
            message: '创建成功'
          })
        }
        this.showUserDialog = false
        this.$refs.userList.fetchUsers()
      } catch (error) {
        this.$bkMessage({
          theme: 'error',
          message: error.message || '操作失败'
        })
      }
    },

    handleCancelUser() {
      this.showUserDialog = false
      this.currentUser = null
    }
  }
}
</script>

<style lang="scss" scoped>
.user-management {
  padding: 20px;
  
  .header {
    display: flex;
    justify-content: space-between;
    align-items: center;
    margin-bottom: 20px;
    
    h2 {
      margin: 0;
      color: #313238;
      font-size: 16px;
      font-weight: 600;
    }
  }
  
  .content {
    background: #fff;
    border-radius: 2px;
    box-shadow: 0 1px 2px 0 rgba(0, 0, 0, 0.1);
  }
}
</style>