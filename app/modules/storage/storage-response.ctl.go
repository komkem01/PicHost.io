package storage

import (
	"fmt"
	"os"
	"strings"
	"time"

	"pichost.io/app/modules/entities/ent"

	"github.com/gin-gonic/gin"
)

type StorageResponseController struct {
	ID        string  `json:"id"`
	ShortCode string  `json:"short_code"`
	Provider  string  `json:"provider"`
	Path      *string `json:"path"`
	URL       *string `json:"url"`
	PublicURL *string `json:"public_url,omitempty"`
	FileSize  int64   `json:"file_size"`
	MIMEType  *string `json:"mime_type"`
	CreatedAt string  `json:"created_at"`
}

func toStorageResponse(item *ent.StorageEntity) StorageResponseController {
	return StorageResponseController{
		ID:        item.ID.String(),
		ShortCode: item.ShortCode,
		Provider:  item.Provider,
		Path:      item.Path,
		URL:       nil,
		FileSize:  item.FileSize,
		MIMEType:  item.MIMEType,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
	}
}

func publicBaseURL(ctx *gin.Context) string {
	if base := strings.TrimSpace(os.Getenv("APP_PUBLIC_BASE_URL")); base != "" {
		return strings.TrimRight(base, "/")
	}

	scheme := "http"
	if ctx.Request.TLS != nil || strings.EqualFold(ctx.GetHeader("X-Forwarded-Proto"), "https") {
		scheme = "https"
	}

	return fmt.Sprintf("%s://%s", scheme, ctx.Request.Host)
}

func buildStoragePublicURL(ctx *gin.Context, id string, shortCode string) string {
	base := publicBaseURL(ctx)
	if strings.TrimSpace(shortCode) != "" {
		return fmt.Sprintf("%s/p/%s", base, shortCode)
	}
	return fmt.Sprintf("%s/i/%s", base, id)
}
