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
