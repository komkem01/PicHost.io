package storage

import (
	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
)

type UploadRequestController struct {
	Provider string  `json:"provider"`
	Path     *string `json:"path"`
	URL      *string `json:"url"`
	FileSize int64   `json:"file_size" binding:"required"`
	MIMEType *string `json:"mime_type"`
}

func (c *Controller) Upload(ctx *gin.Context) {
	var req UploadRequestController
	if err := ctx.ShouldBindJSON(&req); err != nil {
		base.BadRequest(ctx, i18n.InvalidRequestForm, nil)
		return
	}

	provider := req.Provider
	if provider == "" {
		provider = "Railway"
	}

	item, err := c.svc.Upload(ctx.Request.Context(), UploadRequestService{
		Provider: provider,
		Path:     req.Path,
		URL:      req.URL,
		FileSize: req.FileSize,
		MIMEType: req.MIMEType,
	})
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	base.Success(ctx, toStorageResponse(item))
}
