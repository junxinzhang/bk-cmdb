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
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"configcenter/src/common/metadata"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

// MockCoreAPI mocks the CoreAPI for testing
type MockCoreAPI struct {
	mock.Mock
}

type MockUserManagement struct {
	mock.Mock
}

func (m *MockUserManagement) ListUsers(ctx interface{}, header http.Header, params *metadata.UserListRequest) (*metadata.UserListResult, error) {
	args := m.Called(ctx, header, params)
	return args.Get(0).(*metadata.UserListResult), args.Error(1)
}

func (m *MockUserManagement) GetUser(ctx interface{}, header http.Header, userID string) (*metadata.User, error) {
	args := m.Called(ctx, header, userID)
	return args.Get(0).(*metadata.User), args.Error(1)
}

func (m *MockUserManagement) CreateUser(ctx interface{}, header http.Header, data *metadata.CreateUserRequest) (*metadata.User, error) {
	args := m.Called(ctx, header, data)
	return args.Get(0).(*metadata.User), args.Error(1)
}

func (m *MockUserManagement) UpdateUser(ctx interface{}, header http.Header, userID string, data *metadata.UpdateUserRequest) (*metadata.User, error) {
	args := m.Called(ctx, header, userID, data)
	return args.Get(0).(*metadata.User), args.Error(1)
}

func (m *MockUserManagement) DeleteUser(ctx interface{}, header http.Header, userID string) error {
	args := m.Called(ctx, header, userID)
	return args.Error(0)
}

func (m *MockUserManagement) ToggleUserStatus(ctx interface{}, header http.Header, userID string, data *metadata.UserStatusRequest) (*metadata.User, error) {
	args := m.Called(ctx, header, userID, data)
	return args.Get(0).(*metadata.User), args.Error(1)
}

func (m *MockUserManagement) BatchDeleteUsers(ctx interface{}, header http.Header, data *metadata.BatchDeleteUsersRequest) error {
	args := m.Called(ctx, header, data)
	return args.Error(0)
}

func (m *MockUserManagement) ResetUserPassword(ctx interface{}, header http.Header, userID string) (*metadata.ResetPasswordResult, error) {
	args := m.Called(ctx, header, userID)
	return args.Get(0).(*metadata.ResetPasswordResult), args.Error(1)
}

func (m *MockUserManagement) GetUserStatistics(ctx interface{}, header http.Header) (*metadata.UserStatistics, error) {
	args := m.Called(ctx, header)
	return args.Get(0).(*metadata.UserStatistics), args.Error(1)
}

func (m *MockUserManagement) ValidateEmail(ctx interface{}, header http.Header, data *metadata.ValidateEmailRequest) (*metadata.ValidateEmailResult, error) {
	args := m.Called(ctx, header, data)
	return args.Get(0).(*metadata.ValidateEmailResult), args.Error(1)
}

func (m *MockUserManagement) ExportUsers(ctx interface{}, header http.Header, params *metadata.UserExportRequest) ([]byte, error) {
	args := m.Called(ctx, header, params)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockUserManagement) ImportUsers(ctx interface{}, header http.Header, data *metadata.UserImportRequest) (*metadata.UserImportResult, error) {
	args := m.Called(ctx, header, data)
	return args.Get(0).(*metadata.UserImportResult), args.Error(1)
}

func (m *MockUserManagement) CreateRolePermission(ctx interface{}, header http.Header, data *metadata.CreateRolePermissionRequest) (*metadata.RolePermission, error) {
	args := m.Called(ctx, header, data)
	return args.Get(0).(*metadata.RolePermission), args.Error(1)
}

func (m *MockUserManagement) UpdateRolePermission(ctx interface{}, header http.Header, roleID string, data *metadata.UpdateRolePermissionRequest) (*metadata.RolePermission, error) {
	args := m.Called(ctx, header, roleID, data)
	return args.Get(0).(*metadata.RolePermission), args.Error(1)
}

