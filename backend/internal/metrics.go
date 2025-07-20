package internal

import (
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	dto "github.com/prometheus/client_model/go"
)

var (
	HttpRequestsTotal       *prometheus.CounterVec
	HttpRequestDuration     *prometheus.HistogramVec
	HttpRequestsSuccess     *prometheus.CounterVec
	HttpRequestsError       *prometheus.CounterVec
	ClusterDeploymentsTotal *prometheus.CounterVec
	ActiveClusters          prometheus.Gauge
	UsersRegisteredTotal    prometheus.Counter
	StripePaymentsTotal     *prometheus.CounterVec
	GormDBConnections       *prometheus.GaugeVec
	HealthDependencyStatus  *prometheus.GaugeVec
	initMetricsOnce         sync.Once
)

// Helper functions for metrics
func IncClusterDeployments(result string) {
	ClusterDeploymentsTotal.WithLabelValues(result).Inc()
}

func SetActiveClusters(count int) {
	ActiveClusters.Set(float64(count))
}

func IncUsersRegistered() {
	UsersRegisteredTotal.Inc()
}

func IncStripePayments(result string) {
	StripePaymentsTotal.WithLabelValues(result).Inc()
}

func SetGormDBConnections(open, idle int) {
	GormDBConnections.WithLabelValues("open").Set(float64(open))
	GormDBConnections.WithLabelValues("idle").Set(float64(idle))
}

func SetHealthDependencyStatus(dep string, healthy bool) {
	if healthy {
		HealthDependencyStatus.WithLabelValues(dep, "healthy").Set(1)
		HealthDependencyStatus.WithLabelValues(dep, "unhealthy").Set(0)
	} else {
		HealthDependencyStatus.WithLabelValues(dep, "healthy").Set(0)
		HealthDependencyStatus.WithLabelValues(dep, "unhealthy").Set(1)
	}
}

// Helper to get the current value of a Gauge
func GetGaugeValue(g prometheus.Gauge) float64 {
	m := &dto.Metric{}
	if err := g.Write(m); err == nil && m.Gauge != nil {
		return m.Gauge.GetValue()
	}
	return 0
}

// Helper to get the current value of a Counter
func GetCounterValue(c prometheus.Counter) float64 {
	m := &dto.Metric{}
	if err := c.Write(m); err == nil && m.Counter != nil {
		return m.Counter.GetValue()
	}
	return 0
}

// InitMetrics registers Go runtime metrics and initializes all custom metrics
func InitMetrics() {
	fmt.Println("InitMetrics called")
	initMetricsOnce.Do(func() {
		HttpRequestsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"handler", "method", "status"},
		)
		HttpRequestDuration = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"handler", "method", "status"},
		)
		HttpRequestsSuccess = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_success_total",
				Help: "Total number of successful HTTP requests",
			},
			[]string{"handler", "method", "status"},
		)
		HttpRequestsError = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_error_total",
				Help: "Total number of failed HTTP requests",
			},
			[]string{"handler", "method", "status"},
		)
		ClusterDeploymentsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "cluster_deployments_total",
				Help: "Total number of cluster deployments by result",
			},
			[]string{"result"},
		)
		ActiveClusters = promauto.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_clusters",
				Help: "Current number of active clusters",
			},
		)
		UsersRegisteredTotal = promauto.NewCounter(
			prometheus.CounterOpts{
				Name: "users_registered_total",
				Help: "Total number of users registered",
			},
		)
		StripePaymentsTotal = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "stripe_payments_total",
				Help: "Total number of Stripe payments by result",
			},
			[]string{"result"},
		)
		GormDBConnections = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "gorm_db_connections",
				Help: "Number of GORM DB connections by state",
			},
			[]string{"state"},
		)
		HealthDependencyStatus = promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "health_dependency_status",
				Help: "Health status of dependencies (1=healthy, 0=unhealthy)",
			},
			[]string{"dependency", "status"},
		)
	})
}

// MetricsHandler returns a Gin handler for /metrics
func MetricsHandler() gin.HandlerFunc {
	h := promhttp.Handler()
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
