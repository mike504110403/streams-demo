package config

import "os"

// Config 應用設定，從環境變數讀取
type Config struct {
	DBURL      string
	RedisURL   string
	JWTSecret  string
	ServerPort string
	SRSURL     string // SRS API 地址
	SRSRtmpURL string // RTMP 推流地址
	SRSHttpURL string // HTTP-FLV/HLS 拉流地址
}

// Load 從環境變數載入設定，未設定則使用預設值
func Load() *Config {
	return &Config{
		DBURL:      getEnv("DB_URL", "postgres://streams:streams_dev@localhost:5432/streams_demo?sslmode=disable"),
		RedisURL:   getEnv("REDIS_URL", "localhost:6379"),
		JWTSecret:  getEnv("JWT_SECRET", "dev-secret-key-change-in-production"),
		ServerPort: getEnv("SERVER_PORT", "8081"),
		SRSURL:     getEnv("SRS_URL", "http://localhost:1985"),
		SRSRtmpURL: getEnv("SRS_RTMP_URL", "rtmp://localhost:1935/live"),
		SRSHttpURL: getEnv("SRS_HTTP_URL", "http://localhost:8080/live"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}
