package model

import (
	"time"

	"github.com/google/uuid"
)

// AssetType 资源类型
type AssetType string

const (
	AssetTypeImage  AssetType = "image"  // 图片
	AssetTypeVideo  AssetType = "video"  // 视频
	AssetTypeAudio  AssetType = "audio"  // 音频
	AssetTypeDocument AssetType = "document" // 文档
	AssetTypeFile   AssetType = "file"   // 其他文件
)

// AssetVisibility 可见性
type AssetVisibility string

const (
	AssetVisibilityPublic AssetVisibility = "public"  // 公开
	AssetVisibilityPrivate AssetVisibility = "private" // 私有
)

// Asset 媒体资源
type Asset struct {
	ID            uuid.UUID       `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SiteID        uuid.UUID       `json:"site_id" gorm:"type:uuid;not null;index"`
	FolderID      *uuid.UUID      `json:"folder_id" gorm:"type:uuid;index"`
	UUID          string          `json:"uuid" gorm:"size:36;uniqueIndex"` // 业务 UUID
	Name          string          `json:"name" gorm:"size:255;not null"`
	OriginalName  string          `json:"original_name" gorm:"size:255;not null"`
	Slug          string          `json:"slug" gorm:"size:255;index"`
	Type          AssetType       `json:"type" gorm:"type:asset_type;not null;index"`
	MimeType      string          `json:"mime_type" gorm:"size:100;not null"`
	Extension     string          `json:"extension" gorm:"size:20;not null;index"`
	Size          int64           `json:"size" gorm:"not null"` // bytes
	Width         *int            `json:"width" gorm:"type:int"`  // 图片/视频宽度
	Height        *int            `json:"height" gorm:"type:int"` // 图片/视频高度
	Duration      *float64        `json:"duration" gorm:"type:float"` // 音视频时长(秒)
	Path          string          `json:"path" gorm:"size:500;not null"` // 存储路径
	URL           string          `json:"url" gorm:"size:500;not null"` // 访问 URL
	ThumbnailURL  *string         `json:"thumbnail_url" gorm:"size:500"` // 缩略图 URL
	Alt           string          `json:"alt" gorm:"type:text"` // 替代文本
	Title         string          `json:"title" gorm:"size:255"` // 标题
	Caption       string          `json:"caption" gorm:"type:text"` // 说明
	AltText        string          `json:"alt_text" gorm:"type:text"` // 辅助说明
	Description   string          `json:"description" gorm:"type:text"`
	Tags          []string        `json:"tags" gorm:"type:text[]"`
	Metadata      JSONB           `json:"metadata" gorm:"type:jsonb;default:'{}'"`
	Visibility    AssetVisibility `json:"visibility" gorm:"type:asset_visibility;not null;default:'private'"`
	FileHash      string          `json:"file_hash" gorm:"size:64;index"` // SHA-256
	Disk          string          `json:"disk" gorm:"size:50;not null;default:'local'"`
	DownloadCount int             `json:"download_count" gorm:"not null;default:0"`
	UsedCount     int             `json:"used_count" gorm:"not null;default:0"` // 被引用次数
	CreatedBy     *uuid.UUID      `json:"created_by" gorm:"type:uuid"`
	CreatedAt     time.Time       `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt     time.Time       `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt     *time.Time      `json:"deleted_at" gorm:"index"`

	// 关联
	Folder *AssetFolder `json:"folder,omitempty" gorm:"foreignKey:FolderID;references:ID"`
}

// TableName 表名
func (Asset) TableName() string {
	return "assets"
}

// AssetFolder 资源文件夹
type AssetFolder struct {
	ID        uuid.UUID  `json:"id" gorm:"type:uuid;primaryKey;default:gen_random_uuid()"`
	SiteID    uuid.UUID  `json:"site_id" gorm:"type:uuid;not null;index"`
	ParentID  *uuid.UUID `json:"parent_id" gorm:"type:uuid;index"`
	Name      string     `json:"name" gorm:"size:255;not null"`
	Slug      string     `json:"slug" gorm:"size:255;not null"`
	Path      string     `json:"path" gorm:"size:500;not null"` // 完整路径
	SortOrder int        `json:"sort_order" gorm:"not null;default:0"`
	CreatedBy *uuid.UUID `json:"created_by" gorm:"type:uuid"`
	CreatedAt time.Time  `json:"created_at" gorm:"autoCreateTime"`
	UpdatedAt time.Time  `json:"updated_at" gorm:"autoUpdateTime"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`

	// 关联
	Parent   *AssetFolder `json:"parent,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Children []AssetFolder `json:"children,omitempty" gorm:"foreignKey:ParentID;references:ID"`
	Assets   []Asset       `json:"assets,omitempty" gorm:"foreignKey:FolderID;references:ID"`
}

// TableName 表名
func (AssetFolder) TableName() string {
	return "asset_folders"
}
