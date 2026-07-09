package router

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/wgl/dify-api/server/pkg/methods"
)

// registerMethodRoutes 注册方法调用与列表路由。
func registerMethodRoutes(app *fiber.App, registry *methods.MethodRegistry, methodTimeout time.Duration, maxConcurrent int) {
	v1 := app.Group("/api/v1")
	v1.Post("/methods", methods.NewMethodHandler(registry, methodTimeout, maxConcurrent))
	v1.Get("/methods", methods.NewListHandler(registry))
}
