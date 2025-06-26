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
	"fmt"
	"math"
	"math/rand"
	"strings"
	"time"

	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/errors"
	"configcenter/src/common/http/rest"
	"configcenter/src/common/mapstr"
	"configcenter/src/common/metadata"
	"configcenter/src/storage/dal"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const (
	// TableNameUser 用户表名
	TableNameUser = "cc_user_management"
	// TableNameRolePermission 角色权限表名
	TableNameRolePermission = "cc_role_permissions"
)

// generateRandomString 生成指定长度的随机字符串
func generateRandomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	source := rand.NewSource(time.Now().UnixNano())
	r := rand.New(source)

	result := make([]byte, length)
	for i := range result {
		result[i] = charset[r.Intn(len(charset))]
	}
	return string(result)
}

// UserManagement 用户管理操作接口
type UserManagement interface {
	// 用户管理
	CreateUser(kit *rest.Kit, data *metadata.CreateUserRequest) (*metadata.User, errors.CCErrorCoder)
	UpdateUser(kit *rest.Kit, userID string, data *metadata.UpdateUserRequest) (*metadata.User, errors.CCErrorCoder)
	DeleteUser(kit *rest.Kit, userID string) errors.CCErrorCoder
	GetUser(kit *rest.Kit, userID string) (*metadata.User, errors.CCErrorCoder)
	ListUsers(kit *rest.Kit, params *metadata.UserListRequest) (*metadata.UserListResult, errors.CCErrorCoder)
	BatchDeleteUsers(kit *rest.Kit, data *metadata.BatchDeleteUsersRequest) errors.CCErrorCoder

	// 用户状态管理
	ToggleUserStatus(kit *rest.Kit, userID string, data *metadata.UserStatusRequest) (*metadata.User, errors.CCErrorCoder)
	ResetUserPassword(kit *rest.Kit, userID string) (*metadata.ResetPasswordResult, errors.CCErrorCoder)

	// 用户统计和查询
	GetUserStatistics(kit *rest.Kit) (*metadata.UserStatistics, errors.CCErrorCoder)
	ValidateEmail(kit *rest.Kit, data *metadata.ValidateEmailRequest) (*metadata.ValidateEmailResult, errors.CCErrorCoder)

	// 角色权限管理
	CreateRolePermission(kit *rest.Kit, data *metadata.CreateRolePermissionRequest) (*metadata.RolePermission, errors.CCErrorCoder)
	UpdateRolePermission(kit *rest.Kit, roleID string, data *metadata.UpdateRolePermissionRequest) (*metadata.RolePermission, errors.CCErrorCoder)
	DeleteRolePermission(kit *rest.Kit, roleID string) errors.CCErrorCoder
	GetRolePermission(kit *rest.Kit, roleID string) (*metadata.RolePermission, errors.CCErrorCoder)
	ListRolePermissions(kit *rest.Kit) ([]metadata.RolePermission, errors.CCErrorCoder)

	// 权限矩阵
	GetPermissionMatrix(kit *rest.Kit) (*metadata.PermissionMatrix, errors.CCErrorCoder)
	GetUserRoles(kit *rest.Kit, roleID string) ([]metadata.UserRoleInfo, errors.CCErrorCoder)
}

// userManagement 用户管理实现
type userManagement struct {
	db dal.RDB
}

// New 创建用户管理实例
func New(db dal.RDB) UserManagement {
	return &userManagement{
		db: db,
	}
}

