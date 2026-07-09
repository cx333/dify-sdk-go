## 背景

本项目从零构建，Go module 已初始化（`dify-api`, Go 1.26.3）。无历史代码，可自由按最佳实践设计。核心约束：
- Go Workspace 三模块：`dify-sdk`（库）、`examples`（示例服务）、`server`（空白服务端）
- 所有模块只用 `.env`，禁止 config.yaml
- 所有模块使用 `dig.Container` 做依赖注入
- 服务端使用 Fiber v3
- SDK 需完整覆盖 Dify 的 Chat、Workflow、Knowledge、文件、模型配置 API

## 目标 / 非目标

**目标：**
- 高性能可复用 HTTP 客户端（连接池、超时、重试、SSE 流式）
- Go Workspace 三模块清晰分离：SDK 库、示例服务、空白服务端
- 完整封装 Dify Chat、Workflow、Knowledge、文件、模型配置 API
- 统一 `.env` 配置，所有模块一致
- `dig.Container` 全链路依赖注入
- Fiber v3 构建 RESTful HTTP 服务

**非目标：**
- 不实现 Dify 的 Admin API（用户管理、权限等）
- 不做持久化数据库（内存存储即可）
- 不做 WebSocket 实时通信（SSE 已覆盖流式场景）
- 不做多租户隔离
- 不做 SDK 的公开发布（go get 可用即可，不追求 semver 严格管理）

## 决策

### 1. Go Workspace 三模块结构

```
.
├── go.work                    # Go Workspace 根
├── dify-sdk/
│   ├── go.mod                 # module github.com/wgl/dify-sdk
│   ├── client/
│   │   ├── http.go            # HTTP Client（连接池、重试、超时）
│   │   ├── chat.go            # Chat API
│   │   ├── workflow.go        # Workflow API
│   │   ├── knowledge.go       # Knowledge API
│   │   ├── file.go            # File API
│   │   └── model.go           # Model Config API
│   ├── config/
│   │   └── env.go             # .env 加载 + Config 结构体
│   ├── store/
│   │   └── memory.go          # 内存元数据存储
│   └── di/
│       └── container.go       # dig.Container 初始化与 Provider 注册
├── examples/
│   ├── go.mod                 # module github.com/wgl/dify-api/examples
│   ├── .env                   # 示例服务配置
│   ├── main.go                # 入口：组装 DI + 启动 Fiber
│   └── handler/
│       └── demo.go            # 演示 Handler（调用 SDK）
└── server/
    ├── go.mod                 # module github.com/wgl/dify-api/server
    ├── .env                   # 空白服务端配置
    ├── main.go                # 入口：组装 DI + 启动 Fiber（空路由）
    └── handler/
        └── placeholder.go     # 占位 Handler，等待业务填充
```

**理由**：
- `dify-sdk` 作为独立 module，可被外部项目 `go get` 引用
- `examples` 和 `server` 依赖 `dify-sdk`，形成清晰的依赖层次
- `go.work` 让开发时跨模块编辑和测试无缝

**备选方案**：
- 单 module 多包 — 拒绝，SDK 无法被外部独立引用
- 三独立 repo — 拒绝，过度拆分，开发同步成本高

### 2. dig.Container 依赖注入架构

```go
// dify-sdk/di/container.go
func BuildContainer(cfg *config.Config) (*dig.Container, error) {
    c := dig.New()
    
    // 提供 Config
    if err := c.Provide(func() *config.Config { return cfg }); err != nil {
        return nil, err
    }
    
    // 提供 HTTP Client
    if err := c.Provide(client.NewHTTPClient); err != nil {
        return nil, err
    }
    
    // 提供各 API Client
    if err := c.Provide(client.NewChatClient); err != nil {
        return nil, err
    }
    if err := c.Provide(client.NewWorkflowClient); err != nil {
        return nil, err
    }
    // ... 其他 Client
    
    // 提供 Store
    if err := c.Provide(store.NewMemoryStore); err != nil {
        return nil, err
    }
    
    return c, nil
}
```

使用 `dig.Invoke` 在 `main.go` 中解析依赖：
```go
// examples/main.go
container, err := di.BuildContainer(cfg)
if err != nil { log.Fatal(err) }

err = container.Invoke(func(chat *client.ChatClient, wf *client.WorkflowClient, app *fiber.App) {
    // 注册路由
    app.Post("/chat", handler.Chat(chat))
    app.Post("/workflow", handler.Workflow(wf))
})
```

**理由**：
- `dig` 是 Uber 出品，与 Go 生态无缝集成，零反射开销（编译期生成）
- 构造函数即 Provider，无需额外注解，Go 原生体验
- 支持生命周期管理（`Shutdowner` 接口）

