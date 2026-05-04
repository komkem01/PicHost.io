package ent

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type StorageEntity struct {
	bun.BaseModel `bun:"table:storages,alias:s"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	Provider  string    `bun:"provider,notnull,default:'Railway'"`
	Path      *string   `bun:"path"`
	URL       *string   `bun:"url"`
	FileSize  int64     `bun:"file_size,notnull"`
	MIMEType  *string   `bun:"mime_type"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
}
