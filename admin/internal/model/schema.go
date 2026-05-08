// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"time"

	"github.com/google/uuid"
)

// ContentSchemaKind 内容模型类型
type ContentSchemaKind string

const (
	ContentSchemaKindCollection ContentSchemaKind = "collection" // 集合类型（多条目）
	ContentSchemaKindSingle    ContentSchemaKind = "single"      // 单条类型
)

// ContentSchema 内容模型
type ContentSchema struct {
	ID                   uuid.UUID        `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SiteID               uuid.UUID        `json:"site_id" gorm:"type:uuid;not null;index"`
	Name                 string           `json:"name" gorm:"size:200;not null"`
	Slug                 string           `json:"slug" gorm:"size:100;not null;index"`
	Description          string           `json:"description" gorm:"type:text"`
	Kind                 ContentSchemaKind  `json:"kind" gorm:"type:schema_kind;not null;default:'collection'"`
	DisplayConfig        JSONB            `json:"display_config" gorm:"type:jsonb;default:'{}'"`
	APISConfig           JSONB            `json:"api_config" gorm:"column:api_config;type:jsonb;default:'{\"publicRead\":false,\"publicWrite\":false}'"`
	PreviewConfig        JSONB            `json:"preview_config" gorm:"type:jsonb;default:'{}'"`
	VersioningEnabled    bool             `json:"versioning_enabled" gorm:"default:false"`
	DraftAutosaveInterval *int            `json:"draft_autosave_interval" gorm:"default:null"`
	SignatureEnabled     bool             `json:"signature_enabled" gorm:"default:false"` // 是否启用数据签名
	IsActive             bool             `json:"is_active" gorm:"default:true"`
	SortOrder            int              `json:"sort_order" gorm:"default:0"`
	CreatedBy            *uuid.UUID       `json:"created_by" gorm:"type:uuid"`
	CreatedTime            time.Time        `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime            time.Time        `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedTime            *time.Time       `json:"deleted_time" gorm:"index"`

	// 关联
	Fields []Field `json:"fields,omitempty" gorm:"foreignKey:schema_id;references:ID"`
}

// TableName 表名
func (ContentSchema) TableName() string {
	return "schemas"
}

// Field 字段定义
type Field struct {
	ID                  uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ContentSchemaID       uuid.UUID  `json:"schema_id" gorm:"column:schema_id;type:uuid;not null;index"`
	Name                string     `json:"name" gorm:"size:100;not null"`
	Label               string     `json:"label" gorm:"size:200;not null"`
	Description         string     `json:"description" gorm:"type:text"`
	FieldType           string     `json:"field_type" gorm:"type:field_type;not null"`
	Config              JSONB      `json:"config" gorm:"type:jsonb;default:'{}'"`
	Validation          JSONB      `json:"validation" gorm:"type:jsonb;default:'{}'"`
	Display             JSONB      `json:"display" gorm:"type:jsonb;default:'{}'"`
	DefaultValue        *JSONB     `json:"default_value" gorm:"type:jsonb"`
	SortOrder           int        `json:"sort_order" gorm:"default:0"`
	ConditionalDisplay  *JSONB      `json:"conditional_display" gorm:"type:jsonb"`
	CreatedTime           time.Time  `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime           time.Time  `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedTime           *time.Time `json:"deleted_time" gorm:"index"`
}

// TableName 表名
func (Field) TableName() string {
	return "fields"
}

// FieldType 字段类型枚举
var FieldTypes = []string{
	"text",      // 单行文本
	"rich_text", // 富文本
	"number",    // 数字
	"boolean",   // 布尔值
	"date",      // 日期
	"datetime",  // 日期时间
	"email",     // 邮箱
	"url",       // URL
	"json",      // JSON
	"media",     // 媒体（图片/文件）
	"relation",  // 关联（指向其他内容模型）
	"enum",      // 枚举
	"password",  // 密码
}
