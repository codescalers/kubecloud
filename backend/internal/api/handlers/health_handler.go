package handlers

import (
	grid "kubecloud/internal/infrastructure/grid"
	"context"
	"encoding/json"
	"fmt"
	"kubecloud/internal/core/models"
	"kubecloud/internal/infrastructure/logger"
	
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/graphql"
	"golang.org/x/sync/errgroup"
)

const healthTimeout = 2 * time.Second

const (
	HealthyStatus   = "healthy"
	UnhealthyStatus = "unhealthy"
)

type HealthHandler struct {
	db              models.DB
	systemNetwork   string
	firesquidClient graphql.GraphQl
	graphql         graphql.GraphQl
}

func NewHealthHandler(systemNetwork string,
	firesquidClient graphql.GraphQl,
	graphql graphql.GraphQl,
	db models.DB,
) HealthHandler {
	return HealthHandler{
		db:              db,
		systemNetwork:   systemNetwork,
		firesquidClient: firesquidClient,
		graphql:         graphql,
	}
}

type HealthStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type HealthChecker func(ctx context.Context) HealthStatus

var healthHTTPClient = &http.Client{Timeout: healthTimeout}

func healthStatusFromError(err error) HealthStatus {
	if err == nil {
		return HealthStatus{Status: HealthyStatus}
	}
	return HealthStatus{Status: UnhealthyStatus, Message: err.Error()}
}

func (h *HealthHandler) checkDatabase(ctx context.Context) HealthStatus {
	type pinger interface {
		Ping(ctx context.Context) error
	}

	dbPinger, ok := h.db.(pinger)
	if !ok {
		return healthStatusFromError(fmt.Errorf("database does not support ping"))
	}

	ctx, cancel := context.WithTimeout(ctx, healthTimeout)
	defer cancel()

	err := dbPinger.Ping(ctx)
	return healthStatusFromError(err)
}

func httpHealthCheck(ctx context.Context, urls []string) HealthStatus {
	if len(urls) == 0 {
		return healthStatusFromError(fmt.Errorf("no URLs provided for health check"))
	}

	for _, urlStr := range urls {
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, urlStr, nil)
		if err != nil {
			continue
		}

		resp, err := healthHTTPClient.Do(req)
		if err != nil {
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusOK {
			return healthStatusFromError(nil)
		}
		logger.GetLogger().Error().Str("url", urlStr).Int("status", resp.StatusCode).Msg("Health check failed")
	}

	return healthStatusFromError(fmt.Errorf("all URLs failed health check"))
}

func healthURL(baseURL string) (string, error) {
	if len(strings.TrimSpace(baseURL)) == 0 {
		return "", fmt.Errorf("url not set")
	}
	return url.JoinPath(baseURL, "health")
}

func (h *HealthHandler) checkGridProxy(ctx context.Context) HealthStatus {
	proxyURLs := deployer.ProxyURLs[h.systemNetwork]
	var validURLs []string

	for _, proxyURL := range proxyURLs {
		url, err := healthURL(proxyURL)
		if err != nil {
			return healthStatusFromError(fmt.Errorf("gridproxy %s", err.Error()))
		}

		validURLs = append(validURLs, url)
	}

	return httpHealthCheck(ctx, validURLs)
}

func (h *HealthHandler) checkTFChainHealth(ctx context.Context) HealthStatus {
	chainURLs := deployer.SubstrateURLs[h.systemNetwork]

	url := strings.Replace(chainURLs[0], "wss://", "https://", 1)
	url = strings.TrimSuffix(url, "/ws")

	payload := `{"id":1,"jsonrpc":"2.0","method":"system_health","params":[]}`

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(payload))
	if err != nil {
		return healthStatusFromError(err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := healthHTTPClient.Do(req)
	if err != nil {
		return healthStatusFromError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return healthStatusFromError(fmt.Errorf("unexpected status: %s", resp.Status))
	}

	var rpcResp struct {
		Result struct {
			Peers     int  `json:"peers"`
			IsSyncing bool `json:"isSyncing"`
		} `json:"result"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&rpcResp); err != nil {
		return healthStatusFromError(err)
	}

	if rpcResp.Result.IsSyncing || rpcResp.Result.Peers == 0 {
		return healthStatusFromError(fmt.Errorf("syncing: %v, peers: %d", rpcResp.Result.IsSyncing, rpcResp.Result.Peers))
	}

	return healthStatusFromError(nil)
}

func (h *HealthHandler) checkActivationService(ctx context.Context) HealthStatus {
	url, err := healthURL(grid.ActivationServiceURLs[h.systemNetwork])
	if err != nil {
		return healthStatusFromError(fmt.Errorf("activation service %s", err.Error()))
	}
	return httpHealthCheck(ctx, []string{url})
}

func checkGraphQLClient(client interface {
	Query(string, map[string]any) (map[string]any, error)
}) HealthStatus {
	_, err := client.Query("{ __typename }", map[string]any{})
	return healthStatusFromError(err)
}

func (h *HealthHandler) checkGraphQL(ctx context.Context) HealthStatus {
	return checkGraphQLClient(&h.graphql)
}

func (h *HealthHandler) checkFiresquid(ctx context.Context) HealthStatus {
	return checkGraphQLClient(&h.firesquidClient)
}

// @Summary Health check endpoint
// @Description Returns the health status of various system components
// @Tags health
// @Produce json
// @Success 200 {object} map[string]HealthStatus "All systems healthy"
// @Failure 503 {object} map[string]HealthStatus "One or more systems unhealthy"
// @Router /health [get]

func (h *HealthHandler) HealthHandler(c *gin.Context) {
	ctx := c.Request.Context()
	checks := map[string]HealthChecker{
		"database":           h.checkDatabase,
		"gridproxy":          h.checkGridProxy,
		"tfchain":            h.checkTFChainHealth,
		"activation_service": h.checkActivationService,
		"graphql":            h.checkGraphQL,
		"firesquid":          h.checkFiresquid,
	}

	results := h.runChecks(ctx, checks)

	statusCode := http.StatusOK
	for _, status := range results {
		if status.Status != HealthyStatus {
			statusCode = http.StatusServiceUnavailable
			break
		}
	}

	if statusCode != http.StatusOK {
		ServiceUnavailable(c, "Health check failed", results)
		return
	}
	OK(c, "Health check passed", results)
}

func (h *HealthHandler) runChecks(ctx context.Context, checks map[string]HealthChecker) map[string]HealthStatus {
	results := make(map[string]HealthStatus, len(checks))
	var mu sync.Mutex
	var g errgroup.Group

	for name, checker := range checks {
		name, checker := name, checker
		g.Go(func() error {
			defer func() {
				if r := recover(); r != nil {
					mu.Lock()
					results[name] = healthStatusFromError(fmt.Errorf("panic: %v\n%s", r, string(debug.Stack())))
					mu.Unlock()
				}
			}()

			status := checker(ctx)
			mu.Lock()
			results[name] = status
			mu.Unlock()
			return nil
		})
	}

	_ = g.Wait()
	return results
}
