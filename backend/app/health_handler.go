package app

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"runtime/debug"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/sync/errgroup"
)

const healthTimeout = 2 * time.Second

const (
	HealthyStatus   = "healthy"
	UnhealthyStatus = "unhealthy"
)

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

func (h *Handler) checkDatabase(ctx context.Context) HealthStatus {
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

func (h *Handler) checkRedis(ctx context.Context) HealthStatus {
	if h.redis == nil || h.redis.Client() == nil {
		return healthStatusFromError(fmt.Errorf("redis client not initialized"))
	}

	err := h.redis.Client().Ping(ctx).Err()
	return healthStatusFromError(err)
}

func httpHealthCheck(ctx context.Context, url string) HealthStatus {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return healthStatusFromError(err)
	}

	resp, err := healthHTTPClient.Do(req)
	if err != nil {
		return healthStatusFromError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return healthStatusFromError(fmt.Errorf("unexpected status: %s", resp.Status))
	}
	return healthStatusFromError(nil)
}

func healthURL(baseURL string) (string, error) {
	if len(strings.TrimSpace(baseURL)) == 0 {
		return "", fmt.Errorf("url not set")
	}
	return url.JoinPath(baseURL, "health")
}

func (h *Handler) checkGridProxy(ctx context.Context) HealthStatus {
	url, err := healthURL(h.config.GridProxyURL)
	if err != nil {
		return healthStatusFromError(fmt.Errorf("gridproxy %s", err.Error()))
	}
	return httpHealthCheck(ctx, url)
}

func tfchainHealthURL(tfchainURL string) (string, error) {
	if tfchainURL == "" {
		return "", fmt.Errorf("tfchain_url is empty")
	}
	url := strings.Replace(tfchainURL, "wss://", "https://", 1)
	url = strings.Replace(url, "/ws", "/health", 1)
	return url, nil
}

type tfchainHealth struct {
	Peers           int  `json:"peers"`
	IsSyncing       bool `json:"isSyncing"`
	ShouldHavePeers bool `json:"shouldHavePeers"`
}

func (h *Handler) checkTFChainHealth(ctx context.Context) HealthStatus {
	url, err := tfchainHealthURL(h.config.TFChainURL)
	if err != nil {
		return healthStatusFromError(err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return healthStatusFromError(err)
	}

	resp, err := healthHTTPClient.Do(req)
	if err != nil {
		return healthStatusFromError(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return healthStatusFromError(fmt.Errorf("unexpected status: %s", resp.Status))
	}

	var health tfchainHealth
	if err := json.NewDecoder(resp.Body).Decode(&health); err != nil {
		return healthStatusFromError(err)
	}

	if health.IsSyncing || !health.ShouldHavePeers || health.Peers == 0 {
		return healthStatusFromError(fmt.Errorf("syncing: %v, shouldHavePeers: %v, peers: %d",
			health.IsSyncing, health.ShouldHavePeers, health.Peers))
	}

	return healthStatusFromError(nil)
}

func (h *Handler) checkActivationService(ctx context.Context) HealthStatus {
	url, err := healthURL(h.config.ActivationServiceURL)
	if err != nil {
		return healthStatusFromError(fmt.Errorf("activation service %s", err.Error()))
	}
	return httpHealthCheck(ctx, url)
}

func checkGraphQLClient(client interface {
	Query(string, map[string]any) (map[string]any, error)
}) HealthStatus {
	_, err := client.Query("{ __typename }", map[string]any{})
	return healthStatusFromError(err)
}

func (h *Handler) checkGraphQL(ctx context.Context) HealthStatus {
	return checkGraphQLClient(&h.graphqlClient)
}

func (h *Handler) checkFiresquid(ctx context.Context) HealthStatus {
	return checkGraphQLClient(&h.firesquidClient)
}

func (h *Handler) HealthHandler(c *gin.Context) {
	ctx := c.Request.Context()
	checks := map[string]HealthChecker{
		"database":           h.checkDatabase,
		"redis":              h.checkRedis,
		"gridproxy":          h.checkGridProxy,
		"tfchain_health":     h.checkTFChainHealth,
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

	c.JSON(statusCode, results)
}

func (h *Handler) runChecks(ctx context.Context, checks map[string]HealthChecker) map[string]HealthStatus {
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
