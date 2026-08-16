package admin

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	entitiesinf "pichost.io/app/modules/entities/inf"
	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type Controller struct {
	tracer   trace.Tracer
	svc      *Service
	auditEnt entitiesinf.AuditEntity
}

func newController(tracer trace.Tracer, svc *Service, auditEnt entitiesinf.AuditEntity) *Controller {
	return &Controller{tracer: tracer, svc: svc, auditEnt: auditEnt}
}

func getAdminID(ctx *gin.Context) uuid.UUID {
	raw, _ := ctx.Get("auth_user_id")
	id, _ := raw.(uuid.UUID)
	return id
}

func (c *Controller) recordAudit(action, status string, adminID uuid.UUID, ctx *gin.Context, meta map[string]any, errCode *string) {
	if c.auditEnt == nil {
		return
	}
	ip := ctx.ClientIP()
	ua := ctx.GetHeader("User-Agent")
	go func() {
		reqCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.auditEnt.CreateAuditLog(reqCtx, entitiesdto.CreateAuditLog{
			UserID:    &adminID,
			Action:    action,
			IPAddress: &ip,
			UserAgent: &ua,
			Status:    status,
			Metadata:  meta,
			ErrorCode: errCode,
		})
	}()
}

func strPtr(s string) *string { return &s }

// GET /admin/stats
func (c *Controller) Stats(ctx *gin.Context) {
	stats, err := c.svc.GetDashboardStats(ctx.Request.Context())
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	base.Success(ctx, stats)
}

// GET /admin/plans
func (c *Controller) ListPlanSettings(ctx *gin.Context) {
	plans, err := c.svc.ListPlanSettings(ctx.Request.Context())
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	base.Success(ctx, plans)
}

// GET /public/plans
func (c *Controller) ListPublicPlanSettings(ctx *gin.Context) {
	plans, err := c.svc.ListPlanSettings(ctx.Request.Context())
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	enabled := make([]AdminPlanSetting, 0, len(plans))
	for _, plan := range plans {
		if plan.IsEnabled {
			enabled = append(enabled, plan)
		}
	}
	base.Success(ctx, enabled)
}

