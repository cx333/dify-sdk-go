// Package config 提供 server 层配置，从环境变量加载。
package config

import (
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

// ServerConfig 服务器完整配置。所有值通过环境变量注入。
type ServerConfig struct {
	Port                 string
	Env                  string
	MethodTimeout        time.Duration
	LogLevel             string
	LogFormat            string
	OutboundRPS          float64
	OutboundBurst        int
	MaxConcurrentMethods int
	LogFile              LogFileConfig
}

// LogFileConfig 日志文件配置。
type LogFileConfig struct {
	Enabled    bool   // 是否开启日志落盘
	Dir        string // 日志保存目录
	MaxSize    int    // 单个日志文件最大大小（MB）
	MaxAge     int    // 日志最大保留天数
	MaxBackups int    // 最多保留的旧日志文件数
}

// Load 从环境变量加载 ServerConfig，缺失项使用默认值。
func Load() *ServerConfig {
	// 优先从 .env 文件加载环境变量（静默忽略文件不存在的场景，
	// 此时回退到系统环境变量）。
	_ = godotenv.Load(".env")

	return &ServerConfig{
		Port:                 envOrDefault("SERVER_PORT", "8080"),
		Env:                  envOrDefault("ENV", "development"),
		MethodTimeout:        parseDuration(envOrDefault("METHOD_TIMEOUT", "30s")),
		LogLevel:             envOrDefault("LOG_LEVEL", "info"),
		LogFormat:            envOrDefault("LOG_FORMAT", "text"),
		OutboundRPS:          parseFloat(envOrDefault("OUTBOUND_RPS", "0")),
		OutboundBurst:        parseInt(envOrDefault("OUTBOUND_BURST", "10")),
		MaxConcurrentMethods: parseInt(envOrDefault("MAX_CONCURRENT_METHODS", "0")),
		LogFile: LogFileConfig{
			Enabled:    parseBool(envOrDefault("LOG_FILE_ENABLED", "false")),
			Dir:        envOrDefault("LOG_DIR", "./logs"),
			MaxSize:    parseInt(envOrDefault("LOG_MAX_SIZE", "100")),
			MaxAge:     parseInt(envOrDefault("LOG_MAX_AGE", "30")),
			MaxBackups: parseInt(envOrDefault("LOG_MAX_BACKUPS", "10")),
		},
	}
}

func envOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func parseDuration(s string) time.Duration {
	d, err := time.ParseDuration(s)
	if err != nil {
		return 30 * time.Second
	}
	return d
}

func parseInt(s string) int {
	n, err := strconv.Atoi(s)
	if err != nil {
		return 0
	}
	return n
}

func parseFloat(s string) float64 {
	f, err := strconv.ParseFloat(s, 64)
	if err != nil {
		return 0
	}
	return f
}

func parseBool(s string) bool {
	b, err := strconv.ParseBool(s)
	if err != nil {
		return false
	}
	return b
}
