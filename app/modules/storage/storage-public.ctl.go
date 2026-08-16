package storage

import (
	"errors"
	"strings"

	"pichost.io/app/modules/entities/ent"
	imagemod "pichost.io/app/modules/image"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type openPublicURI struct {
	ID string `uri:"id" binding:"required"`
}

type openPublicCodeURI struct {
	Code string `uri:"code" binding:"required"`
}

func (c *Controller) render404(ctx *gin.Context) {
	html := `<!DOCTYPE html>
<html lang="th">
<head>
  <meta charset="UTF-8">
  <meta name="viewport" content="width=device-width, initial-scale=1.0">
  <title>404 — ไม่พบรูปภาพ | PicHost.io</title>
  <script src="https://cdn.tailwindcss.com"></script>
  <link rel="preconnect" href="https://fonts.googleapis.com">
  <link rel="preconnect" href="https://fonts.gstatic.com" crossorigin>
  <link href="https://fonts.googleapis.com/css2?family=Plus+Jakarta+Sans:wght@400;500;600;700&display=swap" rel="stylesheet">
  <style>body { font-family: "Plus Jakarta Sans", sans-serif; }</style>
</head>
<body class="bg-[#FAFAFA] text-zinc-900 min-h-screen flex items-center justify-center p-4">
  <div class="max-w-md w-full bg-white border border-zinc-200 rounded-3xl p-8 text-center shadow-sm space-y-5">
    <div class="w-16 h-16 bg-red-50 border border-red-200/80 rounded-2xl flex items-center justify-center mx-auto text-red-600">
      <svg class="w-8 h-8" fill="none" stroke="currentColor" stroke-width="2" viewBox="0 0 24 24">
        <path stroke-linecap="round" stroke-linejoin="round" d="M12 9v3.75m9-.75a9 9 0 11-18 0 9 9 0 0118 0zm-9 3.75h.008v.008H12v-.008z" />
      </svg>
    </div>
    <div class="space-y-1.5">
      <h1 class="text-xl font-bold text-zinc-900">ไม่พบรูปภาพนี้ / Image Not Found</h1>
      <p class="text-zinc-500 text-sm leading-relaxed">รูปภาพนี้ถูกลบออกไปแล้ว หรือลิงก์ไม่ถูกต้อง<br><span class="text-xs text-zinc-400">This image has been deleted or the link is invalid.</span></p>
    </div>
    <div class="pt-2">
      <a href="/" class="inline-flex items-center justify-center px-5 py-2.5 rounded-xl bg-zinc-900 hover:bg-zinc-800 text-white text-sm font-semibold transition-all shadow-xs">
        กลับสู่หน้าหลัก / Go to Homepage
      </a>
    </div>
  </div>
</body>
</html>`
	ctx.Header("Content-Type", "text/html; charset=utf-8")
	ctx.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	ctx.String(404, html)
}

// OpenPublic resolves a friendly public URL and redirects to a short-lived presigned URL.
func (c *Controller) OpenPublic(ctx *gin.Context) {
	var req openPublicURI
	if err := ctx.ShouldBindUri(&req); err != nil {
		c.render404(ctx)
		return
	}

	id, err := uuid.Parse(req.ID)
	if err != nil {
		c.render404(ctx)
		return
	}

	if c.imgSvc != nil {
		img, imgErr := c.imgSvc.GetImage(ctx.Request.Context(), id)
		if imgErr != nil {
			if errors.Is(imgErr, imagemod.ErrImageAccountLocked) {
				c.render404(ctx)
				return
			}
			c.render404(ctx)
			return
		}
		id = img.StorageID
	}

	item, err := c.svc.GetPresignURLByID(ctx.Request.Context(), id)
	if err != nil {
		c.render404(ctx)
		return
	}

	c.redirectToPresigned(ctx, item)
}

// OpenPublicByCode resolves a short public code and redirects to a short-lived presigned URL.
func (c *Controller) OpenPublicByCode(ctx *gin.Context) {
	var req openPublicCodeURI
	if err := ctx.ShouldBindUri(&req); err != nil {
		c.render404(ctx)
		return
	}

	code := strings.TrimSpace(req.Code)
	if code == "" {
		c.render404(ctx)
		return
	}

	item, err := c.svc.GetPresignURLByShortCode(ctx.Request.Context(), code)
	if err != nil {
		c.render404(ctx)
		return
	}

	if c.imgSvc != nil {
		if _, imgErr := c.imgSvc.GetImageByStorageID(ctx.Request.Context(), item.ID); imgErr != nil {
			if errors.Is(imgErr, imagemod.ErrImageAccountLocked) {
				c.render404(ctx)
				return
			}
			if errors.Is(imgErr, imagemod.ErrImageNotFound) {
				c.render404(ctx)
				return
			}
		}
	}

	c.redirectToPresigned(ctx, item)
}

func (c *Controller) redirectToPresigned(ctx *gin.Context, item *ent.StorageEntity) {
	if item == nil || item.URL == nil || *item.URL == "" {
		c.render404(ctx)
		return
	}

	ctx.Header("Cache-Control", "no-store, no-cache, must-revalidate")
	ctx.Redirect(302, *item.URL)
}
