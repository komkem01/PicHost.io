package auth

import (
	"context"
	"database/sql"
	"errors"
	"strings"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgconn"
)

// normalizeEmailForCompare lets us tell a "real" email change apart from a
// no-op resubmission that only differs in case/whitespace.
func normalizeEmailForCompare(email *string) string {
	if email == nil {
		return ""
	}
	return strings.ToLower(strings.TrimSpace(*email))
}

// isEmailUniqueConflict reports whether err is a Postgres unique-violation on
// the users.email column, mirroring storage's isShortCodeUniqueConflict.
func isEmailUniqueConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	if pgErr.Code != "23505" {
		return false
	}
	constraint := strings.ToLower(pgErr.ConstraintName)
	message := strings.ToLower(pgErr.Message)
	detail := strings.ToLower(pgErr.Detail)
	return strings.Contains(constraint, "email") || strings.Contains(message, "email") || strings.Contains(detail, "email")
}

// UpdateMe updates the authenticated user's own username and/or email.
// Privileged fields (plan, is_active, is_guest) are preserved from the existing record.
//
// Emails are compared case-insensitively (and normalized to lowercase for
// storage/lookup — see normalizeEmailForCompare and entities.normalizeEmailForStorage):
// a username-only change (or an email resubmitted with different case/whitespace
// but the same address) preserves email_verified_at. An actual change to a
// different address clears email_verified_at and sends a fresh verification
// email to the new address — so email_verified_at can always be trusted to
// mean "this exact address was verified".
//
// The email change and the email_verified_at reset are applied in a single
// UPDATE (entitiesdto.UpdateUser.ClearEmailVerification), not two separate
// calls: this closes the window where a partial failure (or a reader racing
// the two calls) could otherwise observe the new, unverified address stored
// together with a stale, still-set email_verified_at from the old address.
func (s *Service) UpdateMe(ctx context.Context, userID uuid.UUID, username *string, email *string) (*ent.UserEntity, error) {
	current, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrAuthUnauthorized
	}

	mergedEmail := current.Email
	emailChanged := false
	if email != nil {
		v := strings.ToLower(strings.TrimSpace(*email))
		mergedEmail = &v
		emailChanged = normalizeEmailForCompare(mergedEmail) != normalizeEmailForCompare(current.Email)
	}
	mergedUsername := current.Username
	if username != nil {
		v := strings.TrimSpace(*username)
		mergedUsername = &v
	}

	isActive := current.IsActive
	isGuest := current.IsGuest
	plan := string(current.Plan)

	if emailChanged && mergedEmail != nil && strings.TrimSpace(*mergedEmail) != "" {
		if existing, gErr := s.user.GetUserByEmail(ctx, *mergedEmail); gErr == nil && existing != nil && existing.ID != userID {
			return nil, ErrUserEmailAlreadyExists
		} else if gErr != nil && !errors.Is(gErr, sql.ErrNoRows) {
			return nil, gErr
		}
	}

	updated, err := s.user.UpdateUser(ctx, userID, entitiesdto.UpdateUser{
		Email:                  mergedEmail,
		Username:               mergedUsername,
		IsActive:               &isActive,
		IsGuest:                &isGuest,
		Plan:                   &plan,
		ClearEmailVerification: emailChanged,
	})
	if err != nil {
		if isEmailUniqueConflict(err) {
			return nil, ErrUserEmailAlreadyExists
		}
		return nil, err
	}

	if emailChanged {
		// Send the fresh verification email asynchronously, same convention as
		// registration; don't block the profile-update response on mail delivery.
		go func() {
			_ = s.SendEmailVerification(context.Background(), updated)
		}()
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

// DeleteMe permanently deletes the authenticated user's own account.
func (s *Service) DeleteMe(ctx context.Context, userID uuid.UUID) error {
	_, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		return ErrAuthUnauthorized
	}
	return s.user.DeleteUser(ctx, userID)
}
