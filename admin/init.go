// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"fmt"
	"os"
	"regexp"
	"strings"

	zlog "github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// initSQLPaths 按优先级列出 init_pg.sql 的可能位置。
var initSQLPaths = []string{
	"db/init_pg.sql",        // dev.sh / Docker（工作目录 = 项目根）
	"../db/init_pg.sql",     // 源码编译（go run，工作目录 = admin/）
	"/app/db/init_pg.sql",   // Docker 容器内
}

// entSQLPaths 按优先级列出 init_ent.sql（企业版增量）的可能位置。
var entSQLPaths = []string{
	"db/init_ent.sql",
	"../db/init_ent.sql",
	"/app/db/init_ent.sql",
}

// readInitSQL 按优先级尝试多个路径读取 init_pg.sql。
func readInitSQL() ([]byte, error) {
	for _, p := range initSQLPaths {
		b, err := os.ReadFile(p)
		if err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf("在所有路径中未找到 init_pg.sql: %v", initSQLPaths)
}

// readEntSQL 按优先级尝试多个路径读取 init_ent.sql。
func readEntSQL() ([]byte, error) {
	for _, p := range entSQLPaths {
		b, err := os.ReadFile(p)
		if err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf("在所有路径中未找到 init_ent.sql: %v", entSQLPaths)
}

// initEnterprise 检测并执行企业版增量初始化。
func initEnterprise(db *gorm.DB) {
	var entCount int64
	db.Raw(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE 'contful_ent_%'",
	).Scan(&entCount)

	if entCount > 0 {
		zlog.Logger.Info().Int64("tables", entCount).Msg("企业版表已存在，跳过")
		return
	}

	runSQLFile(db, "init_ent.sql", readEntSQL)
}

// runSQLFile 读取并执行 SQL 文件（带 advisory lock 和注释剥离）。
func runSQLFile(db *gorm.DB, name string, reader func() ([]byte, error)) {
	sqlBytes, err := reader()
	if err != nil {
		zlog.Logger.Error().Err(err).Str("file", name).Msg("无法读取 SQL 文件")
		return
	}

	if err := db.Exec("SELECT pg_try_advisory_lock(12345)").Error; err != nil {
		zlog.Logger.Error().Err(err).Msg("获取初始化锁失败")
		return
	}
	defer db.Exec("SELECT pg_advisory_unlock(12345)")

	zlog.Logger.Info().Str("file", name).Msg("开始执行 SQL 脚本...")
	sql := stripComments(string(sqlBytes))
	statements := splitSQL(sql)
	failed := 0

	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" {
			continue
		}
		if err := db.Exec(stmt).Error; err != nil {
			zlog.Logger.Error().Err(err).Int("statement", i+1).Str("file", name).Msg("SQL 执行失败")
			failed++
		}
	}

	if failed > 0 {
		zlog.Logger.Error().Int("failed", failed).Str("file", name).Msg("部分语句执行失败")
	} else {
		zlog.Logger.Info().Str("file", name).Msg("执行完成")
	}
}
func stripComments(sql string) string {
	// 移除 /* */ 块注释
	sql = regexp.MustCompile(`/\*[\s\S]*?\*/`).ReplaceAllString(sql, "")
	// 移除以 -- 开头的行（注释行），保留空行占位
	lines := strings.Split(sql, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

// autoInit 在服务启动时自动检测并初始化数据库。
// 检查 information_schema 中是否存在 contful_ 前缀的表，
// 不存在则自动执行 init_pg.sql + init_ent.sql。
// 如果基础表已存在但企业版表缺失，仅执行 init_ent.sql。
func autoInit(db *gorm.DB) {
	var count int64
	db.Raw(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE 'contful_%'",
	).Scan(&count)

	if count > 0 {
		zlog.Logger.Info().Int64("tables", count).Msg("数据库已初始化")
		// 检查企业版表是否需要初始化
		initEnterprise(db)
		return
	}

	// 全新安装：先社区版再企业版
	runSQLFile(db, "init_pg.sql", readInitSQL)
	initEnterprise(db)
}

// splitSQL 将 SQL 文本按语句分割（支持字符串和 dollar-quote）。
func splitSQL(sql string) []string {
	var statements []string
	current := strings.Builder{}
	inString := false
	inDollar := false
	dollarTag := ""

	for i := 0; i < len(sql); i++ {
		ch := sql[i]

		if inString {
			current.WriteByte(ch)
			if ch == '\'' {
				if i+1 < len(sql) && sql[i+1] == '\'' {
					current.WriteByte(sql[i+1])
					i++
				} else {
					inString = false
				}
			}
			continue
		}

		if inDollar {
			if ch == '$' && i+len(dollarTag) <= len(sql) && sql[i:i+len(dollarTag)] == dollarTag {
				current.WriteString(dollarTag)
				i += len(dollarTag) - 1
				inDollar = false
				continue
			}
			current.WriteByte(ch)
			continue
		}

		if ch == '$' && i+1 < len(sql) {
			j := i + 1
			for j < len(sql) && sql[j] != '$' {
				j++
			}
			if j < len(sql) {
				tag := sql[i : j+1]
				current.WriteString(tag)
				i = j
				inDollar = true
				dollarTag = tag
				continue
			}
		}

		if ch == '\'' {
			current.WriteByte(ch)
			inString = true
			continue
		}

		if ch == ';' {
			current.WriteByte(ch)
			statements = append(statements, current.String())
			current.Reset()
			continue
		}

		current.WriteByte(ch)
	}

	remaining := strings.TrimSpace(current.String())
	if remaining != "" {
		statements = append(statements, remaining)
	}
	return statements
}
