// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"
	zlog "github.com/rs/zerolog/log"
	"gorm.io/gorm"

	"github.com/contful/contful/admin/internal/audit"
	"github.com/contful/contful/admin/internal/cmd"
	"github.com/contful/contful/admin/internal/config"
	"github.com/contful/contful/admin/internal/database"
	"github.com/contful/contful/admin/internal/handler"
	lic "github.com/contful/contful/admin/internal/license"
	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/crypto"
	"github.com/contful/contful/admin/internal/metrics"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/contful/contful/admin/internal/service"
	"github.com/contful/contful/admin/internal/storage"
)

func main() {
	// CLI 子命令路由
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "check-db":
			cmd.CheckDB()
			return
		case "gen-key":
			cmd.GenKey()
			return
		case "version":
			cmd.PrintVersion()
			return
		}
	}
	runServer()
}

// runServer 启动 HTTP 服务。
func runServer() {
	// 初始化配置
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// 初始化日志
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix
	logger := zerolog.New(zerolog.ConsoleWriter{Out: log.Writer()}).With().Timestamp().Logger()
	// 同步全局 zerolog logger，确保 handler/service 的日志也用相同格式输出
	zlog.Logger = logger
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
	audit.Register(db)
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
	schemaRepo := repository.NewSchemaRepository(db)
	fieldRepo := repository.NewFieldRepository(db)
	entryRepo := repository.NewEntryRepository(db)
	assetRepo := repository.NewAssetRepository(db)
	tokenRepo := repository.NewAPITokenRepository(db)
	systemRoleRepo := repository.NewSystemRoleRepository(db)
	systemConfigRepo := repository.NewSystemConfigRepository(db, redisClient)
	permRepo := repository.NewPermissionRepository(db)

	// 初始化加密器（根据配置选择算法）
	var crypter crypto.Crypter
	if cfg.Security.Secret != "" {
		var err error
		crypter, err = crypto.NewCrypter(cfg.Security.Algorithm, cfg.Security.Secret)
		if err != nil {
			log.Fatalf("创建加密器失败: %v", err)
		}
		logger.Info().Str("algorithm", cfg.Security.Algorithm).Msg("加密器已就绪")
	} else {
		logger.Warn().Msg("警告：SECRET 未设置，敏感数据将无法加密存储")
	}

	// 初始化 Service（不再需要 configRepo）
	configService := service.NewConfigService(crypter, cfg)

	// 初始化数据签名器（根据 crypto_mode 自动选择 HMAC-SHA256 或 HMAC-SM3）
	var auditHasher crypto.Hasher
	if provider := config.GetCryptoProvider(); provider != nil {
		auditHasher = provider
	}
	dataSigner, err := service.NewDefaultSigner(cfg.Audit.SigningKey, auditHasher)
	if err != nil {
		logger.Warn().Err(err).Msg("数据签名器未启用（签名密钥无效或未配置）")
	} else if dataSigner.IsEnabled() {
		logger.Info().Str("alg", dataSigner.Algorithm()).Msg("数据签名器已就绪")
	}

	// 初始化 Audit Service（审计日志）
	auditService := service.NewAuditService(auditRepo, configService)
	logger.Info().Msg("Audit 服务已就绪")

	// 初始化 RBAC 服务（不再需要 siteRoleRepo 和 siteUserRepo）
	rbacService := service.NewRBACService(db, redisClient, systemRoleRepo, userRepo, permRepo)

	authService := service.NewAuthService(userRepo, siteRepo, systemConfigRepo, auditRepo, redisClient, cfg.JWT.Secret, configService)
	siteService := service.NewSiteService(db, siteRepo)
	schemaService := service.NewSchemaService(schemaRepo, fieldRepo, logger)
	entryService := service.NewEntryService(entryRepo, schemaRepo, fieldRepo)
	scheduleService := service.NewScheduleService(db, entryRepo, logger)
	tokenService := service.NewAPITokenService(tokenRepo, crypter)
	cacheService := service.NewCacheService(redisClient)

	// 初始化 MFA 服务（PRE-005）
	mfaService := service.NewMFAService(userRepo, redisClient, crypter)
	authService.SetMFAService(mfaService)
	logger.Info().Msg("MFA/TOTP 服务已就绪")

	// 初始化存储驱动（从 config.yaml + 环境变量读取，全局共用）
	storageProvider, _, err := storage.NewStorage(ctx, cfg)
	if err != nil {
		logger.Fatal().Err(err).Msg("初始化存储驱动失败")
	}
	logger.Info().Str("driver", cfg.Storage.Driver).Msg("存储驱动已就绪（全局单例）")

	// 加载企业 License
	licenseInfo, licenseErr := lic.Load()
	licenseHandler := lic.NewHandler(licenseInfo)
	if licenseErr != nil {
		logger.Warn().Err(licenseErr).Msg("license not loaded — running in unlicensed mode")
	} else {
		logger.Info().
			Str("customer", licenseInfo.Customer).
			Str("product", licenseInfo.ProductName).
			Str("version", licenseInfo.ProductVersion).
			Msg("license loaded")
	}

	userService := service.NewUserService(userRepo, storageProvider)
	assetService := service.NewAssetService(assetRepo, storageProvider)
	assetService.SetConfigService(configService)

	// 初始化非对称加密器（RSA 或 SM2，由 crypto_mode 决定）
	// 优先从文件加载密钥对，不存在时自动生成
	pubPath := resolveConfigPath(cfg.Security.PublicKeyPath)
	privPath := resolveConfigPath(cfg.Security.PrivateKeyPath)
	var asymCrypter crypto.AsymmetricCrypter
	var asyPubKeyPEM, asyPrivKeyPEM string

	// 获取 Provider 提供的 AsymmetricCrypter（与 crypto_mode 一致）
	if provider := config.GetCryptoProvider(); provider != nil {
		asymCrypter = provider
	} else {
		// fallback: 没有 Provider 时使用 RSA
		asymCrypter = crypto.NewRSACrypter()
	}

	// 优先从文件加载密钥对（兼容已有部署）
	pubBytes, pubErr := os.ReadFile(pubPath)
	privBytes, privErr := os.ReadFile(privPath)
	if pubErr == nil && privErr == nil {
		asyPubKeyPEM = string(pubBytes)
		asyPrivKeyPEM = string(privBytes)
		logger.Info().Str("mode", cfg.Security.CryptoMode).Msg("asymmetric key pair loaded from files")
	} else {
		// 文件不存在，自动生成
		pub, priv, genErr := asymCrypter.GenerateKeyPair()
		if genErr != nil {
			logger.Fatal().Err(genErr).Msg("failed to generate key pair")
		}
		asyPubKeyPEM = pub
		asyPrivKeyPEM = priv
		// 写入文件以便下次启动复用
		if err := os.WriteFile(pubPath, []byte(pub), 0600); err != nil {
			logger.Warn().Err(err).Str("path", pubPath).Msg("failed to save public key")
		}
		if err := os.WriteFile(privPath, []byte(priv), 0600); err != nil {
			logger.Warn().Err(err).Str("path", privPath).Msg("failed to save private key")
		}
		logger.Info().Str("mode", cfg.Security.CryptoMode).Msg("asymmetric key pair generated and saved")
	}

	// 初始化 Handler
	authHandler := handler.NewAuthHandler(authService, asymCrypter, asyPubKeyPEM, asyPrivKeyPEM, cfg.Security.CryptoMode)
	mfaHandler := handler.NewMFAHandler(mfaService, authService)
	userHandler := handler.NewUserHandler(userService, auditService)
	siteHandler := handler.NewSiteHandler(siteService, auditService)
	schemaHandler := handler.NewSchemaHandler(schemaService)
	entryHandler := handler.NewEntryHandler(entryService, configService)
	scheduleHandler := handler.NewScheduleHandler(scheduleService)
	metricsHandler := metrics.NewHandler(db)
	assetHandler := handler.NewAssetHandler(assetService)
	tokenHandler := handler.NewAPITokenHandler(tokenService, auditService)
		integrityHandler := handler.NewIntegrityHandler(entryRepo, assetRepo, auditRepo, configService)
	cacheHandler := handler.NewCacheHandler(cacheService)
	systemRoleHandler := handler.NewSystemRoleHandler(rbacService, auditService)
	systemConfigHandler := handler.NewSystemConfigHandler(systemConfigRepo, rbacService, auditService)
	permHandler := handler.NewPermissionHandler(permRepo, rbacService)
	dashboardHandler := handler.NewDashboardHandler(service.NewDashboardService(db))
	auditHandler := handler.NewAuditHandler(auditService)

	// 启动排期调度器（后台 goroutine）
	ctxCron, cancelCron := context.WithCancel(context.Background())
	defer cancelCron()
	go scheduleService.StartCron(ctxCron)

	// 初始化 Gin
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(metrics.Middleware()) // HTTP metrics (enterprise)

	// 注入 DataSigner + Hasher 到请求 context（供 GORM callback 使用）
	if dataSigner != nil && dataSigner.IsEnabled() {
		r.Use(func(c *gin.Context) {
			ctx := audit.WithSigner(c.Request.Context(), dataSigner)
			if auditHasher != nil {
				ctx = audit.WithHasher(ctx, auditHasher)
			}
			c.Request = c.Request.WithContext(ctx)
			c.Next()
		})
	}

	// 注：CORS 由部署环境的反向代理（nginx）处理，不需要在这里注册中间件

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

	// Prometheus metrics（企业版）
	r.GET("/metrics", metricsHandler.Metrics)

	// API 路由组
	api := r.Group("/admin/api/v1")
	{
		// 公开路由
		auth := api.Group("/auth")
		{
			auth.GET("/public/key", authHandler.PublicKey) // 非对称加密公钥 + Anti-Replay Token
			auth.POST("/register", authHandler.Register)
			auth.POST("/login", authHandler.Login)
			auth.POST("/refresh", authHandler.Refresh)
			// MFA 登录步骤 2（无需 JWT，但受限流保护）
			auth.POST("/mfa/verify", mfaHandler.Verify)
			auth.POST("/mfa/recover", mfaHandler.Recover)
	}

	// ─── 系统配置（公开，无需认证，登录/注册页使用）─────────
	// 注意：必须在 protected 组之前注册，避免 :key 通配符路由冲突
	api.GET("/system/config/site", systemConfigHandler.GetSiteConfig)
	api.GET("/system/config/public", systemConfigHandler.GetPublicConfig)

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
		protected.GET("/sites/mine",
			middleware.RequirePermission(rbacService, "sites:read"),
			siteHandler.MySites)
		protected.GET("/sites",
			middleware.RequirePermission(rbacService, "sites:read"),
			siteHandler.List)
		protected.POST("/sites",
			middleware.RequirePermission(rbacService, "sites:write"),
			siteHandler.Create)
		protected.GET("/sites/:id",
			middleware.RequirePermission(rbacService, "sites:read"),
			siteHandler.Get)
		protected.PUT("/sites/:id",
			middleware.RequirePermission(rbacService, "sites:write"),
			siteHandler.Update)
		protected.DELETE("/sites/:id",
			middleware.RequirePermission(rbacService, "sites:delete"),
			siteHandler.Delete)

			// 用户管理
			protected.GET("/users/me", authHandler.Me)
			protected.PATCH("/users/me", userHandler.UpdateMe)
			protected.PUT("/users/me/password", userHandler.UpdatePassword)
			protected.POST("/users/me/avatar", userHandler.UploadAvatar)
			protected.GET("/users",
				middleware.RequirePermission(rbacService, "users:read"),
				userHandler.List)
			protected.POST("/users",
				middleware.RequirePermission(rbacService, "users:write"),
				userHandler.Create)
			protected.GET("/users/:id",
				middleware.RequirePermission(rbacService, "users:read"),
				userHandler.Get)
			protected.PUT("/users/:id",
				middleware.RequirePermission(rbacService, "users:write"),
				userHandler.Update)
			protected.DELETE("/users/:id",
				middleware.RequirePermission(rbacService, "users:delete"),
				userHandler.Delete)
			protected.POST("/users/:id/restore",
				middleware.RequirePermission(rbacService, "users:write"),
				userHandler.Restore)

			// 管理员重置用户密码（不需要旧密码）
			protected.POST("/users/:id/reset-password",
				middleware.RequirePermission(rbacService, "users:write"),
				userHandler.ResetPassword)

			// 用户数据签名/验签
			protected.POST("/users/:id/sign",
				middleware.RequirePermission(rbacService, "users:write"),
				userHandler.Sign)
			protected.POST("/users/:id/verify",
				middleware.RequirePermission(rbacService, "users:read"),
				userHandler.Verify)

			// 用户-角色关联管理
			protected.GET("/users/:id/roles",
				middleware.RequirePermission(rbacService, "roles:read"),
				systemRoleHandler.GetUserRoles)
			protected.PUT("/users/:id/roles/:roleId",
				middleware.RequirePermission(rbacService, "roles:write"),
				systemRoleHandler.AssignUserRole)
			protected.DELETE("/users/:id/roles/:roleId",
				middleware.RequirePermission(rbacService, "roles:write"),
				systemRoleHandler.RemoveUserRole)

			// 内容类型管理 (REST: /content/schemas)
			protected.GET("/content/schemas",
				middleware.RequirePermission(rbacService, "schema:read"),
				schemaHandler.List)
			protected.POST("/content/schemas",
				middleware.RequirePermission(rbacService, "schema:write"),
				schemaHandler.Create)
			protected.GET("/content/schemas/:id",
				middleware.RequirePermission(rbacService, "schema:read"),
				schemaHandler.Get)
			protected.PUT("/content/schemas/:id",
				middleware.RequirePermission(rbacService, "schema:write"),
				schemaHandler.Update)
			protected.DELETE("/content/schemas/:id",
				middleware.RequirePermission(rbacService, "schema:delete"),
				schemaHandler.Delete)
			protected.POST("/content/schemas/:id/fields",
				middleware.RequirePermission(rbacService, "schema:write"),
				schemaHandler.CreateField)
			protected.GET("/content/schemas/:id/fields",
				middleware.RequirePermission(rbacService, "schema:read"),
				schemaHandler.ListFields)
			// 字段操作（嵌套在 :id 之下，避免与 /content/schemas/:id 冲突）
			protected.PUT("/content/schemas/:id/fields/:fieldId",
				middleware.RequirePermission(rbacService, "schema:write"),
				schemaHandler.UpdateField)
			protected.DELETE("/content/schemas/:id/fields/:fieldId",
				middleware.RequirePermission(rbacService, "schema:delete"),
				schemaHandler.DeleteField)
			protected.POST("/content/schemas/:id/fields/reorder",
				middleware.RequirePermission(rbacService, "schema:write"),
				schemaHandler.ReorderFields)

			// 内容模型数据签名/验签
			protected.POST("/content/schemas/:id/sign",
				middleware.RequirePermission(rbacService, "schema:write"),
				schemaHandler.Sign)
			protected.POST("/content/schemas/:id/verify",
				middleware.RequirePermission(rbacService, "schema:read"),
				schemaHandler.Verify)

			// 内容管理 (REST: /content/entries)
			protected.GET("/content/entries",
				middleware.RequirePermission(rbacService, "entry:read"),
				entryHandler.List)
			protected.POST("/content/entries",
				middleware.RequirePermission(rbacService, "entry:write"),
				entryHandler.Create)
			// 排期列表（静态路径必须在 :id 之前注册）
			protected.GET("/content/entries/scheduled",
				middleware.RequirePermission(rbacService, "entry:read"),
				scheduleHandler.ListScheduled)
			protected.GET("/content/entries/:id",
				middleware.RequirePermission(rbacService, "entry:read"),
				entryHandler.Get)
			protected.PUT("/content/entries/:id",
				middleware.RequirePermission(rbacService, "entry:write"),
				entryHandler.Update)
			protected.DELETE("/content/entries/:id",
				middleware.RequirePermission(rbacService, "entry:delete"),
				entryHandler.Delete)
			protected.POST("/content/entries/:id/publish",
				middleware.RequirePermission(rbacService, "entry:publish"),
				entryHandler.Publish)
			protected.POST("/content/entries/:id/unpublish",
				middleware.RequirePermission(rbacService, "entry:publish"),
				entryHandler.Unpublish)
			protected.GET("/content/entries/:id/versions",
				middleware.RequirePermission(rbacService, "entry:read"),
				entryHandler.GetVersions)
			// 排期操作
			protected.PUT("/content/entries/:id/schedule",
				middleware.RequirePermission(rbacService, "entry:write"),
				entryHandler.SetSchedule)
			protected.DELETE("/content/entries/:id/schedule",
				middleware.RequirePermission(rbacService, "entry:write"),
				entryHandler.ClearSchedule)
			// 批量操作（静态路径在前，避免与 :id 冲突）
			protected.POST("/content/entries/batch-publish",
				middleware.RequirePermission(rbacService, "entry:publish"),
				entryHandler.BatchPublish)
			protected.POST("/content/entries/batch-unpublish",
				middleware.RequirePermission(rbacService, "entry:publish"),
				entryHandler.BatchUnpublish)
			protected.DELETE("/content/entries/batch-delete",
				middleware.RequirePermission(rbacService, "entry:delete"),
				entryHandler.BatchDelete)

			// 媒体库
			// 静态文件访问（nginx 直连，Go 兜底）
			protected.GET("/assets/files/*filePath",
				middleware.RequirePermission(rbacService, "asset:read"),
				assetHandler.ServeFile)
			protected.GET("/assets",
				middleware.RequirePermission(rbacService, "asset:read"),
				assetHandler.List)
			protected.POST("/assets",
				middleware.RequirePermission(rbacService, "asset:write"),
				assetHandler.Upload)
			// 文件夹管理（静态路径必须在 :id 之前，否则 folders 会被 :id 捕获）
			protected.POST("/assets/folders",
				middleware.RequirePermission(rbacService, "asset:write"),
				assetHandler.CreateFolder)
			protected.GET("/assets/folders/tree",
				middleware.RequirePermission(rbacService, "asset:read"),
				assetHandler.GetFolderTree)
			protected.GET("/assets/folders",
				middleware.RequirePermission(rbacService, "asset:read"),
				assetHandler.ListFolders)
			protected.GET("/assets/folders/:id",
				middleware.RequirePermission(rbacService, "asset:read"),
				assetHandler.GetFolder)
			protected.PUT("/assets/folders/:id",
				middleware.RequirePermission(rbacService, "asset:write"),
				assetHandler.UpdateFolder)
			protected.DELETE("/assets/folders/:id",
				middleware.RequirePermission(rbacService, "asset:delete"),
				assetHandler.DeleteFolder)
			// 批量删除必须在 :id 路由之前注册（Gin 静态路径优先）
			protected.DELETE("/assets/batch-delete",
				middleware.RequirePermission(rbacService, "asset:delete"),
				assetHandler.BatchDelete)
			// 资产 CRUD（:id 路由放在最后）
			protected.GET("/assets/:id",
				middleware.RequirePermission(rbacService, "asset:read"),
				assetHandler.Get)
			protected.PUT("/assets/:id",
				middleware.RequirePermission(rbacService, "asset:write"),
				assetHandler.Update)
			protected.DELETE("/assets/:id",
				middleware.RequirePermission(rbacService, "asset:delete"),
				assetHandler.Delete)

			// API Token 管理
			protected.POST("/tokens",
				middleware.RequirePermission(rbacService, "api_token:write"),
				tokenHandler.Create)
			protected.GET("/tokens",
				middleware.RequirePermission(rbacService, "api_token:read"),
				tokenHandler.List)
			protected.GET("/tokens/:id",
				middleware.RequirePermission(rbacService, "api_token:read"),
				tokenHandler.Get)
			protected.PUT("/tokens/:id",
				middleware.RequirePermission(rbacService, "api_token:write"),
				tokenHandler.Update)
			protected.DELETE("/tokens/:id",
				middleware.RequirePermission(rbacService, "api_token:delete"),
				tokenHandler.Delete)
			protected.POST("/tokens/:id/regenerate",
				middleware.RequirePermission(rbacService, "api_token:write"),
				tokenHandler.Regenerate)
			protected.POST("/tokens/:id/revoke",
				middleware.RequirePermission(rbacService, "api_token:write"),
				tokenHandler.Revoke)
			protected.POST("/tokens/:id/export",
				middleware.RequirePermission(rbacService, "api_token:read"),
				tokenHandler.Export)

			// 数据完整性验签（PRE-004）
			protected.GET("/integrity/verify",
				middleware.RequirePermission(rbacService, "settings:read"),
				integrityHandler.Verify)
			protected.POST("/integrity/verify/batch",
				middleware.RequirePermission(rbacService, "settings:read"),
				integrityHandler.BatchVerify)

			// 缓存管理
			protected.POST("/cache/invalidate",
				middleware.RequirePermission(rbacService, "settings:write"),
				cacheHandler.InvalidateSite)

			// 仪表盘统计（不依赖 X-Site-ID）
			protected.GET("/dashboard/stats",
				middleware.RequirePermission(rbacService, "dashboard:read"),
				dashboardHandler.Stats)

			// 审计日志查询（P1 可视化）
			protected.GET("/audit/logs",
				middleware.RequirePermission(rbacService, "audit:read"),
				auditHandler.List)
			protected.GET("/audit/logs/:id",
				middleware.RequirePermission(rbacService, "audit:read"),
				auditHandler.Get)

			// ─── RBAC 系统角色管理 ──────────────────────
			// 注：permissions 静态路由必须在 :id 之前注册
			protected.GET("/system/roles/permissions",
				middleware.RequirePermission(rbacService, "roles:read"),
				systemRoleHandler.Permissions)
			protected.GET("/system/roles",
				middleware.RequirePermission(rbacService, "roles:read"),
				systemRoleHandler.List)
			protected.POST("/system/roles",
				middleware.RequirePermission(rbacService, "roles:write"),
				systemRoleHandler.Create)
			protected.GET("/system/roles/:id",
				middleware.RequirePermission(rbacService, "roles:read"),
				systemRoleHandler.Get)
			protected.PUT("/system/roles/:id",
				middleware.RequirePermission(rbacService, "roles:write"),
				systemRoleHandler.Update)
			protected.DELETE("/system/roles/:id",
				middleware.RequirePermission(rbacService, "roles:delete"),
				systemRoleHandler.Delete)

			// ─── 系统配置管理 ─────────────────────────
			// 需要认证的路由
			protected.GET("/system/config",
				middleware.RequirePermission(rbacService, "settings:read"),
				systemConfigHandler.List)
			protected.GET("/system/config/:key",
				middleware.RequirePermission(rbacService, "settings:read"),
				systemConfigHandler.Get)
			protected.PUT("/system/config/:key",
				middleware.RequirePermission(rbacService, "settings:write"),
				systemConfigHandler.Update)
			protected.POST("/system/config",
				middleware.RequirePermission(rbacService, "settings:write"),
				systemConfigHandler.Create)
			protected.DELETE("/system/config/:key",
				middleware.RequirePermission(rbacService, "settings:write"),
				systemConfigHandler.Delete)
			protected.POST("/system/config/cache/clear",
				middleware.RequirePermission(rbacService, "settings:write"),
				systemConfigHandler.ClearCache)

			// ─── 企业 License ─────────────────────────
			protected.GET("/system/license",
				middleware.RequirePermission(rbacService, "settings:read"),
				licenseHandler.GetInfo)

			// ─── 权限元数据 ─────────────────────────────
			protected.GET("/permissions",
				middleware.RequirePermission(rbacService, "roles:read"),
				func(c *gin.Context) {
					c.JSON(200, rbacService.GetPermissionsMeta())
				})

			// ─── 权限分组管理 ─────────────────────────
			protected.GET("/system/permissions",
				middleware.RequirePermission(rbacService, "roles:read"),
				permHandler.List)
			protected.POST("/system/permissions/group",
				middleware.RequirePermission(rbacService, "roles:write"),
				permHandler.CreateGroup)
			protected.PUT("/system/permissions/group/:id",
				middleware.RequirePermission(rbacService, "roles:write"),
				permHandler.UpdateGroup)
			protected.DELETE("/system/permissions/group/:id",
				middleware.RequirePermission(rbacService, "roles:write"),
				permHandler.DeleteGroup)
			protected.POST("/system/permissions",
				middleware.RequirePermission(rbacService, "roles:write"),
				permHandler.CreatePermission)
			protected.PUT("/system/permissions/:id",
				middleware.RequirePermission(rbacService, "roles:write"),
				permHandler.UpdatePermission)
			protected.DELETE("/system/permissions/:id",
				middleware.RequirePermission(rbacService, "roles:write"),
				permHandler.DeletePermission)
			protected.POST("/system/permissions/cache/clear",
				middleware.RequirePermission(rbacService, "roles:write"),
				permHandler.ClearCache)
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

// buildDatabaseURL 构建数据库 URL
func buildDatabaseURL(cfg config.DatabaseConfig) string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=%s",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.Name, cfg.SSLMode,
	)
}

// resolveConfigPath 解析配置文件相对路径，依次搜索 conf/、../conf/、当前目录
func resolveConfigPath(path string) string {
	searchPaths := []string{"./conf", "../conf", "."}
	for _, dir := range searchPaths {
		candidate := filepath.Join(dir, path)
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return path
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
