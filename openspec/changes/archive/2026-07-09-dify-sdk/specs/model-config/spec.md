## ADDED Requirements

### Requirement: 获取模型列表
Model Config Client SHALL 支持查询 Dify 中可用的模型列表。

#### Scenario: 查询可用模型
- **WHEN** 调用 `ModelConfigClient.ListModels`
- **THEN** 返回模型列表，包含模型名称、提供商和类型

### Requirement: 获取模型参数配置
Model Config Client SHALL 支持查询指定模型的参数配置（如温度、最大 token 数等）。

#### Scenario: 查询模型参数
- **WHEN** 调用 `ModelConfigClient.GetParameters` 传入模型名称
- **THEN** 返回该模型的所有可配置参数及默认值
