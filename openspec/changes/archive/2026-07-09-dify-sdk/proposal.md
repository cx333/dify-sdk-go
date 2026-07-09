## 为什么

现有 Dify Go SDK 老旧且功能单一，仅支持基础的元数据查询，缺乏对 Chat、Workflow、Knowledge 等核心 API 的完整支持。同时项目结构为单模块，难以扩展为独立的 SDK 库、示例服务和空白服务端的多模块协作模式。需要从零构建一个**全面、高性能**的 Dify SDK，并以 **Go Workspace** 组织为可复用库 + 示例服务 + 空白服务端的三模块架构。

## 变更内容

- 初始化 Go Workspace，拆分为 `dify-sdk`（库）、`examples`（示例服务）、`server`（空白服务端）三个模块
- 全新设计高性能 HTTP 客户端：连接池复用、请求级超时、重试机制、流式响应支持
- 完整覆盖 Dify API：Chat、Workflow、Knowledge/Dataset、文件上传、模型配置
- 所有模块统一使用 `.env` 文件管理配置，禁止使用 config.yaml
- SDK 与服务端统一使用 `dig.Container` 进行依赖注入
- 空白服务端和示例服务基于 **Fiber v3** 框架构建 HTTP 服务
- 内置内存缓存层，支持 API Key 管理、元数据预加载与查询

## 能力划分

### 新增能力

- `go-workspace`: Go Workspace 多模块管理，包含 `dify-sdk`、`examples`、`server`
- `http-client`: 高性能可复用 HTTP 客户端，支持连接池、超时、重试、流式（SSE）响应
- `env-config`: 统一 `.env` 配置加载（所有模块），禁止 config.yaml
- `chat-api`: Dify Chat API 完整封装（发送消息、流式对话、会话管理、消息反馈）
- `workflow-api`: Dify Workflow API 完整封装（运行工作流、获取执行详情、停止工作流）
- `knowledge-api`: Dify Knowledge/Dataset API（数据集 CRUD、文档管理、段落检索）
- `file-api`: 文件上传与管理 API（上传文件、获取文件信息）
- `model-config`: 模型配置 API（获取模型列表、模型参数配置）
- `api-key-manager`: 多 API Key 管理与自动轮转
- `metadata-store`: 内存元数据存储，支持 Agent/Workflow/Knowledge 等元数据的预加载与查询
- `dependency-injection`: `dig.Container` 依赖注入容器，统一管理 SDK 与服务端依赖
- `fiber-server`: Fiber v3 服务端框架集成（空白服务端 + 示例服务）

### 修改的能力

无（全新项目，无可修改的现有能力）

## 影响范围

- **代码**: 全新 Go Workspace，从零构建三模块架构
- **依赖**: `github.com/gofiber/fiber/v3`、`go.uber.org/dig`、`github.com/joho/godotenv`，Go 标准库 `net/http`、`encoding/json`
- **外部系统**: 本地部署的 Dify 实例
- **配置**: 各模块独立的 `.env` 文件，包含 `DIFY_BASE_URL`、`DIFY_API_KEYS`（逗号分隔）等
- **架构约束**: 所有模块必须使用 `dig.Container` 做 DI，服务端必须使用 Fiber v3
