# 蓝鲸CMDB附件字段类型和在线预览功能设计文档

## 概述

本设计文档基于需求分析，详细描述了蓝鲸CMDB附件字段类型和在线预览功能的技术实现方案。该功能将为CMDB模型增加附件上传能力，支持图片、文档、压缩包等文件类型的上传、存储、下载和在线预览。

## 架构设计

### 总体架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端Vue组件    │    │   Web服务层      │    │   文件存储层     │
│                │    │                │    │                │
│ - 附件上传组件   │◄──►│ - 附件API       │◄──►│ - 本地文件系统   │
│ - 预览弹窗组件   │    │ - 预览服务      │    │ - 文件元数据     │
│ - 文件列表组件   │    │ - 下载服务      │    │                │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         ▲                       ▲                       ▲
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   字段类型系统   │    │   权限验证层     │    │   MongoDB存储    │
│                │    │                │    │                │
│ - 字段类型定义   │    │ - 上传权限      │    │ - 文件元信息     │
│ - 字段验证      │    │ - 下载权限      │    │ - 关联关系      │
│ - 元数据扩展    │    │ - 预览权限      │    │                │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

### 数据流设计

#### 1. 文件上传流程
```
用户选择文件 → 前端验证 → 生成上传请求 → 后端权限验证 → 文件类型验证 
→ 病毒扫描 → 生成文件ID → 存储到本地 → 保存元信息到MongoDB → 返回文件信息
```

#### 2. 文件预览流程
```
用户点击预览 → 权限验证 → 获取文件信息 → 判断文件类型 
→ 生成预览内容 → 返回预览数据 → 前端渲染预览界面
```

#### 3. 文件下载流程
```
用户点击下载 → 权限验证 → 获取文件路径 → 读取文件内容 
→ 设置响应头 → 流式返回文件数据
```

## 组件设计

### 1. 后端组件设计

#### 1.1 字段类型扩展

**文件位置**: `src/common/definitions.go`

**新增字段类型定义**:
```go
const (
    // FieldTypeAttachment 附件字段类型
    FieldTypeAttachment string = "attachment"
)

// 更新FieldTypes数组
var FieldTypes = []string{
    FieldTypeSingleChar, FieldTypeLongChar, FieldTypeInt, FieldTypeFloat, 
    FieldTypeEnum, FieldTypeEnumMulti, FieldTypeDate, FieldTypeTime, 
    FieldTypeUser, FieldTypeOrganization, FieldTypeTimeZone, FieldTypeBool, 
    FieldTypeList, FieldTypeTable, FieldTypeInnerTable, FieldTypeEnumQuote,
    FieldTypeAttachment, // 新增附件类型
}
```

**附件字段配置结构**:
```go
// AttachmentOption 附件字段配置选项
type AttachmentOption struct {
    MaxFileSize    int64    `json:"max_file_size"`    // 最大文件大小(字节)
    AllowedTypes   []string `json:"allowed_types"`    // 允许的文件类型
    MaxFileCount   int      `json:"max_file_count"`   // 最大文件数量
    AllowPreview   bool     `json:"allow_preview"`    // 是否允许预览
    StoragePath    string   `json:"storage_path"`     // 存储路径
}
```

#### 1.2 附件元数据模型

**文件位置**: `src/common/metadata/attachment.go`

```go
package metadata

// AttachmentMeta 附件元数据结构
type AttachmentMeta struct {
    ID           string    `json:"id" bson:"_id"`                    // 文件唯一ID
    ObjectID     string    `json:"bk_obj_id" bson:"bk_obj_id"`       // 所属模型ID  
    InstanceID   int64     `json:"bk_inst_id" bson:"bk_inst_id"`     // 实例ID
    PropertyID   string    `json:"bk_property_id" bson:"bk_property_id"` // 字段ID
    FileName     string    `json:"file_name" bson:"file_name"`       // 原始文件名
    FileSize     int64     `json:"file_size" bson:"file_size"`       // 文件大小
    ContentType  string    `json:"content_type" bson:"content_type"` // MIME类型
    Extension    string    `json:"extension" bson:"extension"`       // 文件扩展名
    StoragePath  string    `json:"storage_path" bson:"storage_path"` // 存储路径
    UploadTime   *Time     `json:"upload_time" bson:"upload_time"`   // 上传时间
    Uploader     string    `json:"uploader" bson:"uploader"`         // 上传者
    MD5Hash      string    `json:"md5_hash" bson:"md5_hash"`         // 文件MD5值
    Status       string    `json:"status" bson:"status"`             // 状态(active/deleted)
    PreviewURL   string    `json:"preview_url,omitempty" bson:"-"`   // 预览URL(运行时生成)
    DownloadURL  string    `json:"download_url,omitempty" bson:"-"`  // 下载URL(运行时生成)
    BizID        int64     `json:"bk_biz_id" bson:"bk_biz_id"`       // 业务ID
    OwnerID      string    `json:"bk_supplier_account" bson:"bk_supplier_account"` // 开发商ID
}

// AttachmentValue 附件字段值结构(存储在实例数据中)
type AttachmentValue struct {
    FileIDs []string `json:"file_ids" bson:"file_ids"` // 文件ID列表
}
```

#### 1.3 附件服务API设计

**文件位置**: `src/web_server/service/attachment.go`

```go
package service

import (
    "mime/multipart"
    "configcenter/src/common/metadata"
)

// AttachmentService 附件服务接口
type AttachmentService interface {
    // UploadFile 上传文件
    UploadFile(kit *rest.Kit, req *UploadFileRequest) (*metadata.AttachmentMeta, error)
    
    // GetFile 获取文件信息
    GetFile(kit *rest.Kit, fileID string) (*metadata.AttachmentMeta, error)
    
    // DownloadFile 下载文件
    DownloadFile(kit *rest.Kit, fileID string) (*FileStream, error)
    
    // PreviewFile 预览文件
    PreviewFile(kit *rest.Kit, fileID string) (*PreviewData, error)
    
    // DeleteFile 删除文件
    DeleteFile(kit *rest.Kit, fileID string) error
    
    // ListFiles 列出文件
    ListFiles(kit *rest.Kit, req *ListFilesRequest) ([]*metadata.AttachmentMeta, error)
    
    // BatchDownload 批量下载
    BatchDownload(kit *rest.Kit, fileIDs []string) (*ZipStream, error)
}

// UploadFileRequest 上传文件请求
type UploadFileRequest struct {
    File       multipart.File   `json:"-"`              // 文件内容
    Header     *multipart.FileHeader `json:"-"`        // 文件头信息
    ObjectID   string           `json:"bk_obj_id"`      // 模型ID
    InstanceID int64            `json:"bk_inst_id"`     // 实例ID
    PropertyID string           `json:"bk_property_id"` // 字段ID
    BizID      int64            `json:"bk_biz_id"`      // 业务ID
    OwnerID    string           `json:"bk_supplier_account"` // 开发商ID
}

// ListFilesRequest 列出文件请求
type ListFilesRequest struct {
    ObjectID   string `json:"bk_obj_id"`      // 模型ID
    InstanceID int64  `json:"bk_inst_id"`     // 实例ID
    PropertyID string `json:"bk_property_id"` // 字段ID
    BizID      int64  `json:"bk_biz_id"`      // 业务ID
}

// FileStream 文件流
type FileStream struct {
    Content     io.ReadCloser `json:"-"`           // 文件内容流
    ContentType string        `json:"content_type"` // 内容类型
    FileName    string        `json:"file_name"`    // 文件名
    Size        int64         `json:"size"`         // 文件大小
}

// PreviewData 预览数据
type PreviewData struct {
    Type        string      `json:"type"`         // 预览类型(image/pdf/text)
    Content     interface{} `json:"content"`      // 预览内容
    OriginalURL string      `json:"original_url"` // 原始文件URL
}

// ZipStream ZIP文件流
type ZipStream struct {
    Content  io.ReadCloser `json:"-"`        // ZIP内容流
    FileName string        `json:"file_name"` // ZIP文件名
    Size     int64         `json:"size"`      // ZIP文件大小
}
```

#### 1.4 文件存储管理器

