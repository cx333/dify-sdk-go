package client

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"
)

func TestDo_Success(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("missing auth header")
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"result": "success"})
	}))
	defer srv.Close()

	c := NewHTTPClient(srv.URL, "test-key", 10*time.Second, DefaultRetryConfig(3))
	var result map[string]string
	err := c.Do(context.Background(), "POST", "/chat-messages", map[string]string{"query": "hi"}, &result)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result["result"] != "success" {
		t.Errorf("got %v", result)
	}
}

func TestDo_ErrorResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		json.NewEncoder(w).Encode(map[string]any{
			"code":    "invalid_param",
			"message": "missing query",
			"status":  400,
		})
	}))
	defer srv.Close()

	c := NewHTTPClient(srv.URL, "test-key", 10*time.Second, DefaultRetryConfig(3))
	err := c.Do(context.Background(), "POST", "/chat-messages", nil, nil)
	if err == nil {
		t.Fatal("expected error")
	}
	difyErr, ok := err.(*DifyError)
	if !ok {
		t.Fatalf("expected *DifyError, got %T", err)
	}
	if difyErr.Code != "invalid_param" {
		t.Errorf("code = %q", difyErr.Code)
	}
}

func TestDo_Timeout(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(500 * time.Millisecond)
	}))
	defer srv.Close()

	c := NewHTTPClient(srv.URL, "test-key", 50*time.Millisecond, DefaultRetryConfig(3))
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
	defer cancel()

	err := c.Do(ctx, "GET", "/info", nil, nil)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}

func TestDo_Retry(t *testing.T) {
	var mu sync.Mutex
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		attempts++
		mu.Unlock()
		w.WriteHeader(503)
	}))
	defer srv.Close()

	cfg := DefaultRetryConfig(3)
	cfg.BaseDelay = 1 * time.Millisecond

	c := NewHTTPClient(srv.URL, "test-key", 10*time.Second, cfg)
	err := c.Do(context.Background(), "GET", "/info", nil, nil)
	if err == nil {
		t.Fatal("expected error after retries")
	}
	if attempts != cfg.MaxRetries+1 {
		t.Errorf("attempts = %d, want %d", attempts, cfg.MaxRetries+1)
	}
}

func TestSSEStream_Parsing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher, ok := w.(http.Flusher)
		if !ok {
			t.Fatal("expected http.Flusher")
		}
		events := []string{
			`data: {"event":"message","message_id":"m1","answer":"Hello"}` + "\n\n",
			`data: {"event":"message","message_id":"m1","answer":" World"}` + "\n\n",
			`data: {"event":"message_end","message_id":"m1","metadata":{"usage":{"total_tokens":10}}}` + "\n\n",
		}
		for _, e := range events {
			w.Write([]byte(e))
			flusher.Flush()
		}
	}))
	defer srv.Close()

	c := NewHTTPClient(srv.URL, "test-key", 10*time.Second, DefaultRetryConfig(3))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, errs := c.Stream(ctx, "POST", "/chat-messages", map[string]string{"query": "hi", "response_mode": "streaming"})

	var received []SseEvent
loop:
	for {
		select {
		case ev, ok := <-events:
			if !ok {
				break loop
			}
			received = append(received, ev)
		case err := <-errs:
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		}
	}

	if len(received) != 3 {
		t.Fatalf("got %d events, want 3", len(received))
	}
	if received[0].Event != "message" || received[0].Answer != "Hello" {
		t.Errorf("event[0] = %+v", received[0])
	}
	if received[2].Event != "message_end" {
		t.Errorf("event[2] = %+v", received[2])
	}
}

func TestSSEStream_SkipsPing(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		flusher := w.(http.Flusher)
		events := []string{
			`data: {"event":"ping"}` + "\n\n",
			`data: {"event":"message","answer":"Hi"}` + "\n\n",
		}
		for _, e := range events {
			w.Write([]byte(e))
			flusher.Flush()
		}
	}))
	defer srv.Close()

	c := NewHTTPClient(srv.URL, "test-key", 10*time.Second, DefaultRetryConfig(3))
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	events, _ := c.Stream(ctx, "POST", "/test", nil)

	var received []SseEvent
	for ev := range events {
		received = append(received, ev)
	}

	if len(received) != 1 {
		t.Fatalf("got %d events, want 1 (ping skipped)", len(received))
	}
}

func TestDifyError_Parse(t *testing.T) {
	body := []byte(`{"code":"not_found","message":"Conversation Not Exists.","status":404}`)
	e := NewDifyError(404, body)
	if e.Code != "not_found" {
		t.Errorf("code = %q", e.Code)
	}
	if e.StatusCode != 404 {
		t.Errorf("statusCode = %d", e.StatusCode)
	}
}

func TestMaskKey(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"app-xxxxxxxxxxxxxxxxxxxx", "app-xxxx***"},
		{"short", "***"},
		{"", "***"},
	}
	for _, tt := range tests {
		got := MaskKey(tt.key)
		if got != tt.want {
			t.Errorf("MaskKey(%q) = %q, want %q", tt.key, got, tt.want)
		}
	}
}

func TestRetryableStatusCodes(t *testing.T) {
	want := map[int]bool{429: true, 502: true, 503: true, 504: true}
	for _, s := range RetryableStatusCodes {
		if !want[s] {
			t.Errorf("unexpected retryable code: %d", s)
		}
	}
}

