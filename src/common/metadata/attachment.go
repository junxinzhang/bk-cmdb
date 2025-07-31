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
	"configcenter/src/common"
	"configcenter/src/common/errors"
)

// AttachmentMeta 附件元数据结构
type AttachmentMeta struct {
	ID           string    `json:"id" bson:"_id"`                              // 文件唯一ID
	ObjectID     string    `json:"bk_obj_id" bson:"bk_obj_id"`                 // 所属模型ID  
	InstanceID   int64     `json:"bk_inst_id" bson:"bk_inst_id"`               // 实例ID
	PropertyID   string    `json:"bk_property_id" bson:"bk_property_id"`       // 字段ID
	FileName     string    `json:"file_name" bson:"file_name"`                 // 原始文件名
	FileSize     int64     `json:"file_size" bson:"file_size"`                 // 文件大小
	ContentType  string    `json:"content_type" bson:"content_type"`           // MIME类型
	Extension    string    `json:"extension" bson:"extension"`                 // 文件扩展名
	StoragePath  string    `json:"storage_path" bson:"storage_path"`           // 存储路径
	UploadTime   *Time     `json:"upload_time" bson:"upload_time"`             // 上传时间
	Uploader     string    `json:"uploader" bson:"uploader"`                   // 上传者
	MD5Hash      string    `json:"md5_hash" bson:"md5_hash"`                   // 文件MD5值
	Status       string    `json:"status" bson:"status"`                       // 状态(active/deleted)
	PreviewURL   string    `json:"preview_url,omitempty" bson:"-"`             // 预览URL(运行时生成)
	DownloadURL  string    `json:"download_url,omitempty" bson:"-"`            // 下载URL(运行时生成)
	BizID        int64     `json:"bk_biz_id" bson:"bk_biz_id"`                 // 业务ID
	OwnerID      string    `json:"bk_supplier_account" bson:"bk_supplier_account"` // 开发商ID
}

// AttachmentValue 附件字段值结构(存储在实例数据中)
type AttachmentValue struct {
	FileIDs []string `json:"file_ids" bson:"file_ids"` // 文件ID列表
}

// AttachmentStatus 附件状态常量
const (
	AttachmentStatusActive  = "active"
	AttachmentStatusDeleted = "deleted"
)

// TableName 返回附件元数据表名
func (a AttachmentMeta) TableName() string {
	return common.BKTableNameAttachmentMeta
}

// GetID 获取附件ID
func (a *AttachmentMeta) GetID() string {
	return a.ID
}

// SetID 设置附件ID
func (a *AttachmentMeta) SetID(id string) {
	a.ID = id
}

// IsActive 检查附件是否为活跃状态
func (a *AttachmentMeta) IsActive() bool {
	return a.Status == AttachmentStatusActive
}

// GeneratePreviewURL 生成预览URL
func (a *AttachmentMeta) GeneratePreviewURL() string {
	return "/api/v3/attachment/preview/" + a.ID
}

// GenerateDownloadURL 生成下载URL
func (a *AttachmentMeta) GenerateDownloadURL() string {
	return "/api/v3/attachment/download/" + a.ID
}

// Validate 验证附件元数据
func (a *AttachmentMeta) Validate() error {
	if a.ObjectID == "" {
		return ErrAttachmentObjectIDRequired
	}
	if a.InstanceID <= 0 {
		return ErrAttachmentInstanceIDRequired
	}
	if a.PropertyID == "" {
		return ErrAttachmentPropertyIDRequired
	}
	if a.FileName == "" {
		return ErrAttachmentFileNameRequired
	}
	if a.FileSize <= 0 {
		return ErrAttachmentFileSizeInvalid
	}
	if a.ContentType == "" {
		return ErrAttachmentContentTypeRequired
	}
	if a.StoragePath == "" {
		return ErrAttachmentStoragePathRequired
	}
	if a.Uploader == "" {
		return ErrAttachmentUploaderRequired
	}
	if a.MD5Hash == "" {
		return ErrAttachmentMD5HashRequired
	}
	return nil
}

// AttachmentListRequest 附件列表查询请求
type AttachmentListRequest struct {
	ObjectID   string `json:"bk_obj_id"`      // 模型ID
	InstanceID int64  `json:"bk_inst_id"`     // 实例ID
	PropertyID string `json:"bk_property_id"` // 字段ID
	BizID      int64  `json:"bk_biz_id"`      // 业务ID
	Page       BasePage `json:"page"`         // 分页信息
}

// AttachmentListResponse 附件列表查询响应
type AttachmentListResponse struct {
	Count int               `json:"count"`
	Info  []*AttachmentMeta `json:"info"`
}

// AttachmentBatchInfoRequest 批量获取附件信息请求
type AttachmentBatchInfoRequest struct {
	FileIDs []string `json:"file_ids"` // 文件ID列表
}

// AttachmentBatchInfoResponse 批量获取附件信息响应
type AttachmentBatchInfoResponse struct {
	Info []*AttachmentMeta `json:"info"`
}

// AttachmentUploadResponse 附件上传响应
type AttachmentUploadResponse struct {
	*AttachmentMeta
}

// AttachmentPreviewData 预览数据
type AttachmentPreviewData struct {
	Type        string      `json:"type"`         // 预览类型(image/pdf/text/unsupported)
	Content     interface{} `json:"content"`      // 预览内容
	OriginalURL string      `json:"original_url"` // 原始文件URL
}

// AttachmentPreviewResponse 预览响应
type AttachmentPreviewResponse struct {
	*AttachmentPreviewData
}

// 附件相关错误定义
var (
	ErrAttachmentObjectIDRequired     = errors.NewCCError(common.CCErrCommParamsInvalid, "attachment object id is required")
	ErrAttachmentInstanceIDRequired   = errors.NewCCError(common.CCErrCommParamsInvalid, "attachment instance id is required")
	ErrAttachmentPropertyIDRequired   = errors.NewCCError(common.CCErrCommParamsInvalid, "attachment property id is required")
	ErrAttachmentFileNameRequired     = errors.NewCCError(common.CCErrCommParamsInvalid, "attachment file name is required")
	ErrAttachmentFileSizeInvalid      = errors.NewCCError(common.CCErrCommParamsInvalid, "attachment file size is invalid")
	ErrAttachmentContentTypeRequired  = errors.NewCCError(common.CCErrCommParamsInvalid, "attachment content type is required")
	ErrAttachmentStoragePathRequired  = errors.NewCCError(common.CCErrCommParamsInvalid, "attachment storage path is required")
	ErrAttachmentUploaderRequired     = errors.NewCCError(common.CCErrCommParamsInvalid, "attachment uploader is required")
	ErrAttachmentMD5HashRequired      = errors.NewCCError(common.CCErrCommParamsInvalid, "attachment md5 hash is required")
)