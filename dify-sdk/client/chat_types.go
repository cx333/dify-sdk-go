package client

// 本文件定义 Chat / Chatflow API 的请求与响应数据结构。
// 对应 openapi_chat.json 和 openapi_chatflow.json。

// ChatRequest 发送对话消息的请求体。
// POST /chat-messages
type ChatRequest struct {
	Query          string                 `json:"query"`                      // 用户输入/问题内容
	Inputs         map[string]interface{} `json:"inputs"`                     // 应用输入变量
	ResponseMode   string                 `json:"response_mode,omitempty"`    // streaming 或 blocking
	User           string                 `json:"user"`                       // 用户标识
	ConversationID string                 `json:"conversation_id,omitempty"`  // 会话 ID，空字符串或省略则新建会话
	Files          []InputFile            `json:"files,omitempty"`            // 附加文件
	AutoGenName    *bool                  `json:"auto_generate_name,omitempty"` // 自动生成会话标题
	WorkflowID     string                 `json:"workflow_id,omitempty"`      // 指定工作流版本（advanced-chat）
}

// InputFile 消息中附加的文件。
type InputFile struct {
	Type           string `json:"type"`                      // image / document / audio / video / custom
	TransferMethod string `json:"transfer_method"`            // remote_url 或 local_file
	URL            string `json:"url,omitempty"`              // remote_url 时的文件地址
	UploadFileID   string `json:"upload_file_id,omitempty"`   // local_file 时的已上传文件 ID
}

// ChatCompletionResponse 阻塞模式的对话响应。
type ChatCompletionResponse struct {
	Event          string        `json:"event"`
	TaskID         string        `json:"task_id"`
	ID             string        `json:"id"`
	MessageID      string        `json:"message_id"`
	ConversationID string        `json:"conversation_id"`
	Mode           string        `json:"mode"` // chat / agent-chat / advanced-chat
	Answer         string        `json:"answer"`
	Metadata       *ChatMetadata `json:"metadata,omitempty"`
	CreatedAt      int64         `json:"created_at"`
}

// Thinking 提取回答中 <think>...</think> 标签内的思考过程。
func (r *ChatCompletionResponse) Thinking() string {
	t, _ := SplitThink(r.Answer)
	return t
}

// CleanAnswer 去除 <think> 标签后的纯回复文本。
func (r *ChatCompletionResponse) CleanAnswer() string {
	_, a := SplitThink(r.Answer)
	return a
}

// ChatMetadata 响应的附加元数据（含用量和检索资源）。
type ChatMetadata struct {
	Usage              *Usage              `json:"usage,omitempty"`
	RetrieverResources []RetrieverResource `json:"retriever_resources,omitempty"`
}

// Usage 模型 token 用量与费用信息。
type Usage struct {
	PromptTokens        int     `json:"prompt_tokens"`
	PromptUnitPrice     string  `json:"prompt_unit_price"`
	PromptPriceUnit     string  `json:"prompt_price_unit"`
	PromptPrice         string  `json:"prompt_price"`
	CompletionTokens    int     `json:"completion_tokens"`
	CompletionUnitPrice string  `json:"completion_unit_price"`
	CompletionPriceUnit string  `json:"completion_price_unit"`
	CompletionPrice     string  `json:"completion_price"`
	TotalTokens         int     `json:"total_tokens"`
	TotalPrice          string  `json:"total_price"`
	Currency            string  `json:"currency"`
	Latency             float64 `json:"latency"`
}

// RetrieverResource 响应中引用的知识库检索资源。
type RetrieverResource struct {
	ID             string  `json:"id"`
	MessageID      string  `json:"message_id"`
	Position       int     `json:"position"`
	DatasetID      string  `json:"dataset_id"`
	DatasetName    string  `json:"dataset_name"`
	DocumentID     string  `json:"document_id"`
	DocumentName   string  `json:"document_name"`
	SegmentID      string  `json:"segment_id"`
	Score          float64 `json:"score"`
	Content        string  `json:"content"`
}

// Conversation 会话信息。
type Conversation struct {
	ID           string                 `json:"id"`
	Name         string                 `json:"name"`
	Inputs       map[string]interface{} `json:"inputs"`
	Status       string                 `json:"status"`
	Introduction string                 `json:"introduction"`
	CreatedAt    int64                  `json:"created_at"`
	UpdatedAt    int64                  `json:"updated_at"`
}

// ConversationListResponse 会话列表分页响应。
type ConversationListResponse struct {
	Limit   int            `json:"limit"`
	HasMore bool           `json:"has_more"`
	Data    []Conversation `json:"data"`
}

// Message 会话中的单条消息。
type Message struct {
	ID                string                 `json:"id"`
	ConversationID    string                 `json:"conversation_id"`
	ParentMessageID   string                 `json:"parent_message_id,omitempty"`
	Inputs            map[string]interface{} `json:"inputs"`
	Query             string                 `json:"query"`
	Answer            string                 `json:"answer"`
	Status            string                 `json:"status"`
	Error             string                 `json:"error,omitempty"`
	MessageFiles      []MessageFile          `json:"message_files,omitempty"`
	Feedback          *Feedback              `json:"feedback,omitempty"`
	RetrieverResources []RetrieverResource   `json:"retriever_resources,omitempty"`
	AgentThoughts     []AgentThought         `json:"agent_thoughts,omitempty"`
	CreatedAt         int64                  `json:"created_at"`
}