**文件位置**: `src/web_server/service/attachment/storage.go`

```go
package attachment

import (
    "crypto/md5"
    "os"
    "path/filepath"
    "time"
)

// StorageManager 存储管理器
type StorageManager struct {
    BasePath string // 基础存储路径
}

// NewStorageManager 创建存储管理器
func NewStorageManager(basePath string) *StorageManager {
    return &StorageManager{
        BasePath: basePath,
    }
}

// SaveFile 保存文件
func (sm *StorageManager) SaveFile(fileID string, content io.Reader, contentType string) (string, error) {
    // 按日期和类型分目录存储
    now := time.Now()
    year := now.Format("2006")
    month := now.Format("01")
    day := now.Format("02")
    
    // 根据Content-Type确定分类目录
    category := sm.getCategoryByContentType(contentType)
    
    // 构建存储路径: /data/attachments/2024/01/15/images/
    dir := filepath.Join(sm.BasePath, year, month, day, category)
    if err := os.MkdirAll(dir, 0755); err != nil {
        return "", err
    }
    
    // 文件路径: /data/attachments/2024/01/15/images/uuid.jpg
    filePath := filepath.Join(dir, fileID)
    
    // 保存文件
    file, err := os.Create(filePath)
    if err != nil {
        return "", err
    }
    defer file.Close()
    
    _, err = io.Copy(file, content)
    if err != nil {
        os.Remove(filePath) // 清理失败的文件
        return "", err
    }
    
    return filePath, nil
}

// GetFile 获取文件
func (sm *StorageManager) GetFile(filePath string) (io.ReadCloser, error) {
    return os.Open(filePath)
}

// DeleteFile 删除文件
func (sm *StorageManager) DeleteFile(filePath string) error {
    return os.Remove(filePath)
}

// GetFileInfo 获取文件信息
func (sm *StorageManager) GetFileInfo(filePath string) (os.FileInfo, error) {
    return os.Stat(filePath)
}

// getCategoryByContentType 根据MIME类型获取分类目录
func (sm *StorageManager) getCategoryByContentType(contentType string) string {
    switch {
    case strings.HasPrefix(contentType, "image/"):
        return "images"
    case contentType == "application/pdf":
        return "documents"
    case strings.HasPrefix(contentType, "text/"):
        return "texts"
    case strings.Contains(contentType, "zip") || strings.Contains(contentType, "rar"):
        return "archives"
    default:
        return "others"
    }
}

// CalculateMD5 计算文件MD5
func (sm *StorageManager) CalculateMD5(content io.Reader) (string, error) {
    hash := md5.New()
    _, err := io.Copy(hash, content)
    if err != nil {
        return "", err
    }
    return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
```

#### 1.5 预览服务

**文件位置**: `src/web_server/service/attachment/preview.go`

```go
package attachment

import (
    "bufio"
    "image"
    "image/jpeg"
    "image/png"
    "strings"
)

// PreviewService 预览服务
type PreviewService struct {
    storage *StorageManager
}

// NewPreviewService 创建预览服务
func NewPreviewService(storage *StorageManager) *PreviewService {
    return &PreviewService{
        storage: storage,
    }
}

// GeneratePreview 生成预览数据
func (ps *PreviewService) GeneratePreview(meta *metadata.AttachmentMeta) (*PreviewData, error) {
    switch {
    case strings.HasPrefix(meta.ContentType, "image/"):
        return ps.previewImage(meta)
    case meta.ContentType == "application/pdf":
        return ps.previewPDF(meta)
    case strings.HasPrefix(meta.ContentType, "text/"):
        return ps.previewText(meta)
    default:
        return &PreviewData{
            Type:        "unsupported",
            Content:     "不支持预览此文件类型",
            OriginalURL: fmt.Sprintf("/api/v3/attachment/download/%s", meta.ID),
        }, nil
    }
}

// previewImage 图片预览
func (ps *PreviewService) previewImage(meta *metadata.AttachmentMeta) (*PreviewData, error) {
    file, err := ps.storage.GetFile(meta.StoragePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    // 对于图片，返回原始URL和基本信息
    img, _, err := image.DecodeConfig(file)
    if err != nil {
        return nil, err
    }
    
    return &PreviewData{
        Type: "image",
        Content: map[string]interface{}{
            "width":  img.Width,
            "height": img.Height,
            "url":    fmt.Sprintf("/api/v3/attachment/download/%s", meta.ID),
        },
        OriginalURL: fmt.Sprintf("/api/v3/attachment/download/%s", meta.ID),
    }, nil
}

// previewPDF PDF预览
func (ps *PreviewService) previewPDF(meta *metadata.AttachmentMeta) (*PreviewData, error) {
    // PDF预览返回文件信息，前端使用PDF.js渲染
    return &PreviewData{
        Type: "pdf",
        Content: map[string]interface{}{
            "url":       fmt.Sprintf("/api/v3/attachment/download/%s", meta.ID),
            "file_name": meta.FileName,
            "file_size": meta.FileSize,
        },
        OriginalURL: fmt.Sprintf("/api/v3/attachment/download/%s", meta.ID),
    }, nil
}

// previewText 文本预览
func (ps *PreviewService) previewText(meta *metadata.AttachmentMeta) (*PreviewData, error) {
    file, err := ps.storage.GetFile(meta.StoragePath)
    if err != nil {
        return nil, err
    }
    defer file.Close()
    
    // 读取文本文件前1000行或50KB内容
    scanner := bufio.NewScanner(file)
    var lines []string
    lineCount := 0
    totalSize := 0
    maxLines := 1000
    maxSize := 50 * 1024 // 50KB
    
    for scanner.Scan() && lineCount < maxLines && totalSize < maxSize {
        line := scanner.Text()
        lines = append(lines, line)
        totalSize += len(line)
        lineCount++
    }
    
    return &PreviewData{
        Type: "text",
        Content: map[string]interface{}{
            "lines":      lines,
            "truncated":  lineCount >= maxLines || totalSize >= maxSize,
            "line_count": lineCount,
            "encoding":   "UTF-8", // 简化处理，实际应检测编码
        },
        OriginalURL: fmt.Sprintf("/api/v3/attachment/download/%s", meta.ID),
    }, nil
}
```

#### 1.6 字段验证扩展

**文件位置**: `src/common/metadata/attribute.go` (扩展现有文件)

```go
// 在现有的validatorMap中添加附件类型验证
func init() {
    // 在现有的attrValidatorMap中添加
    attrValidatorMap[common.FieldTypeAttachment] = attribute.validAttachment
}

// validAttachment 验证附件字段值
func (attribute *Attribute) validAttachment(ctx context.Context, data interface{}, key string) errors.RawErrorInfo {
    rawError := errors.RawErrorInfo{}
    
    if data == nil {
        return rawError
    }
    
    // 验证数据格式
    attachmentValue, ok := data.(metadata.AttachmentValue)
    if !ok {
        rawError.ErrCode = common.CCErrCommParamsInvalid
        rawError.Args = []interface{}{key}
        return rawError
    }
    
    // 获取字段配置
    option, err := attribute.getAttachmentOption()
    if err != nil {
        rawError.ErrCode = common.CCErrCommParamsInvalid
        rawError.Args = []interface{}{key}
        return rawError
    }
    
    // 验证文件数量限制
    if len(attachmentValue.FileIDs) > option.MaxFileCount {
        rawError.ErrCode = common.CCErrCommOverLimit
        rawError.Args = []interface{}{key, option.MaxFileCount}
        return rawError
    }
    
    return rawError
}

// getAttachmentOption 获取附件字段配置
func (attribute *Attribute) getAttachmentOption() (*AttachmentOption, error) {
    optionBytes, err := json.Marshal(attribute.Option)
    if err != nil {
        return nil, err
    }
    
    var option AttachmentOption
    err = json.Unmarshal(optionBytes, &option)
    if err != nil {
        return nil, err
    }
    
    // 设置默认值
    if option.MaxFileSize == 0 {
        option.MaxFileSize = 100 * 1024 * 1024 // 100MB
    }
    if option.MaxFileCount == 0 {
        option.MaxFileCount = 10
    }
    if len(option.AllowedTypes) == 0 {
        option.AllowedTypes = []string{"image/*", "application/pdf", "text/*", "application/zip"}
    }
    
    return &option, nil
}
```

