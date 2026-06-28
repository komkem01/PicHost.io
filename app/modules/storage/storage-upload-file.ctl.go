package storage

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"image/jpeg"
	"image/png"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	imagemod "pichost.io/app/modules/image"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// UploadFileResponse is the combined storage + image response returned after a multipart upload.
type UploadFileResponse struct {
	ID        string  `json:"id"`
	StorageID string  `json:"storage_id"`
	IsPrivate bool    `json:"is_private"`
	ExpiresAt *string `json:"expires_at,omitempty"`
	CreatedAt string  `json:"created_at"`
	ShortCode string  `json:"short_code"`
	Provider  string  `json:"provider"`
	FileSize  int64   `json:"file_size"`
	MIMEType  *string `json:"mime_type"`
	PublicURL string  `json:"public_url"`
}

// UploadFile handles POST /storage/upload-file (multipart/form-data).
// Form fields:
//   - file      (required) – the image binary
//   - is_private (optional) – "true" or "1"
//   - provider  (optional) – defaults to "Railway"
func (c *Controller) UploadFile(ctx *gin.Context) {
	rawID, exists := ctx.Get("auth_user_id")
	if !exists {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}
	userID, ok := rawID.(uuid.UUID)
	if !ok || userID == uuid.Nil {
		base.Unauthorized(ctx, i18n.Unauthorized, nil)
		return
	}

	fh, err := ctx.FormFile("file")
	if err != nil {
		base.BadRequest(ctx, "file is required", nil)
		return
	}

	isPrivate := strings.EqualFold(ctx.PostForm("is_private"), "true") || ctx.PostForm("is_private") == "1"

	provider := strings.TrimSpace(ctx.PostForm("provider"))
	if provider == "" {
		provider = "Railway"
	}

	src, err := formFileToSource(fh)
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}
	defer src.Reader.Close()

	// Check if user is on Free plan (needs watermark)
	isFree := true
	if freeCheck, err := c.imgSvc.IsFreeUser(ctx.Request.Context(), userID); err == nil {
		isFree = freeCheck
	}

	if isFree && (src.ContentType == "image/jpeg" || src.ContentType == "image/png" || src.ContentType == "image/jpg") {
		if fileBytes, err := io.ReadAll(src.Reader); err == nil {
			if watermarkedBytes, wmErr := applyWatermark(fileBytes); wmErr == nil {
				src.Reader = io.NopCloser(bytes.NewReader(watermarkedBytes))
				src.Size = int64(len(watermarkedBytes))
			}
		}
	}

	storage, err := c.svc.UploadFromSource(ctx.Request.Context(), provider, src)
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	img, err := c.imgSvc.CreateImage(ctx.Request.Context(), imagemod.CreateImageSvcRequest{
		UserID:    userID,
		StorageID: storage.ID,
		IsPrivate: isPrivate,
	})
	if err != nil {
		switch {
		case errors.Is(err, imagemod.ErrImageFileTooLarge):
			_ = base.JSON(ctx, 413, i18n.ImageFileTooLarge, nil, nil)
		case errors.Is(err, imagemod.ErrImageStorageFull):
			_ = base.JSON(ctx, 422, i18n.ImageQuotaExceeded, nil, nil)
		case errors.Is(err, imagemod.ErrImageLimitReached):
			_ = base.JSON(ctx, 422, i18n.ImageLimitReached, nil, nil)
		case errors.Is(err, imagemod.ErrImageMIMENotAllowed):
			_ = base.JSON(ctx, 422, i18n.ImageMIMENotAllowed, nil, nil)
		case errors.Is(err, imagemod.ErrImagePrivateNotAllowed):
			_ = base.JSON(ctx, 403, i18n.ImagePrivateNotAllowed, nil, nil)
		default:
			base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		}
		return
	}

	publicURL := buildStoragePublicURL(ctx, storage.ID.String(), storage.ShortCode)

	res := UploadFileResponse{
		ID:        img.ID.String(),
		StorageID: storage.ID.String(),
		IsPrivate: img.IsPrivate,
		CreatedAt: img.CreatedAt.Format(time.RFC3339),
		ShortCode: storage.ShortCode,
		Provider:  storage.Provider,
		FileSize:  storage.FileSize,
		MIMEType:  storage.MIMEType,
		PublicURL: publicURL,
	}

	c.recordAudit("storage.upload", "success", uuidPtr(userID), strPtr("image"), uuidPtr(img.ID),
		ctx.ClientIP(), ctx.GetHeader("User-Agent"),
		map[string]any{
			"storage_id": storage.ID.String(),
			"file_size":  storage.FileSize,
			"mime_type":  storage.MIMEType,
			"is_private": isPrivate,
			"provider":   storage.Provider,
		}, nil)

	base.Success(ctx, res)
}

