# Dify Server

Dify HTTP 请求转发服务。作为 Dify 工作流中的 HTTP 中间层，接收 Dify 的 HTTP 请求组件调用，通过通用方法注册机制动态分发到 Go 业务逻辑。支持多 API Key，启动时自动发现已配置的应用及其类型（聊天助手/智能体/工作流/对话流/文本生成）。

基于 Fiber v3 + dig 依赖注入 + slog 结构化日志。

## 架构

```
                          ┌──────────────────────────────────┐
                          │            server/               │
                          │                                  │
                          │  ┌────────────┐  ┌───────────┐  │
 Dify 平台 ──HTTP─────────→│  │  router/   │  │   di/     │  │        ┌──────────────┐
                          │  │ Fiber v3   │  │ ClientPool│──┼───────→│  Dify 后端    │
                          │  └─────┬──────┘  │ 多 Key 池  │  │  SDK   │  (多应用)     │
                          │        │         └───────────┘  │        │              │
                          │  ┌─────┴──────┐                 │  ┌─────┴──────────────┤
                          │  │  handler/  │  ┌───────────┐  │  │ app-xxx (agent)   │
                          │  │ apps/      │  │  app/     │  │  │ app-yyy (workflow)│
                          │  │ methods/   │  │ 启动编排   │  │  │ app-zzz (chat)    │
                          │  └───────────┘  └───────────┘  │  └────────────────────┘
                          └──────────────────────────────────┘
                                     │
                              pkg/methods/
                              (动态方法注册)
```

### 启动发现流程

```
 加载 .env → 拿到所有 API Key
                  │
                  ├─ 每个 Key 创建 HTTPClient → ClientPool（共享 transport）
                  ├─ 每个 Key 调 GET /info → 获取 Name / Mode / Description
                  └─ 存入 MemoryStore → GET /api/v1/apps 可查
```

### 请求路由流程

```
 POST /api/v1/methods  {"method":"call_dify","params":{"query":"你好","key_index":1}}
          │
          ▼
 MethodRegistry.Execute("call_dify", params)
          │
          ▼
 main.go 闭包 → App.HTTPClientByIndex(1) → ClientPool.Get(1) → 第 2 把 key
          │
          ▼
 HTTPClient.Do("POST", "/chat-messages", ...) → Dify 对应应用
```

## 目录结构

```
server/
├── cmd/
│   ├── server/main.go          # 应用入口（最薄启动层）
│   └── tools/main.go           # 运维工具（开发辅助脚本）
├── internal/
│   ├── app/app.go              # 启动编排：配置→日志→DI→预加载→路由→监听
│   ├── config/config.go        # 服务配置（端口、环境、超时、日志）
│   ├── di/container.go         # 依赖注入容器（ClientPool + HTTPClient + MethodRegistry）
│   ├── handler/
│   │   ├── health.go           # GET  /health
│   │   ├── api.go              # GET  /api/v1/
│   │   └── apps.go             # GET  /api/v1/apps（应用列表）
│   ├── logger/logger.go        # slog 封装（context 注入、双格式输出）
│   ├── middleware/middleware.go # RequestID + 结构化请求日志
│   └── router/
│       ├── router.go           # Fiber 应用创建 + 全局中间件
│       ├── health.go           # 健康检查路由注册
│       ├── api.go              # API v1 + 应用列表路由注册
│       └── methods.go          # 方法调用路由注册
├── pkg/
│   └── methods/                # 通用方法注册表（可复用库）
│       ├── registry.go         # 线程安全注册表
│       ├── handler.go          # HTTP handler（JSON 请求/响应信封）
│       └── doc.go              # 包文档
├── go.mod / go.sum
├── .env.example
└── README.md
```

### 分层职责

| 层级 | 目录 | 职责 |
|------|------|------|
| 入口 | `cmd/` | 启动进程、注册业务方法 |
| 编排 | `internal/app/` | 组装依赖、预加载应用元数据、管理生命周期 |
| 配置 | `internal/config/` | 环境变量 → 结构化配置 |
| 传输 | `internal/router/` + `internal/handler/` | HTTP 路由与处理器 |
| 基础设施 | `internal/middleware/`、`internal/logger/` | 横切关注点 |
| 可复用库 | `pkg/methods/` | 通用方法注册模式 |
| SDK | `../dify-sdk/` | Dify API 客户端封装（AppMode、MemoryStore） |

