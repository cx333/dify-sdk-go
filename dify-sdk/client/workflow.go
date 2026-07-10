package client

import (
	"context"
	"fmt"
)

// WorkflowClient Dify 工作流应用 API 客户端。
//
// 覆盖端点：
//   - POST /workflows/run — 执行工作流
//   - POST /workflows/{workflow_id}/run — 按 ID 执行指定版本
//   - GET /workflows/run/{workflow_run_id} — 获取执行详情
//   - GET /workflows/logs — 查询工作流日志
//   - POST /workflows/tasks/{task_id}/stop — 停止执行
//   - GET /workflow/{workflow_run_id}/events — 流式获取事件
type WorkflowClient struct {
	http        *HTTPClient
	user        string // SetUser 设置的值，优先级高于 defaultUser
	defaultUser string // 来自配置 DIFY_DEFAULT_USER，最终兜底值
}

// NewWorkflowClient 创建 WorkflowClient。
//
// defaultUser 来自配置 DIFY_DEFAULT_USER，作为所有请求的最终兜底用户标识。
// 传空字符串表示不使用默认值。
// 关于 defaultUser 的安全影响，详见 config.Config.DefaultUser 的文档。
func NewWorkflowClient(http *HTTPClient, defaultUser string) *WorkflowClient {
	return &WorkflowClient{http: http, defaultUser: defaultUser}
}

// SetUser 设置用户标识，优先级高于 defaultUser。
func (c *WorkflowClient) SetUser(user string) {
	c.user = user
}

// resolveUser 按优先级解析用户标识：参数 > SetUser > defaultUser。
func (c *WorkflowClient) resolveUser(paramUser string) string {
	if paramUser != "" {
		return paramUser
	}
	if c.user != "" {
		return c.user
	}
	return c.defaultUser
}

// WorkflowRunRequest 执行工作流的请求体。
type WorkflowRunRequest struct {
	Inputs       map[string]any `json:"inputs"`                  // 工作流输入变量
	ResponseMode string         `json:"response_mode,omitempty"` // "streaming" 或 "blocking"
	User         string         `json:"user"`                    // 用户标识
	Files        []InputFile    `json:"files,omitempty"`         // 附加文件
}

// WorkflowResult 工作流执行结果。
type WorkflowResult struct {
	ID          string                 `json:"id"`
	WorkflowID  string                 `json:"workflow_id"`
	Status      string                 `json:"status"` // running / succeeded / failed / stopped / partial-succeeded / paused
	Outputs     map[string]any `json:"outputs,omitempty"`
	Error       string                 `json:"error,omitempty"`
	ElapsedTime float64                `json:"elapsed_time,omitempty"`
	TotalTokens int                    `json:"total_tokens,omitempty"`
	TotalSteps  int                    `json:"total_steps"`
	CreatedAt   int64                  `json:"created_at"`
	FinishedAt  int64                  `json:"finished_at,omitempty"`
}

// WorkflowBlockingResponse 阻塞模式的工作流响应。
type WorkflowBlockingResponse struct {
	TaskID        string          `json:"task_id"`         // 任务 ID，用于停止
	WorkflowRunID string          `json:"workflow_run_id"` // 运行 ID，用于查询
	Data          *WorkflowResult `json:"data"`
}

// WorkflowLogEntry 工作流日志记录。
type WorkflowLogEntry struct {
	ID          string          `json:"id"`
	WorkflowRun *WorkflowResult `json:"workflow_run"`
	CreatedFrom string          `json:"created_from"`
	CreatedAt   int64           `json:"created_at"`
}

// WorkflowLogsResponse 工作流日志分页响应。
type WorkflowLogsResponse struct {
	Page    int                `json:"page"`
	Limit   int                `json:"limit"`
	Total   int                `json:"total"`
	HasMore bool               `json:"has_more"`
	Data    []WorkflowLogEntry `json:"data"`
}

// Run 以阻塞模式执行工作流。
// 使用 POST /workflows/run。
func (c *WorkflowClient) Run(ctx context.Context, req WorkflowRunRequest) (*WorkflowBlockingResponse, error) {
	if req.ResponseMode == "" {
		req.ResponseMode = "blocking"
	}
	if req.User == "" {
		req.User = c.resolveUser(req.User)
	}
	var resp WorkflowBlockingResponse
	if err := c.http.Do(ctx, "POST", "/workflows/run", req, &resp); err != nil {
		return nil, fmt.Errorf("workflow: 执行失败: %w", err)
	}
	return &resp, nil
}