// GET /admin/plans/:key
func (c *Controller) GetPlanSetting(ctx *gin.Context) {
	key := strings.TrimSpace(ctx.Param("key"))
	if key == "" {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	plan, err := c.svc.GetPlanSetting(ctx.Request.Context(), key)
	if err != nil {
		base.BadRequest(ctx, "plan not found", nil)
		return
	}
	base.Success(ctx, plan)
}

// PATCH /admin/plans/:key
func (c *Controller) UpsertPlanSetting(ctx *gin.Context) {
	key := strings.TrimSpace(ctx.Param("key"))
	if key == "" {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	var req struct {
		DisplayName       string `json:"display_name" binding:"required"`
		MonthlyPriceTHB   int    `json:"monthly_price_thb"`
		StorageLimitBytes int64  `json:"storage_limit_bytes"`
		ImageLimit        int    `json:"image_limit"`
		MaxUploadMB       int    `json:"max_upload_mb"`
		IsEnabled         *bool  `json:"is_enabled"`
		AllowPrivate      bool   `json:"allow_private"`
		CustomDomain      bool   `json:"custom_domain"`
		APIAccess         bool   `json:"api_access"`
		PrioritySupport   bool   `json:"priority_support"`
		NoAds             bool   `json:"no_ads"`
		WatermarkRemoval  bool   `json:"watermark_removal"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	if req.MonthlyPriceTHB < 0 || req.StorageLimitBytes < 0 || req.ImageLimit < 0 || req.MaxUploadMB < 0 {
		base.BadRequest(ctx, "numeric fields must be non-negative", nil)
		return
	}
	isEnabled := true
	if req.IsEnabled != nil {
		isEnabled = *req.IsEnabled
	}
	plan, err := c.svc.UpsertPlanSetting(ctx.Request.Context(), entitiesdto.UpsertPlanSetting{
		PlanKey:           key,
		DisplayName:       strings.TrimSpace(req.DisplayName),
		MonthlyPriceTHB:   req.MonthlyPriceTHB,
		StorageLimitBytes: req.StorageLimitBytes,
		ImageLimit:        req.ImageLimit,
		MaxUploadMB:       req.MaxUploadMB,
		IsEnabled:         isEnabled,
		AllowPrivate:      req.AllowPrivate,
		CustomDomain:      req.CustomDomain,
		APIAccess:         req.APIAccess,
		PrioritySupport:   req.PrioritySupport,
		NoAds:             req.NoAds,
		WatermarkRemoval:  req.WatermarkRemoval,
	})
	if err != nil {
		c.recordAudit("admin.upsert_plan_setting", "failure", getAdminID(ctx), ctx, map[string]any{"plan_key": key, "error": err.Error()}, strPtr(err.Error()))
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	c.recordAudit("admin.upsert_plan_setting", "success", getAdminID(ctx), ctx, map[string]any{"plan_key": key}, nil)
	base.Success(ctx, plan)
}

// DELETE /admin/plans/:key
func (c *Controller) DeletePlanSetting(ctx *gin.Context) {
	key := strings.TrimSpace(ctx.Param("key"))
	if key == "" {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	if err := c.svc.DeletePlanSetting(ctx.Request.Context(), key); err != nil {
		c.recordAudit("admin.delete_plan_setting", "failure", getAdminID(ctx), ctx, map[string]any{"plan_key": key, "error": err.Error()}, strPtr(err.Error()))
		if errors.Is(err, errPlanInUse) {
			base.BadRequest(ctx, "plan is currently assigned to users", nil)
			return
		}
		if errors.Is(err, errPlanNotFound) {
			base.BadRequest(ctx, "plan not found", nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	c.recordAudit("admin.delete_plan_setting", "success", getAdminID(ctx), ctx, map[string]any{"plan_key": key}, nil)
	base.Success(ctx, gin.H{"ok": true})
}

// GET /admin/users
func (c *Controller) ListUsers(ctx *gin.Context) {
	var filter UserFilter
	filter.Query = ctx.Query("q")
	filter.Plan = ctx.Query("plan")
	if act := ctx.Query("is_active"); act != "" {
		b := act == "true"
		filter.IsActive = &b
	}
	if adm := ctx.Query("is_admin"); adm != "" {
		b := adm == "true"
		filter.IsAdmin = &b
	}

	users, err := c.svc.ListUsers(ctx.Request.Context(), filter)
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	base.Success(ctx, users)
}


// GET /admin/users/:id
func (c *Controller) GetUser(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	user, err := c.svc.GetUser(ctx.Request.Context(), id)
	if err != nil {
		base.BadRequest(ctx, "user not found", nil)
		return
	}
	base.Success(ctx, user)
}

// PATCH /admin/users/:id/plan
func (c *Controller) SetUserPlan(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	var req struct {
		Plan string `json:"plan" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	if err := c.svc.SetUserPlan(ctx.Request.Context(), id, req.Plan); err != nil {
		c.recordAudit("admin.set_user_plan", "failure", getAdminID(ctx), ctx, map[string]any{"target_user_id": id.String(), "plan": req.Plan, "error": err.Error()}, strPtr(err.Error()))
		if errors.Is(err, errInvalidPlan) {
			base.BadRequest(ctx, "invalid plan value", nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	c.recordAudit("admin.set_user_plan", "success", getAdminID(ctx), ctx, map[string]any{"target_user_id": id.String(), "plan": req.Plan}, nil)
	base.Success(ctx, gin.H{"ok": true})
}

// PATCH /admin/users/:id/active
func (c *Controller) SetUserActive(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	var req struct {
		IsActive bool `json:"is_active"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	if err := c.svc.SetUserActive(ctx.Request.Context(), id, req.IsActive); err != nil {
		c.recordAudit("admin.set_user_active", "failure", getAdminID(ctx), ctx, map[string]any{"target_user_id": id.String(), "is_active": req.IsActive, "error": err.Error()}, strPtr(err.Error()))
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	c.recordAudit("admin.set_user_active", "success", getAdminID(ctx), ctx, map[string]any{"target_user_id": id.String(), "is_active": req.IsActive}, nil)
	base.Success(ctx, gin.H{"ok": true})
}

// PATCH /admin/users/:id/admin
func (c *Controller) SetUserAdmin(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	var req struct {
		IsAdmin bool `json:"is_admin"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	if err := c.svc.SetUserAdmin(ctx.Request.Context(), id, req.IsAdmin); err != nil {
		c.recordAudit("admin.set_user_admin", "failure", getAdminID(ctx), ctx, map[string]any{"target_user_id": id.String(), "is_admin": req.IsAdmin, "error": err.Error()}, strPtr(err.Error()))
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	c.recordAudit("admin.set_user_admin", "success", getAdminID(ctx), ctx, map[string]any{"target_user_id": id.String(), "is_admin": req.IsAdmin}, nil)
	base.Success(ctx, gin.H{"ok": true})
}

// DELETE /admin/users/:id
// PATCH /admin/users/:id/profile
func (c *Controller) UpdateProfile(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	var req struct {
		Email    *string `json:"email"`
		Username *string `json:"username"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	u, err := c.svc.UpdateUserProfile(ctx.Request.Context(), id, req.Email, req.Username)
	if err != nil {
		c.recordAudit("admin.update_user_profile", "failure", getAdminID(ctx), ctx, map[string]any{"target_user_id": id.String(), "error": err.Error()}, strPtr(err.Error()))
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	c.recordAudit("admin.update_user_profile", "success", getAdminID(ctx), ctx, map[string]any{"target_user_id": id.String()}, nil)
	base.Success(ctx, u)
}

func (c *Controller) DeleteUser(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	if err := c.svc.DeleteUser(ctx.Request.Context(), id); err != nil {
		c.recordAudit("admin.delete_user", "failure", getAdminID(ctx), ctx, map[string]any{"target_user_id": id.String(), "error": err.Error()}, strPtr(err.Error()))
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	c.recordAudit("admin.delete_user", "success", getAdminID(ctx), ctx, map[string]any{"target_user_id": id.String()}, nil)
	base.Success(ctx, gin.H{"ok": true})
}

// GET /admin/images
func (c *Controller) ListImages(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.Query("limit"))
	page, _ := strconv.Atoi(ctx.Query("page"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	images, total, err := c.svc.ListAllImages(ctx.Request.Context(), limit, offset)
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	base.Success(ctx, gin.H{
		"items": images,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// DELETE /admin/images/:id
func (c *Controller) DeleteImage(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	var req struct {
		Reason string `json:"reason"`
	}
	_ = ctx.ShouldBindJSON(&req)

	if err := c.svc.DeleteImageByAdmin(ctx.Request.Context(), id, req.Reason); err != nil {
		c.recordAudit("admin.delete_image", "failure", getAdminID(ctx), ctx, map[string]any{"image_id": id.String(), "reason": req.Reason, "error": err.Error()}, strPtr(err.Error()))
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	c.recordAudit("admin.delete_image", "success", getAdminID(ctx), ctx, map[string]any{"image_id": id.String(), "reason": req.Reason}, nil)
	base.Success(ctx, gin.H{"ok": true})
}

// --- Legal Documents ---

// GET /public/legal
func (c *Controller) ListPublicLegalDocuments(ctx *gin.Context) {
	docs, err := c.svc.ListLegalDocuments(ctx.Request.Context())
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	base.Success(ctx, docs)
}

// GET /public/legal/:key
func (c *Controller) GetPublicLegalDocument(ctx *gin.Context) {
	key := ctx.Param("key")
	doc, err := c.svc.GetLegalDocumentByKey(ctx.Request.Context(), key)
	if err != nil {
		base.BadRequest(ctx, "document not found", nil)
		return
	}
	base.Success(ctx, doc)
}

// GET /admin/legal
func (c *Controller) ListLegalDocuments(ctx *gin.Context) {
	docs, err := c.svc.ListLegalDocuments(ctx.Request.Context())
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	base.Success(ctx, docs)
}

// GET /admin/legal/:key
func (c *Controller) GetLegalDocument(ctx *gin.Context) {
	key := ctx.Param("key")
	doc, err := c.svc.GetLegalDocumentByKey(ctx.Request.Context(), key)
	if err != nil {
		base.BadRequest(ctx, "document not found", nil)
		return
	}
	base.Success(ctx, doc)
}

// PUT /admin/legal/:key
func (c *Controller) UpsertLegalDocument(ctx *gin.Context) {
	key := ctx.Param("key")
	var req struct {
		Title   string `json:"title" binding:"required"`
		Content string `json:"content" binding:"required"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	doc, err := c.svc.UpsertLegalDocument(ctx.Request.Context(), entitiesdto.UpsertLegalDocument{
		Key:     key,
		Title:   req.Title,
		Content: req.Content,
	})
	if err != nil {
		c.recordAudit("admin.upsert_legal_document", "failure", getAdminID(ctx), ctx, map[string]any{"key": key, "error": err.Error()}, strPtr(err.Error()))
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	c.recordAudit("admin.upsert_legal_document", "success", getAdminID(ctx), ctx, map[string]any{"key": key}, nil)
	base.Success(ctx, doc)
}

// DELETE /admin/legal/:key
func (c *Controller) DeleteLegalDocument(ctx *gin.Context) {
	key := ctx.Param("key")
	if err := c.svc.DeleteLegalDocumentByKey(ctx.Request.Context(), key); err != nil {
		c.recordAudit("admin.delete_legal_document", "failure", getAdminID(ctx), ctx, map[string]any{"key": key, "error": err.Error()}, strPtr(err.Error()))
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	c.recordAudit("admin.delete_legal_document", "success", getAdminID(ctx), ctx, map[string]any{"key": key}, nil)
	base.Success(ctx, gin.H{"ok": true})
}

// POST /admin/users/:id/reset-password
func (c *Controller) ResetUserPassword(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	var req struct {
		Password string `json:"password" binding:"required,min=8"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	if err := c.svc.ResetUserPassword(ctx.Request.Context(), id, req.Password); err != nil {
		c.recordAudit("admin.reset_user_password", "failure", getAdminID(ctx), ctx, map[string]any{"target_user_id": id.String(), "error": err.Error()}, strPtr(err.Error()))
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	c.recordAudit("admin.reset_user_password", "success", getAdminID(ctx), ctx, map[string]any{"target_user_id": id.String()}, nil)
	base.Success(ctx, gin.H{"ok": true, "message": "password reset successfully"})
}

// POST /admin/images/bulk-delete
func (c *Controller) BulkDeleteImages(ctx *gin.Context) {
	var req struct {
		ImageIDs []string `json:"image_ids" binding:"required"`
		Reason   string   `json:"reason"`
	}
	if err := ctx.ShouldBindJSON(&req); err != nil || len(req.ImageIDs) == 0 {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	uuids := make([]uuid.UUID, 0, len(req.ImageIDs))
	for _, idStr := range req.ImageIDs {
		if uid, err := uuid.Parse(idStr); err == nil {
			uuids = append(uuids, uid)
		}
	}

	count, err := c.svc.BulkDeleteImagesByAdmin(ctx.Request.Context(), uuids, req.Reason)
	if err != nil {
		c.recordAudit("admin.bulk_delete_images", "failure", getAdminID(ctx), ctx, map[string]any{"requested": len(req.ImageIDs), "reason": req.Reason, "error": err.Error()}, strPtr(err.Error()))
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	c.recordAudit("admin.bulk_delete_images", "success", getAdminID(ctx), ctx, map[string]any{"deleted_count": count, "reason": req.Reason}, nil)
	base.Success(ctx, gin.H{"ok": true, "deleted_count": count})
}

// --- Storage Management Endpoints ---

// GET /admin/storage/stats
func (c *Controller) GetStorageStats(ctx *gin.Context) {
	stats, err := c.svc.GetStorageStats(ctx.Request.Context())
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	base.Success(ctx, stats)
}

// GET /admin/storage/orphaned
func (c *Controller) ListOrphanedStorage(ctx *gin.Context) {
	limit, _ := strconv.Atoi(ctx.Query("limit"))
	page, _ := strconv.Atoi(ctx.Query("page"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	offset := (page - 1) * limit

	storages, total, err := c.svc.ListOrphanedStorage(ctx.Request.Context(), limit, offset)
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	base.Success(ctx, gin.H{
		"items": storages,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}

// POST /admin/storage/cleanup
func (c *Controller) CleanupOrphanedStorage(ctx *gin.Context) {
	var req struct {
		StorageIDs []string `json:"storage_ids"`
		IDs        []string `json:"ids"`
	}
	_ = ctx.ShouldBindJSON(&req)

	rawIDs := req.StorageIDs
	if len(rawIDs) == 0 {
		rawIDs = req.IDs
	}

	var uuids []uuid.UUID
	for _, idStr := range rawIDs {
		if uid, err := uuid.Parse(idStr); err == nil {
			uuids = append(uuids, uid)
		}
	}

	affected, err := c.svc.CleanupOrphanedStorage(ctx.Request.Context(), uuids)
	if err != nil {
		c.recordAudit("admin.storage_cleanup", "failure", getAdminID(ctx), ctx, map[string]any{"error": err.Error()}, strPtr(err.Error()))
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	c.recordAudit("admin.storage_cleanup", "success", getAdminID(ctx), ctx, map[string]any{"cleaned_files": affected}, nil)
	base.Success(ctx, gin.H{"ok": true, "cleaned_count": affected, "deleted_count": affected})
}



