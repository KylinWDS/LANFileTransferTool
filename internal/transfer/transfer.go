package transfer

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"lanfiletransfertool/internal/config"
	"lanfiletransfertool/internal/stats"
	"lanfiletransfertool/pkg/errors"
	"lanfiletransfertool/pkg/logger"
	"lanfiletransfertool/pkg/utils"
)

type Service struct {
	config    *config.TransferConfig
	dataDir   string
	files     map[string]*FileInfo
	mu        sync.RWMutex
}

type FileInfo struct {
	ID          string
	Path        string
	Name        string
	Size        int64
	Type        string
	Checksum    string
	ModTime     time.Time
	RegisteredAt time.Time
}

type TransferProgress struct {
	FileID       string
	FileName     string
	TotalSize    int64
	Transferred int64
	Speed        float64
	Progress     float64
	Status       string
}

func NewService(transferConfig *config.TransferConfig, dataDir string) *Service {
	return &Service{
		config:  transferConfig,
		dataDir: dataDir,
		files:   make(map[string]*FileInfo),
	}
}

func (s *Service) RegisterFile(filePath string) (*FileInfo, error) {
	fileInfo, err := os.Stat(filePath)
	if err != nil {
		return nil, errors.ErrFileNotFound
	}

	checksum, err := s.calculateChecksum(filePath)
	if err != nil {
		logger.Warn("计算校验和失败: %v", err)
		checksum = ""
	}

	id := utils.GenerateID()
	info := &FileInfo{
		ID:           id,
		Path:         filePath,
		Name:         fileInfo.Name(),
		Size:         fileInfo.Size(),
		Type:         s.getFileType(filePath),
		Checksum:     checksum,
		ModTime:      fileInfo.ModTime(),
		RegisteredAt: time.Now(),
	}

	s.mu.Lock()
	s.files[id] = info
	s.mu.Unlock()

	logger.Info("文件注册成功: %s (ID: %s)", info.Name, id)
	return info, nil
}

func (s *Service) GetFileInfo(id string) (*FileInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.files[id]
	if !exists {
		return nil, errors.ErrFileNotFound
	}
	return info, nil
}

func (s *Service) GetAvailableFiles() ([]*FileInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	files := make([]*FileInfo, 0, len(s.files))
	for _, info := range s.files {
		files = append(files, info)
	}
	return files, nil
}

func (s *Service) DownloadFile(id, savePath string, progressChan chan float64) (*TransferResult, error) {
	info, err := s.GetFileInfo(id)
	if err != nil {
		return nil, err
	}

	// 检查是否是文件夹
	stat, err := os.Stat(info.Path)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %w", err)
	}

	if stat.IsDir() {
		// 下载整个文件夹
		return s.downloadDirectory(info, savePath, progressChan)
	}

	// 下载单个文件
	return s.downloadSingleFile(info, savePath, progressChan)
}

// DownloadFileByPath 根据文件路径直接下载（支持跨重启解析）
func (s *Service) DownloadFileByPath(filePath, fileName string, fileSize int64, savePath string, progressChan chan float64) (*TransferResult, error) {
	// 检查文件是否存在
	stat, err := os.Stat(filePath)
	if err != nil {
		return nil, fmt.Errorf("文件不存在: %w", err)
	}

	// 创建临时 FileInfo
	info := &FileInfo{
		ID:       "token-download",
		Path:     filePath,
		Name:     fileName,
		Size:     fileSize,
		Checksum: "",
	}

	// 如果传入的文件名为空，使用实际文件名
	if info.Name == "" {
		info.Name = stat.Name()
	}
	// 如果传入的大小为0，使用实际大小
	if info.Size == 0 {
		info.Size = stat.Size()
	}

	if stat.IsDir() {
		// 下载整个文件夹
		return s.downloadDirectory(info, savePath, progressChan)
	}

	// 下载单个文件
	return s.downloadSingleFile(info, savePath, progressChan)
}

