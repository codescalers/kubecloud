package workers

import (
	"kubecloud/internal/infrastructure/logger"
	"time"
)

func (w Workers) TrackClusterHealth() {
	log := logger.ForOperation("health_tracker", "track_cluster_health")

	ticker := time.NewTicker(w.svc.GetClusterHealthCheckInterval())
	defer ticker.Stop()
	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			log.Info().Msg("Cluster health check started")

			clusters, err := w.svc.ListAllClusters()
			if err != nil {
				log.Error().Err(err).Msg("failed to list clusters")
				continue
			}

			if len(clusters) == 0 {
				log.Info().Msg("No clusters to check health for")
				continue
			}

			for _, cluster := range clusters {
				if err := w.svc.AsyncTrackClusterHealth(cluster); err != nil {
					log.Error().Err(err).Msg("failed to track cluster health")
					continue
				}
			}
		}
	}
}

func (w Workers) TrackReservedNodeHealth() {
	log := logger.ForOperation("health_tracker", "track_reserved_node_health")

	ticker := time.NewTicker(w.svc.GetReservedNodeHealthCheckInterval())
	defer ticker.Stop()

	for {
		select {
		case <-w.ctx.Done():
			return
		case <-ticker.C:
			log.Info().Msg("Reserved node health check started")

			reservedNodes, err := w.svc.ListAllReservedNodes()
			if err != nil {
				log.Error().Err(err).Msg("failed to get reserved nodes for health check")
				continue
			}

			if len(reservedNodes) == 0 {
				log.Info().Msg("No reserved nodes to check health for")
				continue
			}

			log.Info().Int("count", len(reservedNodes)).Msg("Starting health check for reserved nodes")

			w.svc.CheckNodesWithWorkerPool(reservedNodes)

			log.Info().Int("count", len(reservedNodes)).Msg("Reserved node health check workflows started")
		}
	}
}
