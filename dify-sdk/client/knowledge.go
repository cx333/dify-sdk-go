package client

import (
	"context"
	"fmt"
)

// KnowledgeClient Dify 知识库 API 客户端。
//
// 覆盖端点：
//   - POST /datasets — 创建空知识库
//   - GET /datasets — 获取知识库列表
//   - GET /datasets/{id} — 获取知识库详情
//   - PATCH /datasets/{id} — 更新知识库
//   - DELETE /datasets/{id} — 删除知识库
//   - POST /datasets/{id}/document/create-by-text — 从文本创建文档
//   - GET /datasets/{id}/documents — 获取文档列表
//   - GET /datasets/{id}/documents/{doc_id} — 获取文档详情
//   - DELETE /datasets/{id}/documents/{doc_id} — 删除文档
//   - POST /datasets/{id}/retrieve — 检索段落
//   - POST/GET/DELETE /datasets/{id}/documents/{doc_id}/segments — 段落管理
type KnowledgeClient struct {
	http *HTTPClient
}

// NewKnowledgeClient 创建 KnowledgeClient。
func NewKnowledgeClient(http *HTTPClient) *KnowledgeClient {
	return &KnowledgeClient{http: http}
}

// ---- 数据集 CRUD ----

// CreateDataset 创建一个空知识库。
// 使用 POST /datasets。
func (c *KnowledgeClient) CreateDataset(ctx context.Context, req CreateDatasetRequest) (*Dataset, error) {
	var resp Dataset
	if err := c.http.Do(ctx, "POST", "/datasets", req, &resp); err != nil {
		return nil, fmt.Errorf("knowledge: create dataset failed: %w", err)
	}
	return &resp, nil
}

// ListDatasets 获取知识库分页列表。
// 使用 GET /datasets。
func (c *KnowledgeClient) ListDatasets(ctx context.Context, page, limit int) (*DatasetListResponse, error) {
	path := fmt.Sprintf("/datasets?page=%d&limit=%d", page, limit)
	var resp DatasetListResponse
	if err := c.http.Do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, fmt.Errorf("knowledge: list datasets failed: %w", err)
	}
	return &resp, nil
}

// GetDataset 获取知识库详情。
// 使用 GET /datasets/{dataset_id}。
func (c *KnowledgeClient) GetDataset(ctx context.Context, datasetID string) (*Dataset, error) {
	var resp Dataset
	if err := c.http.Do(ctx, "GET", "/datasets/"+datasetID, nil, &resp); err != nil {
		return nil, fmt.Errorf("knowledge: get dataset failed: %w", err)
	}
	return &resp, nil
}

// UpdateDataset 更新知识库配置。
// 使用 PATCH /datasets/{dataset_id}。
func (c *KnowledgeClient) UpdateDataset(ctx context.Context, datasetID string, req UpdateDatasetRequest) (*Dataset, error) {
	var resp Dataset
	if err := c.http.Do(ctx, "PATCH", "/datasets/"+datasetID, req, &resp); err != nil {
		return nil, fmt.Errorf("knowledge: update dataset failed: %w", err)
	}
	return &resp, nil
}

// DeleteDataset 删除知识库及其所有文档。
// 使用 DELETE /datasets/{dataset_id}。
func (c *KnowledgeClient) DeleteDataset(ctx context.Context, datasetID string) error {
	if err := c.http.Do(ctx, "DELETE", "/datasets/"+datasetID, nil, nil); err != nil {
		return fmt.Errorf("knowledge: delete dataset failed: %w", err)
	}
	return nil
}

// ---- 文档操作 ----

// CreateDocumentFromText 从文本内容创建文档。
// 使用 POST /datasets/{dataset_id}/document/create-by-text。
func (c *KnowledgeClient) CreateDocumentFromText(ctx context.Context, datasetID string, req CreateDocByTextRequest) (*DocumentCreationResponse, error) {
	var resp DocumentCreationResponse
	path := "/datasets/" + datasetID + "/document/create-by-text"
	if err := c.http.Do(ctx, "POST", path, req, &resp); err != nil {
		return nil, fmt.Errorf("knowledge: create document from text failed: %w", err)
	}
	return &resp, nil
}

