// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/contful/contful/admin/internal/model"
	"gorm.io/gorm"
)

var (
	ErrUserNotFound      = errors.New("user not found")
	ErrUserAlreadyExists = errors.New("user already exists")
	ErrInvalidPassword   = errors.New("invalid password")
	ErrUserInactive      = errors.New("user is inactive")
	ErrUserSuspended     = errors.New("user is suspended")
)

type UserRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewUserRepository(db *gorm.DB, redis *redis.Client) *UserRepository {
	return &UserRepository{db: db, redis: redis}
}

// Create 创建用户
func (r *UserRepository) Create(ctx context.Context, user *model.SystemUser) error {
	result := r.db.WithContext(ctx).Create(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return ErrUserAlreadyExists
		}
		return result.Error
	}
	return nil
}

// FindByID 根据 ID 查找用户
func (r *UserRepository) FindByID(ctx context.Context, id uuid.UUID) (*model.SystemUser, error) {
	var user model.SystemUser
	result := r.db.WithContext(ctx).Where("id = ? AND deleted_time IS NULL", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindByEmail 根据邮箱查找用户
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*model.SystemUser, error) {
	var user model.SystemUser
	result := r.db.WithContext(ctx).Where("email = ? AND deleted_time IS NULL", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// Update 更新用户信息
func (r *UserRepository) Update(ctx context.Context, user *model.SystemUser) error {
	return r.db.WithContext(ctx).
		Where("id = ?", user.ID).
		Updates(user).Error
}

// Delete 软删除用户
func (r *UserRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&model.SystemUser{}).Error
}

// PermanentDelete 永久删除用户（绕过软删除）
func (r *UserRepository) PermanentDelete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Where("id = ?", id).
		Delete(&model.SystemUser{}).Error
}

// Restore 恢复软删除的用户
func (r *UserRepository) Restore(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).
		Unscoped().
		Model(&model.SystemUser{}).
		Where("id = ?", id).
		Update("deleted_at", nil).Error
}

// FindByIDWithDeleted 根据 ID 查找用户（包含已删除的）
func (r *UserRepository) FindByIDWithDeleted(ctx context.Context, id uuid.UUID) (*model.SystemUser, error) {
	var user model.SystemUser
	result := r.db.WithContext(ctx).Unscoped().Where("id = ?", id).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// FindByEmailWithDeleted 根据邮箱查找用户（包含已删除的）
func (r *UserRepository) FindByEmailWithDeleted(ctx context.Context, email string) (*model.SystemUser, error) {
	var user model.SystemUser
	result := r.db.WithContext(ctx).Unscoped().Where("email = ?", email).First(&user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, ErrUserNotFound
		}
		return nil, result.Error
	}
	return &user, nil
}

// UpdateLastLogin 更新最后登录信息
func (r *UserRepository) UpdateLastLogin(ctx context.Context, id uuid.UUID, ip string) error {
	now := time.Now()
	updates := map[string]interface{}{
		"last_login_time": now,
	}
	// inet 类型不接受空字符串，只有有效 IP 才更新
	if ip != "" {
		updates["last_login_ip"] = ip
	}
	return r.db.WithContext(ctx).Model(&model.SystemUser{}).
		Where("id = ?", id).
		Updates(updates).Error
}

