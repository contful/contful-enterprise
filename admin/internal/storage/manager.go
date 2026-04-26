package storage

import (
	"context"

	"github.com/google/uuid"
)

// StorageConfigFunc 根据 siteID 返回 ProviderConfig 的函数（保留用于潜在扩展）。
type StorageConfigFunc func(ctx context.Context, siteID uuid.UUID) (*ProviderConfig, string, error)

// StorageManager 存储驱动管理器。
// 全局单例模式：从 config.yaml + 环境变量初始化，所有站点共用。
type StorageManager struct {
	provider StorageProvider
}

// NewGlobalManager 创建全局存储管理器（配置在 config.yaml，所有站点共用）。
func NewGlobalManager(provider StorageProvider) *StorageManager {
	return &StorageManager{provider: provider}
}

// ProviderFor 获取存储驱动（全局单例，直接返回）
func (m *StorageManager) ProviderFor(ctx context.Context, siteID uuid.UUID) (StorageProvider, error) {
	return m.provider, nil
}

// Invalidate 不再需要（全局模式配置变更需重启服务）
func (m *StorageManager) Invalidate(siteID uuid.UUID) {}

// InvalidateAll 不再需要（全局模式配置变更需重启服务）
func (m *StorageManager) InvalidateAll() {}

// defaultStorageConfig 默认配置（fallback）
func defaultStorageConfig(_ context.Context, _ uuid.UUID) (*ProviderConfig, string, error) {
	return &ProviderConfig{
		RootDir: "./uploads",
		BaseURL: "/assets",
	}, "local", nil
}
