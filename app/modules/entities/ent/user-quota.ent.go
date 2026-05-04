package ent

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type UserQuotaEntity struct {
	bun.BaseModel `bun:"table:user_quotas,alias:uq"`

	ID               uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID           uuid.UUID `bun:"user_id,type:uuid,notnull,unique"`
	UsedStorageBytes int64     `bun:"used_storage_bytes,notnull,default:0"`
	ImageCount       int       `bun:"image_count,notnull,default:0"`
	UpdatedAt        time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}
