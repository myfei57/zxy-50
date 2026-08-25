package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Addr              string
	DataDir           string
	PeakRainThreshold float64
	PeakWindow        time.Duration
	FullSpeedRate     float64
	ReverseFlowDelta  float64
	LevelRateWindow   time.Duration
}

func Load() Config {
	return Config{
		Addr:              envString("DRAINNET_ADDR", "127.0.0.1:8080"),
		DataDir:           envString("DRAINNET_DATA", "data"),
		PeakRainThreshold: envFloat("DRAINNET_PEAK_RAIN_MM", 50),
		PeakWindow:        envDuration("DRAINNET_PEAK_WINDOW", 10*time.Minute),
		FullSpeedRate:     envFloat("DRAINNET_FULL_SPEED_RATE", 5),
		ReverseFlowDelta:  envFloat("DRAINNET_REVERSE_DELTA", 8),
		LevelRateWindow:   envDuration("DRAINNET_LEVEL_WINDOW", 5*time.Minute),
	}
}

func envString(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envFloat(key string, fallback float64) float64 {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return fallback
	}
	return parsed
}

func envDuration(key string, fallback time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := time.ParseDuration(value)
	if err != nil {
		return fallback
	}
	return parsed
}
