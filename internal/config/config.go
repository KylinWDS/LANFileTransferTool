package config

import (
	"os"
	"time"

	"lanfiletransfertool/pkg/constants"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构体 - 对应 config.yaml 的完整结构
type Config struct {
	App         AppConfig         `yaml:"app"`
	Server      ServerConfig      `yaml:"server"`
	Discovery   DiscoveryConfig   `yaml:"discovery"`
	WebSocket   WebSocketConfig   `yaml:"websocket"`
	UDP         UDPConfig         `yaml:"udp"`
	P2P         P2PConfig         `yaml:"p2p"`
	Transfer    TransferConfig    `yaml:"transfer"`
	Security    SecurityConfig    `yaml:"security"`
	History     HistoryConfig     `yaml:"history"`
	Performance PerformanceConfig `yaml:"performance"`
}

// AppConfig 应用基础配置
type AppConfig struct {
	Name      string `yaml:"name"`
	ShortName string `yaml:"short_name"`
	Version   string `yaml:"version"`
}

// ServerConfig HTTP服务器配置
type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

// DiscoveryConfig UDP发现服务配置
type DiscoveryConfig struct {
	Enabled           bool `yaml:"enabled"`
	Port              int  `yaml:"port"`
	BroadcastInterval int  `yaml:"broadcast_interval"`
	PeerTimeout       int  `yaml:"peer_timeout"`
}

// WebSocketConfig WebSocket传输配置
type WebSocketConfig struct {
	Enabled    bool `yaml:"enabled"`
	PortOffset int  `yaml:"port_offset"`
	ChunkSize  int  `yaml:"chunk_size"`
}

// UDPConfig UDP传输配置
type UDPConfig struct {
	Enabled   bool `yaml:"enabled"`
	Port      int  `yaml:"port"`
	ChunkSize int  `yaml:"chunk_size"`
}

// P2PConfig P2P传输配置
type P2PConfig struct {
	Enabled   bool `yaml:"enabled"`
	Port      int  `yaml:"port"`
	ChunkSize int  `yaml:"chunk_size"`
}

// TransferConfig 文件传输配置
type TransferConfig struct {
	MaxConnections int   `yaml:"max_connections"`
	ChunkSize      int64 `yaml:"chunk_size"`
	EnableResume   bool  `yaml:"enable_resume"`
}

// SecurityConfig 安全配置
type SecurityConfig struct {
	TokenExpiry int      `yaml:"token_expiry"`
	SecretKey   string   `yaml:"secret_key"`
	Whitelist   []string `yaml:"whitelist"`
	Blacklist   []string `yaml:"blacklist"`
}

// HistoryConfig 历史记录配置
type HistoryConfig struct {
	MaxRecords int `yaml:"max_records"`
}

// PerformanceConfig 性能配置
type PerformanceConfig struct {
	PoolSize        int `yaml:"pool_size"`
	MonitorInterval int `yaml:"monitor_interval"`
}

// =============================================================================
// 配置加载与默认值处理
// 所有默认值现在都在 constants.go 中定义
// 优先级：config.yaml > constants.DefaultXXX
// =============================================================================

// DefaultConfig 创建默认配置
// 所有默认值都来自 constants.go
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:      constants.AppName,
			ShortName: "LANftt",
			Version:   constants.AppVersion,
		},
		Server: ServerConfig{
			Port: constants.DefaultServerPort,
			Host: constants.DefaultServerHost,
		},
		Discovery: DiscoveryConfig{
			Enabled:           constants.DefaultDiscoveryEnabled,
			Port:              constants.DefaultDiscoveryPort,
			BroadcastInterval: constants.DefaultDiscoveryBroadcastInterval,
			PeerTimeout:       constants.DefaultDiscoveryPeerTimeout,
		},
		WebSocket: WebSocketConfig{
			Enabled:    constants.DefaultWebSocketEnabled,
			PortOffset: constants.DefaultWebSocketPortOffset,
			ChunkSize:  constants.DefaultWebSocketChunkSize,
		},
		UDP: UDPConfig{
			Enabled:   constants.DefaultUDPEnabled,
			Port:      constants.DefaultUDPPort,
			ChunkSize: constants.DefaultUDPChunkSize,
		},
		P2P: P2PConfig{
			Enabled:   constants.DefaultP2PEnabled,
			Port:      constants.DefaultP2PPort,
			ChunkSize: constants.DefaultP2PChunkSize,
		},
		Transfer: TransferConfig{
			MaxConnections: constants.DefaultTransferMaxConnections,
			ChunkSize:      constants.DefaultTransferChunkSize,
			EnableResume:   constants.DefaultTransferEnableResume,
		},
		Security: SecurityConfig{
			TokenExpiry: constants.DefaultTokenExpiry,
			SecretKey:   "", // 空字符串表示使用 constants.DefaultSecretKey
			Whitelist:   []string{},
			Blacklist:   []string{},
		},
		History: HistoryConfig{
			MaxRecords: constants.DefaultMaxHistoryRecords,
		},
		Performance: PerformanceConfig{
			PoolSize:        constants.DefaultPoolSize,
			MonitorInterval: constants.DefaultMonitorInterval,
		},
	}
}

