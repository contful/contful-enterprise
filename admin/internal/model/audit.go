package model

import "github.com/google/uuid"

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
	ID           uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SiteID       *uuid.UUID `json:"site_id" gorm:"type:uuid;index"`
	UserID       *uuid.UUID `json:"user_id" gorm:"type:uuid;index"`
	Action       string     `json:"action" gorm:"type:varchar(100);not null"`
	ResourceType string     `json:"resource_type" gorm:"type:varchar(100)"`
	ResourceID   *uuid.UUID `json:"resource_id" gorm:"type:uuid"`
	Level        AuditLevel `json:"level" gorm:"type:audit_level;not null;default:'info'"`
	Category     AuditType  `json:"category" gorm:"type:audit_type;not null;index"`
	Details      string     `json:"details" gorm:"type:jsonb"`
	IPAddress    string     `json:"ip_address" gorm:"type:inet"`
	UserAgent    string     `json:"user_agent" gorm:"type:text"`
	CreatedTime    string     `json:"created_time" gorm:"type:timestamptz;not null;default:now()"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}
