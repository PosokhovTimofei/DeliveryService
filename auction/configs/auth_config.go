package configs

import (
	"os"

	"gopkg.in/yaml.v3"
)

type AuthConfig struct {
	ExcludedMethods []string `yaml:"excluded_methods"`
}

var authConfig AuthConfig

func LoadAuthConfig() error {
	data, err := os.ReadFile("./auction/configs/auth.yaml")
	if err != nil {
		return err
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
