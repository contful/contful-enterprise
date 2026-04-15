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
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/contful/contful/admin/internal/config"
	"github.com/contful/contful/admin/internal/handler"
	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/contful/contful/admin/internal/service"
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
	logger.Info().
		Str("service", "admin").
		Str("port", cfg.Server.Port).
		Str("mode", cfg.Server.Mode).
		Msg("starting")

	// 设置 Gin 模式
	if cfg.Server.Mode == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	// 初始化数据库
	db, err := initDB(cfg.Database)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect database")
	}

	// 初始化 Redis
	redisClient := redis.NewClient(&redis.Options{
		Addr:     cfg.Redis.GetAddr(),
		Password: cfg.Redis.Password,
		DB:       cfg.Redis.DB,
		PoolSize: cfg.Redis.PoolSize,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		logger.Fatal().Err(err).Msg("failed to connect redis")
	}
	logger.Info().Msg("redis connected")

	// 初始化 Repository
	userRepo := repository.NewUserRepository(db, redisClient)
	auditRepo := repository.NewAuditRepository(db)
	contentTypeRepo := repository.NewContentTypeRepository(db)
	fieldRepo := repository.NewFieldRepository(db)
	entryRepo := repository.NewEntryRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	tokenRepo := repository.NewAPITokenRepository(db)

	// 初始化 Service
	authService := service.NewAuthService(userRepo, auditRepo, cfg.JWT.Secret)
	ctService := service.NewContentTypeService(contentTypeRepo, fieldRepo, logger)
	entryService := service.NewEntryService(entryRepo, contentTypeRepo, fieldRepo)
	assetService := service.NewAssetService(assetRepo, cfg.Storage.UploadDir, "/admin/v1")
	tokenService := service.NewAPITokenService(tokenRepo)

	// 初始化 Handler
	authHandler := handler.NewAuthHandler(authService)
	ctHandler := handler.NewContentTypeHandler(ctService)
	entryHandler := handler.NewEntryHandler(entryService)
	assetHandler := handler.NewAssetHandler(assetService)
	tokenHandler := handler.NewAPITokenHandler(tokenService)

	// 初始化 Gin
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.CORSMiddleware())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":  "ok",
			"service": "admin",
		})
	})

	// Ready check (检查数据库和 Redis)
	r.GET("/ready", func(c *gin.Context) {
		if err := db.Exec("SELECT 1").Error; err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "db not ready"})
			return
		}
		if err := redisClient.Ping(context.Background()).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "redis not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// API 路由组
	api := r.Group("/admin/v1")
	{
		// 公开路由
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
		}

		// 需要认证的路由
		protected := api.Group("")
		protected.Use(middleware.JWTAuth(authService))
		{
			// 认证相关
			protected.POST("/auth/logout", authHandler.Logout)

			// 用户相关
			protected.GET("/users/me", authHandler.Me)
			protected.GET("/users", authHandler.ListUsers)

			// 内容类型管理 (REST: /content/types)
			protected.GET("/content/types", ctHandler.List)
			protected.POST("/content/types", ctHandler.Create)
			protected.GET("/content/types/:id", ctHandler.Get)
			protected.PUT("/content/types/:id", ctHandler.Update)
			protected.DELETE("/content/types/:id", ctHandler.Delete)
			protected.POST("/content/types/:id/fields", ctHandler.CreateField)
			protected.GET("/content/types/:id/fields", ctHandler.ListFields)
			protected.PUT("/content/types/fields/:fieldId", ctHandler.UpdateField)
			protected.DELETE("/content/types/fields/:fieldId", ctHandler.DeleteField)
			protected.POST("/content/types/:id/fields/reorder", ctHandler.ReorderFields)

			// 内容管理 (REST: /content/entries)
			protected.GET("/content/entries", entryHandler.List)
			protected.POST("/content/entries", entryHandler.Create)
			protected.GET("/content/entries/:id", entryHandler.Get)
			protected.PUT("/content/entries/:id", entryHandler.Update)
			protected.DELETE("/content/entries/:id", entryHandler.Delete)
			protected.POST("/content/entries/:id/publish", entryHandler.Publish)
			protected.POST("/content/entries/:id/unpublish", entryHandler.Unpublish)
			protected.GET("/content/entries/:id/versions", entryHandler.GetVersions)

			// 媒体库
			protected.GET("/assets", assetHandler.List)
			protected.POST("/assets", assetHandler.Upload)
			protected.GET("/assets/:id", assetHandler.Get)
			protected.PUT("/assets/:id", assetHandler.Update)
			protected.DELETE("/assets/:id", assetHandler.Delete)
			protected.DELETE("/assets", assetHandler.BatchDelete)

			// 文件夹管理
			protected.POST("/assets/folders", assetHandler.CreateFolder)
			protected.GET("/assets/folders/tree", assetHandler.GetFolderTree)
			protected.GET("/assets/folders", assetHandler.ListFolders)
			protected.GET("/assets/folders/:id", assetHandler.GetFolder)
			protected.PUT("/assets/folders/:id", assetHandler.UpdateFolder)
			protected.DELETE("/assets/folders/:id", assetHandler.DeleteFolder)

			// API Token 管理
			protected.POST("/api-tokens", tokenHandler.Create)
			protected.GET("/api-tokens", tokenHandler.List)
			protected.GET("/api-tokens/:id", tokenHandler.Get)
			protected.PUT("/api-tokens/:id", tokenHandler.Update)
			protected.DELETE("/api-tokens/:id", tokenHandler.Delete)
			protected.POST("/api-tokens/:id/regenerate", tokenHandler.Regenerate)
			protected.POST("/api-tokens/:id/revoke", tokenHandler.Revoke)
		}
	}

	// 启动服务
	addr := ":" + cfg.Server.Port
	srv := &http.Server{
		Addr:         addr,
		Handler:      r,
		ReadTimeout:  time.Duration(cfg.Server.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(cfg.Server.WriteTimeout) * time.Second,
	}

	// Graceful shutdown
	go func() {
		logger.Info().Str("addr", addr).Msg("server listening")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal().Err(err).Msg("failed to start server")
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	logger.Info().Msg("shutting down server...")

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), time.Duration(cfg.Server.ShutdownTimeout)*time.Second)
	defer shutdownCancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Fatal().Err(err).Msg("server forced to shutdown")
	}

	logger.Info().Msg("server exited")
}

func initDB(cfg config.DatabaseConfig) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}

	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}

	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	sqlDB.SetConnMaxLifetime(cfg.GetConnMaxLifetime())

	return db, nil
}
