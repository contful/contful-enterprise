// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// EntryStatus 条目状态
type EntryStatus string

const (
	EntryStatusDraft     EntryStatus = "draft"     // 草稿
	EntryStatusPublished EntryStatus = "published" // 已发布
	EntryStatusArchived  EntryStatus = "archived"  // 已归档
)

// Entry 内容条目
type Entry struct {
	ID             uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	ContentSchemaID  uuid.UUID  `json:"schema_id" gorm:"column:schema_id;type:uuid;not null;index"`
	SiteID         uuid.UUID  `json:"site_id" gorm:"type:uuid;not null;index"`
	Locale         string     `json:"locale" gorm:"size:20;not null;default:'zh-CN'"`
	Status         EntryStatus `json:"status" gorm:"type:entry_status;not null;default:'draft'"`
	Version        int        `json:"version" gorm:"not null;default:1"`
	VersionHistory JSONArray  `json:"version_history" gorm:"type:jsonb"`
	PublishedTime  *time.Time `json:"published_time" gorm:"column:published_time;type:timestamptz"`
	PublishedBy    *uuid.UUID `json:"published_by" gorm:"type:uuid"`
	Relations      JSONBSlice `json:"relations" gorm:"type:jsonb;default:'[]'"`
	SEOTitle       string     `json:"seo_title" gorm:"size:255"`
	SEODescription string     `json:"seo_description" gorm:"type:text"`
	SEOKeywords    []string   `json:"seo_keywords" gorm:"type:text[]"`
	SortWeight     int        `json:"sort_weight" gorm:"not null;default:0"`
	CreatedBy      *uuid.UUID `json:"created_by" gorm:"type:uuid"`
	CreatedTime      time.Time  `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime      time.Time  `json:"updated_time" gorm:"autoUpdateTime"`
	DeletedAt        gorm.DeletedAt `json:"deleted_time" gorm:"column:deleted_time;index"`
	DataSignature    JSONB      `json:"data_signature" gorm:"type:jsonb"` // 数据完整性签名

	// 关联
	ContentSchema *ContentSchema    `json:"content_schema,omitempty" gorm:"foreignKey:schema_id;references:ID"`
	Values      []EntryValue    `json:"values,omitempty" gorm:"foreignKey:EntryID;references:ID"`
}

// TableName 表名
func (Entry) TableName() string {
	return "contful_entries"
}

// EntryValue 内容字段值
type EntryValue struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EntryID     uuid.UUID  `json:"entry_id" gorm:"type:uuid;not null;index"`
	FieldID     uuid.UUID  `json:"field_id" gorm:"type:uuid;not null;index"`
	Value       JSONB      `json:"value" gorm:"type:jsonb;not null"`
	TextValue   *string    `json:"text_value,omitempty" gorm:"type:text"`
	NumberValue *float64   `json:"number_value,omitempty" gorm:"type:numeric"`
	BoolValue   *bool      `json:"bool_value,omitempty" gorm:"type:boolean"`
	DateValue   *time.Time `json:"date_value,omitempty" gorm:"type:date"`
	DatetimeValue *time.Time `json:"datetime_value,omitempty" gorm:"type:timestamptz"`
	CreatedTime   time.Time  `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime   time.Time  `json:"updated_time" gorm:"autoUpdateTime"`
	DataSignature JSONB      `json:"data_signature" gorm:"type:jsonb"` // 联动签名

	// 关联
	Field *Field `json:"field,omitempty" gorm:"foreignKey:FieldID;references:ID"`
}

// TableName 表名
func (EntryValue) TableName() string {
	return "contful_entry_values"
}

// EntryVersion 内容版本历史
type EntryVersion struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	EntryID       uuid.UUID `json:"entry_id" gorm:"type:uuid;not null;index"`
	Version       int       `json:"version" gorm:"not null"`
	ValuesSnapshot JSONB    `json:"values_snapshot" gorm:"type:jsonb;not null"`
	CreatedBy     *uuid.UUID `json:"created_by" gorm:"type:uuid"`
	CreatedTime     time.Time `json:"created_time" gorm:"autoCreateTime"`
	ChangeSummary string    `json:"change_summary" gorm:"type:text"`
}

// TableName 表名
func (EntryVersion) TableName() string {
	return "contful_entry_versions"
}
