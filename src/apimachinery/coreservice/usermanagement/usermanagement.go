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
	"fmt"
	"net/http"

	"configcenter/src/apimachinery/rest"
	"configcenter/src/common/errors"
	"configcenter/src/common/metadata"
)

type userManagement struct {
	client rest.ClientInterface
}

// NewUserManagementInterface 创建用户管理接口客户端
func NewUserManagementInterface(client rest.ClientInterface) UserManagementInterface {
	return &userManagement{
		client: client,
	}
}

// CreateUser 创建用户
func (u *userManagement) CreateUser(ctx context.Context, h http.Header, data *metadata.CreateUserRequest) (*metadata.User, errors.CCErrorCoder) {
	resp := new(metadata.CreateUserResponse)
	subPath := "/create/user"
	
	err := u.client.Post().
		WithContext(ctx).
		Body(data).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}

// UpdateUser 更新用户
func (u *userManagement) UpdateUser(ctx context.Context, h http.Header, userID string, data *metadata.UpdateUserRequest) (*metadata.User, errors.CCErrorCoder) {
	resp := new(metadata.UpdateUserResponse)
	subPath := fmt.Sprintf("/update/user/%s", userID)
	
	err := u.client.Put().
		WithContext(ctx).
		Body(data).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}

// DeleteUser 删除用户
func (u *userManagement) DeleteUser(ctx context.Context, h http.Header, userID string) errors.CCErrorCoder {
	resp := new(metadata.BaseResp)
	subPath := fmt.Sprintf("/delete/user/%s", userID)
	
	err := u.client.Delete().
		WithContext(ctx).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return err
	}
	
	return nil
}

// GetUser 获取用户详情
func (u *userManagement) GetUser(ctx context.Context, h http.Header, userID string) (*metadata.User, errors.CCErrorCoder) {
	resp := new(metadata.UserDetailResponse)
	subPath := fmt.Sprintf("/find/user/%s", userID)
	
	err := u.client.Get().
		WithContext(ctx).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}

// ListUsers 获取用户列表
func (u *userManagement) ListUsers(ctx context.Context, h http.Header, params *metadata.UserListRequest) (*metadata.UserListResult, errors.CCErrorCoder) {
	resp := new(metadata.UserListResponse)
	subPath := "/find/users"
	
	err := u.client.Post().
		WithContext(ctx).
		Body(params).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}

// BatchDeleteUsers 批量删除用户
func (u *userManagement) BatchDeleteUsers(ctx context.Context, h http.Header, data *metadata.BatchDeleteUsersRequest) errors.CCErrorCoder {
	resp := new(metadata.BaseResp)
	subPath := "/delete/users/batch"
	
	err := u.client.Delete().
		WithContext(ctx).
		Body(data).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return err
	}
	
	return nil
}

// ToggleUserStatus 切换用户状态
func (u *userManagement) ToggleUserStatus(ctx context.Context, h http.Header, userID string, data *metadata.UserStatusRequest) (*metadata.User, errors.CCErrorCoder) {
	resp := new(metadata.UpdateUserResponse)
	subPath := fmt.Sprintf("/update/user/%s/status", userID)
	
	err := u.client.Put().
		WithContext(ctx).
		Body(data).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}

// ResetUserPassword 重置用户密码
func (u *userManagement) ResetUserPassword(ctx context.Context, h http.Header, userID string) (*metadata.ResetPasswordResult, errors.CCErrorCoder) {
	resp := new(metadata.ResetPasswordResponse)
	subPath := fmt.Sprintf("/update/user/%s/reset-password", userID)
	
	err := u.client.Post().
		WithContext(ctx).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}

// GetUserStatistics 获取用户统计信息
func (u *userManagement) GetUserStatistics(ctx context.Context, h http.Header) (*metadata.UserStatistics, errors.CCErrorCoder) {
	resp := new(metadata.UserStatisticsResponse)
	subPath := "/find/users/statistics"
	
	err := u.client.Get().
		WithContext(ctx).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}

// ValidateEmail 验证邮箱是否可用
func (u *userManagement) ValidateEmail(ctx context.Context, h http.Header, data *metadata.ValidateEmailRequest) (*metadata.ValidateEmailResult, errors.CCErrorCoder) {
	resp := new(metadata.ValidateEmailResponse)
	subPath := "/validate/email"
	
	err := u.client.Post().
		WithContext(ctx).
		Body(data).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}

