package configs

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Kafka      KafkaConfig      `yaml:"kafka"`
	Calculator CalculatorConfig `yaml:"calculator"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
}

type KafkaConfig struct {
	Brokers []string `yaml:"brokers"`
	Topic   string   `yaml:"topic"`
	GroupID string   `yaml:"groupID"`
}

type CalculatorConfig struct {
	ClientType string `yaml:"client_type"`
	Address    string `yaml:"address"`
}

func LoadConfig() *Config {
	data, err := os.ReadFile("./consumer/configs/config.yaml")
	if err != nil {
		panic(fmt.Sprintf("Error reading config file: %v", err))
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("Error parsing config: %v", err))
	}

	return &cfg
}