### 2. 前端组件设计

#### 2.1 附件字段表单组件

**文件位置**: `src/ui/src/components/ui/form/attachment.vue`

```vue
<template>
  <div class="cmdb-form-attachment">
    <!-- 文件上传区域 -->
    <div class="upload-area" 
         :class="{ 'drag-over': dragOver, 'disabled': disabled }"
         @drop.prevent="handleDrop"
         @dragover.prevent="dragOver = true"
         @dragleave.prevent="dragOver = false">
      
      <!-- 上传按钮和拖拽提示 -->
      <div class="upload-trigger" v-if="!disabled && canAddMore">
        <bk-upload
          ref="upload"
          :multiple="multiple"
          :accept="acceptTypes"
          :size="maxFileSize"
          :limit="maxFileCount - fileList.length"
          :before-upload="handleBeforeUpload"
          :on-progress="handleProgress"
          :on-success="handleSuccess"
          :on-error="handleError"
          :show-file-list="false"
          :auto-upload="true"
          :url="uploadUrl"
          :headers="uploadHeaders">
          <div class="upload-content">
            <i class="bk-icon icon-plus"></i>
            <div class="upload-text">
              <p>{{ $t('点击上传或拖拽文件到此区域') }}</p>
              <p class="upload-tips">{{ uploadTips }}</p>
            </div>
          </div>
        </bk-upload>
      </div>
      
      <!-- 文件列表 -->
      <div class="file-list" v-if="fileList.length > 0">
        <div class="file-item" 
             v-for="(file, index) in fileList" 
             :key="file.id || index"
             :class="{ 'uploading': file.status === 'uploading', 'error': file.status === 'error' }">
          
          <!-- 文件图标 -->
          <div class="file-icon">
            <i :class="getFileIcon(file)"></i>
          </div>
          
          <!-- 文件信息 -->
          <div class="file-info">
            <div class="file-name" :title="file.name">{{ file.name }}</div>
            <div class="file-meta">
              <span class="file-size">{{ formatFileSize(file.size) }}</span>
              <span class="file-time" v-if="file.upload_time">
                {{ formatTime(file.upload_time) }}
              </span>
            </div>
          </div>
          
          <!-- 上传进度 -->
          <div class="file-progress" v-if="file.status === 'uploading'">
            <bk-progress 
              :percent="file.progress" 
              :show-text="false"
              size="small">
            </bk-progress>
            <span class="progress-text">{{ file.progress }}%</span>
          </div>
          
          <!-- 操作按钮 -->
          <div class="file-actions" v-else>
            <!-- 预览按钮 -->
            <bk-button 
              v-if="canPreview(file)"
              text
              size="small"
              @click="handlePreview(file)">
              <i class="bk-icon icon-eye"></i>
            </bk-button>
            
            <!-- 下载按钮 -->
            <bk-button 
              text
              size="small"
              @click="handleDownload(file)">
              <i class="bk-icon icon-download"></i>
            </bk-button>
            
            <!-- 删除按钮 -->
            <bk-button 
              v-if="!disabled"
              text
              size="small"
              @click="handleRemove(file, index)">
              <i class="bk-icon icon-delete"></i>
            </bk-button>
          </div>
        </div>
      </div>
    </div>
    
    <!-- 批量操作 -->
    <div class="batch-actions" v-if="fileList.length > 1">
      <bk-button size="small" @click="handleBatchDownload">
        <i class="bk-icon icon-download"></i>
        {{ $t('批量下载') }}
      </bk-button>
    </div>
    
    <!-- 预览弹窗 -->
    <attachment-preview-modal
      v-model="previewVisible"
      :file="currentPreviewFile"
      :file-list="fileList"
      @download="handleDownload"
      @delete="handleRemove">
    </attachment-preview-modal>
  </div>
</template>

<script>
import AttachmentPreviewModal from './attachment-preview-modal.vue'
import { fileIcons, previewableTypes } from './attachment-constants.js'

export default {
  name: 'cmdb-form-attachment',
  components: {
    AttachmentPreviewModal
  },
  props: {
    value: {
      type: [Array, Object],
      default: () => []
    },
    disabled: {
      type: Boolean,
      default: false
    },
    multiple: {
      type: Boolean,
      default: true
    },
    options: {
      type: Object,
      default: () => ({})
    },
    // 从属性配置中获取的选项
    maxFileSize: {
      type: Number,
      default: 100 * 1024 * 1024 // 100MB
    },
    maxFileCount: {
      type: Number,
      default: 10
    },
    allowedTypes: {
      type: Array,
      default: () => ['image/*', 'application/pdf', 'text/*', 'application/zip']
    }
  },
  data() {
    return {
      fileList: [],
      dragOver: false,
      previewVisible: false,
      currentPreviewFile: null,
      uploadUrl: '/api/v3/attachment/upload',
      uploadHeaders: {
        'X-Requested-With': 'XMLHttpRequest'
      }
    }
  },
  computed: {
    canAddMore() {
      return this.fileList.length < this.maxFileCount
    },
    acceptTypes() {
      return this.allowedTypes.join(',')
    },
    uploadTips() {
      const sizeLimit = this.formatFileSize(this.maxFileSize)
      const typeList = this.allowedTypes.map(type => {
        if (type.includes('image')) return '图片'
        if (type.includes('pdf')) return 'PDF'
        if (type.includes('text')) return '文本'
        if (type.includes('zip')) return '压缩包'
        return type
      }).join('、')
      return `支持${typeList}，单文件不超过${sizeLimit}，最多${this.maxFileCount}个文件`
    }
  },
  watch: {
    value: {
      immediate: true,
      handler(newVal) {
        this.initFileList(newVal)
      }
    }
  },
  methods: {
    // 初始化文件列表
    async initFileList(value) {
      if (!value) {
        this.fileList = []
        return
      }
      
      // 如果value是附件值对象，获取文件详情
      if (value.file_ids && Array.isArray(value.file_ids)) {
        try {
          const response = await this.$http.post('/api/v3/attachment/batch-info', {
            file_ids: value.file_ids
          })
          this.fileList = response.data || []
        } catch (error) {
          console.error('获取附件信息失败:', error)
          this.fileList = []
        }
      } else if (Array.isArray(value)) {
        this.fileList = value
      }
    },
    
    // 文件拖放处理
    handleDrop(event) {
      this.dragOver = false
      if (this.disabled || !this.canAddMore) return
      
      const files = Array.from(event.dataTransfer.files)
      this.uploadFiles(files)
    },
    
    // 上传前验证
    handleBeforeUpload(file) {
      // 验证文件类型
      if (!this.isValidFileType(file)) {
        this.$bkMessage({
          theme: 'error',
          message: `不支持的文件类型: ${file.name}`
        })
        return false
      }
      
      // 验证文件大小
      if (file.size > this.maxFileSize) {
        this.$bkMessage({
          theme: 'error',
          message: `文件大小超过限制: ${this.formatFileSize(this.maxFileSize)}`
        })
        return false
      }
      
      // 验证文件数量
      if (this.fileList.length >= this.maxFileCount) {
        this.$bkMessage({
          theme: 'error',
          message: `文件数量不能超过${this.maxFileCount}个`
        })
        return false
      }
      
      // 添加到文件列表
      const fileItem = {
        id: this.generateTempId(),
        name: file.name,
        size: file.size,
        type: file.type,
        status: 'uploading',
        progress: 0,
        file: file
      }
      this.fileList.push(fileItem)
      
      return true
    },
    
    // 上传进度
    handleProgress(event, file) {
      const fileItem = this.fileList.find(item => item.file === file)
      if (fileItem) {
        fileItem.progress = Math.round(event.percent)
      }
    },
    
    // 上传成功
    handleSuccess(response, file) {
      const fileItem = this.fileList.find(item => item.file === file)
      if (fileItem && response.result) {
        // 更新文件信息
        Object.assign(fileItem, response.data, {
          status: 'success',
          progress: 100
        })
        delete fileItem.file
        
        // 触发值更新
        this.emitValue()
        
        this.$bkMessage({
          theme: 'success',
          message: `${file.name} 上传成功`
        })
      }
    },
    
    // 上传失败
    handleError(error, file) {
      const fileItem = this.fileList.find(item => item.file === file)
      if (fileItem) {
        fileItem.status = 'error'
        fileItem.progress = 0
      }
      
      this.$bkMessage({
        theme: 'error',
        message: `${file.name} 上传失败: ${error.message || '未知错误'}`
      })
    },
    
    // 预览文件
    handlePreview(file) {
      this.currentPreviewFile = file
      this.previewVisible = true
    },
    
    // 下载文件
    handleDownload(file) {
      const url = `/api/v3/attachment/download/${file.id}`
      const link = document.createElement('a')
      link.href = url
      link.download = file.file_name || file.name
      document.body.appendChild(link)
      link.click()
      document.body.removeChild(link)
    },
    
    // 删除文件
    handleRemove(file, index) {
      this.$bkInfo({
        title: '确认删除',
        subTitle: `确定要删除文件 "${file.name}" 吗？`,
        confirmFn: async () => {
          try {
            if (file.id && file.status === 'success') {
              await this.$http.delete(`/api/v3/attachment/${file.id}`)
            }
            this.fileList.splice(index, 1)
            this.emitValue()
            this.$bkMessage({
              theme: 'success',
              message: '删除成功'
            })
          } catch (error) {
            this.$bkMessage({
              theme: 'error',
              message: `删除失败: ${error.message}`
            })
          }
        }
      })
    },
    
    // 批量下载
    async handleBatchDownload() {
      const fileIds = this.fileList
        .filter(file => file.id && file.status === 'success')
        .map(file => file.id)
      
      if (fileIds.length === 0) {
        this.$bkMessage({
          theme: 'warning',
          message: '没有可下载的文件'
        })
        return
      }
      
      try {
        const response = await this.$http.post('/api/v3/attachment/batch-download', {
          file_ids: fileIds
        }, {
          responseType: 'blob'
        })
        
        const blob = new Blob([response.data], { type: 'application/zip' })
        const url = window.URL.createObjectURL(blob)
        const link = document.createElement('a')
        link.href = url
        link.download = `attachments_${Date.now()}.zip`
        document.body.appendChild(link)
        link.click()
        document.body.removeChild(link)
        window.URL.revokeObjectURL(url)
      } catch (error) {
        this.$bkMessage({
          theme: 'error',
          message: `批量下载失败: ${error.message}`
        })
      }
    },
    
    // 发送值更新事件
    emitValue() {
      const successFiles = this.fileList.filter(file => file.status === 'success')
      const value = {
        file_ids: successFiles.map(file => file.id)
      }
      this.$emit('input', value)
      this.$emit('change', value)
    },
    
    // 工具方法
    isValidFileType(file) {
      return this.allowedTypes.some(type => {
        if (type.endsWith('/*')) {
          const prefix = type.slice(0, -2)
          return file.type.startsWith(prefix)
        }
        return file.type === type
      })
    },
    
    getFileIcon(file) {
      const extension = this.getFileExtension(file.name || file.file_name)
      return fileIcons[extension] || fileIcons.default
    },
    
    getFileExtension(fileName) {
      return fileName.split('.').pop().toLowerCase()
    },
    
    canPreview(file) {
      if (!file.content_type && !file.type) return false
      const contentType = file.content_type || file.type
      return previewableTypes.some(type => {
        if (type.endsWith('/*')) {
          return contentType.startsWith(type.slice(0, -2))
        }
        return contentType === type
      })
    },
    
    formatFileSize(bytes) {
      if (bytes === 0) return '0 B'
      const sizes = ['B', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(1024))
      return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i]
    },
    
    formatTime(timeStr) {
      return new Date(timeStr).toLocaleString()
    },
    
    generateTempId() {
      return 'temp_' + Date.now() + '_' + Math.random().toString(36).substr(2, 9)
    }
  }
}
</script>

<style lang="scss" scoped>
.cmdb-form-attachment {
  .upload-area {
    border: 1px dashed #dcdee5;
    border-radius: 4px;
    transition: all 0.3s;
    
    &.drag-over {
      border-color: #3a84ff;
      background-color: #f0f8ff;
    }
    
    &.disabled {
      background-color: #fafbfd;
      border-color: #dcdee5;
      cursor: not-allowed;
    }
  }
  
  .upload-trigger {
    padding: 20px;
    text-align: center;
    
    .upload-content {
      .bk-icon {
        font-size: 24px;
        color: #979ba5;
        margin-bottom: 8px;
      }
      
      .upload-text {
        color: #63656e;
        font-size: 14px;
        
        .upload-tips {
          font-size: 12px;
          color: #979ba5;
          margin-top: 4px;
        }
      }
    }
  }
  
  .file-list {
    padding: 10px;
  }
  
  .file-item {
    display: flex;
    align-items: center;
    padding: 8px 0;
    border-bottom: 1px solid #f0f1f5;
    
    &:last-child {
      border-bottom: none;
    }
    
    &.uploading {
      opacity: 0.7;
    }
    
    &.error {
      color: #ea3636;
      
      .file-icon i {
        color: #ea3636;
      }
    }
  }
  
  .file-icon {
    margin-right: 8px;
    
    i {
      font-size: 18px;
      color: #3a84ff;
    }
  }
  
  .file-info {
    flex: 1;
    min-width: 0;
    
    .file-name {
      font-size: 14px;
      color: #313238;
      margin-bottom: 4px;
      overflow: hidden;
      text-overflow: ellipsis;
      white-space: nowrap;
    }
    
    .file-meta {
      font-size: 12px;
      color: #979ba5;
      
      .file-time {
        margin-left: 8px;
      }
    }
  }
  
  .file-progress {
    width: 100px;
    margin-right: 8px;
    
    .progress-text {
      font-size: 12px;
      color: #979ba5;
      margin-left: 8px;
    }
  }
  
  .file-actions {
    display: flex;
    gap: 4px;
    
    .bk-button {
      padding: 4px;
      min-width: auto;
      
      .bk-icon {
        font-size: 14px;
      }
    }
  }
  
  .batch-actions {
    margin-top: 10px;
    text-align: right;
  }
}
</style>
```

