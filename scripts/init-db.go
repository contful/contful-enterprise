// +build ignore

package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	user := getEnv("DB_USER", "postgres")
	pass := getEnv("DB_PASSWORD", "")
	ssl := getEnv("DB_SSL_MODE", "disable")

	// 先连接默认 postgres 数据库检查/创建 contful
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=postgres sslmode=%s",
		host, port, user, pass, ssl)

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接 postgres 失败: %v", err)
	}
	log.Println("✅ 数据库连接成功！")

	// 检查 contful 数据库是否存在
	var exists bool
	err = db.Raw("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname='contful')").Scan(&exists).Error
	if err != nil {
		log.Fatalf("检查数据库失败: %v", err)
	}

	if exists {
		log.Println("ℹ️  数据库 'contful' 已存在，跳过创建")
	} else {
		result := db.Exec("CREATE DATABASE contful")
		if result.Error != nil {
			log.Fatalf("创建数据库失败: %v", result.Error)
		}
		log.Println("✅ 数据库 'contful' 创建成功")
	}

	// 连接 contful 数据库
	sqlDB, _ := db.DB()
	sqlDB.Close()

	dsn2 := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=contful sslmode=%s",
		host, port, user, pass, ssl)
	db2, err := gorm.Open(postgres.Open(dsn2), &gorm.Config{})
	if err != nil {
		log.Fatalf("连接 contful 失败: %v", err)
	}
	log.Println("✅ 连接 contful 数据库成功")

	// 检查表是否已存在
	var count int64
	db2.Raw("SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='public'").Scan(&count)
	if count > 0 {
		log.Printf("ℹ️  数据库已有 %d 张表，跳过初始化（已有数据）", count)
		os.Exit(0)
	}

	// 读取并执行 SQL 文件（相对于当前工作目录）
	sqlFile := getEnv("SQL_FILE", "sql/init.sql")

	// 如果是相对路径，转换为绝对路径
	if !filepath.IsAbs(sqlFile) {
		cwd, _ := os.Getwd()
		sqlFile = filepath.Join(cwd, sqlFile)
	}

	content, err := os.ReadFile(sqlFile)
	if err != nil {
		log.Fatalf("读取 SQL 文件失败 [%s]: %v", sqlFile, err)
	}

	err = db2.Exec(string(content)).Error
	if err != nil {
		log.Fatalf("执行 SQL 失败: %v", err)
	}
	log.Println("✅ SQL 初始化完成: " + sqlFile)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
