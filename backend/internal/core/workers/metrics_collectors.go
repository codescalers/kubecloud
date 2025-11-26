package workers

import (
	"runtime"
	"time"

	"kubecloud/internal/infrastructure/logger"
	"kubecloud/internal/infrastructure/metrics"
)

// CollectGORMMetrics runs a worker that periodically updates GORM metrics
func (w Workers) CollectGORMMetrics() {
	log := logger.ForOperation("metrics", "gorm_collector")
	ticker := time.NewTicker(metrics.MetricsCollectorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			log.Info().Msg("GORM metrics collector stopped")
			return
		case <-ticker.C:
			w.metrics.UpdateGORMMetrics(w.db)
		}
	}
}

// CollectGoRuntimeMetrics runs a worker that periodically collects Go runtime metrics
func (w Workers) CollectGoRuntimeMetrics() {
	log := logger.ForOperation("metrics", "runtime_collector")
	var memStats runtime.MemStats
	ticker := time.NewTicker(metrics.MetricsCollectorInterval)
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			log.Info().Msg("Go runtime metrics collector stopped")
			return
		case <-ticker.C:
			runtime.ReadMemStats(&memStats)
		}
	}
}
