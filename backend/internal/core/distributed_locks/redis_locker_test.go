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

func TestRedisLocker_AcquireNodesLocks_Success(t *testing.T) {
	client := newTestRedisClient(t)
	locker := &RedisLocker{client: client, lockTimeout: time.Minute}

	err := locker.AcquireNodesLocks(context.Background(), []uint32{1, 2})

	require.NoError(t, err)
	require.Equal(t, int64(1), client.Exists(context.Background(), "locked:1").Val())
	require.Equal(t, int64(1), client.Exists(context.Background(), "locked:2").Val())
}

func TestRedisLocker_AcquireNodesLocks_NodeAlreadyLocked(t *testing.T) {
	client := newTestRedisClient(t)
	locker := &RedisLocker{client: client, lockTimeout: time.Minute}

	require.NoError(t, client.Set(context.Background(), "locked:2", 1, 0).Err())

	err := locker.AcquireNodesLocks(context.Background(), []uint32{1, 2})

	require.Error(t, err)
	require.Contains(t, err.Error(), "failed to acquire lock for key locked:2")
	require.Equal(t, int64(0), client.Exists(context.Background(), "locked:1").Val(), "previous locks should be rolled back")
}

func TestRedisLocker_AcquireWorkflowLock(t *testing.T) {
	client := newTestRedisClient(t)
	locker := &RedisLocker{client: client, lockTimeout: time.Minute}

	err := locker.AcquireWorkflowLock(context.Background(), []uint32{1}, "wf-1")
	require.NoError(t, err)

	err = locker.AcquireWorkflowLock(context.Background(), []uint32{1}, "wf-1")
	require.Error(t, err)
}

func TestRedisLocker_ReleaseLock(t *testing.T) {
	client := newTestRedisClient(t)
	locker := &RedisLocker{client: client, lockTimeout: time.Minute}

	require.NoError(t, client.Set(context.Background(), "locked:1", 1, 0).Err())
	require.NoError(t, client.Set(context.Background(), "used:1:wf-1", 1, 0).Err())

	err := locker.ReleaseLock(context.Background(), 1, "wf-1")

	require.NoError(t, err)
	require.Equal(t, int64(0), client.Exists(context.Background(), "locked:1").Val())
	require.Equal(t, int64(0), client.Exists(context.Background(), "used:1:wf-1").Val())
}

func TestRedisLocker_GetAllWorkflowsLocks(t *testing.T) {
	client := newTestRedisClient(t)
	locker := &RedisLocker{client: client, lockTimeout: time.Minute}

	require.NoError(t, client.Set(context.Background(), "used:1:wf-1", 1, 0).Err())
	require.NoError(t, client.Set(context.Background(), "used:2:wf-2", 1, 0).Err())
	require.NoError(t, client.Set(context.Background(), "locked:99", 1, 0).Err())

	keys, err := locker.GetAllWorkflowsLocks(context.Background())

	require.NoError(t, err)
	require.ElementsMatch(t, []string{"used:1:wf-1", "used:2:wf-2"}, keys)
}
