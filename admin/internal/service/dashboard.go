// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package service

import (
	"context"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/pkg/uid"
	"gorm.io/gorm"
)

// DashboardStats 仪表盘统计数据
type DashboardStats struct {
	Sites    int64 `json:"sites"`
	Schemas  int64 `json:"schemas,omitempty"`
	Entries  int64 `json:"entries,omitempty"`
	Assets   int64 `json:"assets"`
	Users    int64 `json:"users"`
	APITokens int64 `json:"api_tokens"`
}

// DashboardService 仪表盘统计服务（聚合查询，不依赖单个领域 service）
type DashboardService struct {
	db *gorm.DB
}

// NewDashboardService 新建仪表盘服务
func NewDashboardService(db *gorm.DB) *DashboardService {
	return &DashboardService{db: db}
}

// GetStats 获取仪表盘统计
// siteID 可为 nil（首次进入未创建站点时）
func (s *DashboardService) GetStats(ctx context.Context, siteID *uid.UID) *DashboardStats {
	stats := &DashboardStats{}

	// 全局统计（不依赖 X-Site-ID）
	s.db.WithContext(ctx).Model(&model.Site{}).Where("deleted_time IS NULL").Count(&stats.Sites)
	s.db.WithContext(ctx).Model(&model.SystemUser{}).Where("deleted_time IS NULL").Count(&stats.Users)
	s.db.WithContext(ctx).Model(&model.Asset{}).Count(&stats.Assets)
	s.db.WithContext(ctx).Model(&model.APIToken{}).Where("status = ?", model.TokenStatusActive).Count(&stats.APITokens)

	// 站点相关统计（仅当已选择站点时）
	if siteID != nil {
		s.db.WithContext(ctx).Model(&model.ContentSchema{}).Where("site_id = ?", *siteID).Count(&stats.Schemas)
		s.db.WithContext(ctx).Model(&model.Entry{}).Where("site_id = ?", *siteID).Count(&stats.Entries)
	}

	return stats
}
