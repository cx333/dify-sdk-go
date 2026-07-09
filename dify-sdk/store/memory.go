// Package store 提供 Dify 应用元数据的内存存储与查询能力。
package store

import (
	"context"
	"sync"
	"time"

	"github.com/wgl/dify-sdk/client"
)

// AppMeta Dify 应用的元数据摘要。
type AppMeta struct {
	ID          string    // 应用 ID
	Name        string    // 应用名称
	Mode        string    // 应用模式：chat / agent-chat / workflow / completion / advanced-chat
	Description string    // 应用描述
	APIKey      string    // 关联的 API Key
	Tags        []string  // 标签
	UpdatedAt   time.Time // 最后更新时间
}

// MetadataFetcher 元数据拉取函数签名。
// 传入 HTTPClient（已配置对应 API Key），返回该 Key 下的所有应用元数据。
type MetadataFetcher func(ctx context.Context, httpClient *client.HTTPClient) ([]AppMeta, error)

// MemoryStore 并发安全的内存元数据存储。
//
// 支持三层索引：
//   - byID: 按应用 ID 精确查询
//   - byMode: 按应用模式（类型）列表查询
//   - byKey: 按 API Key 查询归属
type MemoryStore struct {
	mu     sync.RWMutex
	apps   map[string]*AppMeta   // key: app_id
	byMode map[string][]*AppMeta // key: mode (chat/workflow/...)
	byKey  map[string][]*AppMeta // key: api_key
}

// NewMemoryStore 创建空的 MemoryStore。
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		apps:   make(map[string]*AppMeta),
		byMode: make(map[string][]*AppMeta),
		byKey:  make(map[string][]*AppMeta),
	}
}

// Upsert 插入或更新应用元数据（相同 ID 覆盖）。
func (s *MemoryStore) Upsert(app *AppMeta) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.apps[app.ID] = app
	s.byMode[app.Mode] = append(s.byMode[app.Mode], app)
	s.byKey[app.APIKey] = append(s.byKey[app.APIKey], app)
}

// GetByID 按应用 ID 查询。
func (s *MemoryStore) GetByID(id string) (*AppMeta, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	app, ok := s.apps[id]
	return app, ok
}

// GetByMode 按应用模式查询所有匹配的应用。
func (s *MemoryStore) GetByMode(mode string) []*AppMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	apps := s.byMode[mode]
	result := make([]*AppMeta, len(apps))
	copy(result, apps)
	return result
}

// GetByKey 按 API Key 查询归属的应用。
func (s *MemoryStore) GetByKey(key string) []*AppMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	apps := s.byKey[key]
	result := make([]*AppMeta, len(apps))
	copy(result, apps)
	return result
}

// All 返回所有已存储的应用元数据。
func (s *MemoryStore) All() []*AppMeta {
	s.mu.RLock()
	defer s.mu.RUnlock()
	result := make([]*AppMeta, 0, len(s.apps))
	for _, app := range s.apps {
		result = append(result, app)
	}
	return result
}

// Preload 并发拉取所有 API Key 对应的元数据并写入存储。
//
// 参数：
//   - keys: 要遍历的 API Key 列表
//   - fetcher: 元数据拉取函数
//   - baseURL: Dify API 基础地址
//   - timeout: 每个请求的超时时间
//   - concurrency: 最大并发数（≤0 则默认 5）
//
// 使用 semaphore 限制并发，避免请求风暴。
func (s *MemoryStore) Preload(ctx context.Context, keys []string, fetcher MetadataFetcher, baseURL string, timeout time.Duration, concurrency int) error {
	if concurrency <= 0 {
		concurrency = 5
	}

	sem := make(chan struct{}, concurrency)
	var wg sync.WaitGroup
	var firstErr error
	var errMu sync.Mutex

	for _, key := range keys {
		wg.Add(1)
		go func(k string) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()

			httpClient := client.NewHTTPClient(baseURL, k, timeout, client.DefaultRetryConfig())
			apps, err := fetcher(ctx, httpClient)
			if err != nil {
				errMu.Lock()
				if firstErr == nil {
					firstErr = err
				}
				errMu.Unlock()
				return
			}

			for _, app := range apps {
				app.APIKey = k
				s.Upsert(&app)
			}
		}(key)
	}

	wg.Wait()
	return firstErr
}
