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

package attachment

import (
	"crypto/md5"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"configcenter/src/common/blog"

	"github.com/google/uuid"
)

// StorageManager 存储管理器
type StorageManager struct {
	BasePath string        // 基础存储路径
	TempPath string        // 临时文件路径
	mutex    sync.RWMutex  // 读写锁
}

// NewStorageManager 创建存储管理器
func NewStorageManager(basePath, tempPath string) *StorageManager {
	return &StorageManager{
		BasePath: basePath,
		TempPath: tempPath,
	}
}

// SaveFile 保存文件
func (sm *StorageManager) SaveFile(fileID string, content io.Reader, contentType string) (string, error) {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

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
		blog.Errorf("failed to create storage directory %s: %v", dir, err)
		return "", fmt.Errorf("failed to create storage directory: %v", err)
	}
	
	// 生成安全的文件扩展名
	ext := sm.getExtensionByContentType(contentType)
	if ext != "" {
		fileID = fileID + "." + ext
	}
	
	// 文件路径: /data/attachments/2024/01/15/images/uuid.jpg
	filePath := filepath.Join(dir, fileID)
	
	// 检查文件是否已存在
	if _, err := os.Stat(filePath); err == nil {
		return "", fmt.Errorf("file already exists: %s", filePath)
	}
	
	// 保存文件
	file, err := os.Create(filePath)
	if err != nil {
		blog.Errorf("failed to create file %s: %v", filePath, err)
		return "", fmt.Errorf("failed to create file: %v", err)
	}
	defer file.Close()
	
	// 设置文件权限为只读（除了所有者）
	if err := file.Chmod(0644); err != nil {
		blog.Warnf("failed to set file permissions for %s: %v", filePath, err)
	}
	
	_, err = io.Copy(file, content)
	if err != nil {
		os.Remove(filePath) // 清理失败的文件
		blog.Errorf("failed to copy file content to %s: %v", filePath, err)
		return "", fmt.Errorf("failed to copy file content: %v", err)
	}
	
	return filePath, nil
}

// GetFile 获取文件
func (sm *StorageManager) GetFile(filePath string) (io.ReadCloser, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 安全检查：确保文件路径在允许的目录内
	if !sm.isPathSafe(filePath) {
		return nil, fmt.Errorf("unsafe file path: %s", filePath)
	}
	
	file, err := os.Open(filePath)
	if err != nil {
		blog.Errorf("failed to open file %s: %v", filePath, err)
		return nil, fmt.Errorf("failed to open file: %v", err)
	}
	
	return file, nil
}

// DeleteFile 删除文件
func (sm *StorageManager) DeleteFile(filePath string) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	// 安全检查：确保文件路径在允许的目录内
	if !sm.isPathSafe(filePath) {
		return fmt.Errorf("unsafe file path: %s", filePath)
	}
	
	err := os.Remove(filePath)
	if err != nil {
		blog.Errorf("failed to delete file %s: %v", filePath, err)
		return fmt.Errorf("failed to delete file: %v", err)
	}
	
	return nil
}

// GetFileInfo 获取文件信息
func (sm *StorageManager) GetFileInfo(filePath string) (os.FileInfo, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	// 安全检查：确保文件路径在允许的目录内
	if !sm.isPathSafe(filePath) {
		return nil, fmt.Errorf("unsafe file path: %s", filePath)
	}
	
	info, err := os.Stat(filePath)
	if err != nil {
		blog.Errorf("failed to get file info for %s: %v", filePath, err)
		return nil, fmt.Errorf("failed to get file info: %v", err)
	}
	
	return info, nil
}

