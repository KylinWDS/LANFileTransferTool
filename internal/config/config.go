package config

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 应用配置结构体
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
	Enabled          bool `yaml:"enabled"`
	Port             int  `yaml:"port"`
	BroadcastInterval int  `yaml:"broadcast_interval"`
	PeerTimeout      int  `yaml:"peer_timeout"`
}

// WebSocketConfig WebSocket传输配置
type WebSocketConfig struct {
	Enabled   bool `yaml:"enabled"`
	PortOffset int  `yaml:"port_offset"`
	ChunkSize int  `yaml:"chunk_size"`
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

// DefaultConfig 创建默认配置
func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:      "LAN-File-Transfer-Tool",
			ShortName: "LANftt",
			Version:   "0.2.0",
		},
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		Discovery: DiscoveryConfig{
			Enabled:          true,
			Port:             37021,
			BroadcastInterval: 5,
			PeerTimeout:      30,
		},
		WebSocket: WebSocketConfig{
			Enabled:   true,
			PortOffset: 0,
			ChunkSize: 65536,
		},
		UDP: UDPConfig{
			Enabled:   true,
			Port:      37022,
			ChunkSize: 32768,
		},
		P2P: P2PConfig{
			Enabled:   true,
			Port:      37023,
			ChunkSize: 65536,
		},
		Transfer: TransferConfig{
			MaxConnections: 10,
			ChunkSize:      1048576,
			EnableResume:   true,
		},
		Security: SecurityConfig{
			TokenExpiry: 86400,
			SecretKey:   generateSecretKey(),
			Whitelist:   []string{},
			Blacklist:   []string{},
		},
		History: HistoryConfig{
			MaxRecords: 100,
		},
		Performance: PerformanceConfig{
			PoolSize:        10,
			MonitorInterval: 2,
		},
	}
}

// LoadConfig 从文件加载配置
func LoadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	config := DefaultConfig()
	if err := yaml.Unmarshal(data, config); err != nil {
		return nil, err
	}

	return config, nil
}

// Save 保存配置到文件
func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// generateSecretKey 生成随机密钥
func generateSecretKey() string {
	if key := os.Getenv("LANFTT_SECRET_KEY"); key != "" {
		return key
	}

	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "lanftt-default-secret-key-" + hex.EncodeToString([]byte(time.Now().String()))[:16]
	}
	return hex.EncodeToString(bytes)
}

// GetTokenExpiryDuration 获取Token过期时长
func GetTokenExpiryDuration() time.Duration {
	return time.Hour * 24
}

// GetDiscoveryPort 获取发现服务端口
func (c *Config) GetDiscoveryPort() int {
	if c.Discovery.Port <= 0 {
		return 37021
	}
	return c.Discovery.Port
}

// GetUDPPort 获取UDP传输端口
func (c *Config) GetUDPPort() int {
	if c.UDP.Port <= 0 {
		return 37022
	}
	return c.UDP.Port
}

// GetP2PPort 获取P2P传输端口
func (c *Config) GetP2PPort() int {
	if c.P2P.Port <= 0 {
		return 37023
	}
	return c.P2P.Port
}

// GetBroadcastInterval 获取广播间隔（秒）
func (c *Config) GetBroadcastInterval() time.Duration {
	if c.Discovery.BroadcastInterval <= 0 {
		return 5 * time.Second
	}
	return time.Duration(c.Discovery.BroadcastInterval) * time.Second
}

// GetPeerTimeout 获取节点超时时间（秒）
func (c *Config) GetPeerTimeout() time.Duration {
	if c.Discovery.PeerTimeout <= 0 {
		return 30 * time.Second
	}
	return time.Duration(c.Discovery.PeerTimeout) * time.Second
}
