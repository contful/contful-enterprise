// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// ConfigService 站点配置服务（从 sites.settings JSONB 读取）
type ConfigService struct {
	db *gorm.DB
}

// NewConfigService 创建 ConfigService
func NewConfigService(db *gorm.DB) *ConfigService {
	return &ConfigService{db: db}
}

// ErrConfigNotFound 配置不存在
var ErrConfigNotFound = errors.New("config not found")

// GetValue 从 sites.settings JSONB 中获取指定 key 的值
func (s *ConfigService) GetValue(ctx context.Context, siteID uuid.UUID, key string) (string, error) {
	var value string
	// PostgreSQL: 从 JSONB 列中提取 key 的值
	err := s.db.WithContext(ctx).
		Raw("SELECT settings->? FROM sites WHERE id = ?", key, siteID).
		Scan(&value).Error
	if err != nil {
		return "", err
	}
	if value == "" || value == "null" {
		return "", ErrConfigNotFound
	}
	return value, nil
}
