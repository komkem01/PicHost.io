package users

import (
	"errors"

	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func (c *Controller) Delete(ctx *gin.Context) {
	var uri GetUserRequestController
	if err := ctx.ShouldBindUri(&uri); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	id, err := uuid.Parse(uri.ID)
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	if err := c.svc.Delete(ctx.Request.Context(), id); err != nil {
		if errors.Is(err, ErrUserNotFound) {
			_ = base.JSON(ctx, 404, i18n.UserNotFound, nil, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, nil, i18n.UserDeleted)
}
