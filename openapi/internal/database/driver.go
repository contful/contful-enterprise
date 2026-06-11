// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "gitee.com/chunanyong/dm"
	"github.com/contful/contful/openapi/pkg/uid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

// DBType 当前数据库类型（运行时由 Open 设置）
var currentDBType = "postgres"

// CurrentDBType 返回当前数据库类型
func CurrentDBType() string { return currentDBType }

// DSNConfig 数据库连接参数（PG + DM 通用）
type DSNConfig struct {
	DBType   string
	Host     string
	Port     int
	User     string
	Password string
	Name     string
	Schema   string
	SSLMode  string
}

// Open 打开数据库连接，根据 DBType 选择 PG 或 DM 驱动
func Open(cfg *DSNConfig, maxOpen, maxIdle int, maxLifetime int) (*gorm.DB, error) {
	switch cfg.DBType {
	case "dm":
		return openDM(cfg, maxOpen, maxIdle, maxLifetime)
	default:
		return openPG(cfg, maxOpen, maxIdle, maxLifetime)
	}
}

// ============================ PostgreSQL ============================

func openPG(cfg *DSNConfig, maxOpen, maxIdle int, maxLifetime int) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		return nil, err
	}
	applyDBPool(db, maxOpen, maxIdle, maxLifetime)
	currentDBType = "postgres"
	uid.SetDBType("postgres")
	return db, nil
}

// ============================ 达梦 DM8 ============================

func openDM(cfg *DSNConfig, maxOpen, maxIdle int, maxLifetime int) (*gorm.DB, error) {
	schema := cfg.Schema
	if schema == "" {
		schema = cfg.User
	}
	dsn := fmt.Sprintf("dm://%s:%s@%s:%d?schema=%s&compatibleMode=oracle",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, schema)

	sqlDB, err := sql.Open("dm", dsn)
	if err != nil {
		return nil, fmt.Errorf("dm open: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("dm ping: %w", err)
	}
	applyDBPoolSQL(sqlDB, maxOpen, maxIdle, maxLifetime)

	db, err := gorm.Open(&dmDialector{Conn: sqlDB}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	if err != nil {
		sqlDB.Close()
		return nil, err
	}
	currentDBType = "dm"
	uid.SetDBType("dm")
	return db, nil
}

func applyDBPool(db *gorm.DB, maxOpen, maxIdle int, maxLifetime int) {
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.SetMaxIdleConns(maxIdle)
		sqlDB.SetMaxOpenConns(maxOpen)
		sqlDB.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
	}
}

func applyDBPoolSQL(db *sql.DB, maxOpen, maxIdle int, maxLifetime int) {
	db.SetMaxIdleConns(maxIdle)
	db.SetMaxOpenConns(maxOpen)
	db.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
}

// ============================ DM GORM Dialector ============================

type dmDialector struct {
	Conn *sql.DB
}

func (d dmDialector) Name() string                           { return "dm" }
func (d dmDialector) Initialize(_ *gorm.DB) error             { return nil }
func (d dmDialector) Explain(_ string, _ ...interface{}) string { return "" }

func (d dmDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return migrator.Migrator{Config: migrator.Config{DB: db}}
}

func (d dmDialector) BindVarTo(writer clause.Writer, _ *gorm.Statement, _ interface{}) {
	writer.WriteByte('?')
}

func (d dmDialector) QuoteTo(writer clause.Writer, s string) {
	// DM8 默认大小写不敏感，双引号反而强制区分大小写
	// GORM 生成小写表名 → 不加引号让 DM8 自动匹配大写表名
	writer.WriteString(s)
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
