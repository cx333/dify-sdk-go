package router

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wgl/dify-api/server/internal/handler"
	"github.com/wgl/dify-api/server/internal/logger"
)

// registerHealthRoutes 注册健康检查路由 /health。
func registerHealthRoutes(app *fiber.App, log *logger.Logger) {
	h := handler.NewHealthHandler(log)
	app.Get("/health", h.Check)
}
