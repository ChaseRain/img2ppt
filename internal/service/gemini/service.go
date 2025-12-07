package gemini

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/ChaseRain/img2ppt/internal/infra/httpclient"
	"github.com/ChaseRain/img2ppt/internal/infra/logger"
	"github.com/ChaseRain/img2ppt/pkg/errors"
)

type SlideSpec struct {
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	Bullets     []string `json:"bullets"`
	Notes       string   `json:"notes"`
	ImagePrompt string   `json:"image_prompt"`
	Style       string   `json:"style"`
}

type Service struct {
	apiKey     string
	model      string
	httpClient *httpclient.Client
	logger     *logger.Logger
}

func New(apiKey, model string, client *httpclient.Client, log *logger.Logger) *Service {
	return &Service{
		apiKey:     apiKey,
		model:      model,
		httpClient: client,
		logger:     log,
	}
}

func (s *Service) AnalyzeImage(ctx context.Context, imageBytes []byte, language, style string) (*SlideSpec, error) {
	prompt := s.buildPrompt(language, style)

	imageBase64 := base64.StdEncoding.EncodeToString(imageBytes)
	mimeType := detectMimeType(imageBytes)

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"inline_data": map[string]string{
							"mime_type": mimeType,
							"data":      imageBase64,
						},
					},
					{
						"text": prompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"temperature":     0.7,
			"maxOutputTokens": 2048,
			"responseMimeType": "application/json",
		},
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to marshal request")
	}

	url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/%s:generateContent?key=%s",
		s.model, s.apiKey)

	resp, err := s.httpClient.PostJSON(ctx, url, bodyBytes)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeGeminiAPI, "gemini API request failed")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to read response")
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("gemini API error", "status", resp.StatusCode, "body", string(respBody))
		return nil, errors.New(errors.ErrCodeGeminiAPI, fmt.Sprintf("gemini API returned %d", resp.StatusCode))
	}

	return s.parseResponse(respBody)
}

func (s *Service) buildPrompt(language, style string) string {
	return fmt.Sprintf(`你是 PPT 设计助手。输入是一张图片。
分析图片内容，提取关键信息，输出 JSON：
{
  "title": "简洁有力的标题",
  "subtitle": "副标题（可选）",
  "bullets": ["要点1", "要点2", "要点3"],
  "notes": "演讲者备注",
  "image_prompt": "用于生成插图的描述。禁止出现文字。风格为 %s，16:9，适合作为PPT插图。描述应该与图片内容相关但更加抽象艺术化。",
  "style": "%s"
}
语言：%s。
请确保输出是有效的 JSON 格式。`, style, style, language)
}

func (s *Service) parseResponse(body []byte) (*SlideSpec, error) {
	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to parse gemini response")
	}

	if len(response.Candidates) == 0 || len(response.Candidates[0].Content.Parts) == 0 {
		return nil, errors.New(errors.ErrCodeGeminiAPI, "empty response from gemini")
	}

	text := response.Candidates[0].Content.Parts[0].Text
	text = strings.TrimPrefix(text, "```json")
	text = strings.TrimPrefix(text, "```")
	text = strings.TrimSuffix(text, "```")
	text = strings.TrimSpace(text)

	var spec SlideSpec
	if err := json.Unmarshal([]byte(text), &spec); err != nil {
		s.logger.Error("failed to parse slide spec", "text", text, "error", err)
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to parse slide spec JSON")
	}

	return &spec, nil
}

func detectMimeType(data []byte) string {
	if len(data) < 4 {
		return "application/octet-stream"
	}

	if data[0] == 0xFF && data[1] == 0xD8 {
		return "image/jpeg"
	}
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return "image/png"
	}
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return "image/gif"
	}
	if data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
		return "image/webp"
	}

	return "image/jpeg"
}
