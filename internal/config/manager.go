package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"lanfiletransfertool/pkg/logger"
	"lanfiletransfertool/pkg/utils"
)

// Manager 配置管理器，负责管理配置的加载顺序和优先级
type Manager struct {
	mu         sync.RWMutex
	systemCfg  *Config
	userCfg    map[string]interface{}
	mergedCfg  *Config
	configPath string
	userCfgPath string
}

// NewManager 创建配置管理器
// 配置加载顺序：
// 1. 初始化默认配置
// 2. 加载系统配置（config.yaml）
// 3. 加载用户配置（user_config.json）覆盖系统配置
// 4. 应用临时配置（如果有）
func NewManager() (*Manager, error) {
	m := &Manager{
		systemCfg: DefaultConfig(),
		userCfg:   make(map[string]interface{}),
		mergedCfg: DefaultConfig(),
	}

	// 设置配置路径 (使用用户可写目录)
	appDataDir := utils.GetAppDataDir()
	os.MkdirAll(appDataDir, 0755)
	m.configPath = filepath.Join(appDataDir, "config.yaml")
	m.userCfgPath = filepath.Join(appDataDir, "user_config.json")
	logger.Info("配置路径: %s", m.configPath)
	logger.Info("用户配置路径: %s", m.userCfgPath)

	// 执行配置加载
	if err := m.Load(); err != nil {
		logger.Warn("配置加载过程中出现错误: %v", err)
	}

	return m, nil
}

// Load 执行完整的配置加载流程
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// 步骤1: 初始化默认配置
	m.systemCfg = DefaultConfig()
	m.mergedCfg = DefaultConfig()

	logger.Info("步骤1: 初始化默认配置完成")

	// 步骤2: 加载系统配置（config.yaml）
	if err := m.loadSystemConfig(); err != nil {
		logger.Warn("加载系统配置失败，使用默认配置: %v", err)
	} else {
		logger.Info("步骤2: 加载系统配置完成")
	}

	// 步骤3: 加载用户配置并覆盖系统配置
	if err := m.loadUserConfig(); err != nil {
		logger.Warn("加载用户配置失败: %v", err)
	} else {
		logger.Info("步骤3: 加载用户配置并覆盖系统配置完成")
	}

	// 步骤4: 合并配置
	m.mergeConfigs()

	logger.Info("配置加载完成 - 端口: %d, 版本: %s", m.mergedCfg.Server.Port, m.mergedCfg.App.Version)
	return nil
}

// loadSystemConfig 加载系统配置文件
func (m *Manager) loadSystemConfig() error {
	_, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 配置文件不存在，创建默认配置
			logger.Info("系统配置文件不存在，创建默认配置: %s", m.configPath)
			return m.systemCfg.Save(m.configPath)
		}
		return fmt.Errorf("读取系统配置文件失败: %w", err)
	}

	// 解析YAML配置
	loadedCfg, err := LoadConfig(m.configPath)
	if err != nil {
		return fmt.Errorf("解析系统配置文件失败: %w", err)
	}

	m.systemCfg = loadedCfg
	return nil
}

// loadUserConfig 加载用户配置文件
func (m *Manager) loadUserConfig() error {
	data, err := os.ReadFile(m.userCfgPath)
	if err != nil {
		if os.IsNotExist(err) {
			// 用户配置文件不存在，这是正常的
			logger.Info("用户配置文件不存在，使用系统配置")
			return nil
		}
		return fmt.Errorf("读取用户配置文件失败: %w", err)
	}

	var userConfig map[string]interface{}
	if err := json.Unmarshal(data, &userConfig); err != nil {
		return fmt.Errorf("解析用户配置文件失败: %w", err)
	}

	m.userCfg = userConfig
	return nil
}

