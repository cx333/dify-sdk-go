// Package di 使用 dig.Container 管理依赖注入，组装 SDK 各组件。
// 所有 Provider 通过构造函数自动注册，dig 负责解析依赖图并实例化。
package di

import (
	"fmt"

	"github.com/wgl/dify-sdk/client"
	"github.com/wgl/dify-sdk/config"
	"go.uber.org/dig"
)

// BuildContainer 创建并配置 dig.Container，注册所有 SDK 依赖。
//
// 注册的 Provider（按依赖顺序）：
//   - *config.Config — 全局配置
//   - *client.HTTPClient — 共享 HTTP 传输层（使用第一个 API Key）
func BuildContainer(cfg *config.Config) (*dig.Container, error) {
	c := dig.New()

	// 注册全局配置
	if err := c.Provide(func() *config.Config { return cfg }); err != nil {
		return nil, fmt.Errorf("di: register Config failed: %w", err)
	}

	// 注册 HTTPClient（单例，连接池复用）
	if err := c.Provide(provideHTTPClient); err != nil {
		return nil, fmt.Errorf("di: register HTTPClient failed: %w", err)
	}

	return c, nil
}

// BuildContainerWithKey 使用指定的 API Key 构建容器。
// 用于多 Key 场景下为每个 Key 创建独立的 HTTPClient。
func BuildContainerWithKey(cfg *config.Config, apiKey string) (*dig.Container, error) {
	c := dig.New()

	if err := c.Provide(func() *config.Config { return cfg }); err != nil {
		return nil, fmt.Errorf("di: register Config failed: %w", err)
	}

	if err := c.Provide(func(cfg *config.Config) *client.HTTPClient {
		return client.NewHTTPClient(cfg.BaseURL, apiKey, cfg.Timeout, client.DefaultRetryConfig(cfg.MaxRetries))
	}); err != nil {
		return nil, fmt.Errorf("di: register HTTPClient failed: %w", err)
	}

	return c, nil
}

// provideHTTPClient 是 HTTPClient 的 Provider 函数。
// 使用 cfg 中的第一个 API Key 作为默认认证凭据。
func provideHTTPClient(cfg *config.Config) *client.HTTPClient {
	key := ""
	if len(cfg.APIKeys) > 0 {
		key = cfg.APIKeys[0]
	}
	return client.NewHTTPClient(cfg.BaseURL, key, cfg.Timeout, client.DefaultRetryConfig(cfg.MaxRetries))
}
