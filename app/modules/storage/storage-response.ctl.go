package storage

import (
	"time"

	"pichost.io/app/modules/entities/ent"
)

type StorageResponseController struct {
	ID        string  `json:"id"`
	Provider  string  `json:"provider"`
	Path      *string `json:"path"`
	URL       *string `json:"url"`
	FileSize  int64   `json:"file_size"`
	MIMEType  *string `json:"mime_type"`
	CreatedAt string  `json:"created_at"`
}

func toStorageResponse(item *ent.StorageEntity) StorageResponseController {
	return StorageResponseController{
		ID:        item.ID.String(),
		Provider:  item.Provider,
		Path:      item.Path,
		URL:       item.URL,
		FileSize:  item.FileSize,
		MIMEType:  item.MIMEType,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
	}
}
