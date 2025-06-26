package app

import (
	"kubecloud/internal"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

// DeploymentInput represents the input for deployment requests
type DeploymentInput struct {
	BlueprintID string                 `json:"blueprint_id" binding:"required"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// DeploymentResponse represents the response for deployment requests
type DeploymentResponse struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

// TaskStatusResponse represents the response for task status requests
type TaskStatusResponse struct {
	TaskID      string                 `json:"task_id"`
	Status      string                 `json:"status"`
	Message     string                 `json:"message"`
	BlueprintID string                 `json:"blueprint_id"`
	Parameters  map[string]interface{} `json:"parameters"`
	CreatedAt   time.Time              `json:"created_at"`
	StartedAt   *time.Time             `json:"started_at,omitempty"`
	CompletedAt *time.Time             `json:"completed_at,omitempty"`
}

func (h *Handler) DeployHandler(c *gin.Context) {
	// parse request to client deployment
	var cluster workloads.K8sCluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request json format"})
		return
	}

	// TODO: accept for now
	// // make sure it is valid
	// if err := cluster.Validate(); err != nil {
	// 	c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment object"})
	// 	return
	// }

	// create task and add to queue
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}
	taskID := uuid.New().String()

	task := &internal.DeploymentTask{
		ID:        taskID,
		UserID:    userID.(string),
		Status:    internal.TaskStatusPending,
		CreatedAt: time.Now(),
		Payload:   cluster,
	}

	if err := h.redis.AddTask(c.Request.Context(), task); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue deployment task"})
		return
	}

	// create response and send it to client
	response := DeploymentResponse{
		TaskID:    taskID,
		Status:    string(internal.TaskStatusPending),
		Message:   "Deployment task queued successfully",
		CreatedAt: task.CreatedAt,
	}

	c.JSON(http.StatusAccepted, response)
}

// DeployBlueprintHandler handles deployment requests
func (h *Handler) DeployBlueprintHandler(c *gin.Context) {
	var request DeploymentInput

	// Parse request
	if err := c.ShouldBindJSON(&request); err != nil {
		log.Error().Err(err).Msg("Invalid deployment request format")
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
		return
	}

	// Get user ID from context (set by auth middleware)
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// Validate blueprint ID
	if !isValidBlueprint(request.BlueprintID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid blueprint ID"})
		return
	}

	// Set default parameters if not provided
	if request.Parameters == nil {
		request.Parameters = make(map[string]interface{})
	}

	// Set default namespace if not provided
	if _, exists := request.Parameters["namespace"]; !exists {
		request.Parameters["namespace"] = "default"
	}

	// Generate task ID
	taskID := uuid.New().String()

	// Create deployment task
	task := &internal.DeploymentTask{
		ID:          taskID,
		UserID:      userID.(string),
		BlueprintID: request.BlueprintID,
		Parameters:  request.Parameters,
		Status:      internal.TaskStatusPending,
		CreatedAt:   time.Now(),
	}

	// Add task to Redis queue
	if err := h.redis.AddTask(c.Request.Context(), task); err != nil {
		log.Error().Err(err).Str("task_id", taskID).Msg("Failed to add task to queue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue deployment task"})
		return
	}

	// Send immediate response
	response := DeploymentResponse{
		TaskID:    taskID,
		Status:    string(internal.TaskStatusPending),
		Message:   "Deployment task queued successfully",
		CreatedAt: task.CreatedAt,
	}

	log.Info().
		Str("task_id", taskID).
		Str("blueprint_id", request.BlueprintID).
		Uint("user_id", userID.(uint)).
		Msg("Deployment task queued")

	c.JSON(http.StatusAccepted, response)
}

// GetTaskStatusHandler returns the status of a deployment task
func (h *Handler) GetTaskStatusHandler(c *gin.Context) {
	taskID := c.Param("task_id")
	if taskID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Task ID is required"})
		return
	}

	// Get user ID from context
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// In a real implementation, you would query the task status from Redis or database
	// For now, we'll return a mock response
	response := TaskStatusResponse{
		TaskID:      taskID,
		Status:      string(internal.TaskStatusProcessing),
		Message:     "Task is being processed",
		BlueprintID: "nginx", // This would come from stored task data
		Parameters: map[string]interface{}{
			"namespace": "default",
			"replicas":  3,
		},
		CreatedAt: time.Now().Add(-time.Minute * 5),
		StartedAt: func() *time.Time { t := time.Now().Add(-time.Minute * 3); return &t }(),
	}

	c.JSON(http.StatusOK, response)
}

// ListUserTasksHandler returns all tasks for the authenticated user
func (h *Handler) ListUserTasksHandler(c *gin.Context) {
	// Get user ID from context
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	// In a real implementation, you would query tasks from Redis or database
	// For now, we'll return mock data
	tasks := []TaskStatusResponse{
		{
			TaskID:      "task-1",
			Status:      string(internal.TaskStatusCompleted),
			Message:     "NGINX deployment completed successfully",
			BlueprintID: "nginx",
			Parameters: map[string]interface{}{
				"namespace": "default",
				"replicas":  3,
			},
			CreatedAt:   time.Now().Add(-time.Hour * 2),
			StartedAt:   func() *time.Time { t := time.Now().Add(-time.Hour*2 + time.Minute*1); return &t }(),
			CompletedAt: func() *time.Time { t := time.Now().Add(-time.Hour*2 + time.Minute*5); return &t }(),
		},
		{
			TaskID:      "task-2",
			Status:      string(internal.TaskStatusProcessing),
			Message:     "PostgreSQL deployment in progress",
			BlueprintID: "postgresql",
			Parameters: map[string]interface{}{
				"namespace": "database",
				"storage":   "10Gi",
			},
			CreatedAt: time.Now().Add(-time.Minute * 10),
			StartedAt: func() *time.Time { t := time.Now().Add(-time.Minute * 8); return &t }(),
		},
	}

	log.Info().Uint("user_id", userID.(uint)).Msg("Listing user tasks")
	c.JSON(http.StatusOK, gin.H{"tasks": tasks})
}

// GetAvailableBlueprintsHandler returns available blueprints
func (h *Handler) GetAvailableBlueprintsHandler(c *gin.Context) {
	blueprints := []map[string]interface{}{
		{
			"id":          "nginx",
			"name":        "NGINX Web Server",
			"description": "Deploy a scalable NGINX web server",
			"parameters": map[string]interface{}{
				"replicas": map[string]interface{}{
					"type":        "integer",
					"default":     3,
					"description": "Number of replicas",
					"min":         1,
					"max":         10,
				},
				"namespace": map[string]interface{}{
					"type":        "string",
					"default":     "default",
					"description": "Kubernetes namespace",
				},
			},
		},
		{
			"id":          "postgresql",
			"name":        "PostgreSQL Database",
			"description": "Deploy a PostgreSQL database with persistent storage",
			"parameters": map[string]interface{}{
				"storage": map[string]interface{}{
					"type":        "string",
					"default":     "10Gi",
					"description": "Storage size for the database",
				},
				"namespace": map[string]interface{}{
					"type":        "string",
					"default":     "database",
					"description": "Kubernetes namespace",
				},
			},
		},
		{
			"id":          "fullstack",
			"name":        "Full-Stack Application",
			"description": "Deploy a complete full-stack application with frontend, backend, and database",
			"parameters": map[string]interface{}{
				"frontend_replicas": map[string]interface{}{
					"type":        "integer",
					"default":     2,
					"description": "Number of frontend replicas",
				},
				"backend_replicas": map[string]interface{}{
					"type":        "integer",
					"default":     3,
					"description": "Number of backend replicas",
				},
				"namespace": map[string]interface{}{
					"type":        "string",
					"default":     "fullstack",
					"description": "Kubernetes namespace",
				},
			},
		},
	}

	c.JSON(http.StatusOK, gin.H{"blueprints": blueprints})
}

// isValidBlueprint validates if the blueprint ID is supported
func isValidBlueprint(blueprintID string) bool {
	validBlueprints := map[string]bool{
		"nginx":      true,
		"postgresql": true,
		"fullstack":  true,
	}
	return validBlueprints[blueprintID]
}
