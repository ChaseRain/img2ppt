package orchestrator

import (
	"context"

	"github.com/ChaseRain/img2ppt/internal/infra/limiter"
	"github.com/ChaseRain/img2ppt/internal/infra/logger"
	"github.com/ChaseRain/img2ppt/internal/service/gemini"
	"github.com/ChaseRain/img2ppt/internal/service/imagegen"
	"github.com/ChaseRain/img2ppt/internal/service/ppt"
	"github.com/ChaseRain/img2ppt/internal/service/storage"
	"github.com/ChaseRain/img2ppt/pkg/errors"
)

type GeneratePPTRequest struct {
	RequestID  string
	ImageBytes []byte
	Language   string
	Style      string
}

type GeneratePPTResponse struct {
	RequestID string
	PPTURL    string
	Title     string
}

// ProgressEvent 进度事件
type ProgressEvent struct {
	Stage    string
	Message  string
	Progress int
	Data     interface{}
}

// SlideSpecData 分析结果数据
type SlideSpecData struct {
	Title       string   `json:"title"`
	Subtitle    string   `json:"subtitle"`
	Bullets     []string `json:"bullets"`
	ImagePrompt string   `json:"image_prompt"`
}

// ProgressCallback 进度回调函数
type ProgressCallback func(event ProgressEvent)

type Orchestrator struct {
	geminiSvc   *gemini.Service
	imageGenSvc *imagegen.Service
	pptSvc      *ppt.Service
	storageSvc  *storage.Service
	limiter     *limiter.Limiter
	logger      *logger.Logger
}

func New(
	geminiSvc *gemini.Service,
	imageGenSvc *imagegen.Service,
	pptSvc *ppt.Service,
	storageSvc *storage.Service,
	lim *limiter.Limiter,
	log *logger.Logger,
) *Orchestrator {
	return &Orchestrator{
		geminiSvc:   geminiSvc,
		imageGenSvc: imageGenSvc,
		pptSvc:      pptSvc,
		storageSvc:  storageSvc,
		limiter:     lim,
		logger:      log,
	}
}

// GenerateSingleSlidePPT 同步生成（保持兼容）
func (o *Orchestrator) GenerateSingleSlidePPT(ctx context.Context, req *GeneratePPTRequest) (*GeneratePPTResponse, error) {
	return o.GenerateSingleSlidePPTWithProgress(ctx, req, nil)
}

// GenerateSingleSlidePPTWithProgress 带进度回调的生成
func (o *Orchestrator) GenerateSingleSlidePPTWithProgress(ctx context.Context, req *GeneratePPTRequest, onProgress ProgressCallback) (*GeneratePPTResponse, error) {
	release, err := o.limiter.Acquire(ctx)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrCodeRateLimited, "rate limit exceeded")
	}
	defer release()

	emit := func(stage, message string, progress int, data interface{}) {
		if onProgress != nil {
			onProgress(ProgressEvent{
				Stage:    stage,
				Message:  message,
				Progress: progress,
				Data:     data,
			})
		}
	}

	o.logger.Info("starting PPT generation",
		"request_id", req.RequestID,
		"language", req.Language,
		"style", req.Style,
	)

	// Step 1: Analyze image with Gemini
	emit("analyzing", "正在分析图片内容...", 10, nil)

	slideSpec, err := o.geminiSvc.AnalyzeImage(ctx, req.ImageBytes, req.Language, req.Style)
	if err != nil {
		o.logger.Error("failed to analyze image", "request_id", req.RequestID, "error", err)
		return nil, err
	}

	emit("analyzed", "内容分析完成", 40, SlideSpecData{
		Title:       slideSpec.Title,
		Subtitle:    slideSpec.Subtitle,
		Bullets:     slideSpec.Bullets,
		ImagePrompt: slideSpec.ImagePrompt,
	})

	o.logger.Info("image analysis completed",
		"request_id", req.RequestID,
		"title", slideSpec.Title,
		"bullets_count", len(slideSpec.Bullets),
	)

	// Step 2: Generate slide image
	emit("generating", "正在生成配图...", 50, map[string]string{
		"image_prompt": slideSpec.ImagePrompt,
	})

	genImg, err := o.imageGenSvc.GenerateSlideImage(ctx, slideSpec.ImagePrompt, req.ImageBytes, req.Style)
	if err != nil {
		o.logger.Warn("failed to generate image, continuing without image",
			"request_id", req.RequestID,
			"error", err,
		)
		genImg = nil
		emit("generated", "配图生成跳过（将使用默认样式）", 70, nil)
	} else {
		emit("generated", "配图生成完成", 70, nil)
		o.logger.Info("slide image generated", "request_id", req.RequestID)
	}

	// Step 3: Render PPT
	emit("rendering", "正在渲染 PPT...", 80, nil)

	pptBytes, err := o.pptSvc.RenderSingleSlide(slideSpec, genImg)
	if err != nil {
		o.logger.Error("failed to render PPT", "request_id", req.RequestID, "error", err)
		return nil, err
	}
	o.logger.Info("PPT rendered", "request_id", req.RequestID, "size_bytes", len(pptBytes))

	// Step 4: Save to storage
	emit("rendering", "正在保存文件...", 90, nil)

	url, err := o.storageSvc.SavePPT(ctx, req.RequestID, pptBytes)
	if err != nil {
		o.logger.Error("failed to save PPT", "request_id", req.RequestID, "error", err)
		return nil, err
	}

	emit("complete", "生成完成！", 100, map[string]string{
		"ppt_url": url,
		"title":   slideSpec.Title,
	})

	o.logger.Info("PPT saved successfully",
		"request_id", req.RequestID,
		"url", url,
	)

	return &GeneratePPTResponse{
		RequestID: req.RequestID,
		PPTURL:    url,
		Title:     slideSpec.Title,
	}, nil
}
