package token

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"sync"
	"time"

	"lanfiletransfertool/pkg/constants"
	"lanfiletransfertool/pkg/errors"
)

type Manager struct {
	secretKey     string
	defaultExpiry time.Duration
	tokens        map[string]*TokenInfo
	mu            sync.RWMutex
}

type TokenInfo struct {
	Token     string
	FileID    string
	ExpiresAt time.Time
	CreatedAt time.Time
	Type      string
}

type TokenData struct {
	FileID   string `json:"file_id"`
	Expiry   int64  `json:"expiry"`
	Type     string `json:"type"`
	Checksum string `json:"checksum"`
	// 文件元数据（支持跨重启解析）
	FileName string `json:"file_name,omitempty"`
	FileSize int64  `json:"file_size,omitempty"`
	FilePath string `json:"file_path,omitempty"`
}

type EncryptedToken struct {
	Data      []byte `json:"data"`
	Nonce     []byte `json:"nonce"`
	Timestamp int64  `json:"timestamp"`
}

func NewManager(defaultExpiry int, secretKey string) *Manager {
	return &Manager{
		secretKey:     padKey(secretKey),
		defaultExpiry: time.Duration(defaultExpiry) * time.Second,
		tokens:        make(map[string]*TokenInfo),
	}
}

func padKey(key string) string {
	if len(key) == 0 {
		key = constants.DefaultSecretKey
	}
	for len(key) < 32 {
		key += key
	}
	if len(key) > 32 {
		key = key[:32]
	}
	return key
}

func (m *Manager) GenerateToken(fileID string, expiry time.Duration) string {
	return m.GenerateTokenWithFileInfo(fileID, expiry, "", 0, "")
}

// GenerateTokenWithFileInfo 生成包含文件元数据的Token（支持跨重启解析）
func (m *Manager) GenerateTokenWithFileInfo(fileID string, expiry time.Duration, fileName string, fileSize int64, filePath string) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if expiry == 0 {
		expiry = constants.DefaultTokenExpiryDownload
	}

	tokenData := TokenData{
		FileID:   fileID,
		Expiry:   time.Now().Add(expiry).Unix(),
		Type:     "download",
		Checksum: generateChecksum(fileID),
		FileName: fileName,
		FileSize: fileSize,
		FilePath: filePath,
	}

	encryptedToken, err := m.encryptToken(&tokenData, m.secretKey)
	if err != nil {
		tokenStr := generateRandomString()
		m.tokens[tokenStr] = &TokenInfo{
			Token:     tokenStr,
			FileID:    fileID,
			ExpiresAt: time.Now().Add(expiry),
			CreatedAt: time.Now(),
			Type:      "download",
		}
		return tokenStr
	}

	m.tokens[encryptedToken] = &TokenInfo{
		Token:     encryptedToken,
		FileID:    fileID,
		ExpiresAt: time.Now().Add(expiry),
		CreatedAt: time.Now(),
		Type:      "download",
	}

	return encryptedToken
}

func (m *Manager) ValidateToken(token string) (*TokenData, error) {
	return m.ValidateTokenWithKey(token, "")
}

func (m *Manager) ValidateTokenWithKey(token, customKey string) (*TokenData, error) {
	// 首先尝试解密token（支持跨重启解析）
	key := m.secretKey
	if customKey != "" {
		key = padKey(customKey)
	}

	tokenData, err := m.decryptToken(token, key)
	if err == nil {
		// 解密成功，验证有效期和校验和
		if time.Now().Unix() > tokenData.Expiry {
			return nil, errors.ErrTokenExpired
		}

		expectedChecksum := generateChecksum(tokenData.FileID)
		if tokenData.Checksum != expectedChecksum {
			return nil, errors.ErrInvalidToken
		}

		return tokenData, nil
	}

	// 解密失败，尝试从内存中查找（向后兼容）
	m.mu.RLock()
	if tokenInfo, exists := m.tokens[token]; exists {
		m.mu.RUnlock()
		if time.Now().After(tokenInfo.ExpiresAt) {
			m.mu.Lock()
			delete(m.tokens, token)
			m.mu.Unlock()
			return nil, errors.ErrTokenExpired
		}
		return &TokenData{
			FileID: tokenInfo.FileID,
			Expiry: tokenInfo.ExpiresAt.Unix(),
			Type:   tokenInfo.Type,
		}, nil
	}
	m.mu.RUnlock()

	return nil, errors.ErrInvalidToken
}

func (m *Manager) ParseEncryptedToken(token, customKey string) (*TokenData, error) {
	key := m.secretKey
	if customKey != "" {
		key = padKey(customKey)
	}

	tokenData, err := m.decryptToken(token, key)
	if err != nil {
		return nil, fmt.Errorf("解密token失败: %w", err)
	}

	return tokenData, nil
}

func (m *Manager) encryptToken(data *TokenData, key string) (string, error) {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	ciphertext := gcm.Seal(nonce, nonce, jsonData, nil)

	return base64.URLEncoding.EncodeToString(ciphertext), nil
}

func (m *Manager) decryptToken(encryptedToken, key string) (*TokenData, error) {
	ciphertext, err := base64.URLEncoding.DecodeString(encryptedToken)
	if err != nil {
		ciphertext, err = base64.StdEncoding.DecodeString(encryptedToken)
		if err != nil {
			return nil, err
		}
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		return nil, fmt.Errorf("密文太短")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return nil, err
	}

	var tokenData TokenData
	if err := json.Unmarshal(plaintext, &tokenData); err != nil {
		return nil, err
	}

	return &tokenData, nil
}

func (m *Manager) RevokeToken(token string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.tokens[token]; !exists {
		return errors.ErrInvalidToken
	}

	delete(m.tokens, token)
	return nil
}

func (m *Manager) CleanupExpiredTokens() {
	m.mu.Lock()
	defer m.mu.Unlock()

	now := time.Now()
	for token, info := range m.tokens {
		if now.After(info.ExpiresAt) {
			delete(m.tokens, token)
		}
	}
}

func (m *Manager) GetActiveTokens() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var activeTokens []string
	for token, info := range m.tokens {
		if time.Now().Before(info.ExpiresAt) {
			activeTokens = append(activeTokens, token)
		}
	}
	return activeTokens
}

func generateRandomString() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return fmt.Sprintf("%d", time.Now().UnixNano())
	}
	return base64.URLEncoding.EncodeToString(bytes)
}

func generateChecksum(fileID string) string {
	hash := fmt.Sprintf("%x", len(fileID))
	if len(hash) < 8 {
		for len(hash) < 8 {
			hash += hash
		}
	}
	return hash[:8]
}

func GenerateUploadToken(fileID string, expiry time.Duration) string {
	return fmt.Sprintf("upload_%s_%d", fileID, time.Now().UnixNano())
}
