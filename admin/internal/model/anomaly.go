// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"encoding/json"
	"time"

	"github.com/contful/contful-enterprise/shared/uid"
)

// AuditAnomaly 异常事件
type AuditAnomaly struct {
	ID             uid.UID          `json:"id" gorm:"primary_key;default:gen_random_uuid()"`
	AuditLogID     *uid.UID         `json:"audit_log_id" gorm:"index"`
	AnomalyType    AnomalyType      `json:"anomaly_type" gorm:"type:varchar(50);not null;index"`
	Severity       AnomalySeverity  `json:"severity" gorm:"type:varchar(20);not null"`
	Score          float64          `json:"score" gorm:"type:decimal(5,2);not null"`
	BaselineValue  json.RawMessage  `json:"baseline_value" gorm:"type:jsonb"`
	ActualValue    json.RawMessage  `json:"actual_value" gorm:"type:jsonb"`
	Description    string           `json:"description" gorm:"type:text;not null"`
	DetectedTime   time.Time        `json:"detected_time" gorm:"type:timestamptz;not null;index"`
	ResolvedTime   *time.Time       `json:"resolved_time,omitempty" gorm:"type:timestamptz"`
	ResolutionNote *string          `json:"resolution_note,omitempty" gorm:"type:text"`
	CreatedTime    time.Time        `json:"created_time" gorm:"type:timestamptz;not null;default:now()"`
}

func (AuditAnomaly) TableName() string {
	return "contful_audit_anomalies"
}

// AnomalyFilter 异常事件筛选条件
type AnomalyFilter struct {
	AnomalyType AnomalyType
	Severity    AnomalySeverity
	StartTime   time.Time
	EndTime     time.Time
}

// AnomalySummaryResponse 异常概览响应
type AnomalySummaryResponse struct {
	TotalAnomalies  int64                    `json:"total_anomalies"`
	ByType          map[AnomalyType]int64    `json:"by_type"`
	BySeverity      map[AnomalySeverity]int64 `json:"by_severity"`
	Trend           []AnomalyTrendPoint      `json:"trend"`
	LatestAnomalies []AuditAnomaly           `json:"latest_anomalies"`
	LastScanTime    *time.Time               `json:"last_scan_time"`
}

// AnomalyTrendPoint 异常趋势数据点
type AnomalyTrendPoint struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// AnomalyListResponse 异常事件列表响应
type AnomalyListResponse struct {
	Items    []AuditAnomaly `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}
