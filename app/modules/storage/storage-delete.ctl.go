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
		errStr := err.Error()
		c.recordAudit("storage.delete_file", "failure", uuidPtr(userID), strPtr("storage"), uuidPtr(id), ctx.ClientIP(), ctx.GetHeader("User-Agent"), map[string]any{"file_id": id.String(), "error": errStr}, &errStr)
		if errors.Is(err, ErrStorageNotFound) {
			_ = base.JSON(ctx, 404, i18n.BadRequest, nil, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	c.recordAudit("storage.delete_file", "success", uuidPtr(userID), strPtr("storage"), uuidPtr(id), ctx.ClientIP(), ctx.GetHeader("User-Agent"), map[string]any{"file_id": id.String()}, nil)
	base.Success(ctx, nil)
}
