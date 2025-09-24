package app

import (
	"context"
	"fmt"
	"kubecloud/internal/constants"
	"kubecloud/internal/logger"
	"kubecloud/internal/notification"
	"kubecloud/models"
	"strings"
	"sync"
	"time"

	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
	"github.com/xmonader/ewf"
)

type NodeHealthResult struct {
	userID          int
	unhealthyNodeID uint32
}

func (h *Handler) TrackClusterHealth() {

	interval := time.Duration(h.config.ClusterHealthCheckIntervalInHours) * time.Hour

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		logger.GetLogger().Info().Msg("Cluster health check test started")
		clusters, err := h.db.ListAllClusters()
		if err != nil {
			logger.GetLogger().Error().Err(err)
			continue
		}

		if len(clusters) == 0 {
			logger.GetLogger().Info().Msg("No clusters to check health for")
			continue
		}

		for _, cluster := range clusters {

			wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowTrackClusterHealth)
			if err != nil {
				logger.GetLogger().Error().
					Err(err).
					Msg("Failed to create health tracking workflow")
				continue
			}
			cl, err := cluster.GetClusterResult()
			if err != nil {
				logger.GetLogger().Error().
					Err(err).
					Msg("Failed to get cluster result during health tracking")
				continue
			}
			wf.State = ewf.State{
				"cluster": cl,
				"config": map[string]interface{}{
					"user_id": cluster.UserID,
				},
			}

			h.ewfEngine.RunAsync(context.Background(), wf)
		}

	}
}

func (h *Handler) TrackReservedNodeHealth(notificationService *notification.NotificationService, grid proxy.Client) {
	interval := time.Duration(h.config.ReservedNodeHealthCheckIntervalInHours) * time.Hour

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		logger.GetLogger().Info().Msg("Reserved node health check started")

		reservedNodes, err := h.db.ListAllReservedNodes()
		if err != nil {
			logger.GetLogger().Error().Err(err).Msg("Failed to get reserved nodes for health check")
			continue
		}

		if len(reservedNodes) == 0 {
			logger.GetLogger().Info().Msg("No reserved nodes to check health for")
			continue
		}

		logger.GetLogger().Info().
			Int("count", len(reservedNodes)).
			Msg("Starting health check for reserved nodes")

		h.checkNodesWithWorkerPool(reservedNodes, grid, notificationService)

		logger.GetLogger().Info().
			Int("count", len(reservedNodes)).
			Msg("Reserved node health check workflows started")
	}
}

// checkNodesWithWorkerPool uses a worker pool to check node health concurrently
func (h *Handler) checkNodesWithWorkerPool(reservedNodes []models.UserNodes, grid proxy.Client, notificationService *notification.NotificationService) {
	timeout := time.Duration(h.config.ReservedNodeHealthCheckTimeoutInMinutes) * time.Minute
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	workerCount := h.config.ReservedNodeHealthCheckWorkersNum
	if workerCount > len(reservedNodes) {
		workerCount = len(reservedNodes)
	}

	jobs := make(chan models.UserNodes, len(reservedNodes))
	results := make(chan NodeHealthResult, len(reservedNodes))

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go h.healthCheckWorker(ctx, &wg, jobs, results, grid)
	}

	go func() {
		defer close(jobs)
		for _, userNode := range reservedNodes {
			select {
			case <-ctx.Done():
				logger.GetLogger().Info().Msg("Context done, stopping health check worker")
				return
			case jobs <- userNode:
			}
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	userNodes := make(map[int][]uint32)
	for res := range results {
		userNodes[res.userID] = append(userNodes[res.userID], res.unhealthyNodeID)
	}

	for userID, nodeIDs := range userNodes {
		if len(nodeIDs) == 0 {
			continue
		}

		subject := "Reserved Node Health Check Failed"

		var b strings.Builder
		for i, id := range nodeIDs {
			if i > 0 {
				b.WriteString("\n")
			}
			b.WriteString(fmt.Sprintf("Node ID: %d", id))
		}
		message := fmt.Sprintf(
			"You have %d reserved node(s) that are currently unhealthy",
			len(nodeIDs),
		)

		payloadData := map[string]string{
			"unhealthy_count": fmt.Sprintf("%d", len(nodeIDs)),
			"timestamp":       time.Now().UTC().Format("2006-01-02 15:04:05 UTC"),
		}
		payloadData["nodes_list"] = b.String()

		payload := notification.MergePayload(notification.CommonPayload{
			Subject: subject,
			Message: message,
			Status:  "unhealthy",
		}, payloadData)

		notif := models.NewNotification(
			userID,
			models.NotificationTypeNode,
			payload,
			models.WithSeverity(models.NotificationSeverityError),
			models.WithChannels(notification.ChannelEmail),
		)

		if err := notificationService.Send(context.Background(), notif); err != nil {
			logger.GetLogger().Error().Err(err).Int("user_id", userID).Msg("Failed to send consolidated notification")
		}
	}
}

func (h *Handler) healthCheckWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan models.UserNodes, results chan<- NodeHealthResult, grid proxy.Client) {
	defer wg.Done()

	for userNode := range jobs {
		node, err := grid.Node(ctx, userNode.NodeID)
		if err != nil {
			logger.GetLogger().Error().Err(err).Uint32("node_id", userNode.NodeID).Msg("Failed to get node for health check")
			continue
		}

		if node.Healthy {
			continue
		}

		results <- NodeHealthResult{
			userID:          userNode.UserID,
			unhealthyNodeID: userNode.NodeID,
		}
	}
}
