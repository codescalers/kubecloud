package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// DashboardPayload is the full request body for the API
type DashboardPayload struct {
	Dashboard map[string]interface{} `json:"dashboard"`
	FolderID  int                    `json:"folderId"`
	Overwrite bool                   `json:"overwrite"`
}

func graphPanel(title, expr, panelType string, id, row, col, w, h int, isFailure bool) map[string]interface{} {
	panel := map[string]interface{}{
		"id":    id,
		"type":  panelType,
		"title": title,
		"targets": []map[string]string{
			{"expr": expr},
		},
		"datasource": "Prometheus",
		"gridPos": map[string]int{
			"h": h,
			"w": w,
			"x": col,
			"y": row,
		},
	}
	// Add red color thresholds if this is a "failure" panel
	if isFailure {
		panel["fieldConfig"] = map[string]interface{}{
			"defaults": map[string]interface{}{
				"color": map[string]string{"mode": "thresholds"},
				"thresholds": map[string]interface{}{
					"mode": "absolute",
					"steps": []map[string]interface{}{
						{"color": "green", "value": nil}, // default
						{"color": "red", "value": 1},     // red if >= 1
					},
				},
			},
			"overrides": []interface{}{},
		}
	}

	return panel
}

func rowPanel(title string, id, y int) map[string]interface{} {
	return map[string]interface{}{
		"id":        id,
		"type":      "row",
		"title":     title,
		"collapsed": false,
		"gridPos": map[string]int{
			"h": 1,
			"w": 24,
			"x": 0,
			"y": y,
		},
	}
}

func main() {
	id := 1
	y := 0
	panels := []interface{}{
		// HTTP Metrics
		rowPanel("HTTP Metrics", id, y),
		graphPanel("HTTP Requests Total", "sum by (method, endpoint) (http_requests_total)", "graph", id+1, y+1, 0, 12, 8, false),
		graphPanel("HTTP Requests Success", "sum by (method, endpoint, status) (http_requests_success)", "graph", id+2, y+1, 12, 12, 8, false),
		graphPanel("HTTP Requests Failed", "sum by (method, endpoint, status) (http_requests_failed)", "graph", id+3, y+9, 0, 12, 8, true),
		graphPanel("Request Duration", "histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, method, endpoint, status))", "graph", id+4, y+9, 12, 12, 8, false),

		// Cluster Metrics
		rowPanel("Cluster Metrics", id+5, y+17),
		graphPanel("Cluster Deployment Successes", "increase(cluster_deployment_successes[$__range])", "stat", id+6, y+18, 0, 8, 6, false),
		graphPanel("Cluster Deployment Failures", "increase(cluster_deployment_failures[$__range])", "stat", id+7, y+18, 8, 8, 6, true),
		graphPanel("Active Clusters", "increase(active_clusters[$__range])", "stat", id+8, y+18, 16, 8, 6, false),

		// Users & Payments
		rowPanel("Users & Payments", id+9, y+25),
		graphPanel("User Registrations", "increase(user_registrations[$__range])", "stat", id+10, y+26, 0, 8, 6, false),
		graphPanel("Stripe Payment Successes", "increase(stripe_payment_successes[$__range])", "stat", id+11, y+26, 8, 8, 6, false),
		graphPanel("Stripe Payment Failures", "increase(stripe_payment_failures[$__range])", "stat", id+12, y+26, 16, 8, 6, true),

		// Email Metrics
		rowPanel("Email Metrics", id+13, y+33),
		graphPanel("Emails Sent (rate)", "rate(email_sent[5m])", "graph", id+14, y+34, 0, 12, 8, false),
		graphPanel("Emails Failed (rate)", "rate(email_failed[5m])", "graph", id+15, y+34, 12, 12, 8, true),

		// GORM
		rowPanel("Database (GORM)", id+16, y+41),
		graphPanel("GORM Open Connections", "gorm_open_connections", "stat", id+17, y+42, 0, 12, 6, false),
		graphPanel("GORM Idle Connections", "gorm_idle_connections", "stat", id+18, y+42, 12, 12, 6, false),

		// Go Runtime
		rowPanel("Go Runtime", id+19, y+50),
		graphPanel("Go Goroutines", "go_goroutines", "graph", id+17, y+42, 0, 12, 8, false),
		graphPanel("Go Memory Usage", "go_memstats_alloc_bytes", "graph", id+18, y+42, 12, 12, 8, false),
		graphPanel("Go GC Cycles", "go_gc_duration_seconds_count", "graph", id+19, y+50, 0, 12, 8, false),

		// Loki Logs
		rowPanel("Loki Logs", id+23, y+66),
		map[string]interface{}{
			"id":    id + 24,
			"type":  "logs",
			"title": "Application Logs",
			"targets": []map[string]interface{}{
				{
					"expr":     `{job="app-logs"}`,
					"refId":    "A",
					"datasource": "Loki",
				},
			},
			"gridPos": map[string]int{
				"h": 8,
				"w": 24,
				"x": 0,
				"y": y + 67,
			},
		},
	}

	dashboard := map[string]interface{}{
		"id":                    nil,
		"title":                 "Mycelium Cloud Dashboard",
		"editable":              true,
		"updateIntervalSeconds": 5,
		"refresh":               "5s",
		"time": map[string]string{
			"from": "now-1h",
			"to":   "now",
		},
		"timepicker": map[string]interface{}{
			"hidden": false,
		},
		"panels":        panels,
		"schemaVersion": 36,
		"version":       0,
	}

	grafanaRequestURL := fmt.Sprintf("%s/api/dashboards/db", os.Getenv("GRAFANA_URL"))
	grafanaUser := os.Getenv("GRAFANA_USER")
	grafanaPass := os.Getenv("GRAFANA_PASSWORD")

	payload := DashboardPayload{
		Dashboard: dashboard,
		FolderID:  0,
		Overwrite: true,
	}

	payloadBytes, _ := json.Marshal(payload)

	time.Sleep(10 * time.Second)

	req, err := http.NewRequest("POST", grafanaRequestURL, bytes.NewBuffer(payloadBytes))
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to create HTTP request: %v\n", err)
		os.Exit(1)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(grafanaUser, grafanaPass)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to send request to Grafana: %v\n", err)
		os.Exit(1)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Fprintf(os.Stderr, "failed to create Grafana dashboard: %s\n", resp.Status)
		os.Exit(1)
	}

	fmt.Println("Grafana dashboard is created successfully")
}