// ExportUsers 导出用户列表
func (u *userManagement) ExportUsers(ctx context.Context, h http.Header, params *metadata.UserExportRequest) ([]byte, errors.CCErrorCoder) {
	subPath := "/export/users"
	
	result := u.client.Get().
		WithContext(ctx).
		WithParam("role", string(params.Role)).
		WithParam("status", string(params.Status)).
		WithParam("format", params.Format).
		SubResourcef(subPath).
		WithHeaders(h).
		Do()
		
	if result.Err != nil {
		return nil, errors.CCHttpError
	}
	
	return result.Body, nil
}

// ImportUsers 导入用户
func (u *userManagement) ImportUsers(ctx context.Context, h http.Header, data *metadata.UserImportRequest) (*metadata.UserImportResult, errors.CCErrorCoder) {
	resp := new(metadata.UserImportResponse)
	subPath := "/import/users"
	
	err := u.client.Post().
		WithContext(ctx).
		Body(data).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}

// CreateRolePermission 创建角色权限
func (u *userManagement) CreateRolePermission(ctx context.Context, h http.Header, data *metadata.CreateRolePermissionRequest) (*metadata.RolePermission, errors.CCErrorCoder) {
	resp := new(metadata.RolePermissionResponse)
	subPath := "/create/role-permission"
	
	err := u.client.Post().
		WithContext(ctx).
		Body(data).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}

// UpdateRolePermission 更新角色权限
func (u *userManagement) UpdateRolePermission(ctx context.Context, h http.Header, roleID string, data *metadata.UpdateRolePermissionRequest) (*metadata.RolePermission, errors.CCErrorCoder) {
	resp := new(metadata.RolePermissionResponse)
	subPath := fmt.Sprintf("/update/role-permission/%s", roleID)
	
	err := u.client.Put().
		WithContext(ctx).
		Body(data).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}

// DeleteRolePermission 删除角色权限
func (u *userManagement) DeleteRolePermission(ctx context.Context, h http.Header, roleID string) errors.CCErrorCoder {
	resp := new(metadata.BaseResp)
	subPath := fmt.Sprintf("/delete/role-permission/%s", roleID)
	
	err := u.client.Delete().
		WithContext(ctx).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return err
	}
	
	return nil
}

// GetRolePermission 获取角色权限详情
func (u *userManagement) GetRolePermission(ctx context.Context, h http.Header, roleID string) (*metadata.RolePermission, errors.CCErrorCoder) {
	resp := new(metadata.RolePermissionResponse)
	subPath := fmt.Sprintf("/find/role-permission/%s", roleID)
	
	err := u.client.Get().
		WithContext(ctx).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}

// ListRolePermissions 获取角色权限列表
func (u *userManagement) ListRolePermissions(ctx context.Context, h http.Header) ([]metadata.RolePermission, errors.CCErrorCoder) {
	resp := new(metadata.RolePermissionListResponse)
	subPath := "/find/role-permissions"
	
	err := u.client.Get().
		WithContext(ctx).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}

// GetPermissionMatrix 获取权限矩阵
func (u *userManagement) GetPermissionMatrix(ctx context.Context, h http.Header) (*metadata.PermissionMatrix, errors.CCErrorCoder) {
	resp := new(metadata.PermissionMatrixResponse)
	subPath := "/find/permission-matrix"
	
	err := u.client.Get().
		WithContext(ctx).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}

// GetUserRoles 获取角色下的用户列表
func (u *userManagement) GetUserRoles(ctx context.Context, h http.Header, roleID string) ([]metadata.UserRoleInfo, errors.CCErrorCoder) {
	resp := new(metadata.GetUserRolesResponse)
	subPath := fmt.Sprintf("/find/role/%s/users", roleID)
	
	err := u.client.Get().
		WithContext(ctx).
		SubResourcef(subPath).
		WithHeaders(h).
		Do().
		Into(resp)
		
	if err != nil {
		return nil, errors.CCHttpError
	}
	
	if err := resp.CCError(); err != nil {
		return nil, err
	}
	
	return resp.Data, nil
}