package internal

import (
	"context"
	"fmt"
	"kubecloud/kubedeployer"
	"kubecloud/models"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"
)

// Worker represents a deployment worker
type Worker struct {
	ID         string
	redis      *RedisClient
	sseManager *SSEManager
	sshKey     string
	db         models.DB
	gridNet    string // Network name for the grid
	// Metrics callback: func(resultStatus string, activeClusters int)
	metricsCallback func(resultStatus string, activeClusters int)
}

// NewWorker creates a new worker instance
func NewWorker(id string, redis *RedisClient, sseManager *SSEManager, sshKey string, db models.DB, gridNet string) *Worker {
	return &Worker{
		ID:         id,
		redis:      redis,
		sseManager: sseManager,
		sshKey:     sshKey,
		db:         db,
		gridNet:    gridNet,
	}
}

// Start begins processing tasks - context cancellation stops the worker
func (w *Worker) Start(ctx context.Context) {
	log.Debug().Str("worker_id", w.ID).Msg("Worker started")
	defer log.Debug().Str("worker_id", w.ID).Msg("Worker stopped")

	w.handleUnacknowledgedTasks(ctx)

	w.listenForTasks(ctx)
}

// handleUnacknowledgedTasks processes any pending tasks on startup
func (w *Worker) handleUnacknowledgedTasks(ctx context.Context) {
	if err := w.redis.HandleUnacknowledgedTasks(ctx, w.ID, w.processTaskCallback); err != nil {
		log.Error().Err(err).Str("worker_id", w.ID).Msg("Failed to handle unacknowledged tasks")
	}
}

// listenForTasks subscribes to new tasks and processes them
func (w *Worker) listenForTasks(ctx context.Context) {
	if err := w.redis.SubscribeToTasks(ctx, w.ID, w.processTaskCallback); err != nil {
		if ctx.Err() != nil {
			log.Debug().Str("worker_id", w.ID).Msg("Worker stopped due to context cancellation")
			return
		}
		log.Error().Err(err).Str("worker_id", w.ID).Msg("Task subscription failed")
	}
}

// processTaskCallback is the callback function for processing tasks
func (w *Worker) processTaskCallback(ctx context.Context, task *DeploymentTask) {
	log.Info().Str("worker_id", w.ID).Str("task_id", task.TaskID).Msg("Processing task")

	if err := w.processTask(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Str("worker_id", w.ID).Msg("Task processing failed")
	}
}

// processTask processes a single deployment task
func (w *Worker) processTask(ctx context.Context, task *DeploymentTask) error {
	task.Status = TaskStatusProcessing
	now := time.Now()
	task.StartedAt = &now

	result := w.performDeployment(ctx, task)

	if err := w.redis.AddResult(ctx, result); err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to add result to stream")
	}

	if w.sseManager != nil {
		notificationData := map[string]interface{}{
			"type":    "deployment_update",
			"task_id": task.TaskID,
			"status":  result.Status,
			"message": result.Message,
		}

		if result.Error != "" {
			notificationData["error"] = result.Error
		}

		w.sseManager.Notify(task.UserID, "deployment_update", notificationData, task.TaskID)
	}

	if err := w.redis.AckTask(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to acknowledge task")
		return err
	}

	if result.Status == TaskStatusCompleted {
		cluster := &models.Cluster{
			UserID:      task.UserID,
			ProjectName: result.Result.Name,
		}

		if err := cluster.SetClusterResult(result.Result); err != nil {
			log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to serialize cluster result")
		} else {
			existingCluster, err := w.db.GetClusterByName(task.UserID, result.Result.Name)
			if err != nil {
				if err := w.db.CreateCluster(cluster); err != nil {
					log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to create cluster in database")
				} else {
					log.Info().Str("task_id", task.TaskID).Str("project_name", result.Result.Name).Msg("Cluster saved to database")
				}
			} else {
				existingCluster.Result = cluster.Result
				if err := w.db.UpdateCluster(&existingCluster); err != nil {
					log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to update cluster in database")
				} else {
					log.Info().Str("task_id", task.TaskID).Str("project_name", result.Result.Name).Msg("Cluster updated in database")
				}
			}
		}
		if w.metricsCallback != nil {
			if count, err := w.countActiveClusters(); err == nil {
				w.metricsCallback(string(result.Status), count)
			} else {
				w.metricsCallback(string(result.Status), 0)
			}
		}
	}

	log.Info().Str("task_id", task.TaskID).Str("status", string(result.Status)).Msg("Task completed")
	return nil
}

