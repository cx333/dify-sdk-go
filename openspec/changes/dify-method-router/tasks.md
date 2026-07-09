## 1. MethodRegistry 核心实现

- [x] 1.1 创建 `server/pkg/methods/` 包，定义 `MethodDef` 结构体和 `MethodHandler` 函数签名
- [x] 1.2 实现 `MethodRegistry` 结构体（map + sync.RWMutex），包含 `Register`、`Get`、`List` 方法
- [x] 1.3 编写 `MethodRegistry` 单元测试（注册、查找、列表、重复注册、并发安全 `-race`）

## 2. Fiber Handler 实现

- [x] 2.1 实现 `POST /api/v1/methods` handler：解析 method + params、查找 registry、调用 handler、返回统一响应
- [x] 2.2 实现 `GET /api/v1/methods` handler：调用 `registry.List()` 返回方法列表
- [x] 2.3 实现请求超时控制：从 context 创建超时子 context，超时后返回 504
- [x] 2.4 编写 handler 单元测试（使用 `httptest` + Fiber test，覆盖所有错误场景）

## 3. DI 容器集成

- [x] 3.1 在 `server/internal/di/container.go` 中注册 `MethodRegistry` Provider
- [x] 3.2 在 `server/internal/app/app.go` 中集成 MethodRegistry 并传递给路由

## 4. 配置

- [x] 4.1 在 `server/.env.example` 中新增 `METHOD_TIMEOUT=30s` 配置项
- [x] 4.2 在 `server/internal/config/config.go` 的 `ServerConfig` 结构体中新增 `MethodTimeout` 字段

## 5. 集成验证

- [x] 5.1 在 `server/cmd/server/main.go` 中注册示例方法 `ping`，验证端到端调用链路
- [x] 5.2 运行 `go vet ./...` 和 `go test -race ./...` 确保全部通过
