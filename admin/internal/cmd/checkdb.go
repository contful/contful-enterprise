// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"database/sql"
	"fmt"
	"os"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// CheckDB 检查数据库连接和核心表是否存在。
//
// 退出码:
//
//	0 — 数据库可达且至少存在一张核心表（system_users, sites, system_config, system_roles 之一）
//	1 — 数据库可达但核心表不存在
//	2 — 数据库连接失败
func CheckDB() {
	dsn := buildDSN()
	if dsn == "" {
		fmt.Fprintln(os.Stderr, "CHECK_DB: missing required env vars (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)")
		os.Exit(2)
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "CHECK_DB: failed to open database: %v\n", err)
		os.Exit(2)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "CHECK_DB: database unreachable: %v\n", err)
		os.Exit(2)
	}

	query := `SELECT COUNT(*) FROM information_schema.tables
		WHERE table_name IN ('contful_system_users','contful_sites','contful_system_config','contful_system_roles')
		AND table_schema = 'public'
		AND table_type = 'BASE TABLE'`

	var count int
	if err := db.QueryRow(query).Scan(&count); err != nil {
		fmt.Fprintf(os.Stderr, "CHECK_DB: query failed: %v\n", err)
		os.Exit(2)
	}

	if count > 0 {
		fmt.Fprintf(os.Stderr, "CHECK_DB: OK — %d core table(s) found\n", count)
		os.Exit(0)
	}
	fmt.Fprintln(os.Stderr, "CHECK_DB: NO core tables found — database needs initialization")
	os.Exit(1)
}

// buildDSN 从环境变量构建 PostgreSQL DSN。
func buildDSN() string {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")
	sslmode := os.Getenv("DB_SSLMODE")

	if host == "" || port == "" || user == "" || password == "" || name == "" {
		return ""
	}
	if sslmode == "" {
		sslmode = "disable"
	}

	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=5",
		host, port, user, password, name, sslmode,
	)
}
