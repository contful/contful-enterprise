// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"time"

	"github.com/google/uuid"
)

// ReportStatus 导出任务状态
type ReportStatus string

const (
	ReportStatusPending    ReportStatus = "pending"
	ReportStatusProcessing ReportStatus = "processing"
	ReportStatusCompleted  ReportStatus = "completed"
	ReportStatusFailed     ReportStatus = "failed"
	ReportStatusExpired    ReportStatus = "expired"
)

// ReportFormat 导出格式
type ReportFormat string

const (
	ReportFormatCSV  ReportFormat = "csv"
	ReportFormatXLSX ReportFormat = "xlsx"
)

// AuditReportExport 审计报告导出任务
type AuditReportExport struct {
	ID            uuid.UUID    `json:"id"             gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Status        ReportStatus `json:"status"         gorm:"type:varchar(20);not null;default:'pending'"`
	FilterJSON    string       `json:"filter_json"    gorm:"type:jsonb;not null;default:'{}'"`
	Format        ReportFormat `json:"format"         gorm:"type:varchar(10);not null;default:'csv'"`
	FilePath      string       `json:"file_path"      gorm:"type:varchar(500)"`
	FileSize      int64        `json:"file_size"`
	RecordCount   int64        `json:"record_count"`
	TotalCount    int64        `json:"total_count"`
	ErrorMsg      string       `json:"error_msg"      gorm:"type:text"`
	ExpiresTime   time.Time    `json:"expires_time"   gorm:"type:timestamptz;not null"`
	CreatedBy     *uuid.UUID   `json:"created_by"     gorm:"type:uuid"`
	CreatedTime   time.Time    `json:"created_time"   gorm:"type:timestamptz;not null;default:now()"`
	CompletedTime *time.Time   `json:"completed_time" gorm:"type:timestamptz"`
}

func (AuditReportExport) TableName() string {
	return "ent_audit_report_exports"
}

// AuditReportExportListResponse 导出任务列表响应
type AuditReportExportListResponse struct {
	Items    []AuditReportExport `json:"items"`
	Total    int64               `json:"total"`
	Page     int                 `json:"page"`
	PageSize int                 `json:"page_size"`
}
