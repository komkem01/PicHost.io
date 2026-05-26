package storage

import (
	"errors"
	"strings"

	"pichost.io/app/modules/entities/ent"
	imagemod "pichost.io/app/modules/image"
	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type openPublicURI struct {
	ID string `uri:"id" binding:"required"`
}

type openPublicCodeURI struct {
	Code string `uri:"code" binding:"required"`
}

// OpenPublic resolves a friendly public URL and redirects to a short-lived presigned URL.
func (c *Controller) OpenPublic(ctx *gin.Context) {
	var req openPublicURI
	if err := ctx.ShouldBindUri(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	if c.imgSvc != nil {
		img, imgErr := c.imgSvc.GetImage(ctx.Request.Context(), id)
		if imgErr != nil {
			if errors.Is(imgErr, imagemod.ErrImageAccountLocked) {
				_ = base.JSON(ctx, 423, "account is locked because usage exceeds plan limits", nil, nil)
				return
			}
			if errors.Is(imgErr, imagemod.ErrImageNotFound) || errors.Is(imgErr, imagemod.ErrImageExpired) {
				_ = base.JSON(ctx, 404, i18n.BadRequest, nil, nil)
				return
			}
			base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": imgErr.Error()})
			return
		}
		id = img.StorageID
	}

	item, err := c.svc.GetPresignURLByID(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrStorageNotFound) {
			_ = base.JSON(ctx, 404, i18n.BadRequest, nil, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	c.redirectToPresigned(ctx, item)
}

// OpenPublicByCode resolves a short public code and redirects to a short-lived presigned URL.
func (c *Controller) OpenPublicByCode(ctx *gin.Context) {
	var req openPublicCodeURI
	if err := ctx.ShouldBindUri(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	item, err := c.svc.GetPresignURLByShortCode(ctx.Request.Context(), code)
	if err != nil {
		if errors.Is(err, ErrStorageNotFound) {
			_ = base.JSON(ctx, 404, i18n.BadRequest, nil, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	if c.imgSvc != nil {
		if _, imgErr := c.imgSvc.GetImageByStorageID(ctx.Request.Context(), item.ID); imgErr != nil {
			if errors.Is(imgErr, imagemod.ErrImageAccountLocked) {
				_ = base.JSON(ctx, 423, "account is locked because usage exceeds plan limits", nil, nil)
				return
			}
			// If this storage is not mapped to an image record, keep existing behavior.
			if !errors.Is(imgErr, imagemod.ErrImageNotFound) {
				base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": imgErr.Error()})
				return
			}
		}
	}

	c.redirectToPresigned(ctx, item)
}

func (c *Controller) redirectToPresigned(ctx *gin.Context, item *ent.StorageEntity) {
	if item == nil || item.URL == nil || *item.URL == "" {
		_ = base.JSON(ctx, 404, i18n.BadRequest, nil, nil)
		return
	}

	ctx.Header("Cache-Control", "no-store")
	ctx.Redirect(302, *item.URL)
}
