package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// newTestServer 创建模拟 Dify API 的 httptest 服务器。
func newTestServer(handler http.HandlerFunc) (*httptest.Server, *HTTPClient) {
	srv := httptest.NewServer(handler)
	client := NewHTTPClient(srv.URL, "test-key", 10*time.Second, DefaultRetryConfig())
	return srv, client
}

// ---- ChatClient 测试 ----

func TestChatClient_SendMessage(t *testing.T) {
	srv, httpClient := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/chat-messages" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(ChatCompletionResponse{
			Event:          "message",
			MessageID:      "msg-1",
			ConversationID: "conv-1",
			Mode:           "chat",
			Answer:         "你好，有什么可以帮助你的？",
		})
	})
	defer srv.Close()

	c := NewChatClient(httpClient)
	resp, err := c.SendMessage(context.Background(), ChatRequest{
		Query:        "你好",
		User:         "user-1",
		ResponseMode: "blocking",
		Inputs:       map[string]interface{}{},
	})
	if err != nil {
		t.Fatalf("SendMessage 失败: %v", err)
	}
	if resp.Answer != "你好，有什么可以帮助你的？" {
		t.Errorf("Answer = %q", resp.Answer)
	}
	if resp.ConversationID != "conv-1" {
		t.Errorf("ConversationID = %q", resp.ConversationID)
	}
}

func TestChatClient_GetConversations(t *testing.T) {
	srv, httpClient := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(ConversationListResponse{
			Limit:   20,
			HasMore: false,
			Data: []Conversation{
				{ID: "conv-1", Name: "测试会话", Status: "normal"},
			},
		})
	})
	defer srv.Close()

	c := NewChatClient(httpClient)
	resp, err := c.GetConversations(context.Background(), "user-1", "", 20)
	if err != nil {
		t.Fatalf("GetConversations 失败: %v", err)
	}
	if len(resp.Data) != 1 || resp.Data[0].ID != "conv-1" {
		t.Errorf("Data = %+v", resp.Data)
	}
}

func TestChatClient_Feedback(t *testing.T) {
	srv, httpClient := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(SuccessResult{Result: "success"})
	})
	defer srv.Close()

	c := NewChatClient(httpClient)
	err := c.Feedback(context.Background(), "msg-1", FeedbackRequest{
		Rating: "like",
		User:   "user-1",
	})
	if err != nil {
		t.Fatalf("Feedback 失败: %v", err)
	}
}

func TestChatClient_GetAppInfo(t *testing.T) {
	srv, httpClient := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(AppInfo{
			Name: "测试应用",
			Mode: "chat",
			Tags: []string{"test"},
		})
	})
	defer srv.Close()

	c := NewChatClient(httpClient)
	info, err := c.GetAppInfo(context.Background())
	if err != nil {
		t.Fatalf("GetAppInfo 失败: %v", err)
	}
	if info.Mode != "chat" {
		t.Errorf("Mode = %q", info.Mode)
	}
}

// ---- WorkflowClient 测试 ----

func TestWorkflowClient_Run(t *testing.T) {
	srv, httpClient := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(WorkflowBlockingResponse{
			TaskID:        "task-1",
			WorkflowRunID: "run-1",
			Data: &WorkflowResult{
				ID:     "run-1",
				Status: "succeeded",
				Outputs: map[string]interface{}{
					"result": "处理完成",
				},
			},
		})
	})
	defer srv.Close()

	c := NewWorkflowClient(httpClient)
	resp, err := c.Run(context.Background(), WorkflowRunRequest{
		Inputs: map[string]interface{}{"text": "hello"},
		User:   "user-1",
	})
	if err != nil {
		t.Fatalf("Run 失败: %v", err)
	}
	if resp.Data.Status != "succeeded" {
		t.Errorf("Status = %q", resp.Data.Status)
	}
}

func TestWorkflowClient_GetRunDetail(t *testing.T) {
	srv, httpClient := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(WorkflowResult{
			ID:      "run-1",
			Status:  "succeeded",
			Outputs: map[string]interface{}{"text": "ok"},
		})
	})
	defer srv.Close()

	c := NewWorkflowClient(httpClient)
	resp, err := c.GetRunDetail(context.Background(), "run-1")
	if err != nil {
		t.Fatalf("GetRunDetail 失败: %v", err)
	}
	if resp.ID != "run-1" {
		t.Errorf("ID = %q", resp.ID)
	}
}