#### 2.2 预览弹窗组件

**文件位置**: `src/ui/src/components/ui/form/attachment-preview-modal.vue`

```vue
<template>
  <bk-dialog
    v-model="visible"
    :title="dialogTitle"
    :width="800"
    :height="600"
    :show-footer="false"
    :close-icon="true"
    @value-change="handleVisibleChange">
    
    <div class="attachment-preview" v-if="file">
      <!-- 预览头部工具栏 -->
      <div class="preview-header">
        <div class="file-info">
          <span class="file-name">{{ file.file_name || file.name }}</span>
          <span class="file-size">{{ formatFileSize(file.file_size || file.size) }}</span>
        </div>
        
        <div class="preview-actions">
          <!-- 文件切换 -->
          <template v-if="fileList.length > 1">
            <bk-button
              text
              :disabled="currentIndex <= 0"
              @click="switchFile(-1)">
              <i class="bk-icon icon-angle-left"></i>
            </bk-button>
            <span class="file-index">{{ currentIndex + 1 }} / {{ fileList.length }}</span>
            <bk-button
              text
              :disabled="currentIndex >= fileList.length - 1"
              @click="switchFile(1)">
              <i class="bk-icon icon-angle-right"></i>
            </bk-button>
          </template>
          
          <!-- 操作按钮 -->
          <bk-button text @click="handleDownload">
            <i class="bk-icon icon-download"></i>
            {{ $t('下载') }}
          </bk-button>
          
          <bk-button text @click="handleDelete">
            <i class="bk-icon icon-delete"></i>
            {{ $t('删除') }}
          </bk-button>
        </div>
      </div>
      
      <!-- 预览内容区域 -->
      <div class="preview-content" v-bkloading="{ isLoading: loading }">
        <!-- 图片预览 -->
        <div class="image-preview" v-if="previewType === 'image'">
          <div class="image-container">
            <img
              :src="previewData.content.url"
              :alt="file.file_name || file.name"
              @load="handleImageLoad"
              @error="handleImageError"
              :style="imageStyle">
          </div>
          
          <!-- 图片操作工具栏 -->
          <div class="image-toolbar">
            <bk-button text @click="zoomIn">
              <i class="bk-icon icon-plus"></i>
            </bk-button>
            <span class="zoom-level">{{ Math.round(zoomLevel * 100) }}%</span>
            <bk-button text @click="zoomOut">
              <i class="bk-icon icon-minus"></i>
            </bk-button>
            <bk-button text @click="resetZoom">
              <i class="bk-icon icon-reset"></i>
            </bk-button>
            <bk-button text @click="rotateImage">
              <i class="bk-icon icon-refresh"></i>
            </bk-button>
            <bk-button text @click="toggleFullscreen">
              <i class="bk-icon icon-full-screen"></i>
            </bk-button>
          </div>
        </div>
        
        <!-- PDF预览 -->
        <div class="pdf-preview" v-else-if="previewType === 'pdf'">
          <iframe
            :src="pdfViewerUrl"
            width="100%"
            height="500px"
            frameborder="0">
          </iframe>
        </div>
        
        <!-- 文本预览 -->
        <div class="text-preview" v-else-if="previewType === 'text'">
          <div class="text-content">
            <pre><code>{{ textContent }}</code></pre>
          </div>
          <div class="text-info" v-if="previewData.content.truncated">
            <p class="truncated-tip">
              <i class="bk-icon icon-info"></i>
              {{ $t('文件内容过长，仅显示前') }} {{ previewData.content.line_count }} {{ $t('行') }}
            </p>
            <bk-button size="small" @click="handleDownload">
              {{ $t('下载完整文件') }}
            </bk-button>
          </div>
        </div>
        
        <!-- 不支持预览 -->
        <div class="unsupported-preview" v-else>
          <div class="unsupported-content">
            <i class="bk-icon icon-file"></i>
            <p>{{ $t('不支持预览此文件类型') }}</p>
            <bk-button theme="primary" @click="handleDownload">
              <i class="bk-icon icon-download"></i>
              {{ $t('下载文件') }}
            </bk-button>
          </div>
        </div>
      </div>
    </div>
  </bk-dialog>
</template>

<script>
export default {
  name: 'attachment-preview-modal',
  props: {
    value: {
      type: Boolean,
      default: false
    },
    file: {
      type: Object,
      default: null
    },
    fileList: {
      type: Array,
      default: () => []
    }
  },
  data() {
    return {
      visible: false,
      loading: false,
      previewData: null,
      previewType: null,
      
      // 图片预览相关
      zoomLevel: 1,
      rotation: 0,
      
      // 文本预览相关
      textContent: ''
    }
  },
  computed: {
    dialogTitle() {
      return this.file ? (this.file.file_name || this.file.name) : '文件预览'
    },
    
    currentIndex() {
      if (!this.file || this.fileList.length === 0) return -1
      return this.fileList.findIndex(f => f.id === this.file.id)
    },
    
    imageStyle() {
      return {
        transform: `scale(${this.zoomLevel}) rotate(${this.rotation}deg)`,
        transition: 'transform 0.3s ease'
      }
    },
    
    pdfViewerUrl() {
      if (this.previewType !== 'pdf' || !this.previewData) return ''
      return `/pdf-viewer.html?file=${encodeURIComponent(this.previewData.content.url)}`
    }
  },
  watch: {
    value: {
      immediate: true,
      handler(val) {
        this.visible = val
        if (val && this.file) {
          this.loadPreview()
        }
      }
    },
    
    file(newFile) {
      if (newFile && this.visible) {
        this.loadPreview()
      }
    }
  },
  methods: {
    handleVisibleChange(visible) {
      this.$emit('input', visible)
      if (!visible) {
        this.resetPreview()
      }
    },
    
    async loadPreview() {
      if (!this.file || !this.file.id) return
      
      this.loading = true
      try {
        const response = await this.$http.get(`/api/v3/attachment/preview/${this.file.id}`)
        this.previewData = response.data
        this.previewType = this.previewData.type
        
        if (this.previewType === 'text') {
          this.textContent = this.previewData.content.lines.join('\n')
        }
      } catch (error) {
        console.error('加载预览失败:', error)
        this.previewType = 'unsupported'
        this.$bkMessage({
          theme: 'error',
          message: '预览加载失败'
        })
      } finally {
        this.loading = false
      }
    },
    
    switchFile(direction) {
      const newIndex = this.currentIndex + direction
      if (newIndex >= 0 && newIndex < this.fileList.length) {
        const newFile = this.fileList[newIndex]
        this.$emit('file-change', newFile)
      }
    },
    
    handleDownload() {
      this.$emit('download', this.file)
    },
    
    handleDelete() {
      const index = this.currentIndex
      this.$emit('delete', this.file, index)
      
      // 如果还有其他文件，切换到下一个文件
      if (this.fileList.length > 1) {
        const nextIndex = index < this.fileList.length - 1 ? index : index - 1
        if (nextIndex >= 0) {
          this.$emit('file-change', this.fileList[nextIndex])
        } else {
          this.handleVisibleChange(false)
        }
      } else {
        this.handleVisibleChange(false)
      }
    },
    
    // 图片操作方法
    zoomIn() {
      this.zoomLevel = Math.min(this.zoomLevel * 1.2, 5)
    },
    
    zoomOut() {
      this.zoomLevel = Math.max(this.zoomLevel / 1.2, 0.1)
    },
    
    resetZoom() {
      this.zoomLevel = 1
      this.rotation = 0
    },
    
    rotateImage() {
      this.rotation = (this.rotation + 90) % 360
    },
    
    toggleFullscreen() {
      // 实现全屏预览
      const element = this.$el.querySelector('.image-preview')
      if (element.requestFullscreen) {
        element.requestFullscreen()
      }
    },
    
    handleImageLoad() {
      // 图片加载完成
    },
    
    handleImageError() {
      this.$bkMessage({
        theme: 'error',
        message: '图片加载失败'
      })
    },
    
    resetPreview() {
      this.previewData = null
      this.previewType = null
      this.textContent = ''
      this.zoomLevel = 1
      this.rotation = 0
    },
    
    formatFileSize(bytes) {
      if (bytes === 0) return '0 B'
      const sizes = ['B', 'KB', 'MB', 'GB']
      const i = Math.floor(Math.log(bytes) / Math.log(1024))
      return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i]
    }
  }
}
</script>

<style lang="scss" scoped>
.attachment-preview {
  height: 100%;
  display: flex;
  flex-direction: column;
}

.preview-header {
  display: flex;
  justify-content: space-between;
  align-items: center;
  padding: 10px 0;
  border-bottom: 1px solid #dcdee5;
  margin-bottom: 10px;
  
  .file-info {
    .file-name {
      font-size: 16px;
      font-weight: 500;
      color: #313238;
      margin-right: 10px;
    }
    
    .file-size {
      font-size: 12px;
      color: #979ba5;
    }
  }
  
  .preview-actions {
    display: flex;
    align-items: center;
    gap: 8px;
    
    .file-index {
      font-size: 12px;
      color: #979ba5;
      margin: 0 4px;
    }
  }
}

.preview-content {
  flex: 1;
  overflow: auto;
}

.image-preview {
  height: 100%;
  display: flex;
  flex-direction: column;
  
  .image-container {
    flex: 1;
    display: flex;
    justify-content: center;
    align-items: center;
    overflow: hidden;
    background: #f5f7fa;
    
    img {
      max-width: 100%;
      max-height: 100%;
      object-fit: contain;
    }
  }
  
  .image-toolbar {
    display: flex;
    justify-content: center;
    align-items: center;
    gap: 8px;
    padding: 10px;
    border-top: 1px solid #dcdee5;
    background: #fafbfd;
    
    .zoom-level {
      font-size: 12px;
      color: #979ba5;
      min-width: 40px;
      text-align: center;
    }
  }
}

.pdf-preview {
  height: 500px;
}

.text-preview {
  .text-content {
    max-height: 400px;
    overflow: auto;
    border: 1px solid #dcdee5;
    border-radius: 4px;
    
    pre {
      margin: 0;
      padding: 12px;
      font-size: 12px;
      line-height: 1.5;
      background: #fafbfd;
      
      code {
        background: none;
        color: #313238;
      }
    }
  }
  
  .text-info {
    margin-top: 10px;
    padding: 10px;
    background: #f0f8ff;
    border-radius: 4px;
    
    .truncated-tip {
      display: flex;
      align-items: center;
      margin-bottom: 8px;
      font-size: 12px;
      color: #3a84ff;
      
      .bk-icon {
        margin-right: 4px;
      }
    }
  }
}

.unsupported-preview {
  height: 100%;
  display: flex;
  justify-content: center;
  align-items: center;
  
  .unsupported-content {
    text-align: center;
    color: #979ba5;
    
    .bk-icon {
      font-size: 48px;
      margin-bottom: 16px;
    }
    
    p {
      font-size: 14px;
      margin-bottom: 16px;
    }
  }
}
</style>
```

