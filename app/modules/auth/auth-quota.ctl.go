package auth

import (
	"errors"

	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// GetQuota handles GET /auth/quota — returns the authenticated user's plan usage and limits.
func (c *Controller) GetQuota(ctx *gin.Context) {
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

	result, err := c.svc.GetQuota(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrAuthUnauthorized) {
			base.Unauthorized(ctx, i18n.Unauthorized, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, gin.H{
		"plan":                  result.Plan,
		"used_storage_bytes":    result.UsedStorageBytes,
		"storage_limit_bytes":   result.StorageLimitBytes,
		"image_count":           result.ImageCount,
		"max_images":            result.MaxImages,
		"file_size_limit_bytes": result.FileSizeLimitBytes,
		"allow_private":         result.AllowPrivate,
	})
}