## 技术栈

| 组件 | 选型 | 说明 |
|------|------|------|
| HTTP 框架 | Fiber v3 | fasthttp 驱动，高性能 |
| 依赖注入 | uber/dig | 构造函数注入，自动解析依赖图 |
| 日志 | log/slog | Go 标准库结构化日志 |
| 配置 | godotenv | .env 文件加载 |
| UUID | google/uuid | 请求 ID 生成 |

## 快速开始

### 前置条件

- Go 1.26+
- 可访问的 Dify 实例（或本地 `docker compose` 部署）
- 至少一个 Dify 应用的 API Key

### 安装与运行

```bash
cd server

# 1. 配置环境变量
cp .env.example .env
# 编辑 .env，按编号填写各应用的 API Key

# 2. 安装依赖
go mod tidy

# 3. 启动服务
go run ./cmd/server/

# 4. 验证
curl http://localhost:8081/health
# → {"status":"ok"}

curl http://localhost:8081/api/v1/apps
# → [{"index":0,"name":"一致性披露检查","mode":"advanced-chat","mode_label":"对话流",...}]
```

### 构建

```bash
# 开发构建
go build -o server ./cmd/server/

# 生产构建（静态链接，无 CGO 依赖）
CGO_ENABLED=0 go build -ldflags="-s -w" -o server ./cmd/server/
```

## 配置

所有配置通过环境变量或 `.env` 文件注入：

| 变量 | 默认值 | 说明 |
|------|--------|------|
| `SERVER_PORT` | `8080` | 服务监听端口 |
| `ENV` | `development` | 运行环境（development / production） |
| `METHOD_TIMEOUT` | `30s` | 方法执行超时上限 |
| `LOG_LEVEL` | `info` | 日志级别（debug / info / warn / error） |
| `LOG_FORMAT` | `text` | 日志格式（text / json，生产建议 json） |
| `DIFY_BASE_URL` | — | Dify API 基础地址（必填） |
| `DIFY_API_KEY_0` | — | 第 1 个 API Key（必填，编号从 0 开始） |
| `DIFY_API_KEY_1` | — | 第 2 个 API Key（可选） |
| `DIFY_API_KEY_N` | — | 更多 Key 按编号递增（可选） |
| `DIFY_API_KEYS` | — | 逗号分隔的备用格式（仅无编号变量时生效） |
| `DIFY_TIMEOUT` | `30s` | SDK 请求超时 |
| `DIFY_MAX_RETRIES` | `3` | SDK 重试次数 |
| `OUTBOUND_RPS` | `0` | 出站限流每秒最大请求数（0=不限） |
| `OUTBOUND_BURST` | `10` | 出站限流突发容量 |
| `MAX_CONCURRENT_METHODS` | `0` | 方法执行并发上限（0=不限） |

### 多 Key 配置

推荐使用编号环境变量，一个 Key 对应一个 Dify 应用：

```env
# 推荐：编号方式，清晰直观
DIFY_API_KEY_0=app-workflow-xxxxxxxx    # 工作流
DIFY_API_KEY_1=app-agent-yyyyyyyy       # 智能体
DIFY_API_KEY_2=app-chat-zzzzzzzz        # 聊天助手
```

启动时服务会逐个调用 `/info` 自动发现每个应用的名称和类型，启动日志中可看到：

```
发现应用  key_index=0  name=一致性披露检查  mode=advanced-chat  mode_label=对话流
发现应用  key_index=1  name=客服智能体      mode=agent-chat     mode_label=智能体
```

## API 参考

### 健康检查

```
GET /health
→ 200 {"status":"ok"}
```

### API 信息

```
GET /api/v1/
→ 200 {"message":"Dify server ready"}
```

### 应用列表

