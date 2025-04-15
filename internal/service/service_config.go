package service

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ilyakaznacheev/cleanenv"
)

type ServiceConfig struct {
	KafkaAddress string   `json:"port_addr"`
	Brokers      []string `json:"brokers"`
}

func NewServiceConfig() (*ServiceConfig, error) {
	cfg := &ServiceConfig{}

	// Get the CONFIG_PATH environment variable.
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		return nil, fmt.Errorf("CONFIG_PATH is not set")
	}

	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		return nil, fmt.Errorf("CONFIG_PATH conversion error: %v", err)
	}

	// Check if file exists.
	if _, err := os.Stat(absConfigPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %s", absConfigPath)
	}

	if err := cleanenv.ReadConfig(absConfigPath, cfg); err != nil {
		return nil, fmt.Errorf("cannot read config: %s", err)
	}

	return cfg, nil
}
