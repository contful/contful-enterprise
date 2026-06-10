// +build dm
// go run -tags dm dm_test.go

package main

import (
	"fmt"
	"log"

	dm "gitee.com/chunanyong/dm"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func main() {
	dsn := "dm://SYSDBA:SYSDBA008@139.198.171.102:5236?schema=CONTFUL_ENT&compatibleMode=oracle"
	db, err := gorm.Open(dm.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("❌ 连接失败: %v", err)
	}
	var result struct {
		DmVersion string `gorm:"column:V$"`
	}
	db.Raw("SELECT BANNER AS DM_VERSION FROM V$VERSION").Scan(&result)
	fmt.Printf("✅ 达梦连接成功\n版本: %s\n", result.DmVersion)

	// 测试表
	db.Exec("SELECT 1 FROM DUAL")
	fmt.Println("✅ SELECT 1 FROM DUAL 成功")

	sqlDB, _ := db.DB()
	sqlDB.Close()
}
