package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

type Config struct {
	Port            string
	MetricsInterval time.Duration

	// Kubernetes scaler settings
	K8sNamespace       string
	K8sDeployment      string
	ScaleUpThreshold   float64 // CPU% above which we scale up
	ScaleDownThreshold float64 // CPU% below which we scale down
	MaxReplicas        int32
	MinReplicas        int32

	// CORS
	AllowedOrigins string
}

func Load() *Config {
	_ = godotenv.Load() // silently ignore missing .env in production

	return &Config{
		Port:               getEnv("PORT", "8080"),
		MetricsInterval:    getDuration("METRICS_INTERVAL", time.Second),
		K8sNamespace:       getEnv("K8S_NAMESPACE", "default"),
		K8sDeployment:      getEnv("K8S_DEPLOYMENT", "go-monitor-app"),
		ScaleUpThreshold:   getFloat("SCALE_UP_THRESHOLD", 75.0),
		ScaleDownThreshold: getFloat("SCALE_DOWN_THRESHOLD", 25.0),
		MaxReplicas:        int32(getInt("MAX_REPLICAS", 10)),
		MinReplicas:        int32(getInt("MIN_REPLICAS", 1)),
		AllowedOrigins:     getEnv("ALLOWED_ORIGINS", "*"),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func getDuration(key string, fallback time.Duration) time.Duration {
	if v := os.Getenv(key); v != "" {
		if ms, err := strconv.Atoi(v); err == nil {
			return time.Duration(ms) * time.Millisecond
		}
	}
	return fallback
}

func getFloat(key string, fallback float64) float64 {
	if v := os.Getenv(key); v != "" {
		if f, err := strconv.ParseFloat(v, 64); err == nil {
			return f
		}
	}
	return fallback
}

func getInt(key string, fallback int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return fallback
}
