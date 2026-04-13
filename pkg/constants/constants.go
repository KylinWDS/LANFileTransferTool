package constants

import "time"

const (
	AppName    = "LAN-File-Transfer-Tool"
	AppVersion = "0.2.0"

	DefaultPort      = 8080
	DefaultHost      = "0.0.0.0"
	DefaultChunkSize = 1048576

	TokenExpiryDownload = time.Hour * 24
	TokenExpiryUpload   = time.Hour * 1

	MaxConnections    = 10
	MaxHistoryRecords = 10

	StatusPending      = "pending"
	StatusTransferring = "transferring"
	StatusCompleted    = "completed"
	StatusFailed       = "failed"
	StatusPaused       = "paused"

	ActionUpload   = "upload"
	ActionDownload = "download"

	ThemeLight = "light"
	ThemeDark  = "dark"

	LangZhCN = "zh-CN"
	LangEn   = "en"
	LangRu   = "ru"
)
