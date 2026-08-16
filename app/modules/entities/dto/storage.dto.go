package entitiesdto

type CreateStorage struct {
	ShortCode string  `json:"short_code"`
	Provider  string  `json:"provider"`
	Path      *string `json:"path"`
	URL       *string `json:"url"`
	FileSize  int64   `json:"file_size"`
	MIMEType  *string `json:"mime_type"`
}

type UpdateStorage struct {
	ShortCode *string `json:"short_code"`
	ID        *string `json:"id"`
	Provider  *string `json:"provider"`
	Path      *string `json:"path"`
	URL       *string `json:"url"`
	FileSize  *int64  `json:"file_size"`
	MIMEType  *string `json:"mime_type"`
}

type StorageResponse struct {
	ID        string  `json:"id"`
	ShortCode string  `json:"short_code"`
	Provider  string  `json:"provider"`
	Path      *string `json:"path"`
	URL       *string `json:"url"`
	FileSize  int64   `json:"file_size"`
	MIMEType  *string `json:"mime_type"`
	CreatedAt string  `json:"created_at"`
}

type StorageStats struct {
	TotalFiles        int64            `json:"total_files"`
	TotalBytes        int64            `json:"total_bytes"`
	OrphanFiles       int64            `json:"orphan_files"`
	OrphanBytes       int64            `json:"orphan_bytes"`
	ProviderBreakdown map[string]int64 `json:"provider_breakdown"`
}

