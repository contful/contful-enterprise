// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/crypto"
	"github.com/contful/contful/admin/internal/repository"
)

var (
	ErrConfigNotFound    = errors.New("配置不存在")
	ErrConfigReadonly    = errors.New("配置为只读，禁止修改")
	ErrCrypterEmpty      = errors.New("加密器未初始化")
	ErrInvalidConfigType = errors.New("无效的 config_type")
)

// ConfigService 站点配置服务（含加解密）
type ConfigService struct {
	repo    *repository.SiteConfigRepository
	crypter crypto.Crypter
}

// NewConfigService 新建配置服务
// crypter: 加密器，由 NewCrypter(algorithm, secret) 创建
func NewConfigService(repo *repository.SiteConfigRepository, crypter crypto.Crypter) *ConfigService {
	return &ConfigService{
		repo:    repo,
		crypter: crypter,
	}
}

// Get 读取配置，自动解密加密字段
func (s *ConfigService) Get(ctx context.Context, siteID uuid.UUID, key string) (string, error) {
	cfg, err := s.repo.GetByKey(ctx, siteID, key)
	if err != nil {
		return "", ErrConfigNotFound
	}
	if cfg.IsEncrypted {
		return s.decrypt(cfg.ConfigValue)
	}
	return cfg.ConfigValue, nil
}

// Set 写入配置，自动加密（如果 is_encrypted=true）
// updatedBy 传 nil 则不更新 updated_by 字段
func (s *ConfigService) Set(ctx context.Context, siteID uuid.UUID, key, value string, opts *model.CreateSiteConfig, updatedBy *uuid.UUID) error {
	if s.crypter == nil {
		return ErrCrypterEmpty
	}

	// 检查是否存在 + 只读状态
	if existing, err := s.repo.GetByKey(ctx, siteID, key); err == nil {
		if existing.IsReadonly {
			return ErrConfigReadonly
		}
	}

	cfg := &model.SiteConfig{
		SiteID:      siteID,
		ConfigKey:   key,
		ConfigValue: value,
		ConfigType:  opts.ConfigType,
		ConfigGroup: opts.ConfigGroup,
		IsEncrypted: opts.IsEncrypted,
		IsReadonly:  opts.IsReadonly,
		Description: opts.Description,
		UpdatedBy:   updatedBy,
	}

	// 自动加密
	if cfg.ConfigType == "" {
		cfg.ConfigType = model.ConfigTypeString
	}
	if cfg.IsEncrypted || cfg.ConfigType == model.ConfigTypeEncrypted {
		cfg.IsEncrypted = true
		cfg.ConfigType = model.ConfigTypeEncrypted
		encrypted, err := s.encrypt(value)
		if err != nil {
			return fmt.Errorf("加密配置值失败: %w", err)
		}
		cfg.ConfigValue = encrypted
	}

	return s.repo.Upsert(ctx, cfg)
}

// Delete 删除配置（readonly 配置禁止删除）
func (s *ConfigService) Delete(ctx context.Context, siteID uuid.UUID, key string) error {
	cfg, err := s.repo.GetByKey(ctx, siteID, key)
	if err != nil {
		return ErrConfigNotFound
	}
	if cfg.IsReadonly {
		return ErrConfigReadonly
	}
	return s.repo.Delete(ctx, siteID, key)
}

// ListByGroup 按组列出配置
func (s *ConfigService) ListByGroup(ctx context.Context, siteID uuid.UUID, group string) ([]model.SiteConfig, error) {
	if group == "" {
		return s.repo.ListBySite(ctx, siteID)
	}
	return s.repo.ListByGroup(ctx, siteID, group)
}

// ListAll 列出站点所有配置（加密字段返回占位符）
func (s *ConfigService) ListAll(ctx context.Context, siteID uuid.UUID) ([]model.SiteConfig, error) {
	configs, err := s.repo.ListBySite(ctx, siteID)
	if err != nil {
		return nil, err
	}
	// 加密字段返回占位符，不暴露密文
	for i := range configs {
		if configs[i].IsEncrypted {
			configs[i].ConfigValue = "••••••••"
		}
	}
	return configs, nil
}

// encrypt 加密
func (s *ConfigService) encrypt(plaintext string) (string, error) {
	return s.crypter.Encrypt([]byte(plaintext))
}

// decrypt 解密
func (s *ConfigService) decrypt(ciphertext string) (string, error) {
	plaintext, err := s.crypter.Decrypt(ciphertext)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

// GetCrypter 返回加密器（用于 MFA 等其他服务共享）
func (s *ConfigService) GetCrypter() crypto.Crypter {
	return s.crypter
}
