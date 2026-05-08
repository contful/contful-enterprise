// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"errors"

	"github.com/contful/contful/openapi/internal/repository"
	"github.com/google/uuid"
)

// ConfigService 站点配置服务（仅 default 分组，用于单页面内容配置）
type ConfigService struct {
	repo *repository.SiteConfigRepository
}

// NewConfigService 创建 ConfigService
func NewConfigService(repo *repository.SiteConfigRepository) *ConfigService {
	return &ConfigService{repo: repo}
}

// ErrConfigNotFound 配置不存在
var ErrConfigNotFound = errors.New("config not found")

// GetValue 获取站点配置值（仅 default 分组）
// 用于单页面内容（如官网关于我们）的 JSON 配置
func (s *ConfigService) GetValue(ctx context.Context, siteID uuid.UUID, key string) (string, error) {
	value, err := s.repo.GetValue(ctx, siteID, key)
	if err != nil {
		return "", err
	}
	if value == "" {
		return "", ErrConfigNotFound
	}
	return value, nil
}
