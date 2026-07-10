package client

import (
	"context"
	"encoding/json"
	"fmt"
	"time"
)

// ChatStreamEvent 流式对话的结构化事件，已分离 think/answer、提取用量、计时。
type ChatStreamEvent struct {
	Type           string `json:"type"`                      // start / thinking / answer / done / error
	Content        string `json:"content,omitempty"`         // 文本片段
	ConversationID string `json:"conversation_id,omitempty"` // 会话 ID
	MessageID      string `json:"message_id,omitempty"`      // 消息 ID
	Usage          *Usage `json:"usage,omitempty"`           // token 用量（仅 done 事件）
	ElapsedMs      int64  `json:"elapsed_ms,omitempty"`      // 耗时毫秒（done 事件为总耗时，start 为 0）
	Error          string `json:"error,omitempty"`           // 错误信息（仅 error 事件）
}

// SendChatStream 发送对话消息，返回结构化事件流。
// 自动完成 <think> 标签分离、token 用量提取、请求计时。
func (c *ChatClient) SendChatStream(ctx context.Context, req ChatRequest) (<-chan ChatStreamEvent, <-chan error) {
	eventCh := make(chan ChatStreamEvent, 100)
	errCh := make(chan error, 1)

	go func() {
		defer close(eventCh)
		defer close(errCh)

		rawEvents, rawErrs := c.SendMessageStream(ctx, req)

		var (
			parser    ThinkParser
			convID    string
			msgID     string
			usage     *Usage
			startedAt = time.Now()
		)

		// 流开始
		eventCh <- ChatStreamEvent{Type: "start"}

		for {
			select {
			case ev, ok := <-rawEvents:
				if !ok {
					eventCh <- ChatStreamEvent{
						Type:           "done",
						ConversationID: convID,
						MessageID:      msgID,
						Usage:          usage,
						ElapsedMs:      time.Since(startedAt).Milliseconds(),
					}
					return
				}

				if ev.ConversationID != "" {
					convID = ev.ConversationID
				}
				if ev.MessageID != "" {
					msgID = ev.MessageID
				}

				if u := extractUsage(ev.Metadata); u != nil {
					usage = u
				}

				if ev.Status != 0 || ev.Code != "" {
					eventCh <- ChatStreamEvent{
						Type:  "error",
						Error: ev.Message,
					}
					return
				}

				for _, te := range parser.Feed(ev.Answer) {
					eventCh <- ChatStreamEvent{
						Type:           te.Type,
						Content:        te.Content,
						ConversationID: convID,
						MessageID:      msgID,
					}
				}

			case err := <-rawErrs:
				if err != nil {
					eventCh <- ChatStreamEvent{
						Type:  "error",
						Error: err.Error(),
					}
				}
				return
			}
		}
	}()

	return eventCh, errCh
}

// extractUsage 从 SSE metadata 中提取 token 用量。
// Dify 在 message_end 事件中通过 metadata.usage 返回用量。
func extractUsage(raw json.RawMessage) *Usage {
	if len(raw) == 0 {
		return nil
	}

	// 标准格式：{"usage": {...}, "retriever_resources": [...]}
	var meta ChatMetadata
	if err := json.Unmarshal(raw, &meta); err == nil && meta.Usage != nil {
		return meta.Usage
	}

	// 兼容格式：{"usage": {...}}（无外层 metadata 包裹）
	var alt struct {
		Usage Usage `json:"usage"`
	}
	if err := json.Unmarshal(raw, &alt); err == nil && alt.Usage.TotalTokens > 0 {
		return &alt.Usage
	}

	return nil
}

// DebugRawEvent 打印原始 SSE 事件的 metadata 到 stdout，用于排查 Dify 返回格式。
//
// SendChatStream 默认不会调用此函数。如需调试，在调用 SendChatStream 之前，
// 自行在事件循环中调用 DebugRawEvent。
//
//	events, errs := chat.SendChatStream(ctx, req)
//	for ev := range events {
//	    DebugRawEvent(ev.Type, ev.Content, nil)
//	    // ...
//	}
func DebugRawEvent(event, answer string, metadata json.RawMessage) {
	if len(metadata) > 0 {
		fmt.Printf("[DEBUG] event=%s answer_len=%d metadata=%s\n", event, len(answer), string(metadata))
	}
}
