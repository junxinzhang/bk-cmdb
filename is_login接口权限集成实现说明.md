# is_login接口权限集成实现说明

## 修改概述

根据需求，对现有的 `is_login` 接口进行了扩展，使其在返回登录状态的同时，也返回用户的权限列表，用于控制前端菜单的显示。

## 实现方案

选择了**修改现有 `is_login` 接口**的方案，而不是创建新接口，原因如下：

1. **向后兼容**：现有接口已有 `permission` 字段（原本固定为 `null`），直接填充数据不会破坏现有逻辑
2. **减少前端改动**：前端无需修改调用逻辑，只需处理返回的权限数据
3. **符合RESTful设计**：登录状态检查本身就应该包含用户权限信息
4. **渐进式实现**：当前基于用户名提供默认权限，将来可以轻松集成用户管理服务

## 修改内容

### 1. 文件修改位置

**主要修改文件**：`src/web_server/service/login.go`

### 2. 核心修改

#### 2.1 修改 `IsLogin` 函数

```go
// IsLogin user is login
func (s *Service) IsLogin(c *gin.Context) {
	user := user.NewUser(*s.Config, s.Engine, s.CacheCli, s.ApiCli)
	isLogin := user.LoginUser(c)
	if isLogin {
		// 获取用户权限
		permissions := s.getUserPermissions(c)
		c.JSON(200, gin.H{
			common.HTTPBKAPIErrorCode:    0,
			common.HTTPBKAPIErrorMessage: nil,
			"permission":                 permissions, // 从 nil 改为实际权限数据
			"result":                     true,
		})
		return
	}
	// 未登录时权限依然为 nil
	c.JSON(200, gin.H{
		common.HTTPBKAPIErrorCode:    0,
		common.HTTPBKAPIErrorMessage: "Unauthorized",
		"permission":                 nil,
		"result":                     false,
	})
	return
}
```

#### 2.2 新增 `getUserPermissions` 函数

```go
// getUserPermissions 获取用户权限
func (s *Service) getUserPermissions(c *gin.Context) []string {
	rid := httpheader.GetRid(c.Request.Header)
	
	// 从cookie中获取用户名
	cookieUser, err := c.Cookie(common.BKUser)
	if err != nil || cookieUser == "" {
		blog.Warnf("Failed to get user from cookie, rid: %s", rid)
		return []string{}
	}
	
	// 尝试从用户管理服务获取用户权限
	userInfo, err := s.getUserFromDatabase(c, cookieUser, rid)
	if err != nil {
		blog.Warnf("Failed to get user from database, user: %s, err: %v, fallback to default permissions, rid: %s", 
			cookieUser, err, rid)
		return s.getDefaultPermissionsByRole(cookieUser)
	}
	
	if userInfo == nil {
		blog.Warnf("User not found in database, user: %s, fallback to default permissions, rid: %s", cookieUser, rid)
		return s.getDefaultPermissionsByRole(cookieUser)
	}
	
	// 如果用户权限为空，使用默认权限
	if len(userInfo.Permissions) == 0 {
		blog.Infof("User permissions is empty, user: %s, using default permissions, rid: %s", cookieUser, rid)
		return s.getDefaultPermissionsByRole(cookieUser)
	}
	
	// 返回用户的权限列表
	blog.V(4).Infof("Get user permissions from database, user: %s, permissions: %v, rid: %s", 
		cookieUser, userInfo.Permissions, rid)
	return userInfo.Permissions
}
```

#### 2.3 新增 `getUserFromDatabase` 函数

```go
// getUserFromDatabase 从cc_user_management表中获取用户信息
func (s *Service) getUserFromDatabase(c *gin.Context, username string, rid string) (*metadata.User, error) {
	// 构建请求头
	requestHeader := make(http.Header)
	for k, v := range c.Request.Header {
		requestHeader[k] = v
	}
	requestHeader.Set("BK_User", username)
	requestHeader.Set("HTTP_BLUEKING_SUPPLIER_ID", common.BKDefaultOwnerID)

	kit := rest.NewKitFromHeader(requestHeader, s.Engine.CCErr)

	// 尝试直接通过用户ID获取用户信息
	userInfo, err := s.Engine.CoreAPI.CoreService().UserManagement().GetUser(kit.Ctx, requestHeader, username)
	if err == nil && userInfo != nil {
		return userInfo, nil
	}

	// 如果直接获取失败，尝试通过列表查询
	userListRequest := &metadata.UserListRequest{
		Search: username, // 按用户名或邮箱搜索
		Limit:  10,       // 获取多个结果以便进行精确匹配
	}

	userListResult, err := s.Engine.CoreAPI.CoreService().UserManagement().ListUsers(kit.Ctx, requestHeader, userListRequest)
	if err != nil {
		return nil, err
	}

	// 检查是否找到用户，并进行精确匹配
	if userListResult != nil && userListResult.Total > 0 && len(userListResult.Items) > 0 {
		for _, u := range userListResult.Items {
			if u.UserID == username || strings.EqualFold(u.Email, username) {
				return &u, nil
			}
		}
	}

	return nil, nil
}
```

