<template>
  <div class="user-management">
    <div class="header">
      <h2>{{ $t('用户管理') }}</h2>
      <bk-button
        theme="primary"
        icon="plus"
        @click="handleCreateUser">
        {{ $t('新增用户') }}
      </bk-button>
    </div>

    <div class="content">
      <bk-tab
        :active.sync="activeTab"
        type="unborder-card">
        <bk-tab-panel name="users" :label="$t('用户列表')">
          <user-list
            ref="userList"
            @edit="handleEditUser"
            @delete="handleDeleteUser" />
        </bk-tab-panel>
        <bk-tab-panel name="roles" :label="$t('角色管理')">
          <role-management />
        </bk-tab-panel>
      </bk-tab>
    </div>

    <!-- 用户编辑弹窗 -->
    <bk-dialog
      v-model="showUserDialog"
      :title="isEditMode ? $t('编辑用户') : $t('新增用户')"
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
  import { mapActions } from 'vuex'

  export default {
    name: 'UserManagement',
    components: {
      UserList,
      UserForm,
      RoleManagement
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
            title: this.$t('确认删除'),
            subTitle: this.$t('确认删除该用户？').replace('该用户', user.name),
            confirmFn: async () => {
              await this.deleteUser(user._id || user.id || user.user_id)
              this.$bkMessage({
                theme: 'success',
                message: this.$t('删除成功')
              })
              this.$refs.userList.fetchUsers()
            }
          })
        } catch (error) {
          this.$bkMessage({
            theme: 'error',
            message: error.message || this.$t('操作失败')
          })
        }
      },

      async handleConfirmUser() {
        try {
          const formData = await this.$refs.userForm.validate()
          if (this.isEditMode) {
            await this.updateUser({ id: this.currentUser._id || this.currentUser.id || this.currentUser.user_id, ...formData })
            this.$bkMessage({
              theme: 'success',
              message: this.$t('保存成功')
            })
          } else {
            await this.createUser(formData)
            this.$bkMessage({
              theme: 'success',
              message: this.$t('保存成功')
            })
          }
          this.showUserDialog = false
          this.$refs.userList.fetchUsers()
        } catch (error) {
          this.$bkMessage({
            theme: 'error',
            message: error.message || this.$t('操作失败')
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
