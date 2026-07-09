## ADDED Requirements

### Requirement: 运行工作流
Workflow Client SHALL 支持触发 Dify Workflow 执行，返回执行结果或执行 ID。

#### Scenario: 成功运行工作流
- **WHEN** 调用 `WorkflowClient.Run` 传入 `WorkflowRunRequest`
- **THEN** 返回 `*Response[WorkflowResult]` 包含执行结果

### Requirement: 获取工作流结果
Workflow Client SHALL 支持通过工作流执行 ID 查询执行详情和结果。

#### Scenario: 查询执行结果
- **WHEN** 调用 `WorkflowClient.GetResult` 传入工作流执行 ID
- **THEN** 返回 `*Response[WorkflowResult]` 包含完整执行结果

### Requirement: 停止工作流
Workflow Client SHALL 支持停止正在运行的工作流实例。

#### Scenario: 停止执行
- **WHEN** 调用 `WorkflowClient.Stop` 传入工作流执行 ID
- **THEN** 工作流实例 SHALL 被终止
- **AND** 返回 nil 错误
