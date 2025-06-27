<template>
  <div class="user-form">
    <bk-form
      ref="userForm"
      :label-width="120"
      :model="formData"
      :rules="rules">

      <bk-form-item label="邮箱" property="email" required>
        <bk-input
          v-model="formData.email"
          placeholder="请输入用户邮箱">
        </bk-input>
      </bk-form-item>

      <bk-form-item label="姓名" property="name" required>
        <bk-input
          v-model="formData.name"
          placeholder="请输入用户姓名">
        </bk-input>
      </bk-form-item>

      <bk-form-item label="角色" property="role" required>
        <bk-radio-group v-model="formData.role">
          <bk-radio value="admin">
            <span>管理员</span>
            <div class="role-desc">拥有所有权限，可以管理用户和系统配置</div>
          </bk-radio>
          <bk-radio value="operator">
            <span>操作员</span>
            <div class="role-desc">拥有基本操作权限，可以查看和操作业务数据</div>
          </bk-radio>
        </bk-radio-group>
      </bk-form-item>

      <bk-form-item label="权限设置" v-if="formData.role">
        <div class="permission-grid">
          <div
            v-for="permission in availablePermissions"
            :key="permission.key"
            class="permission-item">
            <bk-checkbox
              :value="formData.permissions.includes(permission.key)"
              :disabled="isPermissionDisabled(permission.key)"
              @change="handlePermissionChange(permission.key, $event)">
              {{ permission.label }}
            </bk-checkbox>
            <div class="permission-desc">{{ permission.description }}</div>
          </div>
        </div>
      </bk-form-item>

      <bk-form-item label="状态" property="status" v-if="isEdit">
        <bk-radio-group v-model="formData.status">
          <bk-radio value="active">启用</bk-radio>
          <bk-radio value="inactive">禁用</bk-radio>
        </bk-radio-group>
      </bk-form-item>

    </bk-form>
  </div>
</template>

<script>
  export default {
    name: 'UserForm',
    props: {
      userData: {
        type: Object,
        default: null
      },
      isEdit: {
        type: Boolean,
        default: false
      }
    },
    data() {
      return {
        formData: {
          email: '',
          name: '',
          role: '',
          permissions: [],
          status: 'active'
        },
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
        ],
        rules: {
          email: [
            {
              required: true,
              message: '邮箱不能为空',
              trigger: 'blur'
            },
            {
              type: 'email',
              message: '请输入正确的邮箱格式',
              trigger: 'blur'
            }
          ],
          name: [
            {
              required: true,
              message: '姓名不能为空',
              trigger: 'blur'
            },
            {
              min: 2,
              max: 20,
              message: '姓名长度应在2-20个字符之间',
              trigger: 'blur'
            }
          ],
          role: [
            {
              required: true,
              message: '请选择用户角色',
              trigger: 'change'
            }
          ]
        }
      }
    },
    watch: {
      userData: {
        handler(newVal) {
          if (newVal) {
            // 确保 permissions 始终是数组
            let permissions = newVal.permissions || []
            if (!Array.isArray(permissions)) {
              permissions = []
            }
            
            this.formData = {
              email: newVal.email || '',
              name: newVal.name || '',
              role: newVal.role || '',
              permissions: permissions,
              status: newVal.status || 'active'
            }
          } else {
            this.resetForm()
          }
        },
        immediate: true
      },
      'formData.role'(newRole) {
        if (newRole) {
          this.setDefaultPermissions(newRole)
        }
      }
    },
    methods: {
      validate() {
        return new Promise((resolve, reject) => {
          this.$refs.userForm.validate((valid) => {
            if (valid) {
              // 确保提交的数据中 permissions 是数组
              const submitData = { ...this.formData }
              if (!Array.isArray(submitData.permissions)) {
                submitData.permissions = []
              }
              resolve(submitData)
            } else {
              reject(new Error('表单验证失败'))
            }
          })
        })
      },

      resetForm() {
        this.formData = {
          email: '',
          name: '',
          role: '',
          permissions: [],
          status: 'active'
        }
        this.$nextTick(() => {
          this.$refs.userForm?.clearError()
        })
      },

      setDefaultPermissions(role) {
        // 确保 permissions 始终是数组
        if (!Array.isArray(this.formData.permissions)) {
          this.formData.permissions = []
        }
        
        if (role === 'admin') {
          this.formData.permissions = this.availablePermissions.map(p => p.key)
        } else if (role === 'operator') {
          this.formData.permissions = ['home', 'business', 'resource']
        }
      },

      isPermissionDisabled(permissionKey) {
        if (this.formData.role === 'admin') {
          return permissionKey === 'admin'
        }
        return false
      },

      handlePermissionChange(permissionKey, checked) {
        // 确保 permissions 是数组
        if (!Array.isArray(this.formData.permissions)) {
          this.formData.permissions = []
        }

        if (checked) {
          // 添加权限（如果不存在）
          if (!this.formData.permissions.includes(permissionKey)) {
            this.formData.permissions.push(permissionKey)
          }
        } else {
          // 移除权限
          const index = this.formData.permissions.indexOf(permissionKey)
          if (index > -1) {
            this.formData.permissions.splice(index, 1)
          }
        }
      }
    }
  }
</script>

<style lang="scss" scoped>
.user-form {
  .role-desc {
    font-size: 12px;
    color: #979ba5;
    margin-top: 4px;
    line-height: 1.4;
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
</style>