// downloadSingleFile 下载单个文件
func (s *Service) downloadSingleFile(info *FileInfo, savePath string, progressChan chan float64) (*TransferResult, error) {
	srcFile, err := os.Open(info.Path)
	if err != nil {
		return nil, fmt.Errorf("打开源文件失败: %w", err)
	}
	defer srcFile.Close()

	destPath := filepath.Join(savePath, info.Name)
	destFile, err := os.Create(destPath)
	if err != nil {
		return nil, fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer destFile.Close()

	buffer := make([]byte, s.config.ChunkSize)
	var transferred int64
	startTime := time.Now()

	for {
		n, err := srcFile.Read(buffer)
		if n > 0 {
			_, writeErr := destFile.Write(buffer[:n])
			if writeErr != nil {
				return nil, fmt.Errorf("写入文件失败: %w", writeErr)
			}

			transferred += int64(n)

			// 下载文件时：从磁盘读取，发送给客户端
			stats.RecordSend(int64(n))
			stats.RecordDiskRead(int64(n))
			stats.RecordDiskWrite(int64(n))

			progress := float64(transferred) / float64(info.Size) * 100

			if progressChan != nil {
				select {
				case progressChan <- progress:
				default:
				}
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("读取文件失败: %w", err)
		}
	}

	duration := time.Since(startTime).Seconds()
	speed := float64(transferred) / (1024 * 1024) / duration

	result := &TransferResult{
		FileID:        info.ID,
		FileName:      info.Name,
		TotalSize:     info.Size,
		ReceivedBytes: transferred,
		Duration:      duration,
		Speed:         speed,
		Checksum:      info.Checksum,
	}

	logger.Info("文件下载完成: %s (%.2f MB/s)", info.Name, speed)
	return result, nil
}

// downloadDirectory 下载整个文件夹
func (s *Service) downloadDirectory(info *FileInfo, savePath string, progressChan chan float64) (*TransferResult, error) {
	// 创建目标文件夹
	destDir := filepath.Join(savePath, info.Name)
	if err := os.MkdirAll(destDir, 0755); err != nil {
		return nil, fmt.Errorf("创建目标文件夹失败: %w", err)
	}

	var totalTransferred int64
	var totalFiles int
	startTime := time.Now()

	// 遍历文件夹中的所有文件
	err := filepath.Walk(info.Path, func(path string, fileInfo os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录本身
		if path == info.Path {
			return nil
		}

		// 计算相对路径
		relPath, err := filepath.Rel(info.Path, path)
		if err != nil {
			return err
		}

		destPath := filepath.Join(destDir, relPath)

		if fileInfo.IsDir() {
			// 创建子目录
			return os.MkdirAll(destPath, fileInfo.Mode())
		}

		// 复制文件
		srcFile, err := os.Open(path)
		if err != nil {
			return fmt.Errorf("打开源文件失败 %s: %w", path, err)
		}

		// 确保目标目录存在
		if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
			srcFile.Close()
			return fmt.Errorf("创建目标目录失败: %w", err)
		}

		destFile, err := os.Create(destPath)
		if err != nil {
			srcFile.Close()
			return fmt.Errorf("创建目标文件失败 %s: %w", destPath, err)
		}

		// 复制文件内容
		buffer := make([]byte, s.config.ChunkSize)
		var fileTransferred int64
		for {
			n, err := srcFile.Read(buffer)
			if n > 0 {
				_, writeErr := destFile.Write(buffer[:n])
				if writeErr != nil {
					srcFile.Close()
					destFile.Close()
					return fmt.Errorf("写入文件失败: %w", writeErr)
				}
				fileTransferred += int64(n)
				totalTransferred += int64(n)

				// 记录统计
				stats.RecordSend(int64(n))
				stats.RecordDiskRead(int64(n))
				stats.RecordDiskWrite(int64(n))
			}
			if err == io.EOF {
				break
			}
			if err != nil {
				srcFile.Close()
				destFile.Close()
				return fmt.Errorf("读取文件失败: %w", err)
			}
		}

		srcFile.Close()
		destFile.Close()
		totalFiles++

		logger.Info("文件夹下载进度: %s -> %s (%d bytes)", path, destPath, fileTransferred)
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("下载文件夹失败: %w", err)
	}

	duration := time.Since(startTime).Seconds()
	speed := float64(totalTransferred) / (1024 * 1024) / duration

	result := &TransferResult{
		FileID:        info.ID,
		FileName:      info.Name,
		TotalSize:     info.Size,
		ReceivedBytes: totalTransferred,
		Duration:      duration,
		Speed:         speed,
		Checksum:      info.Checksum,
	}

	logger.Info("文件夹下载完成: %s (%d 个文件, %.2f MB/s)", info.Name, totalFiles, speed)
	return result, nil
}

func (s *Service) BatchDownload(fileIDs []string, savePath string, progressChan chan map[string]interface{}) (*BatchTransferResult, error) {
	results := make([]*TransferResult, 0, len(fileIDs))

	for i, id := range fileIDs {
		result, err := s.DownloadFile(id, savePath, nil)
		if err != nil {
			logger.Error("批量下载文件 %d 失败: %v", i+1, err)
			if progressChan != nil {
				progressChan <- map[string]interface{}{
					"index":    i,
					"file_id":  id,
					"progress": -1,
					"error":    err.Error(),
				}
			}
			continue
		}

		results = append(results, result)

		if progressChan != nil {
			progressChan <- map[string]interface{}{
				"index":          i,
				"file_id":        id,
				"file_name":      result.FileName,
				"total_size":     result.TotalSize,
				"received_bytes": result.ReceivedBytes,
				"progress":       100,
				"status":         "completed",
			}
		}
	}

	batchResult := &BatchTransferResult{
		TotalFiles: len(fileIDs),
		Completed:  len(results),
		Failed:     len(fileIDs) - len(results),
		Results:    results,
	}

	return batchResult, nil
}

func (s *Service) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (s *Service) getFileType(path string) string {
	ext := filepath.Ext(path)
	types := map[string]string{
		".txt":  "text",
		".pdf":  "pdf",
		".doc":  "document",
		".docx": "document",
		".xls":  "spreadsheet",
		".xlsx": "spreadsheet",
		".jpg":  "image",
		".jpeg": "image",
		".png":  "image",
		".gif":  "image",
		".mp4":  "video",
		".avi":  "video",
		".mkv":  "video",
		".mp3":  "audio",
		".wav":  "audio",
		".zip":  "archive",
		".rar":  "archive",
		".7z":   "archive",
		".exe":  "executable",
		".msi":  "executable",
	}

	if t, ok := types[ext]; ok {
		return t
	}
	return "unknown"
}

type TransferResult struct {
	FileID        string
	FileName      string
	TotalSize     int64
	ReceivedBytes int64
	Duration      float64
	Speed         float64
	Checksum      string
}

type BatchTransferResult struct {
	TotalFiles int
	Completed  int
	Failed     int
	Results    []*TransferResult
}
