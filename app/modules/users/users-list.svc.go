package users

import (
	"context"

	"pichost.io/app/modules/entities/ent"
)

func (s *Service) List(ctx context.Context) ([]*ent.UserEntity, error) {
	return s.user.GetListUser(ctx)
}
