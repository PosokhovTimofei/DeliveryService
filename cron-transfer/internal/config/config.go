package config

import "time"

type Config struct {
	GRPCAddress string
	Interval    time.Duration
}

func Load() Config {
	return Config{
		GRPCAddress: "localhost:50054",
		Interval:    24 * time.Hour,
	}
}
