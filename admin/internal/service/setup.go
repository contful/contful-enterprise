// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/contful/contful/admin/internal/crypto"
	"github.com/contful/contful/admin/internal/model"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SetupService 安装向导业务逻辑。
// 注意：Setup 期间不使用 Repository 层，Service 直接操作 GORM。
type SetupService struct {
	db        *gorm.DB    // 服务器的 DB 连接（config.yaml 配置的）
	initSQL   string      // init_pg.sql 完整内容
	cryptoMode string     // 加密模式（rsa / sm）
	confDir   string      // 配置目录（密钥写入路径）
}

// NewSetupService 创建 SetupService 实例。
func NewSetupService(db *gorm.DB, initSQL, cryptoMode, confDir string) *SetupService {
	return &SetupService{
		db:         db,
		initSQL:    initSQL,
		cryptoMode: cryptoMode,
		confDir:    confDir,
	}
}

// Version 返回当前 Contful 版本号。
const Version = "1.3.0"

// CheckStatus 检查是否需要安装。
func (s *SetupService) CheckStatus() (*model.SetupStatusResponse, error) {
	hasTable := s.db.Migrator().HasTable("contful_system_users")
	var count int64
	if hasTable {
		s.db.Table("contful_system_users").Count(&count)
	}
	return &model.SetupStatusResponse{
		SetupRequired: !hasTable || count == 0,
		Version:       Version,
	}, nil
}

// TestDatabase 测试用户提供的数据库连接是否可用。
func (s *SetupService) TestDatabase(req *model.SetupDatabaseRequest) error {
	dsn := buildDSN(req.Host, req.Port, req.User, req.Password, req.DBName, req.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("数据库连接失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层连接失败: %w", err)
	}
	defer sqlDB.Close()

	if err := sqlDB.Ping(); err != nil {
		return fmt.Errorf("数据库 ping 失败: %w", err)
	}

	return nil
}

// InitializeDatabase 在目标数据库执行 init_pg.sql 完整初始化脚本。
// 使用 advisory lock 防止并发初始化。
func (s *SetupService) InitializeDatabase(req *model.SetupDatabaseRequest) error {
	dsn := buildDSN(req.Host, req.Port, req.User, req.Password, req.DBName, req.SSLMode)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return fmt.Errorf("连接目标数据库失败: %w", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		return fmt.Errorf("获取底层连接失败: %w", err)
	}
	defer sqlDB.Close()

	// 获取 advisory lock（与 Docker entrypoint.sh 一致，使用 lock ID 12345）
	if err := db.Exec("SELECT pg_try_advisory_lock(12345)").Error; err != nil {
		return fmt.Errorf("获取初始化锁失败: %w", err)
	}
	defer db.Exec("SELECT pg_advisory_unlock(12345)")

	// 按语句分割 SQL 并逐条执行
	statements := splitSQL(s.initSQL)
	for i, stmt := range statements {
		stmt = strings.TrimSpace(stmt)
		if stmt == "" || strings.HasPrefix(stmt, "--") {
			continue
		}
		if err := db.Exec(stmt).Error; err != nil {
			return fmt.Errorf("执行 SQL 语句 #%d 失败: %w", i+1, err)
		}
	}

	return nil
}

