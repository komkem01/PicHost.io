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
		Column("email", "password", "username", "plan", "is_active", "is_guest", "created_at", "updated_at").
		Returning("*").
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
	_, err := s.db.NewUpdate().
		TableExpr("users").
		Set("plan = ?, updated_at = ?", ent.PlanType(*plan.Plan), now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return s.GetUserByID(ctx, id)
}

// SetUserPlanExpiry sets or clears the plan_expires_at column.
// Pass nil to clear the expiry (e.g. when the user is on the Free plan).
// Set clearCancellation = true when renewing to also clear plan_cancelled_at.
func (s *Service) SetUserPlanExpiry(ctx context.Context, id uuid.UUID, expiresAt *time.Time, clearCancellation bool) (*ent.UserEntity, error) {
	now := time.Now()
	q := s.db.NewUpdate().TableExpr("users")
	if clearCancellation {
		q = q.Set("plan_expires_at = ?, plan_cancelled_at = NULL, updated_at = ?", expiresAt, now)
	} else {
		q = q.Set("plan_expires_at = ?, updated_at = ?", expiresAt, now)
	}
	_, err := q.Where("id = ?", id).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return s.GetUserByID(ctx, id)
}

// CancelUserPlan records the cancellation timestamp. The plan remains active until plan_expires_at.
func (s *Service) CancelUserPlan(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	now := time.Now()
	_, err := s.db.NewUpdate().
		TableExpr("users").
		Set("plan_cancelled_at = ?, updated_at = ?", now, now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return s.GetUserByID(ctx, id)
}

// DowngradeUserPlanToFree sets the plan to Free and clears both plan_expires_at and plan_cancelled_at.
func (s *Service) DowngradeUserPlanToFree(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	now := time.Now()
	_, err := s.db.NewUpdate().
		TableExpr("users").
		Set("plan = ?, plan_expires_at = NULL, plan_cancelled_at = NULL, updated_at = ?", ent.PlanTypeFree, now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return s.GetUserByID(ctx, id)
}

func (s *Service) UpdateUserPassword(ctx context.Context, id uuid.UUID, password entitiesdto.UpdateUserPassword) (*ent.UserEntity, error) {
	now := time.Now()
	data := &ent.UserEntity{
		Password:  &password.NewPassword,
		UpdatedAt: now,
	}
	_, err := s.db.NewUpdate().
		Model(data).
		Column("password", "updated_at").
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	_, _ = s.db.NewDelete().TableExpr("email_verification_tokens").Where("user_id = ?", id).Exec(ctx)
	_, _ = s.db.NewDelete().TableExpr("images").Where("user_id = ?", id).Exec(ctx)
	_, _ = s.db.NewDelete().TableExpr("storage").Where("user_id = ?", id).Exec(ctx)
	_, _ = s.db.NewDelete().TableExpr("payments").Where("user_id = ?", id).Exec(ctx)

	_, err := s.db.NewDelete().
		Model((*ent.UserEntity)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (s *Service) SetUserActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	now := time.Now()
	_, err := s.db.NewUpdate().
		TableExpr("users").
		Set("is_active = ?, updated_at = ?", isActive, now).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (s *Service) UpdateUserProfile(ctx context.Context, id uuid.UUID, profile entitiesdto.UpdateUserProfile) (*ent.UserEntity, error) {
	now := time.Now()
	_, err := s.db.NewUpdate().
		TableExpr("users").
		Set("email = ?, username = ?, updated_at = ?", profile.Email, profile.Username, now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return s.GetUserByID(ctx, id)
}

func (s *Service) SetUserAdmin(ctx context.Context, id uuid.UUID, isAdmin bool) error {
	now := time.Now()
	_, err := s.db.NewUpdate().
		TableExpr("users").
		Set("is_admin = ?, updated_at = ?", isAdmin, now).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (s *Service) SetUserEmailVerified(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	now := time.Now()
	_, err := s.db.NewUpdate().
		TableExpr("users").
		Set("email_verified_at = ?, updated_at = ?", now, now).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return s.GetUserByID(ctx, id)
}

func (s *Service) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*ent.PasswordResetTokenEntity, error) {
	data := &ent.PasswordResetTokenEntity{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	_, err := s.db.NewInsert().
		Model(data).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*ent.PasswordResetTokenEntity, error) {
	var token ent.PasswordResetTokenEntity
	err := s.db.NewSelect().
		Model(&token).
		Where("token_hash = ?", tokenHash).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (s *Service) MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error {
	now := time.Now()
	_, err := s.db.NewUpdate().
		Model((*ent.PasswordResetTokenEntity)(nil)).
		Set("used_at = ?", now).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (s *Service) CreateEmailVerificationToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*ent.EmailVerificationTokenEntity, error) {
	// Delete any old tokens for this user first
	_, _ = s.db.NewDelete().
		Model((*ent.EmailVerificationTokenEntity)(nil)).
		Where("user_id = ?", userID).
		Exec(ctx)

	data := &ent.EmailVerificationTokenEntity{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	_, err := s.db.NewInsert().
		Model(data).
		Exec(ctx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) GetEmailVerificationTokenByHash(ctx context.Context, tokenHash string) (*ent.EmailVerificationTokenEntity, error) {
	var token ent.EmailVerificationTokenEntity
	err := s.db.NewSelect().
		Model(&token).
		Where("token_hash = ?", tokenHash).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &token, nil
}

func (s *Service) DeleteEmailVerificationToken(ctx context.Context, id uuid.UUID) error {
	_, err := s.db.NewDelete().
		Model((*ent.EmailVerificationTokenEntity)(nil)).
		Where("id = ?", id).
		Exec(ctx)
	return err
}

func (s *Service) RecordUserLogin(ctx context.Context, id uuid.UUID, ip *string) error {
	now := time.Now()
	_, err := s.db.NewUpdate().
		TableExpr("users").
		Set("last_login_at = ?, login_count = login_count + 1, last_login_ip = ?, updated_at = ?", now, ip, now).
		Where("id = ?", id).
		Exec(ctx)
	return err
}


