// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"fmt"
	"time"

	"github.com/contful/contful-enterprise/shared/uid"
	"github.com/contful/contful/admin/internal/audit"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/rs/zerolog/log"
)

// AnomalyService 异常检测服务
type AnomalyService struct {
	anomalyRepo *repository.AnomalyRepository
	auditRepo   *repository.AuditRepository
	ruleEngine  *audit.RuleEngine
	detector    *audit.Detector
}

// NewAnomalyService 创建异常检测服务
func NewAnomalyService(
	anomalyRepo *repository.AnomalyRepository,
	auditRepo *repository.AuditRepository,
	ruleEngine *audit.RuleEngine,
	detector *audit.Detector,
) *AnomalyService {
	return &AnomalyService{
		anomalyRepo: anomalyRepo,
		auditRepo:   auditRepo,
		ruleEngine:  ruleEngine,
		detector:    detector,
	}
}

// Scan 执行异常扫描
func (s *AnomalyService) Scan(ctx context.Context) error {
	startTime := time.Now()

	// 获取最近 5 分钟的审计日志
	since := 5 * time.Minute
	logs, err := s.auditRepo.GetRecentLogs(ctx, since)
	if err != nil {
		return fmt.Errorf("fetch recent logs: %w", err)
	}

	if len(logs) == 0 {
		log.Debug().
			Str("module", "anomaly_scanner").
			Msg("no recent logs to scan")
		return nil
	}

	// 加载规则
	rules := s.ruleEngine.GetRules()
	if len(rules) == 0 {
		log.Warn().
			Str("module", "anomaly_scanner").
			Msg("no enabled rules found")
		return nil
	}

	totalDetected := 0
	for _, rule := range rules {
		anomalies := s.detector.Detect(ctx, rule, logs)
		if len(anomalies) > 0 {
			if err := s.anomalyRepo.BatchCreate(ctx, anomalies); err != nil {
				log.Error().
					Err(err).
					Str("module", "anomaly_scanner").
					Str("rule", rule.Type).
					Msg("failed to save anomalies")
				continue
			}
			totalDetected += len(anomalies)
			log.Info().
				Str("module", "anomaly_scanner").
				Str("rule", rule.Type).
				Int("count", len(anomalies)).
				Msg("anomalies detected")
		}
	}

	log.Info().
		Str("module", "anomaly_scanner").
		Int("total_logs", len(logs)).
		Int("total_anomalies", totalDetected).
		Dur("duration", time.Since(startTime)).
		Msg("anomaly scan completed")

	return nil
}

// List 查询异常事件列表
func (s *AnomalyService) List(ctx context.Context, filter *model.AnomalyFilter, page, pageSize int) ([]model.AuditAnomaly, int64, error) {
	return s.anomalyRepo.List(ctx, filter, page, pageSize)
}

// GetByID 根据 ID 获取异常事件
func (s *AnomalyService) GetByID(ctx context.Context, id uid.UID) (*model.AuditAnomaly, error) {
	return s.anomalyRepo.GetByID(ctx, id)
}

// Summary 获取异常概览
func (s *AnomalyService) Summary(ctx context.Context) (*model.AnomalySummaryResponse, error) {
	return s.anomalyRepo.Summary(ctx)
}
