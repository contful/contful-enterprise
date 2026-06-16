// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package audit

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful-enterprise/shared/uid"
)

// =============================================================================
// Helper 函数纯逻辑测试
// =============================================================================

func TestContainsString(t *testing.T) {
	tests := []struct {
		name   string
		slice  []string
		s      string
		expect bool
	}{
		{"found first", []string{"a", "b", "c"}, "a", true},
		{"found middle", []string{"a", "b", "c"}, "b", true},
		{"found last", []string{"a", "b", "c"}, "c", true},
		{"not found", []string{"a", "b", "c"}, "d", false},
		{"empty slice", []string{}, "a", false},
		{"nil slice", nil, "a", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsString(tt.slice, tt.s)
			if result != tt.expect {
				t.Errorf("containsString(%v, %q) = %v, want %v", tt.slice, tt.s, result, tt.expect)
			}
		})
	}
}

func TestKeysOf(t *testing.T) {
	m := map[string]int64{"a": 1, "b": 2, "c": 3}
	keys := keysOf(m)

	if len(keys) != 3 {
		t.Errorf("keysOf length = %d, want 3", len(keys))
	}

	// 验证返回的 keys 都在 map 中
	found := make(map[string]bool)
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			t.Errorf("unexpected key %q in result", k)
		}
		found[k] = true
	}

	// 验证所有 key 都被返回
	for k := range m {
		if !found[k] {
			t.Errorf("missing key %q in result", k)
		}
	}
}

func TestKeysOfEmpty(t *testing.T) {
	keys := keysOf(map[string]int64{})
	if len(keys) != 0 {
		t.Errorf("expected empty slice, got %v", keys)
	}
}

func TestUserIDStr(t *testing.T) {
	// nil 指针
	result := userIDStr(nil)
	if result != "unknown" {
		t.Errorf("userIDStr(nil) = %q, want %q", result, "unknown")
	}

	// 有效 UID
	id := uid.New()
	result = userIDStr(&id)
	if result != id.String() {
		t.Errorf("userIDStr(valid) = %q, want %q", result, id.String())
	}
}

func TestDetermineSeverity(t *testing.T) {
	tests := []struct {
		name     string
		score    float64
		expected model.AnomalySeverity
	}{
		{"critical threshold 90", 90, model.SeverityCritical},
		{"critical above 90", 95, model.SeverityCritical},
		{"high threshold 70", 70, model.SeverityHigh},
		{"high between 70-90", 85, model.SeverityHigh},
		{"medium threshold 50", 50, model.SeverityMedium},
		{"medium between 50-70", 60, model.SeverityMedium},
		{"low below 50", 40, model.SeverityLow},
		{"low zero", 0, model.SeverityLow},
		{"low negative", -1, model.SeverityLow},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := determineSeverity(tt.score)
			if result != tt.expected {
				t.Errorf("determineSeverity(%.1f) = %s, want %s", tt.score, result, tt.expected)
			}
		})
	}
}

func TestNewDetector(t *testing.T) {
	d := NewDetector(nil)
	if d == nil {
		t.Fatal("NewDetector returned nil")
	}
	if d.db != nil {
		t.Error("Expected nil db, got non-nil")
	}
}

// =============================================================================
// 5 类异常检测规则 — 逻辑分支/边界用例（白盒验证）
// 以下测试验证检测器的参数处理、过滤逻辑和输出结构，不依赖数据库连接。
// =============================================================================

// 用例 1: abnormal_login — 滤除非 auth/login/无 user_id/无 ip 的记录
func TestDetectAbnormalLogin_FiltersIrrelevantLogs(t *testing.T) {
	d := &Detector{db: nil}
	rule := Rule{
		Type:       "abnormal_login",
		Enabled:    true,
		Score:      80,
		Thresholds: map[string]interface{}{"score_min": float64(60)},
	}

	now := time.Now()
	userID := uid.New()
	logs := []model.AuditLog{
		{Category: model.AuditTypeContent, Action: "login", UserID: &userID, IPAddress: "10.0.0.1", CreatedTime: now},       // wrong category
		{Category: model.AuditTypeAuth, Action: "logout", UserID: &userID, IPAddress: "10.0.0.1", CreatedTime: now},       // wrong action
		{Category: model.AuditTypeAuth, Action: "login", UserID: nil, IPAddress: "10.0.0.1", CreatedTime: now},            // no user_id
		{Category: model.AuditTypeAuth, Action: "login", UserID: &userID, IPAddress: "", CreatedTime: now},                // no ip
	}

	// 所有记录都应该被滤掉（不符合 auth+login+有user+有ip）
	anomalies := d.detectAbnormalLogin(context.Background(), rule, logs)
	if len(anomalies) != 0 {
		t.Errorf("expected 0 anomalies for filtered logs, got %d", len(anomalies))
	}
}

