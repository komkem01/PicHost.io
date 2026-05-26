package image

import (
	"errors"
	"strings"

	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type CreateImageBody struct {
	StorageID string `json:"storage_id" binding:"required"`
	IsPrivate bool   `json:"is_private"`
}

func (c *Controller) CreateImage(ctx *gin.Context) {
	var req CreateImageBody
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	storageID, err := uuid.Parse(req.StorageID)
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	// Determine user identity (guest or authenticated).
	var userID uuid.UUID
	isGuest := true
	if rawID, exists := ctx.Get("auth_user_id"); exists {
		if uid, ok := rawID.(uuid.UUID); ok && uid != uuid.Nil {
			userID = uid
			isGuest = false
		}
	}

	// For private images, authentication is required.
	if req.IsPrivate && isGuest {
		_ = base.JSON(ctx, 403, i18n.Forbidden, nil, nil)
		return
	}

	item, err := c.svc.CreateImage(ctx.Request.Context(), CreateImageSvcRequest{
		UserID:    userID,
		StorageID: storageID,
		IsPrivate: req.IsPrivate,
		IsGuest:   isGuest,
	})
	if err != nil {
		switch {
		case errors.Is(err, ErrImageAccountLocked):
			_ = base.JSON(ctx, 423, "account is locked because usage exceeds plan limits", nil, nil)
		case errors.Is(err, ErrImageStorageNotFound):
			_ = base.JSON(ctx, 404, i18n.ImageNotFound, nil, nil)
		case errors.Is(err, ErrImageFileTooLarge):
			_ = base.JSON(ctx, 413, i18n.ImageFileTooLarge, nil, nil)
		case errors.Is(err, ErrImageStorageFull):
			_ = base.JSON(ctx, 422, i18n.ImageQuotaExceeded, nil, nil)
		case errors.Is(err, ErrImageLimitReached):
			_ = base.JSON(ctx, 422, i18n.ImageLimitReached, nil, nil)
		case errors.Is(err, ErrImageMIMENotAllowed):
			_ = base.JSON(ctx, 422, i18n.ImageMIMENotAllowed, nil, nil)
		case errors.Is(err, ErrImagePrivateNotAllowed):
			_ = base.JSON(ctx, 403, i18n.ImagePrivateNotAllowed, nil, nil)
		default:
			base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		}
		return
	}

	base.Success(ctx, toImageResponse(item), i18n.ImageCreated)
}

// imageIDFromPath reads :id from URL and parses it.
func imageIDFromPath(ctx *gin.Context) (uuid.UUID, bool) {
	raw := strings.TrimSpace(ctx.Param("id"))
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.Nil, false
	}
	return id, true
}
