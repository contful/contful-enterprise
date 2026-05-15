// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// SystemConfig 系统配置表模型
type SystemConfig struct {
	ID          uuid.UUID `gorm:"type:uuid;primaryKey;default:gen_random_uuid()" json:"id"`
	ConfigKey   string    `gorm:"type:varchar(100);unique;not null" json:"config_key"`
	ConfigValue string    `gorm:"type:text" json:"config_value"`
	ValueType   string    `gorm:"type:varchar(20);not null;default:'string'" json:"value_type"`
	Description string    `gorm:"type:text" json:"description"`
	IsPublic    bool      `gorm:"not null;default:false" json:"is_public"`
	IsSystem    bool      `gorm:"not null;default:false" json:"is_system"`
	CreatedTime string    `gorm:"type:timestamptz;not null;default:now()" json:"created_time"`
	UpdatedTime string    `gorm:"type:timestamptz;not null;default:now()" json:"updated_time"`
}

func (SystemConfig) TableName() string {
	return "system_config"
}

const (
	configCachePrefix = "system_config:"
	configCacheTTL    = 5 * time.Minute
)

// SystemConfigRepository 系统配置仓库（带 Redis 缓存）
type SystemConfigRepository struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewSystemConfigRepository(db *gorm.DB, rdb *redis.Client) *SystemConfigRepository {
	return &SystemConfigRepository{db: db, redis: rdb}
}

// configCacheKey 生成 Redis 缓存 key
func configCacheKey(key string) string {
	return configCachePrefix + key
}

// GetByKey 根据键名获取配置（Cache-Aside: 先 Redis 后 DB）
func (r *SystemConfigRepository) GetByKey(ctx context.Context, key string) (*SystemConfig, error) {
	// 1. 先查 Redis 缓存
	if r.redis != nil {
		val, err := r.redis.Get(context.Background(), configCacheKey(key)).Result()
		if err == nil {
			var config SystemConfig
			if jsonErr := json.Unmarshal([]byte(val), &config); jsonErr == nil {
				return &config, nil
			}
			// 反序列化失败，删除脏缓存，继续走 DB
			r.redis.Del(context.Background(), configCacheKey(key))
		}
		// redis.Nil 表示缓存 miss，继续走 DB
		if err != nil && err != redis.Nil {
			log.Warn().Err(err).Str("key", key).Msg("redis get config cache failed, falling back to DB")
		}
	}

	// 2. 查 DB
	var config SystemConfig
	err := r.db.WithContext(ctx).Where("config_key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
	}

	// 3. 写入 Redis 缓存
	if r.redis != nil {
		if data, jsonErr := json.Marshal(&config); jsonErr == nil {
			if setErr := r.redis.Set(context.Background(), configCacheKey(key), data, configCacheTTL).Err(); setErr != nil {
				log.Warn().Err(setErr).Str("key", key).Msg("redis set config cache failed")
			}
		}
	}

	return &config, nil
}

// GetInt 获取整数类型的配置值
func (r *SystemConfigRepository) GetInt(ctx context.Context, key string, defaultValue int) int {
	config, err := r.GetByKey(ctx, key)
	if err != nil {
		log.Warn().Err(err).Str("key", key).Int("default", defaultValue).Msg("failed to get config, using default")
		return defaultValue
	}

	if config.ValueType != "number" {
		log.Warn().Str("key", key).Str("type", config.ValueType).Msg("config value type is not number")
		return defaultValue
	}

	val, err := strconv.Atoi(config.ConfigValue)
	if err != nil {
		log.Warn().Err(err).Str("key", key).Str("value", config.ConfigValue).Msg("failed to parse config value as int")
		return defaultValue
	}

	return val
}

// GetString 获取字符串类型的配置值
func (r *SystemConfigRepository) GetString(ctx context.Context, key string, defaultValue string) string {
	config, err := r.GetByKey(ctx, key)
	if err != nil {
		return defaultValue
	}
	return config.ConfigValue
}

// GetBool 获取布尔类型的配置值
func (r *SystemConfigRepository) GetBool(ctx context.Context, key string, defaultValue bool) bool {
	config, err := r.GetByKey(ctx, key)
	if err != nil {
		return defaultValue
	}

	return config.ConfigValue == "true"
}

// List 获取所有配置（不缓存，因为列表变更需要复杂的缓存失效策略）
func (r *SystemConfigRepository) List(ctx context.Context) ([]SystemConfig, error) {
	var configs []SystemConfig
	err := r.db.WithContext(ctx).Order("config_key ASC").Find(&configs).Error
	return configs, err
}

// ListByPublic 根据是否公开获取配置
func (r *SystemConfigRepository) ListByPublic(ctx context.Context, isPublic bool) ([]SystemConfig, error) {
	var configs []SystemConfig
	err := r.db.WithContext(ctx).Where("is_public = ?", isPublic).Order("config_key ASC").Find(&configs).Error
	return configs, err
}

// Set 设置配置值（写入 DB 后删除 Redis 缓存）
func (r *SystemConfigRepository) Set(ctx context.Context, key, value, valueType, description string, isPublic bool) error {
	var config SystemConfig
	err := r.db.WithContext(ctx).Where("config_key = ?", key).First(&config).Error

	if err == gorm.ErrRecordNotFound {
		// 创建新配置
		config = SystemConfig{
			ConfigKey:   key,
			ConfigValue: value,
			ValueType:   valueType,
			Description: description,
			IsPublic:    isPublic,
		}
		if createErr := r.db.WithContext(ctx).Create(&config).Error; createErr != nil {
			return createErr
		}
	} else if err != nil {
		return err
	} else {
		// 更新现有配置
		config.ConfigValue = value
		config.ValueType = valueType
		config.Description = description
		config.IsPublic = isPublic
		if saveErr := r.db.WithContext(ctx).Save(&config).Error; saveErr != nil {
			return saveErr
		}
	}

	// 删除对应的 Redis 缓存
	if r.redis != nil {
		if delErr := r.redis.Del(context.Background(), configCacheKey(key)).Err(); delErr != nil {
			log.Warn().Err(delErr).Str("key", key).Msg("redis del config cache failed after set")
		}
	}

	return nil
}

// Create 创建新配置（仅自定义配置）
func (r *SystemConfigRepository) Create(ctx context.Context, key, value, valueType, description string, isPublic bool) (*SystemConfig, error) {
	config := SystemConfig{
		ConfigKey:   key,
		ConfigValue: value,
		ValueType:   valueType,
		Description: description,
		IsPublic:    isPublic,
		IsSystem:    false,
	}
	if err := r.db.WithContext(ctx).Create(&config).Error; err != nil {
		return nil, err
	}
	return &config, nil
}

// Delete 删除配置（仅非系统配置可删除）
func (r *SystemConfigRepository) Delete(ctx context.Context, key string) error {
	result := r.db.WithContext(ctx).Where("config_key = ? AND is_system = FALSE", key).Delete(&SystemConfig{})
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return gorm.ErrRecordNotFound
	}
	// 删除 Redis 缓存
	if r.redis != nil {
		r.redis.Del(context.Background(), configCacheKey(key))
	}
	return nil
}

// ClearCache 清除所有 system_config 的 Redis 缓存
func (r *SystemConfigRepository) ClearCache(ctx context.Context) error {
	if r.redis == nil {
		return nil
	}
	keys, err := r.redis.Keys(context.Background(), configCachePrefix+"*").Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return r.redis.Del(context.Background(), keys...).Err()
	}
	return nil
}
