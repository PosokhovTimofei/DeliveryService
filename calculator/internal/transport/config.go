package transport

type Config struct {
	HTTPPort string `yaml:"http_port"`
	GRPCPort string `yaml:"grpc_port"`
}

func Load() *Config {
	return &Config{
		HTTPPort: "8121",
		GRPCPort: "50051",
	}
}