**备选方案**：
- Wire（Google）— 拒绝，需要代码生成步骤，增加构建复杂度
- 手写工厂 — 拒绝，模块增多后工厂函数爆炸
- fx（Uber）— 拒绝，fx 是更高层框架，引入过多概念，dig 更轻量

### 3. HTTP Client：连接池 + 重试 + 超时 + SSE

```go
type HTTPClient struct {
    client  *http.Client
    baseURL string
    retry   RetryConfig
}

type RetryConfig struct {
    MaxRetries  int
    BaseDelay   time.Duration
    MaxDelay    time.Duration
    StatusCodes []int // 需要重试的状态码，如 429, 502, 503, 504
}
```

- **连接池**：复用 `http.Client`（已内置连接池），设置 `MaxIdleConns: 100`、`MaxIdleConnsPerHost: 10`
- **超时**：请求级 `context.WithTimeout`，默认 30s，可配置
- **重试**：指数退避重试，仅对幂等方法（GET）和网络错误自动重试；POST 仅在 429/5xx 时重试（需确认 Dify API 幂等性）
- **SSE**：`Accept: text/event-stream`，使用 `bufio.Scanner` 逐行读取，通过 channel 返回 `Event` 结构体

**理由**：
- `http.Client` 是 Go 标准库，连接池已优化，无需第三方 HTTP 库
- 请求级超时通过 context 控制，比 Client 级 `Timeout` 更灵活
- 指数退避避免请求风暴

**备选方案**：
- `resty` — 拒绝，功能丰富但引入不必要的依赖，标准库已足够
- `fasthttp` — 拒绝，与 `net/http` 不兼容，且 Dify API 非高并发场景

### 4. .env 配置：统一结构化 Config

```go
// dify-sdk/config/env.go
type Config struct {
    BaseURL    string        `env:"DIFY_BASE_URL"`
    APIKeys    []string      `env:"DIFY_API_KEYS"`    // 逗号分隔
    Timeout    time.Duration `env:"DIFY_TIMEOUT" default:"30s"`
    MaxRetries int           `env:"DIFY_MAX_RETRIES" default:"3"`
}

func Load(path string) (*Config, error) {
    if err := godotenv.Load(path); err != nil {
        return nil, fmt.Errorf("load .env: %w", err)
    }
    // 解析环境变量到结构体
}
```

- 每个 module 独立 `.env` 文件
- `examples/.env` 和 `server/.env` 各自维护
- 禁止 config.yaml，强制 `.env`

**理由**：
- `godotenv` 是 Go 生态标准，简单可靠
- `.env` 便于 Docker/K8s 注入，12-Factor 推荐
- 结构化 Config 比零散 `os.Getenv` 更易维护

**备选方案**：
- `viper` — 拒绝，过度设计，且支持 yaml 违背约束
- `envconfig` — 可考虑，但 `godotenv` + 手写解析更可控

### 5. Fiber v3 服务端架构

```go
// examples/main.go
func main() {
    cfg := config.Load(".env")
    container := di.BuildContainer(cfg)
    
    app := fiber.New(fiber.Config{
        Prefork:      false,
        ServerHeader: "dify-examples",
    })
    
    // 注册路由
    container.Invoke(func(chat *client.ChatClient) {
        app.Post("/api/chat/completions", handlers.ChatCompletion(chat))
        app.Post("/api/chat/stream", handlers.ChatStream(chat))
    })
    
    app.Listen(":3000")
}
```

- **空白服务端** (`server/`)：仅注册 Fiber 和 DI，路由为空，供后续业务填充
- **示例服务** (`examples/`)：注册完整演示路由，展示 SDK 用法
- 统一使用 `app.Group("/api/v1")` 做 API 版本前缀

**理由**：
- Fiber v3 性能优异，API 简洁，中间件生态丰富
- `Prefork: false` 开发友好，生产可开启
- DI 在路由注册阶段解析，避免 Handler 中重复创建 Client

**备选方案**：
- Gin — 拒绝，Fiber 性能更好，v3 API 更现代
- Echo — 拒绝，Fiber 路由性能更优
- 标准库 `net/http` — 拒绝，缺少路由、中间件、热重载等必要功能

### 6. API 封装设计：面向接口 + 泛型响应

