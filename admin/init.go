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

// stripComments 移除以 -- 开头的行注释和 /* */ 块注释。
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
//
// 触发条件（两个必须同时满足）：
//  1. 环境变量 CONTFUL_AUTO_INIT=true
//  2. information_schema 中不存在 contful_ 前缀的表
//
// CONTFUL_AUTO_INIT 默认为 false，避免每次启动都查 information_schema。
// 首次部署时设置 CONFTUL_AUTO_INIT=true，初始化完成后可移除该变量。
func autoInit(db *gorm.DB) {
	if os.Getenv("CONTFUL_AUTO_INIT") != "true" {
		zlog.Logger.Debug().Msg("CONTFUL_AUTO_INIT != true，跳过自动初始化")
		return
	}

	var count int64
	if err := db.Raw(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE 'contful_%'",
	).Scan(&count).Error; err != nil {
		zlog.Logger.Warn().Err(err).Msg("无法查询 information_schema，跳过自动初始化（DB 可能未就绪）")
		return
	}

	if count > 0 {
		zlog.Logger.Info().Int64("tables", count).Msg("数据库已初始化，跳过")
		return
	}

	zlog.Logger.Info().Msg("检测到数据库未初始化，自动执行 init_pg.sql...")

	sqlBytes, err := readInitSQL()
	if err != nil {
		zlog.Logger.Error().Err(err).Msg("无法读取 init_pg.sql，自动初始化失败")
		return
	}

	// advisory lock 防止多实例并发初始化
	if err := db.Exec("SELECT pg_try_advisory_lock(12345)").Error; err != nil {
		zlog.Logger.Error().Err(err).Msg("获取初始化锁失败")
		return
	}
	defer db.Exec("SELECT pg_advisory_unlock(12345)")

	sql := string(sqlBytes)
	// 先剥离注释再按分号分割，避免注释内的分号被误切
	sql = stripComments(sql)
	statements := splitSQL(sql)
	failed := 0

	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if err := db.Exec(stmt).Error; err != nil {
			zlog.Logger.Error().Err(err).Int("statement", i+1).Msg("SQL 执行失败")
			failed++
		}
	}

	if failed > 0 {
		zlog.Logger.Error().Int("failed", failed).Msg("init_pg.sql 部分语句执行失败")
		return
	}

	zlog.Logger.Info().Msg("数据库初始化完成：默认管理员 admin@contful.com / contful@com")
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
