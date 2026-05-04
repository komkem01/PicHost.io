package id

import (
	"github.com/google/uuid"
)

func NewUUID() uuid.UUID {
	return uuid.Must(uuid.NewV7())
}
