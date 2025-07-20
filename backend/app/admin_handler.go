package app

import (
	"context"
	"fmt"
	"kubecloud/internal"
	"kubecloud/models"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

// GenerateVouchersInput holds all data needed when creating vouchers
type GenerateVouchersInput struct {
	Count       int     `json:"count" binding:"required,gt=0" validate:"required,gt=0"`
	Value       float64 `json:"value" binding:"required,gt=0" validate:"required,gt=0"`
	ExpireAfter int     `json:"expire_after_days" binding:"required,gt=0"`
}

// CreditRequestInput represents a request to credit a user's balance
type CreditRequestInput struct {
	Amount uint64 `json:"amount" binding:"required,gt=0" validate:"required,gt=0"`
	Memo   string `json:"memo" binding:"required,min=3,max=255" validate:"required"`
}

// CreditUserResponse holds the response data after crediting a user
type CreditUserResponse struct {
	User   string `json:"user"`
	Amount uint64 `json:"amount"`
	Memo   string `json:"memo"`
}

// MetricsCollector interface for collecting metrics (Single Responsibility Principle)
type MetricsCollector interface {
	CollectBasicMetrics() map[string]float64
	CollectDBMetrics() map[string]float64
	CollectPaymentMetrics() map[string]float64
	CollectDeploymentMetrics() map[string]float64
}

// DependencyHealthChecker interface for health checks (Single Responsibility Principle)
type DependencyHealthChecker interface {
	CheckDependencies(ctx context.Context) (map[string]bool, error)
}

// defaultMetricsCollector implements MetricsCollector
type defaultMetricsCollector struct{}

func (m *defaultMetricsCollector) CollectBasicMetrics() map[string]float64 {
	return map[string]float64{
		"active_clusters":        internal.GetGaugeValue(internal.ActiveClusters),
		"users_registered_total": internal.GetCounterValue(internal.UsersRegisteredTotal),
	}
}

func (m *defaultMetricsCollector) CollectDBMetrics() map[string]float64 {
	return map[string]float64{
		"open": internal.GetGaugeValue(internal.GormDBConnections.WithLabelValues("open")),
		"idle": internal.GetGaugeValue(internal.GormDBConnections.WithLabelValues("idle")),
	}
}

func (m *defaultMetricsCollector) CollectPaymentMetrics() map[string]float64 {
	return map[string]float64{
		"success": internal.GetCounterValue(internal.StripePaymentsTotal.WithLabelValues("success")),
		"failure": internal.GetCounterValue(internal.StripePaymentsTotal.WithLabelValues("failure")),
	}
}

func (m *defaultMetricsCollector) CollectDeploymentMetrics() map[string]float64 {
	return map[string]float64{
		"success": internal.GetCounterValue(internal.ClusterDeploymentsTotal.WithLabelValues("success")),
		"failure": internal.GetCounterValue(internal.ClusterDeploymentsTotal.WithLabelValues("failure")),
	}
}

// defaultHealthChecker implements DependencyHealthChecker
type defaultHealthChecker struct {
	h *Handler
}

func (h *defaultHealthChecker) CheckDependencies(ctx context.Context) (map[string]bool, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	checks := map[string]HealthChecker{
		"database":           h.h.checkDatabase,
		"redis":              h.h.checkRedis,
		"gridproxy":          h.h.checkGridProxy,
		"tfchain":            h.h.checkTFChain,
		"activation_service": h.h.checkActivationService,
		"graphql":            h.h.checkGraphQL,
		"firesquid":          h.h.checkFiresquid,
	}

	dependencyHealth := make(map[string]bool)
	for name, check := range checks {
		// Use a goroutine with timeout for each health check
		healthChan := make(chan bool, 1)
		go func(checkName string, healthCheck HealthChecker) {
			defer func() {
				if r := recover(); r != nil {
					healthChan <- false
				}
			}()
			status := healthCheck(ctx)
			healthChan <- status.Status == "healthy"
		}(name, check)

		// Wait for result with timeout
		select {
		case healthy := <-healthChan:
			dependencyHealth[name] = healthy
			internal.SetHealthDependencyStatus(name, healthy)
		case <-ctx.Done():
			dependencyHealth[name] = false
			internal.SetHealthDependencyStatus(name, false)
		}
	}

	return dependencyHealth, nil
}

// @Summary Get all users
// @Description Returns a list of all users
// @Tags admin
// @ID get-all-users
// @Accept json
// @Produce json
// @Success 200 {array} models.User
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /users [get]
// ListUsersHandler lists all users
func (h *Handler) ListUsersHandler(c *gin.Context) {
	users, err := h.db.ListAllUsers()
	if err != nil {
		log.Error().Err(err).Msg("failed to list all users")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "Users are retrieved successfully", map[string]interface{}{
		"users": users,
	})
}

// @Summary Delete a user
// @Description Deletes a user from the system
// @Tags admin
// @ID delete-user
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Success 200 {object} APIResponse
// @Failure 400 {object} APIResponse "Invalid user ID"
// @Failure 403 {object} APIResponse "Admins cannot delete their own account"
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /users/{user_id} [delete]
// DeleteUsersHandler deletes user from system
func (h *Handler) DeleteUsersHandler(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		Error(c, http.StatusBadRequest, "User ID is required", "")
		return
	}

	id, err := strconv.Atoi(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid user ID", err.Error())
		return
	}

	authUserID := c.GetInt("user_id")
	if id == authUserID {
		Error(c, http.StatusForbidden, "Admins cannot delete their own account", "")
		return
	}

	err = h.db.DeleteUserByID(id)
	if err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to delete user")
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "User is deleted successfully", nil)
}

