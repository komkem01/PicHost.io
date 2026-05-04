package storage

import (
	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
)

func (c *Controller) ListFiles(ctx *gin.Context) {
	items, err := c.svc.ListFiles(ctx.Request.Context())
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	res := make([]StorageResponseController, 0, len(items))
	for _, item := range items {
		res = append(res, toStorageResponse(item))
	}

	base.Success(ctx, res)
}
