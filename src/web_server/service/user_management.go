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
	"net/http"

	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/http/rest"
	"configcenter/src/common/metadata"

	"github.com/gin-gonic/gin"
)

// InitUserManagement 初始化用户管理路由
func (s *Service) InitUserManagement(web *gin.Engine) {
	userGroup := web.Group("/api/v3/usermgmt")
	{
		// 用户管理
		userGroup.GET("/list", s.getUserList)
		userGroup.GET("/:user_id", s.getUserDetail)
		userGroup.POST("/create", s.createUser)
		userGroup.PUT("/:user_id", s.updateUser)
		userGroup.DELETE("/:user_id", s.deleteUser)
		userGroup.DELETE("/batch", s.batchDeleteUsers)
		
		// 用户状态管理
		userGroup.PATCH("/:user_id/status", s.toggleUserStatus)
		userGroup.PUT("/:user_id/disable", s.disableUser)
		userGroup.PUT("/:user_id/enable", s.enableUser)
		userGroup.POST("/:user_id/reset-password", s.resetUserPassword)
		
		// 用户统计和查询
		userGroup.GET("/statistics", s.getUserStatistics)
		userGroup.POST("/validate-email", s.validateEmail)
		
		// 用户导入导出
		userGroup.GET("/export", s.exportUsers)
		userGroup.POST("/import", s.importUsers)
	}
	
	// 角色权限管理
	roleGroup := web.Group("/api/v3/role")
	{
		roleGroup.GET("/list", s.listRolePermissions)
		roleGroup.GET("/:role_id", s.getRolePermission)
		roleGroup.POST("/create", s.createRolePermission)
		roleGroup.PUT("/:role_id", s.updateRolePermission)
		roleGroup.DELETE("/:role_id", s.deleteRolePermission)
		roleGroup.GET("/:role_id/users", s.getUserRoles)
		roleGroup.GET("/permission-matrix", s.getPermissionMatrix)
	}
}

// createUser 创建用户
func (s *Service) createUser(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	
	// 解析请求参数
	data := &metadata.CreateUserRequest{}
	if err := c.ShouldBindJSON(data); err != nil {
		blog.Errorf("create user failed, parse request body failed, err: %v, rid: %s", err, kit.Rid)
		c.JSON(http.StatusBadRequest, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommJSONUnmarshalFailed,
			ErrMsg: s.CCErr.Error("en", common.CCErrCommJSONUnmarshalFailed).Error(),
		})
		return
	}
	
	// 调用核心服务
	user, err := s.CoreAPI.CoreService().UserManagement().CreateUser(kit.Ctx, c.Request.Header, data)
	if err != nil {
		blog.Errorf("create user failed, err: %v, data: %+v, rid: %s", err, data, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.CreateUserResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.CreateUserResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: user,
	})
}

// updateUser 更新用户
func (s *Service) updateUser(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	userID := c.Param("user_id")
	
	// 解析请求参数
	data := &metadata.UpdateUserRequest{}
	if err := c.ShouldBindJSON(data); err != nil {
		blog.Errorf("update user failed, parse request body failed, err: %v, rid: %s", err, kit.Rid)
		c.JSON(http.StatusBadRequest, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommJSONUnmarshalFailed,
			ErrMsg: s.CCErr.Error("en", common.CCErrCommJSONUnmarshalFailed).Error(),
		})
		return
	}
	
	// 调用核心服务
	user, err := s.CoreAPI.CoreService().UserManagement().UpdateUser(kit.Ctx, c.Request.Header, userID, data)
	if err != nil {
		blog.Errorf("update user failed, err: %v, user_id: %s, data: %+v, rid: %s", err, userID, data, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.UpdateUserResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.UpdateUserResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: user,
	})
}

