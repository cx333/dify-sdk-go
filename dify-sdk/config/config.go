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
	// DefaultUser 默认用户标识。当业务层未传入 user 参数且未调用 SetUser 时，
	// SDK 将使用此值作为兜底。通过环境变量 DIFY_DEFAULT_USER 设置。
	//
	// 安全警告（务必阅读）：
	//
	// Dify 平台通过 "user + conversation_id" 组合来隔离会话记忆。
	// 如果所有请求使用相同的 user，Dify 会将它们视为同一用户，导致：
	//
	//   1. 会话记忆串扰 — 用户 A 的对话可能被注入到用户 B 的上下文中
	//   2. 会话列表混乱 — 所有终端用户的对话混在一起
	//   3. 消息反馈污染 — like/dislike 统计失真
	//
	// 推荐做法：
	//
	//   1. 始终由业务层在请求中传入真实用户标识（如登录用户的 ID）
	//   2. DIFY_DEFAULT_USER 仅作为开发/测试环境的兜底，或单用户内部工具场景
	//   3. 生产环境务必使用每请求独立传入 user 参数
	//
	// 使用场景：
	//   - 单用户内部工具（如公司内部知识库问答机器人）— 可以依赖此默认值
	//   - 多用户 SaaS 产品 — 严禁依赖此默认值，必须逐请求传入真实 user
	//   - 开发/测试环境 — 方便快速调试，无需每次指定 user
	//
	// 示例：
	//   DIFY_DEFAULT_USER=internal-tool     # 内部工具，只有一种用户
	//   DIFY_DEFAULT_USER=dev-test          # 开发环境，数据隔离不重要
	//   # 生产多租户环境：不设此值，由业务代码逐请求传入 user
	DefaultUser string
}

// Load 从指定路径读取 .env 文件，解析后返回 Config。
// 缺失必填字段（DIFY_BASE_URL、DIFY_API_KEYS）时返回错误。
func Load(path string) (*Config, error) {
	if err := godotenv.Load(path); err != nil {
		return nil, fmt.Errorf("load .env file failed: %w", err)
	}

	cfg := &Config{
		BaseURL:    os.Getenv("DIFY_BASE_URL"),
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}

	if cfg.BaseURL == "" {
		return nil, fmt.Errorf("DIFY_BASE_URL is required")
	}

	// 优先读取编号环境变量 DIFY_API_KEY_0, DIFY_API_KEY_1, ...
	// 未设置时回退到逗号分隔的 DIFY_API_KEYS。
	cfg.APIKeys = readNumberedKeys()
	if len(cfg.APIKeys) == 0 {
		keys := os.Getenv("DIFY_API_KEYS")
		if keys == "" {
			return nil, fmt.Errorf("DIFY_API_KEYS or DIFY_API_KEY_N is required")
		}
		cfg.APIKeys = splitKeys(keys)
	}

	if v := os.Getenv("DIFY_TIMEOUT"); v != "" {
		d, err := time.ParseDuration(v)
		if err != nil {
			return nil, fmt.Errorf("DIFY_TIMEOUT: invalid format: %w", err)
		}
		cfg.Timeout = d
	}

	if v := os.Getenv("DIFY_MAX_RETRIES"); v != "" {
		var n int
		if _, err := fmt.Sscanf(v, "%d", &n); err != nil {
			return nil, fmt.Errorf("DIFY_MAX_RETRIES: invalid format: %w", err)
		}
		cfg.MaxRetries = n
	}

	// 默认用户标识（可选）。不设置时客户端不会自动填充 user 字段。
	cfg.DefaultUser = os.Getenv("DIFY_DEFAULT_USER")

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
