package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server struct {
		Port int
	}

	MongoDB struct {
		URI      string
		Database string
	}

	GitHub struct {
		Token string
	}
}

func Load() (*Config, error) {
	cfg := &Config{}

	// Server
	port, err := strconv.Atoi(getEnv("SERVER_PORT", "50051"))
	if err != nil {
		return nil, err
	}
	cfg.Server.Port = port

	// MongoDB
	cfg.MongoDB.URI = getEnv("MONGODB_URI", "mongodb://mongo:27017")
	cfg.MongoDB.Database = getEnv("MONGODB_DATABASE", "github_parser")

	// GitHub
	cfg.GitHub.Token = getEnv("GITHUB_TOKEN", "")

	return cfg, nil
}

func getEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
