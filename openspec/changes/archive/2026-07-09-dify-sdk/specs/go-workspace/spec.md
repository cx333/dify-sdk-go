## ADDED Requirements

### Requirement: Workspace 包含三个模块
Go Workspace SHALL 包含 `dify-sdk`、`examples`、`server` 三个模块，并通过 `go.work` 统一管理。

#### Scenario: Workspace 初始化
- **WHEN** 开发者运行 `go work init`
- **THEN** `go.work` 文件引用 `dify-sdk`、`examples`、`server` 三个模块目录

### Requirement: 模块间依赖关系
`examples` 和 `server` 模块 SHALL 依赖 `dify-sdk` 模块，`dify-sdk` 不依赖其他两个模块。

#### Scenario: 依赖验证
- **WHEN** 检查 `examples/go.mod` 和 `server/go.mod`
- **THEN** 两者均包含 `require github.com/wgl/dify-sdk` 依赖
- **AND** `dify-sdk/go.mod` 不包含 `examples` 或 `server` 依赖

### Requirement: 独立构建
每个模块 SHALL 支持独立构建，通过 `go build` 在各自目录下生成可执行文件或库。

#### Scenario: 构建 SDK 库
- **WHEN** 在 `dify-sdk` 目录运行 `go build ./...`
- **THEN** 编译成功，无依赖缺失错误

#### Scenario: 构建示例服务
- **WHEN** 在 `examples` 目录运行 `go build ./...`
- **THEN** 编译成功，正确链接 `dify-sdk`
