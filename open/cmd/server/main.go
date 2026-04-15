package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"

	"github.com/contful/contful/open/internal/config"
)

func main() {
	// 初始化配置
	cfg := config.Load()

	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(zerolog.ConsoleWriter{Out: log.Writer()}).With().Timestamp().Logger()
	logger.Info().Str("service", "open").Str("port", cfg.Server.Port).Msg("starting")

	// 初始化 Gin
	r := gin.Default()

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "open",
		})
	})

	// API 路由组
	api := r.Group("/api/v1")
	{
		// TODO: 添加业务路由
	}

	// 启动服务
	addr := ":" + cfg.Server.Port
	logger.Info().Str("addr", addr).Msg("server listening")
	if err := r.Run(addr); err != nil {
		logger.Fatal().Err(err).Msg("failed to start server")
	}
}
