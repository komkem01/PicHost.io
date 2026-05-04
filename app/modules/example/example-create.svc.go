package example

import (
	"context"

	"github.com/google/uuid"
)

type Example struct {
	ID uuid.UUID `json:"id"`
}

func (s *Service) Create(ctx context.Context, userID uuid.UUID) (*Example, error) {
	example := &Example{
		ID: uuid.New(),
	}

	// Here you would typically save the example to the database
	// For now, we just return the created example

	return example, nil
}
