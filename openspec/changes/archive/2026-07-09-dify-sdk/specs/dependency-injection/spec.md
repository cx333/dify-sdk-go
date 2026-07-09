## ADDED Requirements

### Requirement: Container 初始化
Dependency Injection 模块 SHALL 提供 `BuildContainer` 函数，创建并配置 `dig.Container`，注册所有 Provider。

#### Scenario: 构建容器
- **WHEN** 调用 `di.BuildContainer(cfg)`
- **THEN** 返回已配置好的 `*dig.Container`
- **AND** 容器中 SHALL 包含 `Config`、`HTTPClient`、`ChatClient`、`WorkflowClient` 等依赖

### Requirement: Provider 注册
所有核心组件（Config、HTTPClient、各 API Client、Store）SHALL 通过构造函数自动注册为 Provider，无需手动逐个注册。

#### Scenario: 自动解析依赖
- **WHEN** `BuildContainer` 中仅提供 `*Config` 和构造函数
- **THEN** `dig` SHALL 自动解析依赖图并实例化所有组件

### Requirement: 生命周期管理
DI 容器 SHALL 支持组件的生命周期管理，启动时验证依赖完整性，关闭时调用 `Shutdowner` 接口释放资源。

#### Scenario: 启动验证
- **WHEN** 调用 `container.Invoke` 解析所有依赖
- **THEN** 如有缺失 Provider SHALL 返回明确错误
- **AND** 所有依赖 SHALL 成功实例化
