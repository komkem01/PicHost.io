package quota

import (
	"context"
	"database/sql"
	"testing"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	"pichost.io/internal/config"

	"github.com/google/uuid"
	"go.opentelemetry.io/otel"
)

type mockUserEnt struct {
	users map[uuid.UUID]*ent.UserEntity
}

func newMockUserEnt() *mockUserEnt {
	return &mockUserEnt{users: make(map[uuid.UUID]*ent.UserEntity)}
}

func (m *mockUserEnt) CreateUser(ctx context.Context, u entitiesdto.CreateUser) (*ent.UserEntity, error) {
	id := uuid.New()
	user := &ent.UserEntity{
		ID:        id,
		Email:     u.Email,
		Plan:      ent.PlanType(u.Plan),
		IsActive:  true,
		CreatedAt: time.Now(),
	}
	m.users[id] = user
	return user, nil
}

func (m *mockUserEnt) GetUserByID(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return u, nil
}

func (m *mockUserEnt) GetListUser(ctx context.Context) ([]*ent.UserEntity, error) { return nil, nil }
func (m *mockUserEnt) GetUserByEmail(ctx context.Context, email string) (*ent.UserEntity, error) {
	return nil, nil
}
func (m *mockUserEnt) UpdateUser(ctx context.Context, id uuid.UUID, user entitiesdto.UpdateUser) (*ent.UserEntity, error) {
	return nil, nil
}
func (m *mockUserEnt) UpdateUserPlan(ctx context.Context, id uuid.UUID, plan entitiesdto.UpdateUserPlan) (*ent.UserEntity, error) {
	return nil, nil
}
func (m *mockUserEnt) SetUserPlanExpiry(ctx context.Context, id uuid.UUID, expiresAt *time.Time, clearCancellation bool) (*ent.UserEntity, error) {
	return nil, nil
}
func (m *mockUserEnt) CancelUserPlan(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	return nil, nil
}
func (m *mockUserEnt) DowngradeUserPlanToFree(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	u, ok := m.users[id]
	if !ok {
		return nil, sql.ErrNoRows
	}
	u.Plan = ent.PlanTypeFree
	u.PlanExpiresAt = nil
	return u, nil
}
func (m *mockUserEnt) UpdateUserProfile(ctx context.Context, id uuid.UUID, profile entitiesdto.UpdateUserProfile) (*ent.UserEntity, error) {
	return nil, nil
}
func (m *mockUserEnt) SetUserActive(ctx context.Context, id uuid.UUID, isActive bool) error {
	return nil
}
func (m *mockUserEnt) UpdateUserPassword(ctx context.Context, id uuid.UUID, password entitiesdto.UpdateUserPassword) (*ent.UserEntity, error) {
	return nil, nil
}
func (m *mockUserEnt) SetUserAdmin(ctx context.Context, id uuid.UUID, isAdmin bool) error {
	return nil
}
func (m *mockUserEnt) SetUserEmailVerified(ctx context.Context, id uuid.UUID) (*ent.UserEntity, error) {
	return nil, nil
}
func (m *mockUserEnt) DeleteUser(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockUserEnt) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*ent.PasswordResetTokenEntity, error) {
	return nil, nil
}
func (m *mockUserEnt) GetPasswordResetTokenByHash(ctx context.Context, tokenHash string) (*ent.PasswordResetTokenEntity, error) {
	return nil, nil
}
func (m *mockUserEnt) MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error {
	return nil
}
func (m *mockUserEnt) CreateEmailVerificationToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) (*ent.EmailVerificationTokenEntity, error) {
	return nil, nil
}
func (m *mockUserEnt) GetEmailVerificationTokenByHash(ctx context.Context, tokenHash string) (*ent.EmailVerificationTokenEntity, error) {
	return nil, nil
}
func (m *mockUserEnt) DeleteEmailVerificationToken(ctx context.Context, id uuid.UUID) error {
	return nil
}

type mockQuotaEnt struct {
	quotas map[uuid.UUID]*ent.UserQuotaEntity
}

func newMockQuotaEnt() *mockQuotaEnt {
	return &mockQuotaEnt{quotas: make(map[uuid.UUID]*ent.UserQuotaEntity)}
}

func (m *mockQuotaEnt) GetUserQuota(ctx context.Context, userID uuid.UUID) (*ent.UserQuotaEntity, error) {
	q, ok := m.quotas[userID]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return q, nil
}
func (m *mockQuotaEnt) UpsertUserQuota(ctx context.Context, userID uuid.UUID) (*ent.UserQuotaEntity, error) {
	q := &ent.UserQuotaEntity{UserID: userID, UsedStorageBytes: 0, ImageCount: 0}
	m.quotas[userID] = q
	return q, nil
}
func (m *mockQuotaEnt) AddToUserQuota(ctx context.Context, userID uuid.UUID, delta entitiesdto.AddToUserQuota) (*ent.UserQuotaEntity, error) {
	q, ok := m.quotas[userID]
	if !ok {
		q = &ent.UserQuotaEntity{UserID: userID}
	}
	q.UsedStorageBytes += delta.StorageDelta
	q.ImageCount += delta.ImageCountDelta
	m.quotas[userID] = q
	return q, nil
}

type mockPlanEnt struct {
	plans map[string]*ent.PlanSettingEntity
}

