package configs

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server ServerConfig `yaml:"server"`
	Kafka  KafkaConfig  `yaml:"kafka"`
}

type ServerConfig struct {
	Address string      `yaml:"address"`
	Kafka   KafkaConfig `yaml:"kafka"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
}

func Load() *Config {
	data, err := os.ReadFile("./producer/configs/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Error reading config file: %v", err))
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("Error parsing config: %v", err))
	}

	return &cfg
}
