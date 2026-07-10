package handler

import (
	"bufio"
	"encoding/json"
	"fmt"
	"strconv"

	"github.com/gofiber/fiber/v3"
	"github.com/wgl/dify-api/server/internal/di"
	"github.com/wgl/dify-api/server/internal/logger"
	"github.com/wgl/dify-api/server/pkg/response"
	"github.com/wgl/dify-sdk/client"
)

// ChatHandler 对话相关请求处理器。
type ChatHandler struct {
	log        *logger.Logger
	clientPool *di.ClientPool
}

// NewChatHandler 创建 ChatHandler。
func NewChatHandler(log *logger.Logger, clientPool *di.ClientPool) *ChatHandler {
	return &ChatHandler{log: log, clientPool: clientPool}
}

// chatRequest 对话请求体。key_index 兼容字符串和数字。
type chatRequest struct {
	Query    string `json:"query"`
	KeyIndex int    `json:"key_index"`
}

// UnmarshalJSON 自定义反序列化，使 key_index 同时支持字符串和数字。
func (r *chatRequest) UnmarshalJSON(data []byte) error {
	var raw struct {
		Query    string      `json:"query"`
		KeyIndex interface{} `json:"key_index"`
	}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	r.Query = raw.Query
	switch v := raw.KeyIndex.(type) {
	case float64:
		r.KeyIndex = int(v)
	case string:
		r.KeyIndex, _ = strconv.Atoi(v)
	}
	return nil
}

// Chat 向 Dify 发送对话消息，流式 SSE 返回。
// POST /api/v1/chat
func (h *ChatHandler) Chat(c fiber.Ctx) error {
	var req chatRequest
	body := c.Body()
	if err := json.Unmarshal(body, &req); err != nil {
		h.log.Warn(c.Context(), "chat 请求解析失败", "body", string(body), "error", err.Error())
		return response.Fail(c, fiber.StatusBadRequest, 40001, "请求体 JSON 解析失败: "+err.Error())
	}
	if req.Query == "" {
		return response.Fail(c, fiber.StatusBadRequest, 40001, "请求参数无效，query 为必填项")
	}

	c2 := h.clientPool.Get(req.KeyIndex)
	if c2 == nil {
		return response.Fail(c, fiber.StatusBadRequest, 40002, "key_index 无效，应用不存在")
	}

	chatClient := client.NewChatClient(c2, "")
	eventCh, _ := chatClient.SendChatStream(c.Context(), client.ChatRequest{
		Query:  req.Query,
		User:   "dify-server",
		Inputs: map[string]any{},
	})

	// SSE 流式响应
	c.Set("Content-Type", "text/event-stream")
	c.Set("Cache-Control", "no-cache")
	c.Set("Connection", "keep-alive")

	c.RequestCtx().SetBodyStreamWriter(func(w *bufio.Writer) {
		for ev := range eventCh {
			h.log.Debug(c.Context(), "chat sse event",
				"type", ev.Type,
				"content_len", len(ev.Content),
				"elapsed_ms", ev.ElapsedMs,
				"has_usage", ev.Usage != nil,
			)
			writeSSE(w, ev.Type, ev)
			w.Flush()
		}
	})

	return nil
}

// writeSSE 写入一条 SSE 事件。
func writeSSE(w *bufio.Writer, event string, data any) {
	b, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, string(b))
}
