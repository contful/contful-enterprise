// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/contful/contful/admin/internal/model"
)

// SetupGuard 安装向导路由守卫：已安装后隐藏 setup 端点（返回 404）
func SetupGuard(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 检查 contful_system_users 表是否存在且含数据
		if !db.Migrator().HasTable("contful_system_users") {
			// 表不存在 → 尚未安装，放行
			c.Next()
			return
		}

		var count int64
		db.Table("contful_system_users").Count(&count)
		if count > 0 {
			// 已安装，返回 404（不暴露 setup 端点是否存在）
			c.AbortWithStatusJSON(http.StatusNotFound,
				model.NewErrorResponse(model.CodeNotFound, "not found"))
			return
		}

		c.Next()
	}
}
