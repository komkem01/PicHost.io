package users

import (
	"github.com/gin-gonic/gin"
	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"
)

type CreateUserRequestController struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required,min=8"`
	Username string `json:"username" binding:"required"`
	Plan     string `json:"plan" binding:"required,oneof=free basic pro enterprise"`
}

type CreateUserResponseController struct {
	ID       string `json:"id"`
	Email    string `json:"email"`
	Username string `json:"username"`
	Plan     string `json:"plan"`
	IsActive bool   `json:"is_active"`
	IsGuest  bool   `json:"is_guest"`
}

func (c *Controller) Create(ctx *gin.Context) {
	var req CreateUserRequestController
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	_, err := c.svc.Create(ctx.Request.Context(), CreateUserRequestService{
		Email:    req.Email,
		Username: req.Username,
		Password: req.Password,
		Plan:     req.Plan,
	})
	if err != nil {
		if err == ErrUserEmailAlreadyExists || err == ErrUserInvalidPlan {
			base.BadRequest(ctx, i18n.BadRequest, gin.H{"error": err.Error()})
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, nil)
}
