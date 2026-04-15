package model

import (
	"time"

	"github.com/google/uuid"
)

// ============ API Token DTO ============

// APITokenCreate 创建 API Token 请求
type APITokenCreate struct {
	Name        string              `json:"name" binding:"required,min=1,max=100"`
	Description string              `json:"description" binding:"max=500"`
	Permissions *EndpointPermission  `json:"permissions"`
	RateLimits  *APIEndpointLimits   `json:"rate_limits"`
	ExpiresAt   *time.Time          `json:"expires_at"`
}

// APITokenUpdate 更新 API Token 请求
type APITokenUpdate struct {
	Name        *string             `json:"name" binding:"max=100"`
	Description *string             `json:"description" binding:"max=500"`
	Permissions *EndpointPermission `json:"permissions"`
	RateLimits  *APIEndpointLimits   `json:"rate_limits"`
	ExpiresAt   *time.Time          `json:"expires_at"`
	Status      *TokenStatus        `json:"status"`
}

// APITokenResponse API Token 响应
type APITokenResponse struct {
	ID          uuid.UUID           `json:"id"`
	SiteID      uuid.UUID           `json:"site_id"`
	Name        string              `json:"name"`
	Description string              `json:"description,omitempty"`
	TokenPrefix string              `json:"token_prefix"`
	Permissions EndpointPermission  `json:"permissions"`
	RateLimits  APIEndpointLimits   `json:"rate_limits"`
	Usage       APIUsage            `json:"usage"`
	ExpiresAt   *time.Time         `json:"expires_at,omitempty"`
	Status      TokenStatus         `json:"status"`
	LastUsedAt  *time.Time         `json:"last_used_at,omitempty"`
	CreatedBy   *uuid.UUID         `json:"created_by,omitempty"`
	CreatedAt   time.Time          `json:"created_at"`
	UpdatedAt   time.Time          `json:"updated_at"`
}

// APITokenCreateResponse 创建 Token 响应（包含明文 Token）
type APITokenCreateResponse struct {
	APITokenResponse
	Token string `json:"token"` // 仅在创建时返回一次
}

// APITokenListResponse Token 列表响应
type APITokenListResponse struct {
	Items    []APITokenResponse `json:"items"`
	Total    int64              `json:"total"`
	Page     int                `json:"page"`
	PageSize int                `json:"page_size"`
}

// APITokenListFilter Token 列表过滤
type APITokenListFilter struct {
	Status *TokenStatus `json:"status"`
	Name   *string      `json:"name"` // 模糊搜索
}

// ToResponse 转换为响应
func (t *APIToken) ToResponse() APITokenResponse {
	return APITokenResponse{
		ID:          t.ID,
		SiteID:      t.SiteID,
		Name:        t.Name,
		Description: t.Description,
		TokenPrefix: t.TokenPrefix,
		Permissions: t.Permissions,
		RateLimits:  t.RateLimits,
		Usage:       t.Usage,
		ExpiresAt:   t.ExpiresAt,
		Status:      t.Status,
		LastUsedAt:  t.LastUsedAt,
		CreatedBy:   t.CreatedBy,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

// DefaultEndpointPermission 默认端点权限
func DefaultEndpointPermission() EndpointPermission {
	return EndpointPermission{
		ContentTypes: []string{"*"}, // 允许全部
		Endpoints: []APIEndpoint{
			{Path: "/content/*", Method: []string{"GET"}},
		},
	}
}

// DefaultRateLimits 默认速率限制
func DefaultRateLimits() APIEndpointLimits {
	return APIEndpointLimits{
		RequestsPerMinute: 60,
		RequestsPerDay:    10000,
	}
}

// EmptyUsage 空使用统计
func EmptyUsage() APIUsage {
	return APIUsage{
		RequestCount:     0,
		DailyRequestCount: 0,
		BandwidthUsed:    0,
	}
}
