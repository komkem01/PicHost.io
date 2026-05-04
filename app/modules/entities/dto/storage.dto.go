package entitiesdto

type CrateStorage struct {
	Provider string  `json:"provider"`
	Path     *string `json:"path"`
	URL      *string `json:"url"`
	FileSize int64   `json:"file_size"`
	MIMEType *string `json:"mime_type"`
}

type StorageResponse struct {
	ID        string  `json:"id"`
	Provider  string  `json:"provider"`
	Path      *string `json:"path"`
	URL       *string `json:"url"`
	FileSize  int64   `json:"file_size"`
	MIMEType  *string `json:"mime_type"`
	CreatedAt string  `json:"created_at"`
}
