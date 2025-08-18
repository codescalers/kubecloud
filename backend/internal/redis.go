package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"kubecloud/kubedeployer"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
)

const (
	TaskStreamKey   = "deployment:tasks"
	ResultStreamKey = "deployment:results"
	ConsumerGroup   = "workers"
)

var (
	ErrConsumerGroupExists = errors.New("BUSYGROUP Consumer Group name already exists")
)

type RedisClient struct {
	client *redis.Client
}

// Client returns the underlying *redis.Client for health checks
func (r *RedisClient) Client() *redis.Client {
	return r.client
}

// NewRedisClient creates a new Redis client with task queue functionality
func NewRedisClient(config Redis) (*RedisClient, error) {
	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	rdb := redis.NewClient(&redis.Options{
		Network:  "tcp",
		Addr:     addr,
		Password: config.Password,
		DB:       config.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	client := &RedisClient{client: rdb}
	if err := client.initializeStreams(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize streams: %w", err)
	}

	return client, nil
}

func (r *RedisClient) initializeStreams(ctx context.Context) error {
	if err := r.client.XGroupCreateMkStream(ctx, TaskStreamKey, ConsumerGroup, "0").Err(); err != nil {
		if err.Error() != ErrConsumerGroupExists.Error() {
			return fmt.Errorf("failed to create task stream consumer group: %w", err)
		}
	}

	if err := r.client.XGroupCreateMkStream(ctx, ResultStreamKey, ConsumerGroup+"_results", "0").Err(); err != nil {
		if err.Error() != ErrConsumerGroupExists.Error() {
			return fmt.Errorf("failed to create result stream consumer group: %w", err)
		}
	}

	return nil
}

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

type DeploymentTask struct {
	TaskID      string               `json:"task_id"`
	UserID      string               `json:"user_id"`
	Status      TaskStatus           `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
	StartedAt   *time.Time           `json:"started_at,omitempty"`
	CompletedAt *time.Time           `json:"completed_at,omitempty"`
	Payload     kubedeployer.Cluster `json:"payload"`
	MessageID   string               `json:"-"`
}

type DeploymentResult struct {
	TaskID      string               `json:"task_id"`
	UserID      string               `json:"user_id"`
	Status      TaskStatus           `json:"status"`
	Message     string               `json:"message"`
	Error       string               `json:"error,omitempty"`
	CompletedAt time.Time            `json:"completed_at"`
	Result      kubedeployer.Cluster `json:"result,omitempty"`
}

// addToStream adds a message to the specified stream
func (r *RedisClient) addToStream(ctx context.Context, streamKey, id string, data interface{}) error {
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: streamKey,
		Values: map[string]interface{}{
			"id":   id,
			"data": string(jsonData),
		},
	}

	return r.client.XAdd(ctx, args).Err()
}

// subscribeToStream subscribes to messages from the specified stream
func (r *RedisClient) subscribeToStream(ctx context.Context, streamKey, consumerGroup, consumerName string, callback func(string, []byte)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		args := &redis.XReadGroupArgs{
			Group:    consumerGroup,
			Consumer: consumerName,
			Streams:  []string{streamKey, ">"}, // ">" means read both NEW and UNDELIVERED messages
			Count:    1,                        // read only one message at a time
			Block:    0,                        // block indefinitely, wait for new messages
		}

		streams, err := r.client.XReadGroup(ctx, args).Result()
		if err != nil {
			if err == redis.Nil {
				continue
			}
			if ctx.Err() != nil {
				return ctx.Err()
			}
			log.Error().Err(err).Str("stream", streamKey).Str("consumer", consumerName).Msg("Failed to read from stream")
			continue
		}

		for _, stream := range streams {
			for _, message := range stream.Messages {
				if data, ok := message.Values["data"].(string); ok {
					callback(message.ID, []byte(data))
				}
			}
		}
	}
}

// AddTask adds a deployment task to the task stream
func (r *RedisClient) AddTask(ctx context.Context, task *DeploymentTask) error {
	return r.addToStream(ctx, TaskStreamKey, task.TaskID, task)
}

// AddResult adds a deployment result to the result stream
func (r *RedisClient) AddResult(ctx context.Context, result *DeploymentResult) error {
	return r.addToStream(ctx, ResultStreamKey, result.TaskID, result)
}

// SubscribeToTasks subscribes to new tasks
func (r *RedisClient) SubscribeToTasks(ctx context.Context, consumerName string, callback func(context.Context, *DeploymentTask)) error {
	return r.subscribeToStream(ctx, TaskStreamKey, ConsumerGroup, consumerName, func(messageID string, data []byte) {
		var task DeploymentTask
		if err := json.Unmarshal(data, &task); err != nil {
			log.Error().Err(err).Str("message_id", messageID).Msg("Failed to unmarshal task")
			return
		}
		task.MessageID = messageID
		callback(ctx, &task)
	})
}

// SubscribeToResults subscribes to deployment results
func (r *RedisClient) SubscribeToResults(ctx context.Context, callback func(*DeploymentResult)) error {
	return r.subscribeToStream(ctx, ResultStreamKey, ConsumerGroup+"_results", "result_consumer", func(messageID string, data []byte) {
		var result DeploymentResult
		if err := json.Unmarshal(data, &result); err != nil {
			log.Error().Err(err).Str("message_id", messageID).Msg("Failed to unmarshal result")
			return
		}
		callback(&result)
	})
}

// AckTask acknowledges that a task has been processed
func (r *RedisClient) AckTask(ctx context.Context, task *DeploymentTask) error {
	if task.MessageID == "" {
		return fmt.Errorf("task MessageID is required for acknowledgment")
	}

	return r.client.XAck(ctx, TaskStreamKey, ConsumerGroup, task.MessageID).Err()
}

// HandleUnacknowledgedTasks processes messages that are DELIVERED but not ACKNOWLEDGED
func (r *RedisClient) HandleUnacknowledgedTasks(ctx context.Context, consumerName string, callback func(context.Context, *DeploymentTask)) error {
	pendingResult := r.client.XPending(ctx, TaskStreamKey, ConsumerGroup)
	if pendingResult.Err() != nil {
		if pendingResult.Err() == redis.Nil {
			return nil
		}
		return fmt.Errorf("failed to get pending messages: %w", pendingResult.Err())
	}

	pending := pendingResult.Val()
	if pending.Count == 0 {
		return nil
	}

	pendingExtResult := r.client.XPendingExt(ctx, &redis.XPendingExtArgs{
		Stream:   TaskStreamKey,
		Group:    ConsumerGroup,
		Start:    "-",
		End:      "+",
		Count:    100,
		Consumer: consumerName,
	})

	if pendingExtResult.Err() != nil {
		return fmt.Errorf("failed to get detailed pending messages: %w", pendingExtResult.Err())
	}

	pendingMessages := pendingExtResult.Val()
	for _, pendingMsg := range pendingMessages {
		readResult := r.client.XRange(ctx, TaskStreamKey, pendingMsg.ID, pendingMsg.ID)
		if readResult.Err() != nil {
			log.Error().Err(readResult.Err()).Str("message_id", pendingMsg.ID).Msg("Failed to read unacknowledged message")
			continue
		}

		messages := readResult.Val()
		for _, message := range messages {
			var task DeploymentTask
			if taskData, ok := message.Values["data"].(string); ok {
				if err := json.Unmarshal([]byte(taskData), &task); err != nil {
					log.Error().Err(err).Str("message_id", message.ID).Msg("Failed to unmarshal unacknowledged task")
					continue
				}
				task.MessageID = message.ID
				callback(ctx, &task)
			}
		}
	}

	return nil
}

func (r *RedisClient) SetMaintenanceMode(ctx context.Context, enabled bool) error {
	return r.client.Set(ctx, "maintenance_mode", enabled, 0).Err()
}

func (r *RedisClient) GetMaintenanceMode(ctx context.Context) (bool, error) {
	res, err := r.client.Get(ctx, "maintenance_mode").Result()

	if err != nil && err != redis.Nil {
		return false, err
	}

	return res == "1", nil
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
