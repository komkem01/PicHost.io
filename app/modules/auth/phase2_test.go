package auth

import (
	"context"
	"database/sql"
	"testing"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
)

// Mock entities for unit testing Phase 2 logic without DB dependency
type mockUserEnt struct {
	users map[uuid.UUID]*ent.UserEntity
	resetTokens map[string]*ent.PasswordResetTokenEntity
	verifyTokens map[string]*ent.EmailVerificationTokenEntity
}

func newMockUserEnt() *mockUserEnt {
	return &mockUserEnt{
		users: make(map[uuid.UUID]*ent.UserEntity),
		resetTokens: make(map[string]*ent.PasswordResetTokenEntity),
		verifyTokens: make(map[string]*ent.EmailVerificationTokenEntity),
	}
}

func (m *mockUserEnt) CreateUser(ctx context.Context, u entitiesdto.CreateUser) (*ent.UserEntity, error) {
	id := uuid.New()
	user := &ent.UserEntity{
		ID:        id,
		Email:     u.Email,
		Password:  u.Password,
		Username:  u.Username,
		Plan:      ent.PlanType(u.Plan),
		IsActive:  true,
		IsGuest:   u.IsGuest,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	m.users[id] = user
	return user, nil
}

func (m *mockUserEnt) GetUserByID(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	return u, nil
}

func (m *mockUserEnt) GetListUser(ctx context.Context) ([]*ent.UserEntity, error) {
	list := make([]*ent.UserEntity, 0, len(m.users))
	for _, u := range m.users {
		list = append(list, u)
	}
	return list, nil
}

func (m *mockUserEnt) GetUserByEmail(ctx context.Context, email string) (*ent.UserEntity, error) {
	for _, u := range m.users {
		if u.Email != nil && *u.Email == email {
			return u, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (m *mockUserEnt) UpdateUser(ctx context.Context, id uuid.UUID, user entitiesdto.UpdateUser) (*ent.UserEntity, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	u.Email = user.Email
	u.Username = user.Username
	return u, nil
}

func (m *mockUserEnt) UpdateUserPlan(ctx context.Context, id uuid.UUID, plan entitiesdto.UpdateUserPlan) (*ent.UserEntity, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	u.Plan = ent.PlanType(*plan.Plan)
	return u, nil
}

func (m *mockUserEnt) SetUserPlanExpiry(ctx context.Context, id uuid.UUID, expiresAt *time.Time, clearCancellation bool) (*ent.UserEntity, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	u.PlanExpiresAt = expiresAt
	if clearCancellation {
		u.PlanCancelledAt = nil
	}
	return u, nil
}

func (m *mockUserEnt) CancelUserPlan(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	now := time.Now()
	u.PlanCancelledAt = &now
	return u, nil
}

func (m *mockUserEnt) DowngradeUserPlanToFree(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	u.Plan = ent.PlanTypeFree
	u.PlanExpiresAt = nil
	u.PlanCancelledAt = nil
	return u, nil
}

func (m *mockUserEnt) UpdateUserProfile(ctx context.Context, id uuid.UUID, profile entitiesdto.UpdateUserProfile) (*ent.UserEntity, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	u.Email = profile.Email
	u.Username = profile.Username
	return u, nil
}

func (m *mockUserEnt) SetUserActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	u, ok := m.users[id]
	if !ok {
		return ErrUserNotFound
	}
	u.IsActive = isActive
	return nil
}

func (m *mockUserEnt) UpdateUserPassword(ctx context.Context, id uuid.UUID, password entitiesdto.UpdateUserPassword) (*ent.UserEntity, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	u.Password = &password.NewPassword
	return u, nil
}

func (m *mockUserEnt) SetUserAdmin(ctx context.Context, id uuid.UUID, isAdmin bool) error {
	u, ok := m.users[id]
	if !ok {
		return ErrUserNotFound
	}
	u.IsAdmin = isAdmin
	return nil
}

func (m *mockUserEnt) SetUserEmailVerified(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, ErrUserNotFound
	}
	now := time.Now()
	u.EmailVerifiedAt = &now
	return u, nil
}

func (m *mockUserEnt) DeleteUser(ctx context.Context, id uuid.UUID) error {
	delete(m.users, id)
	return nil
}

func (m *mockUserEnt) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*ent.PasswordResetTokenEntity, error) {
	t := &ent.PasswordResetTokenEntity{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	m.resetTokens[tokenHash] = t
	return t, nil
}

func (m *mockUserEnt) GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*ent.PasswordResetTokenEntity, error) {
	t, ok := m.resetTokens[tokenHash]
	if !ok {
		return nil, ErrInvalidOrExpiredToken
	}
	return t, nil
}

func (m *mockUserEnt) MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error {
	for _, t := range m.resetTokens {
		if t.ID == id {
			now := time.Now()
			t.UsedAt = &now
			return nil
		}
	}
	return nil
}

func (m *mockUserEnt) CreateEmailVerificationToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*ent.EmailVerificationTokenEntity, error) {
	t := &ent.EmailVerificationTokenEntity{
		ID:        uuid.New(),
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
		CreatedAt: time.Now(),
	}
	m.verifyTokens[tokenHash] = t
	return t, nil
}

