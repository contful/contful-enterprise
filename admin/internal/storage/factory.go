package storage

import (
	"context"
	"fmt"
	"sync"
)

// ProviderFactory 存储驱动工厂
type ProviderFactory struct {
	mu         sync.RWMutex
	providers  map[string]func(ctx context.Context, cfg *ProviderConfig) (StorageProvider, error)
	defaultCfg *ProviderConfig
}

var globalFactory *ProviderFactory

func init() {
	globalFactory = &ProviderFactory{
		providers: make(map[string]func(ctx context.Context, cfg *ProviderConfig) (StorageProvider, error)),
	}
	// 注册内置驱动
	RegisterProvider("local", NewLocalProvider)
	RegisterProvider("s3", NewS3Provider)
}

// RegisterProvider 注册存储驱动（drivers 包内 init() 调用）
func RegisterProvider(name string, fn func(ctx context.Context, cfg *ProviderConfig) (StorageProvider, error)) {
	globalFactory.mu.Lock()
	defer globalFactory.mu.Unlock()
	globalFactory.providers[name] = fn
}

// NewFromConfig 根据驱动名创建 Provider
func NewFromConfig(ctx context.Context, driver string, cfg *ProviderConfig) (StorageProvider, error) {
	globalFactory.mu.RLock()
	defer globalFactory.mu.RUnlock()

	fn, ok := globalFactory.providers[driver]
	if !ok {
		return nil, fmt.Errorf("未知的存储驱动: %s，可用驱动: local/s3/oss/cos/obs", driver)
	}
	return fn(ctx, cfg)
}

// ProviderConfig 驱动配置
type ProviderConfig struct {
	// 通用字段
	Bucket   string
	Region   string
	Endpoint string
	BaseURL  string
	RootDir  string

	// 认证字段（从 site_configs 加密字段读取）
	AccessKey string
	SecretKey string

	// S3 特有
	PathPrefix     string
	ForcePathStyle bool // MinIO 需要开启

	// 阿里云 OSS 特有
	STSToken string // RAM 临时凭证

	// 自定义元数据
	Custom map[string]string
}