// LoadConfig 从文件加载配置
// 加载优先级：用户配置文件 > constants.DefaultXXX
func LoadConfig(path string) (*Config, error) {
	// 1. 从 constants 默认值开始
	config := DefaultConfig()

	// 2. 读取配置文件（如果存在）
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，返回默认值
			config.Security.SecretKey = constants.DefaultSecretKey
			return config, nil
		}
		return nil, err
	}

	// 3. 解析配置文件，覆盖默认值
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	// 4. 特殊处理：如果配置文件中没有设置 secret_key，使用默认密钥
	if config.Security.SecretKey == "" {
		config.Security.SecretKey = constants.DefaultSecretKey
	}

	// 5. 确保数值配置有效（防止配置文件中设置为0或负数）
	config.applySafetyDefaults()

	return config, nil
}

// applySafetyDefaults 应用安全默认值（防止无效配置）
// 使用 constants 中的值作为安全回退
func (c *Config) applySafetyDefaults() {
	// 服务器配置
	if c.Server.Port <= 0 {
		c.Server.Port = constants.DefaultServerPort
	}
	if c.Server.Host == "" {
		c.Server.Host = constants.DefaultServerHost
	}

	// 发现服务配置
	if c.Discovery.Port <= 0 {
		c.Discovery.Port = constants.DefaultDiscoveryPort
	}
	if c.Discovery.BroadcastInterval <= 0 {
		c.Discovery.BroadcastInterval = constants.DefaultDiscoveryBroadcastInterval
	}
	if c.Discovery.PeerTimeout <= 0 {
		c.Discovery.PeerTimeout = constants.DefaultDiscoveryPeerTimeout
	}

	// WebSocket配置
	if c.WebSocket.ChunkSize <= 0 {
		c.WebSocket.ChunkSize = constants.DefaultWebSocketChunkSize
	}

	// UDP配置
	if c.UDP.Port <= 0 {
		c.UDP.Port = constants.DefaultUDPPort
	}
	if c.UDP.ChunkSize <= 0 {
		c.UDP.ChunkSize = constants.DefaultUDPChunkSize
	}

	// P2P配置
	if c.P2P.Port <= 0 {
		c.P2P.Port = constants.DefaultP2PPort
	}
	if c.P2P.ChunkSize <= 0 {
		c.P2P.ChunkSize = constants.DefaultP2PChunkSize
	}

	// 传输配置
	if c.Transfer.MaxConnections <= 0 {
		c.Transfer.MaxConnections = constants.DefaultTransferMaxConnections
	}
	if c.Transfer.ChunkSize <= 0 {
		c.Transfer.ChunkSize = constants.DefaultTransferChunkSize
	}

	// 安全配置
	if c.Security.TokenExpiry <= 0 {
		c.Security.TokenExpiry = constants.DefaultTokenExpiry
	}
	if c.Security.SecretKey == "" {
		c.Security.SecretKey = constants.DefaultSecretKey
	}

	// 历史记录配置
	if c.History.MaxRecords <= 0 {
		c.History.MaxRecords = constants.DefaultMaxHistoryRecords
	}

	// 性能配置
	if c.Performance.PoolSize <= 0 {
		c.Performance.PoolSize = constants.DefaultPoolSize
	}
	if c.Performance.MonitorInterval <= 0 {
		c.Performance.MonitorInterval = constants.DefaultMonitorInterval
	}
}