```
GET /api/v1/apps
→ 200 [
  {
    "index": 0,
    "name": "一致性披露检查",
    "mode": "advanced-chat",
    "mode_label": "对话流",
    "description": "",
    "tags": []
  },
  {
    "index": 1,
    "name": "客服智能体",
    "mode": "agent-chat",
    "mode_label": "智能体",
    "description": "",
    "tags": []
  }
]
```

| 字段 | 说明 |
|------|------|
| `index` | key 索引（对应 `DIFY_API_KEY_N` 中的 N，用于后续调用选择目标） |
| `mode` | Dify 原始模式值（chat / agent-chat / advanced-chat / workflow / completion） |
| `mode_label` | 中文标签（聊天助手 / 智能体 / 对话流 / 工作流 / 文本生成） |

### 方法列表

```
GET /api/v1/methods
→ 200 [{"name":"ping","description":"健康检查：返回 pong",...}]
```

### 调用方法

```
POST /api/v1/methods
Content-Type: application/json

{
  "method": "call_dify",
  "params": {
    "query": "你好",
    "key_index": 0
  }
}

→ 200 {"success":true,"data":{"answer":"你好，有什么可以帮你的？"}}

→ 400 {"success":false,"error":{"code":"INVALID_PARAMS","message":"..."}}
→ 404 {"success":false,"error":{"code":"METHOD_NOT_FOUND","message":"..."}}
→ 500 {"success":false,"error":{"code":"METHOD_ERROR","message":"..."}}
→ 504 {"success":false,"error":{"code":"METHOD_TIMEOUT","message":"..."}}
```

`key_index` 参数用于选择目标应用，对应 `GET /api/v1/apps` 返回的 `index` 字段。不传时默认使用 `DIFY_API_KEY_0`。

## 开发指南

### 添加新路由

1. 在 `internal/handler/` 创建处理器：
```go
func NewEchoHandler(log *logger.Logger) *EchoHandler { ... }
func (h *EchoHandler) Handle(c fiber.Ctx) error { ... }
```

2. 在 `internal/router/` 创建路由注册文件：
```go
func registerEchoRoutes(app *fiber.App, log *logger.Logger) { ... }
```

3. 在 `router.go` 的 `Setup()` 中调用注册函数。

4. 如需访问 ClientPool 或 AppStore，通过 `router.Params` 传入。

### 注册业务方法

在 `cmd/server/main.go` 中注册供 Dify 调用的方法：

```go
application.Registry().Register("my_method", &methods.MethodDef{
    Name:        "my_method",
    Description: "方法说明",
    InputSchema: json.RawMessage(`{"type":"object","properties":{...}}`),
    Handler: func(ctx context.Context, params json.RawMessage) (interface{}, error) {
        // 使用 a.HTTPClientByIndex(keyIndex) 选择目标应用
        c := a.HTTPClientByIndex(0)
        // 或使用默认客户端 a.HTTPClient()
        return result, nil
    },
})
```

方法处理器通过闭包访问 `App`，可直接调用：
- `a.HTTPClient()` — 默认客户端（pool[0]）
- `a.HTTPClientByIndex(n)` — 按 key_index 选择客户端
- `a.AppStore()` — 应用元数据缓存

### 添加中间件

在 `internal/middleware/` 创建文件，然后到 `router.go` 的 `Setup()` 中 `app.Use()` 注册。

### 运行测试

```bash
go test ./...                    # 所有测试
go test ./pkg/methods/...        # methods 包测试
go test -v -count=1 ./...        # 详细输出，禁用缓存
```

## 项目依赖

```
server/                         ← 本模块
  ├── dify-sdk/（本地 replace）   ← Dify HTTP 客户端封装
  ├── gofiber/fiber/v3           ← Web 框架
  ├── go.uber.org/dig             ← 依赖注入
  ├── google/uuid                 ← UUID 生成
  └── joho/godotenv              ← .env 加载（间接依赖）
```

`dify-sdk` 通过 `go.mod` 的 `replace` 指令指向本地路径：
```
replace github.com/wgl/dify-sdk => ../dify-sdk
```