// @Summary Generate vouchers
// @Description Generates a bulk of vouchers
// @Tags admin
// @ID generate-vouchers
// @Accept json
// @Produce json
// @Param body body GenerateVouchersInput true "Generate Vouchers Input"
// @Success 201 {array} models.Voucher
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /vouchers/generate [post]
// GenerateVouchersHandler generates bulk of vouchers
func (h *Handler) GenerateVouchersHandler(c *gin.Context) {
	var request GenerateVouchersInput

	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	var vouchers []models.Voucher
	for i := 0; i < request.Count; i++ {
		voucherCode := internal.GenerateRandomVoucher(h.config.VoucherNameLength)
		timestampPart := fmt.Sprintf("%02d%02d", time.Now().Minute(), time.Now().Second())
		fullCode := fmt.Sprintf("%s-%s", voucherCode, timestampPart)

		voucher := models.Voucher{
			Code:      fullCode,
			Value:     request.Value,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(time.Duration(request.ExpireAfter) * 24 * time.Hour),
		}

		if err := h.db.CreateVoucher(&voucher); err != nil {
			log.Error().Err(err).Msg("failed to create voucher")
			InternalServerError(c)
			return
		}

		vouchers = append(vouchers, voucher)
	}

	Success(c, http.StatusCreated, "Vouchers are generated successfully", map[string]interface{}{
		"vouchers": vouchers,
	})
}

// @Summary List vouchers
// @Description Returns all vouchers in the system
// @Tags admin
// @ID list-vouchers
// @Accept json
// @Produce json
// @Success 200 {array} models.Voucher
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /vouchers [get]
// ListVouchersHandler returns all vouchers in system
func (h *Handler) ListVouchersHandler(c *gin.Context) {
	vouchers, err := h.db.ListAllVouchers()
	if err != nil {
		log.Error().Err(err).Msg("failed to list all vouchers")
		InternalServerError(c)
		return
	}
	Success(c, http.StatusOK, "Vouchers are Retrieved successfully", map[string]interface{}{
		"vouchers": vouchers,
	})
}

