package config

import (
    "os"
    "strconv"
)

type Config struct {
    DatabaseURL string
    Port        string
    SessionKey  string
    
    // AWS
    AWSRegion          string
    AWSAccessKeyID     string
    AWSSecretAccessKey string
    S3Bucket          string
    
    // External API
    ExternalAPIURL string
    ExternalAPIKey string
    
    // App settings
    Environment string
    LogLevel    string
}

func Load() *Config {
    return &Config{
        DatabaseURL: getEnv("DATABASE_URL", "postgres://localhost/adtech_dev"),
        Port:        getEnv("PORT", "8080"),
        SessionKey:  getEnv("SESSION_KEY", "default-dev-key-change-in-production"),
        
        AWSRegion:          getEnv("AWS_REGION", "us-east-1"),
        AWSAccessKeyID:     getEnv("AWS_ACCESS_KEY_ID", ""),
        AWSSecretAccessKey: getEnv("AWS_SECRET_ACCESS_KEY", ""),
        S3Bucket:          getEnv("S3_BUCKET", "adtech-audiences"),
        
        ExternalAPIURL: getEnv("EXTERNAL_API_URL", ""),
        ExternalAPIKey: getEnv("EXTERNAL_API_KEY", ""),
        
        Environment: getEnv("ENVIRONMENT", "development"),
        LogLevel:    getEnv("LOG_LEVEL", "info"),
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intVal, err := strconv.Atoi(value); err == nil {
            return intVal
        }
    }
    return defaultValue
}