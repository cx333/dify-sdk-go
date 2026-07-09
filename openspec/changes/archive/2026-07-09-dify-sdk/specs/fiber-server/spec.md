## ADDED Requirements

### Requirement: Fiber v3 初始化
Fiber Server SHALL 使用 `github.com/gofiber/fiber/v3` 创建应用实例，支持自定义配置。

#### Scenario: 创建应用
- **WHEN** 调用 `fiber.New(fiber.Config{ServerHeader: "dify-sdk"})`
- **THEN** 返回 `*fiber.App` 实例
- **AND** 响应头中包含 `Server: dify-sdk`

### Requirement: API 版本前缀
所有路由 SHALL 以 `/api/v1` 为统一前缀，通过 `app.Group` 注册。

#### Scenario: 路由分组
- **WHEN** 注册 `POST /chat` 到 `api := app.Group("/api/v1")`
- **THEN** 实际访问路径 SHALL 为 `/api/v1/chat`

### Requirement: 健康检查端点
Fiber Server SHALL 提供 `/health` 健康检查端点，返回服务状态。

#### Scenario: 健康检查
- **WHEN** 发送 `GET /health` 请求
- **THEN** 返回 HTTP 200 和 JSON `{"status":"ok"}`

### Requirement: 统一错误响应
Fiber Server SHALL 对所有未捕获错误返回统一格式的 JSON 错误响应，包含错误码和消息。

#### Scenario: 处理 panic
- **WHEN** Handler 中发生 panic
- **THEN** 返回 HTTP 500 和 JSON `{"code":"INTERNAL_ERROR","message":"..."}`
- **AND** 服务 SHALL 不崩溃
