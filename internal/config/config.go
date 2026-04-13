package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App       AppConfig       `yaml:"app"`
	Server    ServerConfig    `yaml:"server"`
	Transfer  TransferConfig  `yaml:"transfer"`
	Security  SecurityConfig  `yaml:"security"`
	History   HistoryConfig   `yaml:"history"`
}

type AppConfig struct {
	Name    string `yaml:"name"`
	Version string `yaml:"version"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type TransferConfig struct {
	MaxConnections int   `yaml:"max_connections"`
	ChunkSize      int64 `yaml:"chunk_size"`
	EnableResume   bool  `yaml:"enable_resume"`
}

type SecurityConfig struct {
	TokenExpiry int      `yaml:"token_expiry"`
	SecretKey   string   `yaml:"secret_key"`
	Whitelist   []string `yaml:"whitelist"`
	Blacklist   []string `yaml:"blacklist"`
}

type HistoryConfig struct {
	MaxRecords int `yaml:"max_records"`
}

func DefaultConfig() *Config {
	return &Config{
		App: AppConfig{
			Name:    "LAN File Transfer Tool",
			Version: "1.0.0",
		},
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
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
			MaxRecords: 10,
		},
	}
}

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

func (c *Config) Save(path string) error {
	data, err := yaml.Marshal(c)
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

func generateSecretKey() string {
	return os.Getenv("SECRET_KEY")
}

func GetTokenExpiryDuration() time.Duration {
	return time.Hour * 24
}
