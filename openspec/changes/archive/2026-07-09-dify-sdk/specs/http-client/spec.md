## ADDED Requirements

### Requirement: 连接池复用
HTTP Client SHALL 复用底层 `http.Client` 连接池，设置 `MaxIdleConns` 和 `MaxIdleConnsPerHost` 以支持高并发场景。

#### Scenario: 并发请求
- **WHEN** 同时发起 100 个 HTTP 请求
- **THEN** 连接复用率 SHALL 大于 80%，TCP 连接数 SHALL 不超过 20

### Requirement: 请求级超时
每个 HTTP 请求 SHALL 支持独立的超时控制，通过 `context.WithTimeout` 实现，默认 30 秒。

#### Scenario: 请求超时
- **WHEN** 发送请求且服务端 60 秒内无响应
- **THEN** 客户端 SHALL 在 30 秒时返回 `context deadline exceeded` 错误

### Requirement: 指数退避重试
HTTP Client SHALL 对符合条件的错误（网络错误、429、502、503、504）执行指数退避重试，默认最多 3 次。

#### Scenario: 服务端临时不可用
- **WHEN** 服务端返回 503 状态码
- **THEN** 客户端 SHALL 自动重试最多 3 次，每次间隔指数递增

### Requirement: SSE 流式响应
HTTP Client SHALL 支持 Server-Sent Events (SSE) 流式响应，通过 `Accept: text/event-stream` 触发，返回 `chan SSEEvent` 供调用方消费。

#### Scenario: 流式对话
- **WHEN** 调用 Chat Stream API
- **THEN** 客户端 SHALL 逐行读取 SSE 数据流
- **AND** 通过 channel 发送每个解析后的 `SSEEvent`

### Requirement: 统一错误处理
HTTP Client SHALL 将所有 HTTP 错误统一包装为包含状态码和错误信息的 `DifyError` 类型。

#### Scenario: API 返回 4xx 错误
- **WHEN** Dify API 返回 400 状态码及错误 JSON
- **THEN** 返回的 error SHALL 为 `*DifyError` 类型
- **AND** `DifyError.StatusCode` SHALL 等于 400
