// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package audit

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful-enterprise/shared/uid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Detector 异常检测器
type Detector struct {
	db *gorm.DB
}

// NewDetector 创建检测器
func NewDetector(db *gorm.DB) *Detector {
	return &Detector{db: db}
}

// Detect 根据规则检测异常
func (d *Detector) Detect(ctx context.Context, rule Rule, logs []model.AuditLog) []model.AuditAnomaly {
	switch rule.Type {
	case "abnormal_login":
		return d.detectAbnormalLogin(ctx, rule, logs)
	case "high_frequency":
		return d.detectHighFrequency(ctx, rule, logs)
	case "permission_escalation":
		return d.detectPermissionEscalation(ctx, rule, logs)
	case "time_series_anomaly":
		return d.detectTimeSeriesAnomaly(ctx, rule, logs)
	case "behavior_deviation":
		return d.detectBehaviorDeviation(ctx, rule, logs)
	default:
		log.Warn().
			Str("module", "detector").
			Str("rule_type", rule.Type).
			Msg("unknown rule type, skipping")
		return nil
	}
}

// detectAbnormalLogin 检测异常登录（地理位置偏差 + 时间偏差）
func (d *Detector) detectAbnormalLogin(ctx context.Context, rule Rule, logs []model.AuditLog) []model.AuditAnomaly {
	scoreMin := 60.0
	if v, ok := rule.Thresholds["score_min"]; ok {
		if f, ok := v.(float64); ok {
			scoreMin = f
		}
	}

	var anomalies []model.AuditAnomaly
	now := time.Now()

	for _, log := range logs {
		if log.Category != model.AuditTypeAuth || log.Action != "login" {
			continue
		}
		if log.UserID == nil || log.IPAddress == "" {
			continue
		}

		// 查询用户最近 30 天的登录地点
		var recentIPs []string
		d.db.WithContext(ctx).Model(&model.AuditLog{}).
			Select("DISTINCT ip_address").
			Where("user_id = ? AND category = ? AND action = ? AND created_time > ? AND ip_address != ?",
				*log.UserID, "auth", "login", now.Add(-30*24*time.Hour), log.IPAddress).
			Pluck("ip_address", &recentIPs)

		// 如果 IP 不在最近 30 天的记录中，标记异常
		if len(recentIPs) > 0 && !containsString(recentIPs, log.IPAddress) {
			baseline := map[string]interface{}{"known_ips": recentIPs}
			baselineJSON, _ := json.Marshal(baseline)
			actual := map[string]interface{}{"ip": log.IPAddress}
			actualJSON, _ := json.Marshal(actual)

			score := rule.Score
			if score < scoreMin {
				score = scoreMin + 10
			}

			anomalies = append(anomalies, model.AuditAnomaly{
				AuditLogID:    &log.ID,
				AnomalyType:   model.AnomalyAbnormalLogin,
				Severity:      determineSeverity(score),
				Score:         score,
				BaselineValue: baselineJSON,
				ActualValue:   actualJSON,
				Description:   fmt.Sprintf("用户从新的 IP 地址登录: %s", log.IPAddress),
				DetectedTime:  now,
			})
		}
	}

	return anomalies
}

// detectHighFrequency 检测高频操作
func (d *Detector) detectHighFrequency(ctx context.Context, rule Rule, logs []model.AuditLog) []model.AuditAnomaly {
	threshold := rule.Threshold
	window := rule.Window
	if window == 0 {
		window = 60 * time.Second
	}
	if threshold == 0 {
		threshold = 50
	}

	var anomalies []model.AuditAnomaly
	now := time.Now()

	// 按用户统计最近窗口内的操作次数
	userCounts := make(map[uid.UID]int64)
	windowStart := now.Add(-window)

	for _, log := range logs {
		if log.CreatedTime.Before(windowStart) {
			continue
		}
		if log.UserID != nil {
			userCounts[*log.UserID]++
		}
	}

	for userID, count := range userCounts {
		if count >= int64(threshold) {
			baseline := map[string]interface{}{"threshold": threshold}
			baselineJSON, _ := json.Marshal(baseline)
			actual := map[string]interface{}{"count": count, "window": window.String(), "user_id": userID.String()}
			actualJSON, _ := json.Marshal(actual)

			anomalies = append(anomalies, model.AuditAnomaly{
				AnomalyType:   model.AnomalyHighFrequency,
				Severity:      determineSeverity(rule.Score),
				Score:         rule.Score,
				BaselineValue: baselineJSON,
				ActualValue:   actualJSON,
				Description:   fmt.Sprintf("用户在 %s 内执行了 %d 次操作，超过阈值 %d", window, count, threshold),
				DetectedTime:  now,
			})
		}
	}

	return anomalies
}

// detectPermissionEscalation 检测越权操作
func (d *Detector) detectPermissionEscalation(ctx context.Context, rule Rule, logs []model.AuditLog) []model.AuditAnomaly {
	var anomalies []model.AuditAnomaly
	now := time.Now()

	for _, log := range logs {
		if !containsString(rule.SensitiveActions, log.Action) {
			continue
		}

		sensitiveActionsJSON, _ := json.Marshal(rule.SensitiveActions)

		anomalies = append(anomalies, model.AuditAnomaly{
			AuditLogID:    &log.ID,
			AnomalyType:   model.AnomalyPermissionEscalation,
			Severity:      determineSeverity(rule.Score),
			Score:         rule.Score,
			BaselineValue: nil,
			ActualValue:   sensitiveActionsJSON,
			Description:   fmt.Sprintf("检测到敏感操作: %s（用户: %s）", log.Action, userIDStr(log.UserID)),
			DetectedTime:  now,
		})
	}

	return anomalies
}

