package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ChaseRain/img2ppt/internal/api"
	"github.com/ChaseRain/img2ppt/internal/infra/config"
	"github.com/ChaseRain/img2ppt/internal/infra/httpclient"
	"github.com/ChaseRain/img2ppt/internal/infra/limiter"
	"github.com/ChaseRain/img2ppt/internal/infra/logger"
	"github.com/ChaseRain/img2ppt/internal/service/gemini"
	"github.com/ChaseRain/img2ppt/internal/service/imagegen"
	"github.com/ChaseRain/img2ppt/internal/service/orchestrator"
	"github.com/ChaseRain/img2ppt/internal/service/ppt"
	"github.com/ChaseRain/img2ppt/internal/service/storage"
)

func main() {
	// Load config
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Init logger
	zapLogger, err := logger.New(cfg.Log.Level, cfg.Log.Format)
	if err != nil {
		log.Fatalf("failed to init logger: %v", err)
	}
	defer zapLogger.Sync()

	// Init HTTP client
	httpClient := httpclient.New(httpclient.Options{
		Timeout:    time.Duration(cfg.HTTPClient.TimeoutSeconds) * time.Second,
		MaxRetries: cfg.HTTPClient.MaxRetries,
	})

	// Init limiter
	lim := limiter.New(cfg.Limiter.MaxConcurrent, cfg.Limiter.RatePerSecond)

	// Init services
	geminiSvc := gemini.New(cfg.Gemini.APIKey, cfg.Gemini.Model, httpClient, zapLogger)
	imageGenSvc := imagegen.New(cfg.ImageGen.APIKey, cfg.ImageGen.Model, httpClient, zapLogger)
	pptSvc := ppt.New(zapLogger)
	storageSvc := storage.New(cfg.Storage.Type, cfg.Storage.BasePath, cfg.Storage.BaseURL, zapLogger)

	// Init orchestrator
	orch := orchestrator.New(geminiSvc, imageGenSvc, pptSvc, storageSvc, lim, zapLogger)

	// Init router
	router := api.NewRouter(orch, zapLogger)

	// Create server
	srv := &http.Server{
		Addr:         cfg.Server.Addr,
		Handler:      router,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeoutSeconds) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeoutSeconds) * time.Second,
	}

	// Start server
	go func() {
		zapLogger.Info("starting server", "addr", cfg.Server.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			zapLogger.Error("server error", "error", err)
		}
	}()

	// Graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	zapLogger.Info("shutting down server...")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		zapLogger.Error("server forced to shutdown", "error", err)
	}
	zapLogger.Info("server stopped")
}
