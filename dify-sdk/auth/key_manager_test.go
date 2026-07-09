package auth

import (
	"testing"
)

func TestKeyManager_RoundRobin(t *testing.T) {
	km := NewKeyManager([]string{"key1", "key2", "key3"})

	order := make([]string, 6)
	for i := range order {
		order[i] = km.Next()
	}

	want := []string{"key1", "key2", "key3", "key1", "key2", "key3"}
	for i := range want {
		if order[i] != want[i] {
			t.Fatalf("at %d: got %q, want %q; order=%v", i, order[i], want[i], order)
		}
	}
}

func TestKeyManager_MarkFailed(t *testing.T) {
	km := NewKeyManager([]string{"key1", "key2", "key3"})
	km.MarkFailed("key2")

	// Should skip key2
	for i := 0; i < 3; i++ {
		k := km.Next()
		if k == "key2" {
			t.Errorf("got failed key at round %d", i)
		}
	}
}

func TestKeyManager_UnmarkFailed(t *testing.T) {
	km := NewKeyManager([]string{"key1", "key2"})
	km.MarkFailed("key2")
	km.UnmarkFailed("key2")

	foundKey2 := false
	for i := 0; i < 4; i++ {
		if km.Next() == "key2" {
			foundKey2 = true
		}
	}
	if !foundKey2 {
		t.Error("key2 should be available after unmark")
	}
}

func TestKeyManager_Empty(t *testing.T) {
	km := NewKeyManager(nil)
	if km.Next() != "" {
		t.Error("empty manager should return empty string")
	}
}

func TestKeyManager_AllFailed(t *testing.T) {
	km := NewKeyManager([]string{"key1", "key2"})
	km.MarkFailed("key1")
	km.MarkFailed("key2")
	if km.Next() != "" {
		t.Error("all-failed should return empty string")
	}
}

func TestKeyManager_All(t *testing.T) {
	km := NewKeyManager([]string{"a", "b"})
	all := km.All()
	if len(all) != 2 || all[0] != "a" || all[1] != "b" {
		t.Errorf("All() = %v", all)
	}
}
