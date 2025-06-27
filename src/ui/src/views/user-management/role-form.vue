<template>
  <bk-dialog
    v-model="visible"
    :title="isEdit ? '编辑角色' : '创建角色'"
    :width="600"
    :loading="loading"
    @confirm="handleSubmit"
    @cancel="handleCancel">

    <bk-form
      ref="roleForm"
      :model="formData"
      :rules="formRules"
      :label-width="120">

      <bk-form-item label="角色名称" property="roleName" required>
        <bk-input
          v-model="formData.roleName"
          placeholder="请输入角色名称"
          :disabled="isEdit"
          maxlength="50">
        </bk-input>
      </bk-form-item>

      <bk-form-item label="角色描述" property="description">
        <bk-input
          v-model="formData.description"
          type="textarea"
          placeholder="请输入角色描述"
          :rows="3"
          maxlength="200">
        </bk-input>
      </bk-form-item>

      <bk-form-item label="权限配置" property="permissions" required>
        <div class="permission-selector">
          <div class="permission-tip">
            <bk-alert type="info">
              <template slot="title">
                请选择该角色拥有的权限，权限变更会立即生效
              </template>
            </bk-alert>
          </div>

          <div class="permission-categories">
            <bk-checkbox-group v-model="formData.permissions">
              <div
                v-for="category in permissionCategories"
                :key="category.key"
                class="permission-category">
                <h4 class="category-title">{{ getCategoryLabel(category.key) }}</h4>
                <div class="permission-items">
                  <bk-checkbox
                    v-for="permission in category.permissions"
                    :key="permission.key"
                    :value="permission.key"
                    class="permission-checkbox">
                    <span class="permission-label">{{ permission.label }}</span>
                    <div class="permission-desc">{{ permission.description }}</div>
                  </bk-checkbox>
                </div>
              </div>
            </bk-checkbox-group>
          </div>
        </div>
      </bk-form-item>
    </bk-form>
  </bk-dialog>
</template>

<script>
  import { mapGetters, mapActions } from 'vuex'

  export default {
    name: 'RoleForm',
    props: {
      show: {
        type: Boolean,
        default: false
      },
      roleData: {
        type: Object,
        default: () => ({})
      },
      isEdit: {
        type: Boolean,
        default: false
      }
    },
    data() {
      return {
        loading: false,
        formData: {
          roleName: '',
          description: '',
          permissions: []
        },
        formRules: {
          roleName: [
            { required: true, message: '请输入角色名称', trigger: 'blur' },
            { min: 2, max: 50, message: '角色名称长度在 2 到 50 个字符', trigger: 'blur' }
          ],
          description: [
            { max: 200, message: '描述长度不能超过 200 个字符', trigger: 'blur' }
          ],
          permissions: [
            { required: true, message: '请至少选择一个权限', trigger: 'change' }
          ]
        },
        categoryLabels: {
          basic: '基础功能',
          business: '业务管理',
          resource: '资源管理',
          config: '配置管理',
          operation: '运营分析',
          admin: '平台管理'
        }
      }
    },
    computed: {
      ...mapGetters('rolePermission', [
        'permissionCategories'
      ]),
      visible: {
        get() {
          return this.show
        },
        set(val) {
          if (!val) {
            this.$emit('update:show', false)
          }
        }
      }
    },
    watch: {
      show(val) {
        if (val) {
          this.initFormData()
        }
      }
    },
    methods: {
      ...mapActions('rolePermission', [
        'createRole',
        'updateRole'
      ]),

      initFormData() {
        if (this.isEdit && this.roleData) {
          this.formData = {
            roleName: this.roleData.role_name || this.roleData.name || '',
            description: this.roleData.description || '',
            permissions: [...(this.roleData.permissions || [])]
          }
        } else {
          this.formData = {
            roleName: '',
            description: '',
            permissions: []
          }
        }
      },

      async handleSubmit() {
        try {
          const valid = await this.$refs.roleForm.validate()
          if (!valid) return

          this.loading = true

          const roleData = {
            roleName: this.formData.roleName,
            description: this.formData.description,
            permissions: this.formData.permissions
          }

          if (this.isEdit) {
            console.log('Updating role with data:', {
              roleKey: this.roleData.id || this.roleData.key,
              ...roleData
            })
            await this.$store.dispatch('rolePermission/updateRole', {
              roleKey: this.roleData.id || this.roleData.key,
              ...roleData
            })
            this.$bkMessage({
              theme: 'success',
              message: '角色更新成功'
            })
          } else {
            console.log('Creating role with data:', roleData)
            const result = await this.$store.dispatch('rolePermission/createRole', roleData)
            console.log('Role created successfully:', result)
            this.$bkMessage({
              theme: 'success',
              message: '角色创建成功'
            })
          }

          this.$emit('success')
          this.handleCancel()
        } catch (error) {
          this.$bkMessage({
            theme: 'error',
            message: error.message || `角色${this.isEdit ? '更新' : '创建'}失败`
          })
        } finally {
          this.loading = false
        }
      },

      handleCancel() {
        this.$emit('update:show', false)
        this.$emit('cancel')
        this.$nextTick(() => {
          this.$refs.roleForm.clearError()
        })
      },

      getCategoryLabel(categoryKey) {
        return this.categoryLabels[categoryKey] || categoryKey
      }
    }
  }
</script>

<style lang="scss" scoped>
.permission-selector {
  .permission-tip {
    margin-bottom: 20px;
  }

  .permission-categories {
    .permission-category {
      margin-bottom: 24px;

      &:last-child {
        margin-bottom: 0;
      }

      .category-title {
        margin: 0 0 12px 0;
        font-size: 14px;
        font-weight: 600;
        color: #313238;
        padding-bottom: 8px;
        border-bottom: 1px solid #f0f1f5;
      }

      .permission-items {
        .permission-checkbox {
          display: block;
          margin-bottom: 16px;

          &:last-child {
            margin-bottom: 0;
          }

          .permission-label {
            font-weight: 500;
            color: #313238;
          }

          .permission-desc {
            color: #979ba5;
            font-size: 12px;
            margin-top: 4px;
            margin-left: 24px;
            line-height: 1.4;
          }
        }
      }
    }
  }
}
</style>
