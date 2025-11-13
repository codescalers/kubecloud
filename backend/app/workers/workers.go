package workers

import (
	"context"
	"kubecloud/app/services"
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
