package app

import (
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	AppEnv          string
	HTTPPort        string
	DatabaseURL     string
	RedisAddr       string
	RedisPassword   string
	RedisDB         int
	KafkaBrokers    []string
	JWTAccessSecret string
	JWTRefreshSecret string
	JWTAccessTTL    time.Duration
	JWTRefreshTTL   time.Duration
	HoldTTL         time.Duration
	SeatmapCacheTTL time.Duration
	LogLevel        string
}

func LoadConfig() (Config, error) {
	redisDB, err := strconv.Atoi(env("REDIS_DB", "0"))
	if err != nil {
		return Config{}, err
	}
	accessTTL, err := time.ParseDuration(env("JWT_ACCESS_TTL", "15m"))
	if err != nil {
		return Config{}, err
	}
	refreshTTL, err := time.ParseDuration(env("JWT_REFRESH_TTL", "720h"))
	if err != nil {
		return Config{}, err
	}
	holdTTL, err := time.ParseDuration(env("HOLD_TTL", "10m"))
	if err != nil {
		return Config{}, err
	}
	seatmapTTL, err := time.ParseDuration(env("SEATMAP_CACHE_TTL", "30s"))
	if err != nil {
		return Config{}, err
	}

	return Config{
		AppEnv:           env("APP_ENV", "local"),
		HTTPPort:         env("HTTP_PORT", "8080"),
		DatabaseURL:      env("DATABASE_URL", "postgres://reserveflow:reserveflow@localhost:5432/reserveflow?sslmode=disable"),
		RedisAddr:        env("REDIS_ADDR", "localhost:6379"),
		RedisPassword:    os.Getenv("REDIS_PASSWORD"),
		RedisDB:          redisDB,
		KafkaBrokers:     splitCSV(env("KAFKA_BROKERS", "localhost:9092")),
		JWTAccessSecret:  env("JWT_ACCESS_SECRET", "change-me-access"),
		JWTRefreshSecret: env("JWT_REFRESH_SECRET", "change-me-refresh"),
		JWTAccessTTL:     accessTTL,
		JWTRefreshTTL:    refreshTTL,
		HoldTTL:          holdTTL,
		SeatmapCacheTTL:  seatmapTTL,
		LogLevel:         env("LOG_LEVEL", "debug"),
	}, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func splitCSV(value string) []string {
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" {
			out = append(out, part)
		}
	}
	return out
}

