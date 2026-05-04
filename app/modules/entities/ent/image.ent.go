package ent

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ImageEntity struct {
	bun.BaseModel `bun:"table:images,alias:i"`

	ID        uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID    uuid.UUID `bun:"user_id,type:uuid,notnull"`
	StorageID uuid.UUID `bun:"storage_id,type:uuid,notnull"`
	IsPrivate bool      `bun:"is_private,notnull,default:false"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp"`
}
