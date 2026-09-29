package utils

import (
	"os"
	"strconv"
)

func GetEnv(key string, fallback string) string {
	var value string
	if value = os.Getenv(key); value == "" {
		return fallback
	}
	return value
}

func GetEnvInt(key string, fallback int) int {
	var valueStr string
	if valueStr = os.Getenv(key); valueStr == "" {
		return fallback
	}
	value, err := strconv.Atoi(valueStr)
	if err != nil {
		return fallback
	}
	return value
}
