package app

import (
	"net/http"

	"kubecloud/internal/logger"
	"kubecloud/models"

	"github.com/gin-gonic/gin"
	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

type statsService struct {
	userRepo    models.UserRepository
	clusterRepo models.ClusterRepository
}

func NewStatsService(
	userRepo models.UserRepository,
	clusterRepo models.ClusterRepository,
) statsService {
	return statsService{
		userRepo:    userRepo,
		clusterRepo: clusterRepo,
	}
}

type statsHandler struct {
	svc             statsService
	gridProxyClient proxy.Client
}

func newStatsHandler(svc statsService, gridProxyClient proxy.Client) statsHandler {
	return statsHandler{
		svc:             svc,
		gridProxyClient: gridProxyClient,
	}
}

type Stats struct {
	TotalUsers    uint32  `json:"total_users"`
	TotalClusters uint32  `json:"total_clusters"`
	UpNodes       uint32  `json:"up_nodes"`
	Countries     uint32  `json:"countries"`
	Cores         uint32  `json:"cores"`
	SSD           float64 `json:"ssd"`
}

// @Summary Get system statistics
// @Description Retrieves comprehensive system statistics.
// @Tags admin
// @ID get-stats
// @Accept json
// @Produce json
// @Success 200 {object} Stats "System statistics retrieved successfully"
// @Failure 500 {object} APIResponse "Internal Server Error - Failed to retrieve statistics"
// @Security AdminMiddleware
// @Router /stats [get]
// GetStatsHandler retrieves and returns system statistics including total users and clusters count.
func (h *statsHandler) GetStatsHandler(c *gin.Context) {
	totalUsers, err := h.svc.userRepo.CountAllUsers()
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to count total users")
		InternalServerError(c)
		return
	}

	totalClusters, err := h.svc.clusterRepo.CountAllClusters()
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to count total clusters")
		InternalServerError(c)
		return
	}

	stats, err := h.gridProxyClient.Stats(c.Request.Context(), types.StatsFilter{Status: []string{"up", "standby"}})
	if err != nil {
		logger.GetLogger().Error().Err(err).Msg("failed to retrieve up nodes count")
		InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, Stats{
		TotalUsers:    uint32(totalUsers),
		TotalClusters: uint32(totalClusters),
		UpNodes:       uint32(stats.Nodes),
		Countries:     uint32(stats.Countries),
		Cores:         uint32(stats.TotalCRU),
		SSD:           float64(stats.TotalSRU) / (1024 * 1024 * 1024),
	})
}
