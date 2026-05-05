package auth

import (
	"errors"

	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type UpdateMeRequestController struct {
	Username *string `json:"username"`
	Email    *string `json:"email"`
}

// UpdateMe handles PATCH /auth/me — updates username and/or email for the logged-in user.
func (c *Controller) UpdateMe(ctx *gin.Context) {
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

	var req UpdateMeRequestController
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	user, err := c.svc.UpdateMe(ctx.Request.Context(), userID, req.Username, req.Email)
	if err != nil {
		if errors.Is(err, ErrAuthUnauthorized) {
			base.Unauthorized(ctx, i18n.Unauthorized, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, gin.H{
		"id":       user.ID.String(),
		"username": user.Username,
		"email":    user.Email,
	})
}

type ChangePasswordRequestController struct {
	CurrentPassword string `json:"current_password" binding:"required"`
	NewPassword     string `json:"new_password" binding:"required,min=8"`
}

// ChangePassword handles PATCH /auth/change-password — changes the authenticated user's password.
func (c *Controller) ChangePassword(ctx *gin.Context) {
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

	var req ChangePasswordRequestController
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, gin.H{"error": err.Error()})
		return
	}

	if err := c.svc.ChangePassword(ctx.Request.Context(), userID, req.CurrentPassword, req.NewPassword); err != nil {
		if errors.Is(err, ErrAuthInvalidCredentials) {
			base.BadRequest(ctx, "Current password is incorrect", gin.H{"error": "current_password_incorrect"})
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, gin.H{"message": "Password changed successfully"})
}