// CreateUser 创建用户
func (u *userManagement) CreateUser(kit *rest.Kit, data *metadata.CreateUserRequest) (*metadata.User, errors.CCErrorCoder) {
	// 验证输入数据
	if err := u.validateCreateUserData(kit, data); err != nil {
		return nil, err
	}

	// 检查邮箱是否已存在
	if exists, err := u.emailExists(kit, data.Email); err != nil {
		return nil, err
	} else if exists {
		return nil, kit.CCError.CCErrorf(common.CCErrCommDuplicateItem, "email")
	}

	// 创建用户对象
	now := time.Now()
	user := &metadata.User{
		ID:          primitive.NewObjectID().Hex(),
		UserID:      u.generateUserID(data.Email),
		Email:       data.Email,
		Name:        data.Name,
		Role:        data.Role,
		Permissions: data.Permissions,
		Status:      data.Status,
		CreatedAt:   now,
		UpdatedAt:   now,
		CreatedBy:   kit.User,
		LoginCount:  0,
		Metadata:    data.Metadata,
	}

	// 设置默认状态
	if user.Status == "" {
		user.Status = metadata.UserStatusActive
	}

	// 插入数据库
	if err := u.db.Table(TableNameUser).Insert(kit.Ctx, user); err != nil {
		blog.Errorf("create user failed, err: %v, user: %+v, rid: %s", err, user, kit.Rid)
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBInsertFailed)
	}

	blog.Infof("create user success, user_id: %s, email: %s, rid: %s", user.UserID, user.Email, kit.Rid)
	return user, nil
}

// UpdateUser 更新用户
func (u *userManagement) UpdateUser(kit *rest.Kit, userID string, data *metadata.UpdateUserRequest) (*metadata.User, errors.CCErrorCoder) {
	// 查找用户
	_, err := u.GetUser(kit, userID)
	if err != nil {
		return nil, err
	}

	// 构建更新数据
	updateData := make(mapstr.MapStr)
	if data.Email != nil {
		// 检查邮箱是否已存在（排除当前用户）
		if exists, err := u.emailExists(kit, *data.Email); err != nil {
			return nil, err
		} else if exists {
			// 检查是否是当前用户的邮箱
			existingUser, err := u.getUserByEmail(kit, *data.Email)
			if err != nil {
				return nil, err
			}
			if existingUser.UserID != userID {
				return nil, kit.CCError.CCErrorf(common.CCErrCommDuplicateItem, "email")
			}
		}
		updateData["email"] = *data.Email
	}
	if data.Name != nil {
		updateData["name"] = *data.Name
	}
	if data.Role != nil {
		updateData["role"] = *data.Role
	}
	if data.Permissions != nil {
		updateData["permissions"] = data.Permissions
	}
	if data.Status != nil {
		updateData["status"] = *data.Status
	}
	if data.Metadata != nil {
		updateData["metadata"] = data.Metadata
	}
	updateData["updated_at"] = time.Now()

	// 更新数据库
	condition := mapstr.MapStr{"user_id": userID}
	if err := u.db.Table(TableNameUser).Update(kit.Ctx, condition, updateData); err != nil {
		blog.Errorf("update user failed, err: %v, user_id: %s, rid: %s", err, userID, kit.Rid)
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBUpdateFailed)
	}

	// 返回更新后的用户
	return u.GetUser(kit, userID)
}

// DeleteUser 删除用户
func (u *userManagement) DeleteUser(kit *rest.Kit, userID string) errors.CCErrorCoder {
	// 检查用户是否存在
	if _, err := u.GetUser(kit, userID); err != nil {
		return err
	}

	// 删除用户
	condition := mapstr.MapStr{"user_id": userID}
	if err := u.db.Table(TableNameUser).Delete(kit.Ctx, condition); err != nil {
		blog.Errorf("delete user failed, err: %v, user_id: %s, rid: %s", err, userID, kit.Rid)
		return kit.CCError.CCErrorf(common.CCErrCommDBDeleteFailed)
	}

	blog.Infof("delete user success, user_id: %s, rid: %s", userID, kit.Rid)
	return nil
}

