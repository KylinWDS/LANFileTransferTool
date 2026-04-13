package constants

import "time"

// =============================================================================
// 应用基础信息 - 编译时常量
// =============================================================================
const (
	AppName    = "LAN-File-Transfer-Tool"
	AppVersion = "0.2.0"
)

// =============================================================================
// 状态/类型常量 - 用于状态标识和类型定义
// =============================================================================
const (
	// 传输状态
	StatusPending      = "pending"
	StatusTransferring = "transferring"
	StatusCompleted    = "completed"
	StatusFailed       = "failed"
	StatusPaused       = "paused"

	// 操作类型
	ActionUpload   = "upload"
	ActionDownload = "download"

	// 主题
	ThemeLight = "light"
	ThemeDark  = "dark"

	// 语言
	LangZhCN = "zh-CN"
	LangEn   = "en"
	LangRu   = "ru"

	// 协议类型
	ProtocolHTTP      = "http"
	ProtocolWebSocket = "websocket"
	ProtocolUDP       = "udp"
	ProtocolP2P       = "p2p"
)

// =============================================================================
// 默认配置值 - 所有可配置项的默认值
// 这些值可以被 config.yaml 中的用户配置覆盖
// =============================================================================

// 服务器默认配置
const (
	DefaultServerPort = 8080
	DefaultServerHost = "0.0.0.0"
)

// 发现服务默认配置
const (
	DefaultDiscoveryEnabled           = true
	DefaultDiscoveryPort              = 37021
	DefaultDiscoveryBroadcastInterval = 5  // 秒
	DefaultDiscoveryPeerTimeout       = 30 // 秒
)

// WebSocket默认配置
const (
	DefaultWebSocketEnabled    = true
	DefaultWebSocketPortOffset = 1
	DefaultWebSocketChunkSize  = 65536 // 64KB
)

// UDP默认配置
const (
	DefaultUDPEnabled   = true
	DefaultUDPPort      = 37022
	DefaultUDPChunkSize = 32768 // 32KB
)

// P2P默认配置
const (
	DefaultP2PEnabled   = true
	DefaultP2PPort      = 37023
	DefaultP2PChunkSize = 65536 // 64KB
)

// 传输默认配置
const (
	DefaultTransferMaxConnections = 10
	DefaultTransferChunkSize      = 1048576 // 1MB
	DefaultTransferEnableResume   = true
)

// 安全默认配置
const (
	DefaultTokenExpiry       = 86400 // 24小时，单位秒
	DefaultSecretKey         = "lanftt-default-secret-key-for-sharing"
	DefaultTokenExpiryDownload = time.Hour * 24
	DefaultTokenExpiryUpload   = time.Hour * 1
)

// 历史记录默认配置
const (
	DefaultMaxHistoryRecords = 100
)

// 性能默认配置
const (
	DefaultPoolSize        = 10
	DefaultMonitorInterval = 2 // 秒
)
