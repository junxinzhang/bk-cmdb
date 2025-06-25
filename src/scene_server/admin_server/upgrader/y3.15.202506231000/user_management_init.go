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

package y3_15_202506231000

import (
	"context"
	"time"

	"configcenter/src/common"
	"configcenter/src/common/metadata"
	"configcenter/src/scene_server/admin_server/upgrader"
	"configcenter/src/storage/dal"
	"configcenter/src/storage/dal/types"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"gopkg.in/mgo.v2"
)

const (
	// 用户管理表名
	TableNameUserManagement  = "cc_user_management"
	TableNameRolePermissions = "cc_role_permissions"
)

// initUserManagementTables 初始化用户管理表和索引
func initUserManagementTables(ctx context.Context, db dal.RDB, conf *upgrader.Config) error {
	// 创建用户管理表
	for tableName, indexes := range userManagementTables {
		exists, err := db.HasTable(ctx, tableName)
		if err != nil {
			return err
		}
		if !exists {
			if err = db.CreateTable(ctx, tableName); err != nil && !mgo.IsDup(err) {
				return err
			}
		}

		// 创建索引
		for _, index := range indexes {
			if err = db.Table(tableName).CreateIndex(ctx, index); err != nil && !db.IsDuplicatedError(err) {
				return err
			}
		}
	}

	// 初始化默认角色权限数据
	if err := initDefaultRolePermissions(ctx, db); err != nil {
		return err
	}

	// 初始化默认管理员用户
	if err := initDefaultAdminUser(ctx, db); err != nil {
		return err
	}

	return nil
}

// userManagementTables 用户管理相关表的索引定义
var userManagementTables = map[string][]types.Index{
	TableNameUserManagement: {
		// 主键索引
		types.Index{
			Name:       "idx_id",
			Keys:       bson.D{{common.BKFieldID, 1}},
			Background: true,
			Unique:     true,
		},
		// 用户ID唯一索引
		types.Index{
			Name:       "idx_user_id",
			Keys:       bson.D{{"user_id", 1}},
			Background: true,
			Unique:     true,
		},
		// 邮箱唯一索引
		types.Index{
			Name:       "idx_email",
			Keys:       bson.D{{"email", 1}},
			Background: true,
			Unique:     true,
		},
		// 状态+角色复合索引（用于筛选查询）
		types.Index{
			Name:       "idx_status_role",
			Keys:       bson.D{{"status", 1}, {"role", 1}},
			Background: true,
			Unique:     false,
		},
		// 创建时间索引（用于排序）
		types.Index{
			Name:       "idx_created_at",
			Keys:       bson.D{{"created_at", -1}},
			Background: true,
			Unique:     false,
		},
		// 更新时间索引
		types.Index{
			Name:       "idx_updated_at",
			Keys:       bson.D{{"updated_at", -1}},
			Background: true,
			Unique:     false,
		},
	},
	TableNameRolePermissions: {
		// 主键索引
		types.Index{
			Name:       "idx_id",
			Keys:       bson.D{{common.BKFieldID, 1}},
			Background: true,
			Unique:     true,
		},
		// 角色名唯一索引
		types.Index{
			Name:       "idx_role_name",
			Keys:       bson.D{{"role_name", 1}},
			Background: true,
			Unique:     true,
		},
		// 系统角色索引
		types.Index{
			Name:       "idx_is_system",
			Keys:       bson.D{{"is_system", 1}},
			Background: true,
			Unique:     false,
		},
		// 创建时间索引
		types.Index{
			Name:       "idx_created_at",
			Keys:       bson.D{{"created_at", -1}},
			Background: true,
			Unique:     false,
		},
	},
}

// initDefaultRolePermissions 初始化默认角色权限
func initDefaultRolePermissions(ctx context.Context, db dal.RDB) error {
	// 检查是否已存在默认角色
	count, err := db.Table(TableNameRolePermissions).Find(bson.M{"is_system": true}).Count(ctx)
	if err != nil {
		return err
	}

	// 如果已存在系统角色，跳过初始化
	if count > 0 {
		return nil
	}

	now := time.Now()
	defaultRoles := []metadata.RolePermission{
		{
			ID:       primitive.NewObjectID().Hex(),
			RoleName: "admin",
			Permissions: []string{
				"user.create", "user.update", "user.delete", "user.view",
				"role.create", "role.update", "role.delete", "role.view",
				"config.admin", "operation.admin",
			},
			Description: "系统管理员，拥有所有权限",
			IsSystem:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:       primitive.NewObjectID().Hex(),
			RoleName: "operator",
			Permissions: []string{
				"user.update", "user.view",
				"role.view",
				"operation.execute",
			},
			Description: "系统操作员，拥有执行权限",
			IsSystem:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
		{
			ID:       primitive.NewObjectID().Hex(),
			RoleName: "readonly",
			Permissions: []string{
				"user.view",
				"role.view",
				"operation.view",
			},
			Description: "只读用户，仅可查看信息",
			IsSystem:    true,
			CreatedAt:   now,
			UpdatedAt:   now,
		},
	}

	// 批量插入默认角色
	docs := make([]interface{}, len(defaultRoles))
	for i, role := range defaultRoles {
		docs[i] = role
	}

	return db.Table(TableNameRolePermissions).Insert(ctx, docs)
}

// initDefaultAdminUser 初始化默认管理员用户
func initDefaultAdminUser(ctx context.Context, db dal.RDB) error {
	// 检查是否已存在admin用户
	count, err := db.Table(TableNameUserManagement).Find(bson.M{"user_id": "admin"}).Count(ctx)
	if err != nil {
		return err
	}

	// 如果已存在admin用户，跳过初始化
	if count > 0 {
		return nil
	}

	now := time.Now()
	adminUser := metadata.User{
		ID:     primitive.NewObjectID().Hex(),
		UserID: "admin",
		Email:  "admin@local.com",
		Name:   "系统管理员",
		Role:   metadata.UserRoleAdmin,
		Permissions: []string{
			"user.create", "user.update", "user.delete", "user.view",
			"role.create", "role.update", "role.delete", "role.view",
			"config.admin", "operation.admin",
		},
		Status:     metadata.UserStatusActive,
		CreatedAt:  now,
		UpdatedAt:  now,
		CreatedBy:  "system",
		LoginCount: 0,
	}

	return db.Table(TableNameUserManagement).Insert(ctx, adminUser)
}

