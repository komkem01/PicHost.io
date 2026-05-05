package storage

import (
	"context"
	"time"

	"go.opentelemetry.io/otel/trace"
	entitiesdto "pichost.io/app/modules/entities/dto"
	entitiesinf "pichost.io/app/modules/entities/inf"
	imagemod "pichost.io/app/modules/image"

	"github.com/google/uuid"
)

type Controller struct {
	tracer   trace.Tracer
	svc      *Service
	imgSvc   *imagemod.Service
	auditEnt entitiesinf.AuditEntity
}

func newController(trace trace.Tracer, svc *Service) *Controller {
	return &Controller{
		tracer: trace,
		svc:    svc,
	}
}

// recordAudit writes an audit event asynchronously (fire-and-forget).
func (c *Controller) recordAudit(
	action string,
	status string,
	userID *uuid.UUID,
	resourceType *string,
	resourceID *uuid.UUID,
	ip string,
	ua string,
	meta map[string]any,
	errCode *string,
) {
	if c.auditEnt == nil {
		return
	}
	var ipPtr, uaPtr *string
	if ip != "" {
		v := ip
		ipPtr = &v
	}
	if ua != "" {
		v := ua
		uaPtr = &v
	}
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = c.auditEnt.CreateAuditLog(ctx, entitiesdto.CreateAuditLog{
			UserID:       userID,
			Action:       action,
			ResourceType: resourceType,
			ResourceID:   resourceID,
			IPAddress:    ipPtr,
			UserAgent:    uaPtr,
			Metadata:     meta,
			Status:       status,
			ErrorCode:    errCode,
		})
	}()
}

func strPtr(s string) *string        { return &s }
func uuidPtr(u uuid.UUID) *uuid.UUID { return &u }
