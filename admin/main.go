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
	"gorm.io/gorm"

	"github.com/contful/contful/admin/internal/audit_callback"
	"github.com/contful/contful/admin/internal/config"
	"github.com/contful/contful/admin/internal/database"
	"github.com/contful/contful/admin/internal/handler"
	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/contful/contful/admin/internal/service"
	"github.com/contful/contful/admin/internal/storage"
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

	// 注册 AuditLog 数据签名 callback
	audit_callback.Register(db)
	logger.Info().Msg("AuditLog 签名 callback 已注册")

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
	siteRepo := repository.NewSiteRepository(db)
	auditRepo := repository.NewAuditRepository(db)
	contentTypeRepo := repository.NewContentTypeRepository(db)
	fieldRepo := repository.NewFieldRepository(db)
	entryRepo := repository.NewEntryRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	tokenRepo := repository.NewAPITokenRepository(db)

	// 初始化 Service
	configRepo := repository.NewSiteConfigRepository(db)
	configService := service.NewConfigService(configRepo, os.Getenv("CONTFUL_CONFIG_MASTER_KEY"))
	if configService.GetMasterKey() != "" {
		logger.Info().Msg("配置中心已启用（AES-256-GCM 主密钥已加载）")
	} else {
		logger.Warn().Msg("警告：CONTFUL_CONFIG_MASTER_KEY 未设置，敏感配置将无法加密存储")
	}

	authService := service.NewAuthService(userRepo, auditRepo, redisClient, cfg.JWT.Secret, configService)
	userService := service.NewUserService(userRepo)
	siteService := service.NewSiteService(db, siteRepo)
	ctService := service.NewContentTypeService(contentTypeRepo, fieldRepo, logger)
	entryService := service.NewEntryService(entryRepo, contentTypeRepo, fieldRepo)
	tokenService := service.NewAPITokenService(tokenRepo)

	// 初始化 MFA 服务（PRE-005）
	mfaService := service.NewMFAService(userRepo, redisClient, cfg.JWT.Secret)
	authService.SetMFAService(mfaService)
	logger.Info().Msg("MFA/TOTP 服务已就绪")

	// 初始化存储驱动（支持 per-site 动态切换）
	storageConfigSvc := service.NewStorageConfigService(configService)
	storageManager := storage.NewStorageManager(
		storageConfigSvc.BuildStorageConfigFunc(),
		0, // 0 = 不过期缓存（配置变更后主动 Invalidate）
	)
	logger.Info().Msg("存储管理器已就绪（支持 per-site 动态存储驱动切换）")

	assetService := service.NewAssetService(assetRepo, storageManager)
	assetService.SetConfigService(configService)

	// 初始化 Handler
	authHandler := handler.NewAuthHandler(authService)
	mfaHandler := handler.NewMFAHandler(mfaService, authService)
	userHandler := handler.NewUserHandler(userService)
	siteHandler := handler.NewSiteHandler(siteService)
	ctHandler := handler.NewContentTypeHandler(ctService)
	entryHandler := handler.NewEntryHandler(entryService, configService)
	assetHandler := handler.NewAssetHandler(assetService)
	tokenHandler := handler.NewAPITokenHandler(tokenService)
	configHandler := handler.NewConfigHandler(configService)
	integrityHandler := handler.NewIntegrityHandler(entryRepo, assetRepo, auditRepo, configService)

	// 初始化 Gin
	r := gin.New()
	r.Use(gin.Recovery())
	// CORS 由部署环境统一处理（反向代理/API 网关）

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
	api := r.Group("/admin/api/v1")
	{
		// 公开路由
		auth := api.Group("/auth")
		{
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			// MFA 登录步骤 2（无需 JWT）
			auth.POST("/mfa/verify", mfaHandler.Verify)
			auth.POST("/mfa/recover", mfaHandler.Recover)
		}

		// 需要认证的路由
		protected := api.Group("")
		protected.Use(middleware.JWTAuth(authHandler))
		{
			// 认证相关
			protected.POST("/auth/logout", authHandler.Logout)

			// MFA 管理（需要 JWT 认证）
			protected.POST("/auth/mfa/setup", mfaHandler.Setup)
			protected.POST("/auth/mfa/enable", mfaHandler.Enable)
			protected.POST("/auth/mfa/disable", mfaHandler.Disable)

			// 站点管理
			protected.GET("/sites/mine", siteHandler.MySites)
			protected.GET("/sites", siteHandler.List)
			protected.POST("/sites", siteHandler.Create)
			protected.GET("/sites/:id", siteHandler.Get)
			protected.PUT("/sites/:id", siteHandler.Update)
			protected.DELETE("/sites/:id", siteHandler.Delete)

			// 用户管理
			protected.GET("/users/me", authHandler.Me)
			protected.GET("/users", userHandler.List)
			protected.POST("/users", userHandler.Create)
			protected.GET("/users/:id", userHandler.Get)
			protected.PUT("/users/:id", userHandler.Update)
			protected.DELETE("/users/:id", userHandler.Delete)

			// 内容类型管理 (REST: /content/types)
			protected.GET("/content/types", ctHandler.List)
			protected.POST("/content/types", ctHandler.Create)
			protected.GET("/content/types/:id", ctHandler.Get)
			protected.PUT("/content/types/:id", ctHandler.Update)
			protected.DELETE("/content/types/:id", ctHandler.Delete)
			protected.POST("/content/types/:id/fields", ctHandler.CreateField)
			protected.GET("/content/types/:id/fields", ctHandler.ListFields)
			// 字段操作（嵌套在 :id 之下，避免与 /content/types/:id 冲突）
			protected.PUT("/content/types/:id/fields/:fieldId", ctHandler.UpdateField)
			protected.DELETE("/content/types/:id/fields/:fieldId", ctHandler.DeleteField)
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
			// 批量操作（静态路径在前，避免与 :id 冲突）
			protected.POST("/content/entries/batch-publish", entryHandler.BatchPublish)
			protected.POST("/content/entries/batch-unpublish", entryHandler.BatchUnpublish)
			protected.DELETE("/content/entries/batch-delete", entryHandler.BatchDelete)

		// 媒体库
			protected.GET("/assets", assetHandler.List)
			protected.POST("/assets", assetHandler.Upload)
			// 文件夹管理（静态路径必须在 :id 之前，否则 folders 会被 :id 捕获）
			protected.POST("/assets/folders", assetHandler.CreateFolder)
			protected.GET("/assets/folders/tree", assetHandler.GetFolderTree)
			protected.GET("/assets/folders", assetHandler.ListFolders)
			protected.GET("/assets/folders/:id", assetHandler.GetFolder)
			protected.PUT("/assets/folders/:id", assetHandler.UpdateFolder)
			protected.DELETE("/assets/folders/:id", assetHandler.DeleteFolder)
			// 批量删除必须在 :id 路由之前注册（Gin 静态路径优先）
			protected.DELETE("/assets/batch-delete", assetHandler.BatchDelete)
			// 资产 CRUD（:id 路由放在最后）
			protected.GET("/assets/:id", assetHandler.Get)
			protected.PUT("/assets/:id", assetHandler.Update)
			protected.DELETE("/assets/:id", assetHandler.Delete)

			// API Token 管理
			protected.POST("/api-tokens", tokenHandler.Create)
			protected.GET("/api-tokens", tokenHandler.List)
			protected.GET("/api-tokens/:id", tokenHandler.Get)
			protected.PUT("/api-tokens/:id", tokenHandler.Update)
			protected.DELETE("/api-tokens/:id", tokenHandler.Delete)
			protected.POST("/api-tokens/:id/regenerate", tokenHandler.Regenerate)
			protected.POST("/api-tokens/:id/revoke", tokenHandler.Revoke)

			// 站点配置管理（PRE-001）
			protected.GET("/sites/:id/configs", configHandler.List)
			protected.GET("/sites/:id/configs/:key", configHandler.Get)
			protected.PUT("/sites/:id/configs/:key", configHandler.Set)
			protected.DELETE("/sites/:id/configs/:key", configHandler.Delete)

			// 数据完整性验签（PRE-004）
			protected.GET("/integrity/verify", integrityHandler.Verify)
			protected.POST("/integrity/verify/batch", integrityHandler.BatchVerify)
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
	dsnCfg := &database.DSNConfig{
		Host:     cfg.Host,
		Port:     cfg.Port,
		User:     cfg.User,
		Password: cfg.Password,
		Name:     cfg.Name,
		SSLMode:  cfg.SSLMode,
	}
	return database.Open(dsnCfg, cfg.MaxOpenConns, cfg.MaxIdleConns, cfg.ConnMaxLifetime)
}