// Save 保存配置到文件
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// =============================================================================
// 配置获取方法 - 提供安全的配置值访问
// 所有方法都使用 constants 作为回退值
// =============================================================================

// GetSecretKey 获取密钥
func (c *Config) GetSecretKey() string {
	if c.Security.SecretKey == "" {
		return constants.DefaultSecretKey
	}
	return c.Security.SecretKey
}

// GetTokenExpiryDuration 获取Token过期时长
func (c *Config) GetTokenExpiryDuration() time.Duration {
	if c.Security.TokenExpiry <= 0 {
		return constants.DefaultTokenExpiryDownload
	}
	return time.Duration(c.Security.TokenExpiry) * time.Second
}

// GetServerPort 获取服务器端口
func (c *Config) GetServerPort() int {
	if c.Server.Port <= 0 {
		return constants.DefaultServerPort
	}
	return c.Server.Port
}

// GetServerHost 获取服务器主机
func (c *Config) GetServerHost() string {
	if c.Server.Host == "" {
		return constants.DefaultServerHost
	}
	return c.Server.Host
}

// GetDiscoveryPort 获取发现服务端口
func (c *Config) GetDiscoveryPort() int {
	if c.Discovery.Port <= 0 {
		return constants.DefaultDiscoveryPort
	}
	return c.Discovery.Port
}

// GetUDPPort 获取UDP传输端口
func (c *Config) GetUDPPort() int {
	if c.UDP.Port <= 0 {
		return constants.DefaultUDPPort
	}
	return c.UDP.Port
}

// GetP2PPort 获取P2P传输端口
func (c *Config) GetP2PPort() int {
	if c.P2P.Port <= 0 {
		return constants.DefaultP2PPort
	}
	return c.P2P.Port
}

// GetBroadcastInterval 获取广播间隔
func (c *Config) GetBroadcastInterval() time.Duration {
	if c.Discovery.BroadcastInterval <= 0 {
		return time.Duration(constants.DefaultDiscoveryBroadcastInterval) * time.Second
	}
	return time.Duration(c.Discovery.BroadcastInterval) * time.Second
}

// GetPeerTimeout 获取节点超时时间
func (c *Config) GetPeerTimeout() time.Duration {
	if c.Discovery.PeerTimeout <= 0 {
		return time.Duration(constants.DefaultDiscoveryPeerTimeout) * time.Second
	}
	return time.Duration(c.Discovery.PeerTimeout) * time.Second
}

// GetTransferChunkSize 获取传输分片大小
func (c *Config) GetTransferChunkSize() int64 {
	if c.Transfer.ChunkSize <= 0 {
		return constants.DefaultTransferChunkSize
	}
	return c.Transfer.ChunkSize
}

// GetMaxConnections 获取最大连接数
func (c *Config) GetMaxConnections() int {
	if c.Transfer.MaxConnections <= 0 {
		return constants.DefaultTransferMaxConnections
	}
	return c.Transfer.MaxConnections
}

// GetMaxHistoryRecords 获取最大历史记录数
func (c *Config) GetMaxHistoryRecords() int {
	if c.History.MaxRecords <= 0 {
		return constants.DefaultMaxHistoryRecords
	}
	return c.History.MaxRecords
}

// GetPoolSize 获取线程池大小
func (c *Config) GetPoolSize() int {
	if c.Performance.PoolSize <= 0 {
		return constants.DefaultPoolSize
	}
	return c.Performance.PoolSize
}

// GetMonitorInterval 获取监控间隔
func (c *Config) GetMonitorInterval() time.Duration {
	if c.Performance.MonitorInterval <= 0 {
		return time.Duration(constants.DefaultMonitorInterval) * time.Second
	}
	return time.Duration(c.Performance.MonitorInterval) * time.Second
}
