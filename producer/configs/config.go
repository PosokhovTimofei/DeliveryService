package configs

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server     ServerConfig     `yaml:"server"`
	Database   DatabaseConfig   `yaml:"database"`
	Kafka      KafkaConfig      `yaml:"kafka"`
	Calculator CalculatorConfig `yaml:"calculator"`
}

type ServerConfig struct {
	Address string `yaml:"address"`
}

type DatabaseConfig struct {
	Uri  string `yaml:"uri"`
	Name string `yaml:"name"`
}

type KafkaConfig struct {
	Brokers      []string `yaml:"brokers"`
	PackageTopic string   `yaml:"packageTopic"`
	PaymentTopic string   `yaml:"paymentTopic"`
	Version      string   `yaml:"version"`
}

type CalculatorConfig struct {
	URL string `yaml:"url"`
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
