package client

import (
	"context"
	"fmt"
)

// ChatClient Dify 对话型应用（Chat / Agent-Chat / Chatflow）API 客户端。
//
// 覆盖端点：
//   - POST /chat-messages — 发送消息（支持 blocking 和 streaming）
//   - POST /chat-messages/{task_id}/stop — 停止流式响应
//   - GET /messages/{message_id}/suggested — 获取建议问题
//   - GET /conversations — 获取会话列表
//   - GET /messages — 获取会话历史消息
//   - POST /messages/{message_id}/feedbacks — 提交消息反馈
//   - POST /conversations/{conversation_id}/name — 重命名会话
//   - DELETE /conversations/{conversation_id} — 删除会话
//   - GET /conversations/{conversation_id}/variables — 获取对话变量
//   - GET /info — 获取应用基本信息
//   - GET /parameters — 获取应用参数配置
type ChatClient struct {
	http        *HTTPClient
	user        string // SetUser 设置的值，优先级高于 defaultUser
	defaultUser string // 来自配置 DIFY_DEFAULT_USER，最终兜底值
}

// NewChatClient 创建 ChatClient。
//
// defaultUser 来自配置 DIFY_DEFAULT_USER，作为所有请求的最终兜底用户标识。
// 传空字符串表示不使用默认值。
// 关于 defaultUser 的安全影响，详见 config.Config.DefaultUser 的文档。
func NewChatClient(http *HTTPClient, defaultUser string) *ChatClient {
	return &ChatClient{http: http, defaultUser: defaultUser}
}

// SetUser 设置用户标识，优先级高于 defaultUser。
// 未单独指定 user 参数的方法将使用此值。
// 同时设置了 SetUser 和 DIFY_DEFAULT_USER 时，SetUser 优先。
func (c *ChatClient) SetUser(user string) {
	c.user = user
}

// resolveUser 按优先级解析用户标识：参数 > SetUser > defaultUser。
func (c *ChatClient) resolveUser(paramUser string) string {
	if paramUser != "" {
		return paramUser
	}
	if c.user != "" {
		return c.user
	}
	return c.defaultUser
}

// SendMessage 以阻塞模式发送对话消息。
// 使用 POST /chat-messages，response_mode 默认 "blocking"。
func (c *ChatClient) SendMessage(ctx context.Context, req ChatRequest) (*ChatCompletionResponse, error) {
	if req.ResponseMode == "" {
		req.ResponseMode = "blocking"
	}
	if req.User == "" {
		req.User = c.resolveUser(req.User)
	}
	var resp ChatCompletionResponse
	if err := c.http.Do(ctx, "POST", "/chat-messages", req, &resp); err != nil {
		return nil, fmt.Errorf("chat: 发送消息失败: %w", err)
	}
	return &resp, nil
}

// SendMessageStream 以流式模式发送对话消息，返回 SSE 事件通道。
// 使用 POST /chat-messages，response_mode "streaming"。
func (c *ChatClient) SendMessageStream(ctx context.Context, req ChatRequest) (<-chan SseEvent, <-chan error) {
	req.ResponseMode = "streaming"
	if req.User == "" {
		req.User = c.resolveUser(req.User)
	}
	return c.http.Stream(ctx, "POST", "/chat-messages", req)
}

// StopGeneration 停止正在进行的流式响应。
// 使用 POST /chat-messages/{task_id}/stop。
func (c *ChatClient) StopGeneration(ctx context.Context, taskID, user string) error {
	if user == "" {
		user = c.resolveUser("")
	}
	var result SuccessResult
	if err := c.http.Do(ctx, "POST", "/chat-messages/"+taskID+"/stop", map[string]string{"user": user}, &result); err != nil {
		return fmt.Errorf("chat: 停止生成失败: %w", err)
	}
	return nil
}

// GetSuggestedQuestions 获取某条消息的推荐后续问题。
// 使用 GET /messages/{message_id}/suggested。
func (c *ChatClient) GetSuggestedQuestions(ctx context.Context, messageID, user string) ([]string, error) {
	if user == "" {
		user = c.resolveUser("")
	}
	path := fmt.Sprintf("/messages/%s/suggested?user=%s", messageID, user)
	var resp struct {
		Result string   `json:"result"`
		Data   []string `json:"data"`
	}
	if err := c.http.Do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, fmt.Errorf("chat: 获取建议问题失败: %w", err)
	}
	return resp.Data, nil
}

