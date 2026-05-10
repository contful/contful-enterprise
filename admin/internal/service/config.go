// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"errors"

	"github.com/contful/contful/admin/internal/config"
	"github.com/contful/contful/admin/internal/crypto"
)

var (
	ErrCrypterEmpty = errors.New("加密器未初始化")
)

// ConfigService 配置服务（简化版，不再依赖 site_configs 表）
type ConfigService struct {
	crypter crypto.Crypter
	cfg     *config.Config
}

// NewConfigService 新建配置服务
// crypter: 加密器，由 NewCrypter(algorithm, secret) 创建
// cfg: 应用配置（从 config.yaml 加载）
func NewConfigService(crypter crypto.Crypter, cfg *config.Config) *ConfigService {
	return &ConfigService{
		crypter: crypter,
		cfg:     cfg,
	}
}

// GetCrypter 返回加密器（用于 MFA 等其他服务共享）
func (s *ConfigService) GetCrypter() crypto.Crypter {
	return s.crypter
}

// GetAuditSigningKey 获取审计日志签名密钥（从 config.yaml 读取，自动派生）
func (s *ConfigService) GetAuditSigningKey() (string, error) {
	if s.cfg.Audit.SigningKey == "" {
		return "", errors.New("审计签名密钥未配置，请检查 config.yaml 中的 audit.signing_key")
	}
	return s.cfg.Audit.SigningKey, nil
}
