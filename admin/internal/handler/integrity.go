package handler

import (
	"context"
	"net/http"

	"github.com/contful/contful/admin/internal/middleware"
	"github.com/contful/contful/admin/internal/repository"
	"github.com/contful/contful/admin/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// IntegrityHandler 验签处理器
type IntegrityHandler struct {
	entryRepo     *repository.EntryRepository
	assetRepo     *repository.AssetRepository
	auditRepo     *repository.AuditRepository
	configService *service.ConfigService
}

// NewIntegrityHandler 新建处理器
func NewIntegrityHandler(
	entryRepo *repository.EntryRepository,
	assetRepo *repository.AssetRepository,
	auditRepo *repository.AuditRepository,
	configService *service.ConfigService,
) *IntegrityHandler {
	return &IntegrityHandler{
		entryRepo:     entryRepo,
		assetRepo:     assetRepo,
		auditRepo:     auditRepo,
		configService: configService,
	}
}

// RegisterRoutes 注册路由
func (h *IntegrityHandler) RegisterRoutes(rg *gin.RouterGroup) {
	integrity := rg.Group("/integrity")
	{
		integrity.GET("/verify", h.Verify)
		integrity.POST("/verify/batch", h.BatchVerify)
	}
}

// VerifyRequest 单条验签请求
type VerifyRequest struct {
	Entity string `json:"entity" form:"entity" binding:"required"` // entry / asset / audit_log
	ID     string `json:"id" form:"id" binding:"required"`
}

// Verify 单条验签
func (h *IntegrityHandler) Verify(c *gin.Context) {
	siteID := middleware.GetSiteID(c)
	if siteID == uuid.Nil {
		middleware.BadRequest(c, "X-Site-ID header is required")
		return
	}

	// 从 config 中心读取签名密钥
	var signingKey, alg string
	if h.configService != nil {
		signingKey, _ = h.configService.Get(c.Request.Context(), siteID, "integrity.signing_key")
		alg, _ = h.configService.Get(c.Request.Context(), siteID, "integrity.algorithm")
		if alg == "" {
			alg = "HMAC-SHA256"
		}
	}

	entity := c.Query("entity")
	idStr := c.Query("id")
	if entity == "" || idStr == "" {
		middleware.BadRequest(c, "entity and id are required")
		return
	}

	id, err := uuid.Parse(idStr)
	if err != nil {
		middleware.BadRequest(c, "invalid id format")
		return
	}

	intSvc, _ := service.NewIntegrityService(siteID, signingKey, alg)
	if intSvc == nil {
		middleware.OK(c, gin.H{
			"entity": entity,
			"id":     idStr,
			"valid":  nil,
			"reason": "not_configured",
		})
		return
	}

	result, err := h.verifyEntity(c.Request.Context(), entity, id, siteID, intSvc)
	if err != nil {
		h.handleError(c, err)
		return
	}

	middleware.OK(c, result)
}

// verifyEntity 验签实体
func (h *IntegrityHandler) verifyEntity(ctx context.Context, entity string, id, siteID uuid.UUID, intSvc *service.IntegrityService) (gin.H, error) {
	switch entity {
	case "entry":
		entry, err := h.entryRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if entry.SiteID != siteID {
			// 站点不匹配
			return gin.H{"entity": entity, "id": id.String(), "valid": nil, "reason": "not_found"}, nil
		}
		values, _ := h.entryRepo.GetValuesByEntry(ctx, id)
		vr, _ := intSvc.VerifyEntry(entry, values)
		return gin.H{
			"entity":       entity,
			"id":           id.String(),
			"valid":        vr.Valid,
			"alg":          vr.Alg,
			"created_at":   vr.CreatedAt,
			"payload_hash": vr.PayloadHash,
			"reason":       vr.Reason,
		}, nil

	case "asset":
		asset, err := h.assetRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		if asset.SiteID != siteID {
			return gin.H{"entity": entity, "id": id.String(), "valid": nil, "reason": "not_found"}, nil
		}
		vr, _ := intSvc.VerifyAsset(asset)
		return gin.H{
			"entity":       entity,
			"id":           id.String(),
			"valid":        vr.Valid,
			"alg":          vr.Alg,
			"created_at":   vr.CreatedAt,
			"payload_hash": vr.PayloadHash,
			"reason":       vr.Reason,
		}, nil

	case "audit_log":
		log, err := h.auditRepo.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		vr, _ := intSvc.VerifyAuditLog(log)
		return gin.H{
			"entity":       entity,
			"id":           id.String(),
			"valid":        vr.Valid,
			"alg":          vr.Alg,
			"created_at":   vr.CreatedAt,
			"payload_hash": vr.PayloadHash,
			"reason":       vr.Reason,
		}, nil

	default:
		return gin.H{"entity": entity, "id": id.String(), "valid": nil, "reason": "unsupported_entity"}, nil
	}
}

// BatchVerifyRequest 批量验签请求
type BatchVerifyItem struct {
	Entity string `json:"entity" binding:"required"`
	ID     string `json:"id" binding:"required"`
}

// BatchVerifyRequest 批量验签请求
type BatchVerifyRequest struct {
	Items []BatchVerifyItem `json:"items" binding:"required"`
}

// BatchVerify 批量验签
func (h *IntegrityHandler) BatchVerify(c *gin.Context) {
	siteID := middleware.GetSiteID(c)
	if siteID == uuid.Nil {
		middleware.BadRequest(c, "X-Site-ID header is required")
		return
	}

	var req BatchVerifyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		middleware.BadRequest(c, err.Error())
		return
	}

	if len(req.Items) == 0 {
		middleware.OK(c, gin.H{"total": 0, "valid": 0, "invalid": 0, "not_signed": 0, "results": []interface{}{}})
		return
	}
	if len(req.Items) > 100 {
		middleware.BadRequest(c, "最多支持 100 条批量验签")
		return
	}

	// 读取签名密钥
	var signingKey, alg string
	if h.configService != nil {
		signingKey, _ = h.configService.Get(c.Request.Context(), siteID, "integrity.signing_key")
		alg, _ = h.configService.Get(c.Request.Context(), siteID, "integrity.algorithm")
		if alg == "" {
			alg = "HMAC-SHA256"
		}
	}

	intSvc, _ := service.NewIntegrityService(siteID, signingKey, alg)

	var valid, invalid, notSigned int
	results := make([]gin.H, 0, len(req.Items))

	for _, item := range req.Items {
		id, err := uuid.Parse(item.ID)
		if err != nil {
			results = append(results, gin.H{"entity": item.Entity, "id": item.ID, "valid": nil, "reason": "invalid_id"})
			continue
		}

		result, err := h.verifyEntity(c.Request.Context(), item.Entity, id, siteID, intSvc)
		if err != nil {
			results = append(results, gin.H{"entity": item.Entity, "id": item.ID, "valid": nil, "reason": "not_found"})
			continue
		}

		results = append(results, result)

		v := result["valid"]
		if v == nil {
			notSigned++
		} else if bv, ok := v.(bool); ok && bv {
			valid++
		} else {
			invalid++
		}
	}

	middleware.OK(c, gin.H{
		"total":      len(req.Items),
		"valid":      valid,
		"invalid":    invalid,
		"not_signed": notSigned,
		"results":    results,
	})
}

// handleError 处理错误
func (h *IntegrityHandler) handleError(c *gin.Context, err error) {
	if err == nil {
		return
	}
	if c.Writer.Written() {
		return
	}
	c.JSON(http.StatusInternalServerError, gin.H{"code": 500, "message": err.Error()})
}
