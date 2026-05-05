package entitiesdto

import "github.com/google/uuid"

// CreateAuditLog is the input DTO for recording an audit event.
type CreateAuditLog struct {
	UserID       *uuid.UUID     // nil for guest / unauthenticated actions
	Action       string         // dot-notation event name, e.g. "auth.login"
	ResourceType *string        // "user" | "image" | "storage" | "session"
	ResourceID   *uuid.UUID     // UUID of the affected resource
	IPAddress    *string        // client IP
	UserAgent    *string        // HTTP User-Agent
	Metadata     map[string]any // arbitrary extra fields (serialised to JSONB)
	Status       string         // "success" | "failure"  (defaults to "success")
	ErrorCode    *string        // application error code when status == "failure"
}