// UploadFileGuest handles POST /storage/upload-file-guest (no auth required).
// Images are auto-deleted after 24 hours per the guest plan.
// Form fields:
//   - file     (required) – the image binary
//   - provider (optional) – defaults to "Railway"
func (c *Controller) UploadFileGuest(ctx *gin.Context) {
	fh, err := ctx.FormFile("file")
	if err != nil {
		base.BadRequest(ctx, "file is required", nil)
		return
	}

	provider := strings.TrimSpace(ctx.PostForm("provider"))
	if provider == "" {
		provider = "Railway"
	}

	src, err := formFileToSource(fh)
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}
	defer src.Reader.Close()

	// Guests always get watermarked
	if src.ContentType == "image/jpeg" || src.ContentType == "image/png" || src.ContentType == "image/jpg" {
		if fileBytes, err := io.ReadAll(src.Reader); err == nil {
			if watermarkedBytes, wmErr := applyWatermark(fileBytes); wmErr == nil {
				src.Reader = io.NopCloser(bytes.NewReader(watermarkedBytes))
				src.Size = int64(len(watermarkedBytes))
			}
		}
	}

	storage, err := c.svc.UploadFromSource(ctx.Request.Context(), provider, src)
	if err != nil {
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	img, err := c.imgSvc.CreateImage(ctx.Request.Context(), imagemod.CreateImageSvcRequest{
		UserID:    uuid.Nil,
		StorageID: storage.ID,
		IsPrivate: false,
		IsGuest:   true,
	})
	if err != nil {
		switch {
		case errors.Is(err, imagemod.ErrImageFileTooLarge):
			_ = base.JSON(ctx, 413, i18n.ImageFileTooLarge, nil, nil)
		case errors.Is(err, imagemod.ErrImageStorageFull):
			_ = base.JSON(ctx, 422, i18n.ImageQuotaExceeded, nil, nil)
		case errors.Is(err, imagemod.ErrImageLimitReached):
			_ = base.JSON(ctx, 422, i18n.ImageLimitReached, nil, nil)
		case errors.Is(err, imagemod.ErrImageMIMENotAllowed):
			_ = base.JSON(ctx, 422, i18n.ImageMIMENotAllowed, nil, nil)
		default:
			base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		}
		return
	}

	publicURL := buildStoragePublicURL(ctx, storage.ID.String(), storage.ShortCode)

	var expiresAt *string
	if img.ExpiresAt != nil {
		t := img.ExpiresAt.Format(time.RFC3339)
		expiresAt = &t
	}

	res := UploadFileResponse{
		ID:        img.ID.String(),
		StorageID: storage.ID.String(),
		IsPrivate: false,
		ExpiresAt: expiresAt,
		CreatedAt: img.CreatedAt.Format(time.RFC3339),
		ShortCode: storage.ShortCode,
		Provider:  storage.Provider,
		FileSize:  storage.FileSize,
		MIMEType:  storage.MIMEType,
		PublicURL: publicURL,
	}

	meta := map[string]any{
		"storage_id": storage.ID.String(),
		"file_size":  storage.FileSize,
		"mime_type":  storage.MIMEType,
		"is_guest":   true,
		"provider":   storage.Provider,
	}
	if expiresAt != nil {
		meta["expires_at"] = *expiresAt
	}
	c.recordAudit("storage.upload_guest", "success", nil, strPtr("image"), uuidPtr(img.ID),
		ctx.ClientIP(), ctx.GetHeader("User-Agent"), meta, nil)

	base.Success(ctx, res)
}

