// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"context"
	"database/sql"
	"fmt"
	"regexp"
	"strings"
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

var currentDBType = "postgres"

func CurrentDBType() string { return currentDBType }

type DSNConfig struct {
	DBType, Host, User, Password, Name, Schema, SSLMode string
	Port                                                 int
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

// ============================ 达梦 DM8 ============================

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

	db, err := gorm.Open(&dmDialector{Conn: sqlDB}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
		// DM8 元数据返回大写列名，NamingStrategy 需匹配
		NamingStrategy: dmNamingStrategy{},
	})
	if err != nil {
		sqlDB.Close()
		return nil, err
	}

	// 替换 ConnPool 为 SQL 改写代理层
	db.ConnPool = &dmConnPool{db: sqlDB}
	db.Statement.ConnPool = db.ConnPool

	currentDBType = "dm"
	uid.SetDBType("dm")
	return db, nil
}

// ============================ DM Dialector ============================

type dmDialector struct{ Conn *sql.DB }

func (d dmDialector) Name() string                                       { return "dm" }
func (d dmDialector) Explain(_ string, _ ...interface{}) string          { return "" }
func (d dmDialector) DefaultValueOf(*schema.Field) clause.Expression     { return nil }
func (d dmDialector) Initialize(db *gorm.DB) error                       { db.ConnPool = d.Conn; return nil }
func (d dmDialector) Migrator(db *gorm.DB) gorm.Migrator                 { return migrator.Migrator{Config: migrator.Config{DB: db}} }
func (d dmDialector) BindVarTo(w clause.Writer, _ *gorm.Statement, v interface{}) {
	w.WriteByte('?')
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

// ============================ DM ConnPool Proxy (SQL Rewrite) ============================

var (
	dmQuoteRe = regexp.MustCompile(`"([^"]+)"`)
	dmLimitRe = regexp.MustCompile(`\bLIMIT\s+(\d+)(\s+OFFSET\s+(\d+))?\b`)
)

func dmFixSQL(sql string) string {
	sql = dmQuoteRe.ReplaceAllString(sql, "$1")
	return dmLimitRe.ReplaceAllStringFunc(sql, func(m string) string {
		parts := dmLimitRe.FindStringSubmatch(m)
		if parts == nil {
			return m
		}
		if parts[3] != "" {
			return fmt.Sprintf("OFFSET %s ROWS FETCH NEXT %s ROWS ONLY", parts[3], parts[1])
		}
		return fmt.Sprintf("FETCH FIRST %s ROWS ONLY", parts[1])
	})
}

type dmConnPool struct{ db *sql.DB }

func (p *dmConnPool) PrepareContext(ctx context.Context, query string) (*sql.Stmt, error) {
	return p.db.PrepareContext(ctx, dmFixSQL(query))
}
func (p *dmConnPool) ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error) {
	return p.db.ExecContext(ctx, dmFixSQL(query), args...)
}
func (p *dmConnPool) QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
	return p.db.QueryContext(ctx, dmFixSQL(query), args...)
}
func (p *dmConnPool) QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row {
	return p.db.QueryRowContext(ctx, dmFixSQL(query), args...)
}

// dmNamingStrategy: 列名转大写 snake_case 匹配 DM8 元数据
// GORM 默认 PasswordHash → password_hash，DM8 返回 PASSWORD_HASH
type dmNamingStrategy struct{ schema.NamingStrategy }

func (s dmNamingStrategy) ColumnName(table, column string) string {
	return strings.ToUpper(s.NamingStrategy.ColumnName(table, column))
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
