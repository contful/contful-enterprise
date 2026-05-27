// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package main

import (
	"os"
	"strings"

	zlog "github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// autoInit 在服务启动时自动检测并初始化数据库。
// 检查 information_schema 中是否存在 contful_ 前缀的表，
// 不存在则自动执行 db/init_pg.sql。
func autoInit(db *gorm.DB) {
	var count int64
	db.Raw(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE 'contful_%'",
	).Scan(&count)

	if count > 0 {
		zlog.Logger.Info().Int64("tables", count).Msg("数据库已初始化，跳过")
		return
	}

	zlog.Logger.Info().Msg("检测到数据库未初始化，自动执行 init_pg.sql...")

	sqlBytes, err := os.ReadFile("../db/init_pg.sql")
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
