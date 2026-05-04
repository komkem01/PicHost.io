package auth

import (
	"strings"

	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
)

func (c *Controller) AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := strings.TrimSpace(ctx.GetHeader("Authorization"))
		if authHeader == "" {
			base.Unauthorized(ctx, i18n.Unauthorized, nil)
			ctx.Abort()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			base.Unauthorized(ctx, i18n.Unauthorized, nil)
			ctx.Abort()
			return
		}

		userID, err := c.svc.ParseAccessToken(parts[1])
		if err != nil {
			base.Unauthorized(ctx, i18n.Unauthorized, nil)
			ctx.Abort()
			return
		}

		ctx.Set("auth_user_id", userID)
		ctx.Next()
	}
}

// OptionalAuthMiddleware sets auth_user_id when a valid bearer token is provided.
// If Authorization header is missing, request proceeds as guest.
func (c *Controller) OptionalAuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authHeader := strings.TrimSpace(ctx.GetHeader("Authorization"))
		if authHeader == "" {
			ctx.Next()
			return
		}

		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			base.Unauthorized(ctx, i18n.Unauthorized, nil)
			ctx.Abort()
			return
		}

		userID, err := c.svc.ParseAccessToken(parts[1])
		if err != nil {
			base.Unauthorized(ctx, i18n.Unauthorized, nil)
			ctx.Abort()
			return
		}

		ctx.Set("auth_user_id", userID)
		ctx.Next()
	}
}
