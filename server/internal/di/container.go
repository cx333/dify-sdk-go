// Package di 管理 server 层依赖注入。注册 Config、ClientPool、HTTPClient、MethodRegistry 等单例组件。
package di

import (
	"net/http"
	"time"

	"github.com/wgl/dify-api/server/internal/config"
	"github.com/wgl/dify-api/server/pkg/methods"
	"github.com/wgl/dify-api/server/pkg/ratelimit"
	"github.com/wgl/dify-sdk/client"
	sdkconfig "github.com/wgl/dify-sdk/config"
	"go.uber.org/dig"
)

// Container 封装 dig.Container，提供类型安全的 Invoke 方法。
type Container struct {
	inner *dig.Container
}

// ClientPool 管理多个 API Key 对应的 HTTPClient 实例池。
// 每个 Key 一个独立 Client，并发安全。
type ClientPool struct {
	clients []*client.HTTPClient
}

// Get 按索引获取 HTTPClient；越界返回 nil。
func (p *ClientPool) Get(idx int) *client.HTTPClient {
	if idx < 0 || idx >= len(p.clients) {
		return nil
	}
	return p.clients[idx]
}

// Len 返回池中的客户端数量。
func (p *ClientPool) Len() int {
	return len(p.clients)
}

// BuildContainer 创建 server 层 DI 容器。
// 注册顺序：SDK Config → ClientPool → 默认 HTTPClient → MethodRegistry。
func BuildContainer(sdkCfg *sdkconfig.Config, srvCfg *config.ServerConfig) (*Container, error) {
	c := dig.New()

	// 全局 SDK 配置
	if err := c.Provide(func() *sdkconfig.Config { return sdkCfg }); err != nil {
		return nil, err
	}

	// ClientPool — 每个 API Key 一个 HTTPClient，共享 transport
	if err := c.Provide(func(cfg *sdkconfig.Config) *ClientPool {
		return newClientPool(cfg, srvCfg)
	}); err != nil {
		return nil, err
	}

	// 默认 HTTPClient — 始终指向 pool[0]，向后兼容
	if err := c.Provide(func(pool *ClientPool) *client.HTTPClient {
		return pool.Get(0)
	}); err != nil {
		return nil, err
	}

	// 方法注册表
	if err := c.Provide(func() *methods.MethodRegistry {
		return methods.NewMethodRegistry()
	}); err != nil {
		return nil, err
	}

	return &Container{inner: c}, nil
}

// Invoke 调用 dig.Container.Invoke，用于从容器提取已注册的依赖。
func (c *Container) Invoke(f any) error {
	return c.inner.Invoke(f)
}

// newClientPool 为每个 API Key 创建一个 HTTPClient。
// 所有 Client 共享同一个 transport 实现连接池复用。
func newClientPool(cfg *sdkconfig.Config, srvCfg *config.ServerConfig) *ClientPool {
	baseTransport := &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}

	var transport http.RoundTripper = baseTransport
	if srvCfg.OutboundRPS > 0 {
		transport = ratelimit.NewTransport(baseTransport, srvCfg.OutboundRPS, srvCfg.OutboundBurst)
	}

	clients := make([]*client.HTTPClient, 0, len(cfg.APIKeys))
	for _, key := range cfg.APIKeys {
		c := client.NewHTTPClientWithTransport(
			cfg.BaseURL,
			key,
			cfg.Timeout,
			client.DefaultRetryConfig(3),
			transport,
		)
		clients = append(clients, c)
	}

	return &ClientPool{clients: clients}
}