// MessageFile 消息中附加的文件。
type MessageFile struct {
	ID             string `json:"id"`
	Filename       string `json:"filename"`
	Type           string `json:"type"`
	URL            string `json:"url,omitempty"`
	MimeType       string `json:"mime_type,omitempty"`
	Size           int    `json:"size,omitempty"`
	TransferMethod string `json:"transfer_method"`
	BelongsTo      string `json:"belongs_to,omitempty"` // user 或 assistant
	UploadFileID   string `json:"upload_file_id,omitempty"`
}

// Feedback 消息的用户反馈。
type Feedback struct {
	Rating string `json:"rating"` // like 或 dislike
}

// AgentThought Agent 模式的推理步骤。
type AgentThought struct {
	ID          string   `json:"id"`
	MessageID   string   `json:"message_id"`
	Position    int      `json:"position"`
	Thought     string   `json:"thought"`
	Tool        string   `json:"tool"`
	ToolInput   string   `json:"tool_input"`
	Observation string   `json:"observation"`
	Files       []string `json:"files"`
	CreatedAt   int64    `json:"created_at"`
}

// MessageListResponse 消息历史分页响应。
type MessageListResponse struct {
	Limit   int       `json:"limit"`
	HasMore bool      `json:"has_more"`
	Data    []Message `json:"data"`
}

// FeedbackRequest 提交消息反馈的请求体。
type FeedbackRequest struct {
	Rating  string `json:"rating"`  // like / dislike，省略则撤销
	User    string `json:"user"`
	Content string `json:"content,omitempty"`
}

// ConversationRenameRequest 重命名会话的请求体。
type ConversationRenameRequest struct {
	Name         string `json:"name,omitempty"`
	AutoGenerate bool   `json:"auto_generate"`
	User         string `json:"user"`
}

// AppMode 应用模式类型。
type AppMode string

const (
	ModeChat         AppMode = "chat"          // 聊天助手
	ModeAgentChat    AppMode = "agent-chat"    // 智能体
	ModeAdvancedChat AppMode = "advanced-chat" // 对话流
	ModeWorkflow     AppMode = "workflow"      // 工作流
	ModeCompletion   AppMode = "completion"    // 文本生成
)

// Label 返回模式的中文标签。
func (m AppMode) Label() string {
	switch m {
	case ModeChat:
		return "聊天助手"
	case ModeAgentChat:
		return "智能体"
	case ModeAdvancedChat:
		return "对话流"
	case ModeWorkflow:
		return "工作流"
	case ModeCompletion:
		return "文本生成"
	default:
		return string(m)
	}
}

// AppInfo 应用基本信息。
type AppInfo struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Tags        []string `json:"tags"`
	Mode        string   `json:"mode"` // AppMode 值：chat / agent-chat / advanced-chat / workflow / completion
	AuthorName  string   `json:"author_name,omitempty"`
}

// AppParameters 应用参数与功能开关配置。
type AppParameters struct {
	OpeningStatement                string        `json:"opening_statement,omitempty"`
	SuggestedQuestions              []string      `json:"suggested_questions,omitempty"`
	SuggestedQuestionsAfterAnswer   *FeatureFlag  `json:"suggested_questions_after_answer,omitempty"`
	SpeechToText                    *FeatureFlag  `json:"speech_to_text,omitempty"`
	TextToSpeech                    *TTSConfig    `json:"text_to_speech,omitempty"`
	RetrieverResource               *FeatureFlag  `json:"retriever_resource,omitempty"`
	AnnotationReply                 *FeatureFlag  `json:"annotation_reply,omitempty"`
	MoreLikeThis                    *FeatureFlag  `json:"more_like_this,omitempty"`
	SensitiveWordAvoidance          *FeatureFlag  `json:"sensitive_word_avoidance,omitempty"`
	UserInputForm                   []interface{} `json:"user_input_form,omitempty"`
	FileUpload                      interface{}   `json:"file_upload,omitempty"`
	SystemParameters                interface{}   `json:"system_parameters,omitempty"`
}

// FeatureFlag 功能启/禁用标志。
type FeatureFlag struct {
	Enabled bool `json:"enabled"`
}

// TTSConfig 文字转语音配置。
type TTSConfig struct {
	Enabled  bool   `json:"enabled"`
	Voice    string `json:"voice,omitempty"`
	Language string `json:"language,omitempty"`
	AutoPlay string `json:"autoPlay,omitempty"`
}

// FileUploadResponse 文件上传成功响应。
type FileUploadResponse struct {
	ID             string `json:"id"`
	Name           string `json:"name"`
	Size           int    `json:"size"`
	Extension      string `json:"extension"`
	MimeType       string `json:"mime_type"`
	CreatedBy      string `json:"created_by"`
	CreatedAt      int64  `json:"created_at"`
	PreviewURL     string `json:"preview_url,omitempty"`
	SourceURL      string `json:"source_url,omitempty"`
	OriginalURL    string `json:"original_url,omitempty"`
	UserID         string `json:"user_id,omitempty"`
	TenantID       string `json:"tenant_id,omitempty"`
	ConversationID string `json:"conversation_id,omitempty"`
	FileKey        string `json:"file_key,omitempty"`
}
