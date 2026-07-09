// Package client 实现 Dify REST API 的 Go 客户端。
//
// 核心能力：
//   - HTTPClient：连接池复用、指数退避重试、SSE 流式响应的 HTTP 传输层
//   - ChatClient：对话型应用 API（chat-messages、会话管理、消息反馈）
//   - WorkflowClient：工作流应用 API（执行、查询、日志）
//   - KnowledgeClient：知识库 API（数据集 CRUD、文档管理、段落检索）
//   - FileClient：文件上传与下载
//
// 所有 Client 共享底层 HTTPClient 实例，通过依赖注入组装。
package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// DifyError Dify API 的标准错误响应。
// 兼容 HTTP 错误（4xx/5xx）和 SSE 流内错误。
type DifyError struct {
	StatusCode int    `json:"-"`    // HTTP 状态码（非 JSON 字段）
	Code       string `json:"code"` // 机器可读错误码，如 "invalid_param"
	Message    string `json:"message"` // 人类可读错误描述
	Status     int    `json:"status"`  // 响应中的 HTTP 状态码副本
}

// Error 实现 error 接口。
func (e *DifyError) Error() string {
	return fmt.Sprintf("dify: [%d] %s: %s", e.StatusCode, e.Code, e.Message)
}

// NewDifyError 从 HTTP 响应体构造 DifyError。
func NewDifyError(statusCode int, body []byte) *DifyError {
	e := &DifyError{StatusCode: statusCode}
	if err := json.Unmarshal(body, e); err != nil {
		e.Code = "parse_error"
		e.Message = string(body)
	}
	if e.Status == 0 {
		e.Status = statusCode
	}
	return e
}

// SseError SSE 流内的错误事件。
type SseError struct {
	Status  int    `json:"status"`
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Error 实现 error 接口。
func (e *SseError) Error() string {
	return fmt.Sprintf("sse: [%d] %s: %s", e.Status, e.Code, e.Message)
}

// Response Dify API 通用分页响应包装（泛型）。
type Response[T any] struct {
	Result  string `json:"result,omitempty"`
	Data    T      `json:"data,omitempty"`
	HasMore bool   `json:"has_more,omitempty"`
	Limit   int    `json:"limit,omitempty"`
	Total   int    `json:"total,omitempty"`
	Page    int    `json:"page,omitempty"`
}

// SuccessResult 操作类接口的通用成功响应。
type SuccessResult struct {
	Result string `json:"result"`
}

// SseEvent 解析后的 SSE（Server-Sent Event）事件。
// 兼容 Chat、Workflow、Chatflow 三种应用的流式事件格式。
type SseEvent struct {
	Event          string          `json:"event"`                      // 事件类型：message、workflow_started、node_finished 等
	TaskID         string          `json:"task_id,omitempty"`          // 任务 ID，用于停止响应
	MessageID      string          `json:"message_id,omitempty"`       // 消息 ID
	ConversationID string          `json:"conversation_id,omitempty"`  // 会话 ID（Chat 应用）
	WorkflowRunID  string          `json:"workflow_run_id,omitempty"`  // 工作流运行 ID
	Answer         string          `json:"answer,omitempty"`           // 文本回复片段
	Audio          string          `json:"audio,omitempty"`            // TTS 音频（Base64）
	CreatedAt      int64           `json:"created_at,omitempty"`       // 创建时间戳
	Status         int             `json:"status,omitempty"`           // 错误状态码
	Code           string          `json:"code,omitempty"`             // 错误码
	Message        string          `json:"message,omitempty"`          // 错误描述
	Metadata       json.RawMessage `json:"metadata,omitempty"`         // 元数据（含 token 用量）
	Data           json.RawMessage `json:"data,omitempty"`             // 嵌套载荷（工作流/节点事件）
}

// MaskKey 对 API Key 进行脱敏处理，用于日志输出。
// 长度 ≤8 时返回 "***"，否则显示前 8 位 + "***"。
func MaskKey(key string) string {
	if len(key) <= 8 {
		return "***"
	}
	return key[:8] + "***"
}

// RetryableStatusCodes 需要自动重试的 HTTP 状态码列表。
var RetryableStatusCodes = []int{
	http.StatusTooManyRequests,    // 429 限流
	http.StatusBadGateway,         // 502 网关错误
	http.StatusServiceUnavailable, // 503 服务不可用
	http.StatusGatewayTimeout,     // 504 网关超时
}
