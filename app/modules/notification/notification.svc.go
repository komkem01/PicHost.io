package notification

import (
	"context"

	"github.com/google/uuid"
	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"
)

type Service struct {
	notifEnt entitiesinf.NotificationEntity
}

func newService(notifEnt entitiesinf.NotificationEntity) *Service {
	return &Service{
		notifEnt: notifEnt,
	}
}

func (s *Service) Create(ctx context.Context, in entitiesdto.CreateNotification) (*ent.NotificationEntity, error) {
	return s.notifEnt.CreateNotification(ctx, in)
}

func (s *Service) List(ctx context.Context, filter entitiesdto.NotificationFilter) ([]*ent.NotificationEntity, int, error) {
	return s.notifEnt.ListNotifications(ctx, filter)
}

func (s *Service) GetUnreadCount(ctx context.Context, userID *uuid.UUID, targetRole string) (int, error) {
	return s.notifEnt.GetUnreadCount(ctx, userID, targetRole)
}

func (s *Service) MarkAsRead(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	return s.notifEnt.MarkAsRead(ctx, id, userID)
}

func (s *Service) MarkAllAsRead(ctx context.Context, userID *uuid.UUID, targetRole string) error {
	return s.notifEnt.MarkAllAsRead(ctx, userID, targetRole)
}

func (s *Service) Delete(ctx context.Context, id uuid.UUID, userID *uuid.UUID) error {
	return s.notifEnt.DeleteNotification(ctx, id, userID)
}
