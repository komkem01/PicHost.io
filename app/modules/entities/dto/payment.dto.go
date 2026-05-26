package entitiesdto

import (
	"time"

	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
)

type CreatePaymentTransaction struct {
	UserID            uuid.UUID
	PlanKey           string
	AmountTHB         int
	Currency          string
	Status            ent.PaymentStatus
	Provider          string
	CheckoutReference string
	ProviderReference *string
	PaymentURL        *string
	ExpiresAt         time.Time
	PaidAt            *time.Time
	Metadata          map[string]any
}

type UpdatePaymentTransactionStatus struct {
	Status            ent.PaymentStatus
	ProviderReference *string
	PaidAt            *time.Time
	ReviewReason      *string
	ReviewedBy        *uuid.UUID
	ReviewedAt        *time.Time
	Metadata          map[string]any
}

type UpdatePaymentSlipStorageID struct {
	SlipStorageID string
}
