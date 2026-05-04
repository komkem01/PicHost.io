package example

import (
	"mcop/app/utils"
	"mcop/app/utils/base"
	"mcop/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/oklog/ulid/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type GetRequest struct {
	ID string `uri:"id" binding:"required"`
}

type GetResponse struct {
	ID ulid.ULID `json:"upload_id"`
}

func (c *Controller) Get(ctx *gin.Context) {
	var req GetRequest
	var res GetResponse
	span, log := utils.LogSpanFromGin(ctx)

	if err := ctx.ShouldBindUri(&req); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	span.AddEvent("example.get.request", trace.WithAttributes(
		attribute.String("id", req.ID),
	))

	id, err := ulid.Parse(req.ID)
	if err != nil {
		log.Errf("example.get.error: %s", err)
		base.BadRequest(ctx, "invalid-id", nil)
		return
	}
	res.ID = id

	base.Success(ctx, res, i18n.ExampleMessageOK)
}