#### 2.3 附件常量和工具

**文件位置**: `src/ui/src/components/ui/form/attachment-constants.js`

```javascript
// 文件类型图标映射
export const fileIcons = {
  // 图片类型
  jpg: 'icon-image',
  jpeg: 'icon-image',
  png: 'icon-image', 
  gif: 'icon-image',
  bmp: 'icon-image',
  webp: 'icon-image',
  
  // 文档类型
  pdf: 'icon-file-pdf',
  doc: 'icon-file-word',
  docx: 'icon-file-word',
  xls: 'icon-file-excel',
  xlsx: 'icon-file-excel',
  ppt: 'icon-file-powerpoint',
  pptx: 'icon-file-powerpoint',
  
  // 文本类型
  txt: 'icon-file-text',
  md: 'icon-file-text',
  json: 'icon-file-code',
  xml: 'icon-file-code',
  html: 'icon-file-code',
  css: 'icon-file-code',
  js: 'icon-file-code',
  
  // 压缩包类型
  zip: 'icon-file-zip',
  rar: 'icon-file-zip',
  '7z': 'icon-file-zip',
  tar: 'icon-file-zip',
  gz: 'icon-file-zip',
  
  // 默认图标
  default: 'icon-file'
}

// 支持预览的文件类型
export const previewableTypes = [
  'image/jpeg',
  'image/png', 
  'image/gif',
  'image/bmp',
  'image/webp',
  'application/pdf',
  'text/plain',
  'text/csv',
  'application/json',
  'text/xml',
  'text/html',
  'text/css',
  'application/javascript'
]

// 默认附件配置
export const defaultAttachmentOptions = {
  maxFileSize: 100 * 1024 * 1024, // 100MB
  maxFileCount: 10,
  allowedTypes: [
    'image/*',
    'application/pdf', 
    'text/*',
    'application/zip',
    'application/x-rar-compressed'
  ],
  allowPreview: true,
  storagePath: '/data/attachments'
}

// 文件大小格式化
export function formatFileSize(bytes) {
  if (bytes === 0) return '0 B'
  const sizes = ['B', 'KB', 'MB', 'GB', 'TB']
  const i = Math.floor(Math.log(bytes) / Math.log(1024))
  return Math.round(bytes / Math.pow(1024, i) * 100) / 100 + ' ' + sizes[i]
}

// 获取文件扩展名
export function getFileExtension(fileName) {
  if (!fileName) return ''
  return fileName.split('.').pop().toLowerCase()
}

// 检查文件类型是否被允许
export function isFileTypeAllowed(file, allowedTypes) {
  return allowedTypes.some(type => {
    if (type.endsWith('/*')) {
      const prefix = type.slice(0, -2)
      return file.type.startsWith(prefix)
    }
    return file.type === type
  })
}

// 检查文件是否可以预览
export function canPreviewFile(file) {
  const contentType = file.content_type || file.type
  if (!contentType) return false
  
  return previewableTypes.some(type => {
    if (type.endsWith('/*')) {
      return contentType.startsWith(type.slice(0, -2))
    }
    return contentType === type
  })
}
```

