package storage

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/spf13/viper"
)

// ProviderFactory 存储驱动工厂
type ProviderFactory struct {
	mu         sync.RWMutex
	providers  map[string]func(ctx context.Context, cfg *ProviderConfig) (StorageProvider, error)
	defaultCfg *ProviderConfig
}

var globalFactory = &ProviderFactory{
	providers: make(map[string]func(ctx context.Context, cfg *ProviderConfig) (StorageProvider, error)),
}

func init() {
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

	// 上传限制
	MaxUploadSize int64 // bytes

	// 认证字段（从环境变量读取）
	AccessKey string
	SecretKey string

	// S3 特有
	PathPrefix     string
	ForcePathStyle bool // MinIO 需要开启

	// 自定义元数据
	Custom map[string]string
}

// NewFromViper 从 viper 配置（config.yaml）和环境变量创建 StorageProvider。
// 这是推荐的全局初始化方式，不依赖 site_configs 表。
func NewFromViper(ctx context.Context) (StorageProvider, *ProviderConfig, error) {
	driver := viper.GetString("storage.driver")
	if driver == "" {
		driver = "local"
	}

	cfg := &ProviderConfig{
		RootDir:        viper.GetString("storage.upload_dir"),
		MaxUploadSize:  viper.GetInt64("storage.max_upload_size_mb") * 1024 * 1024,
		Bucket:         viper.GetString(fmt.Sprintf("storage.%s.bucket", driver)),
		Endpoint:       viper.GetString(fmt.Sprintf("storage.%s.endpoint", driver)),
		Region:         viper.GetString(fmt.Sprintf("storage.%s.region", driver)),
		BaseURL:        viper.GetString(fmt.Sprintf("storage.%s.base_url", driver)),
		PathPrefix:     viper.GetString("storage.s3.path_prefix"),
		ForcePathStyle: viper.GetBool("storage.s3.force_path_style"),
		AccessKey:      os.Getenv(fmt.Sprintf("STORAGE_%s_ACCESS_KEY", normalizeEnvKey(driver))),
		SecretKey:      os.Getenv(fmt.Sprintf("STORAGE_%s_SECRET_KEY", normalizeEnvKey(driver))),
	}

	if cfg.RootDir == "" {
		cfg.RootDir = "./uploads"
	}

	provider, err := NewFromConfig(ctx, driver, cfg)
	if err != nil {
		return nil, nil, err
	}
	return provider, cfg, nil
}

// normalizeEnvKey 将驱动名转为环境变量大写下划线格式
// 例如: "s3" -> "S3", "oss" -> "OSS", "cos" -> "COS"
func normalizeEnvKey(driver string) string {
	// 特殊处理 cos -> COS（本身就是大写）
	switch driver {
	case "cos":
		return "COS"
	case "oss", "obs", "s3":
		return driver
	default:
		return driver
	}
}
