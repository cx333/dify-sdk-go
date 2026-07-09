package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wgl/dify-api/server/internal/logger"
	"github.com/wgl/dify-api/server/pkg/response"
)

// HealthHandler 健康检查处理器。
type HealthHandler struct {
	log *logger.Logger
}

// NewHealthHandler 创建 HealthHandler。
func NewHealthHandler(log *logger.Logger) *HealthHandler {
	return &HealthHandler{log: log}
}

// Check 返回服务健康状态。
// GET /health
func (h *HealthHandler) Check(c fiber.Ctx) error {
	h.log.Debug(c.Context(), "health check")
	return response.OK(c, fiber.Map{"status": "ok"})
}