## 数据模型设计

### 1. 数据库表结构

#### 附件元数据表 (`cc_attachment_meta`)
```javascript
{
  "_id": "6507f1234567890abcdef123",           // MongoDB ObjectID
  "bk_obj_id": "host",                        // 所属模型ID
  "bk_inst_id": 12345,                        // 实例ID
  "bk_property_id": "screenshots",            // 属性字段ID
  "file_name": "server-screenshot.png",       // 原始文件名
  "file_size": 2048576,                       // 文件大小(字节)
  "content_type": "image/png",                // MIME类型
  "extension": "png",                         // 文件扩展名
  "storage_path": "/data/attachments/2024/01/15/images/6507f1234567890abcdef123.png",  // 存储路径
  "upload_time": {"$date": "2024-01-15T10:30:00.000Z"},  // 上传时间
  "uploader": "admin",                        // 上传者
  "md5_hash": "d41d8cd98f00b204e9800998ecf8427e",  // 文件MD5值
  "status": "active",                         // 状态(active/deleted)
  "bk_biz_id": 2,                            // 业务ID
  "bk_supplier_account": "0"                  // 开发商ID
}
```

#### 模型属性表扩展 (`cc_ObjAttDes`)
```javascript
{
  // ... 现有字段
  "bk_property_type": "attachment",           // 字段类型为attachment
  "option": {                                 // 附件字段配置
    "max_file_size": 104857600,              // 最大文件大小(100MB)
    "allowed_types": [                       // 允许的文件类型
      "image/*",
      "application/pdf", 
      "text/*",
      "application/zip"
    ],
    "max_file_count": 10,                    // 最大文件数量
    "allow_preview": true,                   // 是否允许预览
    "storage_path": "/data/attachments"      // 存储路径
  }
}
```

#### 实例数据表 (模型实例表)
```javascript
{
  // ... 其他字段
  "screenshots": {                           // 附件字段值
    "file_ids": [                           // 文件ID数组
      "6507f1234567890abcdef123",
      "6507f1234567890abcdef124"
    ]
  }
}
```

### 2. 文件存储结构

```
/data/attachments/
├── 2024/
│   ├── 01/
│   │   ├── 15/
│   │   │   ├── images/
│   │   │   │   ├── 6507f1234567890abcdef123.png
│   │   │   │   └── 6507f1234567890abcdef124.jpg
│   │   │   ├── documents/
│   │   │   │   └── 6507f1234567890abcdef125.pdf
│   │   │   ├── texts/
│   │   │   │   └── 6507f1234567890abcdef126.txt
│   │   │   └── archives/
│   │   │       └── 6507f1234567890abcdef127.zip
│   │   └── 16/
│   └── 02/
└── temp/                                    // 临时上传目录
    └── upload_temp_files/
```

## 接口设计

### 1. 附件上传接口

**接口路径**: `POST /api/v3/attachment/upload`

**请求格式**: `multipart/form-data`

**请求参数**:
```javascript
{
  "file": File,                    // 文件内容
  "bk_obj_id": "host",            // 模型ID  
  "bk_inst_id": 12345,            // 实例ID
  "bk_property_id": "screenshots", // 属性字段ID
  "bk_biz_id": 2                  // 业务ID
}
```

**响应格式**:
```javascript
{
  "result": true,
  "code": 0,
  "message": "success",
  "data": {
    "id": "6507f1234567890abcdef123",
    "file_name": "server-screenshot.png",
    "file_size": 2048576,
    "content_type": "image/png",
    "upload_time": "2024-01-15T10:30:00Z",
    "preview_url": "/api/v3/attachment/preview/6507f1234567890abcdef123",
    "download_url": "/api/v3/attachment/download/6507f1234567890abcdef123"
  }
}
```

### 2. 文件下载接口

**接口路径**: `GET /api/v3/attachment/download/{file_id}`

**响应格式**: 文件流

**响应头**:
```
Content-Type: application/octet-stream
Content-Disposition: attachment; filename="server-screenshot.png"
Content-Length: 2048576
```

### 3. 文件预览接口

**接口路径**: `GET /api/v3/attachment/preview/{file_id}`

