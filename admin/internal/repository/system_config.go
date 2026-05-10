// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package repository

import (
	"context"
	"strconv"

	"github.com/google/uuid"
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
	CreatedTime string    `gorm:"type:timestamptz;not null;default:now()" json:"created_time"`
	UpdatedTime string    `gorm:"type:timestamptz;not null;default:now()" json:"updated_time"`
}

func (SystemConfig) TableName() string {
	return "system_config"
}

// SystemConfigRepository 系统配置仓库
type SystemConfigRepository struct {
	db *gorm.DB
}

func NewSystemConfigRepository(db *gorm.DB) *SystemConfigRepository {
	return &SystemConfigRepository{db: db}
}

// GetByKey 根据键名获取配置
func (r *SystemConfigRepository) GetByKey(ctx context.Context, key string) (*SystemConfig, error) {
	var config SystemConfig
	err := r.db.WithContext(ctx).Where("config_key = ?", key).First(&config).Error
	if err != nil {
		return nil, err
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

// List 获取所有配置
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

// Set 设置配置值
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
		return r.db.WithContext(ctx).Create(&config).Error
	}

	if err != nil {
		return err
	}

	// 更新现有配置
	config.ConfigValue = value
	config.ValueType = valueType
	config.Description = description
	config.IsPublic = isPublic
	return r.db.WithContext(ctx).Save(&config).Error
}
