// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
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
	ExpiresTime   *time.Time `json:"expires_time"`
}

// APITokenUpdate 更新 API Token 请求
type APITokenUpdate struct {
	Name        *string    `json:"name" binding:"max=100"`
	Description *string    `json:"description" binding:"max=500"`
	ExpiresTime *time.Time `json:"expires_time"`
	Status      *string    `json:"status"` // "active" | "expired" | "revoked"
}

// EndpointPermission 权限配置（与前端 EndpointPermission 对应）
type EndpointPermission struct {
	ContentTypes []string `json:"content_types,omitempty"`
	Endpoints    []APIEndpoint `json:"endpoints,omitempty"`
}

// APIEndpoint API 端点
type APIEndpoint struct {
	Path   string   `json:"path"`
	Method []string `json:"method"`
}

// APIEndpointLimits 限流配置（与前端 APIEndpointLimits 对应）
type APIEndpointLimits struct {
	RequestsPerMinute int `json:"requests_per_minute"`
	RequestsPerDay     int `json:"requests_per_day"`
}

// APITokenResponse API Token 响应（与前端 APITokenResponse 对应）
type APITokenResponse struct {
	ID            uuid.UUID           `json:"id"`
	SiteID        uuid.UUID           `json:"site_id"`
	Name          string              `json:"name"`
	Description   string              `json:"description,omitempty"`
	TokenPrefix   string              `json:"token_prefix"`
	Permissions   EndpointPermission  `json:"permissions"`          // 权限范围（前端期望）
	RateLimits    APIEndpointLimits   `json:"rate_limits"`          // 限流配置（前端期望）
	ExpiresTime   *time.Time          `json:"expires_time,omitempty"`
	Status        TokenStatus         `json:"status"`
	LastUsedTime  *time.Time          `json:"last_used_time,omitempty"`
	RequestCount  int64               `json:"request_count"`
	CreatedBy     *uuid.UUID          `json:"created_by,omitempty"`
	CreatedTime   time.Time           `json:"created_time"`
	UpdatedTime   time.Time           `json:"updated_time"`
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

// ToResponse 转换为响应 DTO（与前端 api-token.ts 字段名一致）
func (t *APIToken) ToResponse() APITokenResponse {
	return APITokenResponse{
		ID:            t.ID,
		SiteID:        t.SiteID,
		Name:          t.Name,
		Description:   t.Description,
		TokenPrefix:   t.TokenPrefix,
		Permissions:   EndpointPermission{
			ContentTypes: []string(t.Scopes),
		},
		RateLimits: APIEndpointLimits{
			RequestsPerMinute: t.RateLimit / 60,
			RequestsPerDay:     t.RateLimit,
		},
		ExpiresTime:  t.ExpiresTime,
		Status:       t.Status,
		LastUsedTime: t.LastUsedTime,
		RequestCount: t.RequestCount,
		CreatedBy:    t.CreatedBy,
		CreatedTime:  t.CreatedTime,
		UpdatedTime:  t.UpdatedTime,
	}
}
