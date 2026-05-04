package image

import (
	"time"

	"pichost.io/app/modules/entities/ent"
)

type ImageResponseController struct {
	ID        string  `json:"id"`
	UserID    *string `json:"user_id,omitempty"`
	StorageID string  `json:"storage_id"`
	IsPrivate bool    `json:"is_private"`
	ExpiresAt *string `json:"expires_at,omitempty"`
	CreatedAt string  `json:"created_at"`
}

func toImageResponse(item *ent.ImageEntity) ImageResponseController {
	var userID *string
	if item.UserID != nil {
		s := item.UserID.String()
		userID = &s
	}
	var expiresAt *string
	if item.ExpiresAt != nil {
		s := item.ExpiresAt.Format(time.RFC3339)
		expiresAt = &s
	}
	return ImageResponseController{
		ID:        item.ID.String(),
		UserID:    userID,
		StorageID: item.StorageID.String(),
		IsPrivate: item.IsPrivate,
		ExpiresAt: expiresAt,
		CreatedAt: item.CreatedAt.Format(time.RFC3339),
	}
}
