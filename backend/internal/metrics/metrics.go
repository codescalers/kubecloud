package metrics

import (
	"net/http"
	"runtime"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"gorm.io/gorm"
)

const MetricsCollectorInterval = 10 * time.Second

// Metrics holds all the Prometheus metrics for the application
type Metrics struct {
	// HTTP metrics
	totalRequests      *prometheus.CounterVec
	successfulRequests *prometheus.CounterVec
	failedRequests     *prometheus.CounterVec
	requestDuration    *prometheus.HistogramVec

	// Cluster metrics
	clusterDeploymentSuccesses prometheus.Counter
	clusterDeploymentFailures  prometheus.Counter
	activeClusterCount         prometheus.Gauge

	// User metrics
	userRegistrations prometheus.Counter

	// Payment metrics
	stripePaymentSuccesses prometheus.Counter
	stripePaymentFailures  prometheus.Counter

	// GORM metrics
	gormOpenConnections prometheus.Gauge
	gormIdleConnections prometheus.Gauge

	// Registry for all metrics
	registry *prometheus.Registry
}

// NewMetrics creates and registers all metrics
func NewMetrics() *Metrics {
	registry := prometheus.NewRegistry()

	m := &Metrics{
		// HTTP metrics
		totalRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_total",
				Help: "Total number of HTTP requests",
			},
			[]string{"method", "endpoint"},
		),
		successfulRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_success",
				Help: "Number of successful HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		failedRequests: prometheus.NewCounterVec(
			prometheus.CounterOpts{
				Name: "http_requests_failed",
				Help: "Number of failed HTTP requests",
			},
			[]string{"method", "endpoint", "status"},
		),
		requestDuration: prometheus.NewHistogramVec(
			prometheus.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Duration of HTTP requests in seconds",
				Buckets: prometheus.DefBuckets,
			},
			[]string{"method", "endpoint", "status"},
		),

		// Cluster metrics
		clusterDeploymentSuccesses: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "cluster_deployment_successes",
				Help: "Number of successful cluster deployments",
			},
		),
		clusterDeploymentFailures: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "cluster_deployment_failures",
				Help: "Number of failed cluster deployments",
			},
		),
		activeClusterCount: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "active_clusters",
				Help: "Number of active clusters",
			},
		),

		// User metrics
		userRegistrations: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "user_registrations",
				Help: "Number of user registrations",
			},
		),

		// Payment metrics
		stripePaymentSuccesses: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "stripe_payment_successes",
				Help: "Number of successful Stripe payments",
			},
		),
		stripePaymentFailures: prometheus.NewCounter(
			prometheus.CounterOpts{
				Name: "stripe_payment_failures",
				Help: "Number of failed Stripe payments",
			},
		),

		// GORM metrics
		gormOpenConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "gorm_open_connections",
				Help: "Number of open GORM connections",
			},
		),
		gormIdleConnections: prometheus.NewGauge(
			prometheus.GaugeOpts{
				Name: "gorm_idle_connections",
				Help: "Number of idle GORM connections",
			},
		),

		registry: registry,
	}

	registry.MustRegister(
		m.totalRequests,
		m.successfulRequests,
		m.failedRequests,
		m.requestDuration,
		m.clusterDeploymentSuccesses,
		m.clusterDeploymentFailures,
		m.activeClusterCount,
		m.userRegistrations,
		m.stripePaymentSuccesses,
		m.stripePaymentFailures,
		m.gormOpenConnections,
		m.gormIdleConnections,

		// Register Go runtime metrics
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return m
}

// RegisterMetricsEndpoint registers the /metrics endpoint with the Gin router
func (m *Metrics) RegisterMetricsEndpoint(router *gin.Engine) {
	router.GET("/metrics", gin.WrapH(promhttp.HandlerFor(m.registry, promhttp.HandlerOpts{})))
}

// Middleware returns a Gin middleware that collects HTTP metrics
func (m *Metrics) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		// Increment the total requests counter
		m.totalRequests.WithLabelValues(method, path).Inc()

		// Process the request
		c.Next()

		// Record the request duration
		requestDuration := time.Since(start).Seconds()
		status := c.Writer.Status()
		statusStr := http.StatusText(status)

		// Record metrics based on the response status
		if status >= 200 && status < 400 {
			m.successfulRequests.WithLabelValues(method, path, statusStr).Inc()
		} else {
			m.failedRequests.WithLabelValues(method, path, statusStr).Inc()
		}

		m.requestDuration.WithLabelValues(method, path, statusStr).Observe(requestDuration)
	}
}

// UpdateGORMMetrics updates the GORM connection metrics
func (m *Metrics) UpdateGORMMetrics(db *gorm.DB) {
	sqlDB, err := db.DB()
	if err != nil {
		return
	}

	stats := sqlDB.Stats()
	m.gormOpenConnections.Set(float64(stats.OpenConnections))
	m.gormIdleConnections.Set(float64(stats.Idle))
}

// IncrementClusterDeploymentSuccess increments the successful cluster deployment counter
func (m *Metrics) IncrementClusterDeploymentSuccess() {
	m.clusterDeploymentSuccesses.Inc()
}

// IncrementClusterDeploymentFailure increments the failed cluster deployment counter
func (m *Metrics) IncrementClusterDeploymentFailure() {
	m.clusterDeploymentFailures.Inc()
}

// IncActiveClusterCount increments the active cluster count gauge
func (m *Metrics) IncActiveClusterCount() {
	m.activeClusterCount.Inc()
}

// DecActiveClusterCount decrements the active cluster count gauge
func (m *Metrics) DecActiveClusterCount() {
	m.activeClusterCount.Dec()
}

// IncrementUserRegistration increments the user registration counter
func (m *Metrics) IncrementUserRegistration() {
	m.userRegistrations.Inc()
}

// IncrementStripePaymentSuccess increments the successful Stripe payment counter
func (m *Metrics) IncrementStripePaymentSuccess() {
	m.stripePaymentSuccesses.Inc()
}

// IncrementStripePaymentFailure increments the failed Stripe payment counter
func (m *Metrics) IncrementStripePaymentFailure() {
	m.stripePaymentFailures.Inc()
}

// StartGORMMetricsCollector starts a goroutine that periodically updates GORM metrics
func (m *Metrics) StartGORMMetricsCollector(db *gorm.DB, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			m.UpdateGORMMetrics(db)
		}
	}()
}

// StartGoRuntimeMetricsCollector starts a goroutine that periodically collects runtime metrics
func (m *Metrics) StartGoRuntimeMetricsCollector(interval time.Duration) {
	var memStats runtime.MemStats

	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for range ticker.C {
			runtime.ReadMemStats(&memStats)
		}
	}()
}
