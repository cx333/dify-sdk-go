// Dify 服务入口 —— 最薄启动层，所有逻辑在 internal/app 中。
package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/wgl/dify-api/server/internal/app"
	"github.com/wgl/dify-api/server/pkg/methods"
)

func main() {
	application := app.New()

	// 注册业务方法。方法可通过 application.HTTPClient() 获得共享的
	// Dify HTTP 客户端（单例、连接池复用、受出站限流保护）。
	registerExamples(application)

	if err := application.Run(); err != nil {
		log.Fatalf("服务异常: %v", err)
	}
}

// registerExamples 注册示例方法供 Dify 调用。
func registerExamples(a *app.App) {
	r := a.Registry()

	// ping — 健康检查
	r.Register("ping", &methods.MethodDef{
		Name:         "ping",
		Description:  "健康检查：返回 pong",
		InputSchema:  json.RawMessage(`{"type":"object","properties":{}}`),
		OutputSchema: json.RawMessage(`{"type":"object","properties":{"message":{"type":"string"}}}`),
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			return map[string]string{"message": "pong"}, nil
		},
	})

	// call_dify — 示例：通过共享 HTTPClient 调用 Dify API。
	// 支持 key_index 参数选择目标应用（默认 0）。
	r.Register("call_dify", &methods.MethodDef{
		Name:         "call_dify",
		Description:  "向 Dify 发送对话消息并返回结果",
		InputSchema:  json.RawMessage(`{"type":"object","properties":{"query":{"type":"string"},"key_index":{"type":"integer","default":0}},"required":["query"]}`),
		OutputSchema: json.RawMessage(`{"type":"object","properties":{"answer":{"type":"string"}}}`),
		Handler: func(ctx context.Context, params json.RawMessage) (any, error) {
			var p struct {
				Query    string `json:"query"`
				KeyIndex int    `json:"key_index"`
			}
			json.Unmarshal(params, &p)

			c := a.HTTPClientByIndex(p.KeyIndex)
			if c == nil {
				c = a.HTTPClient()
			}
			var result struct {
				Answer string `json:"answer"`
			}
			req := map[string]any{
				"inputs":          map[string]any{},
				"query":           p.Query,
				"response_mode":   "blocking",
				"conversation_id": "",
				"user":            "dify-server",
			}
			if err := c.Do(ctx, "POST", "/chat-messages", req, &result); err != nil {
				return nil, err
			}
			return result, nil
		},
	})
}
