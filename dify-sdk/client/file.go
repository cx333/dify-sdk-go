package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// FileClient Dify 文件上传与下载客户端。
//
// 覆盖端点：
//   - POST /files/upload — 上传文件（multipart/form-data）
//   - GET /files/{file_id}/preview — 下载/预览文件
type FileClient struct {
	http        *HTTPClient
	user        string // SetUser 设置的值，优先级高于 defaultUser
	defaultUser string // 来自配置 DIFY_DEFAULT_USER，最终兜底值
}

// NewFileClient 创建 FileClient。
//
// defaultUser 来自配置 DIFY_DEFAULT_USER，作为所有请求的最终兜底用户标识。
// 传空字符串表示不使用默认值。
func NewFileClient(http *HTTPClient, defaultUser string) *FileClient {
	return &FileClient{http: http, defaultUser: defaultUser}
}

// SetUser 设置用户标识，优先级高于 defaultUser。
func (c *FileClient) SetUser(user string) {
	c.user = user
}

// resolveUser 按优先级解析用户标识：参数 > SetUser > defaultUser。
func (c *FileClient) resolveUser(paramUser string) string {
	if paramUser != "" {
		return paramUser
	}
	if c.user != "" {
		return c.user
	}
	return c.defaultUser
}

// UploadFile 上传文件到 Dify。
//
// 参数：
//   - filename: 原始文件名
//   - fileData: 文件内容 Reader
//   - user: 用户标识（空则使用 SetUser 设置的值）
//
// 使用 POST /files/upload（multipart/form-data 编码）。
func (c *FileClient) UploadFile(ctx context.Context, filename string, fileData io.Reader, user string) (*FileUploadResponse, error) {
	if user == "" {
		user = c.resolveUser("")
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	// 添加 user 表单字段
	if err := writer.WriteField("user", user); err != nil {
		return nil, fmt.Errorf("file: 写入 user 字段失败: %w", err)
	}

	// 添加文件字段
	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return nil, fmt.Errorf("file: 创建文件表单字段失败: %w", err)
	}
	if _, err := io.Copy(part, fileData); err != nil {
		return nil, fmt.Errorf("file: 写入文件数据失败: %w", err)
	}
	writer.Close()

	req, err := http.NewRequestWithContext(ctx, "POST", c.http.baseURL+"/files/upload", &buf)
	if err != nil {
		return nil, fmt.Errorf("file: 构建上传请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.http.apiKey)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := c.http.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("file: 上传失败: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("file: 读取响应失败: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, NewDifyError(resp.StatusCode, body)
	}

	var result FileUploadResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("file: 解析响应失败: %w", err)
	}
	return &result, nil
}

// DownloadFile 下载文件。
//
// 参数：
//   - fileID: 文件唯一标识
//   - asAttachment: true 强制下载，false 浏览器预览
//
// 返回文件内容和 MIME 类型。
// 使用 GET /files/{file_id}/preview。
func (c *FileClient) DownloadFile(ctx context.Context, fileID string, asAttachment bool) ([]byte, string, error) {
	path := "/files/" + fileID + "/preview"
	if asAttachment {
		path += "?as_attachment=true"
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.http.baseURL+path, nil)
	if err != nil {
		return nil, "", fmt.Errorf("file: 构建下载请求失败: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.http.apiKey)

	resp, err := c.http.client.Do(req)
	if err != nil {
		return nil, "", fmt.Errorf("file: 下载失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return nil, "", NewDifyError(resp.StatusCode, body)
	}

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, "", fmt.Errorf("file: 读取文件内容失败: %w", err)
	}

	mimeType := resp.Header.Get("Content-Type")
	return data, mimeType, nil
}
