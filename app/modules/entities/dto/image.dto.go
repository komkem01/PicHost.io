package entitiesdto

type CreateImage struct {
	UserID    *string `json:"user_id"`
	StorageID *string `json:"user_id"`
	IsPrivate *bool   `json:"is_private"`
}
