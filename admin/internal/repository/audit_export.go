// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/contful/contful/admin/internal/model"
	"gorm.io/gorm"
)

type AuditExportRepository struct {
	db *gorm.DB
}

func NewAuditExportRepository(db *gorm.DB) *AuditExportRepository {
	return &AuditExportRepository{db: db}
}

func (r *AuditExportRepository) Create(ctx context.Context, report *model.AuditReportExport) error {
	return r.db.WithContext(ctx).Create(report).Error
}

func (r *AuditExportRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.AuditReportExport, error) {
	var report model.AuditReportExport
	if err := r.db.WithContext(ctx).First(&report, "id = ?", id).Error; err != nil {
		return nil, err
	}
	return &report, nil
}

func (r *AuditExportRepository) List(ctx context.Context, page, pageSize int) ([]model.AuditReportExport, int64, error) {
	var reports []model.AuditReportExport
	var total int64

	query := r.db.WithContext(ctx).Model(&model.AuditReportExport{})

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := query.Offset(offset).Limit(pageSize).Order("created_time DESC").Find(&reports).Error; err != nil {
		return nil, 0, err
	}

	return reports, total, nil
}

func (r *AuditExportRepository) Update(ctx context.Context, report *model.AuditReportExport) error {
	return r.db.WithContext(ctx).Save(report).Error
}

func (r *AuditExportRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.AuditReportExport{}, "id = ?", id).Error
}

// FindExpired 查找已过期的导出任务
func (r *AuditExportRepository) FindExpired(ctx context.Context) ([]model.AuditReportExport, error) {
	var reports []model.AuditReportExport
	err := r.db.WithContext(ctx).
		Where("expires_time < ?", time.Now()).
		Where("status IN ?", []string{"completed", "failed", "expired"}).
		Find(&reports).Error
	return reports, err
}

// FindStaleProcessing 查找超时未完成的任务（> 30 min）
func (r *AuditExportRepository) FindStaleProcessing(ctx context.Context) ([]model.AuditReportExport, error) {
	var reports []model.AuditReportExport
	err := r.db.WithContext(ctx).
		Where("status = ?", "processing").
		Where("created_time < ?", time.Now().Add(-30*time.Minute)).
		Find(&reports).Error
	return reports, err
}