func (m *mockUserEnt) GetEmailVerificationTokenByHash(ctx context.Context, tokenHash string) (*ent.EmailVerificationTokenEntity, error) {
	t, ok := m.verifyTokens[tokenHash]
	if !ok {
		return nil, ErrInvalidOrExpiredToken
	}
	return t, nil
}

func (m *mockUserEnt) DeleteEmailVerificationToken(ctx context.Context, id uuid.UUID) error {
	for k, t := range m.verifyTokens {
		if t.ID == id {
			delete(m.verifyTokens, k)
			return nil
		}
	}
	return nil
}

type mockAuthEnt struct {
	sessions map[uuid.UUID]*ent.AuthSessionEntity
}

func newMockAuthEnt() *mockAuthEnt {
	return &mockAuthEnt{
		sessions: make(map[uuid.UUID]*ent.AuthSessionEntity),
	}
}

func (m *mockAuthEnt) CreateAuthSession(ctx context.Context, auth entitiesdto.CreateAuthSession) (*ent.AuthSessionEntity, error) {
	id := uuid.New()
	s := &ent.AuthSessionEntity{
		ID:               id,
		UserID:           auth.UserID,
		RefreshTokenHash: auth.RefreshTokenHash,
		UserAgent:        auth.UserAgent,
		IPAddress:        auth.IPAddress,
		ExpiresAt:        auth.ExpiresAt,
		CreatedAt:        time.Now(),
		UpdatedAt:        time.Now(),
	}
	m.sessions[id] = s
	return s, nil
}

func (m *mockAuthEnt) GetAuthSessionByRefreshTokenHash(ctx context.Context, refreshTokenHash string) (*ent.AuthSessionEntity, error) {
	for _, s := range m.sessions {
		if s.RefreshTokenHash == refreshTokenHash && s.RevokedAt == nil {
			return s, nil
		}
	}
	return nil, ErrAuthSessionNotFound
}

func (m *mockAuthEnt) RotateAuthSession(ctx context.Context, id uuid.UUID, auth entitiesdto.RotateAuthSession) (*ent.AuthSessionEntity, error) {
	s, ok := m.sessions[id]
	if !ok {
		return nil, ErrAuthSessionNotFound
	}
	s.RefreshTokenHash = auth.RefreshTokenHash
	s.ExpiresAt = auth.ExpiresAt
	return s, nil
}

func (m *mockAuthEnt) RevokeAuthSession(ctx context.Context, id uuid.UUID) error {
	s, ok := m.sessions[id]
	if ok {
		now := time.Now()
		s.RevokedAt = &now
	}
	return nil
}

func (m *mockAuthEnt) RevokeAuthSessionsByUserID(ctx context.Context, userID uuid.UUID) error {
	now := time.Now()
	for _, s := range m.sessions {
		if s.UserID == userID {
			s.RevokedAt = &now
		}
	}
	return nil
}

func (m *mockAuthEnt) CreateOAuthAccount(ctx context.Context, account entitiesdto.CreateOAuthAccount) (*ent.OAuthAccountEntity, error) {
	return nil, nil
}

