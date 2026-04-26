package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/contful/contful/admin/internal/model"
)

// SiteConfigRepository 站点配置仓储
type SiteConfigRepository struct {
	db *gorm.DB
}

// NewSiteConfigRepository 新建仓储
func NewSiteConfigRepository(db *gorm.DB) *SiteConfigRepository {
	return &SiteConfigRepository{db: db}
}

// Create 创建配置
func (r *SiteConfigRepository) Create(ctx context.Context, cfg *model.SiteConfig) error {
	return r.db.WithContext(ctx).Create(cfg).Error
}

// GetByID 根据 ID 获取
func (r *SiteConfigRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.SiteConfig, error) {
	var cfg model.SiteConfig
	err := r.db.WithContext(ctx).
		Where("id = ?", id).
		First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// GetByKey 根据站点和 Key 获取
func (r *SiteConfigRepository) GetByKey(ctx context.Context, siteID uuid.UUID, key string) (*model.SiteConfig, error) {
	var cfg model.SiteConfig
	err := r.db.WithContext(ctx).
		Where("site_id = ? AND config_key = ?", siteID, key).
		First(&cfg).Error
	if err != nil {
		return nil, err
	}
	return &cfg, nil
}

// ListBySite 列出站点的所有配置
func (r *SiteConfigRepository) ListBySite(ctx context.Context, siteID uuid.UUID) ([]model.SiteConfig, error) {
	var configs []model.SiteConfig
	err := r.db.WithContext(ctx).
		Where("site_id = ?", siteID).
		Order("config_group, config_key").
		Find(&configs).Error
	return configs, err
}

// ListByGroup 列出站点的指定分组配置
func (r *SiteConfigRepository) ListByGroup(ctx context.Context, siteID uuid.UUID, group string) ([]model.SiteConfig, error) {
	var configs []model.SiteConfig
	err := r.db.WithContext(ctx).
		Where("site_id = ? AND config_group = ?", siteID, group).
		Order("config_key").
		Find(&configs).Error
	return configs, err
}

// Update 更新配置
func (r *SiteConfigRepository) Update(ctx context.Context, cfg *model.SiteConfig) error {
	return r.db.WithContext(ctx).Save(cfg).Error
}

// Delete 删除配置
func (r *SiteConfigRepository) Delete(ctx context.Context, siteID uuid.UUID, key string) error {
	return r.db.WithContext(ctx).
		Where("site_id = ? AND config_key = ?", siteID, key).
		Delete(&model.SiteConfig{}).Error
}

// Upsert 插入或更新（key 存在则更新 value，不存在则创建）
func (r *SiteConfigRepository) Upsert(ctx context.Context, cfg *model.SiteConfig) error {
	return r.db.WithContext(ctx).
		Where("site_id = ? AND config_key = ?", cfg.SiteID, cfg.ConfigKey).
		Assign(model.SiteConfig{
			ConfigValue: cfg.ConfigValue,
			ConfigType:  cfg.ConfigType,
			ConfigGroup: cfg.ConfigGroup,
			IsEncrypted: cfg.IsEncrypted,
			Description: cfg.Description,
		}).
		FirstOrCreate(cfg).Error
}

// GetValue 获取配置值（返回原文或解密后值，需外部注入解密逻辑）
func (r *SiteConfigRepository) GetValue(ctx context.Context, siteID uuid.UUID, key string) (string, bool, error) {
	cfg, err := r.GetByKey(ctx, siteID, key)
	if err != nil {
		return "", false, err
	}
	return cfg.ConfigValue, cfg.IsEncrypted, nil
}
