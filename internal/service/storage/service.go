package storage

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/ChaseRain/img2ppt/internal/infra/logger"
	"github.com/ChaseRain/img2ppt/pkg/errors"
)

type Service struct {
	storageType string
	basePath    string
	baseURL     string
	logger      *logger.Logger
}

func New(storageType, basePath, baseURL string, log *logger.Logger) *Service {
	return &Service{
		storageType: storageType,
		basePath:    basePath,
		baseURL:     baseURL,
		logger:      log,
	}
}

func (s *Service) SavePPT(ctx context.Context, id string, data []byte) (string, error) {
	switch s.storageType {
	case "local":
		return s.saveLocal(id, data)
	case "s3":
		return s.saveS3(ctx, id, data)
	case "gcs":
		return s.saveGCS(ctx, id, data)
	default:
		return s.saveLocal(id, data)
	}
}

func (s *Service) saveLocal(id string, data []byte) (string, error) {
	if err := os.MkdirAll(s.basePath, 0755); err != nil {
		return "", errors.Wrap(err, errors.ErrCodeStorage, "failed to create output directory")
	}

	// 检测文件类型，根据内容决定扩展名
	ext := s.detectExtension(data)
	filename := fmt.Sprintf("%s%s", id, ext)
	filePath := filepath.Join(s.basePath, filename)

	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return "", errors.Wrap(err, errors.ErrCodeStorage, "failed to write file")
	}

	url := fmt.Sprintf("%s/%s", s.baseURL, filename)
	s.logger.Info("saved file locally", "path", filePath, "url", url, "size", len(data))

	return url, nil
}

func (s *Service) detectExtension(data []byte) string {
	if len(data) < 4 {
		return ".bin"
	}
	// PNG
	if data[0] == 0x89 && data[1] == 0x50 && data[2] == 0x4E && data[3] == 0x47 {
		return ".png"
	}
	// JPEG
	if data[0] == 0xFF && data[1] == 0xD8 {
		return ".jpg"
	}
	// GIF
	if data[0] == 0x47 && data[1] == 0x49 && data[2] == 0x46 {
		return ".gif"
	}
	// WebP
	if len(data) >= 12 && data[0] == 0x52 && data[1] == 0x49 && data[2] == 0x46 && data[3] == 0x46 {
		return ".webp"
	}
	// PPTX (ZIP)
	if data[0] == 0x50 && data[1] == 0x4B {
		return ".pptx"
	}
	return ".bin"
}

func (s *Service) saveS3(ctx context.Context, id string, data []byte) (string, error) {
	// TODO: Implement S3 storage
	// This is a placeholder for S3 implementation using AWS SDK
	return "", errors.New(errors.ErrCodeStorage, "S3 storage not implemented")
}

func (s *Service) saveGCS(ctx context.Context, id string, data []byte) (string, error) {
	// TODO: Implement GCS storage
	// This is a placeholder for GCS implementation using Google Cloud SDK
	return "", errors.New(errors.ErrCodeStorage, "GCS storage not implemented")
}

func (s *Service) GetFile(ctx context.Context, id string) ([]byte, error) {
	switch s.storageType {
	case "local":
		return s.getLocal(id)
	default:
		return s.getLocal(id)
	}
}

func (s *Service) getLocal(id string) ([]byte, error) {
	filename := fmt.Sprintf("%s.pptx", id)
	filePath := filepath.Join(s.basePath, filename)

	data, err := os.ReadFile(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errors.New(errors.ErrCodeNotFound, "file not found")
		}
		return nil, errors.Wrap(err, errors.ErrCodeStorage, "failed to read file")
	}

	return data, nil
}
