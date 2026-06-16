// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"fmt"

	"github.com/robfig/cron/v3"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

// Scanner 定时异常扫描调度器
type Scanner struct {
	cron   *cron.Cron
	db     *gorm.DB
	anomalySvc *AnomalyService
}

// lockID pg_try_advisory_lock 锁 ID
const anomalyScanLockID = 12345

// NewScanner 创建扫描调度器
func NewScanner(db *gorm.DB, anomalySvc *AnomalyService) *Scanner {
	return &Scanner{
		cron:       cron.New(cron.WithSeconds()),
		db:         db,
		anomalySvc: anomalySvc,
	}
}

// Start 启动定时扫描（每 5 分钟）
func (s *Scanner) Start() error {
	_, err := s.cron.AddFunc("0 */5 * * * *", s.runScan)
	if err != nil {
		return fmt.Errorf("register scan cron job: %w", err)
	}

	s.cron.Start()

	log.Info().
		Str("module", "scanner").
		Msg("anomaly scanner started (every 5 minutes)")

	return nil
}

// Stop 优雅关闭
func (s *Scanner) Stop() {
	if s.cron != nil {
		ctx := s.cron.Stop()
		<-ctx.Done()
	}
	log.Info().
		Str("module", "scanner").
		Msg("anomaly scanner stopped")
}

// runScan 执行扫描（cron 回调）
func (s *Scanner) runScan() {
	ctx := context.Background()

	// 获取 advisory lock 防止多实例并发扫描
	var locked bool
	if err := s.db.WithContext(ctx).
		Raw("SELECT pg_try_advisory_lock(?)", anomalyScanLockID).
		Scan(&locked).Error; err != nil || !locked {
		if err != nil {
			log.Error().
				Err(err).
				Str("module", "scanner").
				Msg("failed to acquire advisory lock")
		} else {
			log.Info().
				Str("module", "scanner").
				Int64("lock_id", anomalyScanLockID).
				Msg("scan skipped - another instance is running")
		}
		return
	}

	defer func() {
		if err := s.db.WithContext(ctx).
			Exec("SELECT pg_advisory_unlock(?)", anomalyScanLockID).Error; err != nil {
			log.Error().
				Err(err).
				Str("module", "scanner").
				Msg("failed to release advisory lock")
		}
	}()

	if err := s.anomalySvc.Scan(ctx); err != nil {
		log.Error().
			Err(err).
			Str("module", "scanner").
			Msg("scan failed")
	}
}

// ScanNow 手动触发扫描
func (s *Scanner) ScanNow(ctx context.Context) error {
	log.Info().
		Str("module", "scanner").
		Msg("manual scan triggered")

	return s.anomalySvc.Scan(ctx)
}
