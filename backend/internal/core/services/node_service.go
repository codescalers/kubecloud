package services

import (
	"context"
	"errors"
	"fmt"
	cfg "kubecloud/internal/config"
	distributedlocks "kubecloud/internal/core/distributed_locks"
	"kubecloud/internal/core/models"
	"kubecloud/internal/core/persistence"
	"kubecloud/internal/core/workflows"
	"kubecloud/internal/infrastructure/gridclient"
	"kubecloud/internal/infrastructure/telemetry"
	"strconv"

	proxyTypes "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
	"github.com/xmonader/ewf"
	"go.opentelemetry.io/otel/attribute"
)

var Zos3NodeFeatures = []string{
	"zmachine",
	"network",
}

type NodeService struct {
	nodesRepo models.UserNodesRepository
	userRepo  models.UserRepository

	appCtx     context.Context
	ewfEngine  *ewf.Engine
	gridClient gridclient.GridClient
	locker          distributedlocks.DistributedLocks
	tracer     *telemetry.ServiceTracer
}

func NewNodeService(
	userNodesRepo models.UserNodesRepository, userRepo models.UserRepository,
	appCtx context.Context, ewfEngine *ewf.Engine, gridClient gridclient.GridClient,
	locker distributedlocks.DistributedLocks,
) NodeService {
	return NodeService{
		nodesRepo:  userNodesRepo,
		userRepo:   userRepo,
		appCtx:     appCtx,
		ewfEngine:  ewfEngine,
		gridClient: gridClient,
		locker:          locker,
		tracer:     telemetry.NewServiceTracer("node_service"),
	}
}

type Pool struct {
	Name string `json:"name"`
	// free space in bytes
	Free uint64 `json:"free"`
	// type of the disk wither ssd or hdd
	Type string `json:"type"`
}

func (svc *NodeService) GetNodes(ctx context.Context, filter proxyTypes.NodeFilter, limit proxyTypes.Limit) ([]proxyTypes.Node, int, error) {
	ctx, span := svc.tracer.StartSpan(ctx, "GetNodes")
	defer span.End()

	span.SetAttributes(
		attribute.String("filter", fmt.Sprintf("%+v", filter)),
		attribute.String("limit", fmt.Sprintf("%+v", limit)),
	)

	nodes, count, err := svc.gridClient.Nodes(ctx, filter, limit)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, 0, err
	}

	span.SetAttributes(attribute.Int("node_count", count))
	return nodes, count, nil
}

func (svc *NodeService) GetZos3Nodes(ctx context.Context, filter proxyTypes.NodeFilter, limit proxyTypes.Limit) ([]proxyTypes.Node, int, error) {
	filter.Features = Zos3NodeFeatures
	return svc.gridClient.Nodes(ctx, filter, limit)
}

func (svc *NodeService) GetUserByID(userID int) (models.User, error) {
	return svc.userRepo.GetUserByID(userID)
}

func (svc *NodeService) GetUserNodeByNodeID(nodeID uint32) (models.UserNodes, error) {
	return svc.nodesRepo.GetUserNodeByNodeID(uint64(nodeID))
}

func (svc *NodeService) CheckUserBalanceForOneHour(ctx context.Context, userMnemonic string, userDebt uint64, nodePriceUsd float64) error {
	_, span := svc.tracer.StartSpan(ctx, "CheckUserBalanceForOneHour")
	defer span.End()

	span.SetAttributes(
		attribute.Float64("node_price_usd", nodePriceUsd),
		attribute.Int64("user_debt", int64(userDebt)),
	)

	// validate user has enough balance for reserving node
	usdMillicentBalance, err := svc.gridClient.GetUserBalanceUSDMillicent(userMnemonic)
	if err != nil {
		telemetry.RecordError(span, err)
		return err
	}

	span.SetAttributes(attribute.Int64("balance_usd_millicent", int64(usdMillicentBalance)))

	//TODO: check price in month constant
	requiredBalance := gridclient.FromUSDToUSDMillicent(nodePriceUsd) / 24 / 30
	span.SetAttributes(attribute.Int64("required_balance", int64(requiredBalance)))

	if usdMillicentBalance-userDebt < requiredBalance {
		err := fmt.Errorf("you should at least have enough balance for one hour")
		telemetry.RecordError(span, err)
		return err
	}

	return nil
}

func (svc *NodeService) GetUserNodeByContractID(contractID uint64) (models.UserNodes, error) {
	return svc.nodesRepo.GetUserNodeByContractID(contractID)
}

func (svc *NodeService) GetTwinIDFromUserID(ctx context.Context, userID int) (uint64, error) {
	_, span := svc.tracer.StartSpan(ctx, "GetTwinIDFromUserID")
	defer span.End()

	span.SetAttributes(attribute.Int("user_id", userID))

	user, err := svc.userRepo.GetUserByID(userID)
	if err != nil {
		telemetry.RecordError(span, err)
		return 0, err
	}

	twinID, err := svc.gridClient.GetTwinIDFromUserMnemonic(user.Mnemonic)
	if err != nil {
		telemetry.RecordError(span, err)
		return 0, err
	}

	span.SetAttributes(attribute.Int64("twin_id", int64(twinID)))
	return twinID, nil
}

