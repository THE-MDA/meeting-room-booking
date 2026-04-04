package config

import (
    "os"
    "testing"
)

func TestLoad(t *testing.T) {
    os.Setenv("DB_HOST", "testhost")
    os.Setenv("DB_PORT", "5433")
    os.Setenv("DB_USER", "testuser")
    os.Setenv("DB_PASSWORD", "testpass")
    os.Setenv("DB_NAME", "testdb")
    os.Setenv("JWT_SECRET", "testsecret")
    os.Setenv("JWT_EXPIRATION_MINUTES", "60")
    os.Setenv("SERVER_PORT", "9090")
    os.Setenv("ENVIRONMENT", "test")
    os.Setenv("LOG_LEVEL", "debug")

    defer func() {
        os.Unsetenv("DB_HOST")
        os.Unsetenv("DB_PORT")
        os.Unsetenv("DB_USER")
        os.Unsetenv("DB_PASSWORD")
        os.Unsetenv("DB_NAME")
        os.Unsetenv("JWT_SECRET")
        os.Unsetenv("JWT_EXPIRATION_MINUTES")
        os.Unsetenv("SERVER_PORT")
        os.Unsetenv("ENVIRONMENT")
        os.Unsetenv("LOG_LEVEL")
    }()

    cfg, err := Load()
    if err != nil {
        t.Fatalf("Load() error = %v", err)
    }

    if cfg.DBHost != "testhost" {
        t.Errorf("DBHost = %v, want testhost", cfg.DBHost)
    }
    if cfg.DBPort != 5433 {
        t.Errorf("DBPort = %v, want 5433", cfg.DBPort)
    }
    if cfg.JWTSecret != "testsecret" {
        t.Errorf("JWTSecret = %v, want testsecret", cfg.JWTSecret)
    }
    if cfg.ServerPort != "9090" {
        t.Errorf("ServerPort = %v, want 9090", cfg.ServerPort)
    }
}

func TestGetDBConnectionString(t *testing.T) {
    cfg := &Config{
        DBHost:     "localhost",
        DBPort:     5432,
        DBUser:     "user",
        DBPassword: "pass",
        DBName:     "mydb",
    }

    expected := "host=localhost port=5432 user=user password=pass dbname=mydb sslmode=disable"
    if cfg.GetDBConnectionString() != expected {
        t.Errorf("GetDBConnectionString() = %v, want %v", cfg.GetDBConnectionString(), expected)
    }
}

func TestGetDatabaseURL(t *testing.T) {
    cfg := &Config{
        DBHost:     "localhost",
        DBPort:     5432,
        DBUser:     "user",
        DBPassword: "pass",
        DBName:     "mydb",
    }

    expected := "postgres://user:pass@localhost:5432/mydb?sslmode=disable"
    if cfg.GetDatabaseURL() != expected {
        t.Errorf("GetDatabaseURL() = %v, want %v", cfg.GetDatabaseURL(), expected)
    }
}