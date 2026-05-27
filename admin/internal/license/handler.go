// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0

package license

import (
	"net/http"

	"github.com/contful/contful/admin/internal/model"
	"github.com/gin-gonic/gin"
)

// Handler license 信息接口处理器
type Handler struct {
	info *Info
}

// NewHandler 创建 license handler
func NewHandler(info *Info) *Handler {
	return &Handler{info: info}
}

// GetInfo 返回当前 license 信息
// GET /admin/api/v1/system/license
func (h *Handler) GetInfo(c *gin.Context) {
	if h.info == nil {
		c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{
			"status":  "unlicensed",
			"message": "未找到有效的授权文件 (conf/license.dat)",
		}))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{
		"status":          h.info.Status(),
		"customer":        h.info.Customer,
		"product_name":    h.info.ProductName,
		"product_version": h.info.ProductVersion,
		"product_code":    h.info.ProductCode,
		"is_trial":        h.info.IsTrial,
		"issued_date":     h.info.IssuedDate,
		"expiry_date":     h.info.ExpiryDate,
		"is_expired":      h.info.IsExpired(),
	}))
}
