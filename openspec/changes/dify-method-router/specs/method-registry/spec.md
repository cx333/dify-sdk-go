## ADDED Requirements

### Requirement: 方法注册
MethodRegistry SHALL 提供 `Register(name string, def *MethodDef) error` 方法，将方法按名称注册到注册表。

#### Scenario: 注册新方法
- **WHEN** 调用 `registry.Register("greet", &MethodDef{...})`
- **THEN** 注册表 SHALL 包含名为 `greet` 的方法
- **AND** 返回 nil error

#### Scenario: 重复注册同名方法
- **WHEN** 调用 `registry.Register("greet", def1)` 成功后再调用 `registry.Register("greet", def2)`
- **THEN** 返回非 nil error
- **AND** 注册表保留第一个注册的定义

### Requirement: 方法查找
MethodRegistry SHALL 提供 `Get(name string) (*MethodDef, bool)` 方法，按名称查找已注册方法。

#### Scenario: 查找已注册方法
- **WHEN** 调用 `registry.Get("greet")` 且 `greet` 已注册
- **THEN** 返回对应的 `*MethodDef` 和 `true`

#### Scenario: 查找未注册方法
- **WHEN** 调用 `registry.Get("unknown")` 且 `unknown` 未注册
- **THEN** 返回 nil 和 `false`

### Requirement: 列出所有方法
MethodRegistry SHALL 提供 `List() []*MethodDef` 方法，返回所有已注册方法的元数据列表（不含 Handler）。

#### Scenario: 列出含多个方法
- **WHEN** 注册表中有 2 个方法
- **AND** 调用 `registry.List()`
- **THEN** 返回长度为 2 的切片
- **AND** 每个 MethodDef 的 Handler 字段为 nil

#### Scenario: 空注册表列出
- **WHEN** 注册表中无方法
- **AND** 调用 `registry.List()`
- **THEN** 返回空切片（非 nil）

### Requirement: 并发安全
MethodRegistry SHALL 的 Register、Get、List 操作 SHALL 是并发安全的。

#### Scenario: 并发读写
- **WHEN** 多个 goroutine 同时调用 Register 和 Get
- **THEN** 不发生 data race
- **AND** Get 始终返回一致的结果（注册前 false，注册后 true）

### Requirement: MethodDef 元数据结构
`MethodDef` 结构体 SHALL 包含 `Name`（方法名）、`Description`（描述）、`InputSchema`（参数 JSON Schema）、`OutputSchema`（返回值 JSON Schema）、`Handler`（处理函数）。

#### Scenario: MethodDef 完整定义
- **WHEN** 创建一个 MethodDef 实例
- **THEN** 所有五个字段均可设置
- **AND** Handler 字段在 JSON 序列化时 SHALL 被忽略（`json:"-"`）

### Requirement: MethodHandler 签名
`MethodHandler` SHALL 定义为 `func(ctx context.Context, params json.RawMessage) (interface{}, error)`，params 为原始 JSON，result 为任意可 JSON 序列化的值。

#### Scenario: Handler 接收原始 JSON params
- **WHEN** 端点到手 `{"method": "echo", "params": {"key": "val"}}`
- **THEN** Handler 收到的 `params` 参数为 `json.RawMessage(`{"key": "val"}`)`

#### Scenario: Handler 返回错误
- **WHEN** Handler 返回 `nil, errors.New("something went wrong")`
- **THEN** 调用方收到 HTTP 500 响应，`error.code` 为 `"METHOD_ERROR"`
