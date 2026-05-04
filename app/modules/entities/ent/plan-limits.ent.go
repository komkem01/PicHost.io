package ent

// PlanLimits defines usage constraints for a given plan tier.
type PlanLimits struct {
	StorageBytes   int64    // total quota (-1 = unlimited)
	FileSizeBytes  int64    // max bytes per single upload (-1 = unlimited)
	MaxImages      int      // max stored images (0 = unlimited)
	RetentionHours int      // hours before auto-delete (0 = forever)
	AllowPrivate   bool     // allow private (non-public) images
	AllowedMIMEs   []string // allowed MIME types
}

// GuestPlan is the "no account" tier.
var GuestPlan = PlanLimits{
	StorageBytes:   50 * 1024 * 1024, // 50 MB
	FileSizeBytes:  5 * 1024 * 1024,  // 5 MB per file
	MaxImages:      10,
	RetentionHours: 24,
	AllowPrivate:   false,
	AllowedMIMEs:   []string{"image/jpeg", "image/png"},
}

// planLimitsMap maps each PlanType to its limits.
var planLimitsMap = map[PlanType]PlanLimits{
	PlanTypeFree: {
		StorageBytes:   500 * 1024 * 1024, // 500 MB
		FileSizeBytes:  10 * 1024 * 1024,  // 10 MB per file
		MaxImages:      200,
		RetentionHours: 0,
		AllowPrivate:   false,
		AllowedMIMEs:   []string{"image/jpeg", "image/png", "image/webp"},
	},
	PlanTypeBasic: {
		StorageBytes:   10 * 1024 * 1024 * 1024, // 10 GB
		FileSizeBytes:  20 * 1024 * 1024,        // 20 MB per file
		MaxImages:      0,
		RetentionHours: 0,
		AllowPrivate:   true,
		AllowedMIMEs:   []string{"image/jpeg", "image/png", "image/webp", "image/gif", "image/avif"},
	},
	PlanTypePro: {
		StorageBytes:   100 * 1024 * 1024 * 1024, // 100 GB
		FileSizeBytes:  50 * 1024 * 1024,         // 50 MB per file
		MaxImages:      0,
		RetentionHours: 0,
		AllowPrivate:   true,
		AllowedMIMEs:   []string{"image/jpeg", "image/png", "image/webp", "image/gif", "image/avif", "image/bmp", "image/tiff", "image/heic", "image/heif"},
	},
	PlanTypeEnterprise: {
		StorageBytes:   -1, // unlimited
		FileSizeBytes:  -1,
		MaxImages:      0,
		RetentionHours: 0,
		AllowPrivate:   true,
		AllowedMIMEs:   nil, // nil = all formats allowed
	},
}

// GetPlanLimits returns the limits for the given plan. Defaults to Free if unknown.
func GetPlanLimits(plan PlanType) PlanLimits {
	if limits, ok := planLimitsMap[plan]; ok {
		return limits
	}
	return planLimitsMap[PlanTypeFree]
}
