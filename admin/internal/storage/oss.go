// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss"
	"github.com/aliyun/alibabacloud-oss-go-sdk-v2/oss/credentials"
)

// OSSProvider 阿里云 OSS 存储驱动
type OSSProvider struct {
	client     *oss.Client
	bucket     string
	pathPrefix string
	baseURL    string
	region     string
}

// NewOSSProvider 创建阿里云 OSS 存储驱动
func NewOSSProvider(_ context.Context, cfg *ProviderConfig) (StorageProvider, error) {
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("阿里云 OSS 缺少 AccessKey/SecretKey")
	}

	// LoadDefaultConfig() 无参，返回 *oss.Config（不是 ConfigBuilder）
	ossCfg := oss.LoadDefaultConfig()
	ossCfg.Region = oss.Ptr(cfg.Region)
	ossCfg.CredentialsProvider = credentials.NewStaticCredentialsProvider(cfg.AccessKey, cfg.SecretKey, "")

	client := oss.NewClient(ossCfg)

	return &OSSProvider{
		client:     client,
		bucket:     cfg.Bucket,
		pathPrefix: cfg.PathPrefix,
		baseURL:    cfg.BaseURL,
		region:     cfg.Region,
	}, nil
}

func (p *OSSProvider) Name() string { return "oss" }

func (p *OSSProvider) fullKey(key string) string {
	if p.pathPrefix != "" {
		return p.pathPrefix + "/" + key
	}
	return key
}

func (p *OSSProvider) Upload(ctx context.Context, key string, body io.Reader, size int64, opts *WriteOptions) (*ObjectInfo, error) {
	req := &oss.PutObjectRequest{
		Bucket: oss.Ptr(p.bucket),
		Key:    oss.Ptr(p.fullKey(key)),
		Body:   body,
	}
	if opts != nil && opts.ContentType != "" {
		req.ContentType = oss.Ptr(opts.ContentType)
	}
	if size > 0 {
		req.ContentLength = oss.Ptr(size)
	}
	_, err := p.client.PutObject(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OSS 上传失败: %w", err)
	}
	return &ObjectInfo{
		Key:         key,
		Size:        size,
		ContentType: opts.ContentType,
		URL:         p.publicURL(key),
		CreatedAt:   time.Now(),
	}, nil
}

func (p *OSSProvider) Download(ctx context.Context, key string, _ *ReadOptions) (io.ReadCloser, error) {
	result, err := p.client.GetObject(ctx, &oss.GetObjectRequest{
		Bucket: oss.Ptr(p.bucket),
		Key:    oss.Ptr(p.fullKey(key)),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (p *OSSProvider) Delete(ctx context.Context, key string) error {
	_, err := p.client.DeleteObject(ctx, &oss.DeleteObjectRequest{
		Bucket: oss.Ptr(p.bucket),
		Key:    oss.Ptr(p.fullKey(key)),
	})
	return err
}

func (p *OSSProvider) Exists(ctx context.Context, key string) (bool, error) {
	_, err := p.client.HeadObject(ctx, &oss.HeadObjectRequest{
		Bucket: oss.Ptr(p.bucket),
		Key:    oss.Ptr(p.fullKey(key)),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (p *OSSProvider) URL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	if expiresIn > 0 {
		presigned, err := p.client.Presign(ctx, &oss.GetObjectRequest{
			Bucket: oss.Ptr(p.bucket),
			Key:    oss.Ptr(p.fullKey(key)),
		}, oss.PresignExpires(expiresIn))
		if err != nil {
			return "", err
		}
		return presigned.URL, nil
	}
	return p.publicURL(key), nil
}

func (p *OSSProvider) publicURL(key string) string {
	if p.baseURL != "" {
		return p.baseURL + "/" + key
	}
	return fmt.Sprintf("https://%s.oss-%s.aliyuncs.com/%s", p.bucket, p.region, p.fullKey(key))
}

func (p *OSSProvider) Stat(ctx context.Context, key string) (*ObjectInfo, error) {
	result, err := p.client.HeadObject(ctx, &oss.HeadObjectRequest{
		Bucket: oss.Ptr(p.bucket),
		Key:    oss.Ptr(p.fullKey(key)),
	})
	if err != nil {
		return nil, err
	}
	ct := ""
	if result.ContentType != nil {
		ct = *result.ContentType
	}
	etag := ""
	if result.ETag != nil {
		etag = *result.ETag
	}
	return &ObjectInfo{
		Key:         key,
		Size:        result.ContentLength, // int64，直接值
		ContentType: ct,
		ETag:        etag,
	}, nil
}

func (p *OSSProvider) List(ctx context.Context, prefix string, limit int, marker string) ([]ObjectInfo, string, error) {
	req := &oss.ListObjectsRequest{
		Bucket:   oss.Ptr(p.bucket),
		MaxKeys:   int32(limit),
		Prefix:    oss.Ptr(p.fullKey(prefix)),
		Delimiter: nil,
	}
	if marker != "" {
		req.Marker = oss.Ptr(marker)
	}
	result, err := p.client.ListObjects(ctx, req)
	if err != nil {
		return nil, "", err
	}
	items := make([]ObjectInfo, 0, len(result.Contents))
	for _, obj := range result.Contents {
		if obj.Key == nil {
			continue
		}
		// 去掉 pathPrefix 前缀得到相对 key
		relKey := *obj.Key
		if p.pathPrefix != "" {
			relKey = relKey[len(p.pathPrefix)+1:]
		}
		items = append(items, ObjectInfo{
			Key:    relKey,
			Size:   obj.Size,
			ETag:   derefString(obj.ETag),
			URL:    p.publicURL(relKey),
			CreatedAt: func() time.Time {
				if obj.LastModified != nil {
					return *obj.LastModified
				}
				return time.Time{}
			}(),
		})
	}
	next := ""
	if result.IsTruncated && result.NextMarker != nil {
		next = *result.NextMarker
	}
	return items, next, nil
}

func (p *OSSProvider) DeleteMulti(ctx context.Context, keys []string) error {
	var objects []oss.ObjectIdentifier
	for _, key := range keys {
		objects = append(objects, oss.ObjectIdentifier{Key: oss.Ptr(p.fullKey(key))})
	}
	_, err := p.client.DeleteMultipleObjects(ctx, &oss.DeleteMultipleObjectsRequest{
		Bucket: oss.Ptr(p.bucket),
		Delete: &oss.Delete{Objects: objects, Quiet: true},
	})
	return err
}

func derefString(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func init() {
	RegisterProvider("oss", NewOSSProvider)
}
