// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package database

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/lib/pq"
)

// MigrateConfig 迁移配置
type MigrateConfig struct {
	MigrationsPath string // 迁移文件路径（如：file://migrations）
	DatabaseURL   string // 数据库连接 URL
}

// RunMigrations 运行数据库迁移
func RunMigrations(cfg *MigrateConfig) error {
	// 创建迁移源（从文件系统读取迁移文件）
	source, err := (&file.File{}).Open(cfg.MigrationsPath)
	if err != nil {
		return fmt.Errorf("failed to open migrations source: %w", err)
	}
	defer source.Close()

	// 连接数据库
	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}
	defer db.Close()

	// 创建 PostgreSQL 驱动
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	// 创建迁移实例
	m, err := migrate.NewWithInstance("file", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// 运行迁移
	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	return nil
}

// RollbackMigration 回滚一个版本
func RollbackMigration(cfg *MigrateConfig) error {
	source, err := (&file.File{}).Open(cfg.MigrationsPath)
	if err != nil {
		return fmt.Errorf("failed to open migrations source: %w", err)
	}
	defer source.Close()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("file", source, "postgres", driver)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	// 回滚一个版本
	if err := m.Steps(-1); err != nil {
		return fmt.Errorf("failed to rollback migration: %w", err)
	}

	return nil
}

// GetMigrationVersion 获取当前迁移版本
func GetMigrationVersion(cfg *MigrateConfig) (uint, bool, error) {
	source, err := (&file.File{}).Open(cfg.MigrationsPath)
	if err != nil {
		return 0, false, fmt.Errorf("failed to open migrations source: %w", err)
	}
	defer source.Close()

	db, err := sql.Open("postgres", cfg.DatabaseURL)
	if err != nil {
		return 0, false, fmt.Errorf("failed to connect database: %w", err)
	}
	defer db.Close()

	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return 0, false, fmt.Errorf("failed to create postgres driver: %w", err)
	}

	m, err := migrate.NewWithInstance("file", source, "postgres", driver)
	if err != nil {
		return 0, false, fmt.Errorf("failed to create migrate instance: %w", err)
	}

	version, dirty, err := m.Version()
	if err != nil {
		return 0, false, fmt.Errorf("failed to get version: %w", err)
	}

	return version, dirty, nil
}
