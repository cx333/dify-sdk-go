// Package methods 提供通用方法注册与动态调用能力。
// Dify HTTP 请求组件通过传入方法名即可调用 Go 后端任意已注册方法。
//
// 使用示例：
//
//	registry := methods.NewMethodRegistry()
//	registry.Register("greet", &methods.MethodDef{
//	    Name:        "greet",
//	    Description: "返回问候语",
//	    Handler: func(ctx context.Context, params json.RawMessage) (interface{}, error) {
//	        return map[string]string{"msg": "hello"}, nil
//	    },
//	})
package methods
