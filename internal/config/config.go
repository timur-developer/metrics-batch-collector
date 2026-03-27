package config

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPPort      string
	ClickHouseDSN string
	BatchSize     int
	FlushInterval time.Duration
	LogLevel      string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPPort:      os.Getenv("HTTP_PORT"),
		ClickHouseDSN: os.Getenv("CLICKHOUSE_DSN"),
		LogLevel:      getEnv("LOG_LEVEL", "info"),
	}

	if cfg.HTTPPort == "" {
		return Config{}, errors.New("HTTP_PORT is required")
	}

	if cfg.ClickHouseDSN == "" {
		return Config{}, errors.New("CLICKHOUSE_DSN is required")
	}

	batchSizeValue := os.Getenv("BATCH_SIZE")
	if batchSizeValue == "" {
		return Config{}, errors.New("BATCH_SIZE is required")
	}

	batchSize, err := strconv.Atoi(batchSizeValue)
	if err != nil || batchSize <= 0 {
		return Config{}, fmt.Errorf("BATCH_SIZE must be a positive integer")
	}
	cfg.BatchSize = batchSize

	flushIntervalValue := os.Getenv("FLUSH_INTERVAL")
	if flushIntervalValue == "" {
		return Config{}, errors.New("FLUSH_INTERVAL is required")
	}

	flushInterval, err := time.ParseDuration(flushIntervalValue)
	if err != nil || flushInterval <= 0 {
		return Config{}, fmt.Errorf("FLUSH_INTERVAL must be a positive duration")
	}
	cfg.FlushInterval = flushInterval

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}

	return value
}
