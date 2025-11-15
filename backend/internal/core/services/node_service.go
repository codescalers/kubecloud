package services

import (
	"kubecloud/internal/shared"
	"context"
	"errors"
	"fmt"
	"kubecloud/internal/infrastructure/substrate"
	"kubecloud/internal/core/models"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	proxyTypes "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
	"github.com/xmonader/ewf"
)

var Zos3NodeFeatures = []string{
	"zmachine",
	"network",
}

type NodeService struct {
	nodesRepo models.UserNodesRepository
	userRepo  models.UserRepository

	appCtx          context.Context
	ewfEngine       *ewf.Engine
	gridClient      deployer.TFPluginClient
	substrateClient substrate.Substrate
}

func NewNodeService(
	userNodesRepo models.UserNodesRepository, userRepo models.UserRepository,
	appCtx context.Context, ewfEngine *ewf.Engine, gridClient deployer.TFPluginClient,
	substrateClient substrate.Substrate,
) NodeService {
	return NodeService{
		nodesRepo:       userNodesRepo,
		userRepo:        userRepo,
		appCtx:          appCtx,
		ewfEngine:       ewfEngine,
		gridClient:      gridClient,
		substrateClient: substrateClient,
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
	return svc.gridClient.GridProxyClient.Nodes(ctx, filter, limit)
}

func (svc *NodeService) GetZos3Nodes(ctx context.Context, filter proxyTypes.NodeFilter, limit proxyTypes.Limit) ([]proxyTypes.Node, int, error) {
	filter.Features = Zos3NodeFeatures
	return svc.gridClient.GridProxyClient.Nodes(ctx, filter, limit)
}

func (svc *NodeService) GetUserByID(userID int) (models.User, error) {
	return svc.userRepo.GetUserByID(userID)
}

func (svc *NodeService) GetUserNodeByNodeID(nodeID uint32) (models.UserNodes, error) {
	return svc.nodesRepo.GetUserNodeByNodeID(uint64(nodeID))
}

func (svc *NodeService) CheckUserBalanceForOneHour(userMnemonic string, userDebt uint64, nodePriceUsd float64) error {
	// validate user has enough balance for reserving node
	usdMillicentBalance, err := svc.substrateClient.GetUserBalanceUSDMillicent(userMnemonic)
	if err != nil {
		return err
	}

	//TODO: check price in month constant
	if usdMillicentBalance-userDebt < substrate.FromUSDToUSDMillicent(nodePriceUsd)/24/30 {
		return fmt.Errorf("you should at least have enough balance for one hour")
	}

	return nil
}

func (svc *NodeService) GetUserNodeByContractID(contractID uint64) (models.UserNodes, error) {
	return svc.nodesRepo.GetUserNodeByContractID(contractID)
}

func (svc *NodeService) GetTwinIDFromUserID(userID int) (uint64, error) {
	user, err := svc.userRepo.GetUserByID(userID)
	if err != nil {
		return 0, err
	}

	identity, err := svc.substrateClient.NewIdentityFromSr25519Phrase(user.Mnemonic)
	if err != nil {
		return 0, err
	}

	twinID, err := svc.substrateClient.GetTwinByPubKey(identity.PublicKey())
	if err != nil {
		return 0, err
	}

	return uint64(twinID), nil
}

func (svc *NodeService) GetTwins(ctx context.Context, filter proxyTypes.TwinFilter, limit proxyTypes.Limit) ([]proxyTypes.Twin, int, error) {
	return svc.gridClient.GridProxyClient.Twins(ctx, filter, limit)
}

func (h *NodeService) GetNodePools(ctx context.Context, nodeID uint32) ([]Pool, error) {
	nc, err := h.gridClient.NcPool.GetNodeClient(h.gridClient.SubstrateConn, nodeID)
	if err != nil {
		return nil, err
	}

	storagePool, err := nc.Pools(ctx)
	if err != nil {
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

	return pools, nil
}

func (svc *NodeService) GetRentedNodesForUser(ctx context.Context, userID int, healthy bool) ([]proxyTypes.Node, int, error) {
	twinID, err := svc.GetTwinIDFromUserID(userID)
	if err != nil {
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
		return nil, 0, err
	}

	return nodes, count, nil
}

func (svc *NodeService) AsyncReserveNode(userID int, userMnemonic string, nodeID uint32) (string, error) {
	wf, err := svc.ewfEngine.NewWorkflow(shared.WorkflowReserveNode)
	if err != nil {
		return "", err
	}

	wf.State = map[string]interface{}{
		"user_id":       userID,
		"mnemonic":      userMnemonic,
		"node_id":       nodeID,
		"target_status": shared.NodeRented,
	}

	if err = svc.runAsyncWithQueue(userID, wf); err != nil {
		return "", err
	}

	return wf.UUID, nil
}

func (svc *NodeService) AsyncUnreserveNode(userID int, userMnemonic string, contractID uint64, nodeID uint32) (string, error) {
	wf, err := svc.ewfEngine.NewWorkflow(shared.WorkflowUnreserveNode)
	if err != nil {
		return "", err
	}

	wf.State = map[string]interface{}{
		"user_id":       userID,
		"mnemonic":      userMnemonic,
		"contract_id":   contractID,
		"node_id":       nodeID,
		"target_status": shared.NodeRentable,
	}

	if err = svc.runAsyncWithQueue(userID, wf); err != nil {
		return "", err
	}

	return wf.UUID, nil
}

func (svc *NodeService) runAsyncWithQueue(userID int, wf *ewf.Workflow) error {
	queueName := fmt.Sprintf("%s:user_%d", shared.DefaultQueueConfig.Name, userID)

	err := svc.ewfEngine.CreateQueue(svc.appCtx, queueName, shared.DefaultQueueConfig.WorkersDef, shared.DefaultQueueConfig.QueueOptions)
	if err != nil && !errors.Is(err, ewf.ErrQueueAlreadyExists) {
		return err
	}

	return svc.ewfEngine.RunAsync(svc.appCtx, wf, ewf.WithQueue(queueName))
}