```go
// 通用响应包装
type Response[T any] struct {
    Code    string `json:"code"`
    Message string `json:"message"`
    Data    T      `json:"data"`
}

// Chat API
type ChatClient struct {
    http *HTTPClient
}

func (c *ChatClient) SendMessage(ctx context.Context, req ChatRequest) (*Response[ChatMessage], error)
func (c *ChatClient) SendMessageStream(ctx context.Context, req ChatRequest) (<-chan SSEEvent, error)
func (c *ChatClient) GetConversations(ctx context.Context, user string) (*Response[[]Conversation], error)
func (c *ChatClient) GetMessages(ctx context.Context, conversationID string) (*Response[[]Message], error)
func (c *ChatClient) Feedback(ctx context.Context, messageID string, req FeedbackRequest) error

// Workflow API
type WorkflowClient struct {
    http *HTTPClient
}

func (c *WorkflowClient) Run(ctx context.Context, req WorkflowRunRequest) (*Response[WorkflowResult], error)
func (c *WorkflowClient) GetResult(ctx context.Context, workflowID string) (*Response[WorkflowResult], error)
func (c *WorkflowClient) Stop(ctx context.Context, workflowID string) error
```

- 每个 API 领域一个 Client（Chat、Workflow、Knowledge、File、Model）
- 共享底层 `HTTPClient`，通过 DI 注入
- 请求/响应使用强类型结构体，避免 `map[string]interface{}`
- SSE 流式返回 `chan SSEEvent`，调用方可用 `for range` 消费

**理由**：
- 领域分离让 API 变更影响面最小化
- 泛型 `Response[T]` 统一响应格式，减少样板代码
- 流式用 channel 是 Go 惯用法，调用方控制消费节奏

### 7. 内存存储：双层索引 + 预加载

```go
type MemoryStore struct {
    mu      sync.RWMutex
    apps    map[string]*AppMeta      // key: app_id
    byType  map[string][]*AppMeta    // key: "chat" | "workflow" | "agent"
    byKey   map[string][]*AppMeta    // key: api_key
}

type AppMeta struct {
    ID          string
    Name        string
    Type        string    // chat | workflow | agent | knowledge
    APIKey      string
    Description string
    UpdatedAt   time.Time
}

func (s *MemoryStore) Preload(ctx context.Context, keys []string, fetcher MetadataFetcher) error
func (s *MemoryStore) GetByID(id string) (*AppMeta, bool)
func (s *MemoryStore) GetByType(typ string) []*AppMeta
func (s *MemoryStore) GetByKey(key string) []*AppMeta
```

- **预加载**：启动时遍历所有 API Key，并发拉取元数据，写入 Store
- **双层索引**：按 ID 精确查 + 按类型列表查 + 按 Key 查
- **并发安全**：`sync.RWMutex`，预加载时写锁，查询时读锁

**理由**：
- 预加载避免运行时首次查询延迟
- 多索引覆盖常见查询场景
- RWMutex 适合写少读多（预加载后全为读）

### 8. 多 API Key 管理：轮转 + 限流

```go
type KeyManager struct {
    keys    []string
    current atomic.Int32
}

func (km *KeyManager) Next() string       // 轮询获取下一个 Key
func (km *KeyManager) All() []string     // 获取所有 Key
func (km *KeyManager) MarkFailed(key string) // 标记 Key 失效（可选）
```

- 默认轮询（round-robin）分配 Key，避免单 Key 限流
- 支持标记失效（如收到 429 时临时跳过）
- 所有 Key 从 `DIFY_API_KEYS`（逗号分隔）解析

**理由**：
- 轮询是最简单的负载均衡，无需额外状态
- 失效标记可应对 Dify 的速率限制

## 风险 / 权衡

- **[R] Dify API 版本变更导致 SDK 不兼容** → 每个 API Client 封装在独立文件，变更影响面小；Base URL 可配置，支持多版本切换
- **[R] 多 Key 并发拉取触发 Dify 限流** → Semaphore 限制并发（默认 5），指数退避重试，KeyManager 标记失效 Key
- **[R] SSE 流式连接长时间占用** → context 超时控制，客户端可调用 `Cancel()` 中断；服务端设置 `IdleTimeout`
- **[R] dig.Container 启动时依赖缺失导致 panic** → 所有 Provider 在 `BuildContainer` 中显式注册，启动前做 `container.Invoke` 验证依赖完整性
- **[R] Fiber v3 尚处 beta（2025 年）** → v3 API 已稳定，且 Fiber 社区活跃；如 v3 有重大变更，API 封装层可隔离影响
- **[T] 内存存储无持久化** → 进程重启后重新预加载，当前可接受；后续如需持久化，Store 实现 `PersistentStore` 接口即可替换

## 待解决问题

1. Dify API 的完整 OpenAPI/Swagger 定义是否可获取？（影响请求/响应结构体设计精度）
2. 是否需要支持 Dify 的 OAuth2 / API Token 两种认证方式？（当前仅支持 API Key）
3. 示例服务的路由设计是否需要与 Dify 的 REST API 路径保持一致？（还是自定义路径？）
