package auth

import (
	"context"
	"testing"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	"pichost.io/internal/config"

	"github.com/google/uuid"
)

type mockQuotaEnt struct {
	quotas map[uuid.UUID]*ent.UserQuotaEntity
}

func newMockQuotaEnt() *mockQuotaEnt {
	return &mockQuotaEnt{quotas: make(map[uuid.UUID]*ent.UserQuotaEntity)}
}

func (m *mockQuotaEnt) GetUserQuota(ctx context.Context, userID uuid.UUID) (*ent.UserQuotaEntity, error) {
	q, ok := m.quotas[userID]
	if !ok {
		return &ent.UserQuotaEntity{UserID: userID, UsedStorageBytes: 0, ImageCount: 0}, nil
	}
	return q, nil
}

func (m *mockQuotaEnt) UpsertUserQuota(ctx context.Context, userID uuid.UUID) (*ent.UserQuotaEntity, error) {
	q := &ent.UserQuotaEntity{UserID: userID, UsedStorageBytes: 0, ImageCount: 0}
	m.quotas[userID] = q
	return q, nil
}

func (m *mockQuotaEnt) AddToUserQuota(ctx context.Context, userID uuid.UUID, delta entitiesdto.AddToUserQuota) (*ent.UserQuotaEntity, error) {
	q, _ := m.GetUserQuota(ctx, userID)
	q.UsedStorageBytes += delta.StorageDelta
	q.ImageCount += delta.ImageCountDelta
	m.quotas[userID] = q
	return q, nil
}

func TestJWTSignAndParse(t *testing.T) {
	conf := &config.Config[Config]{
		Val: &Config{
			JWTSecret:             "super-secret-jwt-key-32-bytes!!",
			JWTIssuer:             "pichost-test",
			AccessTokenTTLSeconds: 3600,
		},
	}
	mockUsers := newMockUserEnt()
	mockAuth := newMockAuthEnt()
	mockQuota := newMockQuotaEnt()

	svc := newService(&Options{
		Config:   conf,
		user:     mockUsers,
		auth:     mockAuth,
		quotaEnt: mockQuota,
	})

	userID := uuid.New()
	token, err := svc.signAccessToken(userID)
	if err != nil {
		t.Fatalf("signAccessToken failed: %v", err)
	}

	parsedID, err := svc.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken failed: %v", err)
	}
	if parsedID != userID {
		t.Errorf("expected userID %s, got %s", userID, parsedID)
	}
}

func TestInvalidJWT(t *testing.T) {
	conf := &config.Config[Config]{
		Val: &Config{
			JWTSecret:             "secret-key",
			JWTIssuer:             "pichost-test",
			AccessTokenTTLSeconds: 3600,
		},
	}
	svc := newService(&Options{
		Config:   conf,
		user:     newMockUserEnt(),
		auth:     newMockAuthEnt(),
		quotaEnt: newMockQuotaEnt(),
	})

	// Test malformed token
	_, err := svc.ParseAccessToken("invalid.token.string")
	if err == nil {
		t.Error("expected error for malformed token, got nil")
	}

	// Test tampered token signature
	userID := uuid.New()
	token, _ := svc.signAccessToken(userID)
	tamperedToken := token + "tampered"
	_, err = svc.ParseAccessToken(tamperedToken)
	if err == nil {
		t.Error("expected error for tampered token signature, got nil")
	}
}

func TestRegisterAndLoginFlow(t *testing.T) {
	conf := &config.Config[Config]{
		Val: &Config{
			JWTSecret:              "secret-key-for-test-32-chars!!",
			JWTIssuer:              "pichost-test",
			AccessTokenTTLSeconds:  3600,
			RefreshTokenTTLSeconds: 86400 * 30,
		},
	}
	mockUsers := newMockUserEnt()
	mockAuth := newMockAuthEnt()
	mockQuota := newMockQuotaEnt()

	svc := newService(&Options{
		Config:   conf,
		user:     mockUsers,
		auth:     mockAuth,
		quotaEnt: mockQuota,
	})

	ctx := context.Background()
	regRes, err := svc.Register(ctx, RegisterRequestService{
		Email:    "testuser@example.com",
		Password: "SecurePassword123!",
		Username: "testuser",
	}, "TestAgent", "127.0.0.1")
	if err != nil {
		t.Fatalf("Register failed: %v", err)
	}

	if regRes.AccessToken == "" || regRes.RefreshToken == "" {
		t.Errorf("expected tokens to be issued on registration")
	}

	// Test Duplicate Email Register
	_, err = svc.Register(ctx, RegisterRequestService{
		Email:    "testuser@example.com",
		Password: "AnotherPassword123!",
		Username: "testuser2",
	}, "TestAgent", "127.0.0.1")
	if err != ErrUserEmailAlreadyExists {
		t.Errorf("expected ErrUserEmailAlreadyExists, got %v", err)
	}

	// Test Login
	loginRes, err := svc.Login(ctx, LoginRequestService{
		Email:    "testuser@example.com",
		Password: "SecurePassword123!",
	}, "TestAgent", "127.0.0.1")
	if err != nil {
		t.Fatalf("Login failed: %v", err)
	}
	if loginRes.AccessToken == "" {
		t.Error("expected access token on login")
	}

	// Test Login Wrong Password
	_, err = svc.Login(ctx, LoginRequestService{
		Email:    "testuser@example.com",
		Password: "WrongPassword!",
	}, "TestAgent", "127.0.0.1")
	if err != ErrAuthInvalidCredentials {
		t.Errorf("expected ErrAuthInvalidCredentials for wrong password, got %v", err)
	}
}

