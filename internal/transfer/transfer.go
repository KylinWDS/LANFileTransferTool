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
		FileID:        id,
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
