package encryption

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"strings"

	"golang.org/x/crypto/scrypt"
)

type Service struct{}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Encrypt(plainText string, key string) (string, error) {
	if strings.TrimSpace(plainText) == "" {
		return "", errors.New("明文不能为空")
	}
	if strings.TrimSpace(key) == "" {
		return "", errors.New("密钥不能为空")
	}

	salt := make([]byte, 16)
	if _, err := io.ReadFull(rand.Reader, salt); err != nil {
		return "", err
	}

	aesKey, err := scrypt.Key([]byte(key), salt, 32768, 8, 1, 32)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		return "", err
	}

	cipherText := aesGCM.Seal(nonce, nonce, []byte(plainText), nil)

	result := append(salt, cipherText...)
	return base64.StdEncoding.EncodeToString(result), nil
}

func (s *Service) Decrypt(cipherText string, key string) (string, error) {
	if strings.TrimSpace(cipherText) == "" {
		return "", errors.New("密文不能为空")
	}
	if strings.TrimSpace(key) == "" {
		return "", errors.New("密钥不能为空")
	}

	cipherText = strings.TrimSpace(cipherText)
	data, err := base64.StdEncoding.DecodeString(cipherText)
	if err != nil {
		return "", errors.New("密文格式错误：不是有效的Base64编码，请检查输入是否正确")
	}

	if len(data) < 48 {
		return "", errors.New("密文格式错误：数据长度不足，可能是密钥不匹配或密文已损坏")
	}

	salt := data[:16]
	ciphertext := data[16:]

	aesKey, err := scrypt.Key([]byte(key), salt, 32768, 8, 1, 32)
	if err != nil {
		return "", errors.New("密钥派生失败")
	}

	block, err := aes.NewCipher(aesKey)
	if err != nil {
		return "", errors.New("cipher初始化失败")
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", errors.New("GCM初始化失败")
	}

	nonceSize := aesGCM.NonceSize()
	if len(ciphertext) < nonceSize {
		return "", errors.New("密文格式错误：数据不完整")
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plainText, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", errors.New("解密失败：密钥不匹配或密文已损坏")
	}

	return string(plainText), nil
}

func (s *Service) GenerateKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

func (s *Service) ValidateKey(key string) bool {
	_, err := base64.StdEncoding.DecodeString(key)
	return err == nil && len(key) >= 32
}
