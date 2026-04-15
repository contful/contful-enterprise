package model

import (
	"time"

	"github.com/google/uuid"
)

// ============ Asset DTO ============

// AssetCreate 创建资源请求
type AssetCreate struct {
	FolderID   *uuid.UUID      `json:"folder_id"`
	Name       string          `json:"name" binding:"required"`
	Alt        string          `json:"alt"`
	Title      string          `json:"title"`
	Caption    string          `json:"caption"`
	AltText    string          `json:"alt_text"`
	Description string        `json:"description"`
	Tags       []string        `json:"tags"`
	Visibility *AssetVisibility `json:"visibility"`
}

// AssetUpdate 更新资源请求
type AssetUpdate struct {
	FolderID    *uuid.UUID      `json:"folder_id"`
	Name        *string         `json:"name"`
	Alt         *string         `json:"alt"`
	Title       *string         `json:"title"`
	Caption     *string         `json:"caption"`
	AltText     *string         `json:"alt_text"`
	Description *string         `json:"description"`
	Tags        []string        `json:"tags"`
	Visibility  *AssetVisibility `json:"visibility"`
}

// AssetUpload 上传响应
type AssetUpload struct {
	ID           uuid.UUID `json:"id"`
	UUID         string    `json:"uuid"`
	Name         string    `json:"name"`
	OriginalName string    `json:"original_name"`
	Type         AssetType `json:"type"`
	MimeType     string    `json:"mime_type"`
	Extension    string    `json:"extension"`
	Size         int64     `json:"size"`
	Width        *int      `json:"width,omitempty"`
	Height       *int      `json:"height,omitempty"`
	Duration     *float64  `json:"duration,omitempty"`
	Path         string    `json:"path"`
	URL          string    `json:"url"`
	ThumbnailURL *string   `json:"thumbnail_url,omitempty"`
	FileHash     string    `json:"file_hash"`
	CreatedAt    time.Time `json:"created_at"`
}

// AssetResponse 资源响应
type AssetResponse struct {
	ID           uuid.UUID       `json:"id"`
	SiteID       uuid.UUID       `json:"site_id"`
	FolderID     *uuid.UUID      `json:"folder_id,omitempty"`
	UUID         string          `json:"uuid"`
	Name         string          `json:"name"`
	OriginalName string          `json:"original_name"`
	Slug         string          `json:"slug"`
	Type         AssetType       `json:"type"`
	MimeType     string          `json:"mime_type"`
	Extension    string          `json:"extension"`
	Size         int64           `json:"size"`
	Width        *int            `json:"width,omitempty"`
	Height       *int            `json:"height,omitempty"`
	Duration     *float64        `json:"duration,omitempty"`
	Path         string          `json:"path"`
	URL          string          `json:"url"`
	ThumbnailURL *string         `json:"thumbnail_url,omitempty"`
	Alt          string          `json:"alt,omitempty"`
	Title        string          `json:"title,omitempty"`
	Caption      string          `json:"caption,omitempty"`
	AltText      string          `json:"alt_text,omitempty"`
	Description  string          `json:"description,omitempty"`
	Tags         []string        `json:"tags,omitempty"`
	Metadata     JSONB           `json:"metadata,omitempty"`
	Visibility   AssetVisibility `json:"visibility"`
	FileHash     string          `json:"file_hash"`
	Disk         string          `json:"disk"`
	DownloadCount int            `json:"download_count"`
	UsedCount    int             `json:"used_count"`
	CreatedBy    *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// AssetListResponse 资源列表响应
type AssetListResponse struct {
	Items      []AssetResponse `json:"items"`
	Total      int64           `json:"total"`
	Page       int             `json:"page"`
	PageSize   int             `json:"page_size"`
}

// AssetListFilter 资源列表过滤条件
type AssetListFilter struct {
	FolderID  *uuid.UUID      `json:"folder_id"`
	Type      *AssetType      `json:"type"`
	Extension *string         `json:"extension"`
	Tag       *string         `json:"tag"`
	Keyword   *string         `json:"keyword"` // 搜索名称
}

// AssetUploadRequest 上传请求
type AssetUploadRequest struct {
	FolderID *uuid.UUID `form:"folder_id"`
	Alt      string     `form:"alt"`
	Title    string     `form:"title"`
}

// ============ Folder DTO ============

// FolderCreate 创建文件夹请求
type FolderCreate struct {
	ParentID  *uuid.UUID `json:"parent_id"`
	Name      string     `json:"name" binding:"required"`
	SortOrder int        `json:"sort_order"`
}

// FolderUpdate 更新文件夹请求
type FolderUpdate struct {
	ParentID  *uuid.UUID `json:"parent_id"`
	Name      *string    `json:"name"`
	SortOrder *int       `json:"sort_order"`
}

// FolderResponse 文件夹响应
type FolderResponse struct {
	ID        uuid.UUID       `json:"id"`
	SiteID    uuid.UUID       `json:"site_id"`
	ParentID  *uuid.UUID      `json:"parent_id,omitempty"`
	Name      string          `json:"name"`
	Slug      string          `json:"slug"`
	Path      string          `json:"path"`
	SortOrder int             `json:"sort_order"`
	Children  []FolderResponse `json:"children,omitempty"`
	Assets    []AssetResponse  `json:"assets,omitempty"`
	CreatedBy *uuid.UUID      `json:"created_by,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
	UpdatedAt time.Time       `json:"updated_at"`
}

// ToResponse 转换为响应
func (a *Asset) ToResponse() AssetResponse {
	return AssetResponse{
		ID:            a.ID,
		SiteID:        a.SiteID,
		FolderID:      a.FolderID,
		UUID:          a.UUID,
		Name:          a.Name,
		OriginalName:  a.OriginalName,
		Slug:          a.Slug,
		Type:          a.Type,
		MimeType:      a.MimeType,
		Extension:     a.Extension,
		Size:          a.Size,
		Width:         a.Width,
		Height:        a.Height,
		Duration:      a.Duration,
		Path:          a.Path,
		URL:           a.URL,
		ThumbnailURL:  a.ThumbnailURL,
		Alt:           a.Alt,
		Title:         a.Title,
		Caption:       a.Caption,
		AltText:       a.AltText,
		Description:   a.Description,
		Tags:          a.Tags,
		Metadata:      a.Metadata,
		Visibility:    a.Visibility,
		FileHash:      a.FileHash,
		Disk:          a.Disk,
		DownloadCount: a.DownloadCount,
		UsedCount:     a.UsedCount,
		CreatedBy:     a.CreatedBy,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}
}

// ToFolderResponse 转换为文件夹响应
func (f *AssetFolder) ToFolderResponse() FolderResponse {
	resp := FolderResponse{
		ID:        f.ID,
		SiteID:    f.SiteID,
		ParentID:  f.ParentID,
		Name:      f.Name,
		Slug:      f.Slug,
		Path:      f.Path,
		SortOrder: f.SortOrder,
		CreatedBy: f.CreatedBy,
		CreatedAt: f.CreatedAt,
		UpdatedAt: f.UpdatedAt,
	}

	if len(f.Children) > 0 {
		resp.Children = make([]FolderResponse, len(f.Children))
		for i, child := range f.Children {
			r := child.ToFolderResponse()
			resp.Children[i] = r
		}
	}

	if len(f.Assets) > 0 {
		resp.Assets = make([]AssetResponse, len(f.Assets))
		for i, asset := range f.Assets {
			resp.Assets[i] = asset.ToResponse()
		}
	}

	return resp
}
