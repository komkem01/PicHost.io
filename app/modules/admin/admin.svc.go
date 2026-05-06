package admin

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"

	"github.com/google/uuid"
)

var errInvalidPlan = errors.New("invalid plan")
var errPlanInUse = errors.New("plan is currently assigned to users")
var errPlanNotFound = errors.New("plan not found")

type Service struct {
	user     entitiesinf.UserEntity
	quota    entitiesinf.UserQuotaEntity
	image    entitiesinf.ImageEntity
	planConf entitiesinf.PlanSettingEntity
}

func newService(
	user entitiesinf.UserEntity,
	quota entitiesinf.UserQuotaEntity,
	image entitiesinf.ImageEntity,
	planConf entitiesinf.PlanSettingEntity,
) *Service {
	return &Service{user: user, quota: quota, image: image, planConf: planConf}
}

type AdminPlanSetting struct {
	PlanKey           string    `json:"plan_key"`
	DisplayName       string    `json:"display_name"`
	MonthlyPriceTHB   int       `json:"monthly_price_thb"`
	StorageLimitBytes int64     `json:"storage_limit_bytes"`
	ImageLimit        int       `json:"image_limit"`
	MaxUploadMB       int       `json:"max_upload_mb"`
	IsEnabled         bool      `json:"is_enabled"`
	AllowPrivate      bool      `json:"allow_private"`
	CustomDomain      bool      `json:"custom_domain"`
	APIAccess         bool      `json:"api_access"`
	PrioritySupport   bool      `json:"priority_support"`
	NoAds             bool      `json:"no_ads"`
	WatermarkRemoval  bool      `json:"watermark_removal"`
	UpdatedAt         time.Time `json:"updated_at"`
}

func toAdminPlanSetting(row *ent.PlanSettingEntity) AdminPlanSetting {
	return AdminPlanSetting{
		PlanKey:           row.PlanKey,
		DisplayName:       row.DisplayName,
		MonthlyPriceTHB:   row.MonthlyPriceTHB,
		StorageLimitBytes: row.StorageLimitBytes,
		ImageLimit:        row.ImageLimit,
		MaxUploadMB:       row.MaxUploadMB,
		IsEnabled:         row.IsEnabled,
		AllowPrivate:      row.AllowPrivate,
		CustomDomain:      row.CustomDomain,
		APIAccess:         row.APIAccess,
		PrioritySupport:   row.PrioritySupport,
		NoAds:             row.NoAds,
		WatermarkRemoval:  row.WatermarkRemoval,
		UpdatedAt:         row.UpdatedAt,
	}
}

func (s *Service) ListPlanSettings(ctx context.Context) ([]AdminPlanSetting, error) {
	rows, err := s.planConf.ListPlanSettings(ctx)
	if err != nil {
		return nil, err
	}
	out := make([]AdminPlanSetting, 0, len(rows))
	for _, row := range rows {
		out = append(out, toAdminPlanSetting(row))
	}
	return out, nil
}

func (s *Service) GetPlanSetting(ctx context.Context, planKey string) (*AdminPlanSetting, error) {
	row, err := s.planConf.GetPlanSettingByKey(ctx, planKey)
	if err != nil {
		return nil, err
	}
	res := toAdminPlanSetting(row)
	return &res, nil
}

func (s *Service) UpsertPlanSetting(ctx context.Context, in entitiesdto.UpsertPlanSetting) (*AdminPlanSetting, error) {
	row, err := s.planConf.UpsertPlanSetting(ctx, in)
	if err != nil {
		return nil, err
	}
	res := toAdminPlanSetting(row)
	return &res, nil
}

func (s *Service) DeletePlanSetting(ctx context.Context, planKey string) error {
	normalized := strings.ToLower(strings.TrimSpace(planKey))
	if normalized == "" {
		return errPlanNotFound
	}

	if _, err := s.planConf.GetPlanSettingByKey(ctx, normalized); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return errPlanNotFound
		}
		return err
	}

	users, err := s.user.GetListUser(ctx)
	if err != nil {
		return err
	}
	for _, u := range users {
		if strings.EqualFold(string(u.Plan), normalized) {
			return errPlanInUse
		}
	}

	return s.planConf.DeletePlanSettingByKey(ctx, normalized)
}

