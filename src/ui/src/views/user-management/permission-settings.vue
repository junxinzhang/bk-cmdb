<template>
  <div class="permission-settings">
    <div class="settings-header">
      <h3>权限设置</h3>
      <p class="settings-desc">配置不同角色的菜单访问权限</p>
    </div>

    <div class="permission-matrix">
      <div class="matrix-header">
        <div class="matrix-cell header-cell">权限项目</div>
        <div
          v-for="role in roles"
          :key="role.key"
          class="matrix-cell header-cell role-header">
          <bk-tag :theme="role.theme">{{ role.name }}</bk-tag>
        </div>
      </div>

      <div
        v-for="permission in permissions"
        :key="permission.key"
        class="matrix-row">
        <div class="matrix-cell permission-cell">
          <div class="permission-info">
            <div class="permission-name">
              <i :class="permission.icon" class="permission-icon"></i>
              {{ permission.label }}
            </div>
            <div class="permission-desc">{{ permission.description }}</div>
          </div>
        </div>
        
        <div
          v-for="role in roles"
          :key="`${permission.key}-${role.key}`"
          class="matrix-cell checkbox-cell">
          <bk-checkbox
            :value="hasPermission(role.key, permission.key)"
            :disabled="isPermissionLocked(role.key, permission.key)"
            @change="togglePermission(role.key, permission.key, $event)">
          </bk-checkbox>
        </div>
      </div>
    </div>

    <div class="settings-actions">
      <bk-button
        theme="primary"
        :loading="saving"
        @click="savePermissions">
        保存设置
      </bk-button>
      <bk-button @click="resetPermissions">重置</bk-button>
    </div>

    <div class="permission-preview">
      <h4>权限预览</h4>
      <div class="preview-tabs">
        <bk-tab :active.sync="previewRole" type="unborder-card">
          <bk-tab-panel
            v-for="role in roles"
            :key="role.key"
            :name="role.key"
            :label="role.name">
            <div class="preview-content">
              <div class="preview-item">
                <h5>可访问菜单：</h5>
                <div class="menu-list">
                  <bk-tag
                    v-for="permission in getAccessibleMenus(role.key)"
                    :key="permission.key"
                    class="menu-tag">
                    <i :class="permission.icon"></i>
                    {{ permission.label }}
                  </bk-tag>
                </div>
              </div>
              <div class="preview-item">
                <h5>权限说明：</h5>
                <ul class="permission-descriptions">
                  <li
                    v-for="permission in getAccessibleMenus(role.key)"
                    :key="permission.key">
                    <strong>{{ permission.label }}：</strong>{{ permission.description }}
                  </li>
                </ul>
              </div>
            </div>
          </bk-tab-panel>
        </bk-tab>
      </div>
    </div>
  </div>
</template>

<script>
import { mapState, mapActions } from 'vuex'

export default {
  name: 'PermissionSettings',
  data() {
    return {
      saving: false,
      previewRole: 'admin',
      permissions: [
        {
          key: 'home',
          label: '首页',
          description: '访问系统首页，查看系统概览和统计数据',
          icon: 'bk-icon icon-home'
        },
        {
          key: 'business',
          label: '业务',
          description: '管理业务拓扑、服务实例、进程配置等业务相关功能',
          icon: 'bk-icon icon-apps'
        },
        {
          key: 'resource',
          label: '资源',
          description: '管理主机资源、云区域配置、资源池等基础资源',
          icon: 'bk-icon icon-host'
        },
        {
          key: 'model',
          label: '模型',
          description: '管理配置模型、对象属性、模型关联等数据模型',
          icon: 'bk-icon icon-cc-model'
        },
        {
          key: 'operation',
          label: '运营分析',
          description: '查看运营数据分析、审计日志、系统监控等运营信息',
          icon: 'bk-icon icon-bar-chart'
        },
        {
          key: 'admin',
          label: '平台管理',
          description: '系统配置、用户管理、权限设置、全局配置等管理功能',
          icon: 'bk-icon icon-cog'
        }
      ],
      localPermissionMatrix: {}
    }
  },
  computed: {
    ...mapState('rolePermission', [
      'roles',
      'permissionMatrix'
    ])
  },
  watch: {
    permissionMatrix: {
      handler(newMatrix) {
        this.localPermissionMatrix = JSON.parse(JSON.stringify(newMatrix))
      },
      immediate: true,
      deep: true
    }
  },
  created() {
    this.fetchPermissionMatrix()
  },
  methods: {
    ...mapActions('rolePermission', [
      'getPermissionMatrix',
      'updatePermissionMatrix'
    ]),

    async fetchPermissionMatrix() {
      try {
        await this.getPermissionMatrix()
      } catch (error) {
        this.$bkMessage({
          theme: 'error',
          message: error.message || '获取权限矩阵失败'
        })
      }
    },

    hasPermission(roleKey, permissionKey) {
      return this.localPermissionMatrix[roleKey]?.includes(permissionKey) || false
    },

    togglePermission(roleKey, permissionKey, hasPermission) {
      if (!this.localPermissionMatrix[roleKey]) {
        this.$set(this.localPermissionMatrix, roleKey, [])
      }
      
      const permissions = this.localPermissionMatrix[roleKey]
      if (hasPermission) {
        if (!permissions.includes(permissionKey)) {
          permissions.push(permissionKey)
        }
      } else {
        const index = permissions.indexOf(permissionKey)
        if (index > -1) {
          permissions.splice(index, 1)
        }
      }
    },

    isPermissionLocked(roleKey, permissionKey) {
      if (roleKey === 'admin') {
        return true
      }
      return false
    },

    async savePermissions() {
      this.saving = true
      try {
        await this.updatePermissionMatrix(this.localPermissionMatrix)
        this.$bkMessage({
          theme: 'success',
          message: '权限设置保存成功'
        })
      } catch (error) {
        this.$bkMessage({
          theme: 'error',
          message: error.message || '保存失败'
        })
      } finally {
        this.saving = false
      }
    },

    resetPermissions() {
      this.localPermissionMatrix = JSON.parse(JSON.stringify(this.permissionMatrix))
      this.$bkMessage({
        theme: 'success',
        message: '已重置为原始设置'
      })
    },

    getAccessibleMenus(roleKey) {
      const rolePermissions = this.localPermissionMatrix[roleKey] || []
      return this.permissions.filter(permission => 
        rolePermissions.includes(permission.key)
      )
    }
  }
}
</script>

