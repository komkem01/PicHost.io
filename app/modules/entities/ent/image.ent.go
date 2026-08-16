package ent

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type ImageEntity struct {
	bun.BaseModel `bun:"table:images,alias:i"`

	ID        uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()" json:"id"`
	UserID    *uuid.UUID `bun:"user_id,type:uuid" json:"user_id,omitempty"`
	StorageID uuid.UUID  `bun:"storage_id,type:uuid,notnull" json:"storage_id"`
	IsPrivate bool       `bun:"is_private,notnull" json:"is_private"`
	ExpiresAt *time.Time `bun:"expires_at" json:"expires_at,omitempty"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	ViewCount int64      `bun:"view_count,notnull,default:0" json:"view_count"`

	Storage *StorageEntity `bun:"rel:belongs-to,join:storage_id=id" json:"storage,omitempty"`
	User    *UserEntity    `bun:"rel:belongs-to,join:user_id=id" json:"user,omitempty"`
}

