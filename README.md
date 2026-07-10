# Dify SDK for Go

[![Go Version](https://img.shields.io/badge/Go-1.26+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/license-MIT-green)](./LICENSE)

[Dify](https://dify.ai) 平台的 Go 语言 SDK，提供类型安全的 API 客户端、流式响应支持、自动重试和连接池复用。

## 特性

- **多 API 覆盖** — Chat（对话）、Workflow（工作流）、Knowledge（知识库）、File（文件上传）
- **流式 & 阻塞双模式** — SSE 事件流 + 同步阻塞调用，一套请求两种用法
- **连接池复用** — 内置 HTTP 连接池（MaxIdleConns: 100），高频调用零开销
- **指数退避重试** — 网络错误和可重试状态码（429/502/503/504）自动重试，最多 3 次
- **DI 容器集成** — 基于 `uber/dig` 的依赖注入，开箱即用
- **内存密钥管理** — 支持多 API Key 轮换和元数据存储
- **类型安全** — 完整的请求/响应结构体定义，编译期保证正确性

## 安装

```bash
go get github.com/cx333/dify-sdk-go/dify-sdk
```

## 快速开始

```go
package main

import (
    "context"
    "fmt"
    "time"

    "github.com/cx333/dify-sdk-go/dify-sdk/client"
)

func main() {
    // 1. 创建 HTTP 客户端
    httpClient := client.NewHTTPClient(
        "https://api.dify.ai/v1",  // Dify API 地址
        "app-xxxxxxxxxxxxx",        // API Key
        30*time.Second,             // 请求超时
        client.DefaultRetryConfig(),
    )

    // 2. 发送对话消息（阻塞模式）
    chat := client.NewChatClient(httpClient, "")
    resp, err := chat.SendMessage(context.Background(), client.ChatRequest{
        Query:  "你好，请介绍一下自己",
        User:   "user-001",
        Inputs: map[string]any{},
    })
    if err != nil {
        panic(err)
    }
    fmt.Println(resp.Answer)
}
```

### 流式对话

```go
// 流式模式 — 逐 token 输出
events, errs := chat.SendMessageStream(context.Background(), client.ChatRequest{
    Query: "写一首关于夏天的诗",
    User:  "user-001",
})

for ev := range events {
    fmt.Print(ev.Answer) // 实时打印
}
if err := <-errs; err != nil {
    panic(err)
}
```

### 执行工作流

```go
wf := client.NewWorkflowClient(httpClient, "")
result, err := wf.Run(context.Background(), client.WorkflowRunRequest{
    Inputs: map[string]any{
        "topic": "AI 发展趋势",
    },
    ResponseMode: "blocking",
    User:         "user-001",
})
if err != nil {
    panic(err)
}
fmt.Printf("状态: %s, 耗时: %.2fs\n", result.Data.Status, result.Data.ElapsedTime)
```

### 知识库检索

```go
kb := client.NewKnowledgeClient(httpClient)
retrieved, err := kb.RetrieveSegments(context.Background(), "dataset-id", client.RetrieveRequest{
    Query: "如何使用 Dify SDK",
    RetrievalModel: client.RetrievalModel{
        TopK: 5,
    },
})
if err != nil {
    panic(err)
}
for _, seg := range retrieved.Segments {
    fmt.Println(seg.Content)
}
```

### 文件上传

```go
file := client.NewFileClient(httpClient, "")
resp, err := file.Upload(context.Background(), client.UploadFileRequest{
    FilePath: "./document.pdf",
    User:     "user-001",
})
```

## API 覆盖

| 模块 | 端点 | 状态 |
|------|------|------|
| **Chat** | 发送消息（blocking / streaming） | ✅ |
| | 停止生成 | ✅ |
| | 获取建议问题 | ✅ |
| | 会话列表 / 历史消息 | ✅ |
| | 消息反馈 | ✅ |
| | 会话管理（重命名 / 删除） | ✅ |
| | 对话变量 | ✅ |
| | 应用信息 / 参数 | ✅ |
| **Workflow** | 执行工作流（blocking / streaming） | ✅ |
| | 按 ID 执行指定版本 | ✅ |
| | 获取执行详情 / 日志 | ✅ |
| | 停止任务 / 恢复事件流 | ✅ |
| **Knowledge** | 知识库 CRUD | ✅ |
| | 文档管理（创建 / 列表 / 删除） | ✅ |
| | 语义检索 | ✅ |
| | 段落管理 | ✅ |
| **File** | 文件上传 | ✅ |

## 项目结构

```
dify-sdk-go/
├── dify-sdk/           # Go SDK（核心库）
│   ├── auth/           # API Key 管理器
│   ├── client/         # HTTP 客户端 & API Client（Chat/Workflow/Knowledge/File）
│   ├── config/         # 环境变量配置
│   ├── di/             # 依赖注入容器
│   └── store/          # 内存元数据存储
├── server/             # HTTP API 服务端（Fiber v3）
│   ├── cmd/server/     # 服务入口
│   ├── internal/       # 内部实现
│   └── pkg/methods/    # 方法注册器
├── examples/           # 使用示例
├── api-specs/          # Dify OpenAPI 规范文件
└── openspec/           # 设计提案 & 变更记录
```

## 配置

通过环境变量或 `.env` 文件配置：

```env
DIFY_BASE_URL=https://api.dify.ai/v1
DIFY_API_KEY=app-xxxxxxxxxxxxx
# 可选：当请求未传 user 时的兜底值。多用户场景请勿依赖此值，应逐请求传真实用户 ID。
# DIFY_DEFAULT_USER=internal-tool
SERVER_PORT=3000
```

## 运行示例

```bash
cd examples
cp .env.example .env
# 编辑 .env 填入你的 Dify API Key
go run main.go
```

服务启动后访问：
- `POST http://localhost:3000/api/v1/chat` — 对话
- `POST http://localhost:3000/api/v1/workflow` — 工作流
- `GET http://localhost:3000/api/v1/datasets` — 知识库列表
- `GET http://localhost:3000/api/v1/info` — 应用信息

## License

MIT
