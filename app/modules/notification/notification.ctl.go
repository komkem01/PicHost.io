package notification

import (
	"strconv"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
)

type Controller struct {
	tracer trace.Tracer
	svc    *Service
}

func newController(tracer trace.Tracer, svc *Service) *Controller {
	return &Controller{tracer: tracer, svc: svc}
}

func getUserID(ctx *gin.Context) *uuid.UUID {
	if raw, exists := ctx.Get("auth_user_id"); exists {
		if id, ok := raw.(uuid.UUID); ok {
			return &id
		}
	}
	return nil
}

func getIsAdmin(ctx *gin.Context) bool {
	if raw, exists := ctx.Get("auth_is_admin"); exists {
		if b, ok := raw.(bool); ok {
			return b
		}
	}
	return false
}

// GET /notifications
func (c *Controller) ListMyNotifications(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == nil {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	page, _ := strconv.Atoi(ctx.Query("page"))
	limit, _ := strconv.Atoi(ctx.Query("limit"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	unreadOnly := ctx.Query("unread") == "true"
	notifType := ctx.Query("type")

	items, total, err := c.svc.List(ctx.Request.Context(), entitiesdto.NotificationFilter{
		UserID:     uid,
		UnreadOnly: unreadOnly,
		Type:       notifType,
		Page:       page,
		Limit:      limit,
	})
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	unreadCount, _ := c.svc.GetUnreadCount(ctx.Request.Context(), uid, "user")

	base.Success(ctx, gin.H{
		"items":        items,
		"total":        total,
		"unread_count": unreadCount,
		"page":         page,
		"limit":        limit,
	})
}

// GET /notifications/unread-count
func (c *Controller) GetMyUnreadCount(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == nil {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	count, err := c.svc.GetUnreadCount(ctx.Request.Context(), uid, "user")
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	base.Success(ctx, gin.H{
		"unread_count": count,
	})
}

// PATCH /notifications/:id/read
func (c *Controller) MarkAsRead(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == nil {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	if err := c.svc.MarkAsRead(ctx.Request.Context(), id, uid); err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	base.Success(ctx, gin.H{"ok": true})
}

// POST /notifications/read-all
func (c *Controller) MarkAllAsRead(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == nil {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	if err := c.svc.MarkAllAsRead(ctx.Request.Context(), uid, "user"); err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	base.Success(ctx, gin.H{"ok": true})
}

// DELETE /notifications/:id
func (c *Controller) DeleteNotification(ctx *gin.Context) {
	uid := getUserID(ctx)
	if uid == nil {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	id, err := uuid.Parse(ctx.Param("id"))
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	if err := c.svc.Delete(ctx.Request.Context(), id, uid); err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	base.Success(ctx, gin.H{"ok": true})
}

// Admin handlers

// GET /admin/notifications
func (c *Controller) AdminListNotifications(ctx *gin.Context) {
	page, _ := strconv.Atoi(ctx.Query("page"))
	limit, _ := strconv.Atoi(ctx.Query("limit"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}

	unreadOnly := ctx.Query("unread") == "true"
	notifType := ctx.Query("type")

	items, total, err := c.svc.List(ctx.Request.Context(), entitiesdto.NotificationFilter{
		IsAdmin:    true,
		UnreadOnly: unreadOnly,
		Type:       notifType,
		Page:       page,
		Limit:      limit,
	})
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	unreadCount, _ := c.svc.GetUnreadCount(ctx.Request.Context(), nil, "admin")

	base.Success(ctx, gin.H{
		"items":        items,
		"total":        total,
		"unread_count": unreadCount,
		"page":         page,
		"limit":        limit,
	})
}

// POST /admin/notifications/broadcast
func (c *Controller) AdminBroadcast(ctx *gin.Context) {
	var req entitiesdto.BroadcastNotificationRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	targetRole := req.TargetRole
	if targetRole == "" {
		targetRole = "all"
	}

	notifType := req.Type
	if notifType == "" {
		notifType = "announcement"
	}

	created, err := c.svc.Create(ctx.Request.Context(), entitiesdto.CreateNotification{
		UserID:     req.UserID,
		TargetRole: targetRole,
		Type:       notifType,
		Title:      req.Title,
		Message:    req.Message,
		Link:       req.Link,
		Metadata:   req.Metadata,
	})
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	base.Success(ctx, created)
}
