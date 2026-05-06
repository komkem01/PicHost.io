package entities

import (
	"context"
	"strings"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
)

func normalizePlanKey(key string) string {
	return strings.ToLower(strings.TrimSpace(key))
}

func (s *Service) ListPlanSettings(ctx context.Context) ([]*ent.PlanSettingEntity, error) {
	var rows []*ent.PlanSettingEntity
	err := s.db.NewSelect().
		Model(&rows).
		OrderExpr("monthly_price_thb ASC, display_name ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (s *Service) GetPlanSettingByKey(ctx context.Context, key string) (*ent.PlanSettingEntity, error) {
	var row ent.PlanSettingEntity
	err := s.db.NewSelect().
		Model(&row).
		Where("plan_key = ?", normalizePlanKey(key)).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (s *Service) UpsertPlanSetting(ctx context.Context, input entitiesdto.UpsertPlanSetting) (*ent.PlanSettingEntity, error) {
	now := time.Now()
	row := &ent.PlanSettingEntity{
		PlanKey:           normalizePlanKey(input.PlanKey),
		DisplayName:       input.DisplayName,
		MonthlyPriceTHB:   input.MonthlyPriceTHB,
		StorageLimitBytes: input.StorageLimitBytes,
		ImageLimit:        input.ImageLimit,
		MaxUploadMB:       input.MaxUploadMB,
		IsEnabled:         input.IsEnabled,
		AllowPrivate:      input.AllowPrivate,
		CustomDomain:      input.CustomDomain,
		APIAccess:         input.APIAccess,
		PrioritySupport:   input.PrioritySupport,
		NoAds:             input.NoAds,
		WatermarkRemoval:  input.WatermarkRemoval,
		CreatedAt:         now,
		UpdatedAt:         now,
	}

	_, err := s.db.NewInsert().
		Model(row).
		On("CONFLICT (plan_key) DO UPDATE").
		Set("display_name = EXCLUDED.display_name").
		Set("monthly_price_thb = EXCLUDED.monthly_price_thb").
		Set("storage_limit_bytes = EXCLUDED.storage_limit_bytes").
		Set("image_limit = EXCLUDED.image_limit").
		Set("max_upload_mb = EXCLUDED.max_upload_mb").
		Set("is_enabled = EXCLUDED.is_enabled").
		Set("allow_private = EXCLUDED.allow_private").
		Set("custom_domain = EXCLUDED.custom_domain").
		Set("api_access = EXCLUDED.api_access").
		Set("priority_support = EXCLUDED.priority_support").
		Set("no_ads = EXCLUDED.no_ads").
		Set("watermark_removal = EXCLUDED.watermark_removal").
		Set("updated_at = EXCLUDED.updated_at").
		Exec(ctx)
	if err != nil {
		return nil, err
	}

	return s.GetPlanSettingByKey(ctx, row.PlanKey)
}

func (s *Service) DeletePlanSettingByKey(ctx context.Context, key string) error {
	_, err := s.db.NewDelete().
		Model((*ent.PlanSettingEntity)(nil)).
		Where("plan_key = ?", normalizePlanKey(key)).
		Exec(ctx)
	return err
}
