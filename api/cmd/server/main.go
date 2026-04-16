package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/redis/go-redis/v9"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"

	"github.com/contful/contful/api/internal/config"
	"github.com/contful/contful/api/internal/middleware"
	"github.com/contful/contful/api/internal/model"
	"github.com/contful/contful/api/internal/repository"
	"github.com/contful/contful/api/internal/service"
)

func main() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(zerolog.ConsoleWriter{Out: log.Writer()}).With().Timestamp().Logger()
	logger.Info().Str("service", "open").Str("port", cfg.Server.Port).Msg("starting")

	// 初始化 PostgreSQL
	db, err := gorm.Open(postgres.Open(cfg.Database.GetDSN()), &gorm.Config{
		Logger: nil, // GORM 日志由 zerolog 统一输出
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect database")
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(cfg.Database.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.Database.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.Database.GetConnMaxLifetime())
	logger.Info().Msg("database connected")

	// 初始化 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := rdb.Ping(ctx).Err(); err != nil {
		logger.Fatal().Err(err).Msg("failed to connect redis")
	}
	logger.Info().Msg("redis connected")

	// 初始化 Repository / Service
	tokenRepo := repository.NewAPITokenRepository(db)
	tokenSvc := service.NewAPITokenService(tokenRepo)
	// 注册 GORM model
	_ = db.AutoMigrate(&repository.APIToken{})

	// 初始化 Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())

	// 全局中间件
	r.Use(middleware.SecurityHeadersMiddleware())
	// CORS 由 API 网关统一处理，应用层不介入

	// Health check（无需认证）
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{
			"status":  "ok",
			"service": "open",
			"version": "1.0.0",
		}))
	})

	// API 路由组
	api := r.Group("/api/v1")
	api.Use(middleware.TokenAuthMiddleware(tokenSvc, logger))

	// 速率限制（从 config.yaml 读取）
	rateLimiter := middleware.NewRateLimiter(rdb)
	if cfg.RateLimit.Enabled {
		api.Use(rateLimiter.RateLimitByToken(cfg.RateLimit.RequestsPerMin))
	}

	// TODO: 内容读写路由将由 M1-009 后续完善
	// 当前占位路由，演示 Token 认证和 Rate Limit 已生效
	api.GET("/content/:slug", middleware.RequireRead(), func(c *gin.Context) {
		c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{
			"message": "M1-015 Token 验证中间件已就绪",
			"slug":   c.Param("slug"),
		}))
	})

	api.POST("/content/:slug", middleware.RequireWrite(), func(c *gin.Context) {
		c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{
			"message": "M1-015 Token 验证中间件已就绪",
			"slug":   c.Param("slug"),
		}))
	})

	// 启动服务
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info().Str("addr", srv.Addr).Msg("server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("server error")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info().Msg("shutting down server...")
	ctx, cancel = context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Fatal().Err(err).Msg("server forced to shutdown")
	}
	logger.Info().Msg("server exited")
}
