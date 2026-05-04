package image

import (
	"time"

	"pichost.io/app/modules/entities/ent"
)

type ImageResponseController struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	StorageID string `json:"storage_id"`
	IsPrivate bool   `json:"is_private"`
	CreatedAt string `json:"created_at"`
}

func toImageResponse(item *ent.ImageEntity) ImageResponseController {
	return ImageResponseController{
		ID:        item.ID.String(),
		UserID:    item.UserID.String(),
		StorageID: item.StorageID.String(),
		IsPrivate: item.IsPrivate,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
	}
}
