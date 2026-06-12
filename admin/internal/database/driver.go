// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"time"

	_ "gitee.com/chunanyong/dm"
	"github.com/contful/contful/admin/pkg/uid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// currentDBType 运行时数据库类型
var currentDBType = "postgres"

// CurrentDBType 返回当前数据库类型
func CurrentDBType() string { return currentDBType }

// DSNConfig 数据库连接参数
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

// Open 打开数据库连接
func Open(cfg *DSNConfig, maxOpen, maxIdle int, maxLifetime int) (*gorm.DB, error) {
	if cfg.DBType == "dm" {
		return openDM(cfg, maxOpen, maxIdle, maxLifetime)
	}
	return openPG(cfg, maxOpen, maxIdle, maxLifetime)
}

// ============================ PostgreSQL ============================

func openPG(cfg *DSNConfig, maxOpen, maxIdle int, maxLifetime int) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.Password, cfg.Name, cfg.SSLMode)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		return nil, err
	}
	setPool(db, maxOpen, maxIdle, maxLifetime)
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
	setPoolSQL(sqlDB, maxOpen, maxIdle, maxLifetime)

	// 使用 PG Dialector + DM driver — GORM 的模型查询生成 PG 语法
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "dm",
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		sqlDB.Close()
		return nil, err
	}

	// 注册 callback 将 LIMIT/OFFSET 改写为 DM8 兼容语法
	registerDMCallbacks(db)

	currentDBType = "dm"
	uid.SetDBType("dm")
	return db, nil
}

// ============================ DM LIMIT/OFFSET 改写 ============================

var limitRe = regexp.MustCompile(`(?i)\bLIMIT\s+(\d+)(?:\s+OFFSET\s+(\d+))?`)

func registerDMCallbacks(db *gorm.DB) {
	_ = db.Callback().Query().Before("gorm:query").Register("dm:rewrite_limit", func(d *gorm.DB) {
		sql := d.Statement.SQL.String()
		sql = limitRe.ReplaceAllStringFunc(sql, func(match string) string {
			parts := limitRe.FindStringSubmatch(match)
			if parts == nil {
				return match
			}
			limit, _ := strconv.Atoi(parts[1])
			if parts[2] != "" {
				offset, _ := strconv.Atoi(parts[2])
				return fmt.Sprintf("OFFSET %d ROWS FETCH NEXT %d ROWS ONLY", offset, limit)
			}
			return fmt.Sprintf("FETCH FIRST %d ROWS ONLY", limit)
		})
		d.Statement.SQL.Reset()
		d.Statement.SQL.WriteString(sql)
	})
}

// ============================ Pool ============================

func setPool(db *gorm.DB, maxOpen, maxIdle int, maxLifetime int) {
	sqlDB, _ := db.DB()
	if sqlDB != nil {
		sqlDB.SetMaxIdleConns(maxIdle)
		sqlDB.SetMaxOpenConns(maxOpen)
		sqlDB.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
	}
}

func setPoolSQL(db *sql.DB, maxOpen, maxIdle int, maxLifetime int) {
	db.SetMaxIdleConns(maxIdle)
	db.SetMaxOpenConns(maxOpen)
	db.SetConnMaxLifetime(time.Duration(maxLifetime) * time.Second)
}
