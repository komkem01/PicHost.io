package auth

import (
	"strings"

	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminMiddleware requires a valid Bearer token AND is_admin = true on the user row.
// Must be chained AFTER AuthMiddleware (re-reads auth_user_id set by it).
func (c *Controller) AdminMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// --- 1. Parse Bearer token (same as AuthMiddleware) ---
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

		// --- 2. Load user and verify is_admin ---
		user, err := c.svc.user.GetUserByID(ctx.Request.Context(), userID)
		if err != nil || !user.IsAdmin {
			base.Forbidden(ctx, "admin access required", nil)
			ctx.Abort()
			return
		}

		ctx.Set("auth_user_id", userID)
		ctx.Set("auth_user_is_admin", true)
		ctx.Next()
	}
}

// requireAdmin extracts auth_user_id from context inside an admin handler.
func requireAdmin(ctx *gin.Context) (uuid.UUID, bool) {
	raw, exists := ctx.Get("auth_user_id")
	if !exists {
		return uuid.Nil, false
	}
	id, ok := raw.(uuid.UUID)
	return id, ok && id != uuid.Nil
}
