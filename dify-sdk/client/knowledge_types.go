package client

// 本文件定义 Knowledge（知识库）API 的请求与响应数据结构。
// 对应 openapi_knowledge.json。

// ---- 数据集 ----

// CreateDatasetRequest 创建知识库的请求体。
type CreateDatasetRequest struct {
	Name                   string `json:"name"`
	Description            string `json:"description,omitempty"`
	IndexingTechnique      string `json:"indexing_technique,omitempty"`      // high_quality 或 economy
	Permission             string `json:"permission,omitempty"`              // only_me / all_team_members / partial_members
	Provider               string `json:"provider,omitempty"`                // vendor 或 external
	EmbeddingModel         string `json:"embedding_model,omitempty"`
	EmbeddingModelProvider string `json:"embedding_model_provider,omitempty"`
}

// UpdateDatasetRequest 更新知识库的请求体。
type UpdateDatasetRequest struct {
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
	Permission  string `json:"permission,omitempty"`
}

// Dataset 知识库信息。
type Dataset struct {
	ID                     string `json:"id"`
	Name                   string `json:"name"`
	Description            string `json:"description,omitempty"`
	Provider               string `json:"provider"`
	Permission             string `json:"permission"`
	DataSourceType         string `json:"data_source_type,omitempty"`
	IndexingTechnique      string `json:"indexing_technique,omitempty"`
	AppCount               int    `json:"app_count"`
	DocumentCount          int    `json:"document_count"`
	WordCount              int    `json:"word_count"`
	CreatedBy              string `json:"created_by"`
	CreatedAt              int64  `json:"created_at"`
	UpdatedBy              string `json:"updated_by"`
	UpdatedAt              int64  `json:"updated_at"`
	EmbeddingModel         string `json:"embedding_model,omitempty"`
	EmbeddingModelProvider string `json:"embedding_model_provider,omitempty"`
	EmbeddingAvailable     bool   `json:"embedding_available,omitempty"`
}

// DatasetListResponse 知识库列表分页响应。
type DatasetListResponse struct {
	Data    []Dataset `json:"data"`
	HasMore bool      `json:"has_more"`
	Limit   int       `json:"limit"`
	Total   int       `json:"total"`
	Page    int       `json:"page"`
}

// ---- 文档 ----

// CreateDocByTextRequest 从文本创建文档的请求体。
type CreateDocByTextRequest struct {
	Name              string `json:"name"`
	Text              string `json:"text"`
	IndexingTechnique string `json:"indexing_technique,omitempty"` // high_quality 或 economy
	DocForm           string `json:"doc_form,omitempty"`            // text_model / hierarchical_model / qa_model
	DocLanguage       string `json:"doc_language,omitempty"`        // 文档语言，QA 模式重要
}

// Document 知识库中的文档。
type Document struct {
	ID             string `json:"id"`
	Position       int    `json:"position"`
	DataSourceType string `json:"data_source_type"`
	Name           string `json:"name"`
	CreatedFrom    string `json:"created_from"`
	CreatedBy      string `json:"created_by"`
	CreatedAt      int64  `json:"created_at"`
	Tokens         int    `json:"tokens"`
	IndexingStatus string `json:"indexing_status"`
	Enabled        bool   `json:"enabled"`
	Archived       bool   `json:"archived"`
	DisplayStatus  string `json:"display_status"`
	WordCount      int    `json:"word_count"`
	HitCount       int    `json:"hit_count"`
	DocForm        string `json:"doc_form"`
}

// DocumentCreationResponse 文档创建响应，含索引进度批次号。
type DocumentCreationResponse struct {
	Document *Document `json:"document"`
	Batch    string    `json:"batch"` // 用于查询索引进度
}

// DocumentListResponse 文档列表分页响应。
type DocumentListResponse struct {
	Data    []Document `json:"data"`
	HasMore bool       `json:"has_more"`
	Limit   int        `json:"limit"`
	Total   int        `json:"total"`
	Page    int        `json:"page"`
}

// ---- 段落 ----

// SegmentInput 要创建的段落。
type SegmentInput struct {
	Content  string   `json:"content"`
	Answer   string   `json:"answer,omitempty"`   // QA 模式的答案
	Keywords []string `json:"keywords,omitempty"` // 关联关键词
}

// Segment 文档中的段落（块）。
type Segment struct {
	ID         string   `json:"id"`
	Position   int      `json:"position"`
	DocumentID string   `json:"document_id"`
	Content    string   `json:"content"`
	Answer     string   `json:"answer,omitempty"`
	WordCount  int      `json:"word_count"`
	Tokens     int      `json:"tokens"`
	Keywords   []string `json:"keywords,omitempty"`
	HitCount   int      `json:"hit_count"`
	Status     string   `json:"status"`
	CreatedAt  int64    `json:"created_at"`
}

// SegmentListResponse 段落列表响应。
type SegmentListResponse struct {
	Data    []Segment `json:"data"`
	DocForm string    `json:"doc_form,omitempty"`
}

// SegmentPaginatedResponse 段落分页列表响应。
type SegmentPaginatedResponse struct {
	Data    []Segment `json:"data"`
	HasMore bool      `json:"has_more"`
	Limit   int       `json:"limit"`
	Total   int       `json:"total"`
	Page    int       `json:"page"`
}

// ---- 检索 ----

// RetrieveRequest 知识库检索请求体。
type RetrieveRequest struct {
	Query          string          `json:"query"`
	RetrievalModel *RetrievalModel `json:"retrieval_model,omitempty"`
}

// RetrievalModel 检索配置参数。
type RetrievalModel struct {
	SearchMethod          string  `json:"search_method,omitempty"`          // hybrid_search / semantic_search / full_text_search
	RerankingEnable       bool    `json:"reranking_enable,omitempty"`        // 是否启用重排序
	TopK                  int     `json:"top_k,omitempty"`                   // 返回结果数
	ScoreThresholdEnabled bool    `json:"score_threshold_enabled,omitempty"` // 是否启用分数阈值
	ScoreThreshold        float64 `json:"score_threshold,omitempty"`         // 最低分数阈值
}

// RetrievedSegment 检索到的段落及其相关性分数。
type RetrievedSegment struct {
	Segment Segment `json:"segment"`
	Score   float64 `json:"score"`
}

// RetrieveResponse 检索响应。
type RetrieveResponse struct {
	Query   struct {
		Content string `json:"content"`
	} `json:"query"`
	Records []RetrievedSegment `json:"records"`
}
