package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
)

// APITokenHandler API Token 处理器
type APITokenHandler struct {
	tokenService *service.APITokenService
}

// NewAPITokenHandler 新建处理器
func NewAPITokenHandler(tokenService *service.APITokenService) *APITokenHandler {
	return &APITokenHandler{tokenService: tokenService}
}

// RegisterRoutes 注册路由
func (h *APITokenHandler) RegisterRoutes(rg *gin.RouterGroup) {
	tokens := rg.Group("/api-tokens")
	{
		tokens.POST("", h.Create)
		tokens.GET("", h.List)
		tokens.GET("/:id", h.Get)
		tokens.PUT("/:id", h.Update)
		tokens.DELETE("/:id", h.Delete)
		tokens.POST("/:id/regenerate", h.Regenerate)
		tokens.POST("/:id/revoke", h.Revoke)
	}
}

// getTokenUserID 从上下文获取用户 ID
func getTokenUserID(c *gin.Context) (uuid.UUID, error) {
	userIDVal, exists := c.Get("user")
	if !exists {
		return uuid.Nil, errors.New("user not found")
	}

	switch v := userIDVal.(type) {
	case string:
		return uuid.Parse(v)
	case uuid.UUID:
		return v, nil
	default:
		return uuid.Nil, errors.New("invalid user id type")
	}
}

// getTokenSiteID 从上下文获取站点 ID
func getTokenSiteID(c *gin.Context) (uuid.UUID, error) {
	// 暂时使用用户 ID 作为 site ID
	return getTokenUserID(c)
}

// Create 创建 Token
func (h *APITokenHandler) Create(c *gin.Context) {
	userID, err := getTokenUserID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	siteID, err := getTokenSiteID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	var req model.APITokenCreate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "请求参数错误"))
		return
	}

	token, fullToken, err := h.tokenService.Create(c.Request.Context(), siteID, userID, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	// 返回包含明文 Token 的响应
	resp := model.APITokenCreateResponse{
		APITokenResponse: token.ToResponse(),
		Token:            fullToken,
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(resp))
}

// List 列出 Token
func (h *APITokenHandler) List(c *gin.Context) {
	siteID, err := getTokenSiteID(c)
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.NewErrorResponse(model.CodeUnauthorized, "unauthorized"))
		return
	}

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	filter := &model.APITokenListFilter{}
	if statusStr := c.Query("status"); statusStr != "" {
		status := model.TokenStatus(statusStr)
		filter.Status = &status
	}
	if name := c.Query("name"); name != "" {
		filter.Name = &name
	}

	result, err := h.tokenService.List(c.Request.Context(), siteID, filter, page, pageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(result))
}

// Get 获取 Token
func (h *APITokenHandler) Get(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的 Token ID"))
		return
	}

	token, err := h.tokenService.Get(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "Token 不存在"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(token.ToResponse()))
}

// Update 更新 Token
func (h *APITokenHandler) Update(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的 Token ID"))
		return
	}

	var req model.APITokenUpdate
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "请求参数错误"))
		return
	}

	token, err := h.tokenService.Update(c.Request.Context(), id, &req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(token.ToResponse()))
}

// Delete 删除 Token
func (h *APITokenHandler) Delete(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的 Token ID"))
		return
	}

	if err := h.tokenService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}

// Regenerate 重新生成 Token
func (h *APITokenHandler) Regenerate(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的 Token ID"))
		return
	}

	token, fullToken, err := h.tokenService.Regenerate(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	// 返回包含新明文 Token 的响应
	resp := model.APITokenCreateResponse{
		APITokenResponse: token.ToResponse(),
		Token:            fullToken,
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(resp))
}

// Revoke 撤销 Token
func (h *APITokenHandler) Revoke(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "无效的 Token ID"))
		return
	}

	if err := h.tokenService.Revoke(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, err.Error()))
		return
	}

	c.Status(http.StatusNoContent)
}
