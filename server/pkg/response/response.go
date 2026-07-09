// Package response 提供统一的 HTTP JSON 响应格式。
//
// 两种格式：
//   - API（普通接口）：前端/外部调用，使用 code/message/data
//   - Method（Dify 调用）：Dify 平台回调，使用 success/data/error
package response

import "github.com/gofiber/fiber/v3"

// ----- API 响应（普通接口）-----

// API 成功 — code=0, message="ok"。
func OK(c fiber.Ctx, data any) error {
	return c.JSON(apiResp{Code: 0, Message: "ok", Data: data})
}

// API 失败 — 自定义 code 和 message。
func Fail(c fiber.Ctx, status int, code int, msg string) error {
	return c.Status(status).JSON(apiResp{Code: code, Message: msg})
}

type apiResp struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    any    `json:"data,omitempty"`
}

// ----- Method 响应（Dify 平台回调）-----

// MethodError 方法调用错误明细。
type MethodError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// MethodOK 方法执行成功。
func MethodOK(c fiber.Ctx, data any) error {
	return c.JSON(methodResp{Success: true, Data: data})
}

// MethodFail 方法执行失败。
func MethodFail(c fiber.Ctx, status int, code, msg string) error {
	return c.Status(status).JSON(methodResp{
		Success: false,
		Error:   &MethodError{Code: code, Message: msg},
	})
}

type methodResp struct {
	Success bool         `json:"success"`
	Data    any          `json:"data,omitempty"`
	Error   *MethodError `json:"error,omitempty"`
}
