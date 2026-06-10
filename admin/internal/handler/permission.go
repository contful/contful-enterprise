// Copyright © 2026-present reepu.com
// SPDX-License-Identifier: Apache-2.0
package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/contful/contful/admin/pkg/uid"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/contful/contful/admin/internal/service"
)

type PermissionHandler struct {
	permRepo    *repository.PermissionRepository
	rbacService *service.RBACService
}

func NewPermissionHandler(permRepo *repository.PermissionRepository, rbacService *service.RBACService) *PermissionHandler {
	return &PermissionHandler{permRepo: permRepo, rbacService: rbacService}
}

// List 获取权限分组及权限项列表
func (h *PermissionHandler) List(c *gin.Context) {
	groups, perms, err := h.permRepo.ListGroupsWithPermissions(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	// 构建嵌套响应
	type PermItem struct {
		ID        string `json:"id"`
		Action    string `json:"action"`
		Label     string `json:"label"`
		LabelEn   string `json:"label_en"`
		SortOrder int    `json:"sort_order"`
	}
	type GroupWithPerms struct {
		ID          string      `json:"id"`
		GroupKey    string      `json:"group_key"`
		Label       string      `json:"label"`
		LabelEn     string      `json:"label_en"`
		SortOrder   int         `json:"sort_order"`
		Permissions []PermItem  `json:"permissions"`
	}

	permMap := make(map[uid.UID][]model.Permission)
	for _, p := range perms {
		permMap[p.GroupID] = append(permMap[p.GroupID], p)
	}

	result := make([]GroupWithPerms, 0, len(groups))
	for _, g := range groups {
		gwp := GroupWithPerms{
			ID:        g.ID.String(),
			GroupKey:  g.GroupKey,
			Label:     g.Label,
			LabelEn:   g.LabelEn,
			SortOrder: g.SortOrder,
		}
		if items, ok := permMap[g.ID]; ok {
			for _, p := range items {
				gwp.Permissions = append(gwp.Permissions, PermItem{
					ID:        p.ID.String(),
					Action:    p.Action,
					Label:     p.Label,
					LabelEn:   p.LabelEn,
					SortOrder: p.SortOrder,
				})
			}
		}
		result = append(result, gwp)
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(result))
}

// ── 权限分组 CRUD ──

func (h *PermissionHandler) CreateGroup(c *gin.Context) {
	var req struct {
		GroupKey  string `json:"group_key" binding:"required"`
		Label     string `json:"label" binding:"required"`
		LabelEn   string `json:"label_en"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	g := &model.PermissionGroup{
		ID:        uid.New(),
		GroupKey:  req.GroupKey,
		Label:     req.Label,
		LabelEn:   req.LabelEn,
		SortOrder: req.SortOrder,
	}
	if err := h.permRepo.CreateGroup(c.Request.Context(), g); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}
	c.JSON(http.StatusCreated, model.NewSuccessResponse(g))
	h.invalidateCache()
}

func (h *PermissionHandler) UpdateGroup(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Label     *string `json:"label"`
		LabelEn   *string `json:"label_en"`
		SortOrder *int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	g := &model.PermissionGroup{}
	gid, err := uid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid id"))
		return
	}
	g.ID = gid
	if req.Label != nil {
		g.Label = *req.Label
	}
	if req.LabelEn != nil {
		g.LabelEn = *req.LabelEn
	}
	if req.SortOrder != nil {
		g.SortOrder = *req.SortOrder
	}

	if err := h.permRepo.UpdateGroup(c.Request.Context(), g); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"message": "updated"}))
	h.invalidateCache()
}

func (h *PermissionHandler) DeleteGroup(c *gin.Context) {
	id := c.Param("id")
	if err := h.permRepo.DeleteGroup(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"message": "deleted"}))
	h.invalidateCache()
}

// ── 权限项 CRUD ──

func (h *PermissionHandler) CreatePermission(c *gin.Context) {
	var req struct {
		GroupID   string `json:"group_id" binding:"required"`
		Action    string `json:"action" binding:"required"`
		Label     string `json:"label" binding:"required"`
		LabelEn   string `json:"label_en"`
		SortOrder int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	groupID, err := uid.Parse(req.GroupID)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid group_id"))
		return
	}

	p := &model.Permission{
		ID:        uid.New(),
		GroupID:   groupID,
		Action:    req.Action,
		Label:     req.Label,
		LabelEn:   req.LabelEn,
		SortOrder: req.SortOrder,
	}
	if err := h.permRepo.CreatePermission(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}
	c.JSON(http.StatusCreated, model.NewSuccessResponse(p))
	h.invalidateCache()
}

func (h *PermissionHandler) UpdatePermission(c *gin.Context) {
	id := c.Param("id")
	var req struct {
		Label     *string `json:"label"`
		LabelEn   *string `json:"label_en"`
		SortOrder *int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	p := &model.Permission{}
	pid, err := uid.Parse(id)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid id"))
		return
	}
	p.ID = pid
	if req.Label != nil {
		p.Label = *req.Label
	}
	if req.LabelEn != nil {
		p.LabelEn = *req.LabelEn
	}
	if req.SortOrder != nil {
		p.SortOrder = *req.SortOrder
	}

	if err := h.permRepo.UpdatePermission(c.Request.Context(), p); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"message": "updated"}))
	h.invalidateCache()
}

func (h *PermissionHandler) DeletePermission(c *gin.Context) {
	id := c.Param("id")
	if err := h.permRepo.DeletePermission(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}
	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"message": "deleted"}))
	h.invalidateCache()
}

func (h *PermissionHandler) invalidateCache() {
	if h.rbacService != nil {
		h.rbacService.InvalidatePermissionCache()
	}
}

// ClearCache 手动清除权限元数据缓存
func (h *PermissionHandler) ClearCache(c *gin.Context) {
	h.invalidateCache()
	c.JSON(http.StatusOK, model.NewSuccessResponse(gin.H{"message": "缓存已清除"}))
}
