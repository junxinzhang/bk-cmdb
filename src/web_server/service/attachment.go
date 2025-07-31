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
	"archive/zip"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"strings"
	"time"

	"configcenter/src/ac/iam"
	"configcenter/src/common"
	"configcenter/src/common/blog"
	"configcenter/src/common/http/rest"
	"configcenter/src/common/mapstr"
	"configcenter/src/common/metadata"
	"configcenter/src/storage/dal"
	"configcenter/src/web_server/service/attachment"
)

// getClientIP 获取客户端IP地址
func getClientIP(kit *rest.Kit) string {
	clientIP := kit.Header.Get("X-Real-IP")
	if clientIP == "" {
		clientIP = kit.Header.Get("X-Forwarded-For")
	}
	if clientIP == "" {
		clientIP = "unknown"
	}
	return clientIP
}

// AttachmentService 附件服务
type AttachmentService struct {
	storage      *attachment.StorageManager
	db           dal.RDB
	iam          *iam.IAM
	maxFileSize  int64
	allowedTypes []string
}

// NewAttachmentService 创建附件服务
func NewAttachmentService(storage *attachment.StorageManager, db dal.RDB, iam *iam.IAM) *AttachmentService {
	return &AttachmentService{
		storage:      storage,
		db:           db,
		iam:          iam,
		maxFileSize:  100 * 1024 * 1024, // 100MB
		allowedTypes: []string{"image/*", "application/pdf", "text/*", "application/zip"},
	}
}

// UploadFileRequest 上传文件请求
type UploadFileRequest struct {
	File       multipart.File        `json:"-"`              // 文件内容
	Header     *multipart.FileHeader `json:"-"`              // 文件头信息
	ObjectID   string                `json:"bk_obj_id"`      // 模型ID
	InstanceID int64                 `json:"bk_inst_id"`     // 实例ID
	PropertyID string                `json:"bk_property_id"` // 字段ID
	BizID      int64                 `json:"bk_biz_id"`      // 业务ID
	OwnerID    string                `json:"bk_supplier_account"` // 开发商ID
}

// ListFilesRequest 列出文件请求
type ListFilesRequest struct {
	ObjectID   string               `json:"bk_obj_id"`      // 模型ID
	InstanceID int64                `json:"bk_inst_id"`     // 实例ID
	PropertyID string               `json:"bk_property_id"` // 字段ID
	BizID      int64                `json:"bk_biz_id"`      // 业务ID
	Page       metadata.BasePage    `json:"page"`           // 分页信息
}

// FileStream 文件流
type FileStream struct {
	Content     io.ReadCloser `json:"-"`            // 文件内容流
	ContentType string        `json:"content_type"` // 内容类型
	FileName    string        `json:"file_name"`    // 文件名
	Size        int64         `json:"size"`         // 文件大小
}

// PreviewData 预览数据
type PreviewData struct {
	Type        string      `json:"type"`         // 预览类型(image/pdf/text/unsupported)
	Content     interface{} `json:"content"`      // 预览内容
	OriginalURL string      `json:"original_url"` // 原始文件URL
}

// ZipStream ZIP文件流
type ZipStream struct {
	Content  io.ReadCloser `json:"-"`        // ZIP内容流
	FileName string        `json:"file_name"` // ZIP文件名
	Size     int64         `json:"size"`      // ZIP文件大小
}