// GetUser 获取用户详情
func (u *userManagement) GetUser(kit *rest.Kit, userID string) (*metadata.User, errors.CCErrorCoder) {
	condition := mapstr.MapStr{"user_id": userID}
	user := &metadata.User{}

	if err := u.db.Table(TableNameUser).Find(condition).One(kit.Ctx, user); err != nil {
		if u.db.IsNotFoundError(err) {
			return nil, kit.CCError.CCErrorf(common.CCErrCommNotFound, "user")
		}
		blog.Errorf("get user failed, err: %v, user_id: %s, rid: %s", err, userID, kit.Rid)
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBSelectFailed)
	}

	return user, nil
}

// ListUsers 获取用户列表
func (u *userManagement) ListUsers(kit *rest.Kit, params *metadata.UserListRequest) (*metadata.UserListResult, errors.CCErrorCoder) {
	// 构建查询条件
	condition := u.buildUserListCondition(params)

	// 计算总数
	total, err := u.db.Table(TableNameUser).Find(condition).Count(kit.Ctx)
	if err != nil {
		blog.Errorf("count users failed, err: %v, condition: %+v, rid: %s", err, condition, kit.Rid)
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBSelectFailed)
	}

	// 设置分页参数
	if params.Page <= 0 {
		params.Page = 1
	}
	if params.Limit <= 0 {
		params.Limit = 20
	}

	// 查询用户列表
	users := make([]metadata.User, 0)
	start := (params.Page - 1) * params.Limit
	finder := u.db.Table(TableNameUser).Find(condition).Start(uint64(start)).Limit(uint64(params.Limit))

	// 排序
	if params.SortField != "" {
		sortOrder := 1
		if strings.ToLower(params.SortOrder) == "desc" {
			sortOrder = -1
		}
		finder = finder.Sort(fmt.Sprintf("%s:%d", params.SortField, sortOrder))
	} else {
		finder = finder.Sort("created_at:-1")
	}

	if err := finder.All(kit.Ctx, &users); err != nil {
		blog.Errorf("list users failed, err: %v, condition: %+v, rid: %s", err, condition, kit.Rid)
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBSelectFailed)
	}

	// 计算总页数
	totalPages := int(math.Ceil(float64(total) / float64(params.Limit)))

	result := &metadata.UserListResult{
		Items:      users,
		Total:      int64(total),
		Page:       params.Page,
		Limit:      params.Limit,
		TotalPages: totalPages,
	}

	return result, nil
}

// BatchDeleteUsers 批量删除用户
func (u *userManagement) BatchDeleteUsers(kit *rest.Kit, data *metadata.BatchDeleteUsersRequest) errors.CCErrorCoder {
	if len(data.UserIDs) == 0 {
		return kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "user_ids")
	}

	// 检查用户是否存在
	condition := mapstr.MapStr{"user_id": mapstr.MapStr{common.BKDBIN: data.UserIDs}}
	count, err := u.db.Table(TableNameUser).Find(condition).Count(kit.Ctx)
	if err != nil {
		blog.Errorf("count users failed, err: %v, user_ids: %v, rid: %s", err, data.UserIDs, kit.Rid)
		return kit.CCError.CCErrorf(common.CCErrCommDBSelectFailed)
	}

	if int(count) != len(data.UserIDs) {
		return kit.CCError.CCErrorf(common.CCErrCommNotFound, "user")
	}

	// 批量删除
	if err := u.db.Table(TableNameUser).Delete(kit.Ctx, condition); err != nil {
		blog.Errorf("batch delete users failed, err: %v, user_ids: %v, rid: %s", err, data.UserIDs, kit.Rid)
		return kit.CCError.CCErrorf(common.CCErrCommDBDeleteFailed)
	}

	blog.Infof("batch delete users success, user_ids: %v, rid: %s", data.UserIDs, kit.Rid)
	return nil
}

