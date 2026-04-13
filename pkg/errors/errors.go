package errors

import "errors"

var (
	ErrFileNotFound      = errors.New("file not found")
	ErrInvalidToken      = errors.New("invalid token")
	ErrTokenExpired      = errors.New("token expired")
	ErrAccessDenied      = errors.New("access denied")
	ErrFileAlreadyExists = errors.New("file already exists")
	ErrInvalidChecksum   = errors.New("invalid checksum")
	ErrEncryptionFailed  = errors.New("encryption failed")
	ErrDecryptionFailed  = errors.New("decryption failed")
	ErrTransferFailed    = errors.New("transfer failed")
	ErrResumeNotSupported = errors.New("resume not supported")
	ErrConfigNotFound    = errors.New("config not found")
	ErrDatabaseError     = errors.New("database error")
	ErrInvalidParameter  = errors.New("invalid parameter")
	ErrServerNotRunning  = errors.New("server not running")
	ErrPortInUse         = errors.New("port in use")
	ErrNetworkError      = errors.New("network error")
)
