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
