package handlers

import (
	"kubecloud/app/services"

	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	svc services.SettingsService
}

func NewSettingsHandler(svc services.SettingsService) SettingsHandler {
	return SettingsHandler{
		svc: svc,
	}
}

// @Summary Set maintenance mode
// @Description Sets maintenance mode for the system
// @Tags admin
// @ID set-maintenance-mode
// @Accept json
// @Produce json
// @Param body body MaintenanceModeStatus true "Maintenance Mode Status"
// @Success 200 {object} APIResponse
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /system/maintenance/status [put]
// SetMaintenanceModeHandler sets maintenance mode for the system
func (h *SettingsHandler) SetMaintenanceModeHandler(c *gin.Context) {
	reqLog := requestLogger(c, "SetMaintenanceModeHandler")
	var request MaintenanceModeStatus

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		reqLog.Error().Err(err).Msg("Invalid request format")
		BadRequest(c, "Invalid request format")
		return
	}

	if err := h.svc.SetMaintenanceMode(request.Enabled); err != nil {
		reqLog.Error().Err(err).Msg("Failed to set maintenance mode")
		InternalServerError(c)
		return
	}

	OK(c, "Maintenance mode is set successfully", nil)
}

// @Summary Get maintenance mode
// @Description Gets maintenance mode for the system
// @Tags admin
// @ID get-maintenance-mode
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=MaintenanceModeStatus}
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /system/maintenance/status [get]
// GetMaintenanceModeHandler gets maintenance mode for the system
func (h *SettingsHandler) GetMaintenanceModeHandler(c *gin.Context) {
	reqLog := requestLogger(c, "GetMaintenanceModeHandler")

	enabled, err := h.svc.GetMaintenanceMode()
	if err != nil {
		reqLog.Error().Err(err).Msg("Failed to get maintenance mode")
		InternalServerError(c)
		return
	}

	OK(c, "Maintenance mode is retrieved successfully", MaintenanceModeStatus{
		Enabled: enabled,
	})
}
