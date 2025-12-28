package distributedlocks

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func newTestRedisClient(t *testing.T) *redis.Client {
	t.Helper()

	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	require.NoError(t, client.Ping(context.Background()).Err())
	require.NoError(t, client.FlushDB(context.Background()).Err())

	t.Cleanup(func() {
		require.NoError(t, client.FlushDB(context.Background()).Err())
		require.NoError(t, client.Close())
	})

	return client
}

func TestRedisLocker_AcquireLocks_Success(t *testing.T) {
	client := newTestRedisClient(t)
	locker := &RedisLocker{client: client, lockTimeout: time.Minute}

	resourceKeys := []string{"node:1", "node:2"}
	lockedKeys, err := locker.AcquireLocks(context.Background(), resourceKeys)

	require.NoError(t, err)
	require.Len(t, lockedKeys, 2)
	require.Contains(t, lockedKeys, "node:1")
	require.Contains(t, lockedKeys, "node:2")
	// Verify UUID values are stored
	require.NotEmpty(t, lockedKeys["node:1"])
	require.NotEmpty(t, lockedKeys["node:2"])
	require.Equal(t, int64(1), client.Exists(context.Background(), "node:1").Val())
	require.Equal(t, int64(1), client.Exists(context.Background(), "node:2").Val())
	// Verify the stored values match
	val1, _ := client.Get(context.Background(), "node:1").Result()
	val2, _ := client.Get(context.Background(), "node:2").Result()
	require.Equal(t, lockedKeys["node:1"], val1)
	require.Equal(t, lockedKeys["node:2"], val2)
}

func TestRedisLocker_AcquireLocks_ResourceAlreadyLocked(t *testing.T) {
	client := newTestRedisClient(t)
	locker := &RedisLocker{client: client, lockTimeout: time.Minute}

	// Set an existing lock with a UUID value
	existingValue := "existing-uuid-value"
	require.NoError(t, client.Set(context.Background(), "node:2", existingValue, 0).Err())

	resourceKeys := []string{"node:1", "node:2"}
	lockedKeys, err := locker.AcquireLocks(context.Background(), resourceKeys)

	require.Error(t, err)
	require.ErrorIs(t, err, ErrResourceLocked)
	require.Contains(t, err.Error(), "node:2")
	require.Nil(t, lockedKeys)
	require.Equal(t, int64(0), client.Exists(context.Background(), "node:1").Val(), "previous locks should be rolled back")
	// Verify the existing lock is still there
	val, _ := client.Get(context.Background(), "node:2").Result()
	require.Equal(t, existingValue, val)
}

func TestRedisLocker_ReleaseLock_Success(t *testing.T) {
	client := newTestRedisClient(t)
	locker := &RedisLocker{client: client, lockTimeout: time.Minute}

	// Set locks with specific UUID values
	lockValue1 := "uuid-value-1"
	lockValue2 := "uuid-value-2"
	require.NoError(t, client.Set(context.Background(), "node:1", lockValue1, 0).Err())
	require.NoError(t, client.Set(context.Background(), "node:2", lockValue2, 0).Err())

	// Release locks with matching values
	lockedKeys := map[string]string{
		"node:1": lockValue1,
		"node:2": lockValue2,
	}
	err := locker.ReleaseLocks(context.Background(), lockedKeys)

	require.NoError(t, err)
	require.Equal(t, int64(0), client.Exists(context.Background(), "node:1").Val())
	require.Equal(t, int64(0), client.Exists(context.Background(), "node:2").Val())
}

func TestRedisLocker_ReleaseLock_ValueMismatch(t *testing.T) {
	client := newTestRedisClient(t)
	locker := &RedisLocker{client: client, lockTimeout: time.Minute}

	// Set lock with a specific value
	storedValue := "stored-uuid-value"
	require.NoError(t, client.Set(context.Background(), "node:1", storedValue, 0).Err())

	// Try to release with wrong value
	lockedKeys := map[string]string{
		"node:1": "wrong-uuid-value",
	}
	err := locker.ReleaseLocks(context.Background(), lockedKeys)

	require.Error(t, err)
	require.Contains(t, err.Error(), "lock value mismatch")
	// Verify the lock is still there
	require.Equal(t, int64(1), client.Exists(context.Background(), "node:1").Val())
	val, _ := client.Get(context.Background(), "node:1").Result()
	require.Equal(t, storedValue, val)
}

func TestRedisLocker_ReleaseLock_KeyNotExists(t *testing.T) {
	client := newTestRedisClient(t)
	locker := &RedisLocker{client: client, lockTimeout: time.Minute}

	// Try to release a non-existent lock
	lockedKeys := map[string]string{
		"node:999": "some-uuid-value",
	}
	err := locker.ReleaseLocks(context.Background(), lockedKeys)

	// Should not error if key doesn't exist (just skip it)
	require.NoError(t, err)
}

func TestRedisLocker_GetLockedResources(t *testing.T) {
	client := newTestRedisClient(t)
	locker := &RedisLocker{client: client, lockTimeout: time.Minute}

	// Set some locks
	require.NoError(t, client.Set(context.Background(), "node:1", "uuid-1", 0).Err())
	require.NoError(t, client.Set(context.Background(), "node:2", "uuid-2", 0).Err())
	require.NoError(t, client.Set(context.Background(), "node:99", "uuid-99", 0).Err())
	require.NoError(t, client.Set(context.Background(), "cluster:5", "uuid-5", 0).Err())

	// Get all node resources
	resources, err := locker.GetLockedResources(context.Background(), "node:*")

	require.NoError(t, err)
	require.ElementsMatch(t, []string{"node:1", "node:2", "node:99"}, resources)

	// Get all resources
	allResources, err := locker.GetLockedResources(context.Background(), "")
	require.NoError(t, err)
	require.Len(t, allResources, 4)
}
