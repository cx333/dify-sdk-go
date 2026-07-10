package difysdk_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/wgl/dify-sdk/client"
	"github.com/wgl/dify-sdk/config"
	"github.com/wgl/dify-sdk/di"
	"github.com/wgl/dify-sdk/store"
)

// TestIntegration_FullChain 验证完整调用链：Config → DI → API Client → Store。
func TestIntegration_FullChain(t *testing.T) {
	// 准备 mock Dify 服务
	var reqCount atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		reqCount.Add(1)
		switch r.URL.Path {
		case "/info":
			json.NewEncoder(w).Encode(client.AppInfo{
				Name: "测试应用",
				Mode: "chat",
			})
		case "/datasets":
			json.NewEncoder(w).Encode(client.DatasetListResponse{
				Data:  []client.Dataset{{ID: "ds-1", Name: "知识库1"}},
				Total: 1,
			})
		case "/workflows/run":
			json.NewEncoder(w).Encode(client.WorkflowBlockingResponse{
				TaskID:        "task-1",
				WorkflowRunID: "run-1",
				Data: &client.WorkflowResult{
					ID:     "run-1",
					Status: "succeeded",
				},
			})
		case "/chat-messages":
			json.NewEncoder(w).Encode(client.ChatCompletionResponse{
				MessageID:      "msg-1",
				ConversationID: "conv-1",
				Answer:         "集成测试通过",
			})
		default:
			w.WriteHeader(404)
		}
	}))
	defer srv.Close()

	// 1. 配置加载
	cfg := &config.Config{
		BaseURL:    srv.URL,
		APIKeys:    []string{"key1", "key2"},
		Timeout:    10 * time.Second,
		MaxRetries: 3,
	}

	// 2. DI 容器
	container, err := di.BuildContainer(cfg)
	if err != nil {
		t.Fatalf("BuildContainer 失败: %v", err)
	}

	// 3. 验证依赖注入
	if err := container.Invoke(func(h *client.HTTPClient) {
		if h == nil || h.BaseURL() != srv.URL {
			t.Fatal("HTTPClient 注入失败")
		}

		// 4. 测试各 API Client
		chatClient := client.NewChatClient(h, "")
		wfClient := client.NewWorkflowClient(h, "")
		kbClient := client.NewKnowledgeClient(h)

		info, err := chatClient.GetAppInfo(context.Background())
		if err != nil {
			t.Fatalf("GetAppInfo 失败: %v", err)
		}
		if info.Mode != "chat" {
			t.Errorf("mode = %q", info.Mode)
		}

		datasets, err := kbClient.ListDatasets(context.Background(), 1, 20)
		if err != nil {
			t.Fatalf("ListDatasets 失败: %v", err)
		}
		if len(datasets.Data) != 1 {
			t.Errorf("datasets = %d", len(datasets.Data))
		}

		wfResp, err := wfClient.Run(context.Background(), client.WorkflowRunRequest{
			Inputs: map[string]any{"text": "hello"},
			User:   "test-user",
		})
		if err != nil {
			t.Fatalf("Run 失败: %v", err)
		}
		if wfResp.Data.Status != "succeeded" {
			t.Errorf("status = %q", wfResp.Data.Status)
		}

		chatResp, err := chatClient.SendMessage(context.Background(), client.ChatRequest{
			Query:  "hi",
			User:   "test-user",
			Inputs: map[string]any{},
		})
		if err != nil {
			t.Fatalf("SendMessage 失败: %v", err)
		}
		if chatResp.Answer != "集成测试通过" {
			t.Errorf("answer = %q", chatResp.Answer)
		}
	}); err != nil {
		t.Fatalf("Invoke 失败: %v", err)
	}

	// 5. 验证 Store 预加载
	s := store.NewMemoryStore()
	err = s.Preload(context.Background(), cfg.APIKeys,
		func(ctx context.Context, hc *client.HTTPClient) ([]store.AppMeta, error) {
			return []store.AppMeta{
				{ID: "app-1", Name: "应用1", Mode: "chat"},
				{ID: "app-2", Name: "应用2", Mode: "workflow"},
			}, nil
		},
		cfg.BaseURL, cfg.Timeout, 2)
	if err != nil {
		t.Fatalf("Preload 失败: %v", err)
	}
	all := s.All()
	if len(all) != 2 {
		t.Errorf("store.All() = %d, want 2", len(all))
	}
	byMode := s.GetByMode("chat")
	// 相同 ID 的 app 重复 upsert 会原地替换，不产生重复条目
	if len(byMode) != 1 || byMode[0].Name != "应用1" {
		t.Errorf("GetByMode(chat) len=%d name=%q", len(byMode), byMode[0].Name)
	}
}
