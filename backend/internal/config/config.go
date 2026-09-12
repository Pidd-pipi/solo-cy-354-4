// Package config loads runtime configuration from environment variables.
package config

import (
	"os"
	"strconv"
	"strings"
)

// Config holds all runtime settings for the campus-market backend.
type Config struct {
	Port            string
	DSN             string
	JWTSecret       string
	JWTExpireHours  int
	RateLimitPerMin int
	LoginRateLimit  int
	CORSOrigins     []string
	SeedingEnabled  bool
}

// Load reads configuration from environment variables and applies defaults.
func Load() *Config {
	return &Config{
		Port:            getEnv("PORT", "8080"),
		DSN:             getEnv("DB_DSN", "lpmarket_user:lpmarket_pwd@tcp(db:3306)/lpcampusmarket_db?charset=utf8mb4&parseTime=True&loc=Local"),
		JWTSecret:       getEnv("JWT_SECRET", "change_me_to_a_long_random_string"),
		JWTExpireHours:  getEnvInt("JWT_EXPIRE_HOURS", 72),
		RateLimitPerMin: getEnvInt("RATE_LIMIT_PER_MIN", 120),
		LoginRateLimit:  getEnvInt("LOGIN_RATE_LIMIT_PER_MIN", 10),
		CORSOrigins:     splitCSV(getEnv("CORS_ORIGINS", "http://localhost:28514,http://localhost:5173")),
		SeedingEnabled:  getEnvBool("SEEDING_ENABLED", true),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if n, err := strconv.Atoi(v); err == nil {
			return n
		}
	}
	return fallback
}

func getEnvBool(key string, fallback bool) bool {
	if v := os.Getenv(key); v != "" {
		if b, err := strconv.ParseBool(v); err == nil {
			return b
		}
	}
	return fallback
}

func splitCSV(s string) []string {
	var out []string
	for _, part := range strings.Split(s, ",") {
		if t := strings.TrimSpace(part); t != "" {
			out = append(out, t)
		}
	}
	return out
}
