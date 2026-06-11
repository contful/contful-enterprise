// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package service

import (
	"context"
	"time"

	"github.com/contful/contful/admin/internal/database"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/metrics"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/contful/contful/admin/pkg/uid"
	"github.com/rs/zerolog"
	"gorm.io/gorm"
)

// ScheduleService 排期服务（Cron 调度 + 列表查询）
type ScheduleService struct {
	db        *gorm.DB
	entryRepo *repository.EntryRepository
	logger    zerolog.Logger
}

// NewScheduleService 新建排期服务
func NewScheduleService(db *gorm.DB, entryRepo *repository.EntryRepository, logger zerolog.Logger) *ScheduleService {
	return &ScheduleService{
		db:        db,
		entryRepo: entryRepo,
		logger:    logger,
	}
}

// StartCron 启动定时扫描（每 60s 执行一次）
func (s *ScheduleService) StartCron(ctx context.Context) {
	s.logger.Info().Msg("排期调度器已启动（间隔 60s）")

	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	// 立即执行一次
	s.scan(ctx)

	for {
		select {
		case <-ctx.Done():
			s.logger.Info().Msg("排期调度器已停止")
			return
		case <-ticker.C:
			s.scan(ctx)
		}
	}
}

// scan 执行一次完整的排期扫描
func (s *ScheduleService) scan(ctx context.Context) {
	// advisory lock 仅 PostgreSQL 支持，达梦跳过
	if database.CurrentDBType() == "postgres" {
		const lockKey = 1735202518135900000
		var ok bool
		if err := s.db.WithContext(ctx).Raw("SELECT pg_try_advisory_lock(?)", lockKey).Scan(&ok).Error; err != nil {
			s.logger.Error().Err(err).Msg("advisory lock 查询失败")
			return
		}
		if !ok {
			s.logger.Debug().Msg("advisory lock 未获取，其他副本正在执行")
			return
		}
		defer func() {
			_ = s.db.Exec("SELECT pg_advisory_unlock(?)", lockKey)
		}()
	}

	pub, pubSkip, pubErr := s.publishEntries(ctx)
	unpub, unpubSkip, unpubErr := s.unpublishEntries(ctx)
	metrics.G().RecordSchedule(pub, unpub, pubSkip+unpubSkip, pubErr+unpubErr)
}

// publishEntries 处理到期待发布的条目
func (s *ScheduleService) publishEntries(ctx context.Context) (published, skipped, errors int64) {
	entries, err := s.entryRepo.FindDuePublish(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("查询待发布条目失败")
		return
	}

	if len(entries) == 0 {
		return
	}

	s.logger.Info().Int("count", len(entries)).Msg("发现待发布条目")

	for _, entry := range entries {
		if entry.Status == model.EntryStatusDraft {
			now := time.Now()
			entry.Status = model.EntryStatusPublished
			entry.PublishedTime = &now
			entry.Version++
			entry.ScheduledPublishTime = nil

			if err := s.entryRepo.Update(ctx, &entry); err != nil {
				s.logger.Error().Err(err).Str("entry_id", entry.ID.String()).Msg("发布条目失败")
				errors++
				continue
			}
			published++
			s.logger.Info().Str("entry_id", entry.ID.String()).Msg("条目已自动发布")
		} else {
			entry.ScheduledPublishTime = nil
			if err := s.entryRepo.Update(ctx, &entry); err != nil {
				s.logger.Error().Err(err).Str("entry_id", entry.ID.String()).Msg("清除已变更条目的排期失败")
				errors++
				continue
			}
			skipped++
		}
	}
	return
}

// unpublishEntries 处理到期待下架的条目
func (s *ScheduleService) unpublishEntries(ctx context.Context) (unpublished, skipped, errors int64) {
	entries, err := s.entryRepo.FindDueUnpublish(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("查询待下架条目失败")
		return
	}

	if len(entries) == 0 {
		return
	}

	s.logger.Info().Int("count", len(entries)).Msg("发现待下架条目")

	for _, entry := range entries {
		if entry.Status == model.EntryStatusPublished {
			entry.Status = model.EntryStatusDraft
			entry.PublishedTime = nil
			entry.ScheduledUnpublishTime = nil

			if err := s.entryRepo.Update(ctx, &entry); err != nil {
				s.logger.Error().Err(err).Str("entry_id", entry.ID.String()).Msg("下架条目失败")
				errors++
				continue
			}
			unpublished++
			s.logger.Info().Str("entry_id", entry.ID.String()).Msg("条目已自动下架")
		} else {
			entry.ScheduledUnpublishTime = nil
			if err := s.entryRepo.Update(ctx, &entry); err != nil {
				s.logger.Error().Err(err).Str("entry_id", entry.ID.String()).Msg("清除已变更条目的排期失败")
				errors++
				continue
			}
		}
	}
	return
}

// ListScheduled 查询排期条目列表
func (s *ScheduleService) ListScheduled(ctx context.Context, siteID uid.UID, filter *model.ScheduledEntryFilter) ([]model.Entry, int64, error) {
	return s.entryRepo.ListScheduled(ctx, siteID, filter)
}
