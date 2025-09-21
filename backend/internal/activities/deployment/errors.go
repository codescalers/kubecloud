package deployment

import (
	"fmt"
	"strings"

	"github.com/xmonader/ewf"
)

func isWorkloadAlreadyDeployedError(err error) bool {
	return strings.Contains(err.Error(), "exists: conflict")
}

func isWorkloadInvalid(err error) bool {
	return strings.Contains(err.Error(), "invalid deployment")
}

func handleDeploymentError(err error, stepCtx *StepContext, resourceType string, resourceName ...string) error {
	stepCtx.Metrics.IncrementClusterDeploymentFailure()

	switch {
	case isWorkloadAlreadyDeployedError(err):
		return fmt.Errorf("%s already deployed for cluster %s: %w", resourceType, stepCtx.Cluster.Name, ewf.ErrFailWorkflowNow)
	case isWorkloadInvalid(err):
		return fmt.Errorf("%s invalid for cluster %s: %w", resourceType, stepCtx.Cluster.Name, ewf.ErrFailWorkflowNow)
	default:
		return fmt.Errorf("failed to deploy %s: %w", resourceType, err)
	}
}
