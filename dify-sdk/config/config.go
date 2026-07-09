// Package config 提供基于 .env 文件的配置加载能力。
// 所有配置项通过环境变量注入，禁止使用 config.yaml 等配置文件。
// 典型用法：
//
//	cfg, err := config.Load(".env")
//	if err != nil {
//	    log.Fatal(err)
//	}
package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// Config Dify SDK 的完整配置，从 .env 文件加载。
// 所有字段均通过环境变量设置，支持默认值。
type Config struct {
	// BaseURL Dify API 的基础地址，例如 http://localhost:5001/v1
	BaseURL string
	// APIKeys 支持多个 API Key，逗号分隔
	APIKeys []string
	// Timeout HTTP 请求超时时间，默认 30 秒
	Timeout time.Duration
	// MaxRetries 请求失败后的最大重试次数，默认 3 次
	MaxRetries int
}

// Load 从指定路径读取 .env 文件，解析后返回 Config。
// 缺失必填字段（DIFY_BASE_URL、DIFY_API_KEYS）时返回错误。
func Load(path string) (*Config, error) {
	if err := godotenv.Load(path); err != nil {
		return nil, fmt.Errorf("加载 .env 文件失败: %w", err)
	}

	cfg := &Config{
		BaseURL:    os.Getenv("DIFY_BASE_URL"),
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("DIFY_BASE_URL 为必填项")
	}

	// 优先读取编号环境变量 DIFY_API_KEY_0, DIFY_API_KEY_1, ...
	// 未设置时回退到逗号分隔的 DIFY_API_KEYS。
	cfg.APIKeys = readNumberedKeys()
	if len(cfg.APIKeys) == 0 {
		keys := os.Getenv("DIFY_API_KEYS")
		if keys == "" {
			return nil, fmt.Errorf("DIFY_API_KEYS 或 DIFY_API_KEY_N 为必填项")
		}
		cfg.APIKeys = splitKeys(keys)
	}

	if v := os.Getenv("DIFY_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("DIFY_TIMEOUT 格式无效: %w", err)
		}
		cfg.Timeout = d
	}

	if v := os.Getenv("DIFY_MAX_RETRIES"); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
			return nil, fmt.Errorf("DIFY_MAX_RETRIES 格式无效: %w", err)
		}
		cfg.MaxRetries = n
	}

	return cfg, nil
}

// readNumberedKeys 读取 DIFY_API_KEY_0, DIFY_API_KEY_1, ... 编号变量，
// 直到遇到第一个不存在的编号为止。返回收集到的所有 key（可能为空）。
func readNumberedKeys() []string {
	var keys []string
	for i := 0; ; i++ {
		v := os.Getenv("DIFY_API_KEY_" + strconv.Itoa(i))
		if v == "" {
			break
		}
		keys = append(keys, strings.TrimSpace(v))
	}
	return keys
}

// splitKeys 将逗号分隔的 API Key 字符串拆分为切片，去除空白。
func splitKeys(s string) []string {
	parts := strings.Split(s, ",")
	result := make([]string, 0, len(parts))
	for _, p := range parts {
		p = strings.TrimSpace(p)
		if p != "" {
			result = append(result, p)
		}
	}
	return result
}
