package distributedlocks

import (
	"context"
	"errors"
)

var ErrResourceLocked = errors.New("resource is currently locked by another request")

const (
	NodeLockPrefix = "node:"
)

type DistributedLocks interface {
	AcquireLocks(ctx context.Context, resourceKeys []string) (map[string]string, error)
	ReleaseLocks(ctx context.Context, lockedKeys map[string]string) error
	GetLockedResources(ctx context.Context, keyPattern string) ([]string, error)
}
