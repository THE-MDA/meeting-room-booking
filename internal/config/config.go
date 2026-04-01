package config

import (
    "fmt"
    "os"
    "strconv"
    "time"
)

type Config struct {
    DBHost     string
    DBPort     int
    DBUser     string
    DBPassword string
    DBName     string
    
    JWTSecret     string
    JWTExpiration time.Duration
    
    ServerPort string
    
    Environment string
    LogLevel    string
}

func Load() (*Config, error) {
    cfg := &Config{
        DBHost:     getEnv("DB_HOST", "localhost"),
        DBPort:     getEnvAsInt("DB_PORT", 5432),
        DBUser:     getEnv("DB_USER", "app"),
        DBPassword: getEnv("DB_PASSWORD", "secret"),
        DBName:     getEnv("DB_NAME", "meeting_rooms"),
        
        JWTSecret:     getEnv("JWT_SECRET", "super-secret-jwt-key"),
        JWTExpiration: time.Duration(getEnvAsInt("JWT_EXPIRATION_MINUTES", 30)) * time.Minute,
        
        ServerPort:  getEnv("SERVER_PORT", "8080"),
        Environment: getEnv("ENVIRONMENT", "development"),
        LogLevel:    getEnv("LOG_LEVEL", "info"),
    }
    
    return cfg, nil
}

func (c *Config) GetDBConnectionString() string {
    return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
        c.DBHost, c.DBPort, c.DBUser, c.DBPassword, c.DBName)
}

func (c *Config) GetDatabaseURL() string {
    return fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=disable",
        c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)
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