package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"

	"github.com/google/uuid"
)

const (
	// TokenPrefix API Token 前缀
	TokenPrefix = "ctf_"
	// TokenLength Token 随机部分长度
	TokenLength = 32
)

// APITokenService API Token 服务
type APITokenService struct {
	tokenRepo *repository.APITokenRepository
}

// NewAPITokenService 新建服务
func NewAPITokenService(tokenRepo *repository.APITokenRepository) *APITokenService {
	return &APITokenService{tokenRepo: tokenRepo}
}

// GenerateToken 生成新的 Token
func (s *APITokenService) GenerateToken() (fullToken string, tokenHash string, prefix string, err error) {
	// 生成随机字节
	randomBytes := make([]byte, TokenLength)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", "", "", fmt.Errorf("生成随机数失败: %w", err)
	}

	// 生成完整 Token
	fullToken = TokenPrefix + hex.EncodeToString(randomBytes)

	// 计算 Hash
	hash := sha256.Sum256([]byte(fullToken))
	tokenHash = hex.EncodeToString(hash[:])

	// 生成前缀 (ctf_ + 前 6 位)
	prefix = fullToken[:10]

	return fullToken, tokenHash, prefix, nil
}

// Create 创建 API Token
func (s *APITokenService) Create(ctx context.Context, siteID, userID uuid.UUID, req *model.APITokenCreate) (*model.APIToken, string, error) {
	// 生成 Token
	fullToken, tokenHash, prefix, err := s.GenerateToken()
	if err != nil {
		return nil, "", err
	}

	// 设置默认权限
	permissions := req.Permissions
	if permissions == nil {
		defaultPerms := model.DefaultEndpointPermission()
		permissions = &defaultPerms
	}

	// 设置默认速率限制
	rateLimits := req.RateLimits
	if rateLimits == nil {
		defaultLimits := model.DefaultRateLimits()
		rateLimits = &defaultLimits
	}

	// 创建 Token 记录
	token := &model.APIToken{
		ID:          uuid.New(),
		SiteID:      siteID,
		Name:        req.Name,
		Description: req.Description,
		TokenPrefix: prefix,
		TokenHash:   tokenHash,
		Permissions: *permissions,
		RateLimits:  *rateLimits,
		Usage:       model.EmptyUsage(),
		ExpiresAt:   req.ExpiresAt,
		Status:      model.TokenStatusActive,
		CreatedBy:   &userID,
	}

	if err := s.tokenRepo.Create(ctx, token); err != nil {
		return nil, "", fmt.Errorf("创建 Token 失败: %w", err)
	}

	return token, fullToken, nil
}

// Get 获取 Token
func (s *APITokenService) Get(ctx context.Context, id uuid.UUID) (*model.APIToken, error) {
	return s.tokenRepo.GetByID(ctx, id)
}

// List 列出 Token
func (s *APITokenService) List(ctx context.Context, siteID uuid.UUID, filter *model.APITokenListFilter, page, pageSize int) (*model.APITokenListResponse, error) {
	tokens, total, err := s.tokenRepo.List(ctx, siteID, filter, page, pageSize)
	if err != nil {
		return nil, err
	}

	items := make([]model.APITokenResponse, len(tokens))
	for i, token := range tokens {
		items[i] = token.ToResponse()
	}

	return &model.APITokenListResponse{
		Items:    items,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	}, nil
}

// Update 更新 Token
func (s *APITokenService) Update(ctx context.Context, id uuid.UUID, req *model.APITokenUpdate) (*model.APIToken, error) {
	token, err := s.tokenRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		token.Name = *req.Name
	}
	if req.Description != nil {
		token.Description = *req.Description
	}
	if req.Permissions != nil {
		token.Permissions = *req.Permissions
	}
	if req.RateLimits != nil {
		token.RateLimits = *req.RateLimits
	}
	if req.ExpiresAt != nil {
		token.ExpiresAt = req.ExpiresAt
	}
	if req.Status != nil {
		token.Status = *req.Status
	}

	if err := s.tokenRepo.Update(ctx, token); err != nil {
		return nil, err
	}

	return token, nil
}

// Delete 删除 Token
func (s *APITokenService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.tokenRepo.Delete(ctx, id)
}

// Revoke 撤销 Token
func (s *APITokenService) Revoke(ctx context.Context, id uuid.UUID) error {
	return s.tokenRepo.Revoke(ctx, id)
}

// Validate 验证 Token
func (s *APITokenService) Validate(ctx context.Context, tokenStr string) (*model.APIToken, error) {
	// 检查前缀
	if !strings.HasPrefix(tokenStr, TokenPrefix) {
		return nil, errors.New("invalid token format")
	}

	// 计算 Hash
	hash := sha256.Sum256([]byte(tokenStr))
	tokenHash := hex.EncodeToString(hash[:])

	// 查询 Token
	token, err := s.tokenRepo.GetByTokenHash(ctx, tokenHash)
	if err != nil {
		return nil, errors.New("invalid token")
	}

	// 检查状态
	if token.Status != model.TokenStatusActive {
		return nil, fmt.Errorf("token is %s", token.Status)
	}

	// 检查过期
	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		// 更新状态为过期
		token.Status = model.TokenStatusExpired
		s.tokenRepo.Update(ctx, token)
		return nil, errors.New("token expired")
	}

	// 更新最后使用时间
	s.tokenRepo.UpdateLastUsed(ctx, token.ID)

	return token, nil
}

// HasPermission 检查是否有权限访问指定端点
func (s *APITokenService) HasPermission(token *model.APIToken, path string, method string) bool {
	// 检查内容类型权限
	perms := token.Permissions
	if len(perms.ContentTypes) == 0 {
		return false
	}

	// * 表示全部允许
	for _, ct := range perms.ContentTypes {
		if ct == "*" {
			return true
		}
	}

	// 检查端点权限
	for _, ep := range perms.Endpoints {
		if matchPath(ep.Path, path) && matchMethod(ep.Method, method) {
			return true
		}
	}

	return false
}

// matchPath 匹配路径模式
func matchPath(pattern, path string) bool {
	// 简单实现：支持 * 匹配
	if pattern == "*" {
		return true
	}
	if strings.HasSuffix(pattern, "*") {
		prefix := strings.TrimSuffix(pattern, "*")
		return strings.HasPrefix(path, prefix)
	}
	return pattern == path
}

// matchMethod 匹配 HTTP 方法
func matchMethod(allowed []string, method string) bool {
	for _, m := range allowed {
		if strings.EqualFold(m, method) {
			return true
		}
	}
	return false
}

// Regenerate 重新生成 Token (保留原 Token 的 ID 和权限)
func (s *APITokenService) Regenerate(ctx context.Context, id uuid.UUID) (*model.APIToken, string, error) {
	token, err := s.tokenRepo.GetByID(ctx, id)
	if err != nil {
		return nil, "", err
	}

	// 生成新 Token
	fullToken, tokenHash, prefix, err := s.GenerateToken()
	if err != nil {
		return nil, "", err
	}

	// 更新 Token
	token.TokenHash = tokenHash
	token.TokenPrefix = prefix

	if err := s.tokenRepo.Update(ctx, token); err != nil {
		return nil, "", err
	}

	return token, fullToken, nil
}
