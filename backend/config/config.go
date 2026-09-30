package config

import (
	"fmt"
	"os"
	"strconv"
)

// Config holds all configuration for the application.
type Config struct {
	App      AppConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Storage  StorageConfig
}

// AppConfig holds application-level configuration.
type AppConfig struct {
	Name        string
	Env         string
	Port        string
	Debug       bool
	FrontendURL string
}

// DatabaseConfig holds PostgreSQL connection configuration.
type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
	TimeZone string
}

// DSN returns the PostgreSQL connection string.
func (d DatabaseConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s TimeZone=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode, d.TimeZone,
	)
}

// RedisConfig holds Redis connection configuration.
type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// Addr returns the Redis address string.
func (r RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", r.Host, r.Port)
}

// JWTConfig holds JWT token configuration.
type JWTConfig struct {
	AccessSecret  string
	RefreshSecret string
	AccessExpiry  int // in hours
	RefreshExpiry int // in days
}

// StorageConfig holds object storage configuration.
type StorageConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
	UseSSL          bool
}

// Load reads configuration from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		App: AppConfig{
			Name:        getEnv("APP_NAME", "SIMAK Management System"),
			Env:         getEnv("APP_ENV", "development"),
			Port:        getEnv("APP_PORT", "8080"),
			Debug:       getEnvBool("APP_DEBUG", true),
			FrontendURL: getEnv("FRONTEND_URL", "http://localhost:5173"),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "simak_user"),
			Password: getEnv("DB_PASSWORD", "simak_password"),
			DBName:   getEnv("DB_NAME", "simak_db"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
			TimeZone: getEnv("DB_TIMEZONE", "Asia/Jakarta"),
		},
		Redis: RedisConfig{
			Host:     getEnv("REDIS_HOST", "localhost"),
			Port:     getEnv("REDIS_PORT", "6379"),
			Password: getEnv("REDIS_PASSWORD", ""),
			DB:       getEnvInt("REDIS_DB", 0),
		},
		JWT: JWTConfig{
			AccessSecret:  getEnv("JWT_ACCESS_SECRET", "simak-access-secret-change-in-production"),
			RefreshSecret: getEnv("JWT_REFRESH_SECRET", "simak-refresh-secret-change-in-production"),
			AccessExpiry:  getEnvInt("JWT_ACCESS_EXPIRY_HOURS", 24),
			RefreshExpiry: getEnvInt("JWT_REFRESH_EXPIRY_DAYS", 7),
		},
		Storage: StorageConfig{
			Endpoint:        getEnv("STORAGE_ENDPOINT", "localhost:9000"),
			AccessKeyID:     getEnv("STORAGE_ACCESS_KEY", "minioadmin"),
			SecretAccessKey: getEnv("STORAGE_SECRET_KEY", "minioadmin"),
			Bucket:          getEnv("STORAGE_BUCKET", "simak-files"),
			UseSSL:          getEnvBool("STORAGE_USE_SSL", false),
		},
	}
}

func getEnv(key, defaultVal string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if value, exists := os.LookupEnv(key); exists {
		b, err := strconv.ParseBool(value)
		if err == nil {
			return b
		}
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if value, exists := os.LookupEnv(key); exists {
		i, err := strconv.Atoi(value)
		if err == nil {
			return i
		}
	}
	return defaultVal
}
