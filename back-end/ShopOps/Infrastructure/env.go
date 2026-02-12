package Infrastructure

import (
	"github.com/joho/godotenv"
	"os"
)

// LoadEnv loads .env if present
func LoadEnv() error {
	return godotenv.Load()
}

// helper gets env with default
func GetEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
