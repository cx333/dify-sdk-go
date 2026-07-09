# Dify API Specs

Dify 平台的 OpenAPI 3.0 规范文档，用于指导 `server/` 和 `dify-sdk/` 的开发实现。

## 目录结构

```
api-specs/
├── openapi_chat.json              # 对话型应用 API（85 KB）
├── openapi_chatflow.json          # Chatflow 编排应用 API（145 KB）
├── openapi_workflow.json          # Workflow 工作流 API（100 KB）
├── openapi_completion.json        # 文本生成应用 API（49 KB）
├── openapi_knowledge.json         # 知识库管理 API（43 KB）
├── external-knowledge-api.json    # 外部知识库检索 API（3.6 KB）
├── external-knowledge-api 2.json  # 同上（副本）
├── 外部知识库 API.json             # 同上（中文命名副本）
└── README.md
```

## 各规范说明

| 文件 | 对应 Dify 应用类型 | 核心接口 |
|------|-------------------|---------|
| `openapi_chat.json` | 对话型应用 | `POST /chat-messages` 发送对话、流式/阻塞双模式 |
| `openapi_chatflow.json` | Chatflow 编排应用 | `POST /chat-messages`、运行日志、会话变量、工作流管理 |
| `openapi_workflow.json` | Workflow 工作流 | `POST /workflows/run` 执行工作流、日志、停止 |
| `openapi_completion.json` | 文本生成应用 | `POST /completion-messages` 发送补全请求 |
| `openapi_knowledge.json` | 知识库管理 | 数据集 CRUD、文档增删改、分段管理、检索 |
| `external-knowledge-api.json` | 外部知识库 | `POST /retrieval` 知识召回、自定义鉴权 |

## 通用约定

- **鉴权**：所有接口通过 `Authorization: Bearer {api_key}` 认证
- **请求格式**：`Content-Type: application/json`
- **流式响应**：`text/event-stream`（SSE 协议），`response_mode: streaming`
- **阻塞响应**：`application/json`，`response_mode: blocking`
- **分页**：`page` + `limit` 参数，`has_more` 标识是否有下一页
- **错误码**：HTTP 400/403/404/413/415/429/500，响应体含 `code`、`message`、`status`

## 使用方式

### 查看规范

```bash
# 使用 Swagger Editor（Web）
open https://editor.swagger.io/

# 使用 redoc-cli 生成静态文档
npx @redocly/cli build openapi_chat.json -o chat.html

# 使用 openapi-generator 生成客户端代码
openapi-generator generate -i openapi_chat.json -g go -o gen/chat-client
```

### 与服务端的对应关系

```
api-specs/openapi_*.json    →    server/（HTTP 转发层）
                                      ↓
                              dify-sdk/client/（SDK HTTP 客户端）
                                      ↓
                              Dify 后端服务
```

- **`server/`** 作为中间层，接收 Dify HTTP 请求节点的调用，通过 `methods` 注册机制转发到 Go 业务逻辑
- **`dify-sdk/`** 作为客户端 SDK，封装对 Dify API 的调用（鉴权、重试、流式解析）
- **`api-specs/`** 是两者的共同契约，新增端点时应先查阅对应规范

### 注意事项

- 目录中存在 `external-knowledge-api 2.json` 和 `外部知识库 API.json` 两份副本，内容与 `external-knowledge-api.json` 相同，后续应清理
- 规范文件为 Dify 官方导出，修改时注意保持与 Dify 版本的兼容性
- 当前 `server/` 和 `dify-sdk/` 尚未完整实现所有规范接口，开发时以此目录为参考
