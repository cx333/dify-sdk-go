## ADDED Requirements

### Requirement: 数据集 CRUD
Knowledge Client SHALL 支持对 Dataset（知识库）的创建、查询、更新、删除操作。

#### Scenario: 创建数据集
- **WHEN** 调用 `KnowledgeClient.CreateDataset` 传入数据集名称和描述
- **THEN** 返回创建成功的数据集信息

#### Scenario: 删除数据集
- **WHEN** 调用 `KnowledgeClient.DeleteDataset` 传入数据集 ID
- **THEN** 数据集 SHALL 被删除
- **AND** 返回 nil 错误

### Requirement: 文档管理
Knowledge Client SHALL 支持在数据集中添加、更新、删除文档。

#### Scenario: 添加文档
- **WHEN** 调用 `KnowledgeClient.AddDocument` 传入数据集 ID 和文档内容
- **THEN** 文档 SHALL 被添加到指定数据集
- **AND** 返回文档 ID

### Requirement: 段落检索
Knowledge Client SHALL 支持在数据集中执行语义检索，返回匹配段落。

#### Scenario: 检索段落
- **WHEN** 调用 `KnowledgeClient.SearchSegments` 传入数据集 ID 和查询文本
- **THEN** 返回按相关性排序的段落列表
