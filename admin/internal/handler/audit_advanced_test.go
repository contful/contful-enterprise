// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/contful/contful/admin/internal/model"
)

// =============================================================================
// parseAuditFilter — 高级查询参数解析测试
// =============================================================================

func TestParseAuditFilter_Keyword(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/?keyword=%E5%88%A0%E9%99%A4&combinator=and", nil)

	filter := parseAuditFilter(c)

	if filter.Keyword != "删除" {
		t.Errorf("Keyword = %q, want %q", filter.Keyword, "删除")
	}
	if filter.Combinator != "and" {
		t.Errorf("Combinator = %q, want %q", filter.Combinator, "and")
	}
}

func TestParseAuditFilter_TimePresetValid(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		timePreset string
	}{
		{"1h preset", "1h"},
		{"24h preset", "24h"},
		{"7d preset", "7d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodGet, "/?time_preset="+tt.timePreset, nil)

			filter := parseAuditFilter(c)

			if filter.StartTime.IsZero() {
				t.Errorf("StartTime should not be zero for time_preset=%s", tt.timePreset)
			}
			if filter.TimePreset != tt.timePreset {
				t.Errorf("TimePreset = %s, want %s", filter.TimePreset, tt.timePreset)
			}
		})
	}
}

func TestParseAuditFilter_TimePresetCalculations(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		timePreset string
		marginBack time.Duration
	}{
		{"1h offset", "1h", 1*time.Hour + 10*time.Second},
		{"24h offset", "24h", 24*time.Hour + 10*time.Second},
		{"7d offset", "7d", 7*24*time.Hour + 10*time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodGet, "/?time_preset="+tt.timePreset, nil)

			filter := parseAuditFilter(c)

			expectedStart := time.Now().Add(-tt.marginBack)
			if filter.StartTime.Before(expectedStart) {
				t.Errorf("StartTime too early: %v, expected >= %v", filter.StartTime, expectedStart)
			}

			if !filter.EndTime.IsZero() {
				t.Error("EndTime should be zero when only time_preset is set")
			}
		})
	}
}

func TestParseAuditFilter_CombinatorValidation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		combinator string
		expected   string
	}{
		{"valid and", "and", "and"},
		{"valid or", "or", "or"},
		{"invalid xor", "xor", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			queryStr := "/"
			if tt.combinator != "" {
				queryStr = "/?combinator=" + tt.combinator
			}
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodGet, queryStr, nil)

			filter := parseAuditFilter(c)

			if filter.Combinator != tt.expected {
				t.Errorf("Combinator = %q, want %q", filter.Combinator, tt.expected)
			}
		})
	}
}

func TestParseAuditFilter_Fields(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		q        string
		expected []string
	}{
		{"single field", "fields=id", []string{"id"}},
		{"multiple fields", "fields=id,action,created_time", []string{"id", "action", "created_time"}},
		{"no fields", "", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = httptest.NewRequest(http.MethodGet, "/?"+tt.q, nil)

			filter := parseAuditFilter(c)

			if len(filter.Fields) != len(tt.expected) {
				t.Fatalf("Fields length = %d, want %d", len(filter.Fields), len(tt.expected))
			}
			for i, f := range filter.Fields {
				if f != tt.expected[i] {
					t.Errorf("Fields[%d] = %q, want %q", i, f, tt.expected[i])
				}
			}
		})
	}
}

func TestParseAuditFilter_LegacyParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet,
		"/?category=content&level=warn&action=delete_entry&start_time=2026-01-01T00:00:00Z&end_time=2026-06-01T00:00:00Z", nil)

	filter := parseAuditFilter(c)

	if filter.Category != model.AuditTypeContent {
		t.Errorf("Category = %s, want %s", filter.Category, model.AuditTypeContent)
	}
	if filter.Level != model.AuditLevelWarn {
		t.Errorf("Level = %s, want %s", filter.Level, model.AuditLevelWarn)
	}
	if filter.Action != "delete_entry" {
		t.Errorf("Action = %s, want delete_entry", filter.Action)
	}
}

