package entities

import (
	"context"
	"time"

	"github.com/google/uuid"
	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"
)

var _ entitiesinf.NotificationEntity = (*Service)(nil)

func (s *Service) CreateNotification(ctx context.Context, in entitiesdto.CreateNotification) (*ent.NotificationEntity, error) {
	targetRole := in.TargetRole
	if targetRole == "" {
		if in.UserID != nil {
			targetRole = "user"
		} else {
			targetRole = "all"
		}
	}

	data := &ent.NotificationEntity{
		UserID:     in.UserID,
		TargetRole: targetRole,
		Type:       in.Type,
		Title:      in.Title,
		Message:    in.Message,
		Link:       in.Link,
		Metadata:   in.Metadata,
		IsRead:     false,
		CreatedAt:  time.Now(),
	}

	_, err := s.db.NewInsert().Model(data).Exec(ctx)
	if err != nil {
		return nil, err
	}
	return data, nil
}

func (s *Service) GetNotificationByID(ctx context.Context, id uuid.UUID) (*ent.NotificationEntity, error) {
	var notif ent.NotificationEntity
	err := s.db.NewSelect().Model(&notif).Where("id = ?", id).Scan(ctx)
	if err != nil {
		return nil, err
	}
	return &notif, nil
}

func (s *Service) ListNotifications(ctx context.Context, filter entitiesdto.NotificationFilter) ([]*ent.NotificationEntity, int, error) {
	var notifications []*ent.NotificationEntity

	q := s.db.NewSelect().Model(&notifications)

	if filter.IsAdmin {
		// Admin sees admin notifications and broadcasts
		q = q.Where("target_role IN ('admin', 'all')")
	} else if filter.UserID != nil {
		// User sees notifications targeted to their user_id OR broadcast to 'all' or 'user'
		q = q.Where("user_id = ? OR (user_id IS NULL AND target_role IN ('all', 'user'))", *filter.UserID)
	}

	if filter.UnreadOnly {
		q = q.Where("is_read = false")
	}

	if filter.Type != "" && filter.Type != "all" {
		q = q.Where("type = ?", filter.Type)
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := (filter.Page - 1) * limit
	if offset < 0 {
		offset = 0
	}

	count, err := q.Order("created_at DESC").Limit(limit).Offset(offset).ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}

	return notifications, count, nil
}

func (s *Service) GetUnreadCount(ctx context.Context, userID *uuid.UUID, targetRole string) (int, error) {
	q := s.db.NewSelect().Model((*ent.NotificationEntity)(nil)).Where("is_read = false")

	if targetRole == "admin" {
		q = q.Where("target_role IN ('admin', 'all')")
	} else if userID != nil {
		q = q.Where("user_id = ? OR (user_id IS NULL AND target_role IN ('all', 'user'))", *userID)
	}

	return q.Count(ctx)
}

func (s *Service) MarkAsRead(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	now := time.Now()
	q := s.db.NewUpdate().
		Model((*ent.NotificationEntity)(nil)).
		Set("is_read = true").
		Set("read_at = ?", now).
		Where("id = ?", id)

	if userID != nil {
		q = q.Where("user_id = ? OR user_id IS NULL", *userID)
	}

	_, err := q.Exec(ctx)
	return err
}

func (s *Service) MarkAllAsRead(ctx context.Context, userID *uuid.UUID, targetRole string) error {
	now := time.Now()
	q := s.db.NewUpdate().
		Model((*ent.NotificationEntity)(nil)).
		Set("is_read = true").
		Set("read_at = ?", now).
		Where("is_read = false")

	if targetRole == "admin" {
		q = q.Where("target_role IN ('admin', 'all')")
	} else if userID != nil {
		q = q.Where("user_id = ? OR (user_id IS NULL AND target_role IN ('all', 'user'))", *userID)
	}

	_, err := q.Exec(ctx)
	return err
}

func (s *Service) DeleteNotification(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	q := s.db.NewDelete().
		Model((*ent.NotificationEntity)(nil)).
		Where("id = ?", id)

	if userID != nil {
		q = q.Where("user_id = ?", *userID)
	}

	_, err := q.Exec(ctx)
	return err
}