// deleteUser 删除用户
func (s *Service) deleteUser(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	userID := c.Param("user_id")
	
	// 调用核心服务
	err := s.CoreAPI.CoreService().UserManagement().DeleteUser(kit.Ctx, c.Request.Header, userID)
	if err != nil {
		blog.Errorf("delete user failed, err: %v, user_id: %s, rid: %s", err, userID, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommHTTPDoRequestFailed,
			ErrMsg: err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.BaseResp{
		Result: true,
		Code:   0,
		ErrMsg: "",
	})
}

// getUserDetail 获取用户详情
func (s *Service) getUserDetail(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	userID := c.Param("user_id")
	
	// 调用核心服务
	user, err := s.CoreAPI.CoreService().UserManagement().GetUser(kit.Ctx, c.Request.Header, userID)
	if err != nil {
		blog.Errorf("get user detail failed, err: %v, user_id: %s, rid: %s", err, userID, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.UserDetailResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.UserDetailResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: user,
	})
}

// getUserList 获取用户列表
func (s *Service) getUserList(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	
	// 解析查询参数
	params := &metadata.UserListRequest{}
	if err := c.ShouldBindQuery(params); err != nil {
		blog.Errorf("get user list failed, parse query params failed, err: %v, rid: %s", err, kit.Rid)
		c.JSON(http.StatusBadRequest, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommParamsInvalid,
			ErrMsg: s.CCErr.Error("en", common.CCErrCommParamsInvalid).Error(),
		})
		return
	}
	
	// 调用核心服务
	result, err := s.CoreAPI.CoreService().UserManagement().ListUsers(kit.Ctx, c.Request.Header, params)
	if err != nil {
		blog.Errorf("get user list failed, err: %v, params: %+v, rid: %s", err, params, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.UserListResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.UserListResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: result,
	})
}

// batchDeleteUsers 批量删除用户
func (s *Service) batchDeleteUsers(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	
	// 解析请求参数
	data := &metadata.BatchDeleteUsersRequest{}
	if err := c.ShouldBindJSON(data); err != nil {
		blog.Errorf("batch delete users failed, parse request body failed, err: %v, rid: %s", err, kit.Rid)
		c.JSON(http.StatusBadRequest, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommJSONUnmarshalFailed,
			ErrMsg: s.CCErr.Error("en", common.CCErrCommJSONUnmarshalFailed).Error(),
		})
		return
	}
	
	// 调用核心服务
	err := s.CoreAPI.CoreService().UserManagement().BatchDeleteUsers(kit.Ctx, c.Request.Header, data)
	if err != nil {
		blog.Errorf("batch delete users failed, err: %v, data: %+v, rid: %s", err, data, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommHTTPDoRequestFailed,
			ErrMsg: err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.BaseResp{
		Result: true,
		Code:   0,
		ErrMsg: "",
	})
}

// toggleUserStatus 切换用户状态
func (s *Service) toggleUserStatus(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	userID := c.Param("user_id")
	
	// 解析请求参数
	data := &metadata.UserStatusRequest{}
	if err := c.ShouldBindJSON(data); err != nil {
		blog.Errorf("toggle user status failed, parse request body failed, err: %v, rid: %s", err, kit.Rid)
		c.JSON(http.StatusBadRequest, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommJSONUnmarshalFailed,
			ErrMsg: s.CCErr.Error("en", common.CCErrCommJSONUnmarshalFailed).Error(),
		})
		return
	}
	
	// 调用核心服务
	user, err := s.CoreAPI.CoreService().UserManagement().ToggleUserStatus(kit.Ctx, c.Request.Header, userID, data)
	if err != nil {
		blog.Errorf("toggle user status failed, err: %v, user_id: %s, data: %+v, rid: %s", err, userID, data, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.UpdateUserResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.UpdateUserResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: user,
	})
}

// resetUserPassword 重置用户密码
func (s *Service) resetUserPassword(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	userID := c.Param("user_id")
	
	// 调用核心服务
	result, err := s.CoreAPI.CoreService().UserManagement().ResetUserPassword(kit.Ctx, c.Request.Header, userID)
	if err != nil {
		blog.Errorf("reset user password failed, err: %v, user_id: %s, rid: %s", err, userID, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.ResetPasswordResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.ResetPasswordResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: result,
	})
}

// getUserStatistics 获取用户统计信息
func (s *Service) getUserStatistics(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	
	// 调用核心服务
	statistics, err := s.CoreAPI.CoreService().UserManagement().GetUserStatistics(kit.Ctx, c.Request.Header)
	if err != nil {
		blog.Errorf("get user statistics failed, err: %v, rid: %s", err, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.UserStatisticsResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.UserStatisticsResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: statistics,
	})
}

// validateEmail 验证邮箱是否可用
func (s *Service) validateEmail(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	
	// 解析请求参数
	data := &metadata.ValidateEmailRequest{}
	if err := c.ShouldBindJSON(data); err != nil {
		blog.Errorf("validate email failed, parse request body failed, err: %v, rid: %s", err, kit.Rid)
		c.JSON(http.StatusBadRequest, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommJSONUnmarshalFailed,
			ErrMsg: s.CCErr.Error("en", common.CCErrCommJSONUnmarshalFailed).Error(),
		})
		return
	}
	
	// 调用核心服务
	result, err := s.CoreAPI.CoreService().UserManagement().ValidateEmail(kit.Ctx, c.Request.Header, data)
	if err != nil {
		blog.Errorf("validate email failed, err: %v, data: %+v, rid: %s", err, data, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.ValidateEmailResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.ValidateEmailResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: result,
	})
}

