package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/contful/contful/admin/internal/model"
	"github.com/contful/contful/admin/internal/service"
)

type UserHandler struct {
	userService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{userService: userService}
}

// Create 创建用户
func (h *UserHandler) Create(c *gin.Context) {
	var req model.CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	user, err := h.userService.Create(c.Request.Context(), &req)
	if err != nil {
		switch err {
		case service.ErrUserAlreadyExists:
			c.JSON(http.StatusConflict, model.NewErrorResponse(model.CodeConflict, "user already exists"))
		default:
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}

	c.JSON(http.StatusCreated, model.NewSuccessResponse(user))
}

// Get 获取单个用户
func (h *UserHandler) Get(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid user id"))
		return
	}

	user, err := h.userService.Get(c.Request.Context(), id)
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "user not found"))
		} else {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(user))
}

// List 分页列表
func (h *UserHandler) List(c *gin.Context) {
	var req model.PageRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	pageResp, err := h.userService.List(c.Request.Context(), req.Page, req.PageSize)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(pageResp))
}

// Update 更新用户
func (h *UserHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid user id"))
		return
	}

	var req model.UpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, err.Error()))
		return
	}

	user, err := h.userService.Update(c.Request.Context(), id, &req)
	if err != nil {
		if err == service.ErrUserNotFound {
			c.JSON(http.StatusNotFound, model.NewErrorResponse(model.CodeNotFound, "user not found"))
		} else {
			c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		}
		return
	}

	c.JSON(http.StatusOK, model.NewSuccessResponse(user))
}

// Delete 删除用户
func (h *UserHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, model.NewErrorResponse(model.CodeBadRequest, "invalid user id"))
		return
	}

	if err := h.userService.Delete(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, model.NewErrorResponse(model.CodeInternalError, "internal error"))
		return
	}

	c.JSON(http.StatusNoContent, nil)
}
