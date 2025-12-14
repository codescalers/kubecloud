package distributedlocks

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const (
	nodeLockKey = "locked"
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
func (l *RedisLocker) AcquireNodesLocks(ctx context.Context, nodeIDs []uint32) (map[string]string, error) {
	lockedKeys, err := l.acquireKeys(ctx, nodeLockKeys(nodeIDs))
	if err != nil {
		return nil, err
	}
	return lockedKeys, nil
}

func nodeLockKeys(nodeIDs []uint32) []string {
	keys := make([]string, len(nodeIDs))
	for i, id := range nodeIDs {
		keys[i] = fmt.Sprintf("%s:%d", nodeLockKey, id)
	}
	return keys
}

func (l *RedisLocker) acquireKeys(ctx context.Context, keys []string) (map[string]string, error) {
	locked := make(map[string]string, len(keys))

	for _, key := range keys {
		keyValue := uuid.New().String()
		ok, err := l.client.SetNX(ctx, key, keyValue, l.lockTimeout).Result()
		if err != nil {
			if rollErr := l.rollbackLocks(ctx, locked); rollErr != nil {
				return nil, rollErr
			}
			return nil, fmt.Errorf("redis error while acquiring lock for key %s: %w", key, err)
		}

		if !ok {
			if rollErr := l.rollbackLocks(ctx, locked); rollErr != nil {
				return nil, rollErr
			}
			return nil, fmt.Errorf("%w: %s", ErrNodeLocked, key)
		}

		locked[key] = keyValue
	}

	return locked, nil
}

func (l *RedisLocker) rollbackLocks(ctx context.Context, locked map[string]string) error {
	if len(locked) == 0 {
		return nil
	}
	keys := make([]string, 0, len(locked))
	for k := range locked {
		keys = append(keys, k)
	}

	if err := l.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("redis error while rolling back locks: %w", err)
	}

	return nil
}

func (l *RedisLocker) ReleaseLock(ctx context.Context, lockedKeys map[string]string) error {
	if len(lockedKeys) == 0 {
		return nil
	}

	var failedKeys []string
	for key, expectedValue := range lockedKeys {
		storedValue, err := l.client.Get(ctx, key).Result()
		if err == redis.Nil {
			continue
		}
		if err != nil {
			return fmt.Errorf("failed to get lock value for key %s: %w", key, err)
		}

		if storedValue != expectedValue {
			failedKeys = append(failedKeys, key)
			continue
		}

		if err := l.client.Del(ctx, key).Err(); err != nil {
			return fmt.Errorf("failed to delete lock for key %s: %w", key, err)
		}
	}

	if len(failedKeys) > 0 {
		return fmt.Errorf("lock value mismatch for keys: %v", failedKeys)
	}

	return nil
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
