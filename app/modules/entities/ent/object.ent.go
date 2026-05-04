package ent

import (
	"time"

	"github.com/google/uuid"
)

// ExampleStatus represents the status of an example.
type ExampleStatus string

const (
	// ExampleStatusPending indicates that the example is pending.
	ExampleStatusPending ExampleStatus = "pending"
	// ExampleStatusActive indicates that the example is active.
	ExampleStatusActive ExampleStatus = "active"
	// ExampleStatusObsolete indicates that the example is obsolete.
	ExampleStatusObsolete ExampleStatus = "obsolete"
	// ExampleStatusPurged indicates that the example has been purged.
	ExampleStatusPurged ExampleStatus = "purged"
)

// Example represents an entity for storing examples in the database.
type Example struct {
	ID        uuid.UUID     `bun:",pk,type:uuid"`
	UserID    uuid.UUID     `bun:",type:uuid,notnull"`
	Status    ExampleStatus `bun:",notnull"` // pending, active, obsolete
	CreatedAt time.Time     `bun:",default:current_timestamp"`
	UpdatedAt time.Time     `bun:",nullzero"`
	DeletedAt time.Time     `bun:",soft_delete,nullzero"`
}

func ToExampleStatus(status string) ExampleStatus {
	switch status {
	case "pending":
		return ExampleStatusPending
	case "active":
		return ExampleStatusActive
	case "obsolete":
		return ExampleStatusObsolete
	case "purged":
		return ExampleStatusPurged
	default:
		return ExampleStatusPending // Default to pending if unknown status
	}
}
