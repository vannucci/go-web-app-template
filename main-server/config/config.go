package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	Environment string
	Version     string
	SessionKey  string
	Debug       bool
	LogLevel    string
	Database    *DatabaseConfig
	BaseURL     string
	UploadDir   string
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
	SSLMode  string
}

func Load() (*Config, error) {
	env := os.Getenv("GO_ENV")
	if env == "" {
		env = "development"
	}

	// Load .env file
	if err := godotenv.Load(".env." + env); err != nil {
		log.Println("No .env file found, using defaults")
	}

	dbConfig := &DatabaseConfig{
		Host:     getEnv("DB_HOST", "localhost"),
		Port:     getEnv("DB_PORT", "5432"),
		User:     getEnv("DB_USER", "appuser"),
		Password: getEnv("DB_PASSWORD", "apppassword"),
		Name:     getEnv("DB_NAME", "appdb"),
		SSLMode:  getEnv("DB_SSL_MODE", "disable"),
	}

	debug := getEnv("DEBUG", "true") == "true"

	return &Config{
		Port:        getEnv("PORT", "8080"),
		Environment: getEnv("ENVIRONMENT", env),
		Version:     getEnv("VERSION", "1.0.0"),
		SessionKey:  getEnv("SESSION_KEY", "default-dev-key"),
		Debug:       debug,
		LogLevel:    getEnv("LOG_LEVEL", "debug"),
		Database:    dbConfig,
		BaseURL:     getEnv("BASE_URL", "http://localhost:8080"),
		UploadDir:   getEnv("UPLOAD_DIR", "./uploads"),
	}, nil
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
