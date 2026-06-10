// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"time"

	"github.com/contful/contful/admin/pkg/uid"
)

// AuditLevel 审计日志级别
type AuditLevel string

const (
	AuditLevelDebug AuditLevel = "debug"
	AuditLevelInfo  AuditLevel = "info"
	AuditLevelWarn  AuditLevel = "warn"
	AuditLevelError AuditLevel = "error"
)

// AuditType 审计日志类型
type AuditType string

const (
	AuditTypeAuth    AuditType = "auth"
	AuditTypeContent AuditType = "content"
	AuditTypeMedia   AuditType = "media"
	AuditTypeSetting AuditType = "settings"
	AuditTypeUser    AuditType = "user"
	AuditTypeSystem  AuditType = "system"
)

// AuditLog 审计日志
type AuditLog struct {
	ID            uid.UID  `json:"id" gorm:"primary_key;default:gen_random_uuid()"`
	SiteID        *uid.UID `json:"site_id" gorm:"index"`
	UserID        *uid.UID `json:"user_id" gorm:"index"`
	Action        string     `json:"action" gorm:"type:varchar(100);not null"`
	ResourceType  string     `json:"resource_type" gorm:"type:varchar(100)"`
	ResourceID    *uid.UID `json:"resource_id" gorm:"type:uuid"`
	Level         AuditLevel `json:"level" gorm:"type:audit_level;not null;default:'info'"`
	Category      AuditType  `json:"category" gorm:"type:audit_type;not null;index"`
	Details       string     `json:"details" gorm:"type:text"`
	IPAddress     string     `json:"ip_address" gorm:"type:inet"`
	UserAgent     string     `json:"user_agent" gorm:"type:text"`
	CreatedTime   time.Time  `json:"created_time" gorm:"type:timestamptz;not null;default:now()"`
	DataSignature string    `json:"data_signature" gorm:"type:varchar(128);not null;default:''"` // HMAC 防篡改签名
}

func (AuditLog) TableName() string {
	return "contful_audit_logs"
}

// AuditLogFilter 审计日志筛选条件
type AuditLogFilter struct {
	SiteID       *uid.UID
	UserID       *uid.UID
	Action       string
	ResourceType string
	Category     AuditType
	Level        AuditLevel
	StartTime    time.Time
	EndTime      time.Time
}

// AuditLogListResponse 审计日志列表响应
type AuditLogListResponse struct {
	Items []AuditLog `json:"items"`
	Total int64      `json:"total"`
	Page  int        `json:"page"`
	PageSize int     `json:"page_size"`
}
