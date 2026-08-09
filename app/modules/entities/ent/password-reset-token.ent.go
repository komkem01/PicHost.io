package ent

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PasswordResetTokenEntity struct {
	bun.BaseModel `bun:"table:password_reset_tokens,alias:prt"`

	ID        uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID    uuid.UUID  `bun:"user_id,type:uuid,notnull"`
	TokenHash string     `bun:"token_hash,notnull,unique"`
	ExpiresAt time.Time  `bun:"expires_at,notnull"`
	UsedAt    *time.Time `bun:"used_at"`
	CreatedAt time.Time  `bun:"created_at,notnull,default:current_timestamp"`

	User *UserEntity `bun:"rel:belongs-to,join:user_id=id"`
}