// GetConversations 获取用户的会话列表。
// 使用 GET /conversations。
func (c *ChatClient) GetConversations(ctx context.Context, user, lastID string, limit int) (*ConversationListResponse, error) {
	if user == "" {
		user = c.resolveUser("")
	}
	path := fmt.Sprintf("/conversations?user=%s", user)
	if lastID != "" {
		path += "&last_id=" + lastID
	}
	if limit > 0 {
		path += fmt.Sprintf("&limit=%d", limit)
	}
	var resp ConversationListResponse
	if err := c.http.Do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, fmt.Errorf("chat: 获取会话列表失败: %w", err)
	}
	return &resp, nil
}

// GetMessages 获取会话的历史消息（滚动分页，最新在前）。
// 使用 GET /messages。
func (c *ChatClient) GetMessages(ctx context.Context, conversationID, user, firstID string, limit int) (*MessageListResponse, error) {
	if user == "" {
		user = c.resolveUser("")
	}
	path := fmt.Sprintf("/messages?conversation_id=%s&user=%s", conversationID, user)
	if firstID != "" {
		path += "&first_id=" + firstID
	}
	if limit > 0 {
		path += fmt.Sprintf("&limit=%d", limit)
	}
	var resp MessageListResponse
	if err := c.http.Do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, fmt.Errorf("chat: 获取消息历史失败: %w", err)
	}
	return &resp, nil
}

// Feedback 提交消息反馈（like/dislike）。
// 使用 POST /messages/{message_id}/feedbacks。
func (c *ChatClient) Feedback(ctx context.Context, messageID string, req FeedbackRequest) error {
	if req.User == "" {
		req.User = c.resolveUser(req.User)
	}
	var result SuccessResult
	if err := c.http.Do(ctx, "POST", "/messages/"+messageID+"/feedbacks", req, &result); err != nil {
		return fmt.Errorf("chat: 提交反馈失败: %w", err)
	}
	return nil
}

// RenameConversation 重命名会话。
// 使用 POST /conversations/{conversation_id}/name。
func (c *ChatClient) RenameConversation(ctx context.Context, conversationID string, req ConversationRenameRequest) (*Conversation, error) {
	if req.User == "" {
		req.User = c.resolveUser(req.User)
	}
	var resp Conversation
	if err := c.http.Do(ctx, "POST", "/conversations/"+conversationID+"/name", req, &resp); err != nil {
		return nil, fmt.Errorf("chat: 重命名会话失败: %w", err)
	}
	return &resp, nil
}

// DeleteConversation 删除会话。
// 使用 DELETE /conversations/{conversation_id}。
func (c *ChatClient) DeleteConversation(ctx context.Context, conversationID, user string) error {
	if user == "" {
		user = c.resolveUser("")
	}
	var body map[string]string
	if user != "" {
		body = map[string]string{"user": user}
	}
	if err := c.http.Do(ctx, "DELETE", "/conversations/"+conversationID, body, nil); err != nil {
		return fmt.Errorf("chat: 删除会话失败: %w", err)
	}
	return nil
}

// GetConversationVariables 获取会话变量。
// 使用 GET /conversations/{conversation_id}/variables。
func (c *ChatClient) GetConversationVariables(ctx context.Context, conversationID, user string) ([]interface{}, error) {
	if user == "" {
		user = c.resolveUser("")
	}
	path := fmt.Sprintf("/conversations/%s/variables?user=%s", conversationID, user)
	var resp struct {
		Data []interface{} `json:"data"`
	}
	if err := c.http.Do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, fmt.Errorf("chat: 获取对话变量失败: %w", err)
	}
	return resp.Data, nil
}

// GetAppInfo 获取应用基本信息（名称、描述、模式等）。
// 使用 GET /info。
func (c *ChatClient) GetAppInfo(ctx context.Context) (*AppInfo, error) {
	var resp AppInfo
	if err := c.http.Do(ctx, "GET", "/info", nil, &resp); err != nil {
		return nil, fmt.Errorf("chat: 获取应用信息失败: %w", err)
	}
	return &resp, nil
}

// GetAppParameters 获取应用参数配置（输入表单、功能开关等）。
// 使用 GET /parameters。
func (c *ChatClient) GetAppParameters(ctx context.Context) (*AppParameters, error) {
	var resp AppParameters
	if err := c.http.Do(ctx, "GET", "/parameters", nil, &resp); err != nil {
		return nil, fmt.Errorf("chat: 获取应用参数失败: %w", err)
	}
	return &resp, nil
}
