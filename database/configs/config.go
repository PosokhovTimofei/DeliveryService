package configs

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Kafka      KafkaConfig      `yaml:"kafka"`
	Calculator CalculatorConfig `yaml:"calculator"`
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

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   []string `yaml:"topics"`
	GroupID string   `yaml:"groupID"`
}

type CalculatorConfig struct {
	GRPCAddress string `yaml:"grpc_address"`
}

func Load() *Config {
	data, err := os.ReadFile("./database/configs/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Error reading config file: %v", err))
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("Error parsing config: %v", err))
	}

	return &cfg
}
