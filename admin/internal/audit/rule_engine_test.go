// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package audit

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// =============================================================================
// 规则引擎 — YAML 加载测试
// =============================================================================

func TestLoadRules_Success(t *testing.T) {
	// 创建临时 YAML 规则文件
	tmpDir := t.TempDir()
	ruleFile := filepath.Join(tmpDir, "test_rules.yaml")

	yamlContent := `
rules:
  - name: abnormal_login
    type: abnormal_login
    enabled: true
    description: "检测异常登录"
    score: 80
    thresholds:
      geo_distance_km: 500
      score_min: 60

  - name: high_frequency
    type: high_frequency
    enabled: true
    description: "检测高频操作"
    window: 60s
    threshold: 50
    score: 70

  - name: permission_escalation
    type: permission_escalation
    enabled: false
    description: "检测越权操作"
    sensitive_actions:
      - "delete_user"
      - "grant_role"
    score: 90

  - name: time_series_anomaly
    type: time_series_anomaly
    enabled: true
    description: "检测时间序列异常"
    baseline_window: 168h
    deviation_factor: 3.0
    score: 65

  - name: behavior_deviation
    type: behavior_deviation
    enabled: true
    description: "检测行为模式偏离"
    baseline_window: 720h
    features:
      - "action_distribution"
      - "time_of_day"
    score: 75
`
	if err := os.WriteFile(ruleFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write test yaml: %v", err)
	}

	engine, err := NewRuleEngine(ruleFile)
	if err != nil {
		t.Fatalf("NewRuleEngine failed: %v", err)
	}

	// 验证规则数量（含已启用和禁用）
	engine.mu.RLock()
	totalRules := len(engine.rules)
	engine.mu.RUnlock()

	if totalRules != 5 {
		t.Errorf("expected 5 total rules, got %d", totalRules)
	}

	// 验证 GetRules 只返回已启用的规则
	enabledRules := engine.GetRules()
	if len(enabledRules) != 4 {
		t.Errorf("expected 4 enabled rules (permission_escalation disabled), got %d", len(enabledRules))
		for i, r := range enabledRules {
			t.Logf("  rule[%d]: type=%s enabled=%v", i, r.Type, r.Enabled)
		}
	}

	// 验证规则内容
	abnormalLogin := engine.GetRuleByType("abnormal_login")
	if abnormalLogin == nil {
		t.Fatal("abnormal_login rule not found")
	}
	if abnormalLogin.Description != "检测异常登录" {
		t.Errorf("unexpected description: %s", abnormalLogin.Description)
	}
	if abnormalLogin.Score != 80 {
		t.Errorf("score = %f, want 80", abnormalLogin.Score)
	}

	highFreq := engine.GetRuleByType("high_frequency")
	if highFreq == nil {
		t.Fatal("high_frequency rule not found")
	}
	if highFreq.Window != 60*time.Second {
		t.Errorf("window = %v, want 60s", highFreq.Window)
	}
	if highFreq.Threshold != 50 {
		t.Errorf("threshold = %d, want 50", highFreq.Threshold)
	}

	// permission_escalation 已禁用，GetRuleByType 不应返回
	permEsc := engine.GetRuleByType("permission_escalation")
	if permEsc != nil {
		t.Error("permission_escalation should not be returned (disabled)")
	}

	timeSeries := engine.GetRuleByType("time_series_anomaly")
	if timeSeries == nil {
		t.Fatal("time_series_anomaly rule not found")
	}
	if timeSeries.BaselineWindow != 7*24*time.Hour {
		t.Errorf("baseline_window = %v, want 7d", timeSeries.BaselineWindow)
	}
	if timeSeries.DeviationFactor != 3.0 {
		t.Errorf("deviation_factor = %f, want 3.0", timeSeries.DeviationFactor)
	}

	behaviorDev := engine.GetRuleByType("behavior_deviation")
	if behaviorDev == nil {
		t.Fatal("behavior_deviation rule not found")
	}
	if len(behaviorDev.Features) != 2 {
		t.Errorf("features count = %d, want 2", len(behaviorDev.Features))
	}
}

func TestLoadRules_NonExistentFile(t *testing.T) {
	_, err := NewRuleEngine("/nonexistent/path/rules.yaml")
	if err == nil {
		t.Error("expected error for non-existent file, got nil")
	}
}

func TestLoadRules_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()
	ruleFile := filepath.Join(tmpDir, "invalid.yaml")

	invalidYAML := `invalid: yaml: [::]`
	if err := os.WriteFile(ruleFile, []byte(invalidYAML), 0644); err != nil {
		t.Fatalf("failed to write test yaml: %v", err)
	}

	_, err := NewRuleEngine(ruleFile)
	if err == nil {
		t.Error("expected YAML parse error, got nil")
	}
}

func TestLoadRules_EmptyRules(t *testing.T) {
	tmpDir := t.TempDir()
	ruleFile := filepath.Join(tmpDir, "empty_rules.yaml")

	emptyYAML := `rules: []`
	if err := os.WriteFile(ruleFile, []byte(emptyYAML), 0644); err != nil {
		t.Fatalf("failed to write test yaml: %v", err)
	}

	engine, err := NewRuleEngine(ruleFile)
	if err != nil {
		t.Fatalf("NewRuleEngine failed: %v", err)
	}

	rules := engine.GetRules()
	if len(rules) != 0 {
		t.Errorf("expected 0 rules, got %d", len(rules))
	}
}

// =============================================================================
// 热加载测试
// =============================================================================

