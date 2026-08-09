package audit

import (
	"context"

	entitiesdto "pichost.io/app/modules/entities/dto"
	"pichost.io/app/modules/entities/ent"
	entitiesinf "pichost.io/app/modules/entities/inf"
)

type Service struct {
	auditEnt entitiesinf.AuditEntity
}

func newService(auditEnt entitiesinf.AuditEntity) *Service {
	return &Service{auditEnt: auditEnt}
}

func (s *Service) RecordAudit(ctx context.Context, in entitiesdto.CreateAuditLog) error {
	return s.auditEnt.CreateAuditLog(ctx, in)
}

func (s *Service) ListAuditLogs(ctx context.Context, filter entitiesdto.ListAuditLogsFilter) ([]*ent.AuditLogEntity, int, error) {
	return s.auditEnt.ListAuditLogs(ctx, filter)
}
