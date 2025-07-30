# is_login 接口权限集成完整实现方案

## 实现总结

已成功修改 `is_login` 接口，使其能够从 `cc_user_management` 表中获取用户权限并返回给前端，前端根据权限控制菜单显示。

## 后端修改

### 1. 修改 IsLogin 接口 (`src/web_server/service/login.go`)

#### A. 更新响应格式
```go
// IsLogin user is login
func (s *Service) IsLogin(c *gin.Context) {
    rid := httpheader.GetRid(c.Request.Header)
    user := user.NewUser(*s.Config, s.Engine, s.CacheCli, s.ApiCli)
    isLogin := user.LoginUser(c)
    
    if isLogin {
        // 获取当前用户权限
        permissions, err := s.getUserPermissions(c, rid)
        if err != nil {
            blog.Warnf("failed to get user permissions, err: %v, rid: %s", err, rid)
            // 权限获取失败时返回空权限，但不影响登录状态
            permissions = []string{}
        }
        
        c.JSON(200, gin.H{
            common.HTTPBKAPIErrorCode:    0,
            common.HTTPBKAPIErrorMessage: nil,
            "permissions":                permissions,  // 新增权限字段
            "result":                     true,
        })
        return
    }
    // ... 未登录响应
}
```

#### B. 新增 getUserPermissions 方法
```go
// getUserPermissions 获取用户权限
func (s *Service) getUserPermissions(c *gin.Context, rid string) ([]string, error) {
    // 1. 从session获取当前用户名
    session := sessions.Default(c)
    userName, ok := session.Get(common.WEBSessionUinKey).(string)
    
    // 2. 调用用户管理API查询用户信息
    req := &metadata.UserListRequest{
        Search: userName,
        Page:   1,
        Limit:  1,
    }
    
    result, err := s.ApiCli.CoreService().UserManagement().ListUsers(
        c.Request.Context(), 
        c.Request.Header, 
        req
    )
    
    // 3. 返回用户权限
    if result != nil && len(result.Items) > 0 {
        return result.Items[0].Permissions, nil
    }
    
    // 4. 用户不存在时返回默认权限
    return []string{"home"}, nil
}
```

### 2. 数据来源：cc_user_management 表

用户权限存储在 `cc_user_management` 表的 `permissions` 字段中：

```javascript
// cc_user_management 表结构示例
{
  "_id": "...",
  "user_id": "admin",
  "email": "admin@example.com",
  "name": "管理员",
  "role": "admin",
  "permissions": [
    "home",
    "business", 
    "resource",
    "model",
    "operation",
    "admin"
  ],
  "status": "active",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z"
}
```

## 前端修改

### 1. 权限获取和存储 (`src/ui/src/main.js`)

```javascript
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

### 2. Store 权限存储 (`src/ui/src/store/modules/global.js`)

```javascript
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

### 3. 菜单权限配置 (`src/ui/src/dictionary/menu.js`)

```javascript
const menus = [{
  id: MENU_INDEX,
  i18n: '首页',
  permission: 'home'      // 对应权限标识
}, {
  id: MENU_BUSINESS,
  i18n: '业务',
  permission: 'business'  // 对应权限标识
}, {
  id: MENU_RESOURCE,
  i18n: '资源',
  permission: 'resource'  // 对应权限标识
}, {
  id: MENU_MODEL,
  i18n: '模型',
  permission: 'model'     // 对应权限标识
}, {
  id: MENU_ANALYSIS,
  i18n: '运营分析',
  permission: 'operation' // 对应权限标识
}, {
  id: MENU_PLATFORM_MANAGEMENT,
  i18n: '平台管理',
  permission: 'admin'     // 对应权限标识
}]
```

### 4. 菜单权限检查 (`src/ui/src/components/layout/header.vue`)

```javascript
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

## API 响应格式

### 登录成功响应
```json
{
  "code": 0,
  "message": null,
  "permissions": [
    "home",
    "business",
    "resource", 
    "model",
    "operation",
    "admin"
  ],
  "result": true
}
```

### 未登录响应
```json
{
  "code": 0,
  "message": "Unauthorized",
  "permissions": null,
  "result": false
}
```

## 权限映射关系

| Permission Key | 菜单名称 | 说明 |
|---|---|---|
| `home` | 首页 | 系统首页访问权限 |
| `business` | 业务/业务集 | 业务管理相关功能 |
| `resource` | 资源 | 资源管理相关功能 |
| `model` | 模型 | 模型管理相关功能 |
| `operation` | 运营分析 | 运营分析相关功能 |
| `admin` | 平台管理 | 平台管理相关功能 |

## 数据流程

1. **用户访问系统** → 调用 `is_login` 接口
2. **is_login 接口处理**：
   - 检查用户登录状态
   - 从session获取当前用户名
   - 调用 UserManagement API 查询 `cc_user_management` 表
   - 获取用户的 `permissions` 字段
3. **返回响应** → 包含用户权限的JSON响应
4. **前端处理**：
   - main.js 解析响应，保存权限到 store 和 window.User
   - header.vue 从 store 获取权限，过滤菜单
5. **菜单渲染** → 只显示用户有权限的菜单项

## 错误处理

1. **用户不存在**：返回默认权限 `["home"]`
2. **权限获取失败**：返回空权限 `[]`，但不影响登录状态
3. **Session 中无用户名**：记录警告日志，返回错误

## 优势

- ✅ **数据源准确**：直接从 `cc_user_management` 表获取权限
- ✅ **实时性**：每次登录检查时获取最新权限
- ✅ **容错性强**：权限获取失败不影响登录功能
- ✅ **向下兼容**：保持原有登录检查逻辑不变
- ✅ **安全性**：前端权限仅用于界面控制，后端仍需独立验证

## 注意事项

1. **性能考虑**：每次 `is_login` 调用都会查询数据库，建议后续可加入缓存机制
2. **默认权限**：新用户或权限为空时的处理策略可根据业务需求调整
3. **权限更新**：用户权限变更后需要重新登录或刷新页面才能生效