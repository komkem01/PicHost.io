package storage

import (
	"errors"

	"pichost.io/app/modules/entities/ent"
	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type getFileURI struct {
	ID string `uri:"id" binding:"required"`
}

type getPresignURLQuery struct {
	ID   string `form:"id"`
	Code string `form:"code"`
	URL  string `form:"url"`
}

func getCallerID(ctx *gin.Context) *uuid.UUID {
	if rawID, exists := ctx.Get("auth_user_id"); exists {
		if parsedID, ok := rawID.(uuid.UUID); ok && parsedID != uuid.Nil {
			return &parsedID
		}
	}
	return nil
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

	callerID := getCallerID(ctx)
	item, err := c.svc.GetFile(ctx.Request.Context(), id, callerID)
	if err != nil {
		if errors.Is(err, ErrStorageNotFound) {
			_ = base.JSON(ctx, 404, i18n.BadRequest, nil, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	res := toStorageResponse(item)
	publicURL := buildStoragePublicURL(ctx, res.ID, res.ShortCode)
	res.PublicURL = &publicURL

	base.Success(ctx, res)
}

func (c *Controller) GetPresignURL(ctx *gin.Context) {
	var req getPresignURLQuery
	if err := ctx.ShouldBindQuery(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	var item *ent.StorageEntity
	var err error
	callerID := getCallerID(ctx)

	if req.ID != "" {
		id, parseErr := uuid.Parse(req.ID)
		if parseErr != nil {
			base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
			return
		}

		item, err = c.svc.GetPresignURLByID(ctx.Request.Context(), id, callerID)
	} else if req.Code != "" {
		item, err = c.svc.GetPresignURLByShortCode(ctx.Request.Context(), req.Code, callerID)
	} else if req.URL != "" {
		item, err = c.svc.GetPresignURL(ctx.Request.Context(), req.URL, callerID)
	} else {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	if err != nil {
		if errors.Is(err, ErrStorageNotFound) {
			_ = base.JSON(ctx, 404, i18n.BadRequest, nil, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}
	if item == nil || item.URL == nil {
		_ = base.JSON(ctx, 404, i18n.BadRequest, nil, nil)
		return
	}

	ctx.Header("Cache-Control", "no-store")
	base.Success(ctx, gin.H{
		"url":        item.URL,
		"expires_in": int(c.svc.presignExpiry().Seconds()),
	})
}
