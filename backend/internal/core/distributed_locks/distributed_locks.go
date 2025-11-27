package distributedlocks

import (
	"context"
	"errors"
)

var ErrNodeLocked = errors.New("node is currently locked by another request")

// DistributedLocks is an interface that defines the methods for distributed locks.
type DistributedLocks interface {
	AcquireNodesLocks(ctx context.Context, nodeIDs []uint32) error
	AcquireWorkflowLock(ctx context.Context, nodeIDs []uint32, workflowID string) error
	ReleaseLock(ctx context.Context, nodeID uint32, workflowID string) error
	GetAllWorkflowsLocks(ctx context.Context) ([]string, error)
}
