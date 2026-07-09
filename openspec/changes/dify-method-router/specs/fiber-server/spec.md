## ADDED Requirements

### Requirement: 方法路由注册
Fiber Server SHALL 在 `/api/v1` 分组下注册 `POST /methods` 和 `GET /methods` 路由，通过 DI 注入 MethodRegistry。

#### Scenario: POST /api/v1/methods 可用
- **WHEN** 服务启动且 MethodRegistry 已注入
- **AND** 发送 `POST /api/v1/methods` 请求
- **THEN** 请求被路由到方法调用 handler

#### Scenario: GET /api/v1/methods 可用
- **WHEN** 服务启动且 MethodRegistry 已注入
- **AND** 发送 `GET /api/v1/methods` 请求
- **THEN** 返回已注册方法列表

### Requirement: MethodRegistry 依赖注入
Fiber Server 的 DI 容器 SHALL 注册 MethodRegistry 为单例 Provider。

#### Scenario: MethodRegistry 注入到容器
- **WHEN** 调用 `di.BuildContainer(cfg)`
- **THEN** 容器中 SHALL 可解析出 `*methods.MethodRegistry` 实例
- **AND** 解析出的实例为单例（多次 Invoke 获取同一实例）
