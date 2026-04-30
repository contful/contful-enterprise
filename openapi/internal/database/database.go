// Package database 提供数据库驱动的编译时选择支持。
// 通过 build tag 选择编译目标数据库：
//
//   - //go:build pg  — PostgreSQL 驱动
//   - //go:build dm  — 达梦 DM8 驱动
//
// 使用方式：
//
//	go build -tags=pg   # 编译 PostgreSQL 版本
//	go build -tags=dm   # 编译达梦 DM8 版本
//
// init.go 负责在编译时注册对应的 GORM Dialect。
package database
