// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
)

type AuditExportService struct {
	exportRepo *repository.AuditExportRepository
	auditRepo  *repository.AuditRepository
	csvService *AuditService // reuse existing CSV generation
	exportDir  string
}

func NewAuditExportService(
	exportRepo *repository.AuditExportRepository,
	auditRepo *repository.AuditRepository,
	csvService *AuditService,
) *AuditExportService {
	dir := "uploads/exports"
	os.MkdirAll(dir, 0755)
	return &AuditExportService{
		exportRepo: exportRepo,
		auditRepo:  auditRepo,
		csvService: csvService,
		exportDir:  dir,
	}
}

// CreateExport 创建导出任务并异步执行
func (s *AuditExportService) CreateExport(ctx context.Context, filter *model.AuditLogFilter, maxRows int, format string, createdBy uuid.UUID) (*model.AuditReportExport, error) {
	filterJSON, _ := json.Marshal(filter)

	report := &model.AuditReportExport{
		ID:         uuid.New(),
		Status:     model.ReportStatusPending,
		FilterJSON: string(filterJSON),
		Format:     model.ReportFormat(format),
		CreatedBy:  &createdBy,
		CreatedTime: time.Now(),
		ExpiresTime: time.Now().Add(7 * 24 * time.Hour),
	}

	if err := s.exportRepo.Create(ctx, report); err != nil {
		return nil, fmt.Errorf("create export: %w", err)
	}

	// 异步执行导出
	go s.executeExport(report.ID, filter, maxRows, format)

	return report, nil
}

func (s *AuditExportService) executeExport(reportID uuid.UUID, filter *model.AuditLogFilter, maxRows int, format string) {
	ctx := context.Background()

	// 标记 processing
	_ = s.updateStatus(reportID, model.ReportStatusProcessing, nil, 0, 0)

	csvBytes, count, total, err := s.csvService.ExportCSV(ctx, filter, maxRows)
	if err != nil {
		_ = s.updateStatus(reportID, model.ReportStatusFailed, &err, 0, total)
		log.Error().Err(err).Str("report_id", reportID.String()).Msg("导出失败")
		return
	}

	// 写入文件
	filename := fmt.Sprintf("%s.csv", reportID.String())
	filepath := fmt.Sprintf("%s/%s", s.exportDir, filename)
	if err := os.WriteFile(filepath, csvBytes, 0644); err != nil {
		_ = s.updateStatus(reportID, model.ReportStatusFailed, &err, count, total)
		log.Error().Err(err).Str("report_id", reportID.String()).Msg("写入文件失败")
		return
	}

	completed := time.Now()
	fileSize := int64(len(csvBytes))

	_ = s.exportRepo.Update(ctx, &model.AuditReportExport{
		ID:            reportID,
		Status:        model.ReportStatusCompleted,
		FilePath:      filepath,
		FileSize:      fileSize,
		RecordCount:   count,
		TotalCount:    total,
		CompletedTime: &completed,
	})

	log.Info().Str("report_id", reportID.String()).Int64("count", count).Msg("导出完成")
}

func (s *AuditExportService) updateStatus(id uuid.UUID, status model.ReportStatus, err *error, count, total int64) error {
	ctx := context.Background()
	report := &model.AuditReportExport{
		ID:          id,
		Status:      status,
		RecordCount: count,
		TotalCount:  total,
	}
	if err != nil && *err != nil {
		report.ErrorMsg = (*err).Error()
	}
	if status == model.ReportStatusCompleted || status == model.ReportStatusFailed {
		now := time.Now()
		report.CompletedTime = &now
	}
	return s.exportRepo.Update(ctx, report)
}

func (s *AuditExportService) GetByID(ctx context.Context, id uuid.UUID) (*model.AuditReportExport, error) {
	return s.exportRepo.GetByID(ctx, id)
}

func (s *AuditExportService) List(ctx context.Context, page, pageSize int) ([]model.AuditReportExport, int64, error) {
	return s.exportRepo.List(ctx, page, pageSize)
}

func (s *AuditExportService) Delete(ctx context.Context, id uuid.UUID) error {
	report, err := s.exportRepo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// 删除本地文件
	if report.FilePath != "" {
		os.Remove(report.FilePath)
	}

	return s.exportRepo.Delete(ctx, id)
}

// CleanupExpired 清理过期导出任务（Cron 调用）
func (s *AuditExportService) CleanupExpired(ctx context.Context) (int, error) {
	reports, err := s.exportRepo.FindExpired(ctx)
	if err != nil {
		return 0, err
	}

	count := 0
	for _, r := range reports {
		if r.FilePath != "" {
			os.Remove(r.FilePath)
		}
		if err := s.exportRepo.Delete(ctx, r.ID); err != nil {
			log.Error().Err(err).Str("id", r.ID.String()).Msg("清理过期导出失败")
			continue
		}
		count++
	}
	return count, nil
}

// MarkStaleFailed 标记超时未完成的 processing 任务为 failed
func (s *AuditExportService) MarkStaleFailed(ctx context.Context) (int, error) {
	reports, err := s.exportRepo.FindStaleProcessing(ctx)
	if err != nil {
		return 0, err
	}

	for _, r := range reports {
		msg := "timeout after 30 minutes"
		r.Status = model.ReportStatusFailed
		r.ErrorMsg = msg
		r.CompletedTime = timePtr(time.Now())
		if err := s.exportRepo.Update(ctx, &r); err != nil {
			log.Error().Err(err).Str("id", r.ID.String()).Msg("标记超时任务失败")
			continue
		}
	}
	return len(reports), nil
}

func timePtr(t time.Time) *time.Time { return &t }
