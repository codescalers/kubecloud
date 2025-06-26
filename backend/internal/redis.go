package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

const (
	TaskStreamKey   = "deployment:tasks"
	ResultStreamKey = "deployment:results"
	ConsumerGroup   = "workers"
)

// RedisClient wraps redis client with task queue functionality
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client
func NewRedisClient(config Redis) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	client := &RedisClient{client: rdb}

	// Initialize streams and consumer groups
	if err := client.initializeStreams(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize streams: %w", err)
	}

	return client, nil
}

// initializeStreams creates the required streams and consumer groups
func (r *RedisClient) initializeStreams(ctx context.Context) error {
	// Create task stream if it doesn't exist
	if err := r.client.XGroupCreateMkStream(ctx, TaskStreamKey, ConsumerGroup, "0").Err(); err != nil {
		// TODO: const error
		if err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return fmt.Errorf("failed to create task stream consumer group: %w", err)
		}
	}

	// Create result stream if it doesn't exist
	if err := r.client.XGroupCreateMkStream(ctx, ResultStreamKey, ConsumerGroup+"_results", "0").Err(); err != nil {
		if err.Error() != "BUSYGROUP Consumer Group name already exists" {
			return fmt.Errorf("failed to create result stream consumer group: %w", err)
		}
	}

	log.Info().Msg("Redis streams initialized successfully")
	return nil
}

// TaskStatus represents the status of a deployment task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

// TODO: change the deployment task struct to match the actual task struct
// DeploymentTask represents a deployment task
type DeploymentTask struct {
	ID          string               `json:"id"`
	UserID      string               `json:"user_id"`
	Status      TaskStatus           `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
	StartedAt   *time.Time           `json:"started_at,omitempty"`
	CompletedAt *time.Time           `json:"completed_at,omitempty"`
	Payload     workloads.K8sCluster `json:"payload"`

	BlueprintID string                 `json:"blueprint_id"`
	Parameters  map[string]interface{} `json:"parameters"`
}

// DeploymentResult represents the result of a deployment task
type DeploymentResult struct {
	TaskID      string               `json:"task_id"`
	Status      TaskStatus           `json:"status"`
	Message     string               `json:"message"`
	Error       string               `json:"error,omitempty"`
	CompletedAt time.Time            `json:"completed_at"`
	Result      workloads.K8sCluster `json:"result,omitempty"`

	Resources map[string]interface{} `json:"resources,omitempty"`
}

// AddTask adds a new deployment task to the queue
func (r *RedisClient) AddTask(ctx context.Context, task *DeploymentTask) error {
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: TaskStreamKey,
		Values: map[string]interface{}{
			"task_id": task.ID,
			"data":    string(taskData),
		},
	}

	if err := r.client.XAdd(ctx, args).Err(); err != nil {
		return fmt.Errorf("failed to add task to stream: %w", err)
	}

	log.Info().Str("task_id", task.ID).Msg("Task added to queue")
	return nil
}

// GetPendingTasks retrieves pending tasks for workers
func (r *RedisClient) GetPendingTasks(ctx context.Context, consumerName string, count int64) ([]DeploymentTask, error) {
	args := &redis.XReadGroupArgs{
		Group:    ConsumerGroup,
		Consumer: consumerName,
		Streams:  []string{TaskStreamKey, ">"},
		Count:    count,
		Block:    time.Second * 5,
	}

	streams, err := r.client.XReadGroup(ctx, args).Result()
	if err != nil {
		if err == redis.Nil {
			return []DeploymentTask{}, nil // No new messages
		}
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	// TODO: should get one by one
	var tasks []DeploymentTask
	for _, stream := range streams {
		for _, message := range stream.Messages {
			var task DeploymentTask
			if taskData, ok := message.Values["data"].(string); ok {
				if err := json.Unmarshal([]byte(taskData), &task); err != nil {
					log.Error().Err(err).Str("message_id", message.ID).Msg("Failed to unmarshal task")
					continue
				}
				tasks = append(tasks, task)
			}
		}
	}

	return tasks, nil
}

// GetNextPendingTask retrieves the next pending task for a worker
func (r *RedisClient) GetNextPendingTask(ctx context.Context, consumerName string) (*DeploymentTask, error) {
	tasks, err := r.GetPendingTasks(ctx, consumerName, 1)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, nil // No tasks available
	}

	return &tasks[0], nil
}

// AckTask acknowledges that a task has been processed
func (r *RedisClient) AckTask(ctx context.Context, taskID string) error {
	// In a real implementation, you'd need to track message IDs
	// For now, we'll use a simple approach
	return nil
}

// AddResult adds a deployment result to the result stream
func (r *RedisClient) AddResult(ctx context.Context, result *DeploymentResult) error {
	resultData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: ResultStreamKey,
		Values: map[string]interface{}{
			"task_id": result.TaskID,
			"data":    string(resultData),
		},
	}

	if err := r.client.XAdd(ctx, args).Err(); err != nil {
		return fmt.Errorf("failed to add result to stream: %w", err)
	}

	log.Info().Str("task_id", result.TaskID).Msg("Result added to stream")
	return nil
}

// SubscribeToResults subscribes to deployment results
func (r *RedisClient) SubscribeToResults(ctx context.Context, callback func(*DeploymentResult)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			args := &redis.XReadArgs{
				Streams: []string{ResultStreamKey, "$"},
				Count:   10,
				Block:   time.Second * 1,
			}

			streams, err := r.client.XRead(ctx, args).Result()
			if err != nil {
				if err == redis.Nil {
					continue // No new messages
				}
				log.Error().Err(err).Msg("Failed to read from result stream")
				continue
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					var result DeploymentResult
					if resultData, ok := message.Values["data"].(string); ok {
						if err := json.Unmarshal([]byte(resultData), &result); err != nil {
							log.Error().Err(err).Str("message_id", message.ID).Msg("Failed to unmarshal result")
							continue
						}
						callback(&result)
					}
				}
			}
		}
	}
}

// Close closes the Redis connection
func (r *RedisClient) Close() error {
	return r.client.Close()
}
