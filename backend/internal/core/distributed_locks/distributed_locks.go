package distributedlocks

import (
	"context"
	"errors"
)

var ErrNodeLocked = errors.New("node is currently locked by another request")

// DistributedLocks is an interface that defines the methods for distributed locks.
type DistributedLocks interface {
	AcquireNodesLocks(ctx context.Context, nodeIDs []uint32) (map[string]string, error)
	ReleaseLock(ctx context.Context, lockedKeys map[string]string) error
	GetLockedNodes(ctx context.Context) ([]uint32, error)
}
