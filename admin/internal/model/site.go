package model

import (
	"time"

	"github.com/google/uuid"
)

// Site 站点
type Site struct {
	ID            uuid.UUID `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	Name          string    `json:"name" gorm:"size:200;not null"`
	Slug          string    `json:"slug" gorm:"size:100;not null;uniqueIndex"`
	Description   string    `json:"description" gorm:"type:text"`
	LogoURL       *string   `json:"logo_url,omitempty" gorm:"column:logo_url;type:text"`
	FaviconURL    *string   `json:"favicon_url,omitempty" gorm:"column:favicon_url;type:text"`
	Config        JSONB     `json:"config" gorm:"type:jsonb;default:'{\"timezone\":\"Asia/Shanghai\",\"locale\":\"zh-CN\"}'"`
	SEO           JSONB     `json:"seo" gorm:"type:jsonb;default:'{}'"`
	CustomDomains JSONArray `json:"custom_domains" gorm:"type:jsonb;default:'[]'"`
	IsActive      bool      `json:"is_active" gorm:"default:true"`
	TenantID      *uuid.UUID `json:"tenant_id,omitempty" gorm:"type:uuid"`
	Plan          string    `json:"plan" gorm:"size:50;default:'free'"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty" gorm:"type:uuid"`
	CreatedTime     time.Time `json:"created_time" gorm:"type:timestamptz;autoCreateTime"`
	UpdatedTime     time.Time `json:"updated_time" gorm:"type:timestamptz;autoUpdateTime"`
	DeletedTime     *time.Time `json:"deleted_time,omitempty" gorm:"type:timestamptz;index"`
}

// TableName 表名
func (Site) TableName() string {
	return "sites"
}

// ============ DTO ============

// SiteCreate 创建站点请求
type SiteCreate struct {
	Name          string     `json:"name" binding:"required,min=1,max=200"`
	Slug          string     `json:"slug" binding:"required,min=1,max=100"`
	Description   string     `json:"description" binding:"max=2000"`
	LogoURL       *string    `json:"logo_url" binding:"omitempty,url"`
	FaviconURL    *string    `json:"favicon_url" binding:"omitempty,url"`
	Config        *JSONB     `json:"config"`
	SEO           *JSONB     `json:"seo"`
	CustomDomains *JSONArray `json:"custom_domains"`
	IsActive      *bool      `json:"is_active"`
	Plan          *string    `json:"plan" binding:"omitempty,oneof=free pro enterprise"`
}

// SiteUpdate 更新站点请求
type SiteUpdate struct {
	Name          *string    `json:"name" binding:"omitempty,min=1,max=200"`
	Slug          *string    `json:"slug" binding:"omitempty,min=1,max=100"`
	Description   *string    `json:"description" binding:"omitempty,max=2000"`
	LogoURL       *string    `json:"logo_url" binding:"omitempty,url"`
	FaviconURL    *string    `json:"favicon_url" binding:"omitempty,url"`
	Config        *JSONB     `json:"config"`
	SEO           *JSONB     `json:"seo"`
	CustomDomains *JSONArray `json:"custom_domains"`
	IsActive      *bool      `json:"is_active"`
	Plan          *string    `json:"plan" binding:"omitempty,oneof=free pro enterprise"`
}

// SiteResponse 站点响应
type SiteResponse struct {
	ID            uuid.UUID      `json:"id"`
	Name          string         `json:"name"`
	Slug          string         `json:"slug"`
	Description   string         `json:"description,omitempty"`
	LogoURL       *string        `json:"logo_url,omitempty"`
	FaviconURL    *string        `json:"favicon_url,omitempty"`
	Config        JSONB          `json:"config"`
	SEO           JSONB          `json:"seo"`
	CustomDomains JSONArray      `json:"custom_domains"`
	IsActive      bool           `json:"is_active"`
	TenantID      *uuid.UUID     `json:"tenant_id,omitempty"`
	Plan          string         `json:"plan"`
	CreatedBy     *uuid.UUID     `json:"created_by,omitempty"`
	CreatedTime     time.Time      `json:"created_time"`
	UpdatedTime     time.Time      `json:"updated_time"`
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
	return SiteResponse{
		ID:            s.ID,
		Name:          s.Name,
		Slug:          s.Slug,
		Description:   s.Description,
		LogoURL:       s.LogoURL,
		FaviconURL:    s.FaviconURL,
		Config:        s.Config,
		SEO:           s.SEO,
		CustomDomains: s.CustomDomains,
		IsActive:      s.IsActive,
		TenantID:      s.TenantID,
		Plan:          s.Plan,
		CreatedBy:     s.CreatedBy,
		CreatedTime:     s.CreatedTime,
		UpdatedTime:     s.UpdatedTime,
	}
}
