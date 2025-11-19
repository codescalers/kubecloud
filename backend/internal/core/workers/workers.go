package workers

import (
	"context"

	"kubecloud/internal/core/services"
	"kubecloud/internal/infrastructure/logger"
)

type Workers struct {
	ctx context.Context
	svc services.WorkerService
}

func NewWorkers(ctx context.Context, svc services.WorkerService) Workers {
	return Workers{
		ctx: ctx,
		svc: svc,
	}
}

func logWorkerAudit(action logger.AuditActionType, severity logger.AuditSeverity, metadata map[string]any) {
	opts := []logger.AuditEntryOption{
		logger.WithAuditSeverity(severity),
	}
	if len(metadata) > 0 {
		opts = append(opts, logger.WithAuditActionMetadata(metadata))
	}
	logger.LogAudit(logger.AuditActorSystem, action, "", "", opts...)
}