// ListDocuments 获取知识库中的文档列表。
// 使用 GET /datasets/{dataset_id}/documents。
func (c *KnowledgeClient) ListDocuments(ctx context.Context, datasetID string, page, limit int) (*DocumentListResponse, error) {
	path := fmt.Sprintf("/datasets/%s/documents?page=%d&limit=%d", datasetID, page, limit)
	var resp DocumentListResponse
	if err := c.http.Do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, fmt.Errorf("knowledge: list documents failed: %w", err)
	}
	return &resp, nil
}

// GetDocument 获取文档详情。
// 使用 GET /datasets/{dataset_id}/documents/{document_id}。
func (c *KnowledgeClient) GetDocument(ctx context.Context, datasetID, documentID string) (*Document, error) {
	var resp Document
	path := "/datasets/" + datasetID + "/documents/" + documentID
	if err := c.http.Do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, fmt.Errorf("knowledge: get document failed: %w", err)
	}
	return &resp, nil
}

// DeleteDocument 从知识库中删除文档。
// 使用 DELETE /datasets/{dataset_id}/documents/{document_id}。
func (c *KnowledgeClient) DeleteDocument(ctx context.Context, datasetID, documentID string) error {
	path := "/datasets/" + datasetID + "/documents/" + documentID
	if err := c.http.Do(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("knowledge: delete document failed: %w", err)
	}
	return nil
}

// ---- 检索 ----

// RetrieveSegments 在知识库中执行语义检索。
// 使用 POST /datasets/{dataset_id}/retrieve。
func (c *KnowledgeClient) RetrieveSegments(ctx context.Context, datasetID string, req RetrieveRequest) (*RetrieveResponse, error) {
	var resp RetrieveResponse
	path := "/datasets/" + datasetID + "/retrieve"
	if err := c.http.Do(ctx, "POST", path, req, &resp); err != nil {
		return nil, fmt.Errorf("knowledge: retrieve failed: %w", err)
	}
	return &resp, nil
}

// ---- 段落管理 ----

// CreateSegments 向文档添加段落。
// 使用 POST /datasets/{dataset_id}/documents/{document_id}/segments。
func (c *KnowledgeClient) CreateSegments(ctx context.Context, datasetID, documentID string, segments []SegmentInput) (*SegmentListResponse, error) {
	var resp SegmentListResponse
	path := "/datasets/" + datasetID + "/documents/" + documentID + "/segments"
	req := map[string]any{"segments": segments}
	if err := c.http.Do(ctx, "POST", path, req, &resp); err != nil {
		return nil, fmt.Errorf("knowledge: create segments failed: %w", err)
	}
	return &resp, nil
}

// ListSegments 获取文档的段落列表。
// 使用 GET /datasets/{dataset_id}/documents/{document_id}/segments。
func (c *KnowledgeClient) ListSegments(ctx context.Context, datasetID, documentID string, page, limit int) (*SegmentPaginatedResponse, error) {
	path := fmt.Sprintf("/datasets/%s/documents/%s/segments?page=%d&limit=%d", datasetID, documentID, page, limit)
	var resp SegmentPaginatedResponse
	if err := c.http.Do(ctx, "GET", path, nil, &resp); err != nil {
		return nil, fmt.Errorf("knowledge: list segments failed: %w", err)
	}
	return &resp, nil
}

// DeleteSegment 删除文档中的段落。
// 使用 DELETE /datasets/{dataset_id}/documents/{document_id}/segments/{segment_id}。
func (c *KnowledgeClient) DeleteSegment(ctx context.Context, datasetID, documentID, segmentID string) error {
	path := "/datasets/" + datasetID + "/documents/" + documentID + "/segments/" + segmentID
	if err := c.http.Do(ctx, "DELETE", path, nil, nil); err != nil {
		return fmt.Errorf("knowledge: delete segment failed: %w", err)
	}
	return nil
}
