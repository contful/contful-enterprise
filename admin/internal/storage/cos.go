package storage

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/tencentyun/cos-go-sdk-v5"
)

// COSProvider 腾讯云 COS 存储驱动
type COSProvider struct {
	client     *cos.Client
	bucket     string
	pathPrefix string
	baseURL    string
	region     string
}

// NewCOSProvider 创建腾讯云 COS 存储驱动
// bucket 格式: {name}-{appid}
// endpoint 示例: https://<bucket>.cos.<region>.myqcloud.com
func NewCOSProvider(_ context.Context, cfg *ProviderConfig) (StorageProvider, error) {
	if cfg.AccessKey == "" || cfg.SecretKey == "" {
		return nil, fmt.Errorf("腾讯云 COS 缺少 SecretID/SecretKey")
	}

	// 构造 bucket URL
	cosURL := cfg.Endpoint
	if cosURL == "" {
		cosURL = fmt.Sprintf("https://%s.cos.%s.myqcloud.com", cfg.Bucket, cfg.Region)
	}

	u, err := url.Parse(cosURL)
	if err != nil {
		return nil, fmt.Errorf("无效的 COS endpoint: %w", err)
	}

	client := cos.NewClient(&cos.BaseURL{BucketURL: u}, &http.Client{
		Transport: &cos.AuthorizationTransport{
			SecretID:  cfg.AccessKey,
			SecretKey: cfg.SecretKey,
		},
	})

	return &COSProvider{
		client:     client,
		bucket:     cfg.Bucket,
		pathPrefix: cfg.PathPrefix,
		baseURL:    cfg.BaseURL,
		region:     cfg.Region,
	}, nil
}

func (p *COSProvider) Name() string { return "cos" }

func (p *COSProvider) fullKey(key string) string {
	if p.pathPrefix != "" {
		return p.pathPrefix + "/" + key
	}
	return key
}

func (p *COSProvider) Upload(ctx context.Context, key string, body io.Reader, size int64, opts *WriteOptions) (*ObjectInfo, error) {
	_, err := p.client.Object.Put(ctx, p.fullKey(key), body, nil)
	if err != nil {
		return nil, fmt.Errorf("COS 上传失败: %w", err)
	}
	return &ObjectInfo{
		Key:         key,
		Size:        size,
		ContentType: opts.ContentType,
		URL:         p.publicURL(key),
		CreatedAt:   time.Now(),
	}, nil
}

func (p *COSProvider) Download(ctx context.Context, key string, _ *ReadOptions) (io.ReadCloser, error) {
	resp, err := p.client.Object.Get(ctx, p.fullKey(key), nil)
	if err != nil {
		return nil, err
	}
	return resp.Body, nil
}

func (p *COSProvider) Delete(ctx context.Context, key string) error {
	_, err := p.client.Object.Delete(ctx, p.fullKey(key), nil)
	return err
}

func (p *COSProvider) Exists(ctx context.Context, key string) (bool, error) {
	_, err := p.client.Object.Head(ctx, p.fullKey(key), nil)
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (p *COSProvider) URL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	if expiresIn > 0 {
		// 生成签名 URL，COS SDK 不提供直接签名方法，使用公共 URL
		return p.publicURL(key), nil
	}
	return p.publicURL(key), nil
}

func (p *COSProvider) publicURL(key string) string {
	if p.baseURL != "" {
		return p.baseURL + "/" + key
	}
	return fmt.Sprintf("https://%s.cos.%s.myqcloud.com/%s", p.bucket, p.region, p.fullKey(key))
}

func (p *COSProvider) Stat(ctx context.Context, key string) (*ObjectInfo, error) {
	resp, err := p.client.Object.Head(ctx, p.fullKey(key), nil)
	if err != nil {
		return nil, err
	}
	return &ObjectInfo{
		Key:         key,
		Size:        resp.ContentLength,
		ContentType: resp.Header.Get("Content-Type"),
	}, nil
}

func (p *COSProvider) List(_ context.Context, _ string, _ int, _ string) ([]ObjectInfo, string, error) {
	// TODO: 使用 GET Bucket (List Objects) 封装分页列举
	return nil, "", fmt.Errorf("List 方法暂未实现")
}

func (p *COSProvider) DeleteMulti(ctx context.Context, keys []string) error {
	var objects []cos.Object
	for _, k := range keys {
		objects = append(objects, cos.Object{Key: p.fullKey(k)})
	}
	_, _, err := p.client.Object.DeleteMulti(ctx, &cos.ObjectDeleteMultiOptions{
		Quiet:   true,
		Objects: objects,
	})
	return err
}

// Region 返回 COS region，从 baseURL 解析
func (p *COSProvider) Region() string {
	return p.region
}

func init() {
	RegisterProvider("cos", NewCOSProvider)
}
