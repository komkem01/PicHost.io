package entities

import (
	"context"
	"encoding/json"
	"time"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"
)

var _ entitiesinf.AuditEntity = (*Service)(nil)

func (s *Service) CreateAuditLog(ctx context.Context, in entitiesdto.CreateAuditLog) error {
	status := in.Status
	if status == "" {
		status = "success"
	}

	var metaBytes json.RawMessage
	if len(in.Metadata) > 0 {
		b, err := json.Marshal(in.Metadata)
		if err == nil {
			metaBytes = b
		}
	}

	data := &ent.AuditLogEntity{
		UserID:       in.UserID,
		Action:       in.Action,
		ResourceType: in.ResourceType,
		ResourceID:   in.ResourceID,
		IPAddress:    in.IPAddress,
		UserAgent:    in.UserAgent,
		Metadata:     metaBytes,
		Status:       status,
		ErrorCode:    in.ErrorCode,
		CreatedAt:    time.Now(),
	}

	_, err := s.db.NewInsert().Model(data).Exec(ctx)
	return err
}

func (s *Service) ListAuditLogs(ctx context.Context, filter entitiesdto.ListAuditLogsFilter) ([]*ent.AuditLogEntity, int, error) {
	var logs []*ent.AuditLogEntity

	q := s.db.NewSelect().Model(&logs)

	if filter.UserID != nil {
		q = q.Where("user_id = ?", *filter.UserID)
	}
	if filter.Action != nil && *filter.Action != "" {
		q = q.Where("action LIKE ?", "%"+*filter.Action+"%")
	}
	if filter.Status != nil && *filter.Status != "" {
		q = q.Where("status = ?", *filter.Status)
	}
	if filter.FromDate != nil && *filter.FromDate != "" {
		if t, err := time.Parse(time.RFC3339, *filter.FromDate); err == nil {
			q = q.Where("created_at >= ?", t)
		} else if t, err := time.Parse("2006-01-02", *filter.FromDate); err == nil {
			q = q.Where("created_at >= ?", t)
		}
	}
	if filter.ToDate != nil && *filter.ToDate != "" {
		if t, err := time.Parse(time.RFC3339, *filter.ToDate); err == nil {
			q = q.Where("created_at <= ?", t)
		} else if t, err := time.Parse("2006-01-02", *filter.ToDate); err == nil {
			// end of the day
			q = q.Where("created_at <= ?", t.Add(24*time.Hour-time.Nanosecond))
		}
	}

	limit := filter.Limit
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}
	offset := filter.Offset
	if offset < 0 {
		offset = 0
	}

	count, err := q.Order("created_at DESC").Limit(limit).Offset(offset).ScanAndCount(ctx)
	if err != nil {
		return nil, 0, err
	}

	return logs, count, nil
}

func (s *Service) DeleteAuditLogsOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	res, err := s.db.NewDelete().
		Model((*ent.AuditLogEntity)(nil)).
		Where("created_at < ?", cutoff).
		Exec(ctx)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return n, nil
}


