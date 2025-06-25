/*
 * Tencent is pleased to support the open source community by making 蓝鲸 available.
 * Copyright (C) 2017-2018 THL A29 Limited, a Tencent company. All rights reserved.
 * Licensed under the MIT License (the "License"); you may not use this file except
 * in compliance with the License. You may obtain a copy of the License at
 * http://opensource.org/licenses/MIT
 * Unless required by applicable law or agreed to in writing, software distributed under
 * the License is distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND,
 * either express or implied. See the License for the specific language governing permissions and
 * limitations under the License.
 */

package metadata

import (
	"time"

	"configcenter/src/common/mapstr"
)

// UserStatus 用户状态枚举
type UserStatus string

const (
	UserStatusActive   UserStatus = "active"   // 活跃
	UserStatusInactive UserStatus = "inactive" // 非活跃
	UserStatusLocked   UserStatus = "locked"   // 锁定
)

// UserRole 用户角色枚举
type UserRole string

const (
	UserRoleAdmin    UserRole = "admin"    // 管理员
	UserRoleOperator UserRole = "operator" // 操作员
	UserRoleReadonly UserRole = "readonly" // 只读用户
)

// User 用户数据模型
type User struct {
	ID          string            `json:"id" bson:"_id"`
	UserID      string            `json:"user_id" bson:"user_id"`
	Email       string            `json:"email" bson:"email"`
	Name        string            `json:"name" bson:"name"`
	Role        UserRole          `json:"role" bson:"role"`
	Permissions []string          `json:"permissions" bson:"permissions"`
	Status      UserStatus        `json:"status" bson:"status"`
	CreatedAt   time.Time         `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at" bson:"updated_at"`
	CreatedBy   string            `json:"created_by" bson:"created_by"`
	LastLogin   *time.Time        `json:"last_login,omitempty" bson:"last_login,omitempty"`
	LoginCount  int64             `json:"login_count" bson:"login_count"`
	Metadata    mapstr.MapStr     `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// RolePermission 角色权限数据模型
type RolePermission struct {
	ID          string        `json:"id" bson:"_id"`
	RoleName    string        `json:"role_name" bson:"role_name"`
	Permissions []string      `json:"permissions" bson:"permissions"`
	Description string        `json:"description" bson:"description"`
	IsSystem    bool          `json:"is_system" bson:"is_system"`
	CreatedAt   time.Time     `json:"created_at" bson:"created_at"`
	UpdatedAt   time.Time     `json:"updated_at" bson:"updated_at"`
	Metadata    mapstr.MapStr `json:"metadata,omitempty" bson:"metadata,omitempty"`
}

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Email       string            `json:"email" validate:"required,email"`
	Name        string            `json:"name" validate:"required,min=2,max=50"`
	Role        UserRole          `json:"role" validate:"required"`
	Permissions []string          `json:"permissions,omitempty"`
	Status      UserStatus        `json:"status,omitempty"`
	Metadata    mapstr.MapStr     `json:"metadata,omitempty"`
}

// UpdateUserRequest 更新用户请求
type UpdateUserRequest struct {
	Name        *string           `json:"name,omitempty" validate:"omitempty,min=2,max=50"`
	Role        *UserRole         `json:"role,omitempty"`
	Permissions []string          `json:"permissions,omitempty"`
	Status      *UserStatus       `json:"status,omitempty"`
	Metadata    mapstr.MapStr     `json:"metadata,omitempty"`
}

// UserListRequest 用户列表请求
type UserListRequest struct {
	Page       int        `json:"page" form:"page"`
	Limit      int        `json:"limit" form:"limit"`
	Search     string     `json:"search" form:"search"`
	Role       UserRole   `json:"role" form:"role"`
	Status     UserStatus `json:"status" form:"status"`
	SortField  string     `json:"sort_field" form:"sort_field"`
	SortOrder  string     `json:"sort_order" form:"sort_order"`
}

// UserListResponse 用户列表响应
type UserListResponse struct {
	BaseResp
	Data *UserListResult `json:"data"`
}

// UserListResult 用户列表结果
type UserListResult struct {
	Items      []User `json:"items"`
	Total      int64  `json:"total"`
	Page       int    `json:"page"`
	Limit      int    `json:"limit"`
	TotalPages int    `json:"total_pages"`
}

// UserDetailResponse 用户详情响应
type UserDetailResponse struct {
	BaseResp
	Data *User `json:"data"`
}

// CreateUserResponse 创建用户响应
type CreateUserResponse struct {
	BaseResp
	Data *User `json:"data"`
}

// UpdateUserResponse 更新用户响应
type UpdateUserResponse struct {
	BaseResp
	Data *User `json:"data"`
}

// BatchDeleteUsersRequest 批量删除用户请求
type BatchDeleteUsersRequest struct {
	UserIDs []string `json:"user_ids" validate:"required,min=1"`
}

// UserStatusRequest 用户状态切换请求
type UserStatusRequest struct {
	Status UserStatus `json:"status" validate:"required"`
}

// ResetPasswordResponse 重置密码响应
type ResetPasswordResponse struct {
	BaseResp
	Data *ResetPasswordResult `json:"data"`
}

// ResetPasswordResult 重置密码结果
type ResetPasswordResult struct {
	TempPassword string `json:"temp_password"`
	ExpiresAt    string `json:"expires_at"`
}

// UserStatisticsResponse 用户统计信息响应
type UserStatisticsResponse struct {
	BaseResp
	Data *UserStatistics `json:"data"`
}

// UserStatistics 用户统计信息
type UserStatistics struct {
	TotalUsers    int64 `json:"total_users"`
	ActiveUsers   int64 `json:"active_users"`
	InactiveUsers int64 `json:"inactive_users"`
	LockedUsers   int64 `json:"locked_users"`
	AdminUsers    int64 `json:"admin_users"`
	OperatorUsers int64 `json:"operator_users"`
	ReadonlyUsers int64 `json:"readonly_users"`
}

// ValidateEmailRequest 验证邮箱请求
type ValidateEmailRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// ValidateEmailResponse 验证邮箱响应
type ValidateEmailResponse struct {
	BaseResp
	Data *ValidateEmailResult `json:"data"`
}

// ValidateEmailResult 验证邮箱结果
type ValidateEmailResult struct {
	Available bool   `json:"available"`
	Message   string `json:"message,omitempty"`
}

// UserExportRequest 用户导出请求
type UserExportRequest struct {
	Role   UserRole   `json:"role" form:"role"`
	Status UserStatus `json:"status" form:"status"`
	Format string     `json:"format" form:"format"`
}

// UserImportRequest 用户导入请求
type UserImportRequest struct {
	File     []byte `json:"file"`
	FileName string `json:"file_name"`
	Format   string `json:"format"`
}

// UserImportResponse 用户导入响应
type UserImportResponse struct {
	BaseResp
	Data *UserImportResult `json:"data"`
}

// UserImportResult 用户导入结果
type UserImportResult struct {
	Total     int               `json:"total"`
	Success   int               `json:"success"`
	Failed    int               `json:"failed"`
	Errors    []ImportError     `json:"errors,omitempty"`
	CreatedUsers []User         `json:"created_users,omitempty"`
}

// ImportError 导入错误
type ImportError struct {
	Row     int    `json:"row"`
	Email   string `json:"email"`
	Message string `json:"message"`
}

// CreateRolePermissionRequest 创建角色权限请求
type CreateRolePermissionRequest struct {
	RoleName    string        `json:"role_name" validate:"required,min=2,max=50"`
	Permissions []string      `json:"permissions" validate:"required,min=1"`
	Description string        `json:"description,omitempty"`
	Metadata    mapstr.MapStr `json:"metadata,omitempty"`
}

// UpdateRolePermissionRequest 更新角色权限请求
type UpdateRolePermissionRequest struct {
	Permissions []string      `json:"permissions,omitempty"`
	Description *string       `json:"description,omitempty"`
	Metadata    mapstr.MapStr `json:"metadata,omitempty"`
}

// RolePermissionListResponse 角色权限列表响应
type RolePermissionListResponse struct {
	BaseResp
	Data []RolePermission `json:"data"`
}

// RolePermissionResponse 角色权限响应
type RolePermissionResponse struct {
	BaseResp
	Data *RolePermission `json:"data"`
}

// PermissionMatrixResponse 权限矩阵响应
type PermissionMatrixResponse struct {
	BaseResp
	Data *PermissionMatrix `json:"data"`
}

// PermissionMatrix 权限矩阵
type PermissionMatrix struct {
	Permissions []UserPermission `json:"permissions"`
	Roles       []RoleInfo       `json:"roles"`
	Matrix      [][]bool         `json:"matrix"`
}

// UserPermission 权限定义
type UserPermission struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Category    string `json:"category"`
}

// RoleInfo 角色信息
type RoleInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	IsSystem    bool   `json:"is_system"`
}

// GetUserRolesResponse 获取用户角色响应
type GetUserRolesResponse struct {
	BaseResp
	Data []UserRoleInfo `json:"data"`
}

// UserRoleInfo 用户角色信息
type UserRoleInfo struct {
	UserID   string `json:"user_id"`
	Email    string `json:"email"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Status   string `json:"status"`
	JoinDate string `json:"join_date"`
}