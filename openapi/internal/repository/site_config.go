// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SiteConfigRepository 站点配置数据访问层（只读）
type SiteConfigRepository struct {
	db *gorm.DB
}

// NewSiteConfigRepository 创建 SiteConfig Repository
func NewSiteConfigRepository(db *gorm.DB) *SiteConfigRepository {
	return &SiteConfigRepository{db: db}
}

// GetValue 获取站点配置值（仅 default 分组），返回空字符串表示未找到
func (r *SiteConfigRepository) GetValue(ctx context.Context, siteID uuid.UUID, key string) (string, error) {
	var configValue string
	err := r.db.WithContext(ctx).
		Table("site_configs").
		Select("config_value").
		Where("site_id = ? AND config_key = ? AND config_group = 'default'", siteID, key).
		Pluck("config_value", &configValue).Error
	if err != nil {
		return "", err
	}
	return configValue, nil
}
