// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package model

import (
	"encoding/json"
	"time"

	"github.com/contful/contful/admin/pkg/uid"
	"gorm.io/gorm"
)

// Webhook Webhook 配置
type Webhook struct {
	ID          uid.UID `json:"id" gorm:"primaryKey;default:gen_random_uuid()"`
	SiteID      uid.UID `json:"site_id" gorm:"not null;index"`
	Name        string    `json:"name" gorm:"size:255;not null"`
	URL         string    `json:"url" gorm:"size:2000;not null"`
	EventsJSON  string    `json:"-" gorm:"column:events;type:text[];default:'{}'"`
	Secret      string    `json:"secret,omitempty" gorm:"size:255"`
	IsActive    bool      `json:"is_active" gorm:"not null;default:true"`
	CreatedTime time.Time `json:"created_time" gorm:"autoCreateTime"`
	UpdatedTime time.Time `json:"updated_time" gorm:"autoUpdateTime"`

	Events []string `json:"events" gorm:"-"`
}

func (Webhook) TableName() string { return "contful_webhooks" }

func (w *Webhook) AfterFind(tx *gorm.DB) error {
	if w.EventsJSON != "" {
		_ = json.Unmarshal([]byte(w.EventsJSON), &w.Events)
	}
	return nil
}

func (w *Webhook) BeforeSave(tx *gorm.DB) error {
	if len(w.Events) > 0 {
		b, _ := json.Marshal(w.Events)
		w.EventsJSON = string(b)
	} else {
		w.EventsJSON = "{}"
	}
	return nil
}

// WebhookDelivery 投递记录
type WebhookDelivery struct {
	ID             uid.UID `json:"id" gorm:"primaryKey;default:gen_random_uuid()"`
	WebhookID      uid.UID `json:"webhook_id" gorm:"not null;index"`
	Event          string    `json:"event" gorm:"size:50;not null"`
	Payload        string    `json:"payload" gorm:"type:jsonb;not null"`
	ResponseStatus int       `json:"response_status"`
	ResponseBody   string    `json:"response_body,omitempty" gorm:"type:text"`
	Status         string    `json:"status" gorm:"size:20;not null;default:'pending'"`
	Attempt        int       `json:"attempt" gorm:"not null;default:1"`
	ErrorMessage   string    `json:"error_message,omitempty" gorm:"type:text"`
	CreatedTime    time.Time `json:"created_time" gorm:"autoCreateTime"`
}

func (WebhookDelivery) TableName() string { return "contful_webhook_deliveries" }