// List 查询用户列表（分页，可包含已删除记录）
func (r *UserRepository) List(ctx context.Context, page, pageSize int, includeDeleted bool) ([]model.SystemUser, int64, error) {
	var users []model.SystemUser
	var total int64

	db := r.db.WithContext(ctx).Model(&model.SystemUser{})
	
	if !includeDeleted {
		db = db.Where("deleted_time IS NULL")
	}
	
	if err := db.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	if err := db.Offset(offset).Limit(pageSize).Order("created_time DESC").Find(&users).Error; err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// Refresh Token 管理
const refreshTokenPrefix = "refresh_token:"
const refreshTokenTTL = 7 * 24 * time.Hour // 7 days

// StoreRefreshToken 存储 Refresh Token 到 Redis
func (r *UserRepository) StoreRefreshToken(ctx context.Context, userID uuid.UUID, token string) error {
	key := refreshTokenPrefix + token
	return r.redis.Set(ctx, key, userID.String(), refreshTokenTTL).Err()
}

// ValidateRefreshToken 验证 Refresh Token
func (r *UserRepository) ValidateRefreshToken(ctx context.Context, token string) (uuid.UUID, error) {
	key := refreshTokenPrefix + token
	userIDStr, err := r.redis.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return uuid.Nil, errors.New("invalid refresh token")
		}
		return uuid.Nil, err
	}
	return uuid.Parse(userIDStr)
}

// DeleteRefreshToken 删除 Refresh Token
func (r *UserRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	key := refreshTokenPrefix + token
	return r.redis.Del(ctx, key).Err()
}

// DeleteAllUserRefreshTokens 删除用户所有 Refresh Token（登出所有设备）
// P2-003 修复：使用 SCAN + COUNT 分段处理，避免阻塞 Redis
func (r *UserRepository) DeleteAllUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	pattern := refreshTokenPrefix + "*"
	var cursor uint64
	const batchSize = 100 // 每次 SCAN 最多返回 100 个 key

	for {
		var keys []string
		var err error
		keys, cursor, err = r.redis.Scan(ctx, cursor, pattern, batchSize).Result()
		if err != nil {
			return err
		}

		for _, token := range keys {
			userIDStr, err := r.redis.Get(ctx, token).Result()
			if err != nil {
				continue
			}
			if userIDStr == userID.String() {
				if err := r.redis.Del(ctx, token).Err(); err != nil {
					log.Error().Err(err).Str("token", token).Msg("failed to delete refresh token")
				}
			}
		}

		// cursor == 0 表示遍历完成
		if cursor == 0 {
			break
		}
	}
	return nil
}

// ============================================
// MFA 相关方法
// ============================================

// UpdateMFASecret 写入 TOTP Secret 和 Recovery Code（Setup 阶段，mfa_enabled 保持不变）
func (r *UserRepository) UpdateMFASecret(ctx context.Context, userID uuid.UUID, encryptedSecret, encryptedCodes string) error {
	return r.db.WithContext(ctx).Model(&model.SystemUser{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"totp_secret":    encryptedSecret,
			"recovery_codes": encryptedCodes,
			"updated_time":   time.Now(),
		}).Error
}

// UpdateMFAEnabled 设置 mfa_enabled 字段
func (r *UserRepository) UpdateMFAEnabled(ctx context.Context, userID uuid.UUID, enabled bool) error {
	return r.db.WithContext(ctx).Model(&model.SystemUser{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"mfa_enabled":  enabled,
			"updated_time": time.Now(),
		}).Error
}

// ClearMFA 清除 MFA 相关字段（禁用 MFA）
func (r *UserRepository) ClearMFA(ctx context.Context, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Model(&model.SystemUser{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"mfa_enabled":    false,
			"totp_secret":    nil,
			"recovery_codes": nil,
			"updated_time":   time.Now(),
		}).Error
}

// UpdateRecoveryCodes 更新 Recovery Code（使用后标记）
func (r *UserRepository) UpdateRecoveryCodes(ctx context.Context, userID uuid.UUID, encryptedCodes string) error {
	return r.db.WithContext(ctx).Model(&model.SystemUser{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"recovery_codes": encryptedCodes,
			"updated_time":   time.Now(),
		}).Error
}

// UpdateAvatarURL 更新用户头像地址
func (r *UserRepository) UpdateAvatarURL(ctx context.Context, userID uuid.UUID, avatarURL string) error {
	return r.db.WithContext(ctx).Model(&model.SystemUser{}).
		Where("id = ?", userID).
		Updates(map[string]interface{}{
			"avatar_url":   avatarURL,
			"updated_time": time.Now(),
		}).Error
}
