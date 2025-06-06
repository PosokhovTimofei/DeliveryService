package configs

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig   `yaml:"server"`
	Database   DatabaseConfig `yaml:"database"`
	Telegram   Telegram       `yaml:"telegram"`
	GrpcConfig GRPCConfig     `yaml:"clients"`
}

type ServerConfig struct {
	Address         string        `yaml:"address"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout"`
}

type Telegram struct {
	TelegramToken string `yaml:"token"`
}

type DatabaseConfig struct {
	Type    string        `yaml:"type"`
	MongoDB MongoDBConfig `yaml:"mongodb"`
}

type MongoDBConfig struct {
	URI      string `yaml:"uri"`
	Database string `yaml:"database"`
}

type GRPCConfig struct {
	Auth    string `yaml:"auth"`
	Package string `yaml:"package"`
}

func Load() *Config {
	data, err := os.ReadFile("./telegram/configs/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Error reading config file: %v", err))
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("Error parsing config: %v", err))
	}

	return &cfg
}
