package queuing

import (
	"context"
	"sync"

	"github.com/redis/go-redis/v9"
	"github.com/xmonader/ewf"
)

var _ ewf.QueueEngine = (*redisQueueEngine)(nil)

// redisQueueEngine is the Redis implementation of the QueueEngine interface
type redisQueueEngine struct {
	client *redis.Client
	mu     sync.Mutex
	queues map[string]*redisQueue
}

// NewRedisQueueEngine creates a new RedisQueueEngine with the given Redis address
func NewRedisQueueEngine(client *redis.Client) ewf.QueueEngine {
	return &redisQueueEngine{
		client: client,
		queues: make(map[string]*redisQueue),
	}
}

// CreateQueue creates a new queue
func (e *redisQueueEngine) CreateQueue(ctx context.Context, queueName string, queueOptions ewf.QueueOptions) (ewf.Queue, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	if q, ok := e.queues[queueName]; ok {

		// signal queue activity, avoid blocking on full channel
		select {
		case q.activityCh <- struct{}{}:
		default:
		}

		return nil, ewf.ErrQueueAlreadyExists
	}

	q, err := NewRedisQueue(
		queueName,
		queueOptions,
		e.client,
	)

	if err != nil {
		return nil, err
	}

	e.queues[queueName] = q.(*redisQueue)

	return q, nil
}

// GetQueue retrieves an existing queue by its name
func (e *redisQueueEngine) GetQueue(ctx context.Context, queueName string) (ewf.Queue, error) {
	e.mu.Lock()
	defer e.mu.Unlock()

	q, ok := e.queues[queueName]
	if !ok {
		return nil, ewf.ErrQueueNotFound
	}

	// signal queue activity, avoid blocking on full channel
	select {
	case q.activityCh <- struct{}{}:
	default:
	}

	return q, nil
}

// CloseQueue closes and removes a queue by its name
func (e *redisQueueEngine) CloseQueue(ctx context.Context, queueName string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	q, ok := e.queues[queueName]
	if !ok {
		return ewf.ErrQueueNotFound
	}

	if err := q.Close(ctx); err != nil {
		return err
	}

	delete(e.queues, queueName)

	return nil
}

// Close closes all queues and the Redis client
func (e *redisQueueEngine) Close(ctx context.Context) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	for _, q := range e.queues {
		err := q.Close(ctx)
		if err != nil {
			return err
		}
	}

	e.queues = make(map[string]*redisQueue)
	return e.client.Close()
}
