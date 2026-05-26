// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package metrics

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// Handler renders /metrics in Prometheus format.
type Handler struct {
	db *gorm.DB
}

// NewHandler creates a metrics handler with optional DB for pool stats.
func NewHandler(db *gorm.DB) *Handler {
	return &Handler{db: db}
}

// Metrics handles GET /metrics — outputs Prometheus exposition format.
func (h *Handler) Metrics(c *gin.Context) {
	contentType := c.GetHeader("Accept")
	if strings.Contains(contentType, "text/plain") || contentType == "" {
		c.Header("Content-Type", "text/plain; version=0.0.4")
	}

	var b strings.Builder
	b.WriteString(global.PrometheusText())

	// DB pool stats (if available)
	if h.db != nil {
		if sqlDB, err := h.db.DB(); err == nil {
			stats := sqlDB.Stats()
			b.WriteString(fmt.Sprintf("# HELP contful_db_pool_open_connections Open database connections.\n"))
			b.WriteString(fmt.Sprintf("# TYPE contful_db_pool_open_connections gauge\n"))
			b.WriteString(fmt.Sprintf("contful_db_pool_open_connections %d\n", stats.OpenConnections))
			b.WriteString(fmt.Sprintf("# HELP contful_db_pool_in_use In-use database connections.\n"))
			b.WriteString(fmt.Sprintf("# TYPE contful_db_pool_in_use gauge\n"))
			b.WriteString(fmt.Sprintf("contful_db_pool_in_use %d\n", stats.InUse))
			b.WriteString(fmt.Sprintf("# HELP contful_db_pool_idle Idle database connections.\n"))
			b.WriteString(fmt.Sprintf("# TYPE contful_db_pool_idle gauge\n"))
			b.WriteString(fmt.Sprintf("contful_db_pool_idle %d\n", stats.Idle))
			b.WriteString(fmt.Sprintf("# HELP contful_db_pool_wait_count Connections waiting for pool.\n"))
			b.WriteString(fmt.Sprintf("# TYPE contful_db_pool_wait_count counter\n"))
			b.WriteString(fmt.Sprintf("contful_db_pool_wait_count %d\n", stats.WaitCount))
			b.WriteString(fmt.Sprintf("# HELP contful_db_pool_max_idle_closed Closed idle connections.\n"))
			b.WriteString(fmt.Sprintf("# TYPE contful_db_pool_max_idle_closed counter\n"))
			b.WriteString(fmt.Sprintf("contful_db_pool_max_idle_closed %d\n", stats.MaxIdleClosed))
		}
	}

	c.String(http.StatusOK, b.String())
}
