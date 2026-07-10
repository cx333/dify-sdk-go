// Package methods 提供通用方法注册与动态调用能力。
// Dify HTTP 请求组件通过传入方法名即可调用 Go 后端任意已注册方法。
package methods

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
)

// MethodHandler 已注册方法的函数签名。
// params 为原始 JSON，由各方法自行解析为具体结构体。
// result 为任意可 JSON 序列化的值，放入统一响应的 data 字段。
type MethodHandler func(ctx context.Context, params json.RawMessage) (any, error)

// MethodDef 描述一个已注册方法。
type MethodDef struct {
	Name         string          `json:"name"`
	Description  string          `json:"description"`
	InputSchema  json.RawMessage `json:"input_schema"`
	OutputSchema json.RawMessage `json:"output_schema"`
	Handler      MethodHandler   `json:"-"`
}

// MethodRegistry 线程安全的方法注册表。
type MethodRegistry struct {
	mu      sync.RWMutex
	methods map[string]*MethodDef
}

// NewMethodRegistry 创建空注册表。
func NewMethodRegistry() *MethodRegistry {
	return &MethodRegistry{
		methods: make(map[string]*MethodDef),
	}
}

// Register 注册方法。name 重复时返回 error。
func (r *MethodRegistry) Register(name string, def *MethodDef) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.methods[name]; exists {
		return fmt.Errorf("方法 %s 已注册", name)
	}
	r.methods[name] = def
	return nil
}

// Get 按名称查找方法。
func (r *MethodRegistry) Get(name string) (*MethodDef, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	def, ok := r.methods[name]
	return def, ok
}

// List 返回所有已注册方法的元数据（不含 Handler）。
func (r *MethodRegistry) List() []*MethodDef {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]*MethodDef, 0, len(r.methods))
	for _, def := range r.methods {
		result = append(result, &MethodDef{
			Name:         def.Name,
			Description:  def.Description,
			InputSchema:  def.InputSchema,
			OutputSchema: def.OutputSchema,
		})
	}
	return result
}
