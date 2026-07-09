## ADDED Requirements

### Requirement: 文件上传
File Client SHALL 支持上传文件到 Dify，返回文件 ID 和元数据。

#### Scenario: 上传文本文件
- **WHEN** 调用 `FileClient.Upload` 传入文件内容和文件名
- **THEN** 返回文件 ID、文件类型和上传时间

### Requirement: 获取文件信息
File Client SHALL 支持通过文件 ID 查询已上传文件的元数据。

#### Scenario: 查询文件信息
- **WHEN** 调用 `FileClient.GetInfo` 传入文件 ID
- **THEN** 返回文件名称、类型、大小和上传时间
