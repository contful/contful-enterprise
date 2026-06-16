// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package database

import (
	"os"
	"testing"

	"golang.org/x/crypto/bcrypt"
)

// TestPGFullStack 一条龙验证 PG 连接 → 查询 → bcrypt 校验
// 环境变量: PG_HOST, PG_PORT, PG_USER, PG_PASSWORD, PG_NAME
func TestPGFullStack(t *testing.T) {
	if os.Getenv("PG_HOST") == "" {
		t.Skip("跳过: 设置 PG_HOST 等环境变量后运行")
	}
	cfg := &DSNConfig{
		DBType:   "postgres",
		Host:     envOr("PG_HOST", "localhost"),
		Port:     envInt("PG_PORT", 5432),
		User:     envOr("PG_USER", "postgres"),
		Password: envOr("PG_PASSWORD", "contful123"),
		Name:     envOr("PG_NAME", "contful_ent"),
		SSLMode:  "disable",
	}

	db, err := Open(cfg, 10, 5, 3600)
	if err != nil {
		t.Fatalf("PG Open failed: %v", err)
	}

	t.Run("RawCount", func(t *testing.T) {
		var n int
		if err := db.Raw("SELECT COUNT(*) FROM contful_system_users").Scan(&n).Error; err != nil {
			t.Fatalf("PG Raw COUNT: %v", err)
		}
		t.Logf("Users = %d", n)
	})

	t.Run("FindAdmin", func(t *testing.T) {
		type user struct {
			ID           string
			Email        string
			PasswordHash string
		}
		var u user
		if err := db.Table("contful_system_users").
			Where("email = ?", "admin@contful.com").
			First(&u).Error; err != nil {
			t.Fatalf("PG Find admin: %v", err)
		}
		if u.Email == "" {
			t.Fatal("PG Find: empty fields")
		}
		if len(u.PasswordHash) == 0 {
			t.Fatal("PG Find: password_hash is empty")
		}

		if err := bcrypt.CompareHashAndPassword(
			[]byte(u.PasswordHash),
			[]byte("contful@com"),
		); err != nil {
			t.Fatalf("PG bcrypt: %v (hash=%s)", err, u.PasswordHash)
		}
		t.Logf("PG OK: id=%s email=%s", u.ID, u.Email)
	})

	t.Run("Roles", func(t *testing.T) {
		type role struct{ ID, Name string }
		var roles []role
		if err := db.Table("contful_system_roles").Find(&roles).Error; err != nil {
			t.Fatalf("PG Find roles: %v", err)
		}
		t.Logf("Roles = %d", len(roles))
	})

	t.Run("Sites", func(t *testing.T) {
		type site struct{ ID, Name, Slug string }
		var sites []site
		if err := db.Table("contful_sites").Find(&sites).Error; err != nil {
			t.Fatalf("PG Find sites: %v", err)
		}
		t.Logf("Sites = %d", len(sites))
	})

	t.Run("Config", func(t *testing.T) {
		var count int64
		if err := db.Raw("SELECT COUNT(*) FROM contful_system_config").Scan(&count).Error; err != nil {
			t.Fatalf("PG Config: %v", err)
		}
		t.Logf("Config entries = %d", count)
	})
}