// performDeployment executes the actual deployment
func (w *Worker) performDeployment(ctx context.Context, task *DeploymentTask) *DeploymentResult {
	result := &DeploymentResult{
		TaskID: task.TaskID,
		UserID: task.UserID,
	}

	userID, err := strconv.Atoi(task.UserID)
	if err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to parse user ID")
		result.Status = TaskStatusFailed
		return result
	}

	user, err := w.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to get user")
		result.Status = TaskStatusFailed
		return result
	}

	var res kubedeployer.Cluster
	existingCluster, err := w.db.GetClusterByName(task.UserID, task.Payload.Name)
	if err == nil {
		log.Info().Str("task_id", task.TaskID).Str("project_name", task.Payload.Name).Msg("Cluster already exists, updating deployment")

		// Get the existing cluster result from the database
		existingClusterResult, err := existingCluster.GetClusterResult()
		if err != nil {
			log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to get existing cluster result from database")
			result.Status = TaskStatusFailed
			return result
		}

		// get the leader IP from the existing cluster
		leaderIP, err := existingCluster.GetLeaderIP()
		if err != nil {
			log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to get leader IP from existing cluster")
			result.Status = TaskStatusFailed
			return result
		}

		// Pass the existing cluster state from database to AddNodesToCluster
		res, err = kubedeployer.AddNodesToCluster(ctx, w.gridNet, user.Mnemonic, task.Payload, w.sshKey, leaderIP, &existingClusterResult)
		if err != nil {
			log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to deploy cluster")
			result.Status = TaskStatusFailed
			result.Error = "Deployment failed"
			result.Message = "Failed to update cluster"
			result.CompletedAt = time.Now()
			return result
		}
	} else {
		log.Info().Str("task_id", task.TaskID).Str("project_name", task.Payload.Name).Msg("Cluster does not exist, creating new deployment")
		res, err = kubedeployer.DeployCluster(ctx, w.gridNet, user.Mnemonic, task.Payload, w.sshKey)
		if err != nil {
			log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to deploy cluster")
			result.Status = TaskStatusFailed
			result.Error = "Deployment failed"
			result.Message = "Failed to deploy cluster"
			result.CompletedAt = time.Now()
			return result
		}
	}

	result.Result = res
	result.Status = TaskStatusCompleted
	result.Message = "Deployment completed successfully"
	result.CompletedAt = time.Now()

	log.Info().Str("task_id", task.TaskID).Msg("Deployment completed")
	return result
}

// countActiveClusters returns the number of active clusters in the DB
func (w *Worker) countActiveClusters() (int, error) {
	clusters, err := w.db.ListUserClusters("") // empty string to get all clusters
	if err != nil {
		return 0, err
	}
	return len(clusters), nil
}

// SetMetricsCallback sets the metrics callback for the worker
func (w *Worker) SetMetricsCallback(cb func(resultStatus string, activeClusters int)) {
	w.metricsCallback = cb
}

// WorkerManager manages multiple workers
type WorkerManager struct {
	workers []*Worker
	ctx     context.Context
	cancel  context.CancelFunc
}

// NewWorkerManager creates a new worker manager
func NewWorkerManager(redis *RedisClient, sseManager *SSEManager, workerCount int, sshKey string, db models.DB, gridNet string) *WorkerManager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &WorkerManager{
		ctx:    ctx,
		cancel: cancel,
	}

	for i := 0; i < workerCount; i++ {
		workerID := fmt.Sprintf("worker-%d", i+1)
		worker := NewWorker(workerID, redis, sseManager, sshKey, db, gridNet)
		manager.workers = append(manager.workers, worker)
	}

	return manager
}

// Start starts all workers
func (wm *WorkerManager) Start() {
	log.Info().Int("worker_count", len(wm.workers)).Msg("Starting workers")

	for _, worker := range wm.workers {
		go worker.Start(wm.ctx)
	}
}

// Stop stops all workers immediately
func (wm *WorkerManager) Stop() {
	wm.cancel()
}

// Workers returns the slice of workers
func (wm *WorkerManager) Workers() []*Worker {
	return wm.workers
}