// --- Stats ---

type DashboardStats struct {
	TotalUsers    int            `json:"total_users"`
	ActiveUsers   int            `json:"active_users"`
	GuestUsers    int            `json:"guest_users"`
	PlanBreakdown map[string]int `json:"plan_breakdown"`
}

func (s *Service) GetDashboardStats(ctx context.Context) (*DashboardStats, error) {
	users, err := s.user.GetListUser(ctx)
	if err != nil {
		return nil, err
	}
	stats := &DashboardStats{PlanBreakdown: make(map[string]int)}
	for _, u := range users {
		stats.TotalUsers++
		if u.IsActive {
			stats.ActiveUsers++
		}
		if u.IsGuest {
			stats.GuestUsers++
		}
		stats.PlanBreakdown[string(u.Plan)]++
	}
	return stats, nil
}

// --- User management ---

type AdminUser struct {
	ID        uuid.UUID `json:"id"`
	Email     *string   `json:"email"`
	Username  *string   `json:"username"`
	Plan      string    `json:"plan"`
	IsActive  bool      `json:"is_active"`
	IsGuest   bool      `json:"is_guest"`
	IsAdmin   bool      `json:"is_admin"`
	CreatedAt time.Time `json:"created_at"`

	// Quota — populated when requested
	UsedStorageBytes *int64 `json:"used_storage_bytes,omitempty"`
	ImageCount       *int   `json:"image_count,omitempty"`
}

func toAdminUser(u *ent.UserEntity) AdminUser {
	return AdminUser{
		ID:        u.ID,
		Email:     u.Email,
		Username:  u.Username,
		Plan:      string(u.Plan),
		IsActive:  u.IsActive,
		IsGuest:   u.IsGuest,
		IsAdmin:   u.IsAdmin,
		CreatedAt: u.CreatedAt,
	}
}

func (s *Service) ListUsers(ctx context.Context) ([]AdminUser, error) {
	users, err := s.user.GetListUser(ctx)
	if err != nil {
		return nil, err
	}
	result := make([]AdminUser, 0, len(users))
	for _, u := range users {
		au := toAdminUser(u)
		// Attach quota inline (best-effort, ignore errors)
		if q, qErr := s.quota.GetUserQuota(ctx, u.ID); qErr == nil {
			au.UsedStorageBytes = &q.UsedStorageBytes
			au.ImageCount = &q.ImageCount
		}
		result = append(result, au)
	}
	return result, nil
}

func (s *Service) GetUser(ctx context.Context, id uuid.UUID) (*AdminUser, error) {
	u, err := s.user.GetUserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	au := toAdminUser(u)
	if q, qErr := s.quota.GetUserQuota(ctx, id); qErr == nil {
		au.UsedStorageBytes = &q.UsedStorageBytes
		au.ImageCount = &q.ImageCount
	}
	return &au, nil
}

func (s *Service) SetUserPlan(ctx context.Context, id uuid.UUID, plan string) error {
	pt := ent.PlanType(plan)
	switch pt {
	case ent.PlanTypeFree, ent.PlanTypeBasic, ent.PlanTypePro, ent.PlanTypeEnterprise:
	default:
		return errInvalidPlan
	}
	p := string(pt)
	_, err := s.user.UpdateUserPlan(ctx, id, entitiesdto.UpdateUserPlan{Plan: &p})
	return err
}

func (s *Service) SetUserActive(ctx context.Context, id uuid.UUID, active bool) error {
	return s.user.SetUserActive(ctx, id, active)
}

func (s *Service) UpdateUserProfile(ctx context.Context, id uuid.UUID, email *string, username *string) (*AdminUser, error) {
	u, err := s.user.UpdateUserProfile(ctx, id, entitiesdto.UpdateUserProfile{
		Email:    email,
		Username: username,
	})
	if err != nil {
		return nil, err
	}
	au := toAdminUser(u)
	return &au, nil
}

func (s *Service) SetUserAdmin(ctx context.Context, id uuid.UUID, isAdmin bool) error {
	return s.user.SetUserAdmin(ctx, id, isAdmin)
}

func (s *Service) DeleteUser(ctx context.Context, id uuid.UUID) error {
	return s.user.DeleteUser(ctx, id)
}