// ToggleUserStatus 切换用户状态
func (u *userManagement) ToggleUserStatus(kit *rest.Kit, userID string, data *metadata.UserStatusRequest) (*metadata.User, errors.CCErrorCoder) {
	updateData := mapstr.MapStr{
		"status":     data.Status,
		"updated_at": time.Now(),
	}

	condition := mapstr.MapStr{"user_id": userID}
	if err := u.db.Table(TableNameUser).Update(kit.Ctx, condition, updateData); err != nil {
		blog.Errorf("toggle user status failed, err: %v, user_id: %s, status: %s, rid: %s", err, userID, data.Status, kit.Rid)
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBUpdateFailed)
	}

	return u.GetUser(kit, userID)
}

// ResetUserPassword 重置用户密码
func (u *userManagement) ResetUserPassword(kit *rest.Kit, userID string) (*metadata.ResetPasswordResult, errors.CCErrorCoder) {
	// 检查用户是否存在
	if _, err := u.GetUser(kit, userID); err != nil {
		return nil, err
	}

	// 生成临时密码
	tempPassword := generateRandomString(12)
	expiresAt := time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05")

	// 这里应该集成实际的密码重置逻辑
	// 例如发送邮件通知用户新密码

	result := &metadata.ResetPasswordResult{
		TempPassword: tempPassword,
		ExpiresAt:    expiresAt,
	}

	blog.Infof("reset user password success, user_id: %s, rid: %s", userID, kit.Rid)
	return result, nil
}

// GetUserStatistics 获取用户统计信息
func (u *userManagement) GetUserStatistics(kit *rest.Kit) (*metadata.UserStatistics, errors.CCErrorCoder) {
	// 总用户数
	total, err := u.db.Table(TableNameUser).Find(nil).Count(kit.Ctx)
	if err != nil {
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBSelectFailed)
	}

	// 按状态统计
	activeCount, _ := u.db.Table(TableNameUser).Find(mapstr.MapStr{"status": metadata.UserStatusActive}).Count(kit.Ctx)
	inactiveCount, _ := u.db.Table(TableNameUser).Find(mapstr.MapStr{"status": metadata.UserStatusInactive}).Count(kit.Ctx)
	lockedCount, _ := u.db.Table(TableNameUser).Find(mapstr.MapStr{"status": metadata.UserStatusLocked}).Count(kit.Ctx)

	// 按角色统计
	adminCount, _ := u.db.Table(TableNameUser).Find(mapstr.MapStr{"role": metadata.UserRoleAdmin}).Count(kit.Ctx)
	operatorCount, _ := u.db.Table(TableNameUser).Find(mapstr.MapStr{"role": metadata.UserRoleOperator}).Count(kit.Ctx)
	readonlyCount, _ := u.db.Table(TableNameUser).Find(mapstr.MapStr{"role": metadata.UserRoleReadonly}).Count(kit.Ctx)

	statistics := &metadata.UserStatistics{
		TotalUsers:    int64(total),
		ActiveUsers:   int64(activeCount),
		InactiveUsers: int64(inactiveCount),
		LockedUsers:   int64(lockedCount),
		AdminUsers:    int64(adminCount),
		OperatorUsers: int64(operatorCount),
		ReadonlyUsers: int64(readonlyCount),
	}

	return statistics, nil
}

// ValidateEmail 验证邮箱是否可用
func (u *userManagement) ValidateEmail(kit *rest.Kit, data *metadata.ValidateEmailRequest) (*metadata.ValidateEmailResult, errors.CCErrorCoder) {
	exists, err := u.emailExists(kit, data.Email)
	if err != nil {
		return nil, err
	}

	result := &metadata.ValidateEmailResult{
		Available: !exists,
	}

	if exists {
		result.Message = "邮箱已被使用"
	}

	return result, nil
}

