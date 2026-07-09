## Context

当前 `server` 模块是空白 Fiber v3 服务端，仅有 `/health` 和 `/api/v1` 空路由。Dify 工作流中的 HTTP 请求组件可以向外部服务发送请求，但 Go 后端尚无通用端点接收此类调用。

需要在不引入 RPC 框架（gRPC、Twirp 等）的前提下，提供一个轻量级的方法调用机制，让 Dify 通过标准 HTTP JSON 请求即可调用 Go 后端的任意注册方法。

## Goals / Non-Goals

**Goals:**
- 提供 `POST /api/v1/methods` 端点，接受 `{"method": "xxx", "params": {...}}` JSON body
- 提供 `GET /api/v1/methods` 发现端点，返回所有已注册方法的元数据
- 方法注册表支持按名称注册/查找，带有参数/返回值 JSON Schema 描述
- 通过 `dig.Container` 注入 MethodRegistry
- 统一响应格式，方便 Dify HTTP 组件解析输出变量（body/status_code/headers）

**Non-Goals:**
- 不做 gRPC / Protobuf 代码生成
- 不做方法版本管理
- 不做请求鉴权（由 Dify 侧 API Key 或内部网络控制）
- 不做异步方法调用 / 回调
- 不做请求日志持久化

## Decisions

### 1. Handler 签名：`func(ctx, params) (result, error)`

```go
type MethodHandler func(ctx context.Context, params json.RawMessage) (interface{}, error)
```

- params 使用 `json.RawMessage`，由各方法自行解析为具体结构体
- result 返回 `interface{}`，序列化为 JSON 后放入响应
- error 返回时自动映射为失败响应

**理由**：Go 泛型无法在 map 中存储不同签名的函数。`json.RawMessage` + `interface{}` 是最小公约数，每个方法在自己的实现里做类型断言/解析即可。

**备选**：`func(ctx, params map[string]interface{})` — 拒绝，要求所有方法都做 map → struct 的手动转换，易出错。

### 2. 响应信封：统一 success/error 结构

```json
// 成功
{"success": true, "data": <任意 JSON 值>}

// 失败
{"success": false, "error": {"code": "METHOD_NOT_FOUND", "message": "方法 xxx 未注册"}}
```

Dify HTTP 组件输出变量 `body` (string) 可拿到完整 JSON 字符串，`status_code` (number) 可拿到 HTTP 状态码。成功返回 200，方法未找到返回 404，方法执行失败返回 500。

**理由**：`success` 字段让 Dify 工作流可以用条件节点判断调用是否成功，而不依赖 HTTP 状态码。`data` 可以是任意 JSON 值（对象、数组、字符串、数字），灵活适配各种返回值。

### 3. MethodRegistry：map + RWMutex + JSON Schema 元数据

```go
type MethodDef struct {
    Name        string          `json:"name"`
    Description string          `json:"description"`
    InputSchema json.RawMessage `json:"input_schema"`  // JSON Schema
    OutputSchema json.RawMessage `json:"output_schema"` // JSON Schema
    Handler     MethodHandler   `json:"-"`
}

type MethodRegistry struct {
    mu      sync.RWMutex
    methods map[string]*MethodDef
}
```

- 启动时在 `main.go` 中调用 `registry.Register(name, def)` 注册方法
- POST 请求时 `registry.Get(name)` 查找 handler
- GET 请求时 `registry.List()` 返回所有方法的元数据（不含 handler）

**理由**：简单的 map + 锁足够，不需要引入第三方服务注册库。JSON Schema 描述参数和返回值，让 Dify 侧调用方能知道该传什么、会收到什么。

### 4. 超时控制：全局配置 + 请求级 context

```go
// .env
METHOD_TIMEOUT=30s

// handler 中
ctx, cancel := context.WithTimeout(c.Context(), cfg.MethodTimeout)
defer cancel()
result, err := methodDef.Handler(ctx, params)
```

每个方法调用都有独立的 context 超时，防止单个方法阻塞整个请求。超时值从 `.env` 读取，默认 30s。

**理由**：请求级 context.WithTimeout 比 http.Client 级 Timeout 更精确，且方法内部可以检查 ctx.Done() 来提前退出。

### 5. 错误码设计：简短语义码

| 码 | HTTP 状态 | 含义 |
|---|-----------|------|
| `METHOD_NOT_FOUND` | 404 | `method` 字段指定的方法未注册 |
| `INVALID_PARAMS` | 400 | 请求 JSON 格式错误或缺少 `method` 字段 |
| `METHOD_ERROR` | 500 | 方法执行中返回的错误 |
| `METHOD_TIMEOUT` | 504 | 方法执行超时 |

**理由**：简短字符串码便于 Dify 工作流做条件路由，比数字码更具可读性。

### 6. 模块结构

```
server/
├── main.go
├── methods/
│   ├── registry.go    // MethodDef + MethodRegistry
│   ├── handler.go     // Fiber handler: POST + GET /methods
│   └── registry_test.go
└── di/
    └── container.go   // 注册 MethodRegistry Provider
```

放在 `server/methods/` 而非 `dify-sdk/` 下，因为方法注册是服务端关注点，不是 SDK 库的职责。SDK 专注 Dify API 调用，server 专注被 Dify 调用的能力。

**理由**：关注点分离。`dify-sdk` 是"调用 Dify 的库"，`server/methods` 是"被 Dify 调用的端点"。

## Risks / Trade-offs

- **[R] 方法执行 panic 导致服务崩溃** → Fiber recover 中间件已就位，panic → 500 + JSON 错误响应
- **[R] 恶意请求传入超大 params 导致 OOM** → 限制 Body size（Fiber 默认 4MB），方法内部做参数校验
- **[R] GET /methods 暴露内部方法签名** → 可接受，内网部署；如需保护可后续加 API Key 校验
- **[T] json.RawMessage 缺少编译期类型安全** → 每个方法在入口处做 JSON Schema 校验 + 解析，写单测覆盖参数边界
