package config

import (
	"log"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL   string
	DBHost        string
	DBPort        string
	DBName        string
	DBUser        string
	DBPassword    string
	AdminEmail    string
	AdminPassword string
	JWTSecret     string
	JWTTTL        int64
	CRMService    string
	CRMToken      string
	ClientURL     string
	Port          string
}

func Load() *Config {
	if err := godotenv.Load(); err != nil {
		log.Println(".env file was not loaded, environment variables will be used")
	}

	jwtTTL := int64(24)
	if raw := os.Getenv("JWT_TTL_HOURS"); raw != "" {
		if parsed, err := strconv.ParseInt(raw, 10, 64); err == nil && parsed > 0 {
			jwtTTL = parsed
		}
	}

	return &Config{
		DatabaseURL:   env("DATABASE_URL", ""),
		DBHost:        env("DB_HOST", "localhost"),
		DBPort:        env("DB_PORT", "5432"),
		DBName:        env("DB_NAME", "testconstructor"),
		DBUser:        env("DB_USER", "postgres"),
		DBPassword:    env("DB_PASSWORD", "postgres"),
		AdminEmail:    env("ADMIN_EMAIL", "admin@example.com"),
		AdminPassword: env("ADMIN_PASSWORD", "admin"),
		JWTSecret:     env("JWT_SECRET", "17a3229b-e5c6-4ab0-ba86-3d87cb7f23fe"),
		JWTTTL:        jwtTTL,
		CRMService:    env("CRM_SERVICE", "http://127.0.0.1:8000"),
		CRMToken:      env("CRM_TOKEN", ""),
		ClientURL:     env("CLIENT_URL", "http://localhost:5173"),
		Port:          env("PORT", "8080"),
	}
}

func env(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
