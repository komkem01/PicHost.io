package ent

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type AuditLogEntity struct {
	bun.BaseModel `bun:"table:audit_logs,alias:al"`

	ID           uuid.UUID       `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID       *uuid.UUID      `bun:"user_id,type:uuid"`
	Action       string          `bun:"action,notnull"`
	ResourceType *string         `bun:"resource_type"`
	ResourceID   *uuid.UUID      `bun:"resource_id,type:uuid"`
	IPAddress    *string         `bun:"ip_address"`
	UserAgent    *string         `bun:"user_agent"`
	Metadata     json.RawMessage `bun:"metadata,type:jsonb"`
	Status       string          `bun:"status,notnull,default:'success'"`
	ErrorCode    *string         `bun:"error_code"`
	CreatedAt    time.Time       `bun:"created_at,notnull,default:current_timestamp"`
}