// =============================================================================
// AnomalyFilter 解析测试
// =============================================================================

func TestParseAnomalyFilter_AllParams(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet,
		"/?anomaly_type=abnormal_login&severity=high&start_time=2026-06-01T00:00:00Z&end_time=2026-06-16T00:00:00Z", nil)

	filter := parseAnomalyFilter(c)

	if filter.AnomalyType != model.AnomalyAbnormalLogin {
		t.Errorf("AnomalyType = %s, want %s", filter.AnomalyType, model.AnomalyAbnormalLogin)
	}
	if filter.Severity != model.SeverityHigh {
		t.Errorf("Severity = %s, want %s", filter.Severity, model.SeverityHigh)
	}
	if filter.StartTime.IsZero() {
		t.Error("StartTime should not be zero")
	}
	if filter.EndTime.IsZero() {
		t.Error("EndTime should not be zero")
	}
}

func TestParseAnomalyFilter_Empty(t *testing.T) {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	filter := parseAnomalyFilter(c)

	if filter.AnomalyType != "" {
		t.Errorf("AnomalyType should be empty, got %s", filter.AnomalyType)
	}
	if filter.Severity != "" {
		t.Errorf("Severity should be empty, got %s", filter.Severity)
	}
}

// =============================================================================
// Model JSON tag 一致性验证
// =============================================================================

func TestAuditLogJSONTags(t *testing.T) {
	sid := "sess_123"
	dur := 45
	var rs int16 = 200
	geoJSON := json.RawMessage([]byte(`{"country":"CN"}`))
	log := model.AuditLog{
		SessionID:      &sid,
		DurationMs:     &dur,
		ResponseStatus: &rs,
		RequestBody:    strPtr(`{"key":"value"}`),
		GeoIPInfo:      geoJSON,
	}

	data, err := json.Marshal(log)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(data, &result)

	// 验证新增字段使用 snake_case JSON tag（omitempty 时 nil 指针被省略）
	checks := []string{"session_id", "duration_ms", "response_status", "geo_ip_info", "request_body"}
	for _, key := range checks {
		if _, ok := result[key]; !ok {
			t.Errorf("missing %s in JSON output (set non-nil value)", key)
		}
	}
}

func TestAuditAnomalyJSONTags(t *testing.T) {
	now := time.Now()
	a := model.AuditAnomaly{
		AnomalyType:  model.AnomalyAbnormalLogin,
		Severity:     model.SeverityHigh,
		Score:        85.5,
		Description:  "test",
		DetectedTime: now,
	}

	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}

	var result map[string]interface{}
	json.Unmarshal(data, &result)

	checks := []string{"anomaly_type", "severity", "score", "description", "detected_time"}
	for _, key := range checks {
		if _, ok := result[key]; !ok {
			t.Errorf("missing %s in anomaly JSON output", key)
		}
	}
}

// =============================================================================
// TableName 一致性验证
// =============================================================================

func TestTableNames(t *testing.T) {
	var a model.AuditAnomaly
	if tn := a.TableName(); tn != "contful_audit_anomalies" {
		t.Errorf("AuditAnomaly.TableName = %s, want contful_audit_anomalies", tn)
	}
	var qt model.AuditQueryTemplate
	if tn := qt.TableName(); tn != "contful_audit_query_templates" {
		t.Errorf("AuditQueryTemplate.TableName = %s, want contful_audit_query_templates", tn)
	}
	var al model.AuditLog
	if tn := al.TableName(); tn != "contful_audit_logs" {
		t.Errorf("AuditLog.TableName = %s, want contful_audit_logs", tn)
	}
}

// =============================================================================
// helpers
// =============================================================================

func strPtr(s string) *string { return &s }
func intPtr(i int) *int       { return &i }
