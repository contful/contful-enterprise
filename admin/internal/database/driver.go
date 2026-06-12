// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"database/sql"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "gitee.com/chunanyong/dm"
	"github.com/contful/contful/admin/pkg/uid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/migrator"
	"gorm.io/gorm/schema"
)

var currentDBType = "postgres"

func CurrentDBType() string { return currentDBType }

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

func Open(cfg *DSNConfig, maxOpen, maxIdle int, maxLifetime int) (*gorm.DB, error) {
	if cfg.DBType == "dm" {
		return openDM(cfg, maxOpen, maxIdle, maxLifetime)
	}
	return openPG(cfg, maxOpen, maxIdle, maxLifetime)
}

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

func openDM(cfg *DSNConfig, maxOpen, maxIdle int, maxLifetime int) (*gorm.DB, error) {
	schemaName := cfg.Schema
	if schemaName == "" {
		schemaName = cfg.User
	}
	dsn := fmt.Sprintf("dm://%s:%s@%s:%d?schema=%s&compatibleMode=oracle",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, schemaName)

	sqlDB, err := sql.Open("dm", dsn)
	if err != nil {
		return nil, fmt.Errorf("dm open: %w", err)
	}
	if err := sqlDB.Ping(); err != nil {
		sqlDB.Close()
		return nil, fmt.Errorf("dm ping: %w", err)
	}
	setPoolSQL(sqlDB, maxOpen, maxIdle, maxLifetime)

	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:       sqlDB,
		DriverName: "dm",
	}), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		sqlDB.Close()
		return nil, err
	}

	// 改写 LIMIT/OFFSET + 脱引号
	registerDMCallbacks(db)

	currentDBType = "dm"
	uid.SetDBType("dm")
	return db, nil
}

// ============================ DM Query Rewrite Callbacks ============================

var limitRe = regexp.MustCompile(`\bLIMIT\s+(\d+)(?:\s+OFFSET\s+(\d+))?\b`)
var quoteRe = regexp.MustCompile(`"([a-z][a-z0-9_]*)"`) // 匹配带引号的小写标识符

func registerDMCallbacks(db *gorm.DB) {
	// 回调 1: 去掉双引号（PG Dialector 默认加引号 → DM8 区分大小写）
	db.Callback().Query().Before("gorm:query").Register("dm:unquote", func(d *gorm.DB) {
		sql := d.Statement.SQL.String()
		// 只去表名列名的引号，保留字符串内的引号
		sql = quoteRe.ReplaceAllString(sql, "$1")
		d.Statement.SQL.Reset()
		d.Statement.SQL.WriteString(sql)
	})

	// 回调 2: LIMIT/OFFSET → Oracle FETCH FIRST / OFFSET … ROWS FETCH NEXT
	db.Callback().Query().Before("gorm:query").Register("dm:limit", func(d *gorm.DB) {
		sql := d.Statement.SQL.String()
		sql = limitRe.ReplaceAllStringFunc(sql, func(m string) string {
			m = strings.ToUpper(m)
			parts := limitRe.FindStringSubmatch(m)
			if parts == nil {
				return m
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

	// 回调 3: 同样处理 Row 查询
	db.Callback().Row().Before("gorm:row").Register("dm:row_unquote", func(d *gorm.DB) {
		sql := d.Statement.SQL.String()
		sql = quoteRe.ReplaceAllString(sql, "$1")
		d.Statement.SQL.Reset()
		d.Statement.SQL.WriteString(sql)
	})
	db.Callback().Row().Before("gorm:row").Register("dm:row_limit", func(d *gorm.DB) {
		sql := d.Statement.SQL.String()
		sql = limitRe.ReplaceAllStringFunc(sql, func(m string) string {
			parts := limitRe.FindStringSubmatch(m)
			if parts == nil { return m }
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

// 以下类型和方法不再需要，但保留以避免编译错误
type dmDialector struct{ Conn *sql.DB }

func (d dmDialector) Name() string                             { return "dm" }
func (d dmDialector) Initialize(_ *gorm.DB) error               { return nil }
func (d dmDialector) Explain(_ string, _ ...interface{}) string { return "" }
func (d dmDialector) Migrator(db *gorm.DB) gorm.Migrator {
	return migrator.Migrator{Config: migrator.Config{DB: db}}
}
func (d dmDialector) BindVarTo(writer clause.Writer, _ *gorm.Statement, _ interface{}) {
	writer.WriteByte('?')
}
func (d dmDialector) QuoteTo(writer clause.Writer, s string) { writer.WriteString(s) }
func (d dmDialector) DataTypeOf(field *schema.Field) string {
	switch field.DataType {
	case schema.Bool: return "CHAR(1)"
	case schema.Int, schema.Uint: return "INT"
	case schema.Float: return "FLOAT"
	case schema.String: return "VARCHAR(255)"
	case schema.Time: return "TIMESTAMP"
	case schema.Bytes: return "BLOB"
	default: return "VARCHAR(255)"
	}
}
func (d dmDialector) DefaultValueOf(*schema.Field) clause.Expression { return nil }
