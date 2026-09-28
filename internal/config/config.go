package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type EnvType string

const (
	DEVELOPMENT EnvType = "development"
	PRODUCTION  EnvType = "production"
	TESTING     EnvType = "testing"
)

type Config struct {
	Port        string
	Env         EnvType
	DatabaseUrl string
}

func ParseEnv(value string) (EnvType, error) {
	env := EnvType(value)

	switch env {
	case DEVELOPMENT, PRODUCTION, TESTING:
		return env, nil

	default:
		return "", fmt.Errorf("invalid environment: %q", value)
	}
}

func Load() (Config, error) {
	_ = godotenv.Load()

	port := os.Getenv("PORT")
	if port == "" {
		return Config{}, fmt.Errorf("PORT is required")
	}

	parsedEnv, err := ParseEnv(os.Getenv("ENV"))
	if err != nil {
		return Config{}, fmt.Errorf("invalid ENV configuration: %w", err)
	}

	databaseUrl := os.Getenv("DATABASE_URL")
	if databaseUrl == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return Config{
		Port:        port,
		Env:         parsedEnv,
		DatabaseUrl: databaseUrl,
	}, nil
}

// MustLoad loads configuration or terminates the application
func MustLoad() Config {
	cfg, err := Load()
	if err != nil {
		panic(err)
	}

	return cfg
}