func (m *mockPlanEnt) ListPlanSettings(ctx context.Context) ([]*ent.PlanSettingEntity, error) {
	return nil, nil
}
func (m *mockPlanEnt) GetPlanSettingByKey(ctx context.Context, key string) (*ent.PlanSettingEntity, error) {
	p, ok := m.plans[key]
	if !ok {
		return nil, sql.ErrNoRows
	}
	return p, nil
}
func (m *mockPlanEnt) UpsertPlanSetting(ctx context.Context, setting entitiesdto.UpsertPlanSetting) (*ent.PlanSettingEntity, error) {
	return nil, nil
}
func (m *mockPlanEnt) DeletePlanSettingByKey(ctx context.Context, key string) error { return nil }

type mockImageEnt struct {
	expiredImages []*ent.ImageEntity
}

func (m *mockImageEnt) CreateImage(ctx context.Context, image entitiesdto.CreateImage) (*ent.ImageEntity, error) {
	return nil, nil
}
func (m *mockImageEnt) GetImageByID(ctx context.Context, id uuid.UUID) (*ent.ImageEntity, error) {
	return nil, nil
}
func (m *mockImageEnt) GetImageByStorageID(ctx context.Context, storageID uuid.UUID) (*ent.ImageEntity, error) {
	return nil, nil
}
func (m *mockImageEnt) GetImagesByUserID(ctx context.Context, userID uuid.UUID) ([]*ent.ImageEntity, error) {
	return nil, nil
}
func (m *mockImageEnt) UpdateImage(ctx context.Context, id uuid.UUID, image entitiesdto.UpdateImage) (*ent.ImageEntity, error) {
	return nil, nil
}
func (m *mockImageEnt) DeleteImage(ctx context.Context, id uuid.UUID) error { return nil }
func (m *mockImageEnt) ListExpiredImages(ctx context.Context, before time.Time) ([]*ent.ImageEntity, error) {
	return m.expiredImages, nil
}
func (m *mockImageEnt) ListAllImages(ctx context.Context, limit int, offset int) ([]*ent.ImageEntity, int, error) {
	return nil, 0, nil
}
func (m *mockImageEnt) GetGuestStats(ctx context.Context) (int, int64, error) { return 0, 0, nil }
func (m *mockImageEnt) GetUniqueGuestIPCount(ctx context.Context, since time.Time) (int, error) {
	return 0, nil
}

func TestEnsureUsageAllowed_WithinLimit(t *testing.T) {
	userEnt := newMockUserEnt()
	quotaEnt := newMockQuotaEnt()
	planEnt := &mockPlanEnt{plans: make(map[string]*ent.PlanSettingEntity)}

	svc := newService(&Options{
		Config:   &config.Config[Config]{Val: &Config{}},
		userEnt:  userEnt,
		quotaEnt: quotaEnt,
		planEnt:  planEnt,
	})

	ctx := context.Background()
	user, _ := userEnt.CreateUser(ctx, entitiesdto.CreateUser{Plan: "Free"})
	_, _ = quotaEnt.AddToUserQuota(ctx, user.ID, entitiesdto.AddToUserQuota{
		StorageDelta:    1024 * 1024, // 1MB used
		ImageCountDelta: 5,
	})

	err := svc.EnsureUsageAllowed(ctx, user.ID, false)
	if err != nil {
		t.Errorf("expected usage allowed, got error: %v", err)
	}
}

func TestEnsureUsageAllowed_ExceedsLimit(t *testing.T) {
	userEnt := newMockUserEnt()
	quotaEnt := newMockQuotaEnt()
	planEnt := &mockPlanEnt{plans: make(map[string]*ent.PlanSettingEntity)}

	svc := newService(&Options{
		Config:   &config.Config[Config]{Val: &Config{}},
		userEnt:  userEnt,
		quotaEnt: quotaEnt,
		planEnt:  planEnt,
	})

	ctx := context.Background()
	user, _ := userEnt.CreateUser(ctx, entitiesdto.CreateUser{Plan: "Free"})

	// Set usage above Free default limits (Free = 5GB)
	overLimitBytes := int64(6) * 1024 * 1024 * 1024
	_, _ = quotaEnt.AddToUserQuota(ctx, user.ID, entitiesdto.AddToUserQuota{
		StorageDelta:    overLimitBytes,
		ImageCountDelta: 100,
	})

	err := svc.EnsureUsageAllowed(ctx, user.ID, false)
	if err != ErrQuotaAccountLocked {
		t.Errorf("expected ErrQuotaAccountLocked when over quota, got %v", err)
	}
}

func TestPurgeExpiredGuestImages(t *testing.T) {
	imageEnt := &mockImageEnt{
		expiredImages: []*ent.ImageEntity{
			{ID: uuid.New(), StorageID: uuid.New(), UserID: nil},
			{ID: uuid.New(), StorageID: uuid.New(), UserID: nil},
		},
	}

	svc := newService(&Options{
		Config:   &config.Config[Config]{Val: &Config{}},
		tracer:   otel.Tracer("test"),
		imageEnt: imageEnt,
	})

	ctx := context.Background()
	purgedCount, err := svc.PurgeExpiredGuestImages(ctx)
	if err != nil {
		t.Fatalf("PurgeExpiredGuestImages failed: %v", err)
	}
	if purgedCount != 2 {
		t.Errorf("expected 2 purged images, got %d", purgedCount)
	}
}
