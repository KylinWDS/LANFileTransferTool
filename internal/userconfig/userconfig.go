package userconfig

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"

	"lanfiletransfertool/pkg/constants"
)

type Manager struct {
	configPath string
	config     *Config
	mu         sync.RWMutex
}

type Config struct {
	Theme    string                 `json:"theme"`
	Language string                 `json:"language"`
	Settings map[string]interface{} `json:"settings"`
}

const defaultTheme = constants.ThemeLight
const defaultLanguage = constants.LangZhCN

func NewManager(configPath string) (*Manager, error) {
	m := &Manager{
		configPath: configPath,
	}

	if err := m.loadOrCreate(); err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) loadOrCreate() error {
	data, err := os.ReadFile(m.configPath)
	if err != nil {
		if os.IsNotExist(err) {
			m.config = m.defaultConfig()
			return m.save()
		}
		return err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		m.config = m.defaultConfig()
		return m.save()
	}

	m.config = &config
	m.applyDefaults()
	return nil
}

func (m *Manager) GetConfig() (*Config, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config == nil {
		return nil, os.ErrNotExist
	}

	configCopy := *m.config
	if configCopy.Settings != nil {
		settingsCopy := make(map[string]interface{})
		for k, v := range configCopy.Settings {
			settingsCopy[k] = v
		}
		configCopy.Settings = settingsCopy
	}

	return &configCopy, nil
}

func (m *Manager) SaveConfig(cfg *Config) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = cfg
	return m.save()
}

func (m *Manager) ResetConfig() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.config = m.defaultConfig()
	return m.save()
}

func (m *Manager) UpdateTheme(theme string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if theme != constants.ThemeLight && theme != constants.ThemeDark {
		theme = defaultTheme
	}

	m.config.Theme = theme
	return m.save()
}

func (m *Manager) UpdateLanguage(lang string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if lang != constants.LangZhCN && lang != constants.LangEn && lang != constants.LangRu {
		lang = defaultLanguage
	}

	m.config.Language = lang
	return m.save()
}

func (m *Manager) UpdateSetting(key string, value interface{}) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.Settings == nil {
		m.config.Settings = make(map[string]interface{})
	}

	m.config.Settings[key] = value
	return m.save()
}

func (m *Manager) GetSetting(key string) (interface{}, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config.Settings == nil {
		return nil, false
	}

	value, exists := m.config.Settings[key]
	return value, exists
}

func (m *Manager) save() error {
	dir := filepath.Dir(m.configPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}

	data, err := json.MarshalIndent(m.config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(m.configPath, data, 0644)
}

func (m *Manager) defaultConfig() *Config {
	return &Config{
		Theme:    defaultTheme,
		Language: defaultLanguage,
		Settings: map[string]interface{}{
			"auto_start_server": true,
			"default_port":      8080,
			"chunk_size":        1048576,
			"max_connections":   10,
			"enable_resume":     true,
			"window_width":      1200,
			"window_height":     800,
		},
	}
}

func (m *Manager) applyDefaults() {
	if m.config.Theme == "" {
		m.config.Theme = defaultTheme
	}
	if m.config.Language == "" {
		m.config.Language = defaultLanguage
	}
	if m.config.Settings == nil {
		m.config.Settings = make(map[string]interface{})
	}
}

func (m *Manager) GetSelectedIP() string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.config.Settings == nil {
		return ""
	}

	if ip, ok := m.config.Settings["selected_ip"].(string); ok {
		return ip
	}
	return ""
}

func (m *Manager) SetSelectedIP(ip string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.config.Settings == nil {
		m.config.Settings = make(map[string]interface{})
	}

	m.config.Settings["selected_ip"] = ip
	return m.save()
}
