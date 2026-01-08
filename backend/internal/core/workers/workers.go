package workers

import (
	"context"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/services"
	"kubecloud/internal/infrastructure/metrics"
)

type Workers struct {
	ctx            context.Context
	svc            services.WorkerService
	billingService services.BillingService
	metrics        *metrics.Metrics
	db             models.DB
}

func NewWorkers(ctx context.Context, svc services.WorkerService, billingService services.BillingService, metrics *metrics.Metrics, db models.DB) Workers {
	return Workers{
		ctx:     ctx,
		svc:     svc,
		metrics: metrics,
		db:      db,
	}
}
