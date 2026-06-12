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
	"github.com/contful/contful/admin/pkg/uid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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

	// Use PostgreSQL Dialector (correct model query building) + DM driver
	db, err := gorm.Open(postgres.New(postgres.Config{
		Conn:                 sqlDB,
		DriverName:           "dm",
		PreferSimpleProtocol: true, // 禁用 prepared statement → ConnPool proxy 生效
	}), &gorm.Config{
		Logger:         logger.Default.LogMode(logger.Silent),
		NamingStrategy: dmNamingStrategy{},
		PrepareStmt:    false, // 禁用 GORM 级别的 prepared statement → ConnPool 代理可拦截
	})
	if err != nil {
		sqlDB.Close()
		return nil, err
	}

	// Wrap ConnPool with SQL rewriting proxy and force-set on every Statement
	proxy := &dmConnPool{db: sqlDB}
	db.ConnPool = proxy
	hook := func(d *gorm.DB) { d.Statement.ConnPool = d.ConnPool }
	db.Callback().Query().Before("gorm:query").Register("dm:cp", hook)
	db.Callback().Row().Before("gorm:row").Register("dm:cp", hook)
	db.Callback().Raw().Before("gorm:raw").Register("dm:cp", hook)

	currentDBType = "dm"
	uid.SetDBType("dm")
	return db, nil
}

// ============================ SQL Rewrite ============================

var (
	dmQuoteRe = regexp.MustCompile(`"([^"]+)"`)
	dmLimitRe = regexp.MustCompile(`\bLIMIT\s+(\$?\d+)(\s+OFFSET\s+(\$?\d+))?\b`)
	dmDollarN = regexp.MustCompile(`\$\d+`) // PG dialect $N → DM driver ?
)

func dmFixSQL(sql string) string {
	// 0. Replace $N → ? (PG dialect → DM driver)
	sql = dmDollarN.ReplaceAllString(sql, "?")
	// 1. Strip double quotes
	sql = dmQuoteRe.ReplaceAllString(sql, "$1")
	// 2. LIMIT/OFFSET → FETCH FIRST
	return dmLimitRe.ReplaceAllStringFunc(sql, func(m string) string {
		parts := dmLimitRe.FindStringSubmatch(m)
		if parts == nil { return m }
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

// ============================ Naming Strategy ============================

type dmNamingStrategy struct{ schema.NamingStrategy }

func (s dmNamingStrategy) ColumnName(table, column string) string {
	return strings.ToUpper(s.NamingStrategy.ColumnName(table, column))
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
