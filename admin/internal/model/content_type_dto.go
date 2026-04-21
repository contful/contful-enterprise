package model

import (
	"time"

	"github.com/google/uuid"
)

// ============ ContentType DTO ============

// ContentTypeCreate 创建内容类型请求
type ContentTypeCreate struct {
	Name                 string          `json:"name" binding:"required,min=1,max=200"`
	Slug                 string          `json:"slug" binding:"required,min=1,max=100"`
	Description          string          `json:"description"`
	Kind                 ContentTypeKind `json:"kind" binding:"required,oneof=collection single"`
	DisplayConfig        JSONB           `json:"display_config"`
	APISConfig           JSONB           `json:"api_config"`
	PreviewConfig        JSONB           `json:"preview_config"`
	VersioningEnabled    bool            `json:"versioning_enabled"`
	DraftAutosaveInterval *int           `json:"draft_autosave_interval"`
	SortOrder            int             `json:"sort_order"`
}

// ContentTypeUpdate 更新内容类型请求
type ContentTypeUpdate struct {
	Name                 *string          `json:"name" binding:"omitempty,min=1,max=200"`
	Slug                 *string          `json:"slug" binding:"omitempty,min=1,max=100"`
	Description          *string          `json:"description"`
	Kind                 *ContentTypeKind `json:"kind" binding:"omitempty,oneof=collection single"`
	DisplayConfig        *JSONB           `json:"display_config"`
	APISConfig           *JSONB           `json:"api_config"`
	PreviewConfig        *JSONB           `json:"preview_config"`
	VersioningEnabled    *bool            `json:"versioning_enabled"`
	DraftAutosaveInterval *int            `json:"draft_autosave_interval"`
	IsActive             *bool            `json:"is_active"`
	SortOrder            *int             `json:"sort_order"`
}

// ContentTypeResponse 内容类型响应
type ContentTypeResponse struct {
	ID                   uuid.UUID             `json:"id"`
	SiteID               uuid.UUID             `json:"site_id"`
	Name                 string                `json:"name"`
	Slug                 string                `json:"slug"`
	Description          string                `json:"description"`
	Kind                 ContentTypeKind       `json:"kind"`
	DisplayConfig        map[string]interface{} `json:"display_config"`
	APISConfig           map[string]interface{} `json:"api_config"`
	PreviewConfig        map[string]interface{} `json:"preview_config"`
	VersioningEnabled    bool                  `json:"versioning_enabled"`
	DraftAutosaveInterval *int                 `json:"draft_autosave_interval"`
	IsActive             bool                  `json:"is_active"`
	SortOrder            int                   `json:"sort_order"`
	CreatedBy            *uuid.UUID            `json:"created_by"`
	CreatedTime            time.Time             `json:"created_time"`
	UpdatedTime            time.Time             `json:"updated_time"`
	Fields               []FieldResponse       `json:"fields,omitempty"`
}

// ContentTypeListResponse 内容类型列表响应
type ContentTypeListResponse struct {
	Items      []ContentTypeResponse `json:"items"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
}

// ToResponse 转换为响应
func (ct *ContentType) ToResponse() ContentTypeResponse {
	resp := ContentTypeResponse{
		ID:                   ct.ID,
		SiteID:               ct.SiteID,
		Name:                 ct.Name,
		Slug:                 ct.Slug,
		Description:          ct.Description,
		Kind:                 ct.Kind,
		VersioningEnabled:    ct.VersioningEnabled,
		DraftAutosaveInterval: ct.DraftAutosaveInterval,
		IsActive:             ct.IsActive,
		SortOrder:            ct.SortOrder,
		CreatedBy:            ct.CreatedBy,
		CreatedTime:            ct.CreatedTime,
		UpdatedTime:            ct.UpdatedTime,
	}

	// 解析 JSONB 字段
	if ct.DisplayConfig != nil {
		resp.DisplayConfig = ct.DisplayConfig.Map()
	}
	if ct.APISConfig != nil {
		resp.APISConfig = ct.APISConfig.Map()
	}
	if ct.PreviewConfig != nil {
		resp.PreviewConfig = ct.PreviewConfig.Map()
	}

	// 转换字段
	if len(ct.Fields) > 0 {
		resp.Fields = make([]FieldResponse, len(ct.Fields))
		for i, f := range ct.Fields {
			resp.Fields[i] = f.ToResponse()
		}
	}

	return resp
}

// ============ Field DTO ============

// FieldCreate 创建字段请求
type FieldCreate struct {
	Name               string   `json:"name" binding:"required,min=1,max=100"`
	Label              string   `json:"label" binding:"required,min=1,max=200"`
	Description        string   `json:"description"`
	FieldType          string   `json:"field_type" binding:"required,oneof=text rich_text number boolean date datetime email url json media relation enum password"`
	Config             JSONB    `json:"config"`
	Validation         JSONB    `json:"validation"`
	Display            JSONB    `json:"display"`
	DefaultValue       *JSONB   `json:"default_value"`
	SortOrder          int      `json:"sort_order"`
	ConditionalDisplay *JSONB   `json:"conditional_display"`
}

// FieldUpdate 更新字段请求
type FieldUpdate struct {
	Name               *string `json:"name" binding:"omitempty,min=1,max=100"`
	Label              *string `json:"label" binding:"omitempty,min=1,max=200"`
	Description        *string `json:"description"`
	FieldType          *string `json:"field_type" binding:"omitempty,oneof=text rich_text number boolean date datetime email url json media relation enum password"`
	Config             *JSONB  `json:"config"`
	Validation         *JSONB  `json:"validation"`
	Display            *JSONB  `json:"display"`
	DefaultValue       *JSONB  `json:"default_value"`
	SortOrder          *int    `json:"sort_order"`
	ConditionalDisplay *JSONB  `json:"conditional_display"`
}

// FieldResponse 字段响应
type FieldResponse struct {
	ID                 uuid.UUID              `json:"id"`
	ContentTypeID      uuid.UUID              `json:"content_type_id"`
	Name               string                 `json:"name"`
	Label              string                 `json:"label"`
	Description        string                 `json:"description"`
	FieldType          string                 `json:"field_type"`
	Config             map[string]interface{} `json:"config"`
	Validation         map[string]interface{} `json:"validation"`
	Display            map[string]interface{} `json:"display"`
	DefaultValue       interface{}            `json:"default_value,omitempty"`
	SortOrder          int                    `json:"sort_order"`
	ConditionalDisplay interface{}            `json:"conditional_display,omitempty"`
	CreatedTime          time.Time              `json:"created_time"`
	UpdatedTime          time.Time              `json:"updated_time"`
}

// ToResponse 转换为响应
func (f *Field) ToResponse() FieldResponse {
	resp := FieldResponse{
		ID:            f.ID,
		ContentTypeID: f.ContentTypeID,
		Name:          f.Name,
		Label:         f.Label,
		Description:   f.Description,
		FieldType:     f.FieldType,
		SortOrder:     f.SortOrder,
		CreatedTime:     f.CreatedTime,
		UpdatedTime:     f.UpdatedTime,
	}

	if f.Config != nil {
		resp.Config = f.Config.Map()
	}
	if f.Validation != nil {
		resp.Validation = f.Validation.Map()
	}
	if f.Display != nil {
		resp.Display = f.Display.Map()
	}
	if f.DefaultValue != nil {
		resp.DefaultValue = f.DefaultValue.Interface()
	}
	if f.ConditionalDisplay != nil {
		resp.ConditionalDisplay = f.ConditionalDisplay.Interface()
	}

	return resp
}