// CreateAdmin 将默认管理员账号替换为用户提供的凭据，并创建默认站点。
func (s *SetupService) CreateAdmin(req *model.SetupAdminRequest) error {
	// 生成 bcrypt 密码哈希（cost=10）
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return fmt.Errorf("密码加密失败: %w", err)
	}

	// 更新种子数据中的默认管理员（id = 00000000-0000-0000-0000-000000000001）
	result := s.db.Table("contful_system_users").
		Where("email = ?", "admin@contful.com").
		Updates(map[string]interface{}{
			"email":         req.Email,
			"password_hash": string(hash),
		})
	if result.Error != nil {
		return fmt.Errorf("更新管理员账号失败: %w", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.New("未找到默认管理员记录，请确认 init_pg.sql 已成功执行")
	}

	// 创建默认站点
	now := time.Now()
	siteResult := s.db.Table("contful_sites").Create(map[string]interface{}{
		"id":           "00000000-0000-0000-0000-000000000002",
		"name":         req.SiteName,
		"slug":         req.SiteSlug,
		"locale":       "zh-CN",
		"timezone":     "Asia/Shanghai",
		"is_active":    true,
		"created_time": now,
		"updated_time": now,
	})
	if siteResult.Error != nil {
		return fmt.Errorf("创建默认站点失败: %w", siteResult.Error)
	}

	// 标记安装完成
	if err := s.markSetupCompleted(); err != nil {
		return fmt.Errorf("标记安装完成失败: %w", err)
	}

	return nil
}

// GenerateKeys 生成非对称密钥对并写入 conf/keys/ 目录。
// 如果密钥文件已存在则跳过（与 Docker entrypoint.sh ensure_keys() 行为一致）。
func (s *SetupService) GenerateKeys() error {
	keysDir := filepath.Join(s.confDir, "keys")
	if err := os.MkdirAll(keysDir, 0700); err != nil {
		return fmt.Errorf("创建密钥目录失败: %w", err)
	}

	pubPath := filepath.Join(keysDir, "public.pem")
	privPath := filepath.Join(keysDir, "private.pem")

	// 如果密钥文件已存在，跳过生成
	if _, err := os.Stat(pubPath); err == nil {
		if _, err := os.Stat(privPath); err == nil {
			return nil // 密钥对已存在
		}
	}

	// 根据 crypto_mode 选择加密器
	asym, err := crypto.NewAsymmetricCrypter(s.cryptoMode)
	if err != nil {
		return fmt.Errorf("创建非对称加密器失败: %w", err)
	}

	pubPEM, privPEM, err := asym.GenerateKeyPair()
	if err != nil {
		return fmt.Errorf("生成密钥对失败: %w", err)
	}

	if err := os.WriteFile(pubPath, []byte(pubPEM), 0644); err != nil {
		return fmt.Errorf("写入公钥文件失败: %w", err)
	}
	if err := os.WriteFile(privPath, []byte(privPEM), 0600); err != nil {
		return fmt.Errorf("写入私钥文件失败: %w", err)
	}

	return nil
}

// markSetupCompleted 在 contful_system_config 表中插入 setup_completed 标记。
func (s *SetupService) markSetupCompleted() error {
	return s.db.Exec(`
		INSERT INTO contful_system_config (config_key, config_value, value_type, is_public, is_system)
		VALUES ('setup_completed', 'true', 'string', false, true)
		ON CONFLICT (config_key) DO NOTHING
	`).Error
}

// buildDSN 构建 PostgreSQL DSN 连接字符串。
func buildDSN(host string, port int, user, password, dbname, sslmode string) string {
	if sslmode == "" {
		sslmode = "disable"
	}
	return fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		host, port, user, password, dbname, sslmode,
	)
}

// splitSQL 将 SQL 文本按语句分割（以分号分隔，保留多行语句）。
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
				// 检查是否是转义的单引号
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
				// 匹配到闭合 dollar quote
				current.WriteString(dollarTag)
				i += len(dollarTag) - 1
				inDollar = false
				continue
			}
			current.WriteByte(ch)
			continue
		}

		// 检测 dollar-quoted string（如 $$...$$, $func$...$func$）
		if ch == '$' && i+1 < len(sql) {
			// 查找结束 $
			j := i + 1
			for j < len(sql) && sql[j] != '$' {
				j++
			}
			if j < len(sql) {
				tag := sql[i : j+1] // e.g., "$$" or "$func$"
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

	// 处理最后一条语句（可能没有分号结尾）
	remaining := strings.TrimSpace(current.String())
	if remaining != "" {
		statements = append(statements, remaining)
	}

	return statements
}
