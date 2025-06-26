package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
)

// Worker represents a deployment worker
type Worker struct {
	ID         string
	redis      *RedisClient
	sseManager *SSEManager
	stopChan   chan struct{}
	isRunning  bool
	gridClient deployer.TFPluginClient
}

// NewWorker creates a new worker instance
func NewWorker(id string, redis *RedisClient, sseManager *SSEManager, gridClient deployer.TFPluginClient) *Worker {
	return &Worker{
		ID:         id,
		redis:      redis,
		sseManager: sseManager,
		gridClient: gridClient,
		stopChan:   make(chan struct{}),
	}
}

// Start starts the worker to process tasks
func (w *Worker) Start(ctx context.Context) {
	if w.isRunning {
		return
	}

	w.isRunning = true
	log.Info().Str("worker_id", w.ID).Msg("Starting worker")

	go w.processLoop(ctx)
}

// Stop stops the worker
func (w *Worker) Stop() {
	if !w.isRunning {
		return
	}

	log.Info().Str("worker_id", w.ID).Msg("Stopping worker")
	close(w.stopChan)
	w.isRunning = false
}

// processLoop is the main processing loop for the worker
func (w *Worker) processLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case <-ticker.C:
			w.processTasks(ctx)
		}
	}
}

// processTasks retrieves and processes one pending task at a time
func (w *Worker) processTasks(ctx context.Context) {
	// Get one task at a time for better resource management and error isolation
	task, err := w.redis.GetNextPendingTask(ctx, w.ID)
	if err != nil {
		log.Error().Err(err).Str("worker_id", w.ID).Msg("Failed to get pending task")
		return
	}

	// If no task is available, return early
	if task == nil {
		log.Debug().Str("worker_id", w.ID).Msg("No pending tasks available")
		return
	}

	// Process the single task
	if err := w.processTask(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.ID).Str("worker_id", w.ID).Msg("Failed to process task")
	}
}

// processTask processes a single deployment task
func (w *Worker) processTask(ctx context.Context, task *DeploymentTask) error {
	log.Info().Str("task_id", task.ID).Str("worker_id", w.ID).Msg("Processing task")

	// Update task status to processing
	task.Status = TaskStatusProcessing
	now := time.Now()
	task.StartedAt = &now

	// Send status update via SSE
	if w.sseManager != nil {
		w.sseManager.SendTaskUpdate(
			fmt.Sprintf("%d", task.UserID),
			task.ID,
			TaskStatusProcessing,
			"Deployment started",
			map[string]interface{}{
				"blueprint_id": task.BlueprintID,
				"started_at":   task.StartedAt,
			},
		)
	}

	// Simulate deployment process
	result := w.performDeployment(ctx, task)

	// Send final result
	if err := w.redis.AddResult(ctx, result); err != nil {
		log.Error().Err(err).Str("task_id", task.ID).Msg("Failed to add result to stream")
		return err
	}

	// Send final status update via SSE
	if w.sseManager != nil {
		w.sseManager.SendTaskUpdate(
			fmt.Sprintf("%d", task.UserID),
			task.ID,
			result.Status,
			result.Message,
			map[string]interface{}{
				"resources":    result.Resources,
				"completed_at": result.CompletedAt,
				"error":        result.Error,
			},
		)
	}

	// Acknowledge task completion
	return w.redis.AckTask(ctx, task.ID)
}

// performDeployment simulates the actual deployment process
func (w *Worker) performDeployment(ctx context.Context, task *DeploymentTask) *DeploymentResult {
	result := &DeploymentResult{
		TaskID: task.ID,
	}

	if err := w.gridClient.K8sDeployer.Deploy(ctx, &task.Payload); err != nil {
		log.Error().Err(err).Str("task_id", task.ID).Msg("Failed to deploy cluster")
		result.Status = TaskStatusFailed
		result.Error = "Deployment failed"
		result.Message = "Failed to deploy cluster"
		result.CompletedAt = time.Now()
		return result
	}

	result.Status = TaskStatusCompleted
	result.Message = "Deployment completed"
	result.CompletedAt = time.Now()
	result.Result = task.Payload

	return result
}

// WorkerManager manages multiple workers
type WorkerManager struct {
	workers    []*Worker
	redis      *RedisClient
	sseManager *SSEManager
	ctx        context.Context
	cancel     context.CancelFunc
}

// NewWorkerManager creates a new worker manager
func NewWorkerManager(redis *RedisClient, sseManager *SSEManager, workerCount int, gridClient deployer.TFPluginClient) *WorkerManager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &WorkerManager{
		redis:      redis,
		sseManager: sseManager,
		ctx:        ctx,
		cancel:     cancel,
	}

	// Create workers
	for i := 0; i < workerCount; i++ {
		workerID := fmt.Sprintf("worker-%d", i+1)
		worker := NewWorker(workerID, redis, sseManager, gridClient)
		manager.workers = append(manager.workers, worker)
	}

	return manager
}

// Start starts all workers
func (wm *WorkerManager) Start() {
	log.Info().Int("worker_count", len(wm.workers)).Msg("Starting worker manager")

	for _, worker := range wm.workers {
		worker.Start(wm.ctx)
	}
}

// Stop stops all workers
func (wm *WorkerManager) Stop() {
	log.Info().Msg("Stopping worker manager")

	wm.cancel()

	for _, worker := range wm.workers {
		worker.Stop()
	}
}
