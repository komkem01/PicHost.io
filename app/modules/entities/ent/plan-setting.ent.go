package ent

import (
	"time"

	"github.com/uptrace/bun"
)

type PlanSettingEntity struct {
	bun.BaseModel `bun:"table:plan_settings,alias:ps"`

	PlanKey           string    `bun:"plan_key,pk"`
	DisplayName       string    `bun:"display_name,notnull"`
	MonthlyPriceTHB   int       `bun:"monthly_price_thb,notnull"`
	StorageLimitBytes int64     `bun:"storage_limit_bytes,notnull"`
	ImageLimit        int       `bun:"image_limit,notnull"`
	MaxUploadMB       int       `bun:"max_upload_mb,notnull"`
	IsEnabled         bool      `bun:"is_enabled,notnull"`
	AllowPrivate      bool      `bun:"allow_private,notnull"`
	CustomDomain      bool      `bun:"custom_domain,notnull"`
	APIAccess         bool      `bun:"api_access,notnull"`
	PrioritySupport   bool      `bun:"priority_support,notnull"`
	NoAds             bool      `bun:"no_ads,notnull"`
	WatermarkRemoval  bool      `bun:"watermark_removal,notnull"`
	CreatedAt         time.Time `bun:"created_at,notnull"`
	UpdatedAt         time.Time `bun:"updated_at,notnull"`
}
