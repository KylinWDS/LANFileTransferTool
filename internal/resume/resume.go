package resume

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"lanfiletransfertool/pkg/errors"
	"lanfiletransfertool/pkg/logger"
	"lanfiletransfertool/pkg/utils"
)

type Service struct {
	dataDir    string
	resumeData map[string]*ResumeInfo
	mu         sync.RWMutex
}

type ResumeInfo struct {
	TransferID   string
	FilePath     string
	FileName     string
	TotalSize    int64
	Transferred  int64
	Checksum     string
	ChunkSize    int64
	LastModified time.Time
	CreatedAt    time.Time
	UpdatedAt    time.Time
	Status       string
}

func NewService(dataDir string) *Service {
	resumeDir := filepath.Join(dataDir, "resume")
	os.MkdirAll(resumeDir, 0755)

	s := &Service{
		dataDir:    resumeDir,
		resumeData: make(map[string]*ResumeInfo),
	}

	s.loadResumeData()
	return s
}

func (s *Service) SaveResumeInfo(info *ResumeInfo) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	info.UpdatedAt = time.Now()
	s.resumeData[info.TransferID] = info

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("序列化恢复信息失败: %w", err)
	}

	filePath := filepath.Join(s.dataDir, info.TransferID+".json")
	return os.WriteFile(filePath, data, 0644)
}

func (s *Service) GetResumeInfo(transferID string) (*ResumeInfo, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	info, exists := s.resumeData[transferID]
	if !exists {
		return nil, errors.ErrFileNotFound
	}
	return info, nil
}

func (s *Service) UpdateProgress(transferID string, transferred int64) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	info, exists := s.resumeData[transferID]
	if !exists {
		return errors.ErrFileNotFound
	}

	info.Transferred = transferred
	info.UpdatedAt = time.Now()

	data, err := json.Marshal(info)
	if err != nil {
		return fmt.Errorf("序列化更新信息失败: %w", err)
	}

	filePath := filepath.Join(s.dataDir, transferID+".json")
	return os.WriteFile(filePath, data, 0644)
}

func (s *Service) CanResume(transferID string, filePath string) (bool, error) {
	info, err := s.GetResumeInfo(transferID)
	if err != nil {
		return false, nil
	}

	currentInfo, err := os.Stat(filePath)
	if err != nil {
		return false, nil
	}

	if currentInfo.Size() != info.TotalSize {
		return false, nil
	}

	if currentInfo.ModTime().After(info.LastModified) {
		return false, nil
	}

	destPath := filepath.Join(filepath.Dir(filePath), info.FileName)
	destInfo, err := os.Stat(destPath)
	if err != nil {
		return true, nil
	}

	return destInfo.Size() < info.TotalSize, nil
}

func (s *Service) DeleteResumeInfo(transferID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.resumeData, transferID)

	filePath := filepath.Join(s.dataDir, transferID+".json")
	if utils.FileExists(filePath) {
		return os.Remove(filePath)
	}
	return nil
}

func (s *Service) CleanupExpiredResumes(maxAge time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for transferID, info := range s.resumeData {
		if now.Sub(info.UpdatedAt) > maxAge {
			delete(s.resumeData, transferID)
			filePath := filepath.Join(s.dataDir, transferID+".json")
			if utils.FileExists(filePath) {
				os.Remove(filePath)
			}
		}
	}

	return nil
}

func (s *Service) GetAllResumes() []*ResumeInfo {
	s.mu.RLock()
	defer s.mu.RUnlock()

	resumes := make([]*ResumeInfo, 0, len(s.resumeData))
	for _, info := range s.resumeData {
		resumes = append(resumes, info)
	}
	return resumes
}

func (s *Service) loadResumeData() {
	files, err := os.ReadDir(s.dataDir)
	if err != nil {
		return
	}

	for _, file := range files {
		if !file.IsDir() && filepath.Ext(file.Name()) == ".json" {
			filePath := filepath.Join(s.dataDir, file.Name())
			data, err := os.ReadFile(filePath)
			if err != nil {
				logger.Warn("加载恢复信息失败: %s, %v", file.Name(), err)
				continue
			}

			var info ResumeInfo
			if err := json.Unmarshal(data, &info); err != nil {
				logger.Warn("解析恢复信息失败: %s, %v", file.Name(), err)
				continue
			}

			s.resumeData[info.TransferID] = &info
		}
	}
}
