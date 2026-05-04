package storage

import (
	"errors"

	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type getFileURI struct {
	ID string `uri:"id" binding:"required"`
}

type getPresignURLQuery struct {
	URL string `form:"url" binding:"required"`
}

func (c *Controller) GetFile(ctx *gin.Context) {
	var req getFileURI
	if err := ctx.ShouldBindUri(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	item, err := c.svc.GetFile(ctx.Request.Context(), id)
	if err != nil {
		if errors.Is(err, ErrStorageNotFound) {
			_ = base.JSON(ctx, 404, i18n.BadRequest, nil, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, toStorageResponse(item))
}

func (c *Controller) GetPresignURL(ctx *gin.Context) {
	var req getPresignURLQuery
	if err := ctx.ShouldBindQuery(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	item, err := c.svc.GetPresignURL(ctx.Request.Context(), req.URL)
	if err != nil {
		if errors.Is(err, ErrStorageNotFound) {
			_ = base.JSON(ctx, 404, i18n.BadRequest, nil, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, gin.H{"url": item.URL})
}
