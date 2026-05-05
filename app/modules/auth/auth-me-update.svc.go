package auth

import (
	"context"
	"strings"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
)

// UpdateMe updates the authenticated user's own username and/or email.
// Privileged fields (plan, is_active, is_guest) are preserved from the existing record.
func (s *Service) UpdateMe(ctx context.Context, userID uuid.UUID, username *string, email *string) (*ent.UserEntity, error) {
	current, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrAuthUnauthorized
	}

	mergedEmail := current.Email
	if email != nil {
		v := strings.TrimSpace(*email)
		mergedEmail = &v
	}
	mergedUsername := current.Username
	if username != nil {
		v := strings.TrimSpace(*username)
		mergedUsername = &v
	}

	isActive := current.IsActive
	isGuest := current.IsGuest
	plan := string(current.Plan)

	updated, err := s.user.UpdateUser(ctx, userID, entitiesdto.UpdateUser{
		Email:    mergedEmail,
		Username: mergedUsername,
		IsActive: &isActive,
		IsGuest:  &isGuest,
		Plan:     &plan,
	})
	if err != nil {
		return nil, err
	}
	return updated, nil
}

// ChangePassword verifies the current password and replaces it with a new hashed password.
func (s *Service) ChangePassword(ctx context.Context, userID uuid.UUID, currentPwd string, newPwd string) error {
	user, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		return ErrAuthInvalidCredentials
	}

	if user.Password == nil || !verifyPassword(*user.Password, currentPwd) {
		return ErrAuthInvalidCredentials
	}

	hash, err := hashPassword(newPwd)
	if err != nil {
		return err
	}

	_, err = s.user.UpdateUserPassword(ctx, userID, entitiesdto.UpdateUserPassword{
		NewPassword: hash,
	})
	return err
}