// exportUsers 导出用户列表
func (s *Service) exportUsers(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	
	// 解析查询参数
	params := &metadata.UserExportRequest{}
	if err := c.ShouldBindQuery(params); err != nil {
		blog.Errorf("export users failed, parse query params failed, err: %v, rid: %s", err, kit.Rid)
		c.JSON(http.StatusBadRequest, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommParamsInvalid,
			ErrMsg: s.CCErr.Error("en", common.CCErrCommParamsInvalid).Error(),
		})
		return
	}
	
	// 调用核心服务
	data, err := s.CoreAPI.CoreService().UserManagement().ExportUsers(kit.Ctx, c.Request.Header, params)
	if err != nil {
		blog.Errorf("export users failed, err: %v, params: %+v, rid: %s", err, params, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommHTTPDoRequestFailed,
			ErrMsg: err.Error(),
		})
		return
	}
	
	// 设置响应头
	filename := "users_export.xlsx"
	if params.Format == "csv" {
		filename = "users_export.csv"
	}
	
	c.Header("Content-Type", "application/octet-stream")
	c.Header("Content-Disposition", "attachment; filename="+filename)
	c.Data(http.StatusOK, "application/octet-stream", data)
}

// importUsers 导入用户
func (s *Service) importUsers(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	
	// 解析文件上传
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		blog.Errorf("import users failed, get file failed, err: %v, rid: %s", err, kit.Rid)
		c.JSON(http.StatusBadRequest, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommParamsInvalid,
			ErrMsg: "文件上传失败",
		})
		return
	}
	defer file.Close()
	
	// 读取文件内容
	fileData := make([]byte, header.Size)
	if _, err := file.Read(fileData); err != nil {
		blog.Errorf("import users failed, read file failed, err: %v, rid: %s", err, kit.Rid)
		c.JSON(http.StatusBadRequest, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommParamsInvalid,
			ErrMsg: "文件读取失败",
		})
		return
	}
	
	// 构建请求数据
	data := &metadata.UserImportRequest{
		File:     fileData,
		FileName: header.Filename,
		Format:   c.DefaultPostForm("format", "excel"),
	}
	
	// 调用核心服务
	result, err := s.CoreAPI.CoreService().UserManagement().ImportUsers(kit.Ctx, c.Request.Header, data)
	if err != nil {
		blog.Errorf("import users failed, err: %v, filename: %s, rid: %s", err, header.Filename, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.UserImportResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.UserImportResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: result,
	})
}

// 角色权限管理相关接口

// createRolePermission 创建角色权限
func (s *Service) createRolePermission(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	
	// 解析请求参数
	data := &metadata.CreateRolePermissionRequest{}
	if err := c.ShouldBindJSON(data); err != nil {
		blog.Errorf("create role permission failed, parse request body failed, err: %v, rid: %s", err, kit.Rid)
		c.JSON(http.StatusBadRequest, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommJSONUnmarshalFailed,
			ErrMsg: s.CCErr.Error("en", common.CCErrCommJSONUnmarshalFailed).Error(),
		})
		return
	}
	
	// 调用核心服务
	role, err := s.CoreAPI.CoreService().UserManagement().CreateRolePermission(kit.Ctx, c.Request.Header, data)
	if err != nil {
		blog.Errorf("create role permission failed, err: %v, data: %+v, rid: %s", err, data, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.RolePermissionResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.RolePermissionResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: role,
	})
}

// updateRolePermission 更新角色权限
func (s *Service) updateRolePermission(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	roleID := c.Param("role_id")
	
	// 解析请求参数
	data := &metadata.UpdateRolePermissionRequest{}
	if err := c.ShouldBindJSON(data); err != nil {
		blog.Errorf("update role permission failed, parse request body failed, err: %v, rid: %s", err, kit.Rid)
		c.JSON(http.StatusBadRequest, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommJSONUnmarshalFailed,
			ErrMsg: s.CCErr.Error("en", common.CCErrCommJSONUnmarshalFailed).Error(),
		})
		return
	}
	
	// 调用核心服务
	role, err := s.CoreAPI.CoreService().UserManagement().UpdateRolePermission(kit.Ctx, c.Request.Header, roleID, data)
	if err != nil {
		blog.Errorf("update role permission failed, err: %v, role_id: %s, data: %+v, rid: %s", err, roleID, data, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.RolePermissionResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.RolePermissionResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: role,
	})
}

