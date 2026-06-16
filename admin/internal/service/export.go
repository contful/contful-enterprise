// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
)

// ExportService 导出服务
type ExportService struct {
	auditRepo *repository.AuditRepository
}

// NewExportService 创建导出服务
func NewExportService(auditRepo *repository.AuditRepository) *ExportService {
	return &ExportService{auditRepo: auditRepo}
}

// ExportJSONStream 流式 JSON (NDJSON) 导出
// 逐行写入 writer，每行一个 JSON 对象
func (s *ExportService) ExportJSONStream(ctx context.Context, filter *model.AuditLogFilter, maxRows int, writer io.Writer) (int64, int64, error) {
	logs, total, err := s.auditRepo.ExportAll(ctx, filter, maxRows)
	if err != nil {
		return 0, 0, fmt.Errorf("export query: %w", err)
	}

	encoder := json.NewEncoder(writer)
	encoder.SetEscapeHTML(false)

	count := int64(0)
	for _, al := range logs {
		// 字段选择
		record := buildExportRecord(&al, filter.Fields)
		if err := encoder.Encode(record); err != nil {
			return count, total, fmt.Errorf("encode record: %w", err)
		}
		count++
	}

	return count, total, nil
}

// buildExportRecord 根据字段选择构建导出记录
func buildExportRecord(log *model.AuditLog, fields []string) map[string]interface{} {
	allFields := map[string]interface{}{
		"id":              log.ID,
		"action":          log.Action,
		"category":        log.Category,
		"level":           log.Level,
		"resource_type":   log.ResourceType,
		"resource_id":     log.ResourceID,
		"user_id":         log.UserID,
		"site_id":         log.SiteID,
		"ip_address":      log.IPAddress,
		"user_agent":      log.UserAgent,
		"details":         log.Details,
		"created_time":    log.CreatedTime,
		"data_signature":  log.DataSignature,
		"request_body":    log.RequestBody,
		"response_status": log.ResponseStatus,
		"duration_ms":     log.DurationMs,
		"session_id":      log.SessionID,
		"geo_ip_info":     log.GeoIPInfo,
	}

	if len(fields) == 0 {
		return allFields
	}

	selected := make(map[string]interface{})
	for _, f := range fields {
		if v, ok := allFields[f]; ok {
			selected[f] = v
		}
	}

	// 确保至少有一个字段
	if len(selected) == 0 {
		return allFields
	}

	return selected
}
