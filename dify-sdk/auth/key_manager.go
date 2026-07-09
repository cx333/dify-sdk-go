// Package auth 提供 Dify API Key 的轮询管理与失效标记能力。
package auth

import (
	"sync"
	"sync/atomic"
)

// KeyManager 管理多个 Dify API Key，支持 round-robin 轮询分配和失效标记。
//
// 并发安全：所有方法可被多个 goroutine 同时调用。
type KeyManager struct {
	keys    []string
	current atomic.Int32    // 原子计数器，用于轮询
	failed  map[int]struct{} // 失效 Key 的索引集合
	mu      sync.RWMutex     // 保护 failed 和 keys 的读写
}

// NewKeyManager 创建 KeyManager。
// keys 从环境变量 DIFY_API_KEYS（逗号分隔）解析获得。
func NewKeyManager(keys []string) *KeyManager {
	return &KeyManager{
		keys:   keys,
		failed: make(map[int]struct{}),
	}
}

// Next 以 round-robin 方式返回下一个可用的 API Key。
// 自动跳过已标记为失效的 Key。所有 Key 均失效时返回空字符串。
func (km *KeyManager) Next() string {
	km.mu.RLock()
	total := len(km.keys)
	failedCount := len(km.failed)
	km.mu.RUnlock()

	if total == 0 || failedCount >= total {
		return ""
	}

	for range total {
		idx := int(km.current.Add(1)-1) % total
		if idx < 0 {
			idx += total
		}

		km.mu.RLock()
		_, isFailed := km.failed[idx]
		km.mu.RUnlock()

		if !isFailed {
			km.mu.RLock()
			key := km.keys[idx]
			km.mu.RUnlock()
			return key
		}
	}

	return ""
}

// All 返回所有 API Key 的副本（含失效 Key）。
func (km *KeyManager) All() []string {
	km.mu.RLock()
	defer km.mu.RUnlock()
	result := make([]string, len(km.keys))
	copy(result, km.keys)
	return result
}

// MarkFailed 标记指定 API Key 为失效，轮询时自动跳过。
func (km *KeyManager) MarkFailed(key string) {
	km.mu.Lock()
	defer km.mu.Unlock()
	for i, k := range km.keys {
		if k == key {
			km.failed[i] = struct{}{}
			return
		}
	}
}

// UnmarkFailed 移除失效标记，恢复 Key 的轮询可用性。
func (km *KeyManager) UnmarkFailed(key string) {
	km.mu.Lock()
	defer km.mu.Unlock()
	for i, k := range km.keys {
		if k == key {
			delete(km.failed, i)
			return
		}
	}
}

// Len 返回 Key 总数。
func (km *KeyManager) Len() int {
	return len(km.keys)
}
