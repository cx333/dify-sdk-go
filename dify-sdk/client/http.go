package client

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"strings"
	"time"
)

// RetryConfig 控制 HTTPClient 的重试行为。
type RetryConfig struct {
	MaxRetries  int           // 最大重试次数
	BaseDelay   time.Duration // 初始退避延迟
	MaxDelay    time.Duration // 最大退避延迟上限
	StatusCodes []int         // 触发重试的 HTTP 状态码
}

// DefaultRetryConfig 返回标准重试配置：
//   - 最多重试 3 次
//   - 初始延迟 500ms，指数递增至最大 10s
//   - 仅对 429/502/503/504 状态码重试
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:  3,
		BaseDelay:   500 * time.Millisecond,
		MaxDelay:    10 * time.Second,
		StatusCodes: RetryableStatusCodes,
	}
}

// HTTPClient Dify API 的可复用 HTTP 传输层。
//
// 内建能力：
//   - 连接池复用（MaxIdleConns: 100，MaxIdleConnsPerHost: 10）
//   - 请求级超时控制（context.WithTimeout）
//   - 指数退避自动重试
//   - SSE（Server-Sent Events）流式响应解析
//
// 非并发安全：SetAPIKey 修改内部状态，需在请求前设置。
type HTTPClient struct {
	client  *http.Client
	baseURL string
	apiKey  string
	retry   RetryConfig
}

// NewHTTPClient 创建 HTTPClient 实例。
//
// 参数：
//   - baseURL: Dify API 基础地址（例如 "http://localhost:5001/v1"）
//   - apiKey: 当前使用的 API Key（可通过 SetAPIKey 切换）
//   - timeout: 请求超时时间
//   - retry: 重试策略
func NewHTTPClient(baseURL, apiKey string, timeout time.Duration, retry RetryConfig) *HTTPClient {
	return NewHTTPClientWithTransport(baseURL, apiKey, timeout, retry, &http.Transport{
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	})
}

// NewHTTPClientWithTransport 使用自定义 transport 创建 HTTPClient。
// 用于注入限流、代理、TLS 配置等自定义传输层。
func NewHTTPClientWithTransport(baseURL, apiKey string, timeout time.Duration, retry RetryConfig, transport http.RoundTripper) *HTTPClient {
	return &HTTPClient{
		client: &http.Client{
			Timeout:   timeout,
			Transport: transport,
		},
		baseURL: strings.TrimRight(baseURL, "/"),
		apiKey:  apiKey,
		retry:   retry,
	}
}

// SetAPIKey 更换认证使用的 API Key（非并发安全）。
func (c *HTTPClient) SetAPIKey(key string) {
	c.apiKey = key
}

// BaseURL 返回配置的基础 URL。
func (c *HTTPClient) BaseURL() string {
	return c.baseURL
}

// Do 发送 HTTP 请求并解析 JSON 响应。
//
// 请求自动附加：
//   - Authorization: Bearer {apiKey}
//   - Content-Type: application/json
//   - Accept: application/json
//
// 非 2xx 响应返回 *DifyError；网络错误自动重试。
func (c *HTTPClient) Do(ctx context.Context, method, path string, body any, result any) error {
	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("序列化请求体失败: %w", err)
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		return fmt.Errorf("构建请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := c.doWithRetry(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("读取响应失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		return NewDifyError(resp.StatusCode, respBody)
	}

	if result != nil {
		if err := json.Unmarshal(respBody, result); err != nil {
			return fmt.Errorf("解析响应失败: %w", err)
		}
	}
	return nil
}

