package configs

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type AuthConfig struct {
	ExcludedMethods []string `yaml:"excluded_methods"`
}

var authConfig AuthConfig

func LoadAuthConfig() error {
	configDir := filepath.Dir(os.Getenv("AUCTION_CONFIG"))
	if configDir == "" {
		configDir = "./auction/configs"
	}

	authConfigPath := filepath.Join(configDir, "auth.yaml")

	data, err := os.ReadFile(authConfigPath)
	if err != nil {
		return fmt.Errorf("error reading auth config file: %v", err)
	}
	return yaml.Unmarshal(data, &authConfig)
}

func IsExcludedMethod(method string) bool {
	for _, m := range authConfig.ExcludedMethods {
		if m == method {
			return true
		}
	}
	return false
}
