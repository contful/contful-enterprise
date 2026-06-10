// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/contful/contful/admin/pkg/uid"
)

// WebhookHandler Webhook 管理接口
type WebhookHandler struct {
	repo       *repository.WebhookRepository
	dispatcher *WebhookDispatcher
}

func NewWebhookHandler(repo *repository.WebhookRepository) *WebhookHandler {
	return &WebhookHandler{
		repo:       repo,
		dispatcher: NewWebhookDispatcher(repo),
	}
}

func (h *WebhookHandler) GetDispatcher() *WebhookDispatcher { return h.dispatcher }

func (h *WebhookHandler) List(c *gin.Context) {
	siteID, _ := getSiteID(c)
	ws, err := h.repo.ListBySite(c.Request.Context(), siteID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, err.Error()))
		return
	}
	if ws == nil {
		ws = []model.Webhook{}
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(ws))
}

func (h *WebhookHandler) Get(c *gin.Context) {
	id := parseUUID(c.Param("id"))
	w, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(404, "webhook not found"))
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(w))
}

func (h *WebhookHandler) Create(c *gin.Context) {
	var w model.Webhook
	if err := c.ShouldBindJSON(&w); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, err.Error()))
		return
	}
	w.ID = uid.New()
	siteID, _ := getSiteID(c)
	w.SiteID = siteID
	w.IsActive = true
	w.CreatedTime = time.Now()
	w.UpdatedTime = time.Now()
	if err := h.repo.Create(c.Request.Context(), &w); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, err.Error()))
		return
	}
	c.JSON(http.StatusCreated, model.NewSuccessResponse(w))
}

func (h *WebhookHandler) Update(c *gin.Context) {
	id := parseUUID(c.Param("id"))
	exist, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(404, "webhook not found"))
		return
	}
	var input model.Webhook
	if err := c.ShouldBindJSON(&input); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(400, err.Error()))
		return
	}
	if input.Name != "" {
		exist.Name = input.Name
	}
	if input.URL != "" {
		exist.URL = input.URL
	}
	if len(input.Events) > 0 {
		exist.Events = input.Events
	}
	exist.Secret = input.Secret
	exist.IsActive = input.IsActive
	exist.UpdatedTime = time.Now()
	if err := h.repo.Update(c.Request.Context(), exist); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(exist))
}

func (h *WebhookHandler) Delete(c *gin.Context) {
	id := parseUUID(c.Param("id"))
	if err := h.repo.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"deleted": true}))
}

func (h *WebhookHandler) ListDeliveries(c *gin.Context) {
	id := parseUUID(c.Param("id"))
	ds, err := h.repo.ListDeliveries(c.Request.Context(), id, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(500, err.Error()))
		return
	}
	if ds == nil {
		ds = []model.WebhookDelivery{}
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(ds))
}

func (h *WebhookHandler) Test(c *gin.Context) {
	id := parseUUID(c.Param("id"))
	w, err := h.repo.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(404, "webhook not found"))
		return
	}
	go h.dispatcher.Deliver(c.Request.Context(), w, "ping", `{"test":true}`, 1)
	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"message": "test delivery sent"}))
}

// ============================================================================
// WebhookDispatcher
// ============================================================================

type WebhookDispatcher struct {
	repo *repository.WebhookRepository
	mu   sync.Mutex
	sem  chan struct{}
}

func NewWebhookDispatcher(repo *repository.WebhookRepository) *WebhookDispatcher {
	return &WebhookDispatcher{
		repo: repo,
		sem:  make(chan struct{}, 10),
	}
}

func (d *WebhookDispatcher) Emit(siteID uid.UID, event string, payloadJSON string) {
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		ws, err := d.repo.ListActive(ctx, siteID, event)
		if err != nil || len(ws) == 0 {
			return
		}
		for i := range ws {
			d.sem <- struct{}{}
			go func(w *model.Webhook) {
				defer func() { <-d.sem }()
				d.Deliver(ctx, w, event, payloadJSON, 1)
			}(&ws[i])
		}
	}()
}

func (d *WebhookDispatcher) Deliver(ctx context.Context, w *model.Webhook, event string, payloadJSON string, attempt int) {
	if ctx.Err() != nil {
		return
	}

	delivery := &model.WebhookDelivery{
		ID:          uid.New(),
		WebhookID:   w.ID,
		Event:       event,
		Payload:     payloadJSON,
		Status:      "pending",
		Attempt:     attempt,
		CreatedTime: time.Now(),
	}
	_ = d.repo.CreateDelivery(ctx, delivery)

	client := &http.Client{Timeout: 10 * time.Second}
	req, _ := http.NewRequestWithContext(ctx, "POST", w.URL, strings.NewReader(payloadJSON))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Contful-Event", event)
	req.Header.Set("X-Contful-Delivery-ID", delivery.ID.String())
	req.Header.Set("User-Agent", "Contful-Webhook/1.0")

	if w.Secret != "" {
		mac := hmac.New(sha256.New, []byte(w.Secret))
		mac.Write([]byte(payloadJSON))
		sig := "sha256=" + hex.EncodeToString(mac.Sum(nil))
		req.Header.Set("X-Contful-Signature", sig)
	}

	resp, err := client.Do(req)
	if err != nil {
		delivery.Status = "failed"
		delivery.ErrorMessage = err.Error()
		d.retry(ctx, w, event, payloadJSON, attempt, delivery)
		return
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
	delivery.Status = "success"
	delivery.ResponseStatus = resp.StatusCode
	delivery.ResponseBody = string(body)
	_ = d.repo.UpdateDelivery(ctx, delivery)
}

func (d *WebhookDispatcher) retry(ctx context.Context, w *model.Webhook, event, payload string, attempt int, delivery *model.WebhookDelivery) {
	_ = d.repo.UpdateDelivery(ctx, delivery)
	if attempt >= 3 {
		return
	}
	backoff := []time.Duration{1, 4, 16}[attempt-1] * time.Second
	time.Sleep(backoff)
	d.Deliver(ctx, w, event, payload, attempt+1)
}

func parseUUID(s string) uid.UID {
	id, _ := uid.Parse(s)
	return id
}
