// Package router 负责注册所有 HTTP 路由和全局中间件。
package router

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/wgl/dify-api/server/internal/di"
	"github.com/wgl/dify-api/server/internal/logger"
	"github.com/wgl/dify-api/server/internal/middleware"
	"github.com/wgl/dify-api/server/pkg/methods"
	"github.com/wgl/dify-sdk/store"
)

// Params 路由初始化参数。
type Params struct {
	Log            *logger.Logger
	Env            string
	Registry       *methods.MethodRegistry
	MethodTimeout  time.Duration
	MaxConcurrent  int
	AppStore       *store.MemoryStore // 应用元数据缓存
	ClientPool     *di.ClientPool     // 多 Key 客户端池
}

// Setup 创建 Fiber 应用，注册全局中间件和所有路由。
func Setup(p Params) *fiber.App {
	app := fiber.New(fiber.Config{
		ServerHeader: "dify-server",
	})

	// 全局中间件（按顺序执行）
	app.Use(recover.New(recover.Config{
		EnableStackTrace: p.Env != "production",
	}))
	app.Use(middleware.RequestID())
	app.Use(middleware.RequestLogging(p.Log))

	// 路由分组注册
	registerHealthRoutes(app, p.Log)
	registerAPIRoutes(app, p.Log)
	registerAppRoutes(app, p.AppStore, p.Log)
	registerChatRoutes(app, p.AppStore, p.Log, p.ClientPool)
	registerMethodRoutes(app, p.Registry, p.MethodTimeout, p.MaxConcurrent)

	return app
}