func formFileToSource(fh *multipart.FileHeader) (*uploadSource, error) {
	f, err := fh.Open()
	if err != nil {
		return nil, err
	}

	contentType := fh.Header.Get("Content-Type")
	if contentType == "" {
		buf := make([]byte, 512)
		n, _ := f.Read(buf)
		contentType = http.DetectContentType(buf[:n])
		if _, err := f.Seek(0, io.SeekStart); err != nil {
			_ = f.Close()
			return nil, err
		}
	}

	return &uploadSource{
		Reader:      f,
		Size:        fh.Size,
		ContentType: contentType,
		Filename:    filepath.Base(fh.Filename),
	}, nil
}

// 8x12 proportional font dictionary styled like THSarabunPSK (Left-aligned bits)
var font8x12 = map[rune][12]byte{
	'P': {0xF8, 0x84, 0x84, 0xF8, 0x80, 0x80, 0x80, 0x80, 0x80, 0x80, 0x00, 0x00},
	'i': {0x40, 0x40, 0x00, 0xC0, 0x40, 0x40, 0x40, 0x40, 0x40, 0xE0, 0x00, 0x00},
	'c': {0x00, 0x00, 0x00, 0x78, 0x80, 0x80, 0x80, 0x80, 0x80, 0x78, 0x00, 0x00},
	'H': {0x84, 0x84, 0x84, 0xFC, 0x84, 0x84, 0x84, 0x84, 0x84, 0x84, 0x00, 0x00},
	'o': {0x00, 0x00, 0x00, 0x70, 0x88, 0x88, 0x88, 0x88, 0x88, 0x70, 0x00, 0x00},
	's': {0x00, 0x00, 0x00, 0x78, 0x80, 0x70, 0x08, 0x08, 0x88, 0x70, 0x00, 0x00},
	't': {0x40, 0x40, 0xE0, 0x40, 0x40, 0x40, 0x40, 0x40, 0x48, 0x30, 0x00, 0x00},
	'.': {0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0xC0, 0xC0, 0x00, 0x00},
}

var font8x12Width = map[rune]int{
	'P': 6,
	'i': 3,
	'c': 5,
	'H': 6,
	'o': 5,
	's': 5,
	't': 4,
	'.': 2,
}

func applyWatermark(imgData []byte) ([]byte, error) {
	srcImg, format, err := image.Decode(bytes.NewReader(imgData))
	if err != nil {
		return nil, err
	}

	bounds := srcImg.Bounds()
	w, h := bounds.Dx(), bounds.Dy()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, srcImg, bounds.Min, draw.Src)

	// Keep it very small, clean, and proportional
	scale := 1
	if w >= 2000 {
		scale = 2
	}

	text := "PicHost.io"
	wmWidth := 0
	for _, char := range text {
		wmWidth += (font8x12Width[char] + 1) * scale
	}
	wmWidth -= 1 * scale // remove trailing gap
	wmHeight := 12 * scale

	// Position: Bottom Right (offset 16px from corner)
	xStart := w - wmWidth - 16
	yStart := h - wmHeight - 16

	if xStart > 0 && yStart > 0 {
		// Draw text function
		drawText := func(t string, startX, startY int, clr color.Color) {
			currX := startX
			for _, char := range t {
				bitmap := font8x12[char]
				width := font8x12Width[char]
				for r := 0; r < 12; r++ {
					for c := 0; c < width; c++ {
						if (bitmap[r] & (1 << (7 - c))) != 0 {
							for dy := 0; dy < scale; dy++ {
								for dx := 0; dx < scale; dx++ {
									rgba.Set(currX+c*scale+dx, startY+r*scale+dy, clr)
								}
							}
						}
					}
				}
				currX += (width + 1) * scale
			}
		}

		// Colors: Subtle semi-transparent white and black drop shadow
		shadowClr := color.RGBA{0, 0, 0, 130}
		whiteClr := color.RGBA{255, 255, 255, 130} // 50% opacity to be unobtrusive

		// Draw shadow
		drawText(text, xStart+1, yStart+1, shadowClr)

		// Draw main text
		drawText(text, xStart, yStart, whiteClr)
	}

	var buf bytes.Buffer
	if format == "png" {
		err = png.Encode(&buf, rgba)
	} else {
		err = jpeg.Encode(&buf, rgba, &jpeg.Options{Quality: 90})
	}
	if err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

