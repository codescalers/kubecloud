package workers

import (
	"kubecloud/internal/infrastructure/logger"
	"time"
)

// ReleaseWorkflowLocks periodically scans Redis locks and frees those that belong to finished workflows.
func (w Workers) ReleaseWorkflowLocks() {
	log := logger.ForOperation("locks_worker", "release_workflow_locks")
	ticker := time.NewTicker(w.svc.GetLocksReleaseInterval())
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			keys, err := w.svc.GetAllWorkflowsLocks()
			if err != nil {
				log.Error().Err(err).Msg("failed to list workflow locks")
				continue
			}

			if len(keys) == 0 {
				continue
			}

			for _, key := range keys {
				if err := w.svc.ReleaseLocks(key); err != nil {
					log.Error().Err(err).Str("key", key).Msg("failed to release lock")
				}
			}
		}
	}
}
