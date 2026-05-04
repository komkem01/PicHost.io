package ent

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AuthSessionEntity struct {
	bun.BaseModel `bun:"table:auth_sessions,alias:as"`

	ID               uuid.UUID  `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID           uuid.UUID  `bun:"user_id,type:uuid,notnull"`
	RefreshTokenHash string     `bun:"refresh_token_hash,notnull"`
	UserAgent        *string    `bun:"user_agent"`
	IPAddress        *string    `bun:"ip_address"`
	ExpiresAt        time.Time  `bun:"expires_at,notnull"`
	RevokedAt        *time.Time `bun:"revoked_at"`
	CreatedAt        time.Time  `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt        time.Time  `bun:"updated_at,notnull,default:current_timestamp"`
}
