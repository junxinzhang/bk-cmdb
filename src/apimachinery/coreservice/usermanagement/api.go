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

package usermanagement

import (
	"context"
	"net/http"

	"configcenter/src/common/errors"
	"configcenter/src/common/metadata"
)

// UserManagementInterface 用户管理接口定义
type UserManagementInterface interface {
	// 用户管理
	CreateUser(ctx context.Context, h http.Header, data *metadata.CreateUserRequest) (*metadata.User, errors.CCErrorCoder)
	UpdateUser(ctx context.Context, h http.Header, userID string, data *metadata.UpdateUserRequest) (*metadata.User, errors.CCErrorCoder)
	DeleteUser(ctx context.Context, h http.Header, userID string) errors.CCErrorCoder
	GetUser(ctx context.Context, h http.Header, userID string) (*metadata.User, errors.CCErrorCoder)
	ListUsers(ctx context.Context, h http.Header, params *metadata.UserListRequest) (*metadata.UserListResult, errors.CCErrorCoder)
	BatchDeleteUsers(ctx context.Context, h http.Header, data *metadata.BatchDeleteUsersRequest) errors.CCErrorCoder
	
	// 用户状态管理
	ToggleUserStatus(ctx context.Context, h http.Header, userID string, data *metadata.UserStatusRequest) (*metadata.User, errors.CCErrorCoder)
	ResetUserPassword(ctx context.Context, h http.Header, userID string) (*metadata.ResetPasswordResult, errors.CCErrorCoder)
	
	// 用户统计和查询
	GetUserStatistics(ctx context.Context, h http.Header) (*metadata.UserStatistics, errors.CCErrorCoder)
	ValidateEmail(ctx context.Context, h http.Header, data *metadata.ValidateEmailRequest) (*metadata.ValidateEmailResult, errors.CCErrorCoder)
	
	// 用户导入导出
	ExportUsers(ctx context.Context, h http.Header, params *metadata.UserExportRequest) ([]byte, errors.CCErrorCoder)
	ImportUsers(ctx context.Context, h http.Header, data *metadata.UserImportRequest) (*metadata.UserImportResult, errors.CCErrorCoder)
	
	// 角色权限管理
	CreateRolePermission(ctx context.Context, h http.Header, data *metadata.CreateRolePermissionRequest) (*metadata.RolePermission, errors.CCErrorCoder)
	UpdateRolePermission(ctx context.Context, h http.Header, roleID string, data *metadata.UpdateRolePermissionRequest) (*metadata.RolePermission, errors.CCErrorCoder)
	DeleteRolePermission(ctx context.Context, h http.Header, roleID string) errors.CCErrorCoder
	GetRolePermission(ctx context.Context, h http.Header, roleID string) (*metadata.RolePermission, errors.CCErrorCoder)
	ListRolePermissions(ctx context.Context, h http.Header) ([]metadata.RolePermission, errors.CCErrorCoder)
	
	// 权限矩阵
	GetPermissionMatrix(ctx context.Context, h http.Header) (*metadata.PermissionMatrix, errors.CCErrorCoder)
	GetUserRoles(ctx context.Context, h http.Header, roleID string) ([]metadata.UserRoleInfo, errors.CCErrorCoder)
}