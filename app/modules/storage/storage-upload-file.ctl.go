package storage

import (
	"errors"
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