// Stream 发送 SSE 流式请求，返回事件通道。
//
// 调用方应 range 遍历 eventCh 直至通道关闭；若发生错误，从 errCh 读取一次。
//
// 使用示例：
//
//	events, errs := client.Stream(ctx, "POST", "/chat-messages", req)
//	for ev := range events {
//	    fmt.Println(ev.Answer)
//	}
//	if err := <-errs; err != nil {
//	    log.Fatal(err)
//	}
func (c *HTTPClient) Stream(ctx context.Context, method, path string, body interface{}) (<-chan SseEvent, <-chan error) {
	eventCh := make(chan SseEvent, 100)
	errCh := make(chan error, 1)

	var bodyReader io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			errCh <- fmt.Errorf("序列化请求体失败: %w", err)
			close(eventCh)
			return eventCh, errCh
		}
		bodyReader = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, bodyReader)
	if err != nil {
		errCh <- fmt.Errorf("构建请求失败: %w", err)
		close(eventCh)
		return eventCh, errCh
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "text/event-stream")

	go func() {
		defer close(eventCh)
		defer close(errCh)

		resp, err := c.client.Do(req)
		if err != nil {
			errCh <- fmt.Errorf("SSE 请求失败: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode >= 400 {
			body, _ := io.ReadAll(resp.Body)
			errCh <- NewDifyError(resp.StatusCode, body)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024) // 支持大数据行

		var dataBuf bytes.Buffer
		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				// SSE 空行表示事件结束，解析 dataBuf
				if dataBuf.Len() > 0 {
					var event SseEvent
					if err := json.Unmarshal(dataBuf.Bytes(), &event); err != nil {
						errCh <- fmt.Errorf("SSE 解析失败: %w", err)
						return
					}
					// 忽略 ping 保活事件
					if event.Event != "ping" {
						select {
						case eventCh <- event:
						case <-ctx.Done():
							return
						}
					}
					dataBuf.Reset()
				}
				continue
			}

			// 跳过注释行（以 : 开头）
			if strings.HasPrefix(line, ":") {
				continue
			}

			// 跳过 event: 字段行（类型已嵌入 data: JSON）
			if strings.HasPrefix(line, "event:") {
				continue
			}

			// 提取 data: 前缀后的 JSON 内容
			if strings.HasPrefix(line, "data:") {
				data := strings.TrimPrefix(line, "data:")
				data = strings.TrimSpace(data)
				dataBuf.WriteString(data)
			}
		}

		if err := scanner.Err(); err != nil && !strings.Contains(err.Error(), "context") {
			errCh <- fmt.Errorf("SSE 扫描器异常: %w", err)
		}
	}()

	return eventCh, errCh
}

// doWithRetry 执行带指数退避重试的 HTTP 请求。
// 满足以下条件时重试：
//   - 网络错误（连接拒绝、超时、TLS 握手失败等）
//   - 可重试的 HTTP 状态码（429、502、503、504）
func (c *HTTPClient) doWithRetry(req *http.Request) (*http.Response, error) {
	var lastErr error

	for attempt := 0; attempt <= c.retry.MaxRetries; attempt++ {
		if attempt > 0 {
			delay := calcBackoff(c.retry.BaseDelay, c.retry.MaxDelay, attempt-1)
			select {
			case <-time.After(delay):
			case <-req.Context().Done():
				return nil, req.Context().Err()
			}
		}

		// 每次重试需要重新读取请求体（已被前次发送消耗）
		var bodyBytes []byte
		if req.Body != nil {
			bodyBytes, _ = io.ReadAll(req.Body)
			req.Body.Close()
		}

		if bodyBytes != nil {
			req.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		resp, err := c.client.Do(req)
		if err != nil {
			lastErr = err
			// 非可重试错误直接返回
			if !c.isRetryableError(err) {
				return nil, err
			}
			continue
		}

		// 检查是否需要按状态码重试
		if c.isRetryableStatus(resp.StatusCode) {
			resp.Body.Close()
			lastErr = fmt.Errorf("可重试状态码: %d", resp.StatusCode)
			continue
		}

		return resp, nil
	}

	return nil, fmt.Errorf("超过最大重试次数: %w", lastErr)
}

// isRetryableStatus 判断 HTTP 状态码是否触发重试。
func (c *HTTPClient) isRetryableStatus(code int) bool {
	for _, s := range c.retry.StatusCodes {
		if s == code {
			return true
		}
	}
	return false
}

// isRetryableError 判断网络错误是否可重试。
func (c *HTTPClient) isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	msg := err.Error()
	return strings.Contains(msg, "connection refused") ||
		strings.Contains(msg, "connection reset") ||
		strings.Contains(msg, "timeout") ||
		strings.Contains(msg, "TLS handshake")
}

// calcBackoff 计算指数退避延迟：base * 2^attempt，不超过 max。
func calcBackoff(base, max time.Duration, attempt int) time.Duration {
	d := time.Duration(float64(base) * math.Pow(2, float64(attempt)))
	if d > max {
		d = max
	}
	return d
}
