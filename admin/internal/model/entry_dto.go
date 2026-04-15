package model

import (
	"time"

	"github.com/google/uuid"
)

// ============ Entry DTO ============

// EntryCreate 创建条目请求
type EntryCreate struct {
	ContentTypeID  uuid.UUID              `json:"content_type_id" binding:"required"`
	Locale         string                 `json:"locale"`
	Values         map[string]interface{} `json:"values"`
	SEOTitle       string                 `json:"seo_title"`
	SEODescription string                 `json:"seo_description"`
	SEOKeywords    []string               `json:"seo_keywords"`
	SortWeight     int                    `json:"sort_weight"`
}

// EntryUpdate 更新条目请求
type EntryUpdate struct {
	Locale         *string                `json:"locale"`
	Status         *EntryStatus           `json:"status"`
	Values         map[string]interface{} `json:"values"`
	SEOTitle       *string                `json:"seo_title"`
	SEODescription *string                `json:"seo_description"`
	SEOKeywords    []string               `json:"seo_keywords"`
	SortWeight     *int                   `json:"sort_weight"`
	ChangeSummary  string                 `json:"change_summary"` // 版本变更说明
}

// EntryPublish 发布条目请求
type EntryPublish struct {
	ChangeSummary string `json:"change_summary"` // 发布说明
}

// EntryResponse 条目响应
type EntryResponse struct {
	ID             uuid.UUID              `json:"id"`
	ContentTypeID  uuid.UUID              `json:"content_type_id"`
	SiteID         uuid.UUID              `json:"site_id"`
	Locale         string                 `json:"locale"`
	Status         EntryStatus            `json:"status"`
	Version        int                    `json:"version"`
	VersionHistory []EntryVersionInfo     `json:"version_history,omitempty"`
	PublishedAt    *time.Time             `json:"published_at,omitempty"`
	PublishedBy    *uuid.UUID             `json:"published_by,omitempty"`
	Relations      []map[string]interface{} `json:"relations,omitempty"`
	SEOTitle       string                 `json:"seo_title,omitempty"`
	SEODescription string                 `json:"seo_description,omitempty"`
	SEOKeywords    []string               `json:"seo_keywords,omitempty"`
	SortWeight     int                    `json:"sort_weight"`
	CreatedBy      *uuid.UUID             `json:"created_by,omitempty"`
	CreatedAt      time.Time              `json:"created_at"`
	UpdatedAt      time.Time              `json:"updated_at"`
	Values         map[string]interface{} `json:"values,omitempty"`
}

// EntryVersionInfo 版本信息
type EntryVersionInfo struct {
	Version       int       `json:"version"`
	CreatedBy     *uuid.UUID `json:"created_by,omitempty"`
	CreatedAt     time.Time `json:"created_at"`
	ChangeSummary string    `json:"change_summary,omitempty"`
}

// EntryListResponse 条目列表响应
type EntryListResponse struct {
	Items      []EntryResponse `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
}

// EntryListFilter 条目列表过滤条件
type EntryListFilter struct {
	ContentTypeID *uuid.UUID  `json:"content_type_id"`
	Status        *EntryStatus `json:"status"`
	Locale        *string     `json:"locale"`
	Keyword       *string     `json:"keyword"` // 搜索标题或内容
}

// ToResponse 转换为响应
func (e *Entry) ToResponse() EntryResponse {
	resp := EntryResponse{
		ID:             e.ID,
		ContentTypeID:  e.ContentTypeID,
		SiteID:         e.SiteID,
		Locale:         e.Locale,
		Status:         e.Status,
		Version:        e.Version,
		PublishedAt:    e.PublishedAt,
		PublishedBy:    e.PublishedBy,
		SEOTitle:       e.SEOTitle,
		SEODescription: e.SEODescription,
		SEOKeywords:    e.SEOKeywords,
		SortWeight:     e.SortWeight,
		CreatedBy:      e.CreatedBy,
		CreatedAt:      e.CreatedAt,
		UpdatedAt:      e.UpdatedAt,
		Values:         make(map[string]interface{}),
	}

	// 解析 version_history (JSONArray)
	if len(e.VersionHistory) > 0 {
		resp.VersionHistory = make([]EntryVersionInfo, 0, len(e.VersionHistory))
		for _, h := range e.VersionHistory {
			if m, ok := h.(map[string]interface{}); ok {
				vi := EntryVersionInfo{}
				if v, ok := m["version"].(float64); ok {
					vi.Version = int(v)
				}
				if v, ok := m["created_at"].(string); ok {
					vi.CreatedAt, _ = time.Parse(time.RFC3339, v)
				}
				if v, ok := m["change_summary"].(string); ok {
					vi.ChangeSummary = v
				}
				resp.VersionHistory = append(resp.VersionHistory, vi)
			}
		}
	}

	// 解析 relations
	if len(e.Relations) > 0 {
		resp.Relations = make([]map[string]interface{}, len(e.Relations))
		for i, r := range e.Relations {
			resp.Relations[i] = r
		}
	}

	// 解析字段值
	if len(e.Values) > 0 {
		for _, v := range e.Values {
			if v.Field != nil {
				resp.Values[v.Field.Name] = v.Value.Interface()
			}
		}
	}

	return resp
}

// EntryResponseWithType 带内容类型信息的响应
type EntryResponseWithType struct {
	EntryResponse
	ContentType *ContentTypeResponse `json:"content_type,omitempty"`
}

// ToResponseWithType 转换为带内容类型的响应
func (e *Entry) ToResponseWithType(ct *ContentType) EntryResponseWithType {
	return EntryResponseWithType{
		EntryResponse: e.ToResponse(),
		ContentType:   func() *ContentTypeResponse { r := ct.ToResponse(); return &r }(),
	}
}
