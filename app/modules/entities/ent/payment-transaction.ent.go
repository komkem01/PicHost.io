package ent

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type PaymentStatus string

const (
	PaymentStatusPending   PaymentStatus = "pending"
	PaymentStatusPaid      PaymentStatus = "paid"
	PaymentStatusFailed    PaymentStatus = "failed"
	PaymentStatusCancelled PaymentStatus = "cancelled"
	PaymentStatusExpired   PaymentStatus = "expired"
)

type PaymentTransactionEntity struct {
	bun.BaseModel `bun:"table:payment_transactions,alias:pt"`

	ID                uuid.UUID      `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	UserID            uuid.UUID      `bun:"user_id,type:uuid,notnull"`
	PlanKey           string         `bun:"plan_key,notnull"`
	AmountTHB         int            `bun:"amount_thb,notnull"`
	Currency          string         `bun:"currency,notnull"`
	Status            PaymentStatus  `bun:"status,type:payment_status,notnull,default:'pending'"`
	Provider          string         `bun:"provider,notnull"`
	CheckoutReference string         `bun:"checkout_reference,notnull"`
	ProviderReference *string        `bun:"provider_reference"`
	PaymentURL        *string        `bun:"payment_url"`
	SlipStorageID     *string        `bun:"slip_storage_id"`
	ReviewReason      *string        `bun:"review_reason"`
	ReviewedBy        *uuid.UUID     `bun:"reviewed_by,type:uuid"`
	ReviewedAt        *time.Time     `bun:"reviewed_at"`
	ExpiresAt         time.Time      `bun:"expires_at,notnull"`
	PaidAt            *time.Time     `bun:"paid_at"`
	Metadata          map[string]any `bun:"metadata,type:jsonb,notnull"`
	CreatedAt         time.Time      `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt         time.Time      `bun:"updated_at,notnull,default:current_timestamp"`
}
