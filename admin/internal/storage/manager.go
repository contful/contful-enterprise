package storage

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

// StorageManager 站点级存储驱动管理器。
// 每个站点可有独立存储配置（通过 site_configs 存储），Manager 按需初始化并缓存 Provider。
type StorageManager struct {
	cfgFunc StorageConfigFunc // 动态获取站点存储配置

	mu    sync.RWMutex
	cache map[uuid.UUID]StorageProvider // siteID → provider cache
	ttl   time.Duration                // 缓存有效期（0 = 永不过期）
}

// StorageConfigFunc 根据 siteID 返回 ProviderConfig 的函数。
// 由调用方注入，从 site_configs 读取动态配置。
type StorageConfigFunc func(ctx context.Context, siteID uuid.UUID) (*ProviderConfig, string, error)

// NewStorageManager 创建存储管理器。
// cfgFunc: 动态配置获取函数（从 site_configs 读取 storage.* 键值）
// ttl: provider 缓存有效期，0 = 不过期（推荐生产环境）
func NewStorageManager(cfgFunc StorageConfigFunc, ttl time.Duration) *StorageManager {
	return &StorageManager{
		cfgFunc: cfgFunc,
		cache:   make(map[uuid.UUID]StorageProvider),
		ttl:     ttl,
	}
}

// ProviderFor 获取指定站点的存储驱动（优先从缓存，缓存不存在时按需初始化）
func (m *StorageManager) ProviderFor(ctx context.Context, siteID uuid.UUID) (StorageProvider, error) {
	// 快速路径：读缓存
	m.mu.RLock()
	if m.ttl == 0 {
		if p, ok := m.cache[siteID]; ok {
			m.mu.RUnlock()
			return p, nil
		}
	}
	m.mu.RUnlock()

	// 慢路径：获取配置 + 初始化
	cfg, driver, err := m.cfgFunc(ctx, siteID)
	if err != nil {
		return nil, fmt.Errorf("获取存储配置失败: %w", err)
	}

	provider, err := NewFromConfig(ctx, driver, cfg)
	if err != nil {
		return nil, fmt.Errorf("初始化存储驱动 %s 失败: %w", driver, err)
	}

	// 写缓存
	m.mu.Lock()
	m.cache[siteID] = provider
	m.mu.Unlock()

	return provider, nil
}

// Invalidate 清除指定站点的 provider 缓存（配置变更后调用）
func (m *StorageManager) Invalidate(siteID uuid.UUID) {
	m.mu.Lock()
	delete(m.cache, siteID)
	m.mu.Unlock()
}

// InvalidateAll 清除所有 provider 缓存
func (m *StorageManager) InvalidateAll() {
	m.mu.Lock()
	m.cache = make(map[uuid.UUID]StorageProvider)
	m.mu.Unlock()
}

// defaultStorageConfig 默认配置（未配置 site_configs 时 fallback）
func defaultStorageConfig(_ context.Context, _ uuid.UUID) (*ProviderConfig, string, error) {
	return &ProviderConfig{
		RootDir: "/app/uploads",
		BaseURL: "http://localhost:9080/assets",
	}, "local", nil
}
