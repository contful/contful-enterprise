// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package cmd

import (
	"database/sql"
	"fmt"
	"os"

	_ "gitee.com/chunanyong/dm"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// CheckDB 检查数据库连接和核心表是否存在（支持 PG + DM）。
//
// 退出码:
//
//	0 — 数据库可达且至少存在一张核心表
//	1 — 数据库可达但核心表不存在
//	2 — 数据库连接失败 / 配置缺失
func CheckDB() {
	dbType := os.Getenv("DB_TYPE")
	if dbType == "" {
		dbType = "postgres"
	}

	dsn, err := buildDSN(dbType)
	if err != nil {
		fmt.Fprintln(os.Stderr, "CHECK_DB:", err)
		os.Exit(2)
	}

	driverName := driverFor(dbType)
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		fmt.Fprintf(os.Stderr, "CHECK_DB: failed to open database: %v\n", err)
		os.Exit(2)
	}
	defer db.Close()

	if err := db.Ping(); err != nil {
		fmt.Fprintf(os.Stderr, "CHECK_DB: database unreachable: %v\n", err)
		os.Exit(2)
	}

	query := coreTableQuery(dbType)
	var count int
	if err := db.QueryRow(query).Scan(&count); err != nil {
		fmt.Fprintf(os.Stderr, "CHECK_DB: query failed: %v\n", err)
		os.Exit(2)
	}

	if count > 0 {
		fmt.Fprintf(os.Stderr, "CHECK_DB: OK — %d core table(s) found (type=%s)\n", count, dbType)
		os.Exit(0)
	}
	fmt.Fprintln(os.Stderr, "CHECK_DB: NO core tables found — database needs initialization")
	os.Exit(1)
}

func driverFor(dbType string) string {
	if dbType == "dm" {
		return "dm"
	}
	return "pgx"
}

func coreTableQuery(dbType string) string {
	tables := "'CONTFUL_SYSTEM_USERS','CONTFUL_SITES','CONTFUL_SYSTEM_CONFIG','CONTFUL_SYSTEM_ROLES'"
	if dbType == "dm" {
		// DM8: ALL_TABLES, schema 大写
		schema := os.Getenv("DB_SCHEMA")
		if schema == "" {
			schema = os.Getenv("DB_USER")
		}
		return fmt.Sprintf(`SELECT COUNT(*) FROM ALL_TABLES WHERE TABLE_NAME IN (%s) AND OWNER = '%s'`, tables, schema)
	}
	// PostgreSQL: information_schema, schema 小写 public
	return fmt.Sprintf(`SELECT COUNT(*) FROM information_schema.tables WHERE table_name IN ('contful_system_users','contful_sites','contful_system_config','contful_system_roles') AND table_schema = 'public' AND table_type = 'BASE TABLE'`)
}

func buildDSN(dbType string) (string, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	name := os.Getenv("DB_NAME")

	if host == "" || port == "" || user == "" || password == "" || name == "" {
		return "", fmt.Errorf("missing required env vars (DB_HOST, DB_PORT, DB_USER, DB_PASSWORD, DB_NAME)")
	}

	if dbType == "dm" {
		schema := os.Getenv("DB_SCHEMA")
		if schema == "" {
			schema = user
		}
		return fmt.Sprintf("dm://%s:%s@%s:%s?schema=%s&compatibleMode=oracle", user, password, host, port, schema), nil
	}

	sslmode := os.Getenv("DB_SSLMODE")
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s connect_timeout=5",
		host, port, user, password, name, sslmode), nil
}