// detectTimeSeriesAnomaly 检测时间序列异常（操作次数突增）
func (d *Detector) detectTimeSeriesAnomaly(ctx context.Context, rule Rule, logs []model.AuditLog) []model.AuditAnomaly {
	baselineWindow := rule.BaselineWindow
	if baselineWindow == 0 {
		baselineWindow = 7 * 24 * time.Hour
	}
	deviationFactor := rule.DeviationFactor
	if deviationFactor == 0 {
		deviationFactor = 3.0
	}

	var anomalies []model.AuditAnomaly
	now := time.Now()

	// 统计最近 1 小时操作数
	oneHourAgo := now.Add(-1 * time.Hour)
	var recentCount int64
	for _, log := range logs {
		if log.CreatedTime.After(oneHourAgo) {
			recentCount++
		}
	}

	// 查询基线（过去 baseline_window 窗口内每小时平均操作数）
	var totalInBaseline int64
	d.db.WithContext(ctx).Model(&model.AuditLog{}).
		Where("created_time > ? AND created_time <= ?", now.Add(-baselineWindow), now).
		Count(&totalInBaseline)

	baselineHours := baselineWindow.Hours()
	avgPerHour := float64(totalInBaseline) / baselineHours
	if avgPerHour < 1 {
		avgPerHour = 1 // 避免除零
	}

	if float64(recentCount) > avgPerHour*deviationFactor {
		baseline := map[string]interface{}{"avg_per_hour": avgPerHour, "baseline_window": baselineWindow.String()}
		baselineJSON, _ := json.Marshal(baseline)
		actual := map[string]interface{}{"recent_count": recentCount, "deviation_factor": deviationFactor}
		actualJSON, _ := json.Marshal(actual)

		anomalies = append(anomalies, model.AuditAnomaly{
			AnomalyType:   model.AnomalyTimeSeries,
			Severity:      determineSeverity(rule.Score),
			Score:         rule.Score,
			BaselineValue: baselineJSON,
			ActualValue:   actualJSON,
			Description:   fmt.Sprintf("最近 1 小时操作数 %d，超过基线平均值 %.1f 的 %.1f 倍", recentCount, avgPerHour, deviationFactor),
			DetectedTime:  now,
		})
	}

	return anomalies
}

// detectBehaviorDeviation 检测行为模式偏离
func (d *Detector) detectBehaviorDeviation(ctx context.Context, rule Rule, logs []model.AuditLog) []model.AuditAnomaly {
	baselineWindow := rule.BaselineWindow
	if baselineWindow == 0 {
		baselineWindow = 30 * 24 * time.Hour
	}

	var anomalies []model.AuditAnomaly
	now := time.Now()

	// 按用户统计行为分布
	userActions := make(map[uid.UID]map[string]int64)

	for _, log := range logs {
		if log.UserID == nil {
			continue
		}
		if userActions[*log.UserID] == nil {
			userActions[*log.UserID] = make(map[string]int64)
		}
		userActions[*log.UserID][log.Action]++
	}

	// 查询基线（过去 baseline_window 的行为分布）
	for userID := range userActions {
		var baselineLogs []model.AuditLog
		d.db.WithContext(ctx).Model(&model.AuditLog{}).
			Where("user_id = ? AND created_time > ? AND created_time <= ?",
				userID, now.Add(-baselineWindow), now.Add(-5*time.Minute)).
			Find(&baselineLogs)

		baselineActions := make(map[string]int64)
		for _, bl := range baselineLogs {
			baselineActions[bl.Action]++
		}

		// 如果用户有历史行为，检查新行为
		if len(baselineActions) > 0 {
			for action, count := range userActions[userID] {
				if _, exists := baselineActions[action]; !exists {
					// 新出现的操作类型
					baseline := map[string]interface{}{"known_actions": keysOf(baselineActions)}
					baselineJSON, _ := json.Marshal(baseline)
					actual := map[string]interface{}{"new_action": action, "count": count}
					actualJSON, _ := json.Marshal(actual)

					anomalies = append(anomalies, model.AuditAnomaly{
						AnomalyType:   model.AnomalyBehaviorDeviation,
						Severity:      determineSeverity(rule.Score),
						Score:         rule.Score,
						BaselineValue: baselineJSON,
						ActualValue:   actualJSON,
						Description:   fmt.Sprintf("用户执行了罕见操作: %s（历史中未出现）", action),
						DetectedTime:  now,
						CreatedTime:   now,
					})
				}
			}
		}
	}

	return anomalies
}

// determineSeverity 根据分数确定严重程度
func determineSeverity(score float64) model.AnomalySeverity {
	switch {
	case score >= 90:
		return model.SeverityCritical
	case score >= 70:
		return model.SeverityHigh
	case score >= 50:
		return model.SeverityMedium
	default:
		return model.SeverityLow
	}
}

// containsString 检查字符串切片中是否包含指定字符串
func containsString(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

// keysOf 返回 map 的键列表
func keysOf(m map[string]int64) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

func userIDStr(id *uid.UID) string {
	if id == nil {
		return "unknown"
	}
	return id.String()
}