func (m *MockUserManagement) DeleteRolePermission(ctx interface{}, header http.Header, roleID string) error {
	args := m.Called(ctx, header, roleID)
	return args.Error(0)
}

func (m *MockUserManagement) GetRolePermission(ctx interface{}, header http.Header, roleID string) (*metadata.RolePermission, error) {
	args := m.Called(ctx, header, roleID)
	return args.Get(0).(*metadata.RolePermission), args.Error(1)
}

func (m *MockUserManagement) ListRolePermissions(ctx interface{}, header http.Header) ([]*metadata.RolePermission, error) {
	args := m.Called(ctx, header)
	return args.Get(0).([]*metadata.RolePermission), args.Error(1)
}

func (m *MockUserManagement) GetPermissionMatrix(ctx interface{}, header http.Header) (*metadata.PermissionMatrix, error) {
	args := m.Called(ctx, header)
	return args.Get(0).(*metadata.PermissionMatrix), args.Error(1)
}

func (m *MockUserManagement) GetUserRoles(ctx interface{}, header http.Header, roleID string) ([]*metadata.UserRole, error) {
	args := m.Called(ctx, header, roleID)
	return args.Get(0).([]*metadata.UserRole), args.Error(1)
}

func (m *MockCoreAPI) UserManagement() MockUserManagement {
	args := m.Called()
	return args.Get(0).(MockUserManagement)
}

// setupTestService 创建测试用的Service实例
func setupTestService() (*Service, *MockCoreAPI, *MockUserManagement) {
	mockCoreAPI := &MockCoreAPI{}
	mockUserMgmt := &MockUserManagement{}
	
	// 设置mockCoreAPI返回mockUserMgmt
	mockCoreAPI.On("UserManagement").Return(*mockUserMgmt)
	
	service := &Service{
		CoreAPI: mockCoreAPI,
		CCErr:   util.NewCCErrorIf(),
	}
	
	return service, mockCoreAPI, mockUserMgmt
}