// @Summary Credit user balance
// @Description Credits a specific user's balance
// @Tags admin
// @ID credit-user
// @Accept json
// @Produce json
// @Param user_id path string true "User ID"
// @Param body body CreditRequestInput true "Credit Request Input"
// @Success 201 {object} CreditUserResponse
// @Failure 400 {object} APIResponse "Invalid request format or user ID"
// @Failure 500 {object} APIResponse
// @Security AdminMiddleware
// @Router /users/{user_id}/credit [post]
// CreditUserHandler lets admin credit specific user with money
func (h *Handler) CreditUserHandler(c *gin.Context) {
	userID := c.Param("user_id")
	if userID == "" {
		Error(c, http.StatusBadRequest, "User ID is required", "")
		return
	}

	var request CreditRequestInput
	// check on request format
	if err := c.ShouldBindJSON(&request); err != nil {
		Error(c, http.StatusBadRequest, "Invalid request format", err.Error())
		return
	}

	id, err := strconv.Atoi(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Invalid user ID format", "")
		return
	}

	user, err := h.db.GetUserByID(id)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	// get admin ID from middleware context
	adminID := c.GetInt("user_id")

	transaction := models.Transaction{
		UserID:    user.ID,
		AdminID:   adminID,
		Amount:    request.Amount,
		Memo:      request.Memo,
		CreatedAt: time.Now(),
	}

	if err := h.db.CreateTransaction(&transaction); err != nil {
		log.Error().Err(err).Msg("Failed to create credit transaction")
		InternalServerError(c)
		return
	}

	if err := h.db.CreditUserBalance(user.ID, request.Amount); err != nil {
		log.Error().Err(err).Msg("Failed to credit user")
		InternalServerError(c)
		return
	}

	err = internal.TransferTFTs(h.substrateClient, request.Amount, user.Mnemonic, h.config.SystemAccount.Mnemonic)
	if err != nil {
		log.Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	Success(c, http.StatusCreated, "User is credited successfully", CreditUserResponse{
		User:   user.Email,
		Amount: request.Amount,
		Memo:   request.Memo,
	})
}

// GetMetricsSummaryHandler returns a summary of key metrics for the dashboard
func (h *Handler) GetMetricsSummaryHandler(c *gin.Context) {
	metricsCollector := h.createMetricsCollector()
	healthChecker := h.createHealthChecker()

	summary, err := h.buildMetricsSummary(c.Request.Context(), metricsCollector, healthChecker)
	if err != nil {
		log.Error().Err(err).Msg("failed to build metrics summary")
		InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, summary)
}

// createMetricsCollector creates a metrics collector (Factory Pattern)
func (h *Handler) createMetricsCollector() MetricsCollector {
	return &defaultMetricsCollector{}
}

// createHealthChecker creates a health checker (Factory Pattern)
func (h *Handler) createHealthChecker() DependencyHealthChecker {
	return &defaultHealthChecker{h: h}
}

// buildMetricsSummary builds the complete metrics summary (Single Responsibility)
func (h *Handler) buildMetricsSummary(ctx context.Context, collector MetricsCollector, checker DependencyHealthChecker) (map[string]interface{}, error) {
	// Collect all metrics
	basicMetrics := collector.CollectBasicMetrics()
	dbMetrics := collector.CollectDBMetrics()
	paymentMetrics := collector.CollectPaymentMetrics()
	deploymentMetrics := collector.CollectDeploymentMetrics()

	// Get dependency health
	dependencyHealth, err := checker.CheckDependencies(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to check dependencies: %w", err)
	}

	// Build summary
	summary := map[string]interface{}{
		"active_clusters":           basicMetrics["active_clusters"],
		"users_registered_total":    basicMetrics["users_registered_total"],
		"db_connections":            dbMetrics,
		"stripe_payments_total":     paymentMetrics,
		"cluster_deployments_total": deploymentMetrics,
		"dependency_health":         dependencyHealth,
		"system_info": map[string]interface{}{
			"uptime":  time.Since(time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC)).String(),
			"version": "1.0.0",
		},
	}

	return summary, nil
}
