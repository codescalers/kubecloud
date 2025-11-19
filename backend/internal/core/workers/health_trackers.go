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
			logWorkerAudit(
				logger.AuditActionWorkerClusterHealth,
				logger.AuditSeverityInfo,
				map[string]any{
					"event": "cluster_health_check_started",
				},
			)

			clusters, err := w.svc.ListAllClusters()
			if err != nil {
				log.Error().Err(err).Msg("failed to list clusters")
				logWorkerAudit(
					logger.AuditActionWorkerClusterHealth,
					logger.AuditSeverityError,
					map[string]any{
						"reason": "list_clusters_failed",
						"error":  err.Error(),
					},
				)
				continue
			}

			if len(clusters) == 0 {
				log.Info().Msg("No clusters to check health for")
				logWorkerAudit(
					logger.AuditActionWorkerClusterHealth,
					logger.AuditSeverityInfo,
					map[string]any{
						"result": "no_clusters_found",
					},
				)
				continue
			}

			for _, cluster := range clusters {
				if err := w.svc.AsyncTrackClusterHealth(cluster); err != nil {
					log.Error().Err(err).Msg("failed to track cluster health")
					logWorkerAudit(
						logger.AuditActionWorkerClusterHealth,
						logger.AuditSeverityError,
						map[string]any{
							"reason":     "track_cluster_health_failed",
							"cluster_id": cluster.ID,
							"error":      err.Error(),
						},
					)
					continue
				}
				logWorkerAudit(
					logger.AuditActionWorkerClusterHealth,
					logger.AuditSeverityInfo,
					map[string]any{
						"cluster_id": cluster.ID,
						"result":     "health_workflow_started",
					},
				)
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
			logWorkerAudit(
				logger.AuditActionWorkerNodeHealth,
				logger.AuditSeverityInfo,
				map[string]any{
					"event": "reserved_node_health_check_started",
				},
			)

			reservedNodes, err := w.svc.ListAllReservedNodes()
			if err != nil {
				log.Error().Err(err).Msg("failed to get reserved nodes for health check")
				logWorkerAudit(
					logger.AuditActionWorkerNodeHealth,
					logger.AuditSeverityError,
					map[string]any{
						"reason": "list_reserved_nodes_failed",
						"error":  err.Error(),
					},
				)
				continue
			}

			if len(reservedNodes) == 0 {
				log.Info().Msg("No reserved nodes to check health for")
				logWorkerAudit(
					logger.AuditActionWorkerNodeHealth,
					logger.AuditSeverityInfo,
					map[string]any{
						"result": "no_reserved_nodes_found",
					},
				)
				continue
			}

			log.Info().Int("count", len(reservedNodes)).Msg("Starting health check for reserved nodes")
			logWorkerAudit(
				logger.AuditActionWorkerNodeHealth,
				logger.AuditSeverityInfo,
				map[string]any{
					"reserved_nodes": len(reservedNodes),
					"result":         "health_checks_started",
				},
			)

			w.svc.CheckNodesWithWorkerPool(reservedNodes)

			log.Info().Int("count", len(reservedNodes)).Msg("Reserved node health check workflows started")
			logWorkerAudit(
				logger.AuditActionWorkerNodeHealth,
				logger.AuditSeverityInfo,
				map[string]any{
					"reserved_nodes": len(reservedNodes),
					"result":         "health_workflows_dispatched",
				},
			)
		}
	}
}
