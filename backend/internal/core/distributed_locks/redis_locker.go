package distributedlocks

import (
	"context"
	"fmt"
	"strconv"
	"strings"
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
	if err := l.acquireKeys(ctx, lockKeys(nodeIDs, nodeLockKey)); err != nil {
		return err
	}

	return nil
}

// AcquireWorkflowLock acquires a lock for the given workflow ID.
func (l *RedisLocker) AcquireWorkflowLock(ctx context.Context, nodeIDs []uint32, workflowID string) error {
	keys := lockKeys(nodeIDs, func(id uint32) string {
		return workflowLockKey(id, workflowID)
	})

	if err := l.acquireKeys(ctx, keys); err != nil {
		if rollErr := l.rollbackLocks(ctx, keys); rollErr != nil {
			return rollErr
		}
		return err
	}

	return nil
}

func nodeLockKey(nodeID uint32) string {
	return fmt.Sprintf("locked:%d", nodeID)
}

func workflowLockKey(nodeID uint32, workflowID string) string {
	return fmt.Sprintf("used:%d:%s", nodeID, workflowID)
}

func lockKeys(ids []uint32, keyFunc func(uint32) string) []string {
	keys := make([]string, len(ids))
	for i, id := range ids {
		keys[i] = keyFunc(id)
	}
	return keys
}

func (l *RedisLocker) acquireKeys(ctx context.Context, keys []string) error {
	locked := make([]string, 0, len(keys))

	for _, key := range keys {
		ok, err := l.client.SetNX(ctx, key, 1, l.lockTimeout).Result()
		if err != nil {
			if rollErr := l.rollbackLocks(ctx, locked); rollErr != nil {
				return rollErr
			}
			return fmt.Errorf("redis error while acquiring lock for key %s: %w", key, err)
		}

		if !ok {
			if rollErr := l.rollbackLocks(ctx, locked); rollErr != nil {
				return rollErr
			}
			return fmt.Errorf("%w: %s", ErrNodeLocked, key)
		}

		locked = append(locked, key)
	}

	return nil
}

func (l *RedisLocker) rollbackLocks(ctx context.Context, keys []string) error {
	if len(keys) == 0 {
		return nil
	}

	if err := l.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis error while rolling back locks: %w", err)
	}

	return nil
}

func (l *RedisLocker) ReleaseLock(ctx context.Context, nodeID uint32, workflowID string) error {
	lockedKey := nodeLockKey(nodeID)
	usedKey := workflowLockKey(nodeID, workflowID)
	return l.client.Del(ctx, lockedKey, usedKey).Err()
}

// GetAllWorkflowsLocks gets all workflow locks.
func (l *RedisLocker) GetAllWorkflowsLocks(ctx context.Context) ([]string, error) {
	return l.client.Keys(ctx, "used:*").Result()
}

func (l *RedisLocker) GetLockedNodes(ctx context.Context) ([]uint32, error) {
	keys, err := l.client.Keys(ctx, "locked:*").Result()
	if err != nil {
		return nil, err
	}
	nodes := make([]uint32, len(keys))
	for i, key := range keys {
		nodeID := strings.Split(key, ":")[1]
		value, parseErr := strconv.ParseUint(nodeID, 10, 32)
		if parseErr != nil {
			return nil, fmt.Errorf("failed to parse locked node id from %s: %w", key, parseErr)
		}
		nodes[i] = uint32(value)
	}
	return nodes, nil
}
