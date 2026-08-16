package ent

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type NotificationEntity struct {
	bun.BaseModel `bun:"table:notifications,alias:n"`

	ID         uuid.UUID       `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID     *uuid.UUID      `bun:"user_id,type:uuid"`
	TargetRole string          `bun:"target_role,notnull,default:'user'"`
	Type       string          `bun:"type,notnull"`
	Title      string          `bun:"title,notnull"`
	Message    string          `bun:"message,notnull"`
	Link       *string         `bun:"link"`
	Metadata   json.RawMessage `bun:"metadata,type:jsonb"`
	IsRead     bool            `bun:"is_read,notnull,default:false"`
	ReadAt     *time.Time      `bun:"read_at"`
	CreatedAt  time.Time       `bun:"created_at,notnull,default:current_timestamp"`

	// Relational field if needed
	User *UserEntity `bun:"rel:belongs-to,join:user_id=id"`
}
