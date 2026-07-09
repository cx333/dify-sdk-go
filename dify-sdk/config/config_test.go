package config

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func clearEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"DIFY_BASE_URL", "DIFY_API_KEYS", "DIFY_TIMEOUT", "DIFY_MAX_RETRIES"} {
		os.Unsetenv(k)
	}
}

func TestLoad(t *testing.T) {
	clearEnv(t)
	tmp := t.TempDir()
	p := filepath.Join(tmp, ".env")
	writeFile(t, p, `DIFY_BASE_URL=http://localhost:5001/v1
DIFY_API_KEYS=app-1, app-2, app-3
DIFY_TIMEOUT=10s
DIFY_MAX_RETRIES=5
`)

	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}

	if cfg.BaseURL != "http://localhost:5001/v1" {
		t.Errorf("BaseURL = %q, want %q", cfg.BaseURL, "http://localhost:5001/v1")
	}
	if len(cfg.APIKeys) != 3 {
		t.Errorf("got %d keys, want 3", len(cfg.APIKeys))
	}
	if cfg.APIKeys[0] != "app-1" || cfg.APIKeys[1] != "app-2" || cfg.APIKeys[2] != "app-3" {
		t.Errorf("APIKeys = %v", cfg.APIKeys)
	}
	if cfg.Timeout != 10*time.Second {
		t.Errorf("Timeout = %v, want 10s", cfg.Timeout)
	}
	if cfg.MaxRetries != 5 {
		t.Errorf("MaxRetries = %d, want 5", cfg.MaxRetries)
	}
}

func TestLoadDefaults(t *testing.T) {
	clearEnv(t)
	tmp := t.TempDir()
	p := filepath.Join(tmp, ".env")
	writeFile(t, p, `DIFY_BASE_URL=http://localhost
DIFY_API_KEYS=app-1
`)

	cfg, err := Load(p)
	if err != nil {
		t.Fatalf("Load() error = %v", err)
	}
	if cfg.Timeout != 30*time.Second {
		t.Errorf("default Timeout = %v", cfg.Timeout)
	}
	if cfg.MaxRetries != 3 {
		t.Errorf("default MaxRetries = %d", cfg.MaxRetries)
	}
}

func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/.env")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

func TestLoadMissingBaseURL(t *testing.T) {
	clearEnv(t)
	tmp := t.TempDir()
	p := filepath.Join(tmp, ".env")
	writeFile(t, p, "DIFY_API_KEYS=app-1\n")
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected error for missing BaseURL")
	}
}

func TestLoadMissingAPIKeys(t *testing.T) {
	clearEnv(t)
	tmp := t.TempDir()
	p := filepath.Join(tmp, ".env")
	writeFile(t, p, "DIFY_BASE_URL=http://localhost\n")
	_, err := Load(p)
	if err == nil {
		t.Fatal("expected error for missing APIKeys")
	}
}

func TestSplitKeysTrimsWhitespace(t *testing.T) {
	keys := splitKeys(" key1 ,  key2  ,key3 ")
	if len(keys) != 3 || keys[0] != "key1" || keys[1] != "key2" || keys[2] != "key3" {
		t.Errorf("splitKeys = %v", keys)
	}
}

func TestSplitKeysSkipsEmpty(t *testing.T) {
	keys := splitKeys("key1,,key2,  ,key3")
	if len(keys) != 3 {
		t.Errorf("got %d keys, want 3", len(keys))
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
}
