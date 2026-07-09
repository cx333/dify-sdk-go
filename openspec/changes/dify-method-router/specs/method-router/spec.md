## ADDED Requirements

### Requirement: 方法调用端点
服务 SHALL 提供 `POST /api/v1/methods` 端点，接受 JSON 请求体 `{"method": "<name>", "params": {...}}`，动态分发到已注册的方法并返回统一格式的 JSON 响应。

#### Scenario: 成功调用方法
- **WHEN** 发送 `POST /api/v1/methods`，body 为 `{"method": "greet", "params": {"name": "Dify"}}`
- **AND** `greet` 方法已注册且执行成功返回 `{"message": "Hello, Dify"}`
- **THEN** 返回 HTTP 200，body 为 `{"success": true, "data": {"message": "Hello, Dify"}}`

#### Scenario: 方法未注册
- **WHEN** 发送 `POST /api/v1/methods`，body 为 `{"method": "unknown_method", "params": {}}`
- **AND** `unknown_method` 未在注册表中注册
- **THEN** 返回 HTTP 404，body 为 `{"success": false, "error": {"code": "METHOD_NOT_FOUND", "message": "方法 unknown_method 未注册"}}`

#### Scenario: 缺少 method 字段
- **WHEN** 发送 `POST /api/v1/methods`，body 为 `{"params": {}}`
- **THEN** 返回 HTTP 400，body 为 `{"success": false, "error": {"code": "INVALID_PARAMS", "message": "请求体必须包含 method 字段"}}`

#### Scenario: 请求体非 JSON
- **WHEN** 发送 `POST /api/v1/methods`，Content-Type 为 `application/json`，body 为 `not-valid-json`
- **THEN** 返回 HTTP 400，body 中 `success` 为 `false`，`error.code` 为 `"INVALID_PARAMS"`

#### Scenario: 方法执行出错
- **WHEN** 发送 `POST /api/v1/methods` 调用已注册方法
- **AND** 方法 handler 返回 error
- **THEN** 返回 HTTP 500，body 为 `{"success": false, "error": {"code": "METHOD_ERROR", "message": "<error 消息>"}}`

#### Scenario: 方法执行超时
- **WHEN** 发送 `POST /api/v1/methods` 调用已注册方法
- **AND** 方法执行时间超过配置的 `METHOD_TIMEOUT`
- **THEN** 返回 HTTP 504，body 为 `{"success": false, "error": {"code": "METHOD_TIMEOUT", "message": "方法执行超时"}}`

### Requirement: params 字段可选
`params` 字段 SHALL 为可选字段，当方法不需要参数时允许省略。

#### Scenario: params 省略
- **WHEN** 发送 `POST /api/v1/methods`，body 为 `{"method": "ping"}`
- **AND** `ping` 方法已注册且不需要参数
- **THEN** 返回 HTTP 200，`success` 为 `true`

#### Scenario: params 为 null
- **WHEN** 发送 `POST /api/v1/methods`，body 为 `{"method": "ping", "params": null}`
- **THEN** 行为与 params 省略相同，正常调用方法

### Requirement: 方法发现端点
服务 SHALL 提供 `GET /api/v1/methods` 端点，返回所有已注册方法的元数据列表。

#### Scenario: 列出所有方法
- **WHEN** 注册表中有 `greet` 和 `add` 两个方法
- **AND** 发送 `GET /api/v1/methods`
- **THEN** 返回 HTTP 200，body 为包含两个方法元数据对象的数组
- **AND** 每个方法对象包含 `name`、`description`、`input_schema`、`output_schema` 字段
- **AND** 不包含 `handler` 字段

#### Scenario: 无注册方法
- **WHEN** 注册表中无任何方法
- **AND** 发送 `GET /api/v1/methods`
- **THEN** 返回 HTTP 200，body 为空数组 `[]`

### Requirement: 统一错误格式
所有方法调用错误 SHALL 使用统一格式 `{"success": false, "error": {"code": "<code>", "message": "<message>"}}`。

#### Scenario: 错误响应结构完整性
- **WHEN** 任何方法调用失败
- **THEN** 响应 SHALL 包含 `success`（布尔值 false）、`error`（对象）、`error.code`（字符串）、`error.message`（字符串）
