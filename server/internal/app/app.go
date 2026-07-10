// Package app 负责应用程序启动、依赖组装和生命周期管理。
package app

import (
	"context"
	"path/filepath"
	"strconv"
	"time"

	"github.com/wgl/dify-api/server/internal/config"
	"github.com/wgl/dify-api/server/internal/di"
	"github.com/wgl/dify-api/server/internal/logger"
	"github.com/wgl/dify-api/server/internal/router"
	"github.com/wgl/dify-api/server/pkg/methods"
	"github.com/wgl/dify-sdk/client"
	sdkconfig "github.com/wgl/dify-sdk/config"
	"github.com/wgl/dify-sdk/store"
)

// App 封装应用程序顶级依赖，管理启动和关闭。
type App struct {
	Log        *logger.Logger
	Config     *config.ServerConfig
	registry   *methods.MethodRegistry
	httpClient *client.HTTPClient  // 默认 HTTPClient（pool[0]），向后兼容
	clientPool *di.ClientPool      // 多 Key 客户端池
	appStore   *store.MemoryStore  // 应用元数据缓存
}

// New 创建 App 实例，初始化日志和配置。
func New() *App {
	cfg := config.Load()
	log := initLogger(cfg)

	return &App{
		Log:      log,
		Config:   cfg,
		registry: methods.NewMethodRegistry(),
	}
}

// Registry 返回 MethodRegistry，供外部注册业务方法。
func (a *App) Registry() *methods.MethodRegistry {
	return a.registry
}

// HTTPClient 返回默认的 Dify HTTP 客户端（pool[0]），向后兼容。
// 仅在 Run() 之后可用；在此之前返回 nil。
func (a *App) HTTPClient() *client.HTTPClient {
	return a.httpClient
}

// HTTPClientByIndex 按索引选择 HTTP 客户端。
// 仅在 Run() 之后可用；索引越界返回 nil。
func (a *App) HTTPClientByIndex(idx int) *client.HTTPClient {
	if a.clientPool == nil {
		return nil
	}
	return a.clientPool.Get(idx)
}

// AppStore 返回应用元数据缓存，仅在 Run() 之后可用。
func (a *App) AppStore() *store.MemoryStore {
	return a.appStore
}

// Run 启动 HTTP 服务，阻塞直到进程收到信号。
func (a *App) Run() error {
	a.Log.Infof("Dify 服务启动中...")

	// 加载 SDK 配置
	sdkCfg, err := sdkconfig.Load(".env")
	if err != nil {
		a.Log.Fatalf("加载配置失败: %v", err)
	}
	a.Log.Info(context.Background(), "SDK 配置加载成功",
		"base_url", sdkCfg.BaseURL,
		"key_count", len(sdkCfg.APIKeys),
	)

	// 构建 DI 容器（内含 ClientPool + 默认 HTTPClient）
	container, err := di.BuildContainer(sdkCfg, a.Config)
	if err != nil {
		a.Log.Fatalf("构建容器失败: %v", err)
	}

	// 从容器获取 ClientPool 和默认 HTTPClient
	var clientPool *di.ClientPool
	var httpClient *client.HTTPClient
	if err := container.Invoke(func(p *di.ClientPool, c *client.HTTPClient) {
		clientPool = p
		httpClient = c
	}); err != nil {
		a.Log.Fatalf("获取 HTTPClient 失败: %v", err)
	}
	a.clientPool = clientPool
	a.httpClient = httpClient

	// 预加载应用元数据：对每个 Key 调用 /info，缓存到 MemoryStore
	a.appStore = preloadApps(sdkCfg, a.Log)

	a.Log.Info(context.Background(), "DI 容器初始化完成",
		"outbound_rps", a.Config.OutboundRPS,
		"max_concurrent_methods", a.Config.MaxConcurrentMethods,
		"app_count", len(a.appStore.All()),
	)

	// 创建 Fiber 应用
	fiberApp := router.Setup(router.Params{
		Log:           a.Log,
		Env:           a.Config.Env,
		Registry:      a.registry,
		MethodTimeout: a.Config.MethodTimeout,
		MaxConcurrent: a.Config.MaxConcurrentMethods,
		AppStore:      a.appStore,
		ClientPool:    a.clientPool,
	})

	// 打印已注册路由（仅 GET/POST，跳过 Fiber 自动注入的 OPTIONS/405 路由）
	for _, routes := range fiberApp.Stack() {
		for _, r := range routes {
			switch r.Method {
			case "GET", "POST":
				a.Log.Infof("  %s %s", r.Method, r.Path)
			}
		}
	}

	// 启动
	a.Log.Infof("服务启动，监听端口 :%s", a.Config.Port)
	return fiberApp.Listen(":" + a.Config.Port)
}

// preloadApps 遍历所有 API Key，调用 /info 获取应用元数据。
func preloadApps(cfg *sdkconfig.Config, log *logger.Logger) *store.MemoryStore {
	s := store.NewMemoryStore()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	for i, key := range cfg.APIKeys {
		c := client.NewHTTPClient(cfg.BaseURL, key, cfg.Timeout, client.DefaultRetryConfig(3))
		chatClient := client.NewChatClient(c, "")
		info, err := chatClient.GetAppInfo(ctx)
		if err != nil {
			log.Warn(ctx, "获取应用信息失败",
				"key_index", i,
				"error", err.Error(),
			)
			continue
		}

		s.Upsert(&store.AppMeta{
			ID:          strconv.Itoa(i),
			Name:        info.Name,
			Mode:        info.Mode,
			Description: info.Description,
			APIKey:      key,
			Tags:        info.Tags,
			UpdatedAt:   time.Now(),
		})

		modeLabel := client.AppMode(info.Mode).Label()
		log.Info(ctx, "发现应用",
			"key_index", i,
			"name", info.Name,
			"mode", info.Mode,
			"mode_label", modeLabel,
		)
	}

	return s
}

func initLogger(cfg *config.ServerConfig) *logger.Logger {
	logCfg := logger.Config{
		Level:  cfg.LogLevel,
		Format: cfg.LogFormat,
	}

	if cfg.LogFile.Enabled {
		logCfg.File = &logger.FileConfig{
			Filename:   filepath.Join(cfg.LogFile.Dir, "server.log"),
			MaxSize:    cfg.LogFile.MaxSize,
			MaxAge:     cfg.LogFile.MaxAge,
			MaxBackups: cfg.LogFile.MaxBackups,
		}
	}

	return logger.New(logCfg)
}
