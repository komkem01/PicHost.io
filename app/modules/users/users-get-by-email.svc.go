package users

import (
	"context"
	"database/sql"
	"errors"

	"pichost.io/app/modules/entities/ent"
)

func (s *Service) GetByEmail(ctx context.Context, email string) (*ent.UserEntity, error) {
	user, err := s.user.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrUserNotFound
		}
		return nil, err
	}
	return user, nil
}