// RunStream 以流式模式执行工作流，返回 SSE 事件通道。
// 使用 POST /workflows/run（response_mode: streaming）。
func (c *WorkflowClient) RunStream(ctx context.Context, req WorkflowRunRequest) (<-chan SseEvent, <-chan error) {
	req.ResponseMode = "streaming"
	if req.User == "" {
		req.User = c.resolveUser(req.User)
	}
	return c.http.Stream(ctx, "POST", "/workflows/run", req)
}

// RunByID 按工作流 ID 执行指定已发布版本。
// 使用 POST /workflows/{workflow_id}/run。
func (c *WorkflowClient) RunByID(ctx context.Context, workflowID string, req WorkflowRunRequest) (*WorkflowBlockingResponse, error) {
	if req.ResponseMode == "" {
		req.ResponseMode = "blocking"
	}
	if req.User == "" {
		req.User = c.resolveUser(req.User)
	}
	var resp WorkflowBlockingResponse
	path := "/workflows/" + workflowID + "/run"
	if err := c.http.Do(ctx, "POST", path, req, &resp); err != nil {
		return nil, fmt.Errorf("workflow: 按 ID 执行失败: %w", err)
	}
	return &resp, nil
}

// GetRunDetail 获取工作流执行详情。
// 使用 GET /workflows/run/{workflow_run_id}。
func (c *WorkflowClient) GetRunDetail(ctx context.Context, workflowRunID string) (*WorkflowResult, error) {
	var resp WorkflowResult
	path := "/workflows/run/" + workflowRunID
	if err := c.http.Do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, fmt.Errorf("workflow: 获取执行详情失败: %w", err)
	}
	return &resp, nil
}

// WorkflowLogsOptions 工作流日志查询过滤条件。
type WorkflowLogsOptions struct {
	Keyword string // 关键词搜索
	Status  string // 状态筛选：succeeded / failed / stopped
	Page    int    // 页码
	Limit   int    // 每页条数（1-100）
}

// GetLogs 查询工作流日志。
// 使用 GET /workflows/logs。
func (c *WorkflowClient) GetLogs(ctx context.Context, opts WorkflowLogsOptions) (*WorkflowLogsResponse, error) {
	path := "/workflows/logs?"
	if opts.Keyword != "" {
		path += "keyword=" + opts.Keyword + "&"
	}
	if opts.Status != "" {
		path += "status=" + opts.Status + "&"
	}
	if opts.Page > 0 {
		path += fmt.Sprintf("page=%d&limit=%d", opts.Page, opts.Limit)
	}
	var resp WorkflowLogsResponse
	if err := c.http.Do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, fmt.Errorf("workflow: 获取日志失败: %w", err)
	}
	return &resp, nil
}

// StopTask 停止正在执行的工作流任务。
// 使用 POST /workflows/tasks/{task_id}/stop。
func (c *WorkflowClient) StopTask(ctx context.Context, taskID, user string) error {
	if user == "" {
		user = c.resolveUser("")
	}
	var result SuccessResult
	path := "/workflows/tasks/" + taskID + "/stop"
	if err := c.http.Do(ctx, "POST", path, map[string]string{"user": user}, &result); err != nil {
		return fmt.Errorf("workflow: 停止任务失败: %w", err)
	}
	return nil
}

// StreamEvents 恢复工作流事件流（用于暂停后重连）。
// 使用 GET /workflow/{workflow_run_id}/events。
func (c *WorkflowClient) StreamEvents(ctx context.Context, workflowRunID, user string, includeStateSnapshot, continueOnPause bool) (<-chan SseEvent, <-chan error) {
	if user == "" {
		user = c.resolveUser("")
	}
	path := fmt.Sprintf("/workflow/%s/events?user=%s", workflowRunID, user)
	if includeStateSnapshot {
		path += "&include_state_snapshot=true"
	}
	if continueOnPause {
		path += "&continue_on_pause=true"
	}
	return c.http.Stream(ctx, "GET", path, nil)
}