// CreateRolePermission 创建角色权限
func (u *userManagement) CreateRolePermission(kit *rest.Kit, data *metadata.CreateRolePermissionRequest) (*metadata.RolePermission, errors.CCErrorCoder) {
	// 检查角色名是否已存在
	condition := mapstr.MapStr{"role_name": data.RoleName}
	count, err := u.db.Table(TableNameRolePermission).Find(condition).Count(kit.Ctx)
	if err != nil {
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBSelectFailed)
	}
	if count > 0 {
		return nil, kit.CCError.CCErrorf(common.CCErrCommDuplicateItem, "role_name")
	}

	now := time.Now()
	role := &metadata.RolePermission{
		ID:          primitive.NewObjectID().Hex(),
		RoleName:    data.RoleName,
		Permissions: data.Permissions,
		Description: data.Description,
		IsSystem:    false,
		CreatedAt:   now,
		UpdatedAt:   now,
		Metadata:    data.Metadata,
	}

	if err := u.db.Table(TableNameRolePermission).Insert(kit.Ctx, role); err != nil {
		blog.Errorf("create role permission failed, err: %v, role: %+v, rid: %s", err, role, kit.Rid)
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBInsertFailed)
	}

	return role, nil
}

// UpdateRolePermission 更新角色权限
func (u *userManagement) UpdateRolePermission(kit *rest.Kit, roleID string, data *metadata.UpdateRolePermissionRequest) (*metadata.RolePermission, errors.CCErrorCoder) {
	updateData := make(mapstr.MapStr)
	if data.Permissions != nil {
		updateData["permissions"] = data.Permissions
	}
	if data.Description != nil {
		updateData["description"] = *data.Description
	}
	if data.Metadata != nil {
		updateData["metadata"] = data.Metadata
	}
	updateData["updated_at"] = time.Now()

	condition := mapstr.MapStr{"_id": roleID}
	if err := u.db.Table(TableNameRolePermission).Update(kit.Ctx, condition, updateData); err != nil {
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBUpdateFailed)
	}

	return u.GetRolePermission(kit, roleID)
}

// DeleteRolePermission 删除角色权限
func (u *userManagement) DeleteRolePermission(kit *rest.Kit, roleID string) errors.CCErrorCoder {
	condition := mapstr.MapStr{"_id": roleID}
	if err := u.db.Table(TableNameRolePermission).Delete(kit.Ctx, condition); err != nil {
		return kit.CCError.CCErrorf(common.CCErrCommDBDeleteFailed)
	}
	return nil
}

// GetRolePermission 获取角色权限详情
func (u *userManagement) GetRolePermission(kit *rest.Kit, roleID string) (*metadata.RolePermission, errors.CCErrorCoder) {
	condition := mapstr.MapStr{"_id": roleID}
	role := &metadata.RolePermission{}

	if err := u.db.Table(TableNameRolePermission).Find(condition).One(kit.Ctx, role); err != nil {
		if u.db.IsNotFoundError(err) {
			return nil, kit.CCError.CCErrorf(common.CCErrCommNotFound, "role")
		}
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBSelectFailed)
	}

	return role, nil
}

// ListRolePermissions 获取角色权限列表
func (u *userManagement) ListRolePermissions(kit *rest.Kit) ([]metadata.RolePermission, errors.CCErrorCoder) {
	roles := make([]metadata.RolePermission, 0)
	if err := u.db.Table(TableNameRolePermission).Find(nil).All(kit.Ctx, &roles); err != nil {
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBSelectFailed)
	}
	return roles, nil
}

// GetPermissionMatrix 获取权限矩阵
func (u *userManagement) GetPermissionMatrix(kit *rest.Kit) (*metadata.PermissionMatrix, errors.CCErrorCoder) {
	// 这里应该返回系统中定义的权限矩阵
	// 现在返回一个示例结构
	matrix := &metadata.PermissionMatrix{
		Permissions: []metadata.UserPermission{
			{ID: "user.create", Name: "创建用户", Description: "允许创建新用户", Category: "用户管理"},
			{ID: "user.update", Name: "更新用户", Description: "允许更新用户信息", Category: "用户管理"},
			{ID: "user.delete", Name: "删除用户", Description: "允许删除用户", Category: "用户管理"},
			{ID: "user.view", Name: "查看用户", Description: "允许查看用户列表和详情", Category: "用户管理"},
		},
		Roles: []metadata.RoleInfo{
			{ID: "admin", Name: "管理员", Description: "系统管理员", IsSystem: true},
			{ID: "operator", Name: "操作员", Description: "系统操作员", IsSystem: true},
			{ID: "readonly", Name: "只读用户", Description: "只读权限用户", IsSystem: true},
		},
		Matrix: [][]bool{
			{true, true, true, true},    // admin
			{false, true, false, true},  // operator
			{false, false, false, true}, // readonly
		},
	}

	return matrix, nil
}

