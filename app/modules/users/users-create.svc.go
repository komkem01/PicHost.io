package users

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	"pichost.io/app/utils/hashing"
)

type CreateUserRequestService struct {
	Email    string
	Username string
	Password string
	Plan     string
}

func (s *Service) Create(ctx context.Context, user CreateUserRequestService) (*ent.UserEntity, error) {
	data, err := s.user.GetUserByEmail(ctx, user.Email)
	if err == nil && data != nil {
		return nil, ErrUserEmailAlreadyExists
	}
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return nil, err
	}

	plan := "Free"
	switch strings.ToLower(user.Plan) {
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
	hash, err := hashing.HashPasswordArgon2(user.Password, hashing.DefaultArgon2Params())
	if err != nil {
		return nil, err
	}
	created, err := s.user.CreateUser(ctx, entitiesdto.CreateUser{
		Email:    &user.Email,
		Password: &hash,
		Username: &user.Username,
		Plan:     plan,
		IsGuest:  false,
	})
	if err != nil {
		return nil, err
	}

	return created, nil
}
