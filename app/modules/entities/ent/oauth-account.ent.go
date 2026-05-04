package ent

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type OAuthAccountEntity struct {
	bun.BaseModel `bun:"table:oauth_accounts,alias:oa"`

	ID             uuid.UUID `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID         uuid.UUID `bun:"user_id,type:uuid,notnull"`
	Provider       string    `bun:"provider,notnull"`
	ProviderUserID string    `bun:"provider_user_id,notnull"`
	Email          *string   `bun:"email"`
	CreatedAt      time.Time `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt      time.Time `bun:"updated_at,notnull,default:current_timestamp"`
}
