package checksum

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"hash"
	"io"
	"os"
	"strings"

	"lanfiletransfertool/pkg/errors"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) CalculateSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("计算哈希失败: %w", err)
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (s *Service) CalculateWithProgress(filePath string, progressChan chan float64) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", fmt.Errorf("获取文件信息失败: %w", err)
	}

	totalSize := fileInfo.Size()
	hasher := sha256.New()

	buffer := make([]byte, 32*1024)
	var readSize int64

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			hasher.Write(buffer[:n])
			readSize += int64(n)

			if progressChan != nil && totalSize > 0 {
				progress := float64(readSize) / float64(totalSize) * 100
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
			return "", fmt.Errorf("读取文件失败: %w", err)
		}
	}

	return hex.EncodeToString(hasher.Sum(nil)), nil
}

func (s *Service) VerifyFile(filePath string, expectedChecksum string) (bool, error) {
	actualChecksum, err := s.CalculateSHA256(filePath)
	if err != nil {
		return false, err
	}

	return actualChecksum == expectedChecksum, nil
}

func (s *Service) VerifyStream(reader io.Reader, expectedChecksum string, size int64) (bool, error) {
	hasher := sha256.New()

	if _, err := io.Copy(hasher, reader); err != nil && err != io.EOF {
		return false, err
	}

	actualChecksum := hex.EncodeToString(hasher.Sum(nil))
	match := actualChecksum == expectedChecksum

	if !match {
		return false, errors.ErrInvalidChecksum
	}

	return match, nil
}

func (s *Service) GetHashAlgorithm() hash.Hash {
	return sha256.New()
}

func (s *Service) FormatChecksum(checksum string) string {
	checksum = strings.TrimSpace(checksum)
	checksum = strings.ToLower(checksum)
	return checksum
}

func (s *Service) IsValidChecksum(checksum string) bool {
	checksum = s.FormatChecksum(checksum)
	_, err := hex.DecodeString(checksum)
	return err == nil && len(checksum) == 64
}
