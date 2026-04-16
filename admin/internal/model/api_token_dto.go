package model

import (
	"time"

	"github.com/google/uuid"
)

// ============ API Token DTO ============

// APITokenCreate 创建 API Token 请求
type APITokenCreate struct {
	Name        string     `json:"name" binding:"required,min=1,max=100"`
	Description string     `json:"description" binding:"max=500"`
	ExpiresAt   *time.Time `json:"expires_at"`
}

// APITokenUpdate 更新 API Token 请求
type APITokenUpdate struct {
	Name        *string    `json:"name" binding:"max=100"`
	Description *string    `json:"description" binding:"max=500"`
	ExpiresAt   *time.Time `json:"expires_at"`
	Status      *string    `json:"status"` // "active" | "expired" | "revoked"
}

// APITokenResponse API Token 响应
type APITokenResponse struct {
	ID            uuid.UUID  `json:"id"`
	SiteID        uuid.UUID  `json:"site_id"`
	Name          string     `json:"name"`
	Description   string     `json:"description,omitempty"`
	TokenPrefix   string     `json:"token_prefix"`
	Scopes        []string   `json:"scopes"`         // 权限范围
	SiteScope     []string   `json:"site_scope"`     // 站点范围
	ChannelScope  []string   `json:"channel_scope"`  // 频道范围
	AllowedIPs    *string    `json:"allowed_ips,omitempty"`
	RateLimit     int        `json:"rate_limit"`
	ExpiresAt     *time.Time `json:"expires_at,omitempty"`
	Status        TokenStatus `json:"status"`
	LastUsedAt    *time.Time `json:"last_used_at,omitempty"`
	LastUsedIP    *string    `json:"last_used_ip,omitempty"`
	RequestCount  int64      `json:"request_count"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// APITokenCreateResponse 创建 Token 响应（包含明文 Token，仅返回一次）
type APITokenCreateResponse struct {
	APITokenResponse
	Token string `json:"token"`
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

// ToResponse 转换为响应 DTO
func (t *APIToken) ToResponse() APITokenResponse {
	return APITokenResponse{
		ID:           t.ID,
		SiteID:       t.SiteID,
		Name:         t.Name,
		Description:  t.Description,
		TokenPrefix:  t.TokenPrefix,
		Scopes:       []string(t.Scopes),
		SiteScope:    []string(t.SiteScope),
		ChannelScope: []string(t.ChannelScope),
		AllowedIPs:   t.AllowedIPs,
		RateLimit:    t.RateLimit,
		ExpiresAt:    t.ExpiresAt,
		Status:       t.Status,
		LastUsedAt:   t.LastUsedAt,
		LastUsedIP:   t.LastUsedIP,
		RequestCount: t.RequestCount,
		CreatedBy:    t.CreatedBy,
		CreatedAt:    t.CreatedAt,
		UpdatedAt:    t.UpdatedAt,
	}
}
