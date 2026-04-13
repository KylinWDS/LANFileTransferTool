package token

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sync"
	"time"

	"lanfiletransfertool/pkg/constants"
	"lanfiletransfertool/pkg/errors"
	"lanfiletransfertool/pkg/utils"
)

type Manager struct {
	secretKey  string
	defaultExpiry time.Duration
	tokens     map[string]*TokenInfo
	mu         sync.RWMutex
}

type TokenInfo struct {
	Token     string
	FileID    string
	ExpiresAt time.Time
	CreatedAt time.Time
	Type      string
}

type TokenData struct {
	FileID string
	Expiry time.Time
}

func NewManager(defaultExpiry int, secretKey string) *Manager {
	return &Manager{
		secretKey:    secretKey,
		defaultExpiry: time.Duration(defaultExpiry) * time.Second,
		tokens:       make(map[string]*TokenInfo),
	}
}

func (m *Manager) GenerateToken(fileID string, expiry time.Duration) string {
	m.mu.Lock()
	defer m.mu.Unlock()

	if expiry == 0 {
		expiry = constants.TokenExpiryDownload
	}

	tokenStr := generateTokenString()

	tokenInfo := &TokenInfo{
		Token:     tokenStr,
		FileID:    fileID,
		ExpiresAt: time.Now().Add(expiry),
		CreatedAt: time.Now(),
		Type:      "download",
	}

	m.tokens[tokenStr] = tokenInfo
	return tokenStr
}

func (m *Manager) ValidateToken(token string) (*TokenData, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	tokenInfo, exists := m.tokens[token]
	if !exists {
		return nil, errors.ErrInvalidToken
	}

	if time.Now().After(tokenInfo.ExpiresAt) {
		delete(m.tokens, token)
		return nil, errors.ErrTokenExpired
	}

	return &TokenData{
		FileID: tokenInfo.FileID,
		Expiry: tokenInfo.ExpiresAt,
	}, nil
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

func generateTokenString() string {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return utils.GenerateID()
	}
	return hex.EncodeToString(bytes)
}

func signToken(token string, secret string) string {
	h := hmac.New(sha256.New, []byte(secret))
	h.Write([]byte(token))
	return hex.EncodeToString(h.Sum(nil))
}

func verifySignature(token, signature, secret string) bool {
	expectedSign := signToken(token, secret)
	return hmac.Equal([]byte(signature), []byte(expectedSign))
}

func GenerateUploadToken(fileID string, expiry time.Duration) string {
	return fmt.Sprintf("upload_%s_%d", fileID, time.Now().UnixNano())
}