// TestGetUserList 测试获取用户列表接口
func TestGetUserList(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		queryParams    string
		expectedParams *metadata.UserListRequest
		mockResult     *metadata.UserListResult
		mockError      error
		expectedStatus int
		expectedResult bool
	}{
		{
			name:        "成功获取用户列表",
			queryParams: "page=1&limit=20&search=test&role=admin&status=active",
			expectedParams: &metadata.UserListRequest{
				Page:   1,
				Limit:  20,
				Search: "test",
				Role:   "admin",
				Status: "active",
			},
			mockResult: &metadata.UserListResult{
				Items: []*metadata.User{
					{
						UserID:    "user1",
						Email:     "user1@test.com",
						Name:      "User One",
						Role:      "admin",
						Status:    "active",
						CreatedAt: "2024-06-24T14:00:00Z",
						LastLogin: "2024-06-24T13:00:00Z",
					},
					{
						UserID:    "user2",
						Email:     "user2@test.com",
						Name:      "User Two",
						Role:      "operator",
						Status:    "active",
						CreatedAt: "2024-06-24T12:00:00Z",
						LastLogin: "2024-06-24T11:00:00Z",
					},
				},
				Total: 2,
				Page:  1,
				Limit: 20,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedResult: true,
		},
		{
			name:        "无查询参数的默认请求",
			queryParams: "",
			expectedParams: &metadata.UserListRequest{
				Page:   0,
				Limit:  0,
				Search: "",
				Role:   "",
				Status: "",
			},
			mockResult: &metadata.UserListResult{
				Items: []*metadata.User{},
				Total: 0,
				Page:  1,
				Limit: 20,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedResult: true,
		},
		{
			name:        "分页查询",
			queryParams: "page=2&limit=10",
			expectedParams: &metadata.UserListRequest{
				Page:   2,
				Limit:  10,
				Search: "",
				Role:   "",
				Status: "",
			},
			mockResult: &metadata.UserListResult{
				Items: []*metadata.User{
					{
						UserID:    "user3",
						Email:     "user3@test.com",
						Name:      "User Three",
						Role:      "operator",
						Status:    "inactive",
						CreatedAt: "2024-06-24T10:00:00Z",
						LastLogin: "",
					},
				},
				Total: 21,
				Page:  2,
				Limit: 10,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedResult: true,
		},
		{
			name:        "角色筛选",
			queryParams: "role=operator",
			expectedParams: &metadata.UserListRequest{
				Page:   0,
				Limit:  0,
				Search: "",
				Role:   "operator",
				Status: "",
			},
			mockResult: &metadata.UserListResult{
				Items: []*metadata.User{
					{
						UserID:    "user2",
						Email:     "user2@test.com",
						Name:      "User Two",
						Role:      "operator",
						Status:    "active",
						CreatedAt: "2024-06-24T12:00:00Z",
						LastLogin: "2024-06-24T11:00:00Z",
					},
				},
				Total: 1,
				Page:  1,
				Limit: 20,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedResult: true,
		},
		{
			name:        "搜索用户",
			queryParams: "search=john",
			expectedParams: &metadata.UserListRequest{
				Page:   0,
				Limit:  0,
				Search: "john",
				Role:   "",
				Status: "",
			},
			mockResult: &metadata.UserListResult{
				Items: []*metadata.User{
					{
						UserID:    "user4",
						Email:     "john@test.com",
						Name:      "John Doe",
						Role:      "admin",
						Status:    "active",
						CreatedAt: "2024-06-24T09:00:00Z",
						LastLogin: "2024-06-24T08:00:00Z",
					},
				},
				Total: 1,
				Page:  1,
				Limit: 20,
			},
			mockError:      nil,
			expectedStatus: http.StatusOK,
			expectedResult: true,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试服务
			service, _, mockUserMgmt := setupTestService()
			
			// 设置mock期望
			mockUserMgmt.On("ListUsers", mock.Anything, mock.Anything, mock.MatchedBy(func(params *metadata.UserListRequest) bool {
				return params.Page == tt.expectedParams.Page &&
					params.Limit == tt.expectedParams.Limit &&
					params.Search == tt.expectedParams.Search &&
					params.Role == tt.expectedParams.Role &&
					params.Status == tt.expectedParams.Status
			})).Return(tt.mockResult, tt.mockError)
			
			// 创建HTTP请求
			req := httptest.NewRequest("GET", "/api/v3/user/list?"+tt.queryParams, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("BK_USER", "testuser")
			req.Header.Set("HTTP_BLUEKING_SUPPLIER_ID", "0")
			
			// 创建响应记录器
			w := httptest.NewRecorder()
			
			// 创建Gin上下文
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			
			// 调用被测试的方法
			service.getUserList(c)
			
			// 验证响应状态码
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// 解析响应JSON
			var response metadata.UserListResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			
			// 验证响应结果
			assert.Equal(t, tt.expectedResult, response.Result)
			
			if tt.expectedResult {
				assert.Equal(t, 0, response.Code)
				assert.Empty(t, response.ErrMsg)
				
				if tt.mockResult != nil {
					assert.Equal(t, len(tt.mockResult.Items), len(response.Data.Items))
					assert.Equal(t, tt.mockResult.Total, response.Data.Total)
					assert.Equal(t, tt.mockResult.Page, response.Data.Page)
					assert.Equal(t, tt.mockResult.Limit, response.Data.Limit)
					
					// 验证用户数据
					for i, expectedUser := range tt.mockResult.Items {
						actualUser := response.Data.Items[i]
						assert.Equal(t, expectedUser.UserID, actualUser.UserID)
						assert.Equal(t, expectedUser.Email, actualUser.Email)
						assert.Equal(t, expectedUser.Name, actualUser.Name)
						assert.Equal(t, expectedUser.Role, actualUser.Role)
						assert.Equal(t, expectedUser.Status, actualUser.Status)
						assert.Equal(t, expectedUser.CreatedAt, actualUser.CreatedAt)
						assert.Equal(t, expectedUser.LastLogin, actualUser.LastLogin)
					}
				}
			}
			
			// 验证mock调用
			mockUserMgmt.AssertExpectations(t)
		})
	}
}

// TestGetUserListWithError 测试获取用户列表时的错误情况
func TestGetUserListWithError(t *testing.T) {
	gin.SetMode(gin.TestMode)
	
	tests := []struct {
		name           string
		queryParams    string
		mockError      error
		expectedStatus int
		expectedResult bool
		expectedCode   int
	}{
		{
			name:           "核心服务返回错误",
			queryParams:    "page=1&limit=20",
			mockError:      assert.AnError,
			expectedStatus: http.StatusInternalServerError,
			expectedResult: false,
			expectedCode:   -1,
		},
	}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 创建测试服务
			service, _, mockUserMgmt := setupTestService()
			
			// 设置mock期望
			mockUserMgmt.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return((*metadata.UserListResult)(nil), tt.mockError)
			
			// 创建HTTP请求
			req := httptest.NewRequest("GET", "/api/v3/user/list?"+tt.queryParams, nil)
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("BK_USER", "testuser")
			req.Header.Set("HTTP_BLUEKING_SUPPLIER_ID", "0")
			
			// 创建响应记录器
			w := httptest.NewRecorder()
			
			// 创建Gin上下文
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			
			// 调用被测试的方法
			service.getUserList(c)
			
			// 验证响应状态码
			assert.Equal(t, tt.expectedStatus, w.Code)
			
			// 解析响应JSON
			var response metadata.UserListResponse
			err := json.Unmarshal(w.Body.Bytes(), &response)
			assert.NoError(t, err)
			
			// 验证响应结果
			assert.Equal(t, tt.expectedResult, response.Result)
			assert.Equal(t, tt.expectedCode, response.Code)
			assert.NotEmpty(t, response.ErrMsg)
			
			// 验证mock调用
			mockUserMgmt.AssertExpectations(t)
		})
	}
}

