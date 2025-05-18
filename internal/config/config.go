package config

import (
	"errors"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/WebChads/SmsService/internal/pkg/slogerr"
	"github.com/ilyakaznacheev/cleanenv"
)

type ServerConfig struct {
	KafkaAddress string   `yaml:"kafka_address" env:"KAFKA_ADDRESS" env-default:"localhost:9092"`
	Brokers      []string `yaml:"brokers" env:"BROKERS" env-separator:","`
}

// NewServiceConfig loads configuration with the following precedence:
// 1. Reads from config file (if exists, but not required)
// 2. Overrides with environment variables
// 3. Validates final configuration
func NewServiceConfig() (*ServerConfig, error) {
	cfg := &ServerConfig{}

	// Try to read from config file first
	configPath := getConfigPath()
	if configPath != "" {
		absConfigPath, err := filepath.Abs(configPath)
		if err != nil {
			slog.Warn("failed to get absolute config path", slogerr.Error(err))
		} else {
			if _, err := os.Stat(absConfigPath); err == nil {
				if err := cleanenv.ReadConfig(absConfigPath, cfg); err != nil {
					slog.Warn("failed to read config file", slogerr.Error(err))
				} else {
					slog.Info("loaded configuration from file", "path", absConfigPath)
				}
			} else {
				slog.Warn("config file not found", "path", absConfigPath)
			}
		}
	}

	// Override with environment variables
	if err := cleanenv.ReadEnv(cfg); err != nil {
		return nil, errors.New("failed to read environment variables: " + err.Error())
	}

	// Validate final configuration
	if err := validateConfig(cfg); err != nil {
		return nil, errors.New("configuration validation failed: " + err.Error())
	}

	return cfg, nil
}

// getConfigPath determines the config file path with fallbacks:
// 1. CONFIG_PATH environment variable
// 2. Default config file in project's configs directory
func getConfigPath() string {
	if configPath := os.Getenv("CONFIG_PATH"); configPath != "" {
		return configPath
	}

	root, err := findModuleRoot(".")
	if err != nil {
		slog.Warn("failed to find project root", slogerr.Error(err))
		return ""
	}

	return filepath.Join(root, "configs", "config.yaml")
}

func findModuleRoot(dir string) (string, error) {
	for {
		if dir == "" || dir == "/" {
			return "", errors.New("invalid working directory name")
		}

		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		dir = filepath.Dir(dir)
	}
}

func validateConfig(cfg *ServerConfig) error {
	if cfg.KafkaAddress == "" {
		return errors.New("kafka address is required")
	}
	if len(cfg.Brokers) == 0 {
		return errors.New("at least one broker is required")
	}
	return nil
}
