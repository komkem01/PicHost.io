package image

import (
	"errors"

	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GetImageURI struct {
	ID string `uri:"id" binding:"required"`
}

type GetPresignURLQuery struct {
	ImageID string `form:"image_id" binding:"required"`
}

func (c *Controller) GetImage(ctx *gin.Context) {
	var req GetImageURI
	if err := ctx.ShouldBindUri(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	item, err := c.svc.GetImage(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrImageNotFound) || errors.Is(err, ErrImageExpired) {
			_ = base.JSON(ctx, 404, i18n.ImageNotFound, nil, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, toImageResponse(item), i18n.ImageFetched)
}

func (c *Controller) GetPresignURL(ctx *gin.Context) {
	var req GetPresignURLQuery
	if err := ctx.ShouldBindQuery(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	imageID, err := uuid.Parse(req.ImageID)
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	url, err := c.svc.GetPresignURL(ctx.Request.Context(), imageID)
	if err != nil {
		if errors.Is(err, ErrImageNotFound) {
			_ = base.JSON(ctx, 404, i18n.ImageNotFound, nil, nil)
			return
		}
		if errors.Is(err, ErrImageURLNotFound) {
			_ = base.JSON(ctx, 404, i18n.ImageURLNotFound, nil, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, gin.H{"url": url}, i18n.ImagePresignURLFetched)
}
