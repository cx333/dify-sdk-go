package methods

import (
	"context"
	"encoding/json"
	"sync"
	"testing"
)

func TestRegisterAndGet(t *testing.T) {
	r := NewMethodRegistry()

	def := &MethodDef{
		Name:        "greet",
		Description: "返回问候语",
		Handler: func(ctx context.Context, params json.RawMessage) (interface{}, error) {
			return map[string]string{"msg": "hello"}, nil
		},
	}

	err := r.Register("greet", def)
	if err != nil {
		t.Fatalf("Register 失败: %v", err)
	}

	got, ok := r.Get("greet")
	if !ok {
		t.Fatal("Get 未找到已注册方法")
	}
	if got.Name != "greet" {
		t.Errorf("Name = %s, want greet", got.Name)
	}
	if got.Description != "返回问候语" {
		t.Errorf("Description = %s, want 返回问候语", got.Description)
	}
	if got.Handler == nil {
		t.Error("Handler 不应为 nil")
	}
}

func TestGetNotFound(t *testing.T) {
	r := NewMethodRegistry()

	_, ok := r.Get("unknown")
	if ok {
		t.Error("Get 未注册方法应返回 false")
	}
}

func TestRegisterDuplicate(t *testing.T) {
	r := NewMethodRegistry()

	def1 := &MethodDef{Name: "greet", Handler: func(ctx context.Context, params json.RawMessage) (interface{}, error) { return nil, nil }}
	def2 := &MethodDef{Name: "greet", Handler: func(ctx context.Context, params json.RawMessage) (interface{}, error) { return nil, nil }}

	if err := r.Register("greet", def1); err != nil {
		t.Fatalf("首次 Register 失败: %v", err)
	}

	err := r.Register("greet", def2)
	if err == nil {
		t.Fatal("重复注册应返回 error")
	}

	got, _ := r.Get("greet")
	if got != def1 {
		t.Error("重复注册应保留第一个定义")
	}
}

func TestList(t *testing.T) {
	r := NewMethodRegistry()

	r.Register("a", &MethodDef{Name: "a", Handler: func(ctx context.Context, params json.RawMessage) (interface{}, error) { return nil, nil }})
	r.Register("b", &MethodDef{Name: "b", Handler: func(ctx context.Context, params json.RawMessage) (interface{}, error) { return nil, nil }})

	list := r.List()
	if len(list) != 2 {
		t.Fatalf("List 长度 = %d, want 2", len(list))
	}

	for _, def := range list {
		if def.Handler != nil {
			t.Errorf("List 返回的 %s Handler 应为 nil", def.Name)
		}
	}
}

func TestListEmpty(t *testing.T) {
	r := NewMethodRegistry()

	list := r.List()
	if list == nil {
		t.Error("空注册表 List 应返回空切片而非 nil")
	}
	if len(list) != 0 {
		t.Errorf("空注册表 List 长度 = %d, want 0", len(list))
	}
}

func TestConcurrentAccess(t *testing.T) {
	r := NewMethodRegistry()
	var wg sync.WaitGroup

	// 并发注册
	for i := 0; i < 50; i++ {
		i := i
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Register(string(rune('A'+i%26))+string(rune('0'+i/26)), &MethodDef{
				Name:    "test",
				Handler: func(ctx context.Context, params json.RawMessage) (interface{}, error) { return nil, nil },
			})
		}()
	}

	// 并发读取
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.Get("A0")
			r.List()
		}()
	}

	wg.Wait()
}
