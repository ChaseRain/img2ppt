package api

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/ChaseRain/img2ppt/internal/infra/logger"
	"github.com/ChaseRain/img2ppt/internal/service/orchestrator"
	"github.com/ChaseRain/img2ppt/pkg/errors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type Handler struct {
	orchestrator *orchestrator.Orchestrator
	logger       *logger.Logger
}

func NewHandler(orch *orchestrator.Orchestrator, log *logger.Logger) *Handler {
	return &Handler{
		orchestrator: orch,
		logger:       log,
	}
}

func (h *Handler) GeneratePPT(c *gin.Context) {
	var req GeneratePPTRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		h.logger.Error("invalid request", "error", err)
		c.JSON(http.StatusBadRequest, GeneratePPTResponse{
			Status: StatusFailed,
			Error: &GeneratePPTError{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	requestID := req.ClientRequestID
	if requestID == "" {
		requestID = uuid.New().String()
	}

	if req.Language == "" {
		req.Language = "zh-CN"
	}
	if req.Style == "" {
		req.Style = "consulting_minimal"
	}

	// 处理 Data URL 格式: data:image/png;base64,xxxxx
	imageBase64 := req.ImageBase64
	if strings.Contains(imageBase64, ",") {
		parts := strings.SplitN(imageBase64, ",", 2)
		if len(parts) == 2 {
			imageBase64 = parts[1]
		}
	}

	imageBytes, err := base64.StdEncoding.DecodeString(imageBase64)
	if err != nil {
		h.logger.Error("failed to decode image", "error", err)
		c.JSON(http.StatusBadRequest, GeneratePPTResponse{
			RequestID: requestID,
			Status:    StatusFailed,
			Error: &GeneratePPTError{
				Code:    "INVALID_IMAGE",
				Message: "failed to decode base64 image",
			},
		})
		return
	}

	orchReq := &orchestrator.GeneratePPTRequest{
		RequestID:  requestID,
		ImageBytes: imageBytes,
		Language:   req.Language,
		Style:      req.Style,
	}

	// 流式输出
	if req.Stream {
		h.handleStreamingResponse(c, requestID, orchReq)
		return
	}

	// 非流式输出（保持兼容）
	result, err := h.orchestrator.GenerateSingleSlidePPT(c.Request.Context(), orchReq)
	if err != nil {
		h.handleError(c, requestID, err)
		return
	}

	c.JSON(http.StatusOK, GeneratePPTResponse{
		RequestID: requestID,
		Status:    StatusSucceeded,
		PPTURL:    result.PPTURL,
		Meta: &GeneratePPTMeta{
			Title: result.Title,
		},
	})
}

func (h *Handler) handleStreamingResponse(c *gin.Context, requestID string, req *orchestrator.GeneratePPTRequest) {
	// 设置 SSE headers
	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	// 发送事件的辅助函数
	sendEvent := func(eventType string, data interface{}) {
		event := StreamEvent{
			Event:     eventType,
			Data:      data,
			RequestID: requestID,
		}
		jsonData, _ := json.Marshal(event)
		fmt.Fprintf(c.Writer, "event: %s\n", eventType)
		fmt.Fprintf(c.Writer, "data: %s\n\n", jsonData)
		c.Writer.Flush()
	}

	// 发送开始事件
	sendEvent(EventTypeStart, EventStart{
		Message:   "开始处理您的请求...",
		Timestamp: time.Now().Unix(),
	})

	// 进度回调
	onProgress := func(event orchestrator.ProgressEvent) {
		switch event.Stage {
		case "analyzing":
			sendEvent(EventTypeAnalyzing, EventAnalyzing{
				Message:  event.Message,
				Progress: event.Progress,
			})
		case "analyzed":
			if specData, ok := event.Data.(orchestrator.SlideSpecData); ok {
				sendEvent(EventTypeAnalyzed, EventAnalyzed{
					Message:  event.Message,
					Title:    specData.Title,
					Subtitle: specData.Subtitle,
					Bullets:  specData.Bullets,
					Progress: event.Progress,
				})
			}
		case "generating":
			prompt := ""
			if m, ok := event.Data.(map[string]string); ok {
				prompt = m["image_prompt"]
			}
			sendEvent(EventTypeGenerating, EventGenerating{
				Message:     event.Message,
				ImagePrompt: prompt,
				Progress:    event.Progress,
			})
		case "generated":
			sendEvent(EventTypeGenerated, EventGenerated{
				Message:  event.Message,
				Progress: event.Progress,
			})
		case "rendering":
			sendEvent(EventTypeRendering, EventRendering{
				Message:  event.Message,
				Progress: event.Progress,
			})
		case "complete":
			url := ""
			title := ""
			if m, ok := event.Data.(map[string]string); ok {
				url = m["ppt_url"]
				title = m["title"]
			}
			sendEvent(EventTypeComplete, EventComplete{
				Message: event.Message,
				PPTURL:  url,
				Title:   title,
			})
		}
	}

	// 执行生成
	_, err := h.orchestrator.GenerateSingleSlidePPTWithProgress(c.Request.Context(), req, onProgress)
	if err != nil {
		code := "INTERNAL_ERROR"
		if appErr, ok := err.(*errors.AppError); ok {
			code = appErr.Code
		}
		sendEvent(EventTypeError, EventError{
			Code:    code,
			Message: err.Error(),
		})
	}
}

func (h *Handler) handleError(c *gin.Context, requestID string, err error) {
	h.logger.Error("failed to generate PPT", "error", err, "request_id", requestID)

	code := "INTERNAL_ERROR"
	status := http.StatusInternalServerError

	if appErr, ok := err.(*errors.AppError); ok {
		code = appErr.Code
		if appErr.Code == errors.ErrCodeRateLimited {
			status = http.StatusTooManyRequests
		}
	}

	c.JSON(status, GeneratePPTResponse{
		RequestID: requestID,
		Status:    StatusFailed,
		Error: &GeneratePPTError{
			Code:    code,
			Message: err.Error(),
		},
	})
}

func (h *Handler) Health(c *gin.Context) {
	c.JSON(http.StatusOK, HealthResponse{Status: "ok"})
}
