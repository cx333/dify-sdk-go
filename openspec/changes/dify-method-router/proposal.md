## Why

Dify 的 HTTP 请求组件功能有限，复杂的业务逻辑（数据处理、外部系统集成、自定义计算）无法在工作流中直接表达。需要在 Go 后端提供一个通用 RPC 端点，Dify 通过传入方法名即可调用后端任意注册方法，将 Go 的能力作为 Dify 工作流的扩展点。

## What Changes

- 在 `server` 模块新增 `/api/v1/methods` 通用 RPC 端点，接受 `method` + `params` JSON body，动态分发到已注册的处理函数
- 新增 **方法注册表（MethodRegistry）**：支持按名称注册方法，每个方法有明确的参数/返回值 schema 描述
- 新增 **方法发现端点** `GET /api/v1/methods`：返回所有已注册方法的名称、参数 schema、描述，供 Dify 侧查阅
- 支持通过 `.env` 配置端点前缀和方法调用超时
- `MethodRegistry` 通过 `dig.Container` 注入到 Fiber Server

## Capabilities

### New Capabilities

- `method-router`: 通用 HTTP 端点 POST /api/v1/methods，接收 method + params，动态分发到注册的处理函数并返回结果
- `method-registry`: 方法注册表，支持按名称注册/查找方法，每个方法携带参数/返回值 schema 和描述

### Modified Capabilities

- `fiber-server`: 在 `/api/v1` 分组下新增 `POST /methods` 和 `GET /methods` 路由，MethodRegistry 通过 DI 注入

## Impact

- **代码**: `server/main.go` 新增路由注册；新增 `server/methods/` 包（registry + handler）；`server/di/` 新增 Provider
- **依赖**: 无新增外部依赖，使用 Fiber v3 和 dig（已有）
- **配置**: `.env` 新增 `METHOD_TIMEOUT`（默认 30s）
- **API**: 新增 `POST /api/v1/methods`（调用方法）和 `GET /api/v1/methods`（列出方法）
