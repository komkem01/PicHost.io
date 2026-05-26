package image

import (
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"pichost.io/app/modules/entities/ent"
	"pichost.io/app/utils/base"
	"pichost.io/config/i18n"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type ImageWithStorageResponse struct {
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

func (c *Controller) ListImages(ctx *gin.Context) {
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

	items, err := c.svc.ListImages(ctx.Request.Context(), userID)
	if err != nil {
		if errors.Is(err, ErrImageAccountLocked) {
			_ = base.JSON(ctx, 423, "account is locked because usage exceeds plan limits", nil, nil)
			return
		}
		base.InternalServerError(ctx, i18n.InternalError, gin.H{"error": err.Error()})
		return
	}

	res := make([]ImageWithStorageResponse, 0, len(items))
	for _, item := range items {
		res = append(res, toImageWithStorageResponse(ctx, item.Image, item.Storage))
	}

	base.Success(ctx, res)
}

func toImageWithStorageResponse(ctx *gin.Context, img *ent.ImageEntity, s *ent.StorageEntity) ImageWithStorageResponse {
	var expiresAt *string
	if img.ExpiresAt != nil {
		t := img.ExpiresAt.Format(time.RFC3339)
		expiresAt = &t
	}
	return ImageWithStorageResponse{
		ID:        img.ID.String(),
		StorageID: img.StorageID.String(),
		IsPrivate: img.IsPrivate,
		ExpiresAt: expiresAt,
		CreatedAt: img.CreatedAt.Format(time.RFC3339),
		ShortCode: s.ShortCode,
		Provider:  s.Provider,
		FileSize:  s.FileSize,
		MIMEType:  s.MIMEType,
		PublicURL: buildListPublicURL(ctx, img.ID.String(), s.ShortCode),
	}
}

func buildListPublicURL(ctx *gin.Context, id string, shortCode string) string {
	base := listPublicBaseURL(ctx)
	if strings.TrimSpace(shortCode) != "" {
		return fmt.Sprintf("%s/p/%s", base, shortCode)
	}
	return fmt.Sprintf("%s/i/%s", base, id)
}

func listPublicBaseURL(ctx *gin.Context) string {
	if b := strings.TrimSpace(os.Getenv("APP_PUBLIC_BASE_URL")); b != "" {
		return strings.TrimRight(b, "/")
	}
	scheme := "http"
	if ctx.Request.TLS != nil || strings.EqualFold(ctx.GetHeader("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}
	return fmt.Sprintf("%s://%s", scheme, ctx.Request.Host)
}
