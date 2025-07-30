# is_login 接口权限集成方案

## 方案概述

通过修改 `is_login` 接口，在登录状态检查时同时返回用户的 `permissions` 字段，并在前端保存到全局状态中，用于控制菜单的显示。

## 实现细节

### 1. 后端接口修改

`is_login` 接口需要返回用户权限信息：

```json
{
  "result": true,
  "code": 0,
  "data": {
    "is_login": true,
    "user_info": {
      "name": "admin",
      "permissions": [
        "home",
        "business", 
        "resource",
        "model",
        "operation",
        "admin"
      ]
    }
  }
}
```

### 2. 前端权限获取和存储

#### 修改 main.js 中的登录检查逻辑：

```javascript
// src/ui/src/main.js
api.get(`${window.API_HOST}is_login`).then((response) => {
  // 从登录状态检查接口获取用户权限
  const userInfo = response.data || response
  
  // 如果响应中包含权限信息，保存到全局状态
  if (userInfo && userInfo.permissions) {
    // 更新window.User对象
    if (window.User) {
      window.User.permissions = userInfo.permissions
    }
    
    // 更新store中的用户信息
    store.commit('global/setUserPermissions', userInfo.permissions)
  }
  
  window.CMDB_APP = new Vue({
    el: '#app',
    router,
    store,
    i18n,
    render() {
      return !subEnv ? <App /> : <IframeEntry />
    }
  })
})
```

#### 添加 store mutation：

```javascript
// src/ui/src/store/modules/global.js
const mutations = {
  // ... 其他 mutations
  setUserPermissions(state, permissions) {
    // 设置用户权限到user对象中
    if (state.user) {
      state.user.permissions = permissions
    } else {
      state.user = { ...window.User, permissions }
    }
  }
}
```

### 3. 菜单权限配置

#### 菜单配置添加权限标识：

```javascript
// src/ui/src/dictionary/menu.js
const menus = [{
  id: MENU_INDEX,
  i18n: '首页',
  permission: 'home'  // 对应权限标识
}, {
  id: MENU_BUSINESS,
  i18n: '业务',
  permission: 'business'
}, {
  id: MENU_RESOURCE,
  i18n: '资源',
  permission: 'resource'
}, {
  id: MENU_MODEL,
  i18n: '模型',
  permission: 'model'
}, {
  id: MENU_ANALYSIS,
  i18n: '运营分析',
  permission: 'operation'
}, {
  id: MENU_PLATFORM_MANAGEMENT,
  i18n: '平台管理',
  permission: 'admin'
}]
```

### 4. 菜单权限检查逻辑

#### header.vue 中的权限过滤：

```javascript
// src/ui/src/components/layout/header.vue
computed: {
  visibleMenu() {
    return menu.filter((menuItem) => {
      // 检查权限
      if (has(menuItem, 'permission')) {
        // 从store或window.User获取用户权限
        const userPermissions = this.$store.state.global.user?.permissions || window.User?.permissions || []
        
        // 检查是否有权限
        if (!Array.isArray(userPermissions) || !userPermissions.includes(menuItem.permission)) {
          return false
        }
      }

      // 检查原有的可见性逻辑
      if (!has(menuItem, 'visibility')) {
        return true
      }

      if (typeof menuItem.visibility === 'function') {
        return menuItem.visibility(this)
      }
      return menuItem.visibility
    })
  }
}
```

## 权限与菜单映射关系

| Permission Key | 菜单名称 | 说明 |
|---|---|---|
| `home` | 首页 | 系统首页访问权限 |
| `business` | 业务/业务集 | 业务管理相关功能 |
| `resource` | 资源 | 资源管理相关功能 |
| `model` | 模型 | 模型管理相关功能 |
| `operation` | 运营分析 | 运营分析相关功能 |
| `admin` | 平台管理 | 平台管理相关功能 |

## 测试验证

### 1. 开发环境调试

在浏览器控制台中查看权限获取情况：

```javascript
// 查看当前用户权限
console.log('Store user permissions:', window.CMDB_APP.$store.state.global.user?.permissions)
console.log('Window user permissions:', window.User?.permissions)

// 查看菜单过滤结果
console.log('Visible menus:', window.CMDB_APP.$children[0].$refs.header.visibleMenu)
```

### 2. 权限测试用例

#### 管理员权限：
```json
{
  "permissions": ["home", "business", "resource", "model", "operation", "admin"]
}
```
预期：显示所有菜单

#### 操作员权限：
```json
{
  "permissions": ["home", "business", "resource"]
}
```
预期：显示首页、业务、资源菜单

#### 只读用户：
```json
{
  "permissions": ["home"]
}
```
预期：只显示首页菜单

## 数据流程

1. **用户访问系统** → 调用 `is_login` 接口
2. **is_login 接口返回** → 包含用户权限的响应
3. **前端 main.js** → 解析响应，保存权限到 store 和 window.User
4. **header.vue 组件** → 从 store 获取权限，过滤菜单
5. **菜单渲染** → 只显示用户有权限的菜单项

## 注意事项

1. **向下兼容**：没有配置 `permission` 的菜单项会默认显示
2. **安全性**：前端权限控制仅用于界面显示，后端API仍需独立权限验证
3. **实时更新**：权限变更时需要刷新页面或重新调用权限接口
4. **错误处理**：权限获取失败时的降级处理

## 优势

- ✅ 集成简单，只需修改现有的 `is_login` 接口
- ✅ 权限数据在应用启动时一次性获取
- ✅ 支持多种权限获取路径，具有良好的容错性
- ✅ 开发环境提供详细的调试信息
- ✅ 向下兼容现有的菜单可见性逻辑