**响应格式**:
```javascript
{
  "result": true,
  "code": 0, 
  "message": "success",
  "data": {
    "type": "image",                    // 预览类型
    "content": {                        // 预览内容
      "width": 1920,
      "height": 1080,
      "url": "/api/v3/attachment/download/6507f1234567890abcdef123"
    },
    "original_url": "/api/v3/attachment/download/6507f1234567890abcdef123"
  }
}
```

### 4. 批量下载接口

**接口路径**: `POST /api/v3/attachment/batch-download`

**请求参数**:
```javascript
{
  "file_ids": [
    "6507f1234567890abcdef123",
    "6507f1234567890abcdef124"
  ]
}
```

**响应格式**: ZIP文件流

### 5. 文件删除接口

**接口路径**: `DELETE /api/v3/attachment/{file_id}`

**响应格式**:
```javascript
{
  "result": true,
  "code": 0,
  "message": "success"
}
```

## 错误处理

### 1. 错误码定义

**文件位置**: `src/common/errInfo.go` (扩展)

```go
const (
    // 附件相关错误码 1199200-1199299
    CCErrAttachmentFileNotFound           = 1199200  // 文件不存在
    CCErrAttachmentFileSizeExceeded       = 1199201  // 文件大小超出限制
    CCErrAttachmentFileTypeNotSupported   = 1199202  // 不支持的文件类型
    CCErrAttachmentFileCountExceeded      = 1199203  // 文件数量超出限制
    CCErrAttachmentUploadFailed           = 1199204  // 文件上传失败
    CCErrAttachmentStorageError           = 1199205  // 存储错误
    CCErrAttachmentPermissionDenied       = 1199206  // 权限不足
    CCErrAttachmentPreviewNotSupported    = 1199207  // 不支持预览
    CCErrAttachmentVirusDetected          = 1199208  // 检测到病毒
    CCErrAttachmentCorrupted              = 1199209  // 文件损坏
)
```

### 2. 错误处理策略

#### 前端错误处理
- 文件大小超限：上传前验证，显示友好提示
- 文件类型不支持：上传前验证，显示支持的类型列表
- 网络错误：提供重试机制
- 权限错误：显示权限申请引导

#### 后端错误处理
- 存储空间不足：返回明确错误信息，记录日志
- 文件损坏：验证文件完整性，清理临时文件
- 权限验证失败：记录审计日志，返回标准权限错误

## 安全考虑

### 1. 文件上传安全

#### 文件类型验证
```go
// ValidateFileType 验证文件类型
func ValidateFileType(file multipart.File, allowedTypes []string) error {
    // 1. 检查文件头魔数
    buffer := make([]byte, 512)
    _, err := file.Read(buffer)
    if err != nil {
        return err
    }
    file.Seek(0, 0) // 重置文件指针
    
    // 2. 检测真实MIME类型
    contentType := http.DetectContentType(buffer)
    
    // 3. 验证是否在允许列表中
    for _, allowedType := range allowedTypes {
        if matchContentType(contentType, allowedType) {
            return nil
        }
    }
    
    return errors.New("不支持的文件类型")
}
```

#### 文件名安全处理
```go
// SanitizeFileName 清理文件名
func SanitizeFileName(fileName string) string {
    // 移除危险字符
    reg := regexp.MustCompile(`[<>:"/\\|?*]`)
    fileName = reg.ReplaceAllString(fileName, "_")
    
    // 限制长度
    if len(fileName) > 255 {
        ext := filepath.Ext(fileName)
        baseName := fileName[:255-len(ext)]
        fileName = baseName + ext
    }
    
    return fileName
}
```

### 2. 访问权限控制

```go
// CheckAttachmentPermission 检查附件权限
func CheckAttachmentPermission(ctx context.Context, userID string, action string, meta *AttachmentMeta) error {
    // 1. 检查实例访问权限
    hasInstanceAccess, err := checkInstancePermission(ctx, userID, meta.ObjectID, meta.InstanceID, action)
    if err != nil {
        return err
    }
    if !hasInstanceAccess {
        return errors.New("无权访问该实例")
    }
    
    // 2. 检查字段权限
    hasFieldAccess, err := checkFieldPermission(ctx, userID, meta.ObjectID, meta.PropertyID, action)
    if err != nil {
        return err
    }
    if !hasFieldAccess {
        return errors.New("无权访问该字段")
    }
    
    return nil
}
```

### 3. 存储安全

#### 文件路径安全
- 使用UUID作为文件名，避免路径遍历攻击
- 文件存储在指定目录外无法访问
- 禁止执行权限，避免恶意脚本执行

#### 备份和恢复
- 定期备份附件元数据
- 支持文件系统快照
- 提供数据恢复机制

## 性能优化

### 1. 文件上传优化

#### 分块上传
```javascript
// 前端分块上传实现
class ChunkUploader {
  constructor(file, options = {}) {
    this.file = file
    this.chunkSize = options.chunkSize || 2 * 1024 * 1024 // 2MB
    this.totalChunks = Math.ceil(file.size / this.chunkSize)
    this.uploadedChunks = 0
  }
  
  async upload() {
    const uploadId = await this.initializeUpload()
    
    for (let i = 0; i < this.totalChunks; i++) {
      const chunk = this.getChunk(i)
      await this.uploadChunk(uploadId, i, chunk)
      this.uploadedChunks++
      this.onProgress(this.uploadedChunks / this.totalChunks)
    }
    
    return await this.completeUpload(uploadId)
  }
  
  getChunk(index) {
    const start = index * this.chunkSize
    const end = Math.min(start + this.chunkSize, this.file.size)
    return this.file.slice(start, end)
  }
}
```

#### 并发控制
```go
// 后端并发上传限制
type UploadLimiter struct {
    semaphore chan struct{}
}

func NewUploadLimiter(maxConcurrent int) *UploadLimiter {
    return &UploadLimiter{
        semaphore: make(chan struct{}, maxConcurrent),
    }
}

func (ul *UploadLimiter) Acquire() {
    ul.semaphore <- struct{}{}
}

func (ul *UploadLimiter) Release() {
    <-ul.semaphore
}
```

### 2. 存储优化

#### 文件去重
```go
// 文件去重实现
func (sm *StorageManager) DeduplicateFile(md5Hash string, content io.Reader) (string, bool, error) {
    // 检查是否已存在相同MD5的文件
    existingPath, exists := sm.findFileByMD5(md5Hash)
    if exists {
        return existingPath, true, nil // 返回现有文件路径，标记为重复
    }
    
    // 保存新文件
    filePath, err := sm.SaveFile(generateFileID(), content, "")
    if err != nil {
        return "", false, err
    }
    
    // 记录MD5映射
    sm.recordMD5Mapping(md5Hash, filePath)
    
    return filePath, false, nil
}
```

#### 缓存策略
```go
// 预览缓存
type PreviewCache struct {
    cache map[string]*PreviewData
    mutex sync.RWMutex
    ttl   time.Duration
}

func (pc *PreviewCache) Get(fileID string) (*PreviewData, bool) {
    pc.mutex.RLock()
    defer pc.mutex.RUnlock()
    
    data, exists := pc.cache[fileID]
    return data, exists
}

func (pc *PreviewCache) Set(fileID string, data *PreviewData) {
    pc.mutex.Lock()
    defer pc.mutex.Unlock()
    
    pc.cache[fileID] = data
    
    // 设置过期清理
    time.AfterFunc(pc.ttl, func() {
        pc.Delete(fileID)
    })
}
```

### 3. 数据库优化

#### 索引设计
```javascript
// 附件元数据表索引
db.cc_attachment_meta.createIndex({
  "bk_obj_id": 1,
  "bk_inst_id": 1,
  "bk_property_id": 1
})

db.cc_attachment_meta.createIndex({
  "md5_hash": 1
})

