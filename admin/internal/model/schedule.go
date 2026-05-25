// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"time"
)

// ============ Schedule DTO ============

// ScheduleRequest 排期设置请求
type ScheduleRequest struct {
	ScheduledPublishTime   *time.Time `json:"scheduled_publish_time"`
	ScheduledUnpublishTime *time.Time `json:"scheduled_unpublish_time"`
}

// ScheduledEntryFilter 排期列表过滤条件
type ScheduledEntryFilter struct {
	Status   *string    `json:"status"` // pending_publish / pending_unpublish / all
	From     *time.Time `json:"from"`
	To       *time.Time `json:"to"`
	Page     int        `json:"page"`
	PageSize int        `json:"page_size"`
}