// BenchmarkGetUserList 性能测试
func BenchmarkGetUserList(b *testing.B) {
	gin.SetMode(gin.TestMode)
	
	// 创建测试服务
	service, _, mockUserMgmt := setupTestService()
	
	// 准备mock数据
	mockResult := &metadata.UserListResult{
		Items: make([]*metadata.User, 100),
		Total: 100,
		Page:  1,
		Limit: 100,
	}
	
	// 填充用户数据
	for i := 0; i < 100; i++ {
		mockResult.Items[i] = &metadata.User{
			UserID:    "user" + string(rune(i)),
			Email:     "user" + string(rune(i)) + "@test.com",
			Name:      "User " + string(rune(i)),
			Role:      "operator",
			Status:    "active",
			CreatedAt: "2024-06-24T14:00:00Z",
			LastLogin: "2024-06-24T13:00:00Z",
		}
	}
	
	// 设置mock期望
	mockUserMgmt.On("ListUsers", mock.Anything, mock.Anything, mock.Anything).Return(mockResult, nil)
	
	b.ResetTimer()
	
	for i := 0; i < b.N; i++ {
		// 创建HTTP请求
		req := httptest.NewRequest("GET", "/api/v3/user/list?page=1&limit=100", nil)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("BK_USER", "testuser")
		req.Header.Set("HTTP_BLUEKING_SUPPLIER_ID", "0")
		
		// 创建响应记录器
		w := httptest.NewRecorder()
		
		// 创建Gin上下文
		c, _ := gin.CreateTestContext(w)
		c.Request = req
		
		// 调用被测试的方法
		service.getUserList(c)
	}
}