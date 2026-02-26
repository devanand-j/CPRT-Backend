package config

import (
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Port          string
	DatabaseURL   string
	JWTSecret     string
	JWTIssuer     string
	JWTTTLMinutes int
}

func Load() Config {
	loadDotEnv()

	port := getEnv("PORT", "8080")
	dbURL := getEnv("DATABASE_URL", "")
	jwtSecret := getEnv("JWT_SECRET", "change-me")
	jwtIssuer := getEnv("JWT_ISSUER", "cprt-lis")
	ttl := getEnv("JWT_TTL_MINUTES", "60")

	ttlMinutes, err := strconv.Atoi(ttl)
	if err != nil {
		log.Printf("invalid JWT_TTL_MINUTES, using 60")
		ttlMinutes = 60
	}

	if dbURL == "" {
		log.Printf("DATABASE_URL is empty")
	}

	return Config{
		Port:          port,
		DatabaseURL:   dbURL,
		JWTSecret:     jwtSecret,
		JWTIssuer:     jwtIssuer,
		JWTTTLMinutes: ttlMinutes,
	}
}

func loadDotEnv() {
	if err := godotenv.Load(); err == nil {
		return
	}

	cwd, err := os.Getwd()
	if err != nil {
		return
	}

	current := cwd
	for {
		envPath := filepath.Join(current, ".env")
		if _, statErr := os.Stat(envPath); statErr == nil {
			_ = godotenv.Overload(envPath)
			return
		}

		parent := filepath.Dir(current)
		if parent == current {
			return
		}
		current = parent
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}