// ---- KnowledgeClient 测试 ----

func TestKnowledgeClient_CreateDataset(t *testing.T) {
	srv, httpClient := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(Dataset{
			ID:   "ds-1",
			Name: "测试知识库",
		})
	})
	defer srv.Close()

	c := NewKnowledgeClient(httpClient)
	resp, err := c.CreateDataset(context.Background(), CreateDatasetRequest{
		Name: "测试知识库",
	})
	if err != nil {
		t.Fatalf("CreateDataset 失败: %v", err)
	}
	if resp.ID != "ds-1" {
		t.Errorf("ID = %q", resp.ID)
	}
}

func TestKnowledgeClient_ListDatasets(t *testing.T) {
	srv, httpClient := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(DatasetListResponse{
			Data:  []Dataset{{ID: "ds-1", Name: "KB1"}},
			Total: 1,
		})
	})
	defer srv.Close()

	c := NewKnowledgeClient(httpClient)
	resp, err := c.ListDatasets(context.Background(), 1, 20)
	if err != nil {
		t.Fatalf("ListDatasets 失败: %v", err)
	}
	if len(resp.Data) != 1 {
		t.Errorf("len(Data) = %d", len(resp.Data))
	}
}

func TestKnowledgeClient_Retrieve(t *testing.T) {
	srv, httpClient := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(RetrieveResponse{
			Records: []RetrievedSegment{
				{Score: 0.95},
			},
		})
	})
	defer srv.Close()

	c := NewKnowledgeClient(httpClient)
	resp, err := c.RetrieveSegments(context.Background(), "ds-1", RetrieveRequest{
		Query: "搜索关键词",
	})
	if err != nil {
		t.Fatalf("RetrieveSegments 失败: %v", err)
	}
	if len(resp.Records) != 1 || resp.Records[0].Score != 0.95 {
		t.Errorf("Records = %+v", resp.Records)
	}
}

// ---- FileClient 测试 ----

func TestFileClient_Upload(t *testing.T) {
	srv, httpClient := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/files/upload" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(FileUploadResponse{
			ID:        "file-1",
			Name:      "test.txt",
			Size:      100,
			MimeType:  "text/plain",
			Extension: "txt",
		})
	})
	defer srv.Close()

	c := NewFileClient(httpClient)
	resp, err := c.UploadFile(context.Background(), "test.txt", &mockReader{data: []byte("hello")}, "user-1")
	if err != nil {
		t.Fatalf("UploadFile 失败: %v", err)
	}
	if resp.ID != "file-1" || resp.Name != "test.txt" {
		t.Errorf("resp = %+v", resp)
	}
}

// mockReader 实现 io.Reader，用于测试文件上传。
type mockReader struct {
	data []byte
	pos  int
}

func (r *mockReader) Read(p []byte) (n int, err error) {
	if r.pos >= len(r.data) {
		return 0, io.EOF
	}
	n = copy(p, r.data[r.pos:])
	r.pos += n
	if r.pos >= len(r.data) {
		return n, io.EOF
	}
	return n, nil
}

// ---- 错误响应测试 ----

func TestChatClient_ErrorResponse(t *testing.T) {
	srv, httpClient := newTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"code":    "not_chat_app",
			"message": "App mode mismatch",
			"status":  400,
		})
	})
	defer srv.Close()

	c := NewChatClient(httpClient)
	_, err := c.SendMessage(context.Background(), ChatRequest{
		Query: "hi", User: "u1", Inputs: map[string]interface{}{},
	})
	if err == nil {
		t.Fatal("期望错误响应")
	}
	var difyErr *DifyError
	if !errors.As(err, &difyErr) {
		t.Fatalf("期望 *DifyError，实际 %T: %v", err, err)
	}
	if difyErr.Code != "not_chat_app" {
		t.Errorf("Code = %q", difyErr.Code)
	}
}
