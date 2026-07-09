package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wgl/dify-api/server/internal/di"
	"github.com/wgl/dify-api/server/internal/handler"
	"github.com/wgl/dify-api/server/internal/logger"
	"github.com/wgl/dify-sdk/store"
)

// registerAPIRoutes 注册 /api/v1 路由组。
func registerAPIRoutes(app *fiber.App, log *logger.Logger) {
	h := handler.NewAPIHandler(log)
	v1 := app.Group("/api/v1")
	v1.Get("/", h.Index)
}

// registerAppRoutes 注册应用列表路由。
func registerAppRoutes(app *fiber.App, appStore *store.MemoryStore, log *logger.Logger) {
	h := handler.NewAppsHandler(appStore, log)
	v1 := app.Group("/api/v1")
	v1.Get("/apps", h.List)
}

// 对话接口
func registerChatRoutes(app *fiber.App, appStore *store.MemoryStore, log *logger.Logger, clientPool *di.ClientPool) {
	h := handler.NewChatHandler(log, clientPool)
	v1 := app.Group("/api/v1")
	v1.Post("/chat", h.Chat)
}
