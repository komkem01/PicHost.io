package entitiesdto

type CreateImage struct {
	UserID    *string `json:"user_id"`
	StorageID *string `json:"storage_id"`
	IsPrivate *bool   `json:"is_private"`
}

type UpdateImage struct {
	ID        *string `json:"id"`
	UserID    *string `json:"user_id"`
	StorageID *string `json:"storage_id"`
	IsPrivate *bool   `json:"is_private"`
}

type ImageResponse struct {
	ID        string `json:"id"`
	UserID    string `json:"user_id"`
	StorageID string `json:"storage_id"`
	IsPrivate bool   `json:"is_private"`
	CreatedAt string `json:"created_at"`
}
