package users

import (
	"errors"

	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UpdateUserRequestController struct {
	Email    *string `json:"email"`
	Username *string `json:"username"`
	IsActive *bool   `json:"is_active"`
	Plan     *string `json:"plan"`
	IsGuest  *bool   `json:"is_guest"`
}

func (c *Controller) Update(ctx *gin.Context) {
	var uri GetUserRequestController
	var req UpdateUserRequestController

	if err := ctx.ShouldBindUri(&uri); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	id, err := uuid.Parse(uri.ID)
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	user, err := c.svc.Update(ctx.Request.Context(), id, UpdateUserRequestService{
		Email:    req.Email,
		Username: req.Username,
		IsActive: req.IsActive,
		Plan:     req.Plan,
		IsGuest:  req.IsGuest,
	})
	if err != nil {
		if errors.Is(err, ErrUserNotFound) {
			_ = base.JSON(ctx, 404, i18n.UserNotFound, nil, nil)
			return
		}
		if errors.Is(err, ErrUserInvalidPlan) || errors.Is(err, ErrUserInvalidIsActive) || errors.Is(err, ErrUserInvalidIsGuest) {
			base.BadRequest(ctx, i18n.BadRequest, gin.H{"error": err.Error()})
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, toUserResponseController(user), i18n.UserUpdated)
}
