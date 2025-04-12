package configs

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
}

type ServerConfig struct {
	Address         string        `yaml:"address"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type DatabaseConfig struct {
	Type    string        `yaml:"type"`
	MongoDB MongoDBConfig `yaml:"mongodb"`
}

type MongoDBConfig struct {
	URI      string `yaml:"uri"`
	Database string `yaml:"database"`
}

func Load() *Config {
	data, err := os.ReadFile("./database/internal/configs/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Error reading config file: %v", err))
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("Error parsing config: %v", err))
	}

	return &cfg
}
