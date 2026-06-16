// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package database

import (
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestDMFullStack 验证 DM 连接 → 查询 → bcrypt
// 环境变量: DM_HOST, DM_PORT, DM_USER, DM_PASSWORD, DM_NAME, DM_SCHEMA
func TestDMFullStack(t *testing.T) {
	if os.Getenv("DM_HOST") == "" && os.Getenv("PG_HOST") == "" {
		t.Skip("无 DB 环境变量，跳过集成测试")
	}
	cfg := &DSNConfig{
		DBType:   "dm",
		Host:     envOr("DM_HOST", "139.198.171.102"),
		Port:     envInt("DM_PORT", 5236),
		User:     envOr("DM_USER", "SYSDBA"),
		Password: envOr("DM_PASSWORD", "SYSDBA008"),
		Name:     envOr("DM_NAME", "CONTFUL_ENT"),
		Schema:   envOr("DM_SCHEMA", "CONTFUL_ENT"),
	}

	db, err := Open(cfg, 10, 5, 3600)
	if err != nil {
		t.Fatalf("Open failed: %v", err)
	}

	db.Config.PrepareStmt = false

	t.Run("RawExists", func(t *testing.T) {
		var n int
		if err := db.Raw("SELECT COUNT(*) FROM contful_system_users").Scan(&n).Error; err != nil {
			t.Fatalf("Raw COUNT failed: %v", err)
		}
		t.Logf("Users count = %d", n)
	})

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
			t.Fatal("Find returned empty fields")
		}
		t.Logf("Found: id=%s email=%s hash_len=%d", u.ID, u.Email, len(u.PasswordHash))

		if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte("contful@com")); err != nil {
			t.Fatalf("bcrypt mismatch: %v (hash=[%s])", err, u.PasswordHash)
		}
		t.Log("bcrypt OK")
	})
}
