// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

//go:build dm

package database

import (
	"fmt"
	"time"

	dm "gitee.com/chunanyong/dm"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DBType 数据库类型标识
const DBType = "dm"

// DSNConfig 达梦数据库连接参数
type DSNConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string // 数据库名 / 实例名
	Schema   string // 模式名（默认与 User 相同）
	SSLMode  string // 保留兼容性，达梦忽略
}

// GetDSN 构建达梦 DSN
func (c *DSNConfig) GetDSN() string {
	schema := c.Schema
	if schema == "" {
		schema = c.User // 达梦默认 schema = user
	}
	return fmt.Sprintf(
		"dm://%s:%s@%s:%d?schema=%s&compatibleMode=oracle",
		c.User, c.Password, c.Host, c.Port, schema,
	)
}

// Open 打开达梦数据库连接
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
