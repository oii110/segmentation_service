package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Bot     BotConfig
	Storage StorageConfig
}

type BotConfig struct {
	Token   string
	AuthKey string
}

type StorageConfig struct {
	Username string
	Password string
	Host     string
	Port     int
	Database string
}

func LoadConfig() (*Config, error) {
	port, err := strconv.Atoi(getEnv("STORAGE_PORT", "5432"))
	if err != nil {
		return nil, fmt.Errorf("неверный порт базы данных: %w", err)
	}

	return &Config{
		Storage: StorageConfig{
			Username: getEnv("STORAGE_USERNAME", "postgres"),
			Password: getEnv("STORAGE_PASSWORD", "3535"),
			Host:     getEnv("STORAGE_HOST", "localhost"),
			Port:     port,
			Database: getEnv("STORAGE_DATABASE", "segment_service"),
		},
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultValue
}
