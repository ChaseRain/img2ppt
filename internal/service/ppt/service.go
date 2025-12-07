package ppt

import (
	"github.com/ChaseRain/img2ppt/internal/infra/logger"
	"github.com/ChaseRain/img2ppt/internal/service/gemini"
	"github.com/ChaseRain/img2ppt/internal/service/imagegen"
)

type Service struct {
	logger *logger.Logger
}

func New(log *logger.Logger) *Service {
	return &Service{
		logger: log,
	}
}

// RenderSingleSlide - Mock implementation, just returns the generated image bytes
func (s *Service) RenderSingleSlide(spec *gemini.SlideSpec, img *imagegen.GeneratedImage) ([]byte, error) {
	s.logger.Info("mock PPT render",
		"title", spec.Title,
		"subtitle", spec.Subtitle,
		"bullets", len(spec.Bullets),
		"has_image", img != nil && len(img.Bytes) > 0,
	)

	// Mock: 直接返回生成的图片作为输出，方便测试生图功能
	if img != nil && len(img.Bytes) > 0 {
		return img.Bytes, nil
	}

	// 如果没有图片，返回一个空的占位
	return []byte("mock ppt content"), nil
}
