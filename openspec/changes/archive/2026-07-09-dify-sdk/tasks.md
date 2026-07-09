## 1. 项目基础设施

- [x] 1.1 初始化 Go Workspace：创建 `go.work`，添加 `dify-sdk`、`examples`、`server` 三个模块目录
- [x] 1.2 创建 `dify-sdk/go.mod`（module `github.com/wgl/dify-sdk`），添加基础依赖（`godotenv`、`dig`）
- [x] 1.3 创建 `examples/go.mod` 和 `server/go.mod`，添加 `fiber/v3` 及对 `dify-sdk` 的 `replace` 依赖
- [x] 1.4 创建各模块的 `.env` 和 `.env.example` 模板文件，确认 `.gitignore` 忽略 `.env`

## 2. 配置模块（env-config）

- [x] 2.1 实现 `dify-sdk/config/config.go`：`Config` 结构体 + `Load(path string)` 函数，使用 `godotenv` 解析 `.env`
- [x] 2.2 编写 `config` 包的单元测试（正常加载、缺失文件、默认值校验）

## 3. HTTP 客户端（http-client）

- [x] 3.1 实现 `dify-sdk/client/http.go`：创建 `HTTPClient` 结构体，封装 `*http.Client`（连接池、超时配置）
- [x] 3.2 实现 `HTTPClient.doRequest` 方法：统一请求逻辑（设置 Header、JSON 序列化、响应解析）
- [x] 3.3 实现重试机制：指数退避，默认 3 次，仅对网络错误和 429/502/503/504 重试
- [x] 3.4 实现 SSE 流式读取：`streamRequest` 方法，使用 `bufio.Scanner`，通过 `chan SSEEvent` 输出
- [x] 3.5 定义 `DifyError` 类型和通用 `Response[T]` 泛型响应结构体
- [x] 3.6 编写 `http` 包的单元测试（正常请求、超时、重试、SSE 解析）

## 4. 依赖注入容器（dependency-injection）

- [x] 4.1 实现 `dify-sdk/di/container.go`：`BuildContainer(cfg *Config)` 函数，注册 Config、HTTPClient Provider
- [x] 4.2 编写 `di` 包的单元测试（容器构建成功、依赖缺失报错）

## 5. API Key 管理器（api-key-manager）

- [x] 5.1 实现 `dify-sdk/auth/key_manager.go`：`KeyManager` 结构体，轮询 `Next()`、全部获取 `All()`、失效标记 `MarkFailed()`
- [x] 5.2 编写 `auth` 包的单元测试（轮询顺序、失效跳过）

## 6. 内存存储（metadata-store）

- [x] 6.1 实现 `dify-sdk/store/memory.go`：`MemoryStore` 结构体（`sync.RWMutex` + 三层 map 索引）
- [x] 6.2 实现 `Preload(ctx, keys, fetcher)` 方法：并发拉取 + semaphore 限流（默认 5）
- [x] 6.3 编写 `store` 包的单元测试（CRUD、并发安全 `-race`、预加载）

## 7. API 客户端

- [x] 7.1 实现 `dify-sdk/client/chat.go`：`ChatClient`（SendMessage、SendMessageStream、GetConversations、GetMessages、Feedback）
- [x] 7.2 实现 `dify-sdk/client/workflow.go`：`WorkflowClient`（Run、GetResult、Stop）
- [x] 7.3 实现 `dify-sdk/client/knowledge.go`：`KnowledgeClient`（CreateDataset、DeleteDataset、AddDocument、SearchSegments）
- [x] 7.4 实现 `dify-sdk/client/file.go`：`FileClient`（Upload、GetInfo）
- [x] 7.5 实现 `dify-sdk/client/model.go`：`ModelConfigClient`（ListModels、GetParameters）
- [x] 7.6 编写各 API Client 的 mock 测试（使用 `httptest.Server` 模拟 Dify API）

## 8. Fiber 服务端（fiber-server）

- [ ] 8.1 实现 `server/main.go`：加载 `.env`、`dig.BuildContainer`、`fiber.New` 创建空白服务，注册 `/health` 和 `/api/v1` 分组
- [ ] 8.2 实现统一错误处理中间件（recover panic → JSON 错误响应）
- [ ] 8.3 实现 `examples/main.go`：同 server 结构，额外注册 Chat/Workflow 演示路由
- [ ] 8.4 实现 `examples/handler/demo.go`：演示 Handler（调用 SDK 各 API Client 并返回结果）

## 9. 集成测试与验证

- [x] 9.1 编写集成测试：启动 mock Dify server（`httptest`），验证完整调用链（Config → DI → API Client → Store）
- [x] 9.2 验证 `go work` 下所有模块独立构建成功：`cd dify-sdk && go build ./...`、`cd examples && go build ./...`、`cd server && go build ./...`
- [x] 9.3 验证 `go vet ./...` 和 `go test -race ./...` 三模块全部通过
