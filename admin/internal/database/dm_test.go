// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package database

import (
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestDMFullStack 一条龙验证 DM 连接 → 查询 → bcrypt 校验
func TestDMFullStack(t *testing.T) {
	cfg := &DSNConfig{
		DBType:   "dm",
		Host:     "139.198.171.102",
		Port:     5236,
		User:     "SYSDBA",
		Password: "SYSDBA008",
		Name:     "CONTFUL_ENT",
		Schema:   "CONTFUL_ENT",
	}

	db, err := Open(cfg, 10, 5, 3600)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	// 临时为测试禁用 prepared statement（复现 service 行为）
	db.Config.PrepareStmt = false

	// Step 1: raw query — 验 SQL 改写和连接
	t.Run("RawExists", func(t *testing.T) {
		var n int
		if err := db.Raw("SELECT COUNT(*) FROM contful_system_users").Scan(&n).Error; err != nil {
			t.Fatalf("Raw COUNT failed: %v", err)
		}
		t.Logf("Users count = %d", n)
	})

	// Step 2: model Find — 验 Schema 映射
	t.Run("FindAdmin", func(t *testing.T) {
		type user struct {
			ID           string
			Email        string
			PasswordHash string
		}
		var u user
		if err := db.Table("contful_system_users").
			Where("EMAIL = ?", "admin@contful.com").
			First(&u).Error; err != nil {
			t.Fatalf("Find admin failed: %v", err)
		}
		if u.Email == "" {
			t.Fatal("Find returned empty fields — schema mapping broken")
		}
		t.Logf("Found: id=%s email=%s hash_len=%d", u.ID, u.Email, len(u.PasswordHash))

		// Step 3: bcrypt — 验密码
		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("contful@com")); err != nil {
			t.Fatalf("bcrypt mismatch: %v (hash=[%s])", err, u.PasswordHash)
		}
		t.Log("bcrypt OK")
	})
}
