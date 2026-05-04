package example

import (
	"mcop/app/utils"
	"mcop/app/utils/base"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jinzhu/copier"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

type CreateRequest struct {
	Filename   string `json:"filename" binding:"required,filename"`
	Visibility string `json:"visibility" default:"private"`
}

type CreateResponse struct {
	ID string `json:"id"`
}

func (c *Controller) Create(ctx *gin.Context) {
	var req CreateRequest
	var res CreateResponse
	span, log := utils.LogSpanFromGin(ctx)
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(400, gin.H{"error": "Invalid request"})
		return
	}
	span.AddEvent("example.create.request", trace.WithAttributes(
		attribute.String("filename", req.Filename),
		attribute.String("visibility", req.Visibility),
	))

	userID, err := uuid.Parse(ctx.GetString("userID"))
	if err != nil {
		userID = uuid.Nil
	}
	example, err := c.svc.Create(ctx, userID)
	if err != nil {
		log.Errf("example.create.error: %s", err)
		base.BadRequest(ctx, "example-create-failed", nil)
		return
	}

	copier.CopyWithOption(
		&res,
		example,
		copier.Option{
			IgnoreEmpty: true,
			DeepCopy:    true,
		},
	)
	base.Success(ctx, res, "example-created")
}
