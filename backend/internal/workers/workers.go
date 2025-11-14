package workers

import (
	"context"
	"kubecloud/internal/services"
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