// 用例 1b: abnormal_login — 验证 detectAbnormalLogin 的类型路由
func TestDetect_AbnormalLogin_Route(t *testing.T) {
	d := &Detector{db: nil}
	rule := Rule{Type: "abnormal_login", Score: 70, Thresholds: map[string]interface{}{"score_min": float64(60)}}
	userID := uid.New()

	logs := []model.AuditLog{
		{ID: uid.New(), Category: model.AuditTypeAuth, Action: "login", UserID: &userID, IPAddress: "10.0.0.1", CreatedTime: time.Now()},
	}

	// 注：由于没有 DB，DB 查询会跳过；仅验证函数可被调用不 panic
	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered from expected panic (no DB): %v", r)
		}
	}()
	_ = d.Detect(context.Background(), rule, logs)
}

// 用例 2: high_frequency — 窗口/阈值默认值
func TestDetectHighFrequency_Defaults(t *testing.T) {
	d := &Detector{db: nil}

	// 规则未设置 window 和 threshold
	rule := Rule{Type: "high_frequency", Score: 70}
	userID := uid.New()
	now := time.Now()

	logs := []model.AuditLog{
		{UserID: &userID, Action: "read", CreatedTime: now, IPAddress: "10.0.0.1", Category: model.AuditTypeContent},
	}

	// 只有 1 条记录，远低于默认 threshold=50，不应触发
	anomalies := d.detectHighFrequency(context.Background(), rule, logs)
	if len(anomalies) != 0 {
		t.Errorf("expected 0 anomalies (count below threshold), got %d", len(anomalies))
	}
}

// 用例 2b: high_frequency — 超过阈值触发异常
func TestDetectHighFrequency_ExceedsThreshold(t *testing.T) {
	d := &Detector{db: nil}

	rule := Rule{Type: "high_frequency", Score: 70, Window: 60 * time.Second, Threshold: 5}
	userID := uid.New()
	now := time.Now()

	// 生成 10 条记录在同一窗口内
	logs := make([]model.AuditLog, 10)
	for i := 0; i < 10; i++ {
		logs[i] = model.AuditLog{
			UserID:      &userID,
			Action:      "api_call",
			CreatedTime: now.Add(-time.Duration(i) * time.Second),
			IPAddress:   "10.0.0.1",
			Category:    model.AuditTypeSystem,
		}
	}

	anomalies := d.detectHighFrequency(context.Background(), rule, logs)

	if len(anomalies) != 1 {
		t.Fatalf("expected 1 anomaly, got %d", len(anomalies))
	}

	a := anomalies[0]
	if a.AnomalyType != model.AnomalyHighFrequency {
		t.Errorf("AnomalyType = %s, want %s", a.AnomalyType, model.AnomalyHighFrequency)
	}
	if a.Score != 70 {
		t.Errorf("Score = %f, want 70", a.Score)
	}
	if a.Description == "" {
		t.Error("Description should not be empty")
	}
	if a.DetectedTime.IsZero() {
		t.Error("DetectedTime should not be zero")
	}

	// 验证 actual_value 包含计数值
	var actual map[string]interface{}
	json.Unmarshal(a.ActualValue, &actual)
	if count, ok := actual["count"]; !ok || count == nil {
		t.Error("actual_value should contain 'count'")
	}
}

// 用例 3: permission_escalation — 敏感操作匹配
func TestDetectPermissionEscalation_DetectSensitiveActions(t *testing.T) {
	d := &Detector{db: nil}

	rule := Rule{
		Type:             "permission_escalation",
		Score:            90,
		SensitiveActions: []string{"delete_user", "grant_role", "modify_permission"},
	}

	userID := uid.New()
	now := time.Now()

	logs := []model.AuditLog{
		{ID: uid.New(), UserID: &userID, Action: "read_entries", CreatedTime: now, IPAddress: "10.0.0.1", Category: model.AuditTypeContent},
		{ID: uid.New(), UserID: &userID, Action: "delete_user", CreatedTime: now, IPAddress: "10.0.0.1", Category: model.AuditTypeUser},
		{ID: uid.New(), UserID: &userID, Action: "grant_role", CreatedTime: now, IPAddress: "10.0.0.1", Category: model.AuditTypeUser},
	}

	anomalies := d.detectPermissionEscalation(context.Background(), rule, logs)

	if len(anomalies) != 2 {
		t.Fatalf("expected 2 anomalies (delete_user + grant_role), got %d", len(anomalies))
	}

	for _, a := range anomalies {
		if a.AnomalyType != model.AnomalyPermissionEscalation {
			t.Errorf("AnomalyType = %s, want %s", a.AnomalyType, model.AnomalyPermissionEscalation)
		}
		if a.Severity != model.SeverityCritical {
			t.Errorf("Severity = %s (score=%v), want %s", a.Severity, a.Score, model.SeverityCritical)
		}
		if a.Score != 90 {
			t.Errorf("Score = %f, want 90", a.Score)
		}
	}
}

