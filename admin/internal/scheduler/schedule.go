// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package scheduler

import (
	"context"
	"sync"
	"time"

	"github.com/contful/contful/admin/internal/repository"
	"github.com/rs/zerolog"
)

// ScheduleService 定时发布排期调度器
type ScheduleService struct {
	entryRepo *repository.EntryRepository
	logger    zerolog.Logger
	interval  time.Duration
	stopCh    chan struct{}
	wg        sync.WaitGroup
}

// NewScheduleService 创建调度器
func NewScheduleService(
	entryRepo *repository.EntryRepository,
	interval time.Duration,
	logger zerolog.Logger,
) *ScheduleService {
	return &ScheduleService{
		entryRepo: entryRepo,
		logger:    logger,
		interval:  interval,
		stopCh:    make(chan struct{}),
	}
}

// Start 启动调度器（后台 goroutine）
func (s *ScheduleService) Start(ctx context.Context) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		ticker := time.NewTicker(s.interval)
		defer ticker.Stop()

		s.logger.Info().
			Dur("interval", s.interval).
			Msg("schedule service started")

		// 首次立即执行，不等待第一个 tick
		s.run(ctx)

		for {
			select {
			case <-ticker.C:
				s.run(ctx)
			case <-s.stopCh:
				s.logger.Info().Msg("schedule service stopped")
				return
			case <-ctx.Done():
				s.logger.Info().Msg("schedule service context cancelled")
				return
			}
		}
	}()
}

// Stop 停止调度器（优雅关闭）
func (s *ScheduleService) Stop() {
	close(s.stopCh)
	s.wg.Wait()
}

// run 执行一次定时扫描
func (s *ScheduleService) run(ctx context.Context) {
	// 1. 定时发布：draft → published
	toPublish, err := s.entryRepo.FindScheduledToPublish(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("schedule: failed to find entries to publish")
	} else {
		for _, entry := range toPublish {
			_, execErr := s.entryRepo.ExecuteScheduledPublish(ctx, entry.ID)
			if execErr != nil {
				s.logger.Error().
					Err(execErr).
					Str("entry_id", entry.ID.String()).
					Msg("schedule: failed to auto-publish entry")
			} else {
				s.logger.Info().
					Str("entry_id", entry.ID.String()).
					Str("site_id", entry.SiteID.String()).
					Msg("schedule: entry auto-published")
			}
		}
		if len(toPublish) > 0 {
			s.logger.Info().Int("count", len(toPublish)).Msg("schedule: publish batch complete")
		}
	}

	// 2. 定时下架：published → draft
	toUnpublish, err := s.entryRepo.FindScheduledToUnpublish(ctx)
	if err != nil {
		s.logger.Error().Err(err).Msg("schedule: failed to find entries to unpublish")
	} else {
		for _, entry := range toUnpublish {
			_, execErr := s.entryRepo.ExecuteScheduledUnpublish(ctx, entry.ID)
			if execErr != nil {
				s.logger.Error().
					Err(execErr).
					Str("entry_id", entry.ID.String()).
					Msg("schedule: failed to auto-unpublish entry")
			} else {
				s.logger.Info().
					Str("entry_id", entry.ID.String()).
					Str("site_id", entry.SiteID.String()).
					Msg("schedule: entry auto-unpublished")
			}
		}
		if len(toUnpublish) > 0 {
			s.logger.Info().Int("count", len(toUnpublish)).Msg("schedule: unpublish batch complete")
		}
	}
}
