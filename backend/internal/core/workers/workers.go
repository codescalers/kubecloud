package workers

import (
	"context"
	"kubecloud/internal/core/services"
)

type Workers struct {
	ctx            context.Context
	svc            services.WorkerService
	billingService services.BillingService
}

func NewWorkers(ctx context.Context, svc services.WorkerService, billingService services.BillingService) Workers {
	return Workers{
		ctx: ctx,
		svc: svc,

		billingService: billingService,
	}
}
