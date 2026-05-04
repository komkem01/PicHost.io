package example

import (
	"mcop/app/modules/net/httpx"
	"mcop/app/utils"
	"mcop/app/utils/base"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

func (c *Controller) GetHttpReq(ctx *gin.Context) {
	span, log := utils.LogSpanFromGin(ctx)

	span.AddEvent("example.get.http.request")

	log.Infof("Received HTTP request for example.get")

	req, err := httpx.NewRequest(ctx.Request.Context(), "GET", "https://httpbin.org/get", nil)
	if err != nil {
		log.Errf("example.get.http.error: %s", err)
		base.BadRequest(ctx, "http-request-error", nil)
		return
	}
	resp, err := c.cli.Do(req)
	if err != nil {
		log.Errf("example.get.http.error: %s", err)
		base.InternalServerError(ctx, "http-request-failed", nil)
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Errf("example.get.http.error: unexpected status code %d", resp.StatusCode)
		base.InternalServerError(ctx, "unexpected-status-code", nil)
		return
	}
	log.Infof("Successfully processed example.get HTTP request")
	span.AddEvent("example.get.http.response", trace.WithAttributes(
		attribute.Int("status_code", resp.StatusCode),
	))

	ctx.DataFromReader(200, resp.ContentLength, req.Header.Get("Content-Type"), resp.Body, nil)
}
