// Package storage 提供存储驱动抽象，支持 local / OSS / S3 / COS / OBS 多后端。
package storage

import (
	"context"
	"io"
	"time"
)

// ObjectInfo 对象元信息
type ObjectInfo struct {
	Key         string            `json:"key"`
	Size        int64             `json:"size"`
	ContentType string            `json:"content_type"`
	ETag        string            `json:"etag,omitempty"`
	URL         string            `json:"url,omitempty"`
	CreatedAt   time.Time         `json:"created_at"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// WriteOptions 上传选项
type WriteOptions struct {
	ContentType  string
	CacheControl string // e.g. "public, max-age=31536000"
	Metadata    map[string]string
}

// ReadOptions 读取选项
type ReadOptions struct {
	Range *Range
}

// Range HTTP Range 参数
type Range struct {
	Offset int64
	Length int64
}

// StorageProvider 存储驱动接口
type StorageProvider interface {
	// Name 返回驱动名称（用于 disk 字段）
	Name() string

	// Upload 上传文件，返回对象元信息
	Upload(ctx context.Context, key string, body io.Reader, size int64, opts *WriteOptions) (*ObjectInfo, error)

	// Download 下载文件内容
	Download(ctx context.Context, key string, opts *ReadOptions) (io.ReadCloser, error)

	// Delete 删除文件
	Delete(ctx context.Context, key string) error

	// DeleteMulti 批量删除（可选实现）
	DeleteMulti(ctx context.Context, keys []string) error

	// Exists 检查文件是否存在
	Exists(ctx context.Context, key string) (bool, error)

	// URL 生成访问 URL
	// local: 直接返回公开 URL
	// 云存储: 支持签名 URL（expiresIn > 0 时生成临时签名 URL）
	URL(ctx context.Context, key string, expiresIn time.Duration) (string, error)

	// Stat 获取文件元信息
	Stat(ctx context.Context, key string) (*ObjectInfo, error)

	// List 列出前缀下的文件（分页）
	List(ctx context.Context, prefix string, pageSize int, continuationToken string) ([]ObjectInfo, string, error)
}