<style lang="scss" scoped>
.permission-settings {
  padding: 20px;
  
  .settings-header {
    margin-bottom: 24px;
    
    h3 {
      margin: 0 0 8px 0;
      font-size: 16px;
      color: #313238;
    }
    
    .settings-desc {
      margin: 0;
      color: #979ba5;
      font-size: 12px;
    }
  }
  
  .permission-matrix {
    background: #fff;
    border: 1px solid #dcdee5;
    border-radius: 2px;
    margin-bottom: 20px;
    
    .matrix-header {
      display: grid;
      grid-template-columns: 300px repeat(auto-fit, 120px);
      background: #fafbfd;
      border-bottom: 1px solid #dcdee5;
      
      .header-cell {
        padding: 16px;
        font-weight: 600;
        border-right: 1px solid #dcdee5;
        
        &:last-child {
          border-right: none;
        }
        
        &.role-header {
          text-align: center;
        }
      }
    }
    
    .matrix-row {
      display: grid;
      grid-template-columns: 300px repeat(auto-fit, 120px);
      border-bottom: 1px solid #f0f1f5;
      
      &:last-child {
        border-bottom: none;
      }
      
      .matrix-cell {
        padding: 16px;
        border-right: 1px solid #f0f1f5;
        
        &:last-child {
          border-right: none;
        }
        
        &.checkbox-cell {
          display: flex;
          justify-content: center;
          align-items: center;
        }
      }
      
      .permission-cell {
        .permission-info {
          .permission-name {
            display: flex;
            align-items: center;
            font-weight: 500;
            margin-bottom: 4px;
            
            .permission-icon {
              margin-right: 8px;
              color: #3a84ff;
            }
          }
          
          .permission-desc {
            color: #979ba5;
            font-size: 12px;
            line-height: 1.4;
          }
        }
      }
    }
  }
  
  .settings-actions {
    margin-bottom: 32px;
    
    .bk-button {
      margin-right: 8px;
    }
  }
  
  .permission-preview {
    background: #fff;
    border: 1px solid #dcdee5;
    border-radius: 2px;
    padding: 20px;
    
    h4 {
      margin: 0 0 16px 0;
      font-size: 14px;
      color: #313238;
    }
    
    .preview-content {
      .preview-item {
        margin-bottom: 20px;
        
        &:last-child {
          margin-bottom: 0;
        }
        
        h5 {
          margin: 0 0 12px 0;
          font-size: 13px;
          color: #63656e;
        }
        
        .menu-list {
          .menu-tag {
            margin-right: 8px;
            margin-bottom: 8px;
            
            i {
              margin-right: 4px;
            }
          }
        }
        
        .permission-descriptions {
          margin: 0;
          padding-left: 16px;
          
          li {
            margin-bottom: 8px;
            font-size: 12px;
            color: #63656e;
            line-height: 1.5;
            
            &:last-child {
              margin-bottom: 0;
            }
          }
        }
      }
    }
  }
}
</style>