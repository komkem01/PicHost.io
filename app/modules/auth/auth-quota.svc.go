package auth

import (
	"context"
	"database/sql"
	"errors"

	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
)

type QuotaResult struct {
	Plan               ent.PlanType `json:"plan"`
	UsedStorageBytes   int64        `json:"used_storage_bytes"`
	StorageLimitBytes  int64        `json:"storage_limit_bytes"` // -1 = unlimited
	ImageCount         int          `json:"image_count"`
	MaxImages          int          `json:"max_images"` // 0 = unlimited
	FileSizeLimitBytes int64        `json:"file_size_limit_bytes"`
	AllowPrivate       bool         `json:"allow_private"`
	TotalViews         int64        `json:"total_views"`
}

func (s *Service) GetQuota(ctx context.Context, userID uuid.UUID) (*QuotaResult, error) {
	user, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrAuthUnauthorized
	}

	limits := ent.GetPlanLimits(user.Plan)
	if setting, pErr := s.planEnt.GetPlanSettingByKey(ctx, string(user.Plan)); pErr == nil {
		limits.StorageBytes = setting.StorageLimitBytes
		limits.FileSizeBytes = int64(setting.MaxUploadMB) * 1024 * 1024
		limits.MaxImages = setting.ImageLimit
		limits.AllowPrivate = setting.AllowPrivate
	} else if !errors.Is(pErr, sql.ErrNoRows) {
		return nil, pErr
	}

	var usedStorage int64
	var imageCount int

	quota, err := s.quotaEnt.GetUserQuota(ctx, userID)
	if err != nil {
		if !errors.Is(err, sql.ErrNoRows) {
			return nil, err
		}
	} else {
		usedStorage = quota.UsedStorageBytes
		imageCount = quota.ImageCount
	}

	var totalViews int64
	if s.imageEnt != nil {
		if views, vErr := s.imageEnt.GetTotalImageViewsByUserID(ctx, userID); vErr == nil {
			totalViews = views
		}
	}

	return &QuotaResult{
		Plan:               user.Plan,
		UsedStorageBytes:   usedStorage,
		StorageLimitBytes:  limits.StorageBytes,
		ImageCount:         imageCount,
		MaxImages:          limits.MaxImages,
		FileSizeLimitBytes: limits.FileSizeBytes,
		AllowPrivate:       limits.AllowPrivate,
		TotalViews:         totalViews,
	}, nil
}
