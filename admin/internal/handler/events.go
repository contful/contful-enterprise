// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// EventBus 全局事件总线（Webhook 事件触发器）
var EventBus = &eventBus{}

type eventBus struct {
	dispatcher *WebhookDispatcher
}

type WebhookEvent struct {
	Event       string      `json:"event"`
	SiteID      uuid.UUID   `json:"site_id"`
	CreatedTime time.Time   `json:"created_time"`
	Data        interface{} `json:"data"`
	Previous    interface{} `json:"previous,omitempty"`
}

func SetWebhookDispatcher(d *WebhookDispatcher) {
	EventBus.dispatcher = d
}

func EmitWebhookEvent(siteID uuid.UUID, event string, data interface{}, previous interface{}) {
	if EventBus.dispatcher == nil {
		return
	}
	payload := WebhookEvent{
		Event:       event,
		SiteID:      siteID,
		CreatedTime: time.Now(),
		Data:        data,
		Previous:    previous,
	}
	b, err := json.Marshal(payload)
	if err != nil {
		return
	}
	EventBus.dispatcher.Emit(siteID, event, string(b))
}
