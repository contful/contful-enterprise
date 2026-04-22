package model

import (
	"time"

	"github.com/google/uuid"
)

// ConfigType 配置值类型
type ConfigType string

const (
	ConfigTypeString     ConfigType = "string"
	ConfigTypeNumber     ConfigType = "number"
	ConfigTypeBoolean   ConfigType = "boolean"
	ConfigTypeJSON       ConfigType = "json"
	ConfigTypeEncrypted  ConfigType = "encrypted"
)

// ConfigGroup 配置分组（保留组名）
type ConfigGroup string

const (
	ConfigGroupDefault   ConfigGroup = "default"
	ConfigGroupStorage   ConfigGroup = "storage"
	ConfigGroupMail      ConfigGroup = "mail"
	ConfigGroupOAuth     ConfigGroup = "oauth"
	ConfigGroupPayment   ConfigGroup = "payment"
	ConfigGroupFeature  ConfigGroup = "feature"
	ConfigGroupIntegrity ConfigGroup = "integrity"
)

// SiteConfig 站点配置
type SiteConfig struct {
	ID          uuid.UUID  `json:"id" gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SiteID      uuid.UUID  `json:"site_id" gorm:"type:uuid;not null;index"`
	ConfigKey   string     `json:"config_key" gorm:"type:varchar(128);not null;index:uk_site_config_key,priority:2"`
	ConfigValue string     `json:"config_value,omitempty" gorm:"type:text"`
	ConfigType  ConfigType `json:"config_type" gorm:"type:varchar(32);default:'string'"`
	ConfigGroup ConfigGroup `json:"config_group" gorm:"type:varchar(64);default:'default';index:idx_site_configs_group,priority:2"`
	IsEncrypted bool       `json:"is_encrypted" gorm:"default:false"`
	IsReadonly  bool       `json:"is_readonly" gorm:"default:false"`
	Description string     `json:"description" gorm:"type:varchar(255)"`
	CreatedTime time.Time  `json:"created_time" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedTime time.Time  `json:"updated_time" gorm:"type:timestamptz;not null;default:now()"`
	UpdatedBy   *uuid.UUID `json:"updated_by,omitempty" gorm:"type:uuid"`

	// 复合唯一索引 (site_id, config_key)
	// GORM 通过 uk_site_config_key 实现，priority:1 = site_id
}

func (SiteConfig) TableName() string {
	return "site_configs"
}

// SiteConfigListFilter 列表查询过滤
type SiteConfigListFilter struct {
	Group string `form:"group"`
	Key   string `form:"key"`
}

// CreateSiteConfig 创建参数
type CreateSiteConfig struct {
	ConfigKey   string     `json:"config_key" binding:"required"`
	ConfigValue string     `json:"config_value" binding:"required"`
	ConfigType  ConfigType `json:"config_type"`
	ConfigGroup ConfigGroup `json:"config_group"`
	IsEncrypted bool       `json:"is_encrypted"`
	IsReadonly  bool       `json:"is_readonly"`
	Description string     `json:"description"`
}

// UpdateSiteConfig 更新参数
type UpdateSiteConfig struct {
	ConfigValue string `json:"config_value" binding:"required"`
}
