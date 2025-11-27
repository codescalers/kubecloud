package distributedlocks

import (
	"context"
)

type DistributedLocks interface {
	AcquireNodesLocks(ctx context.Context, nodeIDs []uint32) error
	AcquireWorkflowLock(ctx context.Context, nodeID uint32, workflowID string) (bool, error)
	ReleaseLock(ctx context.Context, nodeID uint32, workflowID string) error
	GetAllWorkflowsLocks(ctx context.Context) ([]string, error)
}
