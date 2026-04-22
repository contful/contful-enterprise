package service

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/contful/contful/admin/internal/storage"
)

// StorageConfigService 封装 ConfigService，提供站点级存储配置。
// 负责从 site_configs 读取 storage.driver / storage.* 相关键值。
type StorageConfigService struct {
	configService *ConfigService

	mu   sync.RWMutex
	cache map[uuid.UUID]*storageConfig
}

// storageConfig 缓存的配置项
type storageConfig struct {
	driver   string
	provider *storage.ProviderConfig
}

// NewStorageConfigService 创建存储配置服务
func NewStorageConfigService(configService *ConfigService) *StorageConfigService {
	return &StorageConfigService{
		configService: configService,
		cache:        make(map[uuid.UUID]*storageConfig),
	}
}

// BuildStorageConfigFunc 返回一个可注入 StorageManager 的配置函数。
// 该函数会从 site_configs 读取 storage.driver 等键值，缓存在 StorageConfigService 中。
func (s *StorageConfigService) BuildStorageConfigFunc() storage.StorageConfigFunc {
	return func(ctx context.Context, siteID uuid.UUID) (*storage.ProviderConfig, string, error) {
		s.mu.RLock()
		if cached, ok := s.cache[siteID]; ok {
			s.mu.RUnlock()
			return cached.provider, cached.driver, nil
		}
		s.mu.RUnlock()

		cfg := &storage.ProviderConfig{}
		var driver string

		// 读取驱动名称
		driver, _ = s.configService.Get(ctx, siteID, "storage.driver")
		if driver == "" {
			driver = "local"
		}

		// 根据驱动类型读取对应配置
		switch driver {
		case "local":
			if root, _ := s.configService.Get(ctx, siteID, "storage.local.root"); root != "" {
				cfg.RootDir = root
			}
			if baseURL, _ := s.configService.Get(ctx, siteID, "storage.local.base_url"); baseURL != "" {
				cfg.BaseURL = baseURL
			}

		case "oss":
			cfg.Bucket, _ = s.configService.Get(ctx, siteID, "storage.oss.bucket")
			cfg.Endpoint, _ = s.configService.Get(ctx, siteID, "storage.oss.endpoint")
			cfg.Region, _ = s.configService.Get(ctx, siteID, "storage.oss.region")
			cfg.AccessKey, _ = s.configService.Get(ctx, siteID, "storage.oss.access_key_id")
			cfg.SecretKey, _ = s.configService.Get(ctx, siteID, "storage.oss.access_key_secret")
			if baseURL, _ := s.configService.Get(ctx, siteID, "storage.oss.base_url"); baseURL != "" {
				cfg.BaseURL = baseURL
			}

		case "s3":
			cfg.Bucket, _ = s.configService.Get(ctx, siteID, "storage.s3.bucket")
			cfg.Endpoint, _ = s.configService.Get(ctx, siteID, "storage.s3.endpoint")
			cfg.Region, _ = s.configService.Get(ctx, siteID, "storage.s3.region")
			cfg.AccessKey, _ = s.configService.Get(ctx, siteID, "storage.s3.access_key")
			cfg.SecretKey, _ = s.configService.Get(ctx, siteID, "storage.s3.secret_key")
			cfg.PathPrefix, _ = s.configService.Get(ctx, siteID, "storage.s3.path_prefix")
			// force_path_style（布尔值）
			if fps, _ := s.configService.Get(ctx, siteID, "storage.s3.force_path_style"); fps == "true" {
				cfg.ForcePathStyle = true
			}
			if baseURL, _ := s.configService.Get(ctx, siteID, "storage.s3.base_url"); baseURL != "" {
				cfg.BaseURL = baseURL
			}

		case "cos":
			cfg.Bucket, _ = s.configService.Get(ctx, siteID, "storage.cos.bucket")
			cfg.Region, _ = s.configService.Get(ctx, siteID, "storage.cos.region")
			cfg.AccessKey, _ = s.configService.Get(ctx, siteID, "storage.cos.secret_id")
			cfg.SecretKey, _ = s.configService.Get(ctx, siteID, "storage.cos.secret_key")
			if baseURL, _ := s.configService.Get(ctx, siteID, "storage.cos.base_url"); baseURL != "" {
				cfg.BaseURL = baseURL
			}

		case "obs":
			cfg.Bucket, _ = s.configService.Get(ctx, siteID, "storage.obs.bucket")
			cfg.Endpoint, _ = s.configService.Get(ctx, siteID, "storage.obs.endpoint")
			cfg.Region, _ = s.configService.Get(ctx, siteID, "storage.obs.region")
			cfg.AccessKey, _ = s.configService.Get(ctx, siteID, "storage.obs.access_key")
			cfg.SecretKey, _ = s.configService.Get(ctx, siteID, "storage.obs.secret_key")
			if baseURL, _ := s.configService.Get(ctx, siteID, "storage.obs.base_url"); baseURL != "" {
				cfg.BaseURL = baseURL
			}
		}

		// 通用 public_url_template
		if pubURL, _ := s.configService.Get(ctx, siteID, "storage.public_url_template"); pubURL != "" {
			cfg.BaseURL = pubURL
		}

		// 缓存
		s.mu.Lock()
		s.cache[siteID] = &storageConfig{driver: driver, provider: cfg}
		s.mu.Unlock()

		return cfg, driver, nil
	}
}

// InvalidateCache 清除指定站点的配置缓存（配置变更后调用）
func (s *StorageConfigService) InvalidateCache(siteID uuid.UUID) {
	s.mu.Lock()
	delete(s.cache, siteID)
	s.mu.Unlock()
}
