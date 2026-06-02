// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

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
	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog"

	"github.com/contful/contful/openapi/internal/config"
	"github.com/contful/contful/openapi/internal/database"
	"github.com/contful/contful/openapi/internal/graphql"
	"github.com/contful/contful/openapi/internal/metrics"
	"github.com/contful/contful/openapi/internal/middleware"
	"github.com/contful/contful/openapi/internal/model"
	"github.com/contful/contful/openapi/internal/repository"
	"github.com/contful/contful/openapi/internal/service"
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

	// 初始化数据库（PostgreSQL）
	dsnCfg := &database.DSNConfig{
		Host:     cfg.Database.Host,
		Port:     cfg.Database.Port,
		User:     cfg.Database.User,
		Password: cfg.Database.Password,
		Name:     cfg.Database.Name,
		SSLMode:  cfg.Database.SSLMode,
	}
	db, err := database.Open(dsnCfg, cfg.Database.MaxOpenConns, cfg.Database.MaxIdleConns, cfg.Database.ConnMaxLifetime)
	if err != nil {
		logger.Fatal().Err(err).Msg("failed to connect database")
	}
	logger.Info().Str("db_type", database.DBType).Msg("database connected")

	// 初始化 Metrics（企业版）
	metricsHandler := metrics.NewHandler(db)

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

	// 初始化缓存服务
	cacheSvc := service.NewCacheService(rdb)

	entryRepo := repository.NewEntryRepository(db)
	csRepo := repository.NewContentSchemaRepository(db)
	fieldRepo := repository.NewFieldRepository(db)
	entrySvc := service.NewEntryService(entryRepo, csRepo, cacheSvc)

	// 站点配置服务（从 sites.settings JSONB 读取）

	// 初始化 Gin
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(gin.Logger())
	r.Use(metrics.Middleware()) // HTTP metrics (enterprise)

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

	// Prometheus metrics（企业版）
	r.GET("/metrics", metricsHandler.Metrics)

	// API 路由组
	api := r.Group("/api/v1")
	api.Use(middleware.TokenAuthMiddleware(tokenSvc, logger))

	// 速率限制（从 config.yaml 读取）
	rateLimiter := middleware.NewRateLimiter(rdb)
	if cfg.RateLimit.Enabled {
		api.Use(rateLimiter.RateLimitByToken(cfg.RateLimit.RequestsPerMin))
	}

	// 内容读取路由（需 Token 认证）
	// GET  /api/v1/content/:slug         — 列出指定内容类型的已发布条目
	// GET  /api/v1/content/:slug/:id     — 获取单个已发布条目
	api.GET("/content/:slug", middleware.RequireRead(), func(c *gin.Context) {
		tc := middleware.GetTokenContext(c)
		if tc == nil {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}

		slug := c.Param("slug")
		locale := c.Query("locale")
		sortField := c.Query("sort_field")
		sortOrder := c.Query("sort_order")
		page, pageSize := service.ParsePage(c.Query("page"), c.Query("page_size"))

		resp, err := entrySvc.ListBySlug(c.Request.Context(), tc.SiteID, slug, locale, sortField, sortOrder, page, pageSize)
		if err != nil {
			if err == service.ErrContentSchemaNotFound {
				c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "content type not found"))
				return
			}
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
			return
		}

		c.JSON(http.StatusOK, model.NewSuccessResponse(resp))
	})

	api.GET("/content/:slug/:id", middleware.RequireRead(), func(c *gin.Context) {
		tc := middleware.GetTokenContext(c)
		if tc == nil {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}

		slug := c.Param("slug")
		idStr := c.Param("id")
		entryID, err := parseUUID(idStr)
		if err != nil {
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid id, expected UUID format (e.g. xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx)"))
			return
		}

		item, err := entrySvc.GetByID(c.Request.Context(), tc.SiteID, slug, entryID)
		if err != nil {
			if err == service.ErrContentSchemaNotFound || err == service.ErrEntryNotFound {
				c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "not found"))
				return
			}
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
			return
		}

		c.JSON(http.StatusOK, model.NewSuccessResponse(item))
	})

	api.POST("/content/:slug", middleware.RequireWrite(), func(c *gin.Context) {
		c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{
			"message": "write API coming in M2",
			"slug":    c.Param("slug"),
		}))
	})

	// 站点配置路由（需 Token 认证）
	// GET /api/v1/configs      — 无参数时返回公开配置键值对
	// GET /api/v1/configs/:key — 获取指定 key 的配置值
	api.GET("/configs", func(c *gin.Context) {
		tc := middleware.GetTokenContext(c)
		if tc == nil {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}

		// 查询公开配置
		var configs []struct {
			ConfigKey   string `json:"config_key"`
			ConfigValue string `json:"config_value"`
		}
		if err := db.WithContext(c.Request.Context()).
			Table("system_config").
			Where("is_public = ?", true).
			Select("config_key, config_value").
			Find(&configs).Error; err != nil {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
			return
		}

		result := make(map[string]string, len(configs))
		for _, cfg := range configs {
			result[cfg.ConfigKey] = cfg.ConfigValue
		}
		c.JSON(http.StatusOK, model.NewSuccessResponse(result))
	})

	api.GET("/configs/:key", func(c *gin.Context) {
		tc := middleware.GetTokenContext(c)
		if tc == nil {
			c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
			return
		}
		key := c.Param("key")
		if key == "" {
			c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "key is required"))
			return
		}

		var configValue string
		err := db.WithContext(c.Request.Context()).
			Table("system_config").
			Where("config_key = ?", key).
			Select("config_value").
			Scan(&configValue).Error
		if err != nil || configValue == "" {
			c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "config not found"))
			return
		}
		c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"key": key, "value": configValue}))
	})

	// GraphQL API（需 Token 认证）
	gqlResolver := graphql.NewResolver(entryRepo, csRepo, fieldRepo)
	gqlBuilder := graphql.NewSchemaBuilder(csRepo, fieldRepo, gqlResolver)
	api.POST("/graphql", graphql.GraphQLHandler(gqlBuilder))
	api.GET("/graphql", graphql.GraphQLHandler(gqlBuilder))

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

// parseUUID 解析 UUID 字符串，失败返回 error
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}
