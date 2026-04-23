//go:build dm

package database

import (
	"fmt"
	"time"

	"gorm.io/driver/dm"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBType 数据库类型标识
const DBType = "dm"

// DSNConfig 达梦 DM8 连接参数
type DSNConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	SSLMode  string // 达梦暂不支持 SSL_MODE，固定为空
}

// GetDSN 构建达梦 DM8 DSN
// 达梦使用类 Oracle 连接格式：user/password@host:port/name
func (c *DSNConfig) GetDSN() string {
	return fmt.Sprintf(
		"dm://%s:%s@%s:%d/%s",
		c.User, c.Password, c.Host, c.Port, c.Name,
	)
}

// Open 打开达梦 DM8 连接
func Open(cfg *DSNConfig, maxOpen, maxIdle int, maxLifetime int) (*gorm.DB, error) {
	dsn := cfg.GetDSN()
	db, err := gorm.Open(dm.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
	return db, nil
}
