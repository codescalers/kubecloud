package distributedlocks

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisLocker struct {
	client      *redis.Client
	lockTimeout time.Duration
}

// NewRedisLocker creates a new RedisLocker instance.
func NewRedisLocker(client *redis.Client, lockTimeout time.Duration) *RedisLocker {
	return &RedisLocker{
		client:      client,
		lockTimeout: lockTimeout,
	}
}

// AcquireNodesLocks acquires locks for the given node IDs.
func (l *RedisLocker) AcquireNodesLocks(ctx context.Context, nodeIDs []uint32) error {
	keys := nodeLockKeys(nodeIDs)
	locked := make([]string, 0, len(keys))

	for _, key := range keys {
		ok, err := l.client.SetNX(ctx, key, 1, l.lockTimeout).Result()
		if err != nil {
			err = l.client.Del(ctx, locked...).Err()
			if err != nil {
				return fmt.Errorf("redis error while rolling back locks: %w", err)
			}
			return fmt.Errorf("redis error while acquiring lock for key %s: %w", key, err)
		}

		if !ok {
			err = l.client.Del(ctx, locked...).Err()
			if err != nil {
				return fmt.Errorf("redis error while rolling back locks: %w", err)
			}
			return fmt.Errorf("failed to acquire lock for key %s: node already locked", key)
		}

		locked = append(locked, key)
	}

	return nil
}

// AcquireWorkflowLock acquires a lock for the given workflow ID.
func (l *RedisLocker) AcquireWorkflowLock(ctx context.Context, nodeID uint32, workflowID string) (bool, error) {
	key := workflowLockKey(nodeID, workflowID)
	return l.client.SetNX(ctx, key, 1, l.lockTimeout).Result()
}

// ReleaseLock releases a lock for the given node ID and workflow ID.
func (l *RedisLocker) ReleaseLock(ctx context.Context, nodeID uint32, workflowID string) error {
	lockedKey := nodeLockKey(nodeID)
	usedKey := workflowLockKey(nodeID, workflowID)
	return l.client.Del(ctx, lockedKey, usedKey).Err()
}

// GetAllWorkflowsLocks gets all workflow locks.
func (l *RedisLocker) GetAllWorkflowsLocks(ctx context.Context) ([]string, error) {
	return l.client.Keys(ctx, "used:*").Result()
}

func nodeLockKey(nodeID uint32) string {
	return fmt.Sprintf("locked:%d", nodeID)
}

func nodeLockKeys(nodeIDs []uint32) []string {
	keys := make([]string, len(nodeIDs))
	for i, id := range nodeIDs {
		keys[i] = nodeLockKey(id)
	}
	return keys
}

func workflowLockKey(nodeID uint32, workflowID string) string {
	return fmt.Sprintf("used:%d:%s", nodeID, workflowID)
}
