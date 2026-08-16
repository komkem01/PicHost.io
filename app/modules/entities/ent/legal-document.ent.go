package ent

import (
	"time"

	"github.com/uptrace/bun"
)

type LegalDocumentEntity struct {
	bun.BaseModel `bun:"table:legal_documents,alias:ld"`

	Key       string    `bun:"key,pk" json:"key"`
	Title     string    `bun:"title,notnull" json:"title"`
	Content   string    `bun:"content,notnull" json:"content"`
	UpdatedAt time.Time `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	CreatedAt time.Time `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}
