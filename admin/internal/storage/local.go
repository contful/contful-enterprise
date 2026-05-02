// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"mime"
)

// LocalProvider 本地存储驱动
type LocalProvider struct {
	rootDir string
	baseURL string
}

// NewLocalProvider 创建本地存储驱动
func NewLocalProvider(_ context.Context, cfg *ProviderConfig) (StorageProvider, error) {
	if cfg.RootDir == "" {
		cfg.RootDir = "./uploads"
	}
	if cfg.BaseURL == "" {
		cfg.BaseURL = "/assets"
	}
	// 确保目录存在
	if err := os.MkdirAll(cfg.RootDir, 0755); err != nil {
		return nil, fmt.Errorf("创建存储目录失败: %w", err)
	}
	return &LocalProvider{
		rootDir: cfg.RootDir,
		baseURL: cfg.BaseURL,
	}, nil
}

func (p *LocalProvider) Name() string { return "local" }

func (p *LocalProvider) Upload(ctx context.Context, key string, body io.Reader, size int64, opts *WriteOptions) (*ObjectInfo, error) {
	fullPath := filepath.Join(p.rootDir, key)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	f, err := os.OpenFile(fullPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return nil, fmt.Errorf("创建文件失败: %w", err)
	}
	defer f.Close()

	written, err := io.Copy(f, body)
	if err != nil {
		return nil, fmt.Errorf("写入文件失败: %w", err)
	}

	return &ObjectInfo{
		Key:         key,
		Size:        written,
		ContentType: opts.ContentType,
		URL:         fmt.Sprintf("%s/%s", p.baseURL, key),
		CreatedAt:   time.Now(),
	}, nil
}

func (p *LocalProvider) Download(ctx context.Context, key string, _ *ReadOptions) (io.ReadCloser, error) {
	fullPath := filepath.Join(p.rootDir, key)
	return os.Open(fullPath)
}

func (p *LocalProvider) Delete(ctx context.Context, key string) error {
	fullPath := filepath.Join(p.rootDir, key)
	return os.Remove(fullPath)
}

func (p *LocalProvider) DeleteMulti(ctx context.Context, keys []string) error {
	for _, key := range keys {
		if err := p.Delete(ctx, key); err != nil && !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func (p *LocalProvider) Exists(ctx context.Context, key string) (bool, error) {
	fullPath := filepath.Join(p.rootDir, key)
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (p *LocalProvider) URL(_ context.Context, key string, _ time.Duration) (string, error) {
	return fmt.Sprintf("%s/%s", p.baseURL, key), nil
}

func (p *LocalProvider) Stat(ctx context.Context, key string) (*ObjectInfo, error) {
	fullPath := filepath.Join(p.rootDir, key)
	info, err := os.Stat(fullPath)
	if err != nil {
		return nil, err
	}
	ct := http.DetectContentType(nil) // 本地存储不维护 MIME，需外部传入
	_ = ct
	return &ObjectInfo{
		Key:       key,
		Size:      info.Size(),
		CreatedAt: info.ModTime(),
	}, nil
}

// ServeFile 提供静态文件服务（用于 HTTP 响应）
func (p *LocalProvider) ServeFile(ctx context.Context, key string) (io.ReadCloser, string, error) {
	fullPath := filepath.Join(p.rootDir, key)
	file, err := os.Open(fullPath)
	if err != nil {
		return nil, "", err
	}

	// 检测 MIME 类型
	ext := filepath.Ext(key)
	mimeType := mime.TypeByExtension(ext)
	if mimeType == "" {
		mimeType = "application/octet-stream"
	}

	return file, mimeType, nil
}

func (p *LocalProvider) List(ctx context.Context, prefix string, pageSize int, _ string) ([]ObjectInfo, string, error) {
	dir := filepath.Join(p.rootDir, prefix)
	f, err := os.Open(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, "", nil
		}
		return nil, "", err
	}
	defer f.Close()

	entries, err := f.ReadDir(pageSize + 1)
	if err != nil {
		return nil, "", err
	}

	var result []ObjectInfo
	hasMore := len(entries) > pageSize
	if hasMore {
		entries = entries[:pageSize]
	}

	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		info, _ := e.Info()
		result = append(result, ObjectInfo{
			Key:       filepath.Join(prefix, e.Name()),
			Size:      info.Size(),
			CreatedAt: info.ModTime(),
		})
	}

	var nextToken string
	if hasMore && len(entries) > 0 {
		nextToken = entries[len(entries)-1].Name()
	}
	return result, nextToken, nil
}

func init() {
	RegisterProvider("local", NewLocalProvider)
}
