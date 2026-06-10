// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

//go:build dm

package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "gitee.com/chunanyong/dm"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

// DBType 数据库类型标识
const DBType = "dm"

// DSNConfig 达梦数据库连接参数
type DSNConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	Schema   string
	SSLMode  string // 保留兼容性
}

// GetDSN 构建达梦 DSN
func (c *DSNConfig) GetDSN() string {
	schema := c.Schema
	if schema == "" {
		schema = c.User
	}
	return fmt.Sprintf("dm://%s:%s@%s:%d?schema=%s&compatibleMode=oracle",
		c.User, c.Password, c.Host, c.Port, schema)
}

// Open 打开达梦数据库连接（通过 database/sql + GORM）
func Open(cfg *DSNConfig, maxOpen, maxIdle int, maxLifetime int) (*gorm.DB, error) {
	dsn := cfg.GetDSN()

	sqlDB, err := sql.Open("dm", dsn)
	if err != nil {
		return nil, fmt.Errorf("dm open: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("dm ping: %w", err)
	}
	sqlDB.SetMaxIdleConns(maxIdle)
	sqlDB.SetMaxOpenConns(maxOpen)
	sqlDB.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)

	db, err := gorm.Open(&dmDialector{Conn: sqlDB}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		sqlDB.Close()
		return nil, err
	}
	return db, nil
}

// dmDialector 达梦 GORM Dialector
type dmDialector struct {
	Conn *sql.DB
}

func (d dmDialector) Name() string { return "dm" }

func (d dmDialector) Initialize(db *gorm.DB) error {
	return nil
}

func (d dmDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return migrator.Migrator{Config: migrator.Config{DB: db}}
}

func (d dmDialector) DataTypeOf(field *schema.Field) string {
	switch field.DataType {
	case schema.Bool:
		return "CHAR(1)"
	case schema.Int, schema.Uint:
		return "INT"
	case schema.Float:
		return "FLOAT"
	case schema.String:
		return "VARCHAR(255)"
	case schema.Time:
		return "TIMESTAMP"
	case schema.Bytes:
		return "BLOB"
	default:
		return "VARCHAR(255)"
	}
}

func (d dmDialector) DefaultValueOf(*schema.Field) clause.Expression { return nil }

func (d dmDialector) BindVarTo(writer clause.Writer, _ *gorm.Statement, _ interface{}) {
	writer.WriteByte('?')
}

func (d dmDialector) QuoteTo(writer clause.Writer, s string) {
	writer.WriteByte('"')
	writer.WriteString(s)
	writer.WriteByte('"')
}

func (d dmDialector) Explain(_ string, _ ...interface{}) string { return "" }
