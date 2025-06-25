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

package service

import (
	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/http/rest"
	"configcenter/src/common/metadata"
)

// CreateUser 创建用户
func (s *coreService) CreateUser(ctx *rest.Contexts) {
	data := &metadata.CreateUserRequest{}
	if err := ctx.DecodeInto(data); err != nil {
		ctx.RespAutoError(err)
		return
	}

	result, err := s.core.UserManagementOperation().CreateUser(ctx.Kit, data)
	if err != nil {
		blog.Errorf("create user failed, err: %v, rid: %s", err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(result)
}

// UpdateUser 更新用户
func (s *coreService) UpdateUser(ctx *rest.Contexts) {
	userID := ctx.Request.PathParameter("user_id")
	if userID == "" {
		blog.Errorf("update user failed, user_id is empty, rid: %s", ctx.Kit.Rid)
		ctx.RespAutoError(ctx.Kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "user_id"))
		return
	}

	data := &metadata.UpdateUserRequest{}
	if err := ctx.DecodeInto(data); err != nil {
		ctx.RespAutoError(err)
		return
	}

	result, err := s.core.UserManagementOperation().UpdateUser(ctx.Kit, userID, data)
	if err != nil {
		blog.Errorf("update user failed, userID: %s, err: %v, rid: %s", userID, err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(result)
}

// DeleteUser 删除用户
func (s *coreService) DeleteUser(ctx *rest.Contexts) {
	userID := ctx.Request.PathParameter("user_id")
	if userID == "" {
		blog.Errorf("delete user failed, user_id is empty, rid: %s", ctx.Kit.Rid)
		ctx.RespAutoError(ctx.Kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "user_id"))
		return
	}

	err := s.core.UserManagementOperation().DeleteUser(ctx.Kit, userID)
	if err != nil {
		blog.Errorf("delete user failed, userID: %s, err: %v, rid: %s", userID, err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(nil)
}

// GetUser 获取用户信息
func (s *coreService) GetUser(ctx *rest.Contexts) {
	userID := ctx.Request.PathParameter("user_id")
	if userID == "" {
		blog.Errorf("get user failed, user_id is empty, rid: %s", ctx.Kit.Rid)
		ctx.RespAutoError(ctx.Kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "user_id"))
		return
	}

	result, err := s.core.UserManagementOperation().GetUser(ctx.Kit, userID)
	if err != nil {
		blog.Errorf("get user failed, userID: %s, err: %v, rid: %s", userID, err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(result)
}

// ListUsers 获取用户列表
func (s *coreService) ListUsers(ctx *rest.Contexts) {
	query := &metadata.UserListRequest{}
	if err := ctx.DecodeInto(query); err != nil {
		ctx.RespAutoError(err)
		return
	}

	result, err := s.core.UserManagementOperation().ListUsers(ctx.Kit, query)
	if err != nil {
		blog.Errorf("list users failed, err: %v, rid: %s", err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(result)
}

// BatchDeleteUsers 批量删除用户
func (s *coreService) BatchDeleteUsers(ctx *rest.Contexts) {
	data := &metadata.BatchDeleteUsersRequest{}
	if err := ctx.DecodeInto(data); err != nil {
		ctx.RespAutoError(err)
		return
	}

	err := s.core.UserManagementOperation().BatchDeleteUsers(ctx.Kit, data)
	if err != nil {
		blog.Errorf("batch delete users failed, err: %v, rid: %s", err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(nil)
}

// ToggleUserStatus 切换用户状态
func (s *coreService) ToggleUserStatus(ctx *rest.Contexts) {
	userID := ctx.Request.PathParameter("user_id")
	if userID == "" {
		blog.Errorf("toggle user status failed, user_id is empty, rid: %s", ctx.Kit.Rid)
		ctx.RespAutoError(ctx.Kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "user_id"))
		return
	}

	data := &metadata.UserStatusRequest{}
	if err := ctx.DecodeInto(data); err != nil {
		ctx.RespAutoError(err)
		return
	}

	result, err := s.core.UserManagementOperation().ToggleUserStatus(ctx.Kit, userID, data)
	if err != nil {
		blog.Errorf("toggle user status failed, userID: %s, err: %v, rid: %s", userID, err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(result)
}

// ResetUserPassword 重置用户密码
func (s *coreService) ResetUserPassword(ctx *rest.Contexts) {
	userID := ctx.Request.PathParameter("user_id")
	if userID == "" {
		blog.Errorf("reset user password failed, user_id is empty, rid: %s", ctx.Kit.Rid)
		ctx.RespAutoError(ctx.Kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "user_id"))
		return
	}

	result, err := s.core.UserManagementOperation().ResetUserPassword(ctx.Kit, userID)
	if err != nil {
		blog.Errorf("reset user password failed, userID: %s, err: %v, rid: %s", userID, err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(result)
}

// GetUserStatistics 获取用户统计信息
func (s *coreService) GetUserStatistics(ctx *rest.Contexts) {
	result, err := s.core.UserManagementOperation().GetUserStatistics(ctx.Kit)
	if err != nil {
		blog.Errorf("get user statistics failed, err: %v, rid: %s", err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(result)
}

// ValidateEmail 验证邮箱
func (s *coreService) ValidateEmail(ctx *rest.Contexts) {
	data := &metadata.ValidateEmailRequest{}
	if err := ctx.DecodeInto(data); err != nil {
		ctx.RespAutoError(err)
		return
	}

	result, err := s.core.UserManagementOperation().ValidateEmail(ctx.Kit, data)
	if err != nil {
		blog.Errorf("validate email failed, err: %v, rid: %s", err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(result)
}

// CreateRolePermission 创建角色权限
func (s *coreService) CreateRolePermission(ctx *rest.Contexts) {
	data := &metadata.CreateRolePermissionRequest{}
	if err := ctx.DecodeInto(data); err != nil {
		ctx.RespAutoError(err)
		return
	}

	result, err := s.core.UserManagementOperation().CreateRolePermission(ctx.Kit, data)
	if err != nil {
		blog.Errorf("create role permission failed, err: %v, rid: %s", err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(result)
}

// UpdateRolePermission 更新角色权限
func (s *coreService) UpdateRolePermission(ctx *rest.Contexts) {
	roleID := ctx.Request.PathParameter("role_id")
	if roleID == "" {
		blog.Errorf("update role permission failed, role_id is empty, rid: %s", ctx.Kit.Rid)
		ctx.RespAutoError(ctx.Kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "role_id"))
		return
	}

	data := &metadata.UpdateRolePermissionRequest{}
	if err := ctx.DecodeInto(data); err != nil {
		ctx.RespAutoError(err)
		return
	}

	result, err := s.core.UserManagementOperation().UpdateRolePermission(ctx.Kit, roleID, data)
	if err != nil {
		blog.Errorf("update role permission failed, roleID: %s, err: %v, rid: %s", roleID, err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(result)
}

// DeleteRolePermission 删除角色权限
func (s *coreService) DeleteRolePermission(ctx *rest.Contexts) {
	roleID := ctx.Request.PathParameter("role_id")
	if roleID == "" {
		blog.Errorf("delete role permission failed, role_id is empty, rid: %s", ctx.Kit.Rid)
		ctx.RespAutoError(ctx.Kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "role_id"))
		return
	}

	err := s.core.UserManagementOperation().DeleteRolePermission(ctx.Kit, roleID)
	if err != nil {
		blog.Errorf("delete role permission failed, roleID: %s, err: %v, rid: %s", roleID, err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(nil)
}

// GetPermissionMatrix 获取权限矩阵
func (s *coreService) GetPermissionMatrix(ctx *rest.Contexts) {
	result, err := s.core.UserManagementOperation().GetPermissionMatrix(ctx.Kit)
	if err != nil {
		blog.Errorf("get permission matrix failed, err: %v, rid: %s", err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(result)
}

// ExportUsers 导出用户
func (s *coreService) ExportUsers(ctx *rest.Contexts) {
	// TODO: 实现用户导出功能
	ctx.RespEntity(map[string]interface{}{
		"message": "Export users functionality not implemented yet",
	})
}

// ImportUsers 导入用户
func (s *coreService) ImportUsers(ctx *rest.Contexts) {
	// TODO: 实现用户导入功能
	ctx.RespEntity(map[string]interface{}{
		"message": "Import users functionality not implemented yet",
	})
}

// GetRolePermission 获取角色权限
func (s *coreService) GetRolePermission(ctx *rest.Contexts) {
	roleID := ctx.Request.PathParameter("role_id")
	if roleID == "" {
		blog.Errorf("get role permission failed, role_id is empty, rid: %s", ctx.Kit.Rid)
		ctx.RespAutoError(ctx.Kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "role_id"))
		return
	}

	result, err := s.core.UserManagementOperation().GetRolePermission(ctx.Kit, roleID)
	if err != nil {
		blog.Errorf("get role permission failed, roleID: %s, err: %v, rid: %s", roleID, err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(result)
}

// ListRolePermissions 获取角色权限列表
func (s *coreService) ListRolePermissions(ctx *rest.Contexts) {
	result, err := s.core.UserManagementOperation().ListRolePermissions(ctx.Kit)
	if err != nil {
		blog.Errorf("list role permissions failed, err: %v, rid: %s", err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(result)
}

// GetUserRoles 获取用户角色信息
func (s *coreService) GetUserRoles(ctx *rest.Contexts) {
	roleID := ctx.Request.PathParameter("role_id")
	if roleID == "" {
		blog.Errorf("get user roles failed, role_id is empty, rid: %s", ctx.Kit.Rid)
		ctx.RespAutoError(ctx.Kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "role_id"))
		return
	}

	result, err := s.core.UserManagementOperation().GetUserRoles(ctx.Kit, roleID)
	if err != nil {
		blog.Errorf("get user roles failed, roleID: %s, err: %v, rid: %s", roleID, err, ctx.Kit.Rid)
		ctx.RespAutoError(err)
		return
	}

	ctx.RespEntity(result)
}