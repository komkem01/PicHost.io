package entitiesdto

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type CreateNotification struct {
	UserID     *uuid.UUID      `json:"user_id,omitempty"`
	TargetRole string          `json:"target_role,omitempty"` // 'user', 'admin', 'all'
	Type       string          `json:"type"`                  // 'moderation', 'payment', 'security', 'storage', 'system', 'announcement'
	Title      string          `json:"title"`
	Message    string          `json:"message"`
	Link       *string         `json:"link,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
}

type NotificationResponse struct {
	ID         uuid.UUID       `json:"id"`
	UserID     *uuid.UUID      `json:"user_id,omitempty"`
	TargetRole string          `json:"target_role"`
	Type       string          `json:"type"`
	Title      string          `json:"title"`
	Message    string          `json:"message"`
	Link       *string         `json:"link,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
	IsRead     bool            `json:"is_read"`
	ReadAt     *time.Time      `json:"read_at,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

type BroadcastNotificationRequest struct {
	TargetRole string          `json:"target_role"`           // 'all', 'user', 'admin', or specific plan like 'plan:basic'
	UserID     *uuid.UUID      `json:"user_id,omitempty"`     // optional if sending to specific user
	Type       string          `json:"type"`                  // 'announcement', 'system', 'warning', 'info'
	Title      string          `json:"title" binding:"required"`
	Message    string          `json:"message" binding:"required"`
	Link       *string         `json:"link,omitempty"`
	Metadata   json.RawMessage `json:"metadata,omitempty"`
}

type NotificationFilter struct {
	UserID     *uuid.UUID
	TargetRole string
	IsAdmin    bool
	UnreadOnly bool
	Type       string
	Page       int
	Limit      int
}
