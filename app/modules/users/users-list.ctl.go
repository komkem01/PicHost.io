package users

import (
	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
)

func (c *Controller) List(ctx *gin.Context) {
	users, err := c.svc.List(ctx.Request.Context())
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	items := make([]UserResponseController, 0, len(users))
	for _, user := range users {
		items = append(items, toUserResponseController(user))
	}

	base.Success(ctx, items, i18n.UserListed)
}
