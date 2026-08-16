package auth

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"

	"github.com/google/uuid"
)

// fakeUserEnt is a minimal stub of entitiesinf.UserEntity: it embeds the (nil)
// interface so it satisfies the full interface, and only overrides the
// methods UpdateMe actually exercises (including the async
// SendEmailVerification path it may trigger, so that path never panics on a
// nil embedded interface).
type fakeUserEnt struct {
	entitiesinf.UserEntity

	byID    map[uuid.UUID]*ent.UserEntity
	byEmail map[string]*ent.UserEntity

	updateUserFn func(ctx context.Context, id uuid.UUID, in entitiesdto.UpdateUser) (*ent.UserEntity, error)
}

func (f *fakeUserEnt) GetUserByID(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	if u, ok := f.byID[id]; ok {
		return u, nil
	}
	return nil, sql.ErrNoRows
}

// GetUserByEmail mirrors entities.Service.GetUserByEmail's case-insensitive
// lookup so tests exercise the same "same address, different case" semantics
// the real implementation provides.
func (f *fakeUserEnt) GetUserByEmail(ctx context.Context, email string) (*ent.UserEntity, error) {
	for _, u := range f.byEmail {
		if u.Email != nil && normalizeEmailForCompare(u.Email) == normalizeEmailForCompare(&email) {
			return u, nil
		}
	}
	return nil, sql.ErrNoRows
}

func (f *fakeUserEnt) UpdateUser(ctx context.Context, id uuid.UUID, in entitiesdto.UpdateUser) (*ent.UserEntity, error) {
	if f.updateUserFn != nil {
		return f.updateUserFn(ctx, id, in)
	}
	return nil, errors.New("updateUserFn not stubbed")
}

func (f *fakeUserEnt) CreateEmailVerificationToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*ent.EmailVerificationTokenEntity, error) {
	// UpdateMe fires SendEmailVerification asynchronously on a real email
	// change; stub this so that goroutine can run to completion without a nil
	// embedded-interface panic if it happens to execute during the test.
	return &ent.EmailVerificationTokenEntity{ID: uuid.New(), UserID: userID}, nil
}

func newTestAuthService(userEnt entitiesinf.UserEntity) *Service {
	return &Service{user: userEnt}
}

func TestUpdateMe_UsernameOnlyChange_PreservesVerification(t *testing.T) {
	userID := uuid.New()
	email := "user@example.com"
	verifiedAt := time.Now().Add(-time.Hour)

	var capturedIn entitiesdto.UpdateUser
	userEnt := &fakeUserEnt{
		byID: map[uuid.UUID]*ent.UserEntity{
			userID: {ID: userID, Email: &email, EmailVerifiedAt: &verifiedAt},
		},
		updateUserFn: func(ctx context.Context, id uuid.UUID, in entitiesdto.UpdateUser) (*ent.UserEntity, error) {
			capturedIn = in
			return &ent.UserEntity{ID: id, Email: in.Email, Username: in.Username, EmailVerifiedAt: &verifiedAt}, nil
		},
	}

	svc := newTestAuthService(userEnt)

	newUsername := "newname"
	_, err := svc.UpdateMe(context.Background(), userID, &newUsername, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedIn.ClearEmailVerification {
		t.Fatalf("username-only change must not request clearing email verification")
	}
}

func TestUpdateMe_EmailResubmittedSameCaseInsensitiveAddress_PreservesVerification(t *testing.T) {
	userID := uuid.New()
	email := "User@Example.com"
	verifiedAt := time.Now().Add(-time.Hour)

	var capturedIn entitiesdto.UpdateUser
	userEnt := &fakeUserEnt{
		byID: map[uuid.UUID]*ent.UserEntity{
			userID: {ID: userID, Email: &email, EmailVerifiedAt: &verifiedAt},
		},
		updateUserFn: func(ctx context.Context, id uuid.UUID, in entitiesdto.UpdateUser) (*ent.UserEntity, error) {
			capturedIn = in
			return &ent.UserEntity{ID: id, Email: in.Email, EmailVerifiedAt: &verifiedAt}, nil
		},
	}

	svc := newTestAuthService(userEnt)

	// Same address, different case/whitespace: must NOT count as a real change.
	resubmitted := "  user@example.com "
	_, err := svc.UpdateMe(context.Background(), userID, nil, &resubmitted)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if capturedIn.ClearEmailVerification {
		t.Fatalf("resubmitting the same address in a different case must not clear verification")
	}
	if capturedIn.Email == nil || *capturedIn.Email != "user@example.com" {
		t.Fatalf("expected the stored email to be normalized to lowercase, got %v", capturedIn.Email)
	}
}

func TestUpdateMe_RealEmailChange_ClearsVerificationAtomicallyWithUpdate(t *testing.T) {
	userID := uuid.New()
	oldEmail := "old@example.com"
	verifiedAt := time.Now().Add(-time.Hour)

	var capturedIn entitiesdto.UpdateUser
	userEnt := &fakeUserEnt{
		byID: map[uuid.UUID]*ent.UserEntity{
			userID: {ID: userID, Email: &oldEmail, EmailVerifiedAt: &verifiedAt},
		},
		updateUserFn: func(ctx context.Context, id uuid.UUID, in entitiesdto.UpdateUser) (*ent.UserEntity, error) {
			capturedIn = in
			// Simulate the atomic entities-layer UPDATE: email changes and
			// email_verified_at is cleared in the same call.
			var newVerifiedAt *time.Time
			if !in.ClearEmailVerification {
				newVerifiedAt = &verifiedAt
			}
			return &ent.UserEntity{ID: id, Email: in.Email, EmailVerifiedAt: newVerifiedAt}, nil
		},
	}

	svc := newTestAuthService(userEnt)

	newEmail := "New@Example.com"
	updated, err := svc.UpdateMe(context.Background(), userID, nil, &newEmail)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !capturedIn.ClearEmailVerification {
		t.Fatalf("a genuine email change must request clearing email verification atomically with the update")
	}
	if capturedIn.Email == nil || *capturedIn.Email != "new@example.com" {
		t.Fatalf("expected the new email to be normalized to lowercase before storage, got %v", capturedIn.Email)
	}
	if updated.EmailVerifiedAt != nil {
		t.Fatalf("expected email_verified_at to be cleared after a real email change")
	}
}

func TestUpdateMe_EmailChangeToAnotherUsersAddress_IsRejected(t *testing.T) {
	userID := uuid.New()
	otherID := uuid.New()
	myEmail := "me@example.com"
	otherEmail := "taken@example.com"
	verifiedAt := time.Now().Add(-time.Hour)

	userEnt := &fakeUserEnt{
		byID: map[uuid.UUID]*ent.UserEntity{
			userID: {ID: userID, Email: &myEmail, EmailVerifiedAt: &verifiedAt},
		},
		byEmail: map[string]*ent.UserEntity{
			otherEmail: {ID: otherID, Email: &otherEmail},
		},
		updateUserFn: func(ctx context.Context, id uuid.UUID, in entitiesdto.UpdateUser) (*ent.UserEntity, error) {
			t.Fatalf("UpdateUser must not be called when the target email is already taken by another user")
			return nil, nil
		},
	}

	svc := newTestAuthService(userEnt)

	// Different case of the already-registered address: the case-insensitive
	// GetUserByEmail lookup must still catch this as a conflict.
	attempted := "TAKEN@example.com"
	_, err := svc.UpdateMe(context.Background(), userID, nil, &attempted)
	if !errors.Is(err, ErrUserEmailAlreadyExists) {
		t.Fatalf("expected ErrUserEmailAlreadyExists, got %v", err)
	}
}
