package distributedlocks

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
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

func (l *RedisLocker) AcquireLocks(ctx context.Context, resourceKeys []string) (map[string]string, error) {
	if len(resourceKeys) == 0 {
		return nil, fmt.Errorf("no resource keys provided")
	}

	expiry := int64(l.lockTimeout / time.Millisecond)

	values := make([]string, len(resourceKeys))
	argv := make([]interface{}, 0, len(resourceKeys)+1)
	//expiry of locks
	argv = append(argv, expiry)

	// uuid values for each key
	for i := range resourceKeys {
		val := uuid.New().String()
		values[i] = val
		argv = append(argv, val)
	}

	lua := redis.NewScript(`
local expiry = tonumber(ARGV[1])
local locked = {}

for i = 1, #KEYS do
    local ok = redis.call("SET", KEYS[i], ARGV[i+1], "PX", expiry, "NX")
    if not ok then
        for j = 1, #locked do
            redis.call("DEL", KEYS[j])
        end
        return {"LOCKED", KEYS[i]}
    end
    table.insert(locked, KEYS[i])
end

return {"OK"}
`)

	res, err := lua.Run(ctx, l.client, resourceKeys, argv...).Result()
	if err != nil {
		return nil, err
	}

	out, ok := res.([]interface{})
	if !ok || len(out) == 0 {
		return nil, fmt.Errorf("unexpected script output: %v", res)
	}

	status, _ := out[0].(string)
	if status == "LOCKED" {
		conflict := out[1].(string)
		return nil, fmt.Errorf("%w: %s", ErrResourceLocked, conflict)
	}

	locked := map[string]string{}
	for i, k := range resourceKeys {
		locked[k] = values[i]
	}

	return locked, nil
}

// ReleaseLocks releases the locks for the given keys.
func (l *RedisLocker) ReleaseLocks(ctx context.Context, lockedKeys map[string]string) error {
	if len(lockedKeys) == 0 {
		return nil
	}
	keys := make([]string, 0, len(lockedKeys))
	values := make([]interface{}, 0, len(lockedKeys))

	for k, v := range lockedKeys {
		keys = append(keys, k)
		values = append(values, v)
	}

	luaScript := redis.NewScript(`
local failed = {}
for i = 1, #KEYS do
    local key = KEYS[i]
    local expected = ARGV[i]
    local actual = redis.call("GET", key)

    if actual ~= false then
        if actual ~= expected then
            table.insert(failed, key)
        else
            redis.call("DEL", key)
        end
    end
end
return failed
`)

	// Run the script
	res, err := luaScript.Run(ctx, l.client, keys, values...).Result()
	if err != nil {
		return err
	}

	failedKeys, _ := res.([]interface{})
	if len(failedKeys) > 0 {
		mismatches := make([]string, len(failedKeys))
		for i, v := range failedKeys {
			mismatches[i] = v.(string)
		}
		return fmt.Errorf("lock value mismatch for keys: %v", mismatches)
	}

	return nil
}

// GetLockedResources returns all currently locked resource keys matching the given pattern.
func (l *RedisLocker) GetLockedResources(ctx context.Context, keyPattern string) ([]string, error) {
	if keyPattern == "" {
		keyPattern = "*"
	}
	iter := l.client.Scan(ctx, 0, keyPattern, 0).Iterator()
	resources := make([]string, 0)
	for iter.Next(ctx) {
		resources = append(resources, iter.Val())
	}
	if err := iter.Err(); err != nil {
		return nil, err
	}
	return resources, nil
}