func (svc *NodeService) GetTwins(ctx context.Context, filter proxyTypes.TwinFilter, limit proxyTypes.Limit) ([]proxyTypes.Twin, int, error) {
	return svc.gridClient.Twins(ctx, filter, limit)
}

func (svc *NodeService) GetNodePools(ctx context.Context, nodeID uint32) ([]Pool, error) {
	ctx, span := svc.tracer.StartSpan(ctx, "GetNodePools")
	defer span.End()

	span.SetAttributes(attribute.Int64("node_id", int64(nodeID)))

	nc, err := svc.gridClient.GetNodeClient(nodeID)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, err
	}
	storagePool, err := nc.Pools(ctx)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, err
	}

	var pools []Pool
	for _, pool := range storagePool {
		pools = append(pools, Pool{
			Name: pool.Name,
			Free: uint64(pool.Size - pool.Used),
			Type: string(pool.Type),
		})
	}

	span.SetAttributes(attribute.Int("pool_count", len(pools)))
	return pools, nil
}

func (svc *NodeService) GetRentedNodesForUser(ctx context.Context, userID int, healthy bool) ([]proxyTypes.Node, int, error) {
	ctx, span := svc.tracer.StartSpan(ctx, "GetRentedNodesForUser")
	defer span.End()

	span.SetAttributes(
		attribute.Int("user_id", userID),
		attribute.Bool("healthy", healthy),
	)

	twinID, err := svc.GetTwinIDFromUserID(ctx, userID)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, 0, err
	}

	filter := proxyTypes.NodeFilter{
		RentedBy: &twinID,
		Features: Zos3NodeFeatures,
	}

	if healthy {
		filter.Healthy = &healthy
	}

	limit := proxyTypes.DefaultLimit()

	nodes, count, err := svc.GetNodes(ctx, filter, limit)
	if err != nil {
		telemetry.RecordError(span, err)
		return nil, 0, err
	}

	span.SetAttributes(attribute.Int("node_count", count))
	return nodes, count, nil
}

func (svc *NodeService) AsyncReserveNode(userID int, userMnemonic string, nodeID uint32) (string, error) {
	queueName := fmt.Sprintf("%s:user_%d", cfg.DefaultQueueConfig.Name, userID)
	displayName := fmt.Sprintf("Reserving node %d", nodeID)
	metadata := map[string]string{
		"node_id": strconv.FormatUint(uint64(nodeID), 10),
	}

	wf, err := svc.ewfEngine.NewWorkflow(workflows.WorkflowReserveNode, ewf.WithQueue(queueName), ewf.WithDisplayName(displayName), ewf.WithMetadata(metadata))
	if err != nil {
		return "", err
	}

	wf.State = map[string]interface{}{
		"node_id":       nodeID,
		"target_status": workflows.NodeRented,
		"config": map[string]interface{}{
			"user_id":  userID,
			"mnemonic": userMnemonic,
		},
	}

	if err = persistence.SetStateUserID(&wf, userID); err != nil {
		return "", err
	}

	if err = svc.locker.AcquireWorkflowLock(svc.appCtx, []uint32{nodeID}, wf.UUID); err != nil {
		return "", err
	}

	if err = svc.runWithQueue(queueName, &wf); err != nil {
		return "", err
	}

	return wf.UUID, nil
}

func (svc *NodeService) AsyncUnreserveNode(userID int, userMnemonic string, contractID uint64, nodeID uint32) (string, error) {
	queueName := fmt.Sprintf("%s:user_%d", cfg.DefaultQueueConfig.Name, userID)

	displayName := fmt.Sprintf("Unreserving node %d", nodeID)
	metadata := map[string]string{
		"contract_id": strconv.FormatUint(contractID, 10),
		"node_id":     strconv.FormatUint(uint64(nodeID), 10),
	}
	wf, err := svc.ewfEngine.NewWorkflow(workflows.WorkflowUnreserveNode, ewf.WithQueue(queueName), ewf.WithDisplayName(displayName), ewf.WithMetadata(metadata))
	if err != nil {
		return "", err
	}

	wf.State = map[string]interface{}{
		"contract_id":   contractID,
		"node_id":       nodeID,
		"target_status": workflows.NodeRentable,
		"config": map[string]interface{}{
			"user_id":  userID,
			"mnemonic": userMnemonic,
		},
	}

	if err = persistence.SetStateUserID(&wf, userID); err != nil {
		return "", err
	}

	if err = svc.runWithQueue(queueName, &wf); err != nil {
		return "", err
	}

	return wf.UUID, nil
}

// runWithQueue ensures the workflow is run within the specified queue
// if a queue with the given name does not exist, it creates one
// if a non queued workflow is passed, it sets its queue name to run in the specified queue
func (svc *NodeService) runWithQueue(queueName string, wf *ewf.Workflow) error {

	err := svc.ewfEngine.CreateQueue(svc.appCtx, queueName, cfg.DefaultQueueConfig.WorkersDef, cfg.DefaultQueueConfig.QueueOptions)
	if err != nil && !errors.Is(err, ewf.ErrQueueAlreadyExists) {
		return err
	}

	if wf.QueueName == "" {
		wf.QueueName = queueName
	}

	return svc.ewfEngine.Run(svc.appCtx, *wf)
}
