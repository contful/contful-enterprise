// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package repository

import (
	"context"

	"github.com/contful/contful/admin/internal/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type WebhookRepository struct{ db *gorm.DB }

func NewWebhookRepository(db *gorm.DB) *WebhookRepository { return &WebhookRepository{db: db} }

func (r *WebhookRepository) ListBySite(ctx context.Context, siteID uuid.UUID) ([]model.Webhook, error) {
	var ws []model.Webhook
	err := r.db.WithContext(ctx).Where("site_id = ?", siteID).Order("created_time DESC").Find(&ws).Error
	return ws, err
}

func (r *WebhookRepository) ListActive(ctx context.Context, siteID uuid.UUID, event string) ([]model.Webhook, error) {
	var ws []model.Webhook
	err := r.db.WithContext(ctx).
		Where("site_id = ? AND is_active = true", siteID).
		Find(&ws).Error
	if err != nil {
		return nil, err
	}
	var matched []model.Webhook
	for _, w := range ws {
		for _, e := range w.Events {
			if e == event {
				matched = append(matched, w)
				break
			}
		}
	}
	return matched, nil
}

func (r *WebhookRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Webhook, error) {
	var w model.Webhook
	err := r.db.WithContext(ctx).First(&w, "id = ?", id).Error
	return &w, err
}

func (r *WebhookRepository) Create(ctx context.Context, w *model.Webhook) error {
	return r.db.WithContext(ctx).Create(w).Error
}

func (r *WebhookRepository) Update(ctx context.Context, w *model.Webhook) error {
	return r.db.WithContext(ctx).Save(w).Error
}

func (r *WebhookRepository) Delete(ctx context.Context, id uuid.UUID) error {
	return r.db.WithContext(ctx).Delete(&model.Webhook{}, "id = ?", id).Error
}

func (r *WebhookRepository) CreateDelivery(ctx context.Context, d *model.WebhookDelivery) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *WebhookRepository) UpdateDelivery(ctx context.Context, d *model.WebhookDelivery) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *WebhookRepository) ListDeliveries(ctx context.Context, webhookID uuid.UUID, limit int) ([]model.WebhookDelivery, error) {
	var ds []model.WebhookDelivery
	q := r.db.WithContext(ctx).Where("webhook_id = ?", webhookID).Order("created_time DESC")
	if limit > 0 {
		q = q.Limit(limit)
	}
	err := q.Find(&ds).Error
	return ds, err
}
