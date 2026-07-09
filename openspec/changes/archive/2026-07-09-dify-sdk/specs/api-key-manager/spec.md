## ADDED Requirements

### Requirement: 多 Key 存储
API Key Manager SHALL 支持存储和管理多个 API Key，从环境变量 `DIFY_API_KEYS`（逗号分隔）解析。

#### Scenario: 加载多个 Key
- **WHEN** 环境变量 `DIFY_API_KEYS=app-1,app-2,app-3`
- **THEN** `KeyManager.All()` SHALL 返回 `[]string{"app-1", "app-2", "app-3"}`

### Requirement: 轮询分配
API Key Manager SHALL 支持轮询（round-robin）方式分配 API Key，用于负载均衡。

#### Scenario: 轮询获取 Key
- **WHEN** 连续调用 `KeyManager.Next()` 4 次
- **THEN** 返回的 Key 序列 SHALL 为 `app-1, app-2, app-3, app-1`

### Requirement: 失效 Key 标记
API Key Manager SHALL 支持标记失效的 Key（如收到 429 限流时），在轮询中自动跳过。

#### Scenario: 跳过失效 Key
- **WHEN** 调用 `KeyManager.MarkFailed("app-2")`
- **AND** 连续调用 `KeyManager.Next()` 3 次
- **THEN** 返回的 Key 序列 SHALL 为 `app-1, app-3, app-1`
- **AND** 不包含 `app-2`
