package configs

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerPort    string `yaml:"SERVER_PORT"`
	ProtectedPort string `yaml:"PROTECTED_PORT"`
	DBUri         string `yaml:"DB_URI"`
	DBName        string `yaml:"DB_NAME"`
	JWTSecret     string `yaml:"JWT_SECRET"`
	MetricsPort   string `yaml:"METRICS_PORT"`
	GRPCPort      string `yaml:"GRPC_PORT"`
}

func Load() *Config {
	configPath := os.Getenv("AUTH_CONFIG")
	if configPath == "" {
		configPath = "./auth/configs/config.yaml"
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
