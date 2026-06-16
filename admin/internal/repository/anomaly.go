// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"
	"time"

	"github.com/contful/contful-enterprise/shared/uid"
	"github.com/contful/contful/admin/internal/model"
	"gorm.io/gorm"
)

type AnomalyRepository struct {
	db *gorm.DB
}

func NewAnomalyRepository(db *gorm.DB) *AnomalyRepository {
	return &AnomalyRepository{db: db}
}

// Create 创建异常事件
func (r *AnomalyRepository) Create(ctx context.Context, anomaly *model.AuditAnomaly) error {
	return r.db.WithContext(ctx).Create(anomaly).Error
}

// BatchCreate 批量创建异常事件
func (r *AnomalyRepository) BatchCreate(ctx context.Context, anomalies []model.AuditAnomaly) error {
	if len(anomalies) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Create(&anomalies).Error
}

// GetByID 根据 ID 获取异常事件
func (r *AnomalyRepository) GetByID(ctx context.Context, id uid.UID) (*model.AuditAnomaly, error) {
	var anomaly model.AuditAnomaly
	if err := r.db.WithContext(ctx).First(&anomaly, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &anomaly, nil
}

// List 查询异常事件列表（支持筛选和分页）
func (r *AnomalyRepository) List(ctx context.Context, filter *model.AnomalyFilter, page, pageSize int) ([]model.AuditAnomaly, int64, error) {
	var anomalies []model.AuditAnomaly
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AuditAnomaly{})

	if filter != nil {
		if filter.AnomalyType != "" {
			query = query.Where("anomaly_type = ?", filter.AnomalyType)
		}
		if filter.Severity != "" {
			query = query.Where("severity = ?", filter.Severity)
		}
		if !filter.StartTime.IsZero() {
			query = query.Where("detected_time >= ?", filter.StartTime)
		}
		if !filter.EndTime.IsZero() {
			query = query.Where("detected_time <= ?", filter.EndTime)
		}
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("detected_time DESC").Find(&anomalies).Error; err != nil {
		return nil, 0, err
	}

	return anomalies, total, nil
}

// Summary 获取异常概览统计
func (r *AnomalyRepository) Summary(ctx context.Context) (*model.AnomalySummaryResponse, error) {
	summary := &model.AnomalySummaryResponse{}

	// 总数
	r.db.WithContext(ctx).Model(&model.AuditAnomaly{}).Count(&summary.TotalAnomalies)

	// 按类型统计
	var typeStats []struct {
		AnomalyType model.AnomalyType
		Count       int64
	}
	r.db.WithContext(ctx).Model(&model.AuditAnomaly{}).
		Select("anomaly_type, COUNT(*) as count").
		Group("anomaly_type").
		Find(&typeStats)
	summary.ByType = make(map[model.AnomalyType]int64)
	for _, s := range typeStats {
		summary.ByType[s.AnomalyType] = s.Count
	}

	// 按严重程度统计
	var severityStats []struct {
		Severity model.AnomalySeverity
		Count    int64
	}
	r.db.WithContext(ctx).Model(&model.AuditAnomaly{}).
		Select("severity, COUNT(*) as count").
		Group("severity").
		Find(&severityStats)
	summary.BySeverity = make(map[model.AnomalySeverity]int64)
	for _, s := range severityStats {
		summary.BySeverity[s.Severity] = s.Count
	}

	// 7 日趋势
	sevenDaysAgo := time.Now().Add(-7 * 24 * time.Hour)
	var trend []model.AnomalyTrendPoint
	r.db.WithContext(ctx).Model(&model.AuditAnomaly{}).
		Select("DATE(detected_time) as date, COUNT(*) as count").
		Where("detected_time >= ?", sevenDaysAgo).
		Group("DATE(detected_time)").
		Order("date ASC").
		Find(&trend)
	summary.Trend = trend

	// 最新 5 条异常
	var latest []model.AuditAnomaly
	r.db.WithContext(ctx).Model(&model.AuditAnomaly{}).
		Order("detected_time DESC").
		Limit(5).
		Find(&latest)
	summary.LatestAnomalies = latest

	// 最后扫描时间
	var lastScan model.AuditAnomaly
	if err := r.db.WithContext(ctx).Model(&model.AuditAnomaly{}).
		Order("created_time DESC").
		First(&lastScan).Error; err == nil {
		summary.LastScanTime = &lastScan.CreatedTime
	}

	return summary, nil
}
