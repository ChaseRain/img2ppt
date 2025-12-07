package imagegen

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/ChaseRain/img2ppt/internal/infra/httpclient"
	"github.com/ChaseRain/img2ppt/internal/infra/logger"
	"github.com/ChaseRain/img2ppt/pkg/errors"
)

type GeneratedImage struct {
	Bytes []byte
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

func (s *Service) GenerateSlideImage(ctx context.Context, prompt string, refImage []byte, style string) (*GeneratedImage, error) {
	enhancedPrompt := s.buildImagePrompt(prompt, style)

	requestBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{
				"parts": []map[string]interface{}{
					{
						"text": enhancedPrompt,
					},
				},
			},
		},
		"generationConfig": map[string]interface{}{
			"responseModalities": []string{"TEXT", "IMAGE"},
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
		return nil, errors.Wrap(err, errors.ErrCodeImageGenAPI, "image generation API request failed")
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to read response")
	}

	if resp.StatusCode != http.StatusOK {
		s.logger.Error("image gen API error", "status", resp.StatusCode, "body", string(respBody))
		return nil, errors.New(errors.ErrCodeImageGenAPI, fmt.Sprintf("image generation API returned %d", resp.StatusCode))
	}

	return s.parseResponse(respBody)
}

func (s *Service) buildImagePrompt(prompt, style string) string {
	return fmt.Sprintf(`Generate a high-quality illustration for a PowerPoint slide.

Requirements:
- Aspect ratio: 16:9
- Style: %s, professional, clean
- NO text, letters, words, or numbers in the image
- Abstract, artistic interpretation suitable for business presentation
- High contrast, visually striking

Description: %s`, style, prompt)
}

func (s *Service) parseResponse(body []byte) (*GeneratedImage, error) {
	var response struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text       string `json:"text,omitempty"`
					InlineData *struct {
						MimeType string `json:"mimeType"`
						Data     string `json:"data"`
					} `json:"inlineData,omitempty"`
				} `json:"parts"`
			} `json:"content"`
		} `json:"candidates"`
	}

	if err := json.Unmarshal(body, &response); err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to parse image gen response")
	}

	if len(response.Candidates) == 0 {
		return nil, errors.New(errors.ErrCodeImageGenAPI, "empty response from image generation")
	}

	for _, part := range response.Candidates[0].Content.Parts {
		if part.InlineData != nil && part.InlineData.Data != "" {
			imageBytes, err := base64.StdEncoding.DecodeString(part.InlineData.Data)
			if err != nil {
				return nil, errors.Wrap(err, errors.ErrCodeInternal, "failed to decode image data")
			}
			return &GeneratedImage{Bytes: imageBytes}, nil
		}
	}

	return nil, errors.New(errors.ErrCodeImageGenAPI, "no image in response")
}
