package config

import "os"

// Config 應用設定，從環境變數讀取
type Config struct {
	DBURL      string
	RedisURL   string
	JWTSecret  string
	ServerPort string
}

// Load 從環境變數載入設定，未設定則使用預設值
func Load() *Config {
	return &Config{
		DBURL:      getEnv("DB_URL", "postgres://streams:streams_dev@localhost:5432/streams_demo?sslmode=disable"),
		RedisURL:   getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:  getEnv("JWT_SECRET", "dev-secret-key-change-in-production"),
		ServerPort: getEnv("SERVER_PORT", "8081"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
