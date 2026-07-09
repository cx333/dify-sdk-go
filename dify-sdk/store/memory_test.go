package store

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/wgl/dify-sdk/client"
)

func TestMemoryStore_Basic(t *testing.T) {
	s := NewMemoryStore()

	s.Upsert(&AppMeta{ID: "1", Name: "Chat1", Mode: "chat"})
	s.Upsert(&AppMeta{ID: "2", Name: "Wf1", Mode: "workflow"})

	app, ok := s.GetByID("1")
	if !ok || app.Name != "Chat1" {
		t.Errorf("GetByID = %+v, %v", app, ok)
	}

	chatApps := s.GetByMode("chat")
	if len(chatApps) != 1 || chatApps[0].ID != "1" {
		t.Errorf("GetByMode = %+v", chatApps)
	}

	all := s.All()
	if len(all) != 2 {
		t.Errorf("All() len = %d", len(all))
	}
}

func TestMemoryStore_GetByKey(t *testing.T) {
	s := NewMemoryStore()
	s.Upsert(&AppMeta{ID: "1", Name: "A", APIKey: "key-a"})
	s.Upsert(&AppMeta{ID: "2", Name: "B", APIKey: "key-a"})
	s.Upsert(&AppMeta{ID: "3", Name: "C", APIKey: "key-b"})

	a := s.GetByKey("key-a")
	if len(a) != 2 {
		t.Errorf("GetByKey(key-a) len = %d, want 2", len(a))
	}
	b := s.GetByKey("key-b")
	if len(b) != 1 {
		t.Errorf("GetByKey(key-b) len = %d, want 1", len(b))
	}
}

func TestMemoryStore_Concurrent(t *testing.T) {
	s := NewMemoryStore()
	var wg sync.WaitGroup

	for i := 0; i < 100; i++ {
		wg.Add(2)
		go func(n int) {
			defer wg.Done()
			s.Upsert(&AppMeta{ID: string(rune('0' + n%10)), Name: "x", Mode: "chat"})
		}(i)
		go func() {
			defer wg.Done()
			_ = s.All()
			_ = s.GetByMode("chat")
		}()
	}
	wg.Wait()
}

func TestMemoryStore_Preload(t *testing.T) {
	s := NewMemoryStore()

	err := s.Preload(context.Background(), []string{"key1", "key2"},
		func(ctx context.Context, hc *client.HTTPClient) ([]AppMeta, error) {
			return []AppMeta{
				{ID: "app-a", Name: "AppA", Mode: "chat"},
				{ID: "app-b", Name: "AppB", Mode: "workflow"},
			}, nil
		},
		"http://localhost", 10*time.Second, 2)

	if err != nil {
		t.Fatalf("Preload error: %v", err)
	}
	all := s.All()
	// Same app IDs are deduplicated across keys
	if len(all) != 2 {
		t.Errorf("Preload result len = %d, want 2 (deduplicated by ID)", len(all))
	}
	byKey1 := s.GetByKey("key1")
	if len(byKey1) != 2 {
		t.Errorf("key1 apps = %d, want 2", len(byKey1))
	}
	byKey2 := s.GetByKey("key2")
	if len(byKey2) != 2 {
		t.Errorf("key2 apps = %d, want 2", len(byKey2))
	}
}
