package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

// S3Provider S3 兼容存储驱动（支持 AWS S3 / MinIO / Longhorn / SeaweedFS 等）
type S3Provider struct {
	client     *s3.Client
	bucket     string
	pathPrefix string
	baseURL    string
}

// NewS3Provider 创建 S3 存储驱动
func NewS3Provider(_ context.Context, cfg *ProviderConfig) (StorageProvider, error) {
	// 构造 AWS 配置
	var awsCfg aws.Config
	var err error

	if cfg.Endpoint != "" {
		// 自定义端点（MinIO / 兼容 S3 的存储）
		awsCfg, err = awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(cfg.Region),
			awsconfig.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
				cfg.AccessKey, cfg.SecretKey, "",
			)),
		)
		if err != nil {
			return nil, fmt.Errorf("加载 AWS 配置失败: %w", err)
		}
	} else {
		// AWS S3（使用默认凭证链：环境变量/EC2 Role/Shared Config）
		awsCfg, err = awsconfig.LoadDefaultConfig(context.Background(),
			awsconfig.WithRegion(cfg.Region),
		)
		if err != nil {
			return nil, fmt.Errorf("加载 AWS 默认配置失败: %w", err)
		}
	}

	// 配置自定义端点（用于 MinIO 等）
	var s3Opts []func(*s3.Options)
	if cfg.Endpoint != "" {
		s3Opts = append(s3Opts, func(o *s3.Options) {
			o.BaseEndpoint = aws.String(cfg.Endpoint)
			o.UsePathStyle = cfg.ForcePathStyle
			o.Region = cfg.Region
		})
	}

	client := s3.NewFromConfig(awsCfg, s3Opts...)

	return &S3Provider{
		client:     client,
		bucket:     cfg.Bucket,
		pathPrefix: cfg.PathPrefix,
		baseURL:    cfg.BaseURL,
	}, nil
}

func (p *S3Provider) Name() string { return "s3" }

func (p *S3Provider) fullKey(key string) string {
	if p.pathPrefix != "" {
		return p.pathPrefix + "/" + key
	}
	return key
}

func (p *S3Provider) Upload(ctx context.Context, key string, body io.Reader, size int64, opts *WriteOptions) (*ObjectInfo, error) {
	input := &s3.PutObjectInput{
		Bucket:      aws.String(p.bucket),
		Key:         aws.String(p.fullKey(key)),
		Body:        body,
		ContentType: aws.String(opts.ContentType),
	}
	if opts.CacheControl != "" {
		input.CacheControl = aws.String(opts.CacheControl)
	}

	_, err := p.client.PutObject(ctx, input)
	if err != nil {
		return nil, fmt.Errorf("S3 上传失败: %w", err)
	}

	return &ObjectInfo{
		Key:         key,
		Size:        size,
		ContentType: opts.ContentType,
		URL:         p.publicURL(key),
		CreatedAt:   time.Now(),
	}, nil
}

func (p *S3Provider) Download(ctx context.Context, key string, _ *ReadOptions) (io.ReadCloser, error) {
	result, err := p.client.GetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(p.fullKey(key)),
	})
	if err != nil {
		return nil, err
	}
	return result.Body, nil
}

func (p *S3Provider) Delete(ctx context.Context, key string) error {
	_, err := p.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(p.fullKey(key)),
	})
	return err
}

func (p *S3Provider) DeleteMulti(ctx context.Context, keys []string) error {
	objects := make([]types.ObjectIdentifier, len(keys))
	for i, key := range keys {
		objects[i] = types.ObjectIdentifier{Key: aws.String(p.fullKey(key))}
	}
	_, err := p.client.DeleteObjects(ctx, &s3.DeleteObjectsInput{
		Bucket: aws.String(p.bucket),
		Delete: &types.Delete{Objects: objects},
	})
	return err
}

func (p *S3Provider) Exists(ctx context.Context, key string) (bool, error) {
	_, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(p.fullKey(key)),
	})
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (p *S3Provider) URL(ctx context.Context, key string, expiresIn time.Duration) (string, error) {
	if expiresIn > 0 {
		// 生成签名 URL
		presignClient := s3.NewPresignClient(p.client)
		presigned, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
			Bucket: aws.String(p.bucket),
			Key:    aws.String(p.fullKey(key)),
		}, s3.WithPresignExpires(expiresIn))
		if err != nil {
			return "", err
		}
		return presigned.URL, nil
	}
	return p.publicURL(key), nil
}

func (p *S3Provider) publicURL(key string) string {
	if p.baseURL != "" {
		return p.baseURL + "/" + key
	}
	return "https://" + p.bucket + ".s3.amazonaws.com/" + p.fullKey(key)
}

func (p *S3Provider) Stat(ctx context.Context, key string) (*ObjectInfo, error) {
	result, err := p.client.HeadObject(ctx, &s3.HeadObjectInput{
		Bucket: aws.String(p.bucket),
		Key:    aws.String(p.fullKey(key)),
	})
	if err != nil {
		return nil, err
	}
	return &ObjectInfo{
		Key:         key,
		Size:        *result.ContentLength,
		ContentType: *result.ContentType,
		ETag:        *result.ETag,
		CreatedAt:   *result.LastModified,
	}, nil
}

func (p *S3Provider) List(ctx context.Context, prefix string, pageSize int, _ string) ([]ObjectInfo, string, error) {
	input := &s3.ListObjectsV2Input{
		Bucket:  aws.String(p.bucket),
		Prefix:  aws.String(p.fullKey(prefix)),
		MaxKeys: aws.Int32(int32(pageSize + 1)),
	}

	result, err := p.client.ListObjectsV2(ctx, input)
	if err != nil {
		return nil, "", err
	}

	var infos []ObjectInfo
	hasMore := false
	for i, obj := range result.Contents {
		if i >= pageSize {
			hasMore = true
			break
		}
		infos = append(infos, ObjectInfo{
			Key:       *obj.Key,
			Size:      *obj.Size,
			CreatedAt: *obj.LastModified,
		})
	}

	var nextToken string
	if hasMore && len(result.Contents) > 0 {
		nextToken = *result.Contents[len(result.Contents)-1].Key
	}

	return infos, nextToken, nil
}

func init() {
	RegisterProvider("s3", NewS3Provider)
}