func (m *mockAuthEnt) GetOAuthAccountByProviderUserID(ctx context.Context, provider string, providerUserID string) (*ent.OAuthAccountEntity, error) {
	return nil, ErrUserNotFound
}

func TestForgotPassword_EnumerationDefense(t *testing.T) {
	userEnt := newMockUserEnt()
	authEnt := newMockAuthEnt()

	svc := &Service{
		user: userEnt,
		auth: authEnt,
	}

	// Registered user
	email := "user@example.com"
	_, _ = userEnt.CreateUser(context.Background(), entitiesdto.CreateUser{
		Email: &email,
	})

	// 1. Registered email -> returns nil (no error, 200 OK)
	err := svc.ForgotPassword(context.Background(), email)
	if err != nil {
		t.Fatalf("expected nil error for registered email, got %v", err)
	}
	if len(userEnt.resetTokens) != 1 {
		t.Fatalf("expected 1 reset token created, got %d", len(userEnt.resetTokens))
	}

	// 2. Unregistered email -> also returns nil (no error, 200 OK — enumeration defense!)
	err = svc.ForgotPassword(context.Background(), "unknown@example.com")
	if err != nil {
		t.Fatalf("expected nil error for unregistered email, got %v", err)
	}
}

// Test Password Reset Token Hashing & Expiry Test
func TestResetPassword_TokenHashingAndRevocation(t *testing.T) {
	userEnt := newMockUserEnt()
	authEnt := newMockAuthEnt()

	svc := &Service{
		user: userEnt,
		auth: authEnt,
	}

	email := "reset@example.com"
	u, _ := userEnt.CreateUser(context.Background(), entitiesdto.CreateUser{
		Email: &email,
	})

	// Create an active session for the user
	sess, _ := authEnt.CreateAuthSession(context.Background(), entitiesdto.CreateAuthSession{
		UserID:           u.ID,
		RefreshTokenHash: "hash123",
		ExpiresAt:        time.Now().Add(1 * time.Hour),
	})

	// Trigger forgot password to generate token
	rawToken, tokenHash, _ := newRefreshToken()
	expiresAt := time.Now().Add(1 * time.Hour)
	_, _ = userEnt.CreatePasswordResetToken(context.Background(), u.ID, tokenHash, expiresAt)

	// Verify token stored in DB is hashed and not raw
	if _, exists := userEnt.resetTokens[rawToken]; exists {
		t.Fatalf("raw token must NOT be stored in DB!")
	}
	if _, exists := userEnt.resetTokens[tokenHash]; !exists {
		t.Fatalf("token SHA-256 hash must be stored in DB!")
	}

	// Reset password with valid token
	err := svc.ResetPassword(context.Background(), rawToken, "NewSecurePassword123!")
	if err != nil {
		t.Fatalf("expected password reset success, got %v", err)
	}

	// Verify all existing sessions for this user were revoked!
	if sess.RevokedAt == nil {
		t.Fatalf("sessions must be revoked after password reset!")
	}

	// Attempting to reuse the same token must fail
	err = svc.ResetPassword(context.Background(), rawToken, "AnotherPassword123!")
	if err == nil {
		t.Fatalf("expected token reuse to fail!")
	}
}

func TestEmailVerification(t *testing.T) {
	userEnt := newMockUserEnt()
	authEnt := newMockAuthEnt()

	svc := &Service{
		user: userEnt,
		auth: authEnt,
	}

	email := "verify@example.com"
	u, _ := userEnt.CreateUser(context.Background(), entitiesdto.CreateUser{
		Email: &email,
	})

	if u.EmailVerifiedAt != nil {
		t.Fatalf("new user should not be verified initially")
	}

	// Trigger SendEmailVerification
	err := svc.SendEmailVerification(context.Background(), u)
	if err != nil {
		t.Fatalf("expected verification email trigger success, got %v", err)
	}

	if len(userEnt.verifyTokens) != 1 {
		t.Fatalf("expected 1 verification token in DB")
	}

	for hash := range userEnt.verifyTokens {
		// we know verify tokens store hash, test verification with raw invalid token
		err := svc.VerifyEmail(context.Background(), hash)
		if err == nil {
			t.Fatalf("raw token must match hash calculation")
		}
	}
}
