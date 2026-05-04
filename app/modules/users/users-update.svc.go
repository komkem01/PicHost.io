package users

import (
	"context"
	"strings"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
)

type UpdateUserRequestService struct {
	Email    *string
	Username *string
	IsActive *bool
	Plan     *string
	IsGuest  *bool
}

func (s *Service) Update(ctx context.Context, id uuid.UUID, req UpdateUserRequestService) (*ent.UserEntity, error) {
	if _, err := s.Get(ctx, id); err != nil {
		return nil, err
	}

	if req.IsActive == nil {
		return nil, ErrUserInvalidIsActive
	}
	if req.IsGuest == nil {
		return nil, ErrUserInvalidIsGuest
	}
	if req.Plan == nil {
		return nil, ErrUserInvalidPlan
	}

	plan := "Free"
	switch strings.ToLower(*req.Plan) {
	case "free":
		plan = "Free"
	case "basic":
		plan = "Basic"
	case "pro":
		plan = "Pro"
	case "enterprise":
		plan = "Enterprise"
	default:
		return nil, ErrUserInvalidPlan
	}

	_, err := s.user.UpdateUser(ctx, id, entitiesdto.UpdateUser{
		Email:    req.Email,
		Username: req.Username,
		IsActive: req.IsActive,
		Plan:     &plan,
		IsGuest:  req.IsGuest,
	})
	if err != nil {
		return nil, err
	}

	return s.Get(ctx, id)
}
