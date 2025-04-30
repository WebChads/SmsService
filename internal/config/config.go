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
	KafkaAddress string   `yaml:"kafka_address"`
	Brokers      []string `yaml:"brokers"`
}

func NewServiceConfig() *ServerConfig {

	// Get the CONFIG_PATH environment variable.
	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		root, err := FindModuleRoot(".")
		if err != nil {
			slog.Error("failed to find config file", slogerr.Error(err))
			return nil
		}

		configPath = root + "/configs/local.yaml"
	}

	absConfigPath, err := filepath.Abs(configPath)
	if err != nil {
		slog.Error("CONFIG_PATH conversion error", slogerr.Error(err))
		return nil
	}

	// Check if file exists.
	if _, err := os.Stat(absConfigPath); os.IsNotExist(err) {
		slog.Error("config file does not exist: " + absConfigPath)
		return nil
	}

	var cfg ServerConfig

	if err := cleanenv.ReadConfig(absConfigPath, &cfg); err != nil {
		slog.Error("cannot read config", slogerr.Error(err))
		return nil
	}

	return &cfg
}

func FindModuleRoot(dir string) (string, error) {
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
