// Package middleware 提供 Fiber v3 通用中间件。
// 包含请求 ID 注入、结构化请求日志等。
package middleware

import (
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/wgl/dify-api/server/internal/logger"
)

// RequestID 注入唯一请求 ID 到 context 和响应头。
// 优先使用客户端传入的 X-Request-ID，否则自动生成 UUID v4。
func RequestID() fiber.Handler {
	return func(c fiber.Ctx) error {
		rid := c.Get("X-Request-ID")
		if rid == "" {
			rid = uuid.New().String()
		}
		c.Set("X-Request-ID", rid)
		c.SetContext(logger.WithRequestID(c.Context(), rid))
		return c.Next()
	}
}

// RequestLogging 使用结构化日志记录每个 HTTP 请求。
// 输出字段：method, path, status, duration, request_id。
func RequestLogging(log *logger.Logger) fiber.Handler {
	return func(c fiber.Ctx) error {
		start := time.Now()

		err := c.Next()

		duration := time.Since(start)
		status := c.Response().StatusCode()

		ctx := c.Context()

		logFunc := log.Info
		if status >= 500 {
			logFunc = log.Error
		} else if status >= 400 {
			logFunc = log.Warn
		}

		logFunc(ctx, "http request",
			"method", c.Method(),
			"path", c.Path(),
			"status", status,
			"duration_ms", duration.Milliseconds(),
		)

		return err
	}
}
