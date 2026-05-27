// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

// SetupStatusResponse 安装状态响应
type SetupStatusResponse struct {
	SetupRequired bool   `json:"setup_required"`
	Version       string `json:"version"`
}

// SetupDatabaseRequest 数据库连接测试请求
type SetupDatabaseRequest struct {
	Host     string `json:"host" binding:"required"`
	Port     int    `json:"port" binding:"required"`
	User     string `json:"user" binding:"required"`
	Password string `json:"password" binding:"required"`
	DBName   string `json:"db_name" binding:"required"`
	SSLMode  string `json:"ssl_mode"`
}

// SetupDatabaseResponse 数据库连接测试响应
type SetupDatabaseResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

// SetupAdminRequest 管理员创建请求
type SetupAdminRequest struct {
	Email           string `json:"email" binding:"required,email"`
	Password        string `json:"password" binding:"required,min=8"`
	PasswordConfirm string `json:"password_confirm" binding:"required,eqfield=Password"`
	SiteName        string `json:"site_name" binding:"required"`
	SiteSlug        string `json:"site_slug" binding:"required"`
}

// SetupCompleteResponse 安装完成响应
type SetupCompleteResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}