// CalculateMD5 计算文件MD5
func (sm *StorageManager) CalculateMD5(content io.Reader) (string, error) {
	hash := md5.New()
	_, err := io.Copy(hash, content)
	if err != nil {
		blog.Errorf("failed to calculate MD5: %v", err)
		return "", fmt.Errorf("failed to calculate MD5: %v", err)
	}
	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// ValidateFileType 验证文件类型（基于文件头魔数）
func (sm *StorageManager) ValidateFileType(content io.Reader, allowedTypes []string) (string, error) {
	// 读取文件头部分来检测真实的MIME类型
	buffer := make([]byte, 512)
	n, err := content.Read(buffer)
	if err != nil && err != io.EOF {
		return "", fmt.Errorf("failed to read file header: %v", err)
	}
	
	// 检测MIME类型
	contentType := http.DetectContentType(buffer[:n])
	
	// 验证是否在允许的类型列表中
	if !sm.isContentTypeAllowed(contentType, allowedTypes) {
		return "", fmt.Errorf("file type %s is not allowed", contentType)
	}
	
	return contentType, nil
}

// SanitizeFileName 清理文件名，移除危险字符
func (sm *StorageManager) SanitizeFileName(fileName string) string {
	// 移除危险字符
	reg := regexp.MustCompile(`[<>:"/\\|?*\x00-\x1f]`)
	fileName = reg.ReplaceAllString(fileName, "_")
	
	// 移除前后空格和点
	fileName = strings.Trim(fileName, " .")
	
	// 限制长度
	if len(fileName) > 255 {
		ext := filepath.Ext(fileName)
		baseName := fileName[:255-len(ext)]
		fileName = baseName + ext
	}
	
	// 如果文件名为空，使用默认名称
	if fileName == "" {
		fileName = "unnamed_file"
	}
	
	return fileName
}

// GenerateFileID 生成唯一文件ID
func (sm *StorageManager) GenerateFileID() string {
	return uuid.New().String()
}

// FindFileByMD5 通过MD5查找文件（用于去重）
func (sm *StorageManager) FindFileByMD5(md5Hash string) (string, bool) {
	// 这里可以集成到数据库查询中
	// 暂时返回false，表示未找到重复文件
	return "", false
}

// RecordMD5Mapping 记录MD5映射（用于去重）
func (sm *StorageManager) RecordMD5Mapping(md5Hash, filePath string) {
	// 这里可以集成到数据库中记录MD5到文件路径的映射
	// 用于文件去重功能
}

// CleanupTempFiles 清理临时文件
func (sm *StorageManager) CleanupTempFiles(maxAge time.Duration) error {
	sm.mutex.Lock()
	defer sm.mutex.Unlock()

	if sm.TempPath == "" {
		return nil
	}
	
	return filepath.Walk(sm.TempPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() && time.Since(info.ModTime()) > maxAge {
			if err := os.Remove(path); err != nil {
				blog.Warnf("failed to remove temp file %s: %v", path, err)
			}
		}
		
		return nil
	})
}

// GetStorageStats 获取存储统计信息
func (sm *StorageManager) GetStorageStats() (*StorageStats, error) {
	sm.mutex.RLock()
	defer sm.mutex.RUnlock()

	stats := &StorageStats{}
	
	err := filepath.Walk(sm.BasePath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			stats.TotalFiles++
			stats.TotalSize += info.Size()
		}
		
		return nil
	})
	
	if err != nil {
		return nil, fmt.Errorf("failed to calculate storage stats: %v", err)
	}
	
	return stats, nil
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

// getExtensionByContentType 根据MIME类型获取文件扩展名
func (sm *StorageManager) getExtensionByContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return "jpg"
	case "image/png":
		return "png"
	case "image/gif":
		return "gif"
	case "image/bmp":
		return "bmp"
	case "image/webp":
		return "webp"
	case "application/pdf":
		return "pdf"
	case "text/plain":
		return "txt"
	case "application/zip":
		return "zip"
	case "application/x-rar-compressed":
		return "rar"
	default:
		return ""
	}
}

// isContentTypeAllowed 检查内容类型是否被允许
func (sm *StorageManager) isContentTypeAllowed(contentType string, allowedTypes []string) bool {
	for _, allowedType := range allowedTypes {
		if allowedType == "*/*" {
			return true
		}
		
		if strings.HasSuffix(allowedType, "/*") {
			prefix := strings.TrimSuffix(allowedType, "/*")
			if strings.HasPrefix(contentType, prefix+"/") {
				return true
			}
		} else if contentType == allowedType {
			return true
		}
	}
	return false
}

// isPathSafe 检查文件路径是否安全（防止路径遍历攻击）
func (sm *StorageManager) isPathSafe(filePath string) bool {
	// 获取绝对路径
	absPath, err := filepath.Abs(filePath)
	if err != nil {
		return false
	}
	
	// 获取基础路径的绝对路径
	absBasePath, err := filepath.Abs(sm.BasePath)
	if err != nil {
		return false
	}
	
	// 检查文件是否在基础路径内
	return strings.HasPrefix(absPath, absBasePath)
}

// StorageStats 存储统计信息
type StorageStats struct {
	TotalFiles int64 `json:"total_files"` // 总文件数
	TotalSize  int64 `json:"total_size"`  // 总大小（字节）
}