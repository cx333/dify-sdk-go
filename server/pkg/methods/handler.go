package methods

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gofiber/fiber/v3"
	"github.com/wgl/dify-api/server/pkg/response"
)

// MethodRequest POST 请求体。
type MethodRequest struct {
	Method string          `json:"method"`
	Params json.RawMessage `json:"params"`
}

// NewMethodHandler 创建 POST /api/v1/methods handler。
// timeout 控制每个方法调用的最大执行时间。
// maxConcurrent 控制同时执行的方法数上限，0 表示不限。
func NewMethodHandler(registry *MethodRegistry, timeout time.Duration, maxConcurrent int) fiber.Handler {
	sem := make(chan struct{}, maxConcurrent)

	return func(c fiber.Ctx) error {
		var req MethodRequest
		if err := c.Bind().JSON(&req); err != nil {
			return response.MethodFail(c, fiber.StatusBadRequest,
				"INVALID_PARAMS", "请求体 JSON 格式无效")
		}

		if req.Method == "" {
			return response.MethodFail(c, fiber.StatusBadRequest,
				"INVALID_PARAMS", "请求体必须包含 method 字段")
		}

		def, ok := registry.Get(req.Method)
		if !ok {
			return response.MethodFail(c, fiber.StatusNotFound,
				"METHOD_NOT_FOUND", "方法 "+req.Method+" 未注册")
		}

		// 并发控制：非阻塞获取信号量，满时立即返回 429
		if maxConcurrent > 0 {
			select {
			case sem <- struct{}{}:
				defer func() { <-sem }()
			default:
				return response.MethodFail(c, fiber.StatusTooManyRequests,
					"TOO_MANY_REQUESTS", "服务器繁忙，请稍后重试")
			}
		}

		ctx, cancel := context.WithTimeout(c.Context(), timeout)
		defer cancel()

		result, err := def.Handler(ctx, req.Params)
		if err != nil {
			if ctx.Err() == context.DeadlineExceeded {
				return response.MethodFail(c, fiber.StatusGatewayTimeout,
					"METHOD_TIMEOUT", "方法执行超时")
			}
			return response.MethodFail(c, fiber.StatusInternalServerError,
				"METHOD_ERROR", err.Error())
		}

		return response.MethodOK(c, result)
	}
}

// NewListHandler 创建 GET /api/v1/methods handler。
func NewListHandler(registry *MethodRegistry) fiber.Handler {
	return func(c fiber.Ctx) error {
		return response.OK(c, registry.List())
	}
}