// UploadFile 上传文件
func (as *AttachmentService) UploadFile(kit *rest.Kit, req *UploadFileRequest) (*metadata.AttachmentMeta, error) {
	// 1. 权限验证
	if err := as.checkUploadPermission(kit, req); err != nil {
		return nil, err
	}

	// 2. 基础验证
	if req.File == nil || req.Header == nil {
		return nil, kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "file")
	}

	// 3. 文件大小验证
	if req.Header.Size > as.maxFileSize {
		return nil, kit.CCError.CCErrorf(common.CCErrAttachmentFileSizeExceeded, as.maxFileSize)
	}

	// 4. 文件类型验证
	contentType, err := as.storage.ValidateFileType(req.File, as.allowedTypes)
	if err != nil {
		return nil, kit.CCError.CCErrorf(common.CCErrAttachmentFileTypeNotSupported, err.Error())
	}

	// 5. 重置文件指针
	if seeker, ok := req.File.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	// 6. 计算MD5
	md5Hash, err := as.storage.CalculateMD5(req.File)
	if err != nil {
		return nil, kit.CCError.CCErrorf(common.CCErrAttachmentUploadFailed, err.Error())
	}

	// 7. 重置文件指针
	if seeker, ok := req.File.(io.Seeker); ok {
		seeker.Seek(0, io.SeekStart)
	}

	// 8. 生成文件ID和存储
	fileID := as.storage.GenerateFileID()
	storagePath, err := as.storage.SaveFile(fileID, req.File, contentType)
	if err != nil {
		return nil, kit.CCError.CCErrorf(common.CCErrAttachmentStorageError, err.Error())
	}

	// 9. 创建附件元数据
	now := metadata.Time{Time: time.Now()}
	attachmentMeta := &metadata.AttachmentMeta{
		ID:          fileID,
		ObjectID:    req.ObjectID,
		InstanceID:  req.InstanceID,
		PropertyID:  req.PropertyID,
		FileName:    as.storage.SanitizeFileName(req.Header.Filename),
		FileSize:    req.Header.Size,
		ContentType: contentType,
		Extension:   as.getFileExtension(req.Header.Filename),
		StoragePath: storagePath,
		UploadTime:  &now,
		Uploader:    kit.User,
		MD5Hash:     md5Hash,
		Status:      metadata.AttachmentStatusActive,
		BizID:       req.BizID,
		OwnerID:     req.OwnerID,
	}

	// 10. 生成URL
	attachmentMeta.PreviewURL = attachmentMeta.GeneratePreviewURL()
	attachmentMeta.DownloadURL = attachmentMeta.GenerateDownloadURL()

	// 11. 保存到数据库
	if err := as.saveAttachmentMeta(kit, attachmentMeta); err != nil {
		// 清理已上传的文件
		as.storage.DeleteFile(storagePath)
		return nil, err
	}

	// 12. 记录审计日志
	as.logAttachmentOperation(kit, "upload", attachmentMeta, nil)

	return attachmentMeta, nil
}

// GetFile 获取文件信息
func (as *AttachmentService) GetFile(kit *rest.Kit, fileID string) (*metadata.AttachmentMeta, error) {
	// 1. 参数验证
	if fileID == "" {
		return nil, kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "file_id")
	}

	// 2. 从数据库获取文件元数据
	attachmentMeta, err := as.getAttachmentMeta(kit, fileID)
	if err != nil {
		return nil, err
	}

	// 3. 权限验证
	if err := as.checkDownloadPermission(kit, attachmentMeta); err != nil {
		return nil, err
	}

	// 4. 生成URL
	attachmentMeta.PreviewURL = attachmentMeta.GeneratePreviewURL()
	attachmentMeta.DownloadURL = attachmentMeta.GenerateDownloadURL()

	return attachmentMeta, nil
}

// DownloadFile 下载文件
func (as *AttachmentService) DownloadFile(kit *rest.Kit, fileID string) (*FileStream, error) {
	// 1. 获取文件信息
	attachmentMeta, err := as.GetFile(kit, fileID)
	if err != nil {
		return nil, err
	}

	// 2. 打开文件
	file, err := as.storage.GetFile(attachmentMeta.StoragePath)
	if err != nil {
		return nil, kit.CCError.CCErrorf(common.CCErrAttachmentFileNotFound, fileID)
	}

	// 3. 记录下载日志
	as.logAttachmentOperation(kit, "download", attachmentMeta, nil)

	return &FileStream{
		Content:     file,
		ContentType: attachmentMeta.ContentType,
		FileName:    attachmentMeta.FileName,
		Size:        attachmentMeta.FileSize,
	}, nil
}

// DeleteFile 删除文件
func (as *AttachmentService) DeleteFile(kit *rest.Kit, fileID string) error {
	// 1. 获取文件信息
	attachmentMeta, err := as.GetFile(kit, fileID)
	if err != nil {
		return err
	}

	// 2. 权限验证（删除权限）
	if err := as.checkDeletePermission(kit, attachmentMeta); err != nil {
		return err
	}

	// 3. 标记为已删除（软删除）
	if err := as.markAttachmentDeleted(kit, fileID); err != nil {
		return err
	}

	// 4. 记录审计日志
	as.logAttachmentOperation(kit, "delete", attachmentMeta, nil)

	return nil
}

