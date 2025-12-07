package api

import (
	"github.com/ChaseRain/img2ppt/internal/infra/logger"
	"github.com/ChaseRain/img2ppt/internal/service/orchestrator"
	"github.com/gin-gonic/gin"
)

func NewRouter(orch *orchestrator.Orchestrator, log *logger.Logger) *gin.Engine {
	gin.SetMode(gin.ReleaseMode)

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(requestLogger(log))

	handler := NewHandler(orch, log)

	r.GET("/health", handler.Health)

	v1 := r.Group("/v1")
	{
		v1.POST("/image-to-ppt", handler.GeneratePPT)
	}

	return r
}

func requestLogger(log *logger.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		log.Info("request started",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
		)
		c.Next()
		log.Info("request completed",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status(),
		)
	}
}
