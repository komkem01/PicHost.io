package auth

import (
	"context"
	"database/sql"
	"errors"

	"pichost.io/app/modules/entities/ent"

	"github.com/google/uuid"
)

// QuotaResult holds the user's current plan, usage, and limits for a single response.
type QuotaResult struct {
	Plan               ent.PlanType
	UsedStorageBytes   int64
	StorageLimitBytes  int64 // -1 = unlimited
	ImageCount         int
	MaxImages          int // 0 = unlimited
	FileSizeLimitBytes int64
	AllowPrivate       bool
}

// GetQuota returns the authenticated user's plan limits alongside actual usage.
func (s *Service) GetQuota(ctx context.Context, userID uuid.UUID) (*QuotaResult, error) {
	user, err := s.user.GetUserByID(ctx, userID)
	if err != nil {
		return nil, ErrAuthUnauthorized
	}

	limits := ent.GetPlanLimits(user.Plan)

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

	return &QuotaResult{
		Plan:               user.Plan,
		UsedStorageBytes:   usedStorage,
		StorageLimitBytes:  limits.StorageBytes,
		ImageCount:         imageCount,
		MaxImages:          limits.MaxImages,
		FileSizeLimitBytes: limits.FileSizeBytes,
		AllowPrivate:       limits.AllowPrivate,
	}, nil
}