// ListFiles 列出文件
func (as *AttachmentService) ListFiles(kit *rest.Kit, req *ListFilesRequest) ([]*metadata.AttachmentMeta, error) {
	// 1. 权限验证
	if err := as.checkListPermission(kit, req); err != nil {
		return nil, err
	}

	// 2. 构建查询条件
	condition := mapstr.MapStr{
		"bk_obj_id":      req.ObjectID,
		"bk_inst_id":     req.InstanceID,
		"bk_property_id": req.PropertyID,
		"bk_biz_id":      req.BizID,
		"status":         metadata.AttachmentStatusActive,
	}

	// 3. 从数据库查询
	result := make([]*metadata.AttachmentMeta, 0)
	err := as.db.Table(common.BKTableNameAttachmentMeta).Find(condition).Start(uint64(req.Page.Start)).
		Limit(uint64(req.Page.Limit)).All(kit.Ctx, &result)
	if err != nil {
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBSelectFailed, err.Error())
	}

	// 4. 生成URL
	for _, meta := range result {
		meta.PreviewURL = meta.GeneratePreviewURL()
		meta.DownloadURL = meta.GenerateDownloadURL()
	}

	return result, nil
}

// BatchDownload 批量下载
func (as *AttachmentService) BatchDownload(kit *rest.Kit, fileIDs []string) (*ZipStream, error) {
	if len(fileIDs) == 0 {
		return nil, kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, "file_ids")
	}

	// 1. 创建内存ZIP
	var zipBuffer bytes.Buffer
	zipWriter := zip.NewWriter(&zipBuffer)

	// 2. 添加文件到ZIP
	fileCount := 0
	for _, fileID := range fileIDs {
		// 获取文件信息
		attachmentMeta, err := as.GetFile(kit, fileID)
		if err != nil {
			blog.Warnf("failed to get file %s for batch download: %v, rid: %s", fileID, err, kit.Rid)
			continue
		}

		// 打开文件
		file, err := as.storage.GetFile(attachmentMeta.StoragePath)
		if err != nil {
			blog.Warnf("failed to open file %s for batch download: %v, rid: %s", fileID, err, kit.Rid)
			continue
		}

		// 处理重名文件
		fileName := attachmentMeta.FileName
		for i := 1; ; i++ {
			if _, err := zipWriter.Create(fileName); err == nil {
				break
			}
			// 文件名已存在，添加序号
			ext := as.getFileExtension(attachmentMeta.FileName)
			nameWithoutExt := strings.TrimSuffix(attachmentMeta.FileName, "."+ext)
			if ext != "" {
				fileName = fmt.Sprintf("%s_%d.%s", nameWithoutExt, i, ext)
			} else {
				fileName = fmt.Sprintf("%s_%d", attachmentMeta.FileName, i)
			}
		}

		// 创建ZIP文件条目
		zipFile, err := zipWriter.Create(fileName)
		if err != nil {
			file.Close()
			blog.Errorf("failed to create zip entry for file %s: %v, rid: %s", fileID, err, kit.Rid)
			continue
		}

		// 复制文件内容到ZIP
		_, err = io.Copy(zipFile, file)
		file.Close()
		if err != nil {
			blog.Errorf("failed to copy file %s to zip: %v, rid: %s", fileID, err, kit.Rid)
			continue
		}

		fileCount++
	}

	// 3. 关闭ZIP写入器
	if err := zipWriter.Close(); err != nil {
		return nil, kit.CCError.CCErrorf(common.CCErrAttachmentStorageError, err.Error())
	}

	if fileCount == 0 {
		return nil, kit.CCError.CCErrorf(common.CCErrAttachmentFileNotFound, "no valid files found")
	}

	// 4. 记录批量下载日志
	as.logAttachmentOperation(kit, "batch_download", nil, map[string]interface{}{
		"file_ids":   fileIDs,
		"file_count": fileCount,
	})

	// 5. 创建ZIP流
	zipReader := bytes.NewReader(zipBuffer.Bytes())
	zipReadCloser := io.NopCloser(zipReader)

	return &ZipStream{
		Content:  zipReadCloser,
		FileName: fmt.Sprintf("attachments_%d.zip", time.Now().Unix()),
		Size:     int64(zipBuffer.Len()),
	}, nil
}

// 私有方法

