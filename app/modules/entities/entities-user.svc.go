package entities

import (
	"context"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"

	"github.com/google/uuid"
)

var _ entitiesinf.UserEntity = (*Service)(nil)

func (s *Service) CreateUser(ctx context.Context, user entitiesdto.CreateUser) (*ent.UserEntity, error) {
	now := time.Now()
	data := &ent.UserEntity{
		Email:     user.Email,
		Password:  user.Password,
		Username:  user.Username,
		Plan:      ent.PlanType(user.Plan),
		IsActive:  true,
		IsGuest:   user.IsGuest,
		CreatedAt: now,
		UpdatedAt: now,
	}
	_, err := s.db.NewInsert().
		Model(data).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) GetUserByID(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	var user ent.UserEntity
	err := s.db.NewSelect().
		Model(&user).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) GetListUser(ctx context.Context) ([]*ent.UserEntity, error) {
	var users []*ent.UserEntity
	err := s.db.NewSelect().
		Model(&users).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return users, nil
}

func (s *Service) GetUserByEmail(ctx context.Context, email string) (*ent.UserEntity, error) {
	var user ent.UserEntity
	err := s.db.NewSelect().
		Model(&user).
		Where("email = ?", email).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) UpdateUser(ctx context.Context, id uuid.UUID, user entitiesdto.UpdateUser) (*ent.UserEntity, error) {
	now := time.Now()
	data := &ent.UserEntity{
		Email:     user.Email,
		Username:  user.Username,
		IsActive:  *user.IsActive,
		Plan:      ent.PlanType(*user.Plan),
		IsGuest:   *user.IsGuest,
		UpdatedAt: now,
	}
	_, err := s.db.NewUpdate().
		Model(data).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) UpdateUserPlan(ctx context.Context, id uuid.UUID, plan entitiesdto.UpdateUserPlan) (*ent.UserEntity, error) {
	now := time.Now()
	data := &ent.UserEntity{
		Plan:      ent.PlanType(*plan.Plan),
		UpdatedAt: now,
	}
	_, err := s.db.NewUpdate().
		Model(data).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) UpdateUserPassword(ctx context.Context, id uuid.UUID, password entitiesdto.UpdateUserPassword) (*ent.UserEntity, error) {
	now := time.Now()
	data := &ent.UserEntity{
		Password:  &password.NewPassword,
		UpdatedAt: now,
	}
	_, err := s.db.NewUpdate().
		Model(data).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.NewDelete().
		Model((*ent.UserEntity)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}
