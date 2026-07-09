package handler

import (
	"github.com/gofiber/fiber/v3"
	"github.com/wgl/dify-api/server/internal/logger"
	"github.com/wgl/dify-api/server/pkg/response"
)

// APIHandler API v1 处理器。
type APIHandler struct {
	log *logger.Logger
}

// NewAPIHandler 创建 APIHandler。
func NewAPIHandler(log *logger.Logger) *APIHandler {
	return &APIHandler{log: log}
}

// Index API v1 根路径，返回服务信息。
// GET /api/v1/
func (h *APIHandler) Index(c fiber.Ctx) error {
	h.log.Debug(c.Context(), "api index")
	return response.OK(c, fiber.Map{"message": "Dify server ready"})
}
