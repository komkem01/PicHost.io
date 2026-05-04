package users

import (
	"errors"

	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GetUserRequestController struct {
	ID string `uri:"id" binding:"required"`
}

func (c *Controller) Get(ctx *gin.Context) {
	var req GetUserRequestController
	if err := ctx.ShouldBindUri(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	user, err := c.svc.Get(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			_ = base.JSON(ctx, 404, i18n.UserNotFound, nil, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, toUserResponseController(user), i18n.UserFetched)
}
