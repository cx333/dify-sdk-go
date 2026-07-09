## ADDED Requirements

### Requirement: 三层索引存储
Metadata Store SHALL 支持按应用 ID、应用类型、API Key 三层索引存储元数据。

#### Scenario: 存储并查询
- **WHEN** 存储 3 个 Agent（key=app-1）和 2 个 Workflow（key=app-2）
- **THEN** `GetByType("agent")` SHALL 返回 3 条记录
- **AND** `GetByKey("app-2")` SHALL 返回 2 条记录
- **AND** `GetByID("agent-1")` SHALL 返回对应单条记录

### Requirement: 并发安全
Metadata Store SHALL 在并发读写场景下保持数据一致性，使用 `sync.RWMutex` 保护内部状态。

#### Scenario: 并发读写
- **WHEN** 10 个 goroutine 同时写入元数据
- **AND** 10 个 goroutine 同时读取元数据
- **THEN** 所有读写操作 SHALL 无数据竞争（通过 `go test -race` 验证）

### Requirement: 启动预加载
Metadata Store SHALL 支持在应用启动时预加载所有 API Key 对应的元数据。

#### Scenario: 预加载
- **WHEN** 调用 `Store.Preload(ctx, keys, fetcher)`
- **THEN** 遍历所有 Key 并发拉取元数据
- **AND** 预加载完成后 `GetByType` SHALL 返回非空结果