// GetUserRoles 获取角色下的用户列表
func (u *userManagement) GetUserRoles(kit *rest.Kit, roleID string) ([]metadata.UserRoleInfo, errors.CCErrorCoder) {
	condition := mapstr.MapStr{"role": roleID}
	users := make([]metadata.User, 0)

	if err := u.db.Table(TableNameUser).Find(condition).All(kit.Ctx, &users); err != nil {
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBSelectFailed)
	}

	userRoles := make([]metadata.UserRoleInfo, len(users))
	for i, user := range users {
		userRoles[i] = metadata.UserRoleInfo{
			UserID:   user.UserID,
			Email:    user.Email,
			Name:     user.Name,
			Role:     string(user.Role),
			Status:   string(user.Status),
			JoinDate: user.CreatedAt.Format("2006-01-02"),
		}
	}

	return userRoles, nil
}

// helper methods

func (u *userManagement) validateCreateUserData(kit *rest.Kit, data *metadata.CreateUserRequest) errors.CCErrorCoder {
	if data.Email == "" {
		return kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "email")
	}
	if data.Name == "" {
		return kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "name")
	}
	if data.Role == "" {
		return kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "role")
	}
	return nil
}

func (u *userManagement) emailExists(kit *rest.Kit, email string) (bool, errors.CCErrorCoder) {
	condition := mapstr.MapStr{"email": bson.M{common.BKDBLIKE: "^" + email + "$", common.BKDBOPTIONS: "i"}}
	count, err := u.db.Table(TableNameUser).Find(condition).Count(kit.Ctx)
	if err != nil {
		return false, kit.CCError.CCErrorf(common.CCErrCommDBSelectFailed)
	}
	return count > 0, nil
}

func (u *userManagement) getUserByEmail(kit *rest.Kit, email string) (*metadata.User, errors.CCErrorCoder) {
	condition := mapstr.MapStr{"email": bson.M{common.BKDBLIKE: "^" + email + "$", common.BKDBOPTIONS: "i"}}
	user := &metadata.User{}

	if err := u.db.Table(TableNameUser).Find(condition).One(kit.Ctx, user); err != nil {
		if u.db.IsNotFoundError(err) {
			return nil, kit.CCError.CCErrorf(common.CCErrCommNotFound, "user")
		}
		blog.Errorf("get user by email failed, err: %v, email: %s, rid: %s", err, email, kit.Rid)
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBSelectFailed)
	}

	return user, nil
}

func (u *userManagement) generateUserID(email string) string {
	// 简单的用户ID生成逻辑，实际项目中可能需要更复杂的算法
	return "uid-" + generateRandomString(6)
}

func (u *userManagement) buildUserListCondition(params *metadata.UserListRequest) mapstr.MapStr {
	condition := make(mapstr.MapStr)

	// 关键字搜索
	if params.Search != "" {
		condition[common.BKDBOR] = []mapstr.MapStr{
			{"email": bson.M{common.BKDBLIKE: params.Search, common.BKDBOPTIONS: "i"}},
			{"name": bson.M{common.BKDBLIKE: params.Search, common.BKDBOPTIONS: "i"}},
		}
	}

	// 角色过滤
	if params.Role != "" {
		condition["role"] = params.Role
	}

	// 状态过滤
	if params.Status != "" {
		condition["status"] = params.Status
	}

	return condition
}
