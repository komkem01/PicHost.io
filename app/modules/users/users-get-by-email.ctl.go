package users

import (
	"errors"

	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
)

type GetByEmailRequestController struct {
	Email string `uri:"email" binding:"required,email"`
}

func (c *Controller) GetByEmail(ctx *gin.Context) {
	var req GetByEmailRequestController
	if err := ctx.ShouldBindUri(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	user, err := c.svc.GetByEmail(ctx.Request.Context(), req.Email)
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
