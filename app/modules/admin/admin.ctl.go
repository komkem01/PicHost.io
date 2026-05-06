package admin

import (
	"errors"
	"strings"

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

func (c *Controller) recordAudit(action, status string, adminID uuid.UUID, ctx *gin.Context) {
	if c.auditEnt == nil {
		return
	}
	ip := ctx.ClientIP()
	ua := ctx.GetHeader("User-Agent")
	go func() {
		_ = c.auditEnt.CreateAuditLog(ctx.Request.Context(), entitiesdto.CreateAuditLog{
			UserID:    &adminID,
			Action:    action,
			IPAddress: &ip,
			UserAgent: &ua,
			Status:    status,
		})
	}()
}

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
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	c.recordAudit("admin.upsert_plan_setting", "success", getAdminID(ctx), ctx)
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

	c.recordAudit("admin.delete_plan_setting", "success", getAdminID(ctx), ctx)
	base.Success(ctx, gin.H{"ok": true})
}

// GET /admin/users
func (c *Controller) ListUsers(ctx *gin.Context) {
	users, err := c.svc.ListUsers(ctx.Request.Context())
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
		if errors.Is(err, errInvalidPlan) {
			base.BadRequest(ctx, "invalid plan value", nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	c.recordAudit("admin.set_user_plan", "success", getAdminID(ctx), ctx)
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
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	c.recordAudit("admin.set_user_active", "success", getAdminID(ctx), ctx)
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
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	c.recordAudit("admin.set_user_admin", "success", getAdminID(ctx), ctx)
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
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	c.recordAudit("admin.update_user_profile", "success", getAdminID(ctx), ctx)
	base.Success(ctx, u)
}

func (c *Controller) DeleteUser(ctx *gin.Context) {
	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	if err := c.svc.DeleteUser(ctx.Request.Context(), id); err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}
	c.recordAudit("admin.delete_user", "success", getAdminID(ctx), ctx)
	base.Success(ctx, gin.H{"ok": true})
}
