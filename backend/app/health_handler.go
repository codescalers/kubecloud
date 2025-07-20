package app

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"kubecloud/internal"
	"kubecloud/models/sqlite"

	"github.com/gin-gonic/gin"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	proxyTypes "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

type HealthStatus struct {
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
}

type HealthChecker func(ctx context.Context) HealthStatus

func statusFromError(err error) HealthStatus {
	if err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	return HealthStatus{Status: "healthy"}
}

func simpleHealthCheck(client interface{}, checkFn func() error) HealthStatus {
	if client == nil {
		return HealthStatus{Status: "unhealthy", Message: "client not initialized"}
	}
	if err := checkFn(); err != nil {
		return HealthStatus{Status: "unhealthy", Message: err.Error()}
	}
	return HealthStatus{Status: "healthy"}
}

func (h *Handler) checkDatabase(ctx context.Context) HealthStatus {
	sqliteDB, ok := h.db.(*sqlite.Sqlite)
	if !ok {
		return HealthStatus{Status: "unhealthy", Message: "not a sqlite DB"}
	}
	db, err := sqliteDB.SQLDB()
	if err != nil {
		return statusFromError(err)
	}
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	return statusFromError(db.PingContext(ctx))
}

func (h *Handler) checkRedis(ctx context.Context) HealthStatus {
	return statusFromError(h.redis.Client().Ping(ctx).Err())
}

func (h *Handler) checkGridProxy(ctx context.Context) HealthStatus {
	return simpleHealthCheck(h.proxyClient, func() error {
		_, _, err := h.proxyClient.Nodes(ctx, proxyTypes.NodeFilter{}, proxyTypes.DefaultLimit())
		return err
	})
}

func (h *Handler) checkTFChain(ctx context.Context) HealthStatus {
	if h.substrateClient == nil {
		return HealthStatus{Status: "unhealthy", Message: "client not initialized"}
	}
	identity, idErr := substrate.NewIdentityFromSr25519Phrase(h.config.SystemAccount.Mnemonic)
	if idErr != nil {
		return statusFromError(idErr)
	}
	address := identity.Address()
	account, accErr := substrate.FromAddress(address)
	if accErr != nil {
		return statusFromError(accErr)
	}
	_, err := h.substrateClient.GetBalance(account)
	return statusFromError(err)
}

func (h *Handler) checkActivationService(ctx context.Context) HealthStatus {
	if h.config.ActivationServiceURL == "" {
		return HealthStatus{Status: "unhealthy", Message: "activation service URL not set"}
	}
	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(h.config.ActivationServiceURL + "/health")
	if err != nil {
		return statusFromError(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return HealthStatus{Status: "unhealthy", Message: fmt.Sprintf("unexpected status: %s", resp.Status)}
	}
	return HealthStatus{Status: "healthy"}
}

func (h *Handler) checkGraphQL(ctx context.Context) HealthStatus {
	_, err := h.graphqlClient.Query("{ __typename }", map[string]interface{}{})
	return statusFromError(err)
}

func (h *Handler) checkFiresquid(ctx context.Context) HealthStatus {
	_, err := h.firesquidClient.Query("{ __typename }", map[string]interface{}{})
	return statusFromError(err)
}

func (h *Handler) HealthHandler(c *gin.Context) {
	ctx := c.Request.Context()
	checks := map[string]HealthChecker{
		"database":           h.checkDatabase,
		"redis":              h.checkRedis,
		"gridproxy":          h.checkGridProxy,
		"tfchain":            h.checkTFChain,
		"activation_service": h.checkActivationService,
		"graphql":            h.checkGraphQL,
		"firesquid":          h.checkFiresquid,
	}
	resp := make(map[string]HealthStatus)
	for name, check := range checks {
		status := check(ctx)
		resp[name] = status
		internal.SetHealthDependencyStatus(name, status.Status == "healthy")
	}
	c.JSON(http.StatusOK, resp)
}
