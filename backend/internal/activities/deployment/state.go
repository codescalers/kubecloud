package deployment

import (
	"errors"
	"fmt"
	"kubecloud/internal/metrics"
	"kubecloud/internal/statemanager"
	"kubecloud/kubedeployer"

	"github.com/xmonader/ewf"
)

var (
	ErrConfig     = errors.New("config error")
	ErrKubeClient = errors.New("kube client error")
	ErrCluster    = errors.New("cluster error")
)

// StepContext encapsulates all common dependencies for workflow steps
type StepContext struct {
	State      ewf.State
	Config     statemanager.ClientConfig
	KubeClient *kubedeployer.Client
	Cluster    kubedeployer.Cluster
	Metrics    *metrics.Metrics
}

func NewStepContext(state ewf.State, metrics *metrics.Metrics) (*StepContext, error) {
	config, err := GetFromState[statemanager.ClientConfig](state, "config")
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrConfig, err)
	}

	kubeClient, err := statemanager.GetKubeClient(state, config)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrKubeClient, err)
	}

	cluster, err := statemanager.GetCluster(state)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrCluster, err)
	}

	return &StepContext{
		State:      state,
		Config:     config,
		KubeClient: kubeClient,
		Cluster:    cluster,
		Metrics:    metrics,
	}, nil
}

func (ctx *StepContext) SaveState() {
	statemanager.SaveGridClientState(ctx.State, ctx.KubeClient)
	statemanager.StoreCluster(ctx.State, ctx.Cluster)
}

// Used for other state key retrievals with type safety
func GetFromState[T any](state ewf.State, key string) (T, error) {
	var zero T

	value, ok := state[key]
	if !ok {
		return zero, fmt.Errorf("missing '%s' in state", key)
	}

	typedValue, ok := value.(T)
	if !ok {
		return zero, fmt.Errorf("invalid type for '%s' in state: expected %T, got %T", key, zero, value)
	}

	return typedValue, nil
}
