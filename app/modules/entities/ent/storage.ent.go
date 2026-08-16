package ent

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type StorageEntity struct {
	bun.BaseModel `bun:"table:storages,alias:s"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	ShortCode string    `bun:"short_code,notnull" json:"short_code"`
	Provider  string    `bun:"provider,notnull,default:'Railway'" json:"provider"`
	Path      *string   `bun:"path" json:"path,omitempty"`
	URL       *string   `bun:"url" json:"url,omitempty"`
	FileSize  int64     `bun:"file_size,notnull" json:"file_size"`
	MIMEType  *string   `bun:"mime_type" json:"mime_type,omitempty"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}
