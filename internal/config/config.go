package config

import (
	"encoding/json"
	"os"
)

// Config represents the application configuration
type Config struct {
	Database  DatabaseConfig  `json:"database"`
	Cache     CacheConfig     `json:"cache"`
	Exchanges ExchangesConfig `json:"exchanges"`
	Server    ServerConfig    `json:"server"`
}

// DatabaseConfig represents PostgreSQL configuration
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Database string `json:"database"`
	SSLMode  string `json:"ssl_mode"`
}

// CacheConfig represents Redis configuration
type CacheConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	Password string `json:"password"`
	Database int    `json:"database"`
}

// ExchangesConfig represents exchange configuration
type ExchangesConfig struct {
	Exchange1 ExchangeConfig `json:"exchange1"`
	Exchange2 ExchangeConfig `json:"exchange2"`
	Exchange3 ExchangeConfig `json:"exchange3"`
	Test      ExchangeConfig `json:"test"`
}

// ExchangeConfig represents individual exchange configuration
type ExchangeConfig struct {
	Host string `json:"host"`
	Port int    `json:"port"`
}

// ServerConfig represents server configuration
type ServerConfig struct {
	Port int `json:"port"`
}

// Load loads configuration from file
func Load() (*Config, error) {
	configFile := "configs/config.json"
	if envFile := os.Getenv("CONFIG_FILE"); envFile != "" {
		configFile = envFile
	}

	data, err := os.ReadFile(configFile)
	if err != nil {
		return nil, err
	}

	var config Config
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return &config, nil
}