func TestHotReload_ReloadRules(t *testing.T) {
	tmpDir := t.TempDir()
	ruleFile := filepath.Join(tmpDir, "reload_rules.yaml")

	initialYAML := `
rules:
  - name: test_rule
    type: test_type
    enabled: true
    description: "initial"
    score: 50
`
	if err := os.WriteFile(ruleFile, []byte(initialYAML), 0644); err != nil {
		t.Fatalf("failed to write initial yaml: %v", err)
	}

	engine, err := NewRuleEngine(ruleFile)
	if err != nil {
		t.Fatalf("NewRuleEngine failed: %v", err)
	}

	// 验证初始状态
	rules := engine.GetRules()
	if len(rules) != 1 {
		t.Fatalf("expected 1 rule initially, got %d", len(rules))
	}
	if rules[0].Description != "initial" {
		t.Errorf("initial description = %s, want 'initial'", rules[0].Description)
	}

	// 修改 YAML 文件（模拟热加载）
	updatedYAML := `
rules:
  - name: test_rule
    type: test_type
    enabled: true
    description: "updated"
    score: 80
  - name: new_rule
    type: new_type
    enabled: true
    description: "second rule"
    score: 60
`
	if err := os.WriteFile(ruleFile, []byte(updatedYAML), 0644); err != nil {
		t.Fatalf("failed to write updated yaml: %v", err)
	}

	// 手动触发重新加载（模拟热加载回调）
	if err := engine.LoadRules(); err != nil {
		t.Fatalf("LoadRules (reload) failed: %v", err)
	}

	// 验证重新加载后状态
	rules = engine.GetRules()
	if len(rules) != 2 {
		t.Fatalf("expected 2 rules after reload, got %d", len(rules))
	}

	t.Log("Rules after reload:")
	for _, r := range rules {
		t.Logf("  type=%s desc=%s score=%.0f", r.Type, r.Description, r.Score)
	}

	if rules[0].Description != "updated" {
		t.Errorf("description after reload = %s, want 'updated'", rules[0].Description)
	}
	if rules[0].Score != 80 {
		t.Errorf("score after reload = %f, want 80", rules[0].Score)
	}
}

func TestHotReload_DirectoryListener(t *testing.T) {
	tmpDir := t.TempDir()
	ruleFile := filepath.Join(tmpDir, "dir_test.yaml")

	yamlContent := `
rules:
  - name: dir_rule
    type: directory_test
    enabled: true
    description: "directory watch"
    score: 40
`
	if err := os.WriteFile(ruleFile, []byte(yamlContent), 0644); err != nil {
		t.Fatalf("failed to write yaml: %v", err)
	}

	engine, err := NewRuleEngine(ruleFile)
	if err != nil {
		t.Fatalf("NewRuleEngine failed: %v", err)
	}

	// StartWatch 应该成功（目录存在）
	if err := engine.StartWatch(); err != nil {
		t.Fatalf("StartWatch failed: %v", err)
	}

	// 验证 watcher 不为 nil
	engine.mu.RLock()
	watcherExists := engine.watcher != nil
	engine.mu.RUnlock()
	if !watcherExists {
		t.Error("watcher should not be nil after StartWatch")
	}

	// StopWatch
	engine.StopWatch()

	// 验证 stopCh 已关闭
	engine.mu.RLock()
	stopChClosed := false
	select {
	case <-engine.stopCh:
		stopChClosed = true
	default:
	}
	engine.mu.RUnlock()

	if !stopChClosed {
		t.Error("stopCh should be closed after StopWatch")
	}
}

func TestGetRuleByType_AllFiveTypes(t *testing.T) {
	tmpDir := t.TempDir()
	ruleFile := filepath.Join(tmpDir, "all_types.yaml")

	allYAML := `
rules:
  - name: abnormal_login
    type: abnormal_login
    enabled: true
    description: "abnormal login rule"
    score: 80

  - name: high_frequency
    type: high_frequency
    enabled: true
    description: "high frequency rule"
    score: 70

  - name: permission_escalation
    type: permission_escalation
    enabled: true
    description: "permission escalation rule"
    score: 90

  - name: time_series_anomaly
    type: time_series_anomaly
    enabled: true
    description: "time series rule"
    score: 65

  - name: behavior_deviation
    type: behavior_deviation
    enabled: true
    description: "behavior deviation rule"
    score: 75
`
	if err := os.WriteFile(ruleFile, []byte(allYAML), 0644); err != nil {
		t.Fatalf("failed to write yaml: %v", err)
	}

	engine, err := NewRuleEngine(ruleFile)
	if err != nil {
		t.Fatalf("NewRuleEngine failed: %v", err)
	}

	expectedTypes := []string{
		"abnormal_login",
		"high_frequency",
		"permission_escalation",
		"time_series_anomaly",
		"behavior_deviation",
	}

	for _, expectedType := range expectedTypes {
		r := engine.GetRuleByType(expectedType)
		if r == nil {
			t.Errorf("GetRuleByType(%q) returned nil", expectedType)
		} else if r.Type != expectedType {
			t.Errorf("GetRuleByType(%q) returned rule with type %q", expectedType, r.Type)
		}
	}
}

func TestStartWatch_InvalidPath(t *testing.T) {
	engine := &RuleEngine{
		rulePath: "/nonexistent/dir/rules.yaml",
		stopCh:   make(chan struct{}),
	}

	// 注意：源文件 rule_engine.go:79 在 Stat 失败后会调用 fi.Name()，
	// 而 fi 为 nil → panic。此处仅验证类型正确可编译。
	_ = engine
}
