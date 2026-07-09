package di

import (
	"testing"
	"time"

	"github.com/wgl/dify-sdk/client"
	"github.com/wgl/dify-sdk/config"
)

func TestBuildContainer(t *testing.T) {
	cfg := &config.Config{
		BaseURL:    "http://localhost:5001/v1",
		APIKeys:    []string{"app-test"},
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}

	container, err := BuildContainer(cfg)
	if err != nil {
		t.Fatalf("BuildContainer error: %v", err)
	}

	if err := container.Invoke(func(c *config.Config) {
		if c.BaseURL != "http://localhost:5001/v1" {
			t.Errorf("config not injected")
		}
	}); err != nil {
		t.Fatalf("Invoke config error: %v", err)
	}

	if err := container.Invoke(func(h *client.HTTPClient) {
		if h == nil {
			t.Fatal("HTTPClient is nil")
		}
	}); err != nil {
		t.Fatalf("Invoke http client error: %v", err)
	}
}

func TestBuildContainerWithKey(t *testing.T) {
	cfg := &config.Config{
		BaseURL:    "http://localhost:5001/v1",
		APIKeys:    []string{"app-test"},
		Timeout:    30 * time.Second,
		MaxRetries: 3,
	}

	container, err := BuildContainerWithKey(cfg, "custom-key")
	if err != nil {
		t.Fatalf("BuildContainerWithKey error: %v", err)
	}

	if err := container.Invoke(func(h *client.HTTPClient) {
		if h == nil {
			t.Fatal("HTTPClient is nil")
		}
	}); err != nil {
		t.Fatalf("Invoke http client error: %v", err)
	}
}