db.cc_attachment_meta.createIndex({
  "status": 1,
  "upload_time": 1
})
```

#### 查询优化
```go
// 批量查询文件信息
func (as *AttachmentService) BatchGetFiles(ctx context.Context, fileIDs []string) ([]*AttachmentMeta, error) {
    pipeline := []bson.M{
        {"$match": bson.M{"_id": bson.M{"$in": fileIDs}}},
        {"$sort": bson.M{"upload_time": -1}},
    }
    
    cursor, err := as.db.Collection("cc_attachment_meta").Aggregate(ctx, pipeline)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    
    var files []*AttachmentMeta
    err = cursor.All(ctx, &files)
    return files, err
}
```

## 测试策略

### 1. 单元测试

#### 后端测试
```go
func TestAttachmentUpload(t *testing.T) {
    tests := []struct {
        name        string
        file        *multipart.FileHeader
        options     *AttachmentOption
        expectError bool
    }{
        {
            name: "valid image upload",
            file: createTestImageFile(),
            options: &AttachmentOption{
                MaxFileSize:  10 * 1024 * 1024,
                AllowedTypes: []string{"image/*"},
                MaxFileCount: 5,
            },
            expectError: false,
        },
        {
            name: "file size exceeded",
            file: createLargeTestFile(),
            options: &AttachmentOption{
                MaxFileSize:  1024,
                AllowedTypes: []string{"*/*"},
                MaxFileCount: 5,
            },
            expectError: true,
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            service := NewAttachmentService()
            _, err := service.UploadFile(context.Background(), &UploadFileRequest{
                File:   tt.file,
                // ... other fields
            })
            
            if tt.expectError && err == nil {
                t.Error("expected error but got none")
            }
            if !tt.expectError && err != nil {
                t.Errorf("unexpected error: %v", err)
            }
        })
    }
}
```

#### 前端测试
```javascript
describe('AttachmentComponent', () => {
  it('should validate file type before upload', () => {
    const wrapper = mount(AttachmentComponent, {
      propsData: {
        allowedTypes: ['image/*']
      }
    })
    
    const invalidFile = new File([''], 'test.txt', { type: 'text/plain' })
    const isValid = wrapper.vm.isValidFileType(invalidFile)
    
    expect(isValid).toBe(false)
  })
  
  it('should handle upload progress correctly', async () => {
    const wrapper = mount(AttachmentComponent)
    const progressEvent = { percent: 50 }
    const file = new File([''], 'test.jpg', { type: 'image/jpeg' })
    
    wrapper.vm.handleProgress(progressEvent, file)
    
    expect(wrapper.vm.fileList[0].progress).toBe(50)
  })
})
```

### 2. 集成测试

```go
func TestAttachmentWorkflow(t *testing.T) {
    // 1. 上传文件
    uploadResp, err := uploadTestFile()
    require.NoError(t, err)
    
    fileID := uploadResp.Data.ID
    
    // 2. 获取文件信息
    fileInfo, err := getFileInfo(fileID)
    require.NoError(t, err)
    assert.Equal(t, uploadResp.Data.FileName, fileInfo.FileName)
    
    // 3. 预览文件
    previewData, err := previewFile(fileID)
    require.NoError(t, err)
    assert.NotEmpty(t, previewData.Content)
    
    // 4. 下载文件
    downloadResp, err := downloadFile(fileID)
    require.NoError(t, err)
    assert.NotNil(t, downloadResp.Body)
    
    // 5. 删除文件
    err = deleteFile(fileID)
    require.NoError(t, err)
    
    // 6. 验证文件已删除
    _, err = getFileInfo(fileID)
    assert.Error(t, err)
}
```

### 3. 性能测试

```go
func BenchmarkFileUpload(b *testing.B) {
    service := NewAttachmentService()
    testFile := createTestFile(1024 * 1024) // 1MB file
    
    b.ResetTimer()
    for i := 0; i < b.N; i++ {
        _, err := service.UploadFile(context.Background(), &UploadFileRequest{
            File: testFile,
            // ... other fields
        })
        if err != nil {
            b.Fatal(err)
        }
    }
}

func TestConcurrentUpload(t *testing.T) {
    const numGoroutines = 10
    const uploadsPerGoroutine = 5
    
    var wg sync.WaitGroup
    errors := make(chan error, numGoroutines*uploadsPerGoroutine)
    
    for i := 0; i < numGoroutines; i++ {
        wg.Add(1)
        go func() {
            defer wg.Done()
            
            for j := 0; j < uploadsPerGoroutine; j++ {
                _, err := uploadTestFile()
                if err != nil {
                    errors <- err
                }
            }
        }()
    }
    
    wg.Wait()
    close(errors)
    
    for err := range errors {
        t.Error(err)
    }
}
```

## 部署和运维

### 1. 配置管理

**附件服务配置**:
```yaml
attachment:
  storage:
    type: "local"                           # 存储类型
    base_path: "/data/attachments"          # 基础存储路径
    temp_path: "/tmp/attachment_temp"       # 临时文件路径
  limits:
    max_file_size: 104857600               # 最大文件大小(100MB)
    max_files_per_request: 20              # 单次请求最大文件数
    concurrent_uploads: 10                 # 并发上传限制
  security:
    virus_scan: true                       # 是否启用病毒扫描
    allowed_extensions:                    # 允许的文件扩展名
      - jpg
      - png
      - pdf
      - txt
      - zip
  cache:
    preview_cache_size: 100               # 预览缓存大小(MB)
    preview_cache_ttl: "24h"              # 预览缓存过期时间
  cleanup:
    temp_file_ttl: "1h"                   # 临时文件过期时间
    deleted_file_retention: "30d"         # 已删除文件保留时间
```

### 2. 监控指标

#### 业务指标
- 文件上传成功率
- 文件下载成功率  
- 平均上传时间
- 平均下载时间
- 存储空间使用情况
- 文件类型分布

#### 系统指标
- 磁盘I/O性能
- 内存使用情况
- CPU使用率
- 网络带宽使用

### 3. 日志记录

```go
// 附件操作日志
type AttachmentLog struct {
    Action     string    `json:"action"`      // upload/download/delete/preview
    FileID     string    `json:"file_id"`     // 文件ID
    FileName   string    `json:"file_name"`   // 文件名
    FileSize   int64     `json:"file_size"`   // 文件大小
    UserID     string    `json:"user_id"`     // 用户ID
    ClientIP   string    `json:"client_ip"`   // 客户端IP
    UserAgent  string    `json:"user_agent"`  // 用户代理
    Duration   int64     `json:"duration"`    // 操作耗时(毫秒)
    Success    bool      `json:"success"`     // 是否成功
    Error      string    `json:"error,omitempty"` // 错误信息
    Timestamp  time.Time `json:"timestamp"`   // 时间戳
}
```

### 4. 备份策略

#### 文件备份
```bash
#!/bin/bash
# 附件文件备份脚本

BACKUP_SOURCE="/data/attachments" 
BACKUP_DEST="/backup/attachments"
DATE=$(date +%Y%m%d_%H%M%S)

# 创建增量备份
rsync -avz --link-dest="${BACKUP_DEST}/latest" \
      "${BACKUP_SOURCE}/" \
      "${BACKUP_DEST}/backup_${DATE}/"

# 更新最新备份链接
ln -sfn "backup_${DATE}" "${BACKUP_DEST}/latest"

# 清理7天前的备份
find "${BACKUP_DEST}" -maxdepth 1 -name "backup_*" -mtime +7 -exec rm -rf {} \;
```

#### 元数据备份
```bash
#!/bin/bash
# MongoDB附件元数据备份

MONGO_HOST="localhost:27017"
MONGO_DB="cmdb"
BACKUP_DIR="/backup/mongodb"
DATE=$(date +%Y%m%d_%H%M%S)

# 备份附件相关集合
mongodump --host $MONGO_HOST \
          --db $MONGO_DB \
          --collection cc_attachment_meta \
          --out "${BACKUP_DIR}/attachment_backup_${DATE}"

# 压缩备份文件
tar -czf "${BACKUP_DIR}/attachment_backup_${DATE}.tar.gz" \
    "${BACKUP_DIR}/attachment_backup_${DATE}"

# 清理原始备份目录
rm -rf "${BACKUP_DIR}/attachment_backup_${DATE}"
```

这个设计文档提供了蓝鲸CMDB附件字段功能的完整技术实现方案，涵盖了从数据模型到用户界面的所有层面，确保功能的完整性、安全性和可维护性。