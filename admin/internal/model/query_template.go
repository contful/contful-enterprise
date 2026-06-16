// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"encoding/json"
	"time"

	"github.com/contful/contful-enterprise/shared/uid"
)

// AuditQueryTemplate 查询模板
type AuditQueryTemplate struct {
	ID          uid.UID         `json:"id" gorm:"primary_key;default:gen_random_uuid()"`
	Name        string          `json:"name" gorm:"type:varchar(200);not null"`
	Conditions  json.RawMessage `json:"conditions" gorm:"type:jsonb;not null"`
	CreatedBy   *uid.UID        `json:"created_by" gorm:"index"`
	CreatedTime time.Time       `json:"created_time" gorm:"type:timestamptz;not null;default:now()"`
}

func (AuditQueryTemplate) TableName() string {
	return "contful_audit_query_templates"
}

// AuditTemplateConditions 查询模板条件结构（Conditions 字段的反序列化目标）
type AuditTemplateConditions struct {
	Keyword    string `json:"keyword,omitempty"`
	Combinator string `json:"combinator,omitempty"` // "and" | "or"
	TimePreset string `json:"time_preset,omitempty"` // "1h" | "24h" | "7d"
	Category   string `json:"category,omitempty"`
	Level      string `json:"level,omitempty"`
	Action     string `json:"action,omitempty"`
	StartTime  string `json:"start_time,omitempty"`
	EndTime    string `json:"end_time,omitempty"`
}

// TemplateApplyResponse 模板应用响应
type TemplateApplyResponse struct {
	Template *AuditQueryTemplate   `json:"template"`
	Results  *AuditLogListResponse `json:"results"`
}
