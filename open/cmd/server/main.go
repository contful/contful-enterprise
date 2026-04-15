package main

import (
	"context"
	"fmt"
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
	"gorm.io/gorm/logger"

	"github.com/contful/contful/open/internal/config"
	"github.com/contful/contful/open/internal/middleware"
	"github.com/contful/contful/open/internal/model"
	"github.com/contful/contful/open/internal/repository"
	"github.com/contful/contful/open/internal/service"
)

func main() {
	// 初始化配置
	cfg := config.Load()

	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(zerolog.ConsoleWriter{Out: log.Writer()}).With().Timestamp().Logger()
	logger.Info().Str("service", "open").Str("port", cfg.Server.Port).Msg("starting")

	// 初始化 PostgreSQL
	dsn := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Database.Host, cfg.Database.Port,
		cfg.Database.User, cfg.Database.Password,
		cfg.Database.Name, cfg.Database.SSLMode,
	)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect database")
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	logger.Info().Msg("database connected")

	// 初始化 Redis
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Redis.Host, cfg.Redis.Port),
		Password: cfg.Redis.Password,
		DB:       0,
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

	// 全局中间件（所有路由都走）
	r.Use(middleware.LoggerMiddleware(&logger))
	r.Use(middleware.SecurityHeadersMiddleware())
	r.Use(middleware.CORSMiddleware())

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
	api.Use(middleware.TokenAuthMiddleware(
		service.NewAPITokenService(tokenSvc),
		logger,
	))

	// 速率限制：100次/分钟/Token（Open API 标准）
	rateLimiter := middleware.NewRateLimiter(rdb)
	api.Use(rateLimiter.RateLimitByToken(100))

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
