package configs

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Postgres PostgresConfig `yaml:"postgres"`
	Kafka    KafkaConfig    `yaml:"kafka"`
}

type ServerConfig struct {
	Address     string `yaml:"address"`
	GRPCAddress string `yaml:"grpc_address"`
}

type DatabaseConfig struct {
	Driver string `yaml:"driver"`
	Uri    string `yaml:"uri"`
	Name   string `yaml:"name"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DBName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type KafkaConfig struct {
	Brokers       []string `yaml:"brokers"`
	ConsumerTopic string   `yaml:"consumerTopic"`
	ProducerTopic []string `yaml:"producerTopic"`
	GroupID       string   `yaml:"groupID"`
	Version       string   `yaml:"version"`
}

func Load() *Config {
	configPath := os.Getenv("PAYMENT_CONFIG")
	if configPath == "" {
		configPath = "./payment/configs/config.yaml" // путь по умолчанию
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		panic(fmt.Sprintf("Error reading config file: %v", err))
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		panic(fmt.Sprintf("Error parsing config: %v", err))
	}

	return &cfg
}
