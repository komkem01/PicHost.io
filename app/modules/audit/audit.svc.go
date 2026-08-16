package audit

import (
	"context"
	"log/slog"
	"time"

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

func (s *Service) PurgeOldLogs(ctx context.Context, retentionDays int) (int64, error) {
	if retentionDays <= 0 {
		retentionDays = 30
	}
	cutoff := time.Now().AddDate(0, 0, -retentionDays)
	return s.auditEnt.DeleteAuditLogsOlderThan(ctx, cutoff)
}

func (s *Service) StartRetentionCleaner(ctx context.Context, retentionDays int) {
	go func() {
		// Initial cleanup on startup
		cleanCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		if deleted, err := s.PurgeOldLogs(cleanCtx, retentionDays); err == nil && deleted > 0 {
			slog.Info("Purged expired audit logs", "deleted", deleted, "retention_days", retentionDays)
		} else if err != nil {
			slog.Error("Failed to purge expired audit logs", "err", err)
		}
		cancel()

		// Run periodic daily cleanup ticker
		ticker := time.NewTicker(24 * time.Hour)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				runCtx, runCancel := context.WithTimeout(context.Background(), 30*time.Second)
				if deleted, err := s.PurgeOldLogs(runCtx, retentionDays); err == nil && deleted > 0 {
					slog.Info("Daily audit log cleanup executed", "deleted", deleted)
				} else if err != nil {
					slog.Error("Daily audit log cleanup failed", "err", err)
				}
				runCancel()
			}
		}
	}()
}

