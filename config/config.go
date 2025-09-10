package config

import (
	"os"
	"strconv"
	"time"
)

type Application struct {
	GracefulShutdownTimeout time.Duration
}

type HTTPServer struct {
	Port int
}

type Database struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string
}

type Redis struct {
	Host     string
	Port     int
	Password string
	DB       int
}

type Logger struct {
	Level string
	Mode  string // development or production
}

type Swagger struct {
	Enabled bool `json:"enabled"`
}

type JWT struct {
	Secret         string
	ExpirationTime time.Duration
}

type OTP struct {
	Length         int
	ExpirationTime time.Duration
}

type RateLimit struct {
	MaxRequests    int
	WindowDuration time.Duration
}

type Config struct {
	Application Application
	HTTPServer  HTTPServer
	Database    Database
	Redis       Redis
	Logger      Logger
	Swagger     Swagger
	JWT         JWT
	OTP         OTP
	RateLimit   RateLimit
}

func Load() (*Config, error) {
	cfg := &Config{
		Application: Application{
			GracefulShutdownTimeout: parseDurationWithDefault("APPLICATION_GRACEFUL_SHUTDOWN_TIMEOUT", 30*time.Second),
		},
		HTTPServer: HTTPServer{
			Port: parseIntWithDefault("HTTP_SERVER_PORT", 8080),
		},
		Database: Database{
			Host:     getEnvWithDefault("DATABASE_HOST", "db"),
			Port:     parseIntWithDefault("DATABASE_PORT", 5432),
			User:     getEnvWithDefault("DATABASE_USER", "otp_auth"),
			Password: getEnvWithDefault("DATABASE_PASSWORD", "otp_auth"),
			Name:     getEnvWithDefault("DATABASE_NAME", "otp_auth"),
			SSLMode:  getEnvWithDefault("DATABASE_SSL_MODE", "disable"),
		},
		Logger: Logger{
			Level: getEnvWithDefault("LOGGER_LEVEL", "info"),
			Mode:  getEnvWithDefault("LOGGER_MODE", "production"),
		},
		Swagger: Swagger{
			Enabled: getEnvBoolWithDefault("SWAGGER_ENABLED", true),
		},
		JWT: JWT{
			Secret:         getEnvWithDefault("JWT_SECRET", "your-super-secret-key-change-in-production"),
			ExpirationTime: parseDurationWithDefault("JWT_EXPIRATION_TIME", 24*time.Hour),
		},
		OTP: OTP{
			Length:         parseIntWithDefault("OTP_LENGTH", 6),
			ExpirationTime: parseDurationWithDefault("OTP_EXPIRATION_TIME", 2*time.Minute),
		},
		Redis: Redis{
			Host:     getEnvWithDefault("REDIS_HOST", "redis"),
			Port:     parseIntWithDefault("REDIS_PORT", 6379),
			Password: getEnvWithDefault("REDIS_PASSWORD", ""),
			DB:       parseIntWithDefault("REDIS_DB", 0),
		},
		RateLimit: RateLimit{
			MaxRequests:    parseIntWithDefault("RATE_LIMIT_MAX_REQUESTS", 3),
			WindowDuration: parseDurationWithDefault("RATE_LIMIT_WINDOW_DURATION", 10*time.Minute),
		},
	}

	// Support legacy environment variables for backwards compatibility
	if port := os.Getenv("APP_PORT"); port != "" {
		if p, err := strconv.Atoi(port); err == nil {
			cfg.HTTPServer.Port = p
		}
	}

	return cfg, nil
}

func getEnvWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func parseIntWithDefault(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.Atoi(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func parseDurationWithDefault(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		if parsed, err := time.ParseDuration(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getEnvBoolWithDefault(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		if parsed, err := strconv.ParseBool(value); err == nil {
			return parsed
		}
	}
	return defaultValue
}

func getStringWithDefault(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
