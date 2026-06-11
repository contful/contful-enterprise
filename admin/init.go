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

// initSQLPaths 按优先级列出 init_pg.sql 的可能位置（仅 PostgreSQL）。
var initSQLPaths = []string{
	"db/init_pg.sql",
	"../db/init_pg.sql",
	"/app/db/init_pg.sql",
}

// entSQLPaths 按优先级列出 init_ent_pg.sql 的可能位置（企业版 PostgreSQL 增量）。
var entSQLPaths = []string{
	"db/init_ent_pg.sql",
	"../db/init_ent_pg.sql",
	"/app/db/init_ent_pg.sql",
}

// dmEntSQLPaths 按优先级列出 init_ent_dm.sql 的可能位置（企业版达梦增量）。
var dmEntSQLPaths = []string{
	"db/init_ent_dm.sql",
	"../db/init_ent_dm.sql",
	"/app/db/init_ent_dm.sql",
}

// dmSQLPaths 按优先级列出 init_dm.sql 的可能位置（达梦社区版基础）。
var dmSQLPaths = []string{
	"db/init_dm.sql",
	"../db/init_dm.sql",
	"/app/db/init_dm.sql",
}

// dmSeedPaths 按优先级列出 seed_ent_dm.sql 的可能位置。
var dmSeedPaths = []string{
	"db/seed_ent_dm.sql",
	"../db/seed_ent_dm.sql",
	"/app/db/seed_ent_dm.sql",
}

// isDM 判断当前是否为达梦数据库
func isDM() bool { return os.Getenv("DB_TYPE") == "dm" }

// readInitSQL 按优先级尝试多个路径读取 init_pg.sql。
func readInitSQL() ([]byte, error) { return readFirst(initSQLPaths, "init_pg.sql") }

// readEntSQL 读取 init_ent_pg.sql。
func readEntSQL() ([]byte, error) { return readFirst(entSQLPaths, "init_ent_pg.sql") }

// readDMEntSQL 读取 init_ent_dm.sql。
func readDMEntSQL() ([]byte, error) { return readFirst(dmEntSQLPaths, "init_ent_dm.sql") }

// readDMSQL 读取 init_dm.sql。
func readDMSQL() ([]byte, error) { return readFirst(dmSQLPaths, "init_dm.sql") }

// readDMSeed 读取 seed_ent_dm.sql。
func readDMSeed() ([]byte, error) { return readFirst(dmSeedPaths, "seed_ent_dm.sql") }

func readFirst(paths []string, name string) ([]byte, error) {
	for _, p := range paths {
		b, err := os.ReadFile(p)
		if err == nil {
			return b, nil
		}
	}
	return nil, fmt.Errorf("未找到 %s: %v", name, paths)
}

// initEnterprise 检测并执行企业版增量初始化。
func initEnterprise(db *gorm.DB) {
	if isDM() {
		initEnterpriseDM(db)
		return
	}
	initEnterprisePG(db)
}

func initEnterprisePG(db *gorm.DB) {
	var entCount int64
	db.Raw(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE 'contful_ent_%'",
	).Scan(&entCount)

	if entCount > 0 {
		zlog.Logger.Info().Int64("tables", entCount).Msg("企业版表已存在，跳过")
		return
	}
	runSQLFile(db, "init_ent_pg.sql", readEntSQL)
}

func initEnterpriseDM(db *gorm.DB) {
	var entCount int64
	db.Raw(
		"SELECT COUNT(*) FROM ALL_TABLES WHERE OWNER = 'CONTFUL_ENT' AND TABLE_NAME LIKE 'CONTFUL_AUDIT_%'",
	).Scan(&entCount)

	if entCount > 0 {
		zlog.Logger.Info().Int64("tables", entCount).Msg("企业版表已存在，跳过")
		return
	}
	runSQLFile(db, "init_ent_dm.sql", readDMEntSQL)
}

// runSQLFile 读取并执行 SQL 文件。
func runSQLFile(db *gorm.DB, name string, reader func() ([]byte, error)) {
	sqlBytes, err := reader()
	if err != nil {
		zlog.Logger.Error().Err(err).Str("file", name).Msg("无法读取 SQL 文件")
		return
	}

	// pg_try_advisory_lock 仅 PostgreSQL，达梦跳过
	if !isDM() {
		if err := db.Exec("SELECT pg_try_advisory_lock(12345)").Error; err != nil {
			zlog.Logger.Error().Err(err).Msg("获取初始化锁失败")
			return
		}
		defer db.Exec("SELECT pg_advisory_unlock(12345)")
	}

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
	sql = regexp.MustCompile(`/\*[\s\S]*?\*/`).ReplaceAllString(sql, "")
	lines := strings.Split(sql, "\n")
	var result []string
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "--") {
			continue
		}
		if idx := strings.Index(line, " -- "); idx >= 0 {
			line = line[:idx]
		}
		result = append(result, line)
	}
	return strings.Join(result, "\n")
}

// autoInit 在服务启动时自动检测并初始化数据库。
func autoInit(db *gorm.DB) {
	if os.Getenv("CONTFUL_AUTO_INIT") != "true" {
		zlog.Logger.Debug().Msg("CONTFUL_AUTO_INIT != true，跳过自动初始化")
		return
	}

	if isDM() {
		autoInitDM(db)
		return
	}
	autoInitPG(db)
}

func autoInitPG(db *gorm.DB) {
	var count int64
	if err := db.Raw(
		"SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public' AND table_name LIKE 'contful_%'",
	).Scan(&count).Error; err != nil {
		zlog.Logger.Warn().Err(err).Msg("无法查询 information_schema，跳过自动初始化")
		return
	}
	if count > 0 {
		zlog.Logger.Info().Int64("tables", count).Msg("数据库已初始化")
		initEnterprisePG(db)
		return
	}
	runSQLFile(db, "init_pg.sql", readInitSQL)
	initEnterprisePG(db)
}

func autoInitDM(db *gorm.DB) {
	var count int64
	if err := db.Raw(
		"SELECT COUNT(*) FROM ALL_TABLES WHERE OWNER = 'CONTFUL_ENT' AND TABLE_NAME LIKE 'CONTFUL_%'",
	).Scan(&count).Error; err != nil {
		zlog.Logger.Warn().Err(err).Msg("无法查询 ALL_TABLES，跳过自动初始化")
		return
	}
	if count > 0 {
		zlog.Logger.Info().Int64("tables", count).Msg("达梦数据库已初始化")
		return
	}
	// 先社区版基础表，再企业版增量，最后种子数据
	runSQLFile(db, "init_dm.sql", readDMSQL)
	runSQLFile(db, "init_ent_dm.sql", readDMEntSQL)
	runSQLFile(db, "seed_ent_dm.sql", readDMSeed)
}

// splitSQL 将 SQL 文本按语句分割（支持字符串和 dollar-quote）。
func splitSQL(sql string) []string {
	// 达梦 Oracle 语句用 / 分隔（PG 用 ;）
	if isDM() {
		return splitSQLOracle(sql)
	}
	return splitSQLPG(sql)
}

func splitSQLPG(sql string) []string {
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
			} else {
				current.WriteByte(ch)
			}
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

// splitSQLOracle 按 / 分隔 Oracle/DM PL/SQL 块
func splitSQLOracle(sql string) []string {
	return strings.FieldsFunc(sql, func(r rune) bool { return r == '/' })
}
