package audit

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

// GET /admin/audit-logs
func (c *Controller) ListAuditLogs(ctx *gin.Context) {
	var filter entitiesdto.ListAuditLogsFilter

	if uidStr := ctx.Query("user_id"); uidStr != "" {
		if uid, err := uuid.Parse(uidStr); err == nil {
			filter.UserID = &uid
		}
	}

	if action := ctx.Query("action"); action != "" {
		filter.Action = &action
	}

	if status := ctx.Query("status"); status != "" {
		filter.Status = &status
	}

	if from := ctx.Query("from_date"); from != "" {
		filter.FromDate = &from
	}

	if to := ctx.Query("to_date"); to != "" {
		filter.ToDate = &to
	}

	limit, _ := strconv.Atoi(ctx.Query("limit"))
	page, _ := strconv.Atoi(ctx.Query("page"))
	if page < 1 {
		page = 1
	}
	if limit <= 0 {
		limit = 20
	}
	filter.Limit = limit
	filter.Offset = (page - 1) * limit

	logs, total, err := c.svc.ListAuditLogs(ctx.Request.Context(), filter)
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, nil)
		return
	}

	base.Success(ctx, gin.H{
		"items": logs,
		"total": total,
		"page":  page,
		"limit": limit,
	})
}
