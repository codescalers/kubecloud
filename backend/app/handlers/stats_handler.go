package handlers

import (
	"kubecloud/app/services"

	"github.com/gin-gonic/gin"
)

type StatsHandler struct {
	svc services.StatsService
}

func NewStatsHandler(svc services.StatsService) StatsHandler {
	return StatsHandler{svc: svc}
}

// @Summary Get system statistics
// @Description Retrieves comprehensive system statistics.
// @Tags admin
// @ID get-stats
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=services.Stats} "System statistics retrieved successfully"
// @Failure 500 {object} APIResponse "Internal Server Error - Failed to retrieve statistics"
// @Security AdminMiddleware
// @Router /stats [get]
// GetStatsHandler retrieves and returns system statistics including total users and clusters count.
func (h *StatsHandler) GetStatsHandler(c *gin.Context) {
	reqLog := requestLogger(c, "GetStatsHandler")

	stats, err := h.svc.GetStats(c.Request.Context())
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve up nodes count")
		InternalServerError(c)
		return
	}

	OK(c, "System statistics retrieved successfully", stats)
}