// saveAttachmentMeta 保存附件元数据到数据库
func (as *AttachmentService) saveAttachmentMeta(kit *rest.Kit, meta *metadata.AttachmentMeta) error {
	// 验证元数据
	if err := meta.Validate(); err != nil {
		return kit.CCError.CCErrorf(common.CCErrCommParamsInvalid, err.Error())
	}

	// 保存到数据库
	if err := as.db.Table(common.BKTableNameAttachmentMeta).Insert(kit.Ctx, meta); err != nil {
		return kit.CCError.CCErrorf(common.CCErrCommDBInsertFailed, err.Error())
	}

	return nil
}

// getAttachmentMeta 从数据库获取附件元数据
func (as *AttachmentService) getAttachmentMeta(kit *rest.Kit, fileID string) (*metadata.AttachmentMeta, error) {
	condition := mapstr.MapStr{
		"_id":    fileID,
		"status": metadata.AttachmentStatusActive,
	}

	result := &metadata.AttachmentMeta{}
	err := as.db.Table(common.BKTableNameAttachmentMeta).Find(condition).One(kit.Ctx, result)
	if err != nil {
		if as.db.IsNotFoundError(err) {
			return nil, kit.CCError.CCErrorf(common.CCErrAttachmentFileNotFound, fileID)
		}
		return nil, kit.CCError.CCErrorf(common.CCErrCommDBSelectFailed, err.Error())
	}

	return result, nil
}

// markAttachmentDeleted 标记附件为已删除
func (as *AttachmentService) markAttachmentDeleted(kit *rest.Kit, fileID string) error {
	condition := mapstr.MapStr{"_id": fileID}
	update := mapstr.MapStr{
		"status":    metadata.AttachmentStatusDeleted,
		"last_time": metadata.Time{Time: time.Now()},
	}

	err := as.db.Table(common.BKTableNameAttachmentMeta).Update(kit.Ctx, condition, update)
	if err != nil {
		return kit.CCError.CCErrorf(common.CCErrCommDBUpdateFailed, err.Error())
	}

	return nil
}

// 权限验证方法

// checkUploadPermission 检查上传权限
func (as *AttachmentService) checkUploadPermission(kit *rest.Kit, req *UploadFileRequest) error {
	// TODO: 实现基于IAM的权限验证
	// 这里应该检查用户是否有对指定实例的编辑权限
	return nil
}

// checkDownloadPermission 检查下载权限
func (as *AttachmentService) checkDownloadPermission(kit *rest.Kit, meta *metadata.AttachmentMeta) error {
	// TODO: 实现基于IAM的权限验证
	// 这里应该检查用户是否有对指定实例的查看权限
	return nil
}

// checkDeletePermission 检查删除权限
func (as *AttachmentService) checkDeletePermission(kit *rest.Kit, meta *metadata.AttachmentMeta) error {
	// TODO: 实现基于IAM的权限验证
	// 这里应该检查用户是否有对指定实例的删除权限
	return nil
}

// checkListPermission 检查列表权限
func (as *AttachmentService) checkListPermission(kit *rest.Kit, req *ListFilesRequest) error {
	// TODO: 实现基于IAM的权限验证
	// 这里应该检查用户是否有对指定实例的查看权限
	return nil
}

// 工具方法

// getFileExtension 获取文件扩展名
func (as *AttachmentService) getFileExtension(fileName string) string {
	parts := strings.Split(fileName, ".")
	if len(parts) > 1 {
		return strings.ToLower(parts[len(parts)-1])
	}
	return ""
}

// logAttachmentOperation 记录附件操作日志
func (as *AttachmentService) logAttachmentOperation(kit *rest.Kit, action string, meta *metadata.AttachmentMeta, extra map[string]interface{}) {
	logData := map[string]interface{}{
		"action":    action,
		"user":      kit.User,
		"client_ip": getClientIP(kit),
		"rid":       kit.Rid,
		"timestamp": time.Now(),
	}

	if meta != nil {
		logData["file_id"] = meta.ID
		logData["file_name"] = meta.FileName
		logData["file_size"] = meta.FileSize
		logData["object_id"] = meta.ObjectID
		logData["instance_id"] = meta.InstanceID
		logData["property_id"] = meta.PropertyID
	}

	if extra != nil {
		for k, v := range extra {
			logData[k] = v
		}
	}

	logJSON, _ := json.Marshal(logData)
	blog.Infof("attachment operation: %s, rid: %s", string(logJSON), kit.Rid)
}