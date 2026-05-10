// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Site 站点（混合模式：固定列 + JSONB 动态配置）
type Site struct {
	ID        uuid.UUID      `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name      string           `json:"name" gorm:"size:200;not null"`
	Slug      string           `json:"slug" gorm:"size:100;not null;uniqueIndex"`
	Description string         `json:"description" gorm:"type:text"`
	SiteURL   *string         `json:"site_url,omitempty" gorm:"column:site_url;type:text"`

	// 混合模式：固定列（类型安全，可索引）
	Locale        string         `json:"locale" gorm:"type:varchar(20);default:zh-CN"`
	Timezone      string         `json:"timezone" gorm:"type:varchar(50);default:Asia/Shanghai"`
	SeoTitle      string         `json:"seo_title" gorm:"type:varchar(255)"`
	SeoDescription string        `json:"seo_description" gorm:"type:text"`
	SeoKeywords   datatypes.JSON `json:"seo_keywords" gorm:"type:jsonb;default:'[]'"`

	// 动态配置（JSONB，灵活扩展，无需改表）
	Settings   JSONB          `json:"settings" gorm:"column:settings;type:jsonb;default:'{}'"`

	IsActive  bool            `json:"is_active" gorm:"default:true"`
	CreatedBy *uuid.UUID      `json:"created_by,omitempty" gorm:"type:uuid"`
	CreatedTime time.Time      `json:"created_time" gorm:"type:timestamptz;autoCreateTime"`
	UpdatedTime time.Time      `json:"updated_time" gorm:"type:timestamptz;autoUpdateTime"`
	DeletedTime *time.Time     `json:"deleted_time,omitempty" gorm:"type:timestamptz;index"`
}

// TableName 表名
func (Site) TableName() string {
	return "sites"
}

// ============ DTO ============

// SiteCreate 创建站点请求
type SiteCreate struct {
	Name          string  `json:"name" binding:"required,min=1,max=200"`
	Slug          string  `json:"slug" binding:"required,min=1,max=100"`
	Description   *string `json:"description,omitempty"`
	SiteURL       *string `json:"site_url" binding:"omitempty,url"`
	Locale        *string `json:"locale"`
	Timezone      *string `json:"timezone"`
	SeoTitle      *string `json:"seo_title"`
	SeoDescription *string `json:"seo_description"`
	SeoKeywords   []string `json:"seo_keywords"`
	Settings      *JSONB  `json:"settings"`
	IsActive      *bool   `json:"is_active"`
}

// SiteUpdate 更新站点请求
type SiteUpdate struct {
	Name          *string `json:"name" binding:"omitempty,min=1,max=200"`
	Slug          *string `json:"slug" binding:"omitempty,min=1,max=100"`
	Description   *string `json:"description,omitempty"`
	SiteURL       *string `json:"site_url" binding:"omitempty,url"`
	Locale        *string `json:"locale"`
	Timezone      *string `json:"timezone"`
	SeoTitle      *string `json:"seo_title"`
	SeoDescription *string `json:"seo_description"`
	SeoKeywords   []string `json:"seo_keywords"`
	Settings      *JSONB  `json:"settings"`
	IsActive      *bool   `json:"is_active"`
}

// SiteResponse 站点响应
type SiteResponse struct {
	ID            uuid.UUID      `json:"id"`
	Name          string         `json:"name"`
	Slug          string         `json:"slug"`
	Description   string         `json:"description,omitempty"`
	SiteURL       *string        `json:"site_url,omitempty"`
	Locale        string         `json:"locale"`
	Timezone      string         `json:"timezone"`
	SeoTitle      string         `json:"seo_title"`
	SeoDescription string        `json:"seo_description,omitempty"`
	SeoKeywords   []string       `json:"seo_keywords"`
	Settings      JSONB          `json:"settings"`
	IsActive      bool           `json:"is_active"`
	CreatedBy     *uuid.UUID     `json:"created_by,omitempty"`
	CreatedTime   time.Time      `json:"created_time"`
	UpdatedTime   time.Time      `json:"updated_time"`
}

// SiteListResponse 站点列表响应
type SiteListResponse struct {
	Items    []SiteResponse `json:"items"`
	Total    int64          `json:"total"`
	Page     int            `json:"page"`
	PageSize int            `json:"page_size"`
}

// ToResponse 转换为响应 DTO
func (s *Site) ToResponse() SiteResponse {
	var keywords []string
	if len(s.SeoKeywords) > 0 {
		json.Unmarshal(s.SeoKeywords, &keywords)
	}
	return SiteResponse{
		ID:            s.ID,
		Name:          s.Name,
		Slug:          s.Slug,
		Description:   s.Description,
		SiteURL:       s.SiteURL,
		Locale:        s.Locale,
		Timezone:      s.Timezone,
		SeoTitle:      s.SeoTitle,
		SeoDescription: s.SeoDescription,
		SeoKeywords:   keywords,
		Settings:      s.Settings,
		IsActive:      s.IsActive,
		CreatedBy:     s.CreatedBy,
		CreatedTime:   s.CreatedTime,
		UpdatedTime:   s.UpdatedTime,
	}
}
