// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/huaweicloud/huaweicloud-sdk-go-obs/obs"
)

// OBSProvider 华为云 OBS 存储驱动
type OBSProvider struct {
	client     *obs.ObsClient
	bucket     string
	pathPrefix string
	baseURL    string
	endpoint   string
}

// NewOBSProvider 创建华为云 OBS 存储驱动
// endpoint 示例: obs.cn-north-4.myhuaweicloud.com
func NewOBSProvider(_ context.Context, cfg *ProviderConfig) (StorageProvider, error) {
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("华为云 OBS 缺少 AccessKey/SecretKey")
	}

	endpoint := cfg.Endpoint
	if endpoint == "" {
		endpoint = fmt.Sprintf("obs.%s.myhuaweicloud.com", cfg.Region)
	}

	client, err := obs.New(cfg.AccessKey, cfg.SecretKey, endpoint)
	if err != nil {
		return nil, fmt.Errorf("创建 OBS 客户端失败: %w", err)
	}

	return &OBSProvider{
		client:     client,
		bucket:     cfg.Bucket,
		pathPrefix: cfg.PathPrefix,
		baseURL:    cfg.BaseURL,
		endpoint:   endpoint,
	}, nil
}

func (p *OBSProvider) Name() string { return "obs" }

func (p *OBSProvider) fullKey(key string) string {
	if p.pathPrefix != "" {
		return p.pathPrefix + "/" + key
	}
	return key
}

func (p *OBSProvider) Upload(ctx context.Context, key string, body io.Reader, size int64, opts *WriteOptions) (*ObjectInfo, error) {
	input := &obs.PutObjectInput{
		PutObjectBasicInput: obs.PutObjectBasicInput{
			ObjectOperationInput: obs.ObjectOperationInput{
				Bucket: p.bucket,
				Key:    p.fullKey(key),
			},
		},
		Body: body,
	}
	output, err := p.client.PutObject(input)
	if err != nil {
		return nil, fmt.Errorf("OBS 上传失败: %w", err)
	}
	return &ObjectInfo{
		Key:         key,
		Size:        size,
		ContentType: opts.ContentType,
		ETag:        output.ETag,
		URL:         p.publicURL(key),
		CreatedAt:   time.Now(),
	}, nil
}

func (p *OBSProvider) Download(ctx context.Context, key string, _ *ReadOptions) (io.ReadCloser, error) {
	output, err := p.client.GetObject(&obs.GetObjectInput{
		GetObjectMetadataInput: obs.GetObjectMetadataInput{
			Bucket: p.bucket,
			Key:    p.fullKey(key),
		},
	})
	if err != nil {
		return nil, err
	}
	return output.Body, nil
}

func (p *OBSProvider) Delete(ctx context.Context, key string) error {
	_, err := p.client.DeleteObject(&obs.DeleteObjectInput{
		Bucket: p.bucket,
		Key:    p.fullKey(key),
	})
	return err
}

func (p *OBSProvider) Exists(ctx context.Context, key string) (bool, error) {
	_, err := p.client.GetObjectMetadata(&obs.GetObjectMetadataInput{
		Bucket: p.bucket,
		Key:    p.fullKey(key),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (p *OBSProvider) URL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	if expiresIn > 0 {
		req := &obs.CreateSignedUrlInput{
			Method:  obs.HttpMethodGet,
			Bucket:  p.bucket,
			Key:    p.fullKey(key),
			Expires: int(expiresIn / time.Second),
		}
		output, err := p.client.CreateSignedUrl(req)
		if err != nil {
			return "", err
		}
		return output.SignedUrl, nil
	}
	return p.publicURL(key), nil
}

func (p *OBSProvider) publicURL(key string) string {
	if p.baseURL != "" {
		return p.baseURL + "/" + key
	}
	return fmt.Sprintf("https://%s.%s/%s", p.bucket, p.endpoint, p.fullKey(key))
}

func (p *OBSProvider) Stat(ctx context.Context, key string) (*ObjectInfo, error) {
	output, err := p.client.GetObjectMetadata(&obs.GetObjectMetadataInput{
		Bucket: p.bucket,
		Key:    p.fullKey(key),
	})
	if err != nil {
		return nil, err
	}
	return &ObjectInfo{
		Key:         key,
		Size:        output.ContentLength,
		ContentType: output.ContentType,
		ETag:        output.ETag,
	}, nil
}

func (p *OBSProvider) List(ctx context.Context, prefix string, pageSize int, _ string) ([]ObjectInfo, string, error) {
	input := &obs.ListObjectsInput{
		Bucket: p.bucket,
		ListObjsInput: obs.ListObjsInput{
			Prefix:  p.fullKey(prefix),
			MaxKeys: pageSize + 1,
		},
	}
	output, err := p.client.ListObjects(input)
	if err != nil {
		return nil, "", err
	}

	var infos []ObjectInfo
	hasMore := len(output.Contents) > pageSize
	for i, obj := range output.Contents {
		if i >= pageSize {
			break
		}
		infos = append(infos, ObjectInfo{
			Key:       obj.Key,
			Size:      obj.Size,
			CreatedAt: obj.LastModified,
		})
	}

	var nextToken string
	if hasMore && len(output.Contents) > 0 {
		nextToken = output.Contents[len(output.Contents)-1].Key
	}
	return infos, nextToken, nil
}

func (p *OBSProvider) DeleteMulti(ctx context.Context, keys []string) error {
	var objects []obs.ObjectToDelete
	for _, key := range keys {
		objects = append(objects, obs.ObjectToDelete{Key: p.fullKey(key)})
	}
	_, err := p.client.DeleteObjects(&obs.DeleteObjectsInput{
		Bucket:  p.bucket,
		Objects: objects,
		Quiet:   true,
	})
	return err
}

func init() {
	RegisterProvider("obs", NewOBSProvider)
}