// mergeConfigs 合并系统配置和用户配置
func (m *Manager) mergeConfigs() {
	// 从系统配置开始
	*m.mergedCfg = *m.systemCfg

	// 应用用户配置覆盖
	if theme, ok := m.userCfg["theme"].(string); ok && theme != "" {
		// 主题保存在用户配置中，不在系统配置中
	}

	if settings, ok := m.userCfg["settings"].(map[string]interface{}); ok {
		// 应用服务器设置
		if server, ok := settings["server"].(map[string]interface{}); ok {
			if port, ok := server["port"].(float64); ok {
				m.mergedCfg.Server.Port = int(port)
			}
		}

		// 应用传输设置
		if transfer, ok := settings["transfer"].(map[string]interface{}); ok {
			if chunkSize, ok := transfer["chunkSize"].(float64); ok {
				m.mergedCfg.Transfer.ChunkSize = int64(chunkSize)
			}
			if maxConn, ok := transfer["maxConnections"].(float64); ok {
				m.mergedCfg.Transfer.MaxConnections = int(maxConn)
			}
			if enableResume, ok := transfer["enableResume"].(bool); ok {
				m.mergedCfg.Transfer.EnableResume = enableResume
			}
		}

		// 应用协议设置
		if protocols, ok := settings["protocols"].(map[string]interface{}); ok {
			if ws, ok := protocols["websocket"].(bool); ok {
				m.mergedCfg.WebSocket.Enabled = ws
			}
			if udp, ok := protocols["udp"].(bool); ok {
				m.mergedCfg.UDP.Enabled = udp
			}
			if p2p, ok := protocols["p2p"].(bool); ok {
				m.mergedCfg.P2P.Enabled = p2p
			}
			if discovery, ok := protocols["discovery"].(bool); ok {
				m.mergedCfg.Discovery.Enabled = discovery
			}
		}

		// 应用安全设置
		if security, ok := settings["security"].(map[string]interface{}); ok {
			if tokenExpiry, ok := security["tokenExpiry"].(float64); ok {
				m.mergedCfg.Security.TokenExpiry = int(tokenExpiry)
			}
			if enableEncryption, ok := security["enableEncryption"].(bool); ok {
				// 加密设置
				_ = enableEncryption
			}
		}

		// 应用历史记录设置
		if history, ok := settings["history"].(map[string]interface{}); ok {
			if maxRecords, ok := history["maxRecords"].(float64); ok {
				m.mergedCfg.History.MaxRecords = int(maxRecords)
			}
		}
	}
}

// GetConfig 获取合并后的配置（只读）
func (m *Manager) GetConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// 返回配置的副本，防止外部修改
	cfgCopy := *m.mergedCfg
	return &cfgCopy
}

// GetSystemConfig 获取系统配置（只读）
func (m *Manager) GetSystemConfig() *Config {
	m.mu.RLock()
	defer m.mu.RUnlock()

	cfgCopy := *m.systemCfg
	return &cfgCopy
}

// UpdateSystemConfig 更新系统配置
func (m *Manager) UpdateSystemConfig(cfg *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.systemCfg = cfg
	m.mergeConfigs()

	return m.systemCfg.Save(m.configPath)
}

// ApplyTemporaryConfig 应用临时配置（最高优先级）
func (m *Manager) ApplyTemporaryConfig(key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch key {
	case "server.port":
		if port, ok := value.(int); ok {
			m.mergedCfg.Server.Port = port
		}
	case "transfer.chunk_size":
		if size, ok := value.(int64); ok {
			m.mergedCfg.Transfer.ChunkSize = size
		}
	case "transfer.max_connections":
		if max, ok := value.(int); ok {
			m.mergedCfg.Transfer.MaxConnections = max
		}
	default:
		return fmt.Errorf("未知的临时配置项: %s", key)
	}

	return nil
}

// Reload 重新加载所有配置
func (m *Manager) Reload() error {
	return m.Load()
}

// SaveSystemConfig 保存系统配置到文件
func (m *Manager) SaveSystemConfig() error {
	m.mu.RLock()
	defer m.mu.RUnlock()

	return m.systemCfg.Save(m.configPath)
}

// GetConfigPath 获取系统配置文件路径
func (m *Manager) GetConfigPath() string {
	return m.configPath
}

// GetUserConfigPath 获取用户配置文件路径
func (m *Manager) GetUserConfigPath() string {
	return m.userCfgPath
}