// deleteRolePermission 删除角色权限
func (s *Service) deleteRolePermission(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	roleID := c.Param("role_id")
	
	// 调用核心服务
	err := s.CoreAPI.CoreService().UserManagement().DeleteRolePermission(kit.Ctx, c.Request.Header, roleID)
	if err != nil {
		blog.Errorf("delete role permission failed, err: %v, role_id: %s, rid: %s", err, roleID, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.BaseResp{
			Result: false,
			Code:   common.CCErrCommHTTPDoRequestFailed,
			ErrMsg: err.Error(),
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.BaseResp{
		Result: true,
		Code:   0,
		ErrMsg: "",
	})
}

// getRolePermission 获取角色权限详情
func (s *Service) getRolePermission(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	roleID := c.Param("role_id")
	
	// 调用核心服务
	role, err := s.CoreAPI.CoreService().UserManagement().GetRolePermission(kit.Ctx, c.Request.Header, roleID)
	if err != nil {
		blog.Errorf("get role permission failed, err: %v, role_id: %s, rid: %s", err, roleID, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.RolePermissionResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.RolePermissionResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: role,
	})
}

// listRolePermissions 获取角色权限列表
func (s *Service) listRolePermissions(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	
	// 调用核心服务
	roles, err := s.CoreAPI.CoreService().UserManagement().ListRolePermissions(kit.Ctx, c.Request.Header)
	if err != nil {
		blog.Errorf("list role permissions failed, err: %v, rid: %s", err, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.RolePermissionListResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.RolePermissionListResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: roles,
	})
}

// getPermissionMatrix 获取权限矩阵
func (s *Service) getPermissionMatrix(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	
	// 调用核心服务
	matrix, err := s.CoreAPI.CoreService().UserManagement().GetPermissionMatrix(kit.Ctx, c.Request.Header)
	if err != nil {
		blog.Errorf("get permission matrix failed, err: %v, rid: %s", err, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.PermissionMatrixResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.PermissionMatrixResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: matrix,
	})
}

// getUserRoles 获取角色下的用户列表
func (s *Service) getUserRoles(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	roleID := c.Param("role_id")
	
	// 调用核心服务
	userRoles, err := s.CoreAPI.CoreService().UserManagement().GetUserRoles(kit.Ctx, c.Request.Header, roleID)
	if err != nil {
		blog.Errorf("get user roles failed, err: %v, role_id: %s, rid: %s", err, roleID, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.GetUserRolesResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.GetUserRolesResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: userRoles,
	})
}

// disableUser 禁用用户
func (s *Service) disableUser(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	userID := c.Param("user_id")
	
	// 构建禁用请求数据
	data := &metadata.UserStatusRequest{
		Status: metadata.UserStatusInactive,
	}
	
	// 调用核心服务
	user, err := s.CoreAPI.CoreService().UserManagement().ToggleUserStatus(kit.Ctx, c.Request.Header, userID, data)
	if err != nil {
		blog.Errorf("disable user failed, err: %v, user_id: %s, rid: %s", err, userID, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.UpdateUserResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.UpdateUserResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: user,
	})
}

// enableUser 启用用户
func (s *Service) enableUser(c *gin.Context) {
	kit := rest.NewKitFromHeader(c.Request.Header, s.CCErr)
	userID := c.Param("user_id")
	
	// 构建启用请求数据
	data := &metadata.UserStatusRequest{
		Status: metadata.UserStatusActive,
	}
	
	// 调用核心服务
	user, err := s.CoreAPI.CoreService().UserManagement().ToggleUserStatus(kit.Ctx, c.Request.Header, userID, data)
	if err != nil {
		blog.Errorf("enable user failed, err: %v, user_id: %s, rid: %s", err, userID, kit.Rid)
		c.JSON(http.StatusInternalServerError, metadata.UpdateUserResponse{
			BaseResp: metadata.BaseResp{
				Result: false,
				Code:   common.CCErrCommHTTPDoRequestFailed,
				ErrMsg: err.Error(),
			},
		})
		return
	}
	
	c.JSON(http.StatusOK, metadata.UpdateUserResponse{
		BaseResp: metadata.BaseResp{
			Result: true,
			Code:   0,
			ErrMsg: "",
		},
		Data: user,
	})
}