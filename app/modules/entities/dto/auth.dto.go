package entitiesdto

import (
	"time"

	"github.com/google/uuid"
)

type CreateAuthSession struct {
	UserID           uuid.UUID
	RefreshTokenHash string
	UserAgent        *string
	IPAddress        *string
	ExpiresAt        time.Time
}

type RotateAuthSession struct {
	RefreshTokenHash string
	ExpiresAt        time.Time
}

type CreateOAuthAccount struct {
	UserID         uuid.UUID
	Provider       string
	ProviderUserID string
	Email          *string
}