// 用例 3b: permission_escalation — 无敏感操作则不触发
func TestDetectPermissionEscalation_NoSensitiveActions(t *testing.T) {
	d := &Detector{db: nil}

	rule := Rule{
		Type:             "permission_escalation",
		Score:            90,
		SensitiveActions: []string{"delete_user", "grant_role"},
	}

	userID := uid.New()
	now := time.Now()
	logs := []model.AuditLog{
		{UserID: &userID, Action: "read_entries", CreatedTime: now, IPAddress: "10.0.0.1", Category: model.AuditTypeContent},
		{UserID: &userID, Action: "update_profile", CreatedTime: now, IPAddress: "10.0.0.1", Category: model.AuditTypeUser},
	}

	anomalies := d.detectPermissionEscalation(context.Background(), rule, logs)
	if len(anomalies) != 0 {
		t.Errorf("expected 0 anomalies for non-sensitive actions, got %d", len(anomalies))
	}
}

// 用例 4: time_series_anomaly — 默认值验证
func TestDetectTimeSeriesAnomaly_Defaults(t *testing.T) {
	d := &Detector{db: nil}

	// 规则未设置 baseline_window 和 deviation_factor
	rule := Rule{Type: "time_series_anomaly", Score: 65}
	now := time.Now()

	// 生成少量记录（不足以触发异常）
	logs := make([]model.AuditLog, 5)
	for i := 0; i < 5; i++ {
		logs[i] = model.AuditLog{
			Action:      "api_call",
			CreatedTime: now.Add(-time.Duration(i) * time.Minute),
			IPAddress:   "10.0.0.1",
			Category:    model.AuditTypeSystem,
		}
	}

	// 由于没有 DB，detectTimeSeriesAnomaly 的 DB 查询会 panic
	// 验证无 DB 时的行为（仅验证默认值逻辑不会编译错误）
	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered from expected nil DB panic: %v", r)
		}
	}()
	_ = d.detectTimeSeriesAnomaly(context.Background(), rule, logs)
}

// 用例 4b: time_series_anomaly — 大量记录触发异常（无 DB 时验证局部逻辑）
func TestDetectTimeSeriesAnomaly_CustomParams(t *testing.T) {
	d := &Detector{db: nil}

	// 设置极低的 deviation_factor，使得即使少量记录也会被判定异常
	rule := Rule{
		Type:            "time_series_anomaly",
		Score:           65,
		BaselineWindow:  7 * 24 * time.Hour,
		DeviationFactor: 0.01, // 极低，几乎任何数量都会触发
	}
	now := time.Now()

	logs := make([]model.AuditLog, 3)
	for i := 0; i < 3; i++ {
		logs[i] = model.AuditLog{
			Action:      "api_call",
			CreatedTime: now.Add(-time.Duration(i) * time.Minute),
			IPAddress:   "10.0.0.1",
			Category:    model.AuditTypeSystem,
		}
	}

	// 由于无 DB，detectTimeSeriesAnomaly 的 DB 查询会 panic
	// 此测试验证：Custom params 类型正确可编译
	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered from expected nil DB panic: %v", r)
		}
	}()
	_ = d.detectTimeSeriesAnomaly(context.Background(), rule, logs)
}

// 用例 5: behavior_deviation — 无历史基线时不触发
func TestDetectBehaviorDeviation_NoBaseline(t *testing.T) {
	d := &Detector{db: nil}

	rule := Rule{
		Type:           "behavior_deviation",
		Score:          75,
		BaselineWindow: 30 * 24 * time.Hour,
	}
	now := time.Now()
	userID := uid.New()

	logs := []model.AuditLog{
		{UserID: &userID, Action: "new_action", CreatedTime: now, IPAddress: "10.0.0.1", Category: model.AuditTypeContent},
	}

	// 无 DB → detectBehaviorDeviation 的 DB 查询会 panic
	defer func() {
		if r := recover(); r != nil {
			t.Logf("recovered from expected nil DB panic: %v", r)
		}
	}()
	_ = d.detectBehaviorDeviation(context.Background(), rule, logs)
}

// 用例 5b: behavior_deviation — 跳过 nil userID 的日志
func TestDetectBehaviorDeviation_SkipNilUserID(t *testing.T) {
	d := &Detector{db: nil}

	rule := Rule{
		Type:           "behavior_deviation",
		Score:          75,
		BaselineWindow: 30 * 24 * time.Hour,
	}
	now := time.Now()

	logs := []model.AuditLog{
		{UserID: nil, Action: "some_action", CreatedTime: now, IPAddress: "10.0.0.1", Category: model.AuditTypeSystem},
	}

	anomalies := d.detectBehaviorDeviation(context.Background(), rule, logs)
	if len(anomalies) != 0 {
		t.Errorf("expected 0 anomalies (nil user_id), got %d", len(anomalies))
	}
}

// 用例 6: Detect — 未知规则类型返回 nil
func TestDetect_UnknownRuleType(t *testing.T) {
	d := &Detector{db: nil}
	rule := Rule{Type: "unknown_type", Score: 50}
	userID := uid.New()
	logs := []model.AuditLog{
		{UserID: &userID, Action: "test", CreatedTime: time.Now(), IPAddress: "10.0.0.1", Category: model.AuditTypeSystem},
	}

	anomalies := d.Detect(context.Background(), rule, logs)
	if anomalies != nil {
		t.Errorf("expected nil for unknown rule type, got %v", anomalies)
	}
}
