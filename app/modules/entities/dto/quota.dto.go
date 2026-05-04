package entitiesdto

type UpsertUserQuota struct {
	UserID string `json:"user_id"`
}

type AddToUserQuota struct {
	StorageDelta    int64 `json:"storage_delta"`
	ImageCountDelta int   `json:"image_count_delta"`
}
