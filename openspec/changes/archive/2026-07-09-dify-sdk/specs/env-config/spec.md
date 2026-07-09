## ADDED Requirements

### Requirement: .env 文件加载
配置模块 SHALL 通过 `godotenv` 从 `.env` 文件加载环境变量，禁止读取 `config.yaml` 或任何 YAML 文件。

#### Scenario: 加载 .env 配置
- **WHEN** 调用 `config.Load(".env")`
- **THEN** 读取 `.env` 文件中的键值对
- **AND** 返回 `*Config` 结构体
- **AND** 不读取任何 YAML 文件

### Requirement: 结构化 Config 结构体
Config 结构体 SHALL 包含 `BaseURL`、`APIKeys`、`Timeout`、`MaxRetries` 字段，支持从环境变量自动解析，包括默认值。

#### Scenario: 解析完整配置
- **WHEN** `.env` 文件包含 `DIFY_BASE_URL=http://localhost:5001/v1` 和 `DIFY_API_KEYS=app-1,app-2`
- **THEN** `Config.BaseURL` SHALL 等于 `"http://localhost:5001/v1"`
- **AND** `Config.APIKeys` SHALL 等于 `[]string{"app-1", "app-2"}`
- **AND** `Config.Timeout` SHALL 默认为 30 秒

### Requirement: 多模块独立配置
每个模块（`dify-sdk`、`examples`、`server`）SHALL 拥有独立的 `.env` 文件，配置加载 SHALL 支持指定文件路径。

#### Scenario: 独立配置加载
- **WHEN** `examples` 模块调用 `config.Load("examples/.env")`
- **THEN** 仅读取 `examples/.env` 文件
- **AND** 不影响其他模块的配置
