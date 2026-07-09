## ADDED Requirements

### Requirement: 发送聊天消息
Chat Client SHALL 支持向 Dify Chat API 发送消息，返回包含回复内容的响应。

#### Scenario: 成功发送消息
- **WHEN** 调用 `ChatClient.SendMessage` 传入 `ChatRequest`
- **THEN** 返回 `*Response[ChatMessage]` 包含 AI 回复内容

### Requirement: 流式对话
Chat Client SHALL 支持流式对话，通过 SSE 返回逐字生成的回复。

#### Scenario: 流式回复
- **WHEN** 调用 `ChatClient.SendMessageStream` 传入 `ChatRequest`
- **THEN** 返回 `<-chan SSEEvent`
- **AND** channel 中 SHALL 包含逐步生成的消息片段

### Requirement: 会话管理
Chat Client SHALL 支持获取会话列表和会话历史消息。

#### Scenario: 获取会话列表
- **WHEN** 调用 `ChatClient.GetConversations` 传入用户标识
- **THEN** 返回该用户的所有会话列表

#### Scenario: 获取历史消息
- **WHEN** 调用 `ChatClient.GetMessages` 传入会话 ID
- **THEN** 返回该会话的所有历史消息

### Requirement: 消息反馈
Chat Client SHALL 支持对单条消息进行点赞/点踩反馈。

#### Scenario: 消息点赞
- **WHEN** 调用 `ChatClient.Feedback` 传入消息 ID 和 `like`
- **THEN** 成功提交反馈，返回 nil 错误
