package entitiesdto

type UpsertPlanSetting struct {
	PlanKey           string
	DisplayName       string
	MonthlyPriceTHB   int
	StorageLimitBytes int64
	ImageLimit        int
	MaxUploadMB       int
	IsEnabled         bool
	AllowPrivate      bool
	CustomDomain      bool
	APIAccess         bool
	PrioritySupport   bool
	NoAds             bool
	WatermarkRemoval  bool
}
