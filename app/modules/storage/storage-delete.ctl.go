package storage

import (
	"errors"

	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (c *Controller) DeleteFile(ctx *gin.Context) {
	rawID, exists := ctx.Get("auth_user_id")
	if !exists {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}
	userID, ok := rawID.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	var req getFileURI
	if err := ctx.ShouldBindUri(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	if err := c.svc.DeleteFile(ctx.Request.Context(), id, userID); err != nil {
		if errors.Is(err, ErrStorageNotFound) {
			_ = base.JSON(ctx, 404, i18n.BadRequest, nil, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, nil)
}