#### 2.4 新增 `getDefaultPermissionsByRole` 函数

```go
// getDefaultPermissionsByRole 根据用户角色返回默认权限
func (s *Service) getDefaultPermissionsByRole(username string) []string {
	// 管理员用户返回所有权限
	if username == "admin" || username == "Administrator" || username == "root" {
		return []string{"admin", "home", "business", "model", "resource", "operation"}
	}
	
	// 运维用户权限
	if strings.Contains(strings.ToLower(username), "ops") || strings.Contains(strings.ToLower(username), "operator") {
		return []string{"home", "business", "resource", "operation"}
	}
	
	// 开发用户权限
	if strings.Contains(strings.ToLower(username), "dev") || strings.Contains(strings.ToLower(username), "developer") {
		return []string{"home", "business", "model", "resource"}
	}
	
	// 默认用户权限
	return []string{"home", "business", "resource"}
}
```

## 权限说明

### 权限类型定义

- `"admin"`: 平台管理权限
- `"home"`: 首页访问权限
- `"business"`: 业务模块权限  
- `"model"`: 模型管理权限
- `"resource"`: 资源管理权限
- `"operation"`: 运营分析权限

### 用户角色与权限映射

| 用户类型 | 用户名模式 | 权限列表 |
|---------|-----------|----------|
| 管理员 | admin, Administrator, root | 全部权限 |
| 运维用户 | 包含 "ops", "operator" | home, business, resource, operation |
| 开发用户 | 包含 "dev", "developer" | home, business, model, resource |
| 普通用户 | 其他 | home, business, resource |

## 接口返回格式

### 登录成功时的响应

```json
{
  "bk_error_code": 0,
  "bk_error_msg": null,
  "permission": [
    "admin",
    "home", 
    "business",
    "model",
    "resource",
    "operation"
  ],
  "result": true
}
```

### 未登录时的响应

```json
{
  "bk_error_code": 0,
  "bk_error_msg": "Unauthorized",
  "permission": null,
  "result": false
}
```

## 容错机制

1. **Cookie获取失败**：返回空权限列表，用户需要重新登录
2. **用户名为空**：返回空权限列表
3. **默认权限机制**：基于用户名模式智能判断用户角色并分配相应权限

## 扩展计划

当前实现是一个基础版本，将来可以通过以下方式扩展：

1. **集成用户管理服务**：从数据库获取用户的实际权限配置
2. **角色权限管理**：支持动态角色和权限分配
3. **权限缓存**：增加权限信息缓存以提高性能
4. **实时权限更新**：支持权限变更的实时生效

## 前端集成建议

前端可以基于返回的 `permission` 数组来控制菜单显示：

```javascript
// 检查用户是否有某个权限
function hasPermission(permission) {
  const userPermissions = response.permission || [];
  return userPermissions.includes(permission);
}

// 控制菜单显示
if (hasPermission('admin')) {
  // 显示平台管理菜单
}
if (hasPermission('business')) {
  // 显示业务菜单
}
// ... 其他菜单控制逻辑
```

## 测试说明

创建了测试脚本 `test_is_login_api.sh` 用于验证接口功能。使用方法：

```bash
# 确保webserver服务已启动
./test_is_login_api.sh
```

测试脚本会模拟不同类型的用户来验证权限返回是否正确。

## 兼容性

- **向后兼容**：现有前端代码无需修改即可正常工作
- **前向兼容**：前端可以选择性地使用权限信息来增强用户体验
- **服务降级**：用户管理服务不可用时，系统依然可以提供基础权限支持