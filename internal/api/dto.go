package api

type GeneratePPTRequest struct {
	ImageBase64     string `json:"image_base64" binding:"required"`
	Language        string `json:"language"`
	Style           string `json:"style"`
	Stream          bool   `json:"stream"`
	ClientRequestID string `json:"client_request_id"`
}

type GeneratePPTResponse struct {
	RequestID string            `json:"request_id"`
	Status    string            `json:"status"`
	PPTURL    string            `json:"ppt_url,omitempty"`
	Meta      *GeneratePPTMeta  `json:"meta,omitempty"`
	Error     *GeneratePPTError `json:"error,omitempty"`
}

type GeneratePPTMeta struct {
	Title    string   `json:"title,omitempty"`
	Subtitle string   `json:"subtitle,omitempty"`
	Bullets  []string `json:"bullets,omitempty"`
}

type GeneratePPTError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

// SSE 流式事件类型
type StreamEvent struct {
	Event     string      `json:"event"`
	Data      interface{} `json:"data"`
	RequestID string      `json:"request_id"`
}

// 各阶段事件数据
type EventStart struct {
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type EventAnalyzing struct {
	Message  string `json:"message"`
	Progress int    `json:"progress"`
}

type EventAnalyzed struct {
	Message  string   `json:"message"`
	Title    string   `json:"title"`
	Subtitle string   `json:"subtitle"`
	Bullets  []string `json:"bullets"`
	Progress int      `json:"progress"`
}

type EventGenerating struct {
	Message     string `json:"message"`
	ImagePrompt string `json:"image_prompt"`
	Progress    int    `json:"progress"`
}

type EventGenerated struct {
	Message  string `json:"message"`
	Progress int    `json:"progress"`
}

type EventRendering struct {
	Message  string `json:"message"`
	Progress int    `json:"progress"`
}

type EventComplete struct {
	Message string `json:"message"`
	PPTURL  string `json:"ppt_url"`
	Title   string `json:"title"`
}

type EventError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

const (
	StatusPending   = "PENDING"
	StatusSucceeded = "SUCCEEDED"
	StatusFailed    = "FAILED"

	// SSE 事件类型
	EventTypeStart      = "start"
	EventTypeAnalyzing  = "analyzing"
	EventTypeAnalyzed   = "analyzed"
	EventTypeGenerating = "generating"
	EventTypeGenerated  = "generated"
	EventTypeRendering  = "rendering"
	EventTypeComplete   = "complete"
	EventTypeError      = "error"
)
