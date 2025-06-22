package config

import (
	"os"
	"strconv"
)

type Config struct {
	Server ServerConfig
	Redis  RedisConfig
	STUN   STUNConfig
}

type ServerConfig struct {
	Port         string
	Host         string
	ReadTimeout  int
	WriteTimeout int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type STUNConfig struct {
	URLs []string
}

func Load() *Config {
	return &Config{
		Server: ServerConfig{
			Port:         getEnv("SERVER_PORT", "8080"),
			Host:         getEnv("SERVER_HOST", "0.0.0.0"),
			ReadTimeout:  getEnvAsInt("READ_TIMEOUT", 60),
			WriteTimeout: getEnvAsInt("WRITE_TIMEOUT", 60),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvAsInt("REDIS_DB", 0),
		},
		STUN: STUNConfig{
			URLs: []string{
				getEnv("STUN_URL", "stun:localhost:3478"),
				getEnv("TURN_URL", "turn:localhost:3478"),
			},
		},
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvAsInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if intValue, err := strconv.Atoi(value); err == nil {
			return intValue
		}
	}
	return defaultValue
}
