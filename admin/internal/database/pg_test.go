package database

import (
	"testing"
	"golang.org/x/crypto/bcrypt"
)

func TestPGFullStack(t *testing.T) {
	cfg := &DSNConfig{
		DBType:   "postgres",
		Host:     "47.123.4.8",
		Port:     5432,
		User:     "postgres",
		Password: "3edc&YGV",
		Name:     "contful_ent",
		SSLMode:  "disable",
	}

	db, err := Open(cfg, 10, 5, 3600)
	if err != nil {
		t.Fatalf("PG Open failed: %v", err)
	}

	// Step 1: count users
	t.Run("RawCount", func(t *testing.T) {
		var n int
		if err := db.Raw("SELECT COUNT(*) FROM contful_system_users").Scan(&n).Error; err != nil {
			t.Fatalf("PG Raw COUNT: %v", err)
		}
		t.Logf("Users = %d", n)
	})

	// Step 2: find admin
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

		// bcrypt
		if err := bcrypt.CompareHashAndPassword(
			[]byte(u.PasswordHash),
			[]byte("contful@com"),
		); err != nil {
			t.Fatalf("PG bcrypt: %v (hash=%s)", err, u.PasswordHash)
		}
		t.Logf("PG OK: id=%s email=%s", u.ID, u.Email)
	})

	// Step 3: roles
	t.Run("Roles", func(t *testing.T) {
		type role struct{ ID, Name string }
		var roles []role
		if err := db.Table("contful_system_roles").Find(&roles).Error; err != nil {
			t.Fatalf("PG Find roles: %v", err)
		}
		t.Logf("Roles = %d", len(roles))
	})

	// Step 4: sites
	t.Run("Sites", func(t *testing.T) {
		type site struct{ ID, Name, Slug string }
		var sites []site
		if err := db.Table("contful_sites").Find(&sites).Error; err != nil {
			t.Fatalf("PG Find sites: %v", err)
		}
		t.Logf("Sites = %d", len(sites))
	})

	// Step 5: config
	t.Run("Config", func(t *testing.T) {
		var count int64
		if err := db.Raw("SELECT COUNT(*) FROM contful_system_config").Scan(&count).Error; err != nil {
			t.Fatalf("PG Config: %v", err)
		}
		t.Logf("Config entries = %d", count)
	})
}
