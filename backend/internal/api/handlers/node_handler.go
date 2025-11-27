package handlers

import (
	"errors"
	"fmt"
	distributedlocks "kubecloud/internal/core/distributed_locks"
	"kubecloud/internal/core/models"
	"math/rand/v2"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"kubecloud/internal/core/services"

	"github.com/gin-gonic/gin"
	proxyTypes "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

type NodeHandler struct {
	svc    services.NodeService
	locker distributedlocks.DistributedLocks
}

func NewNodeHandler(svc services.NodeService, locker distributedlocks.DistributedLocks) NodeHandler {
	return NodeHandler{
		svc:    svc,
		locker: locker,
	}
}

// ListNodesResponse holds the response for reserved nodes
type ListNodesResponse struct {
	Total int               `json:"total"`
	Nodes []proxyTypes.Node `json:"nodes"`
}

type NodesWithDiscount struct {
	proxyTypes.Node
	DiscountPrice float64 `json:"discount_price"`
}

type ListNodesWithDiscountResponse struct {
	Total int                 `json:"total"`
	Nodes []NodesWithDiscount `json:"nodes"`
}

// ReserveNodeResponse holds the response for reserve node response
type ReserveNodeResponse struct {
	WorkflowID string `json:"workflow_id"`
	NodeID     uint32 `json:"node_id"`
	Email      string `json:"email"`
}

// UnreserveNodeResponse holds the response for unreserve node response
type UnreserveNodeResponse struct {
	WorkflowID string `json:"workflow_id"`
	ContractID uint64 `json:"contract_id"`
	Email      string `json:"email"`
}

type TwinResponse struct {
	PublicKey string `json:"public_key"`
	AccountID string `json:"account_id"`
	Relay     string `json:"relay"`
	TwinID    uint   `json:"twin_id"`
}

type NodeStoragePoolResponse struct {
	Pools []services.Pool `json:"pools"`
}

// @Summary List all grid nodes
// @Description List all nodes from the grid proxy (no user-specific filtering)
// @Tags nodes
// @ID list-all-grid-nodes
// @Accept json
// @Produce json
// @Param healthy query bool false "Filter by healthy nodes (default: false)"
// @Param size query int false "Limit the number of nodes returned (default: 50)"
// @Param page query int false "page number (default: 1)"
// @Success 200 {object} APIResponse{data=ListNodesResponse} "All grid nodes retrieved successfully"
// @Failure 400 {object} APIResponse "Invalid filter parameters"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /nodes [get]
func (h *NodeHandler) ListAllGridNodesHandler(c *gin.Context) {
	reqLog := requestLogger(c, "ListAllGridNodesHandler")
	query := c.Request.URL.Query()

	limit := proxyTypes.DefaultLimit()
	limit.RetCount = true
	limit.SortBy = "uptime"
	limit.SortOrder = "desc"
	err := queryParamsToStruct(query, &limit)
	if err != nil {
		BadRequest(c, "Invalid limit params")
		return
	}

	filter := proxyTypes.NodeFilter{}
	err = queryParamsToStruct(query, &filter)
	if err != nil {
		BadRequest(c, "Invalid filter params")
		return
	}

	nodes, count, err := h.svc.GetNodes(c.Request.Context(), filter, limit)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get nodes")
		InternalServerError(c)
		return
	}

	OK(c, "All grid nodes retrieved successfully", ListNodesResponse{
		Total: count,
		Nodes: nodes,
	})
}

// @Summary List nodes
// @Description List nodes from proxy [rented nodes first, then available nodes sorted by uptime]
// @Tags nodes
// @ID list-nodes
// @Accept json
// @Produce json
// @Param healthy query bool false "Filter by healthy nodes (default: true)"
// @Param rentable query bool false "Filter by rentable nodes (default: true)"
// @Param limit query int false "Limit the number of nodes returned (default: 50)"
// @Param offset query int false "Offset for pagination (default: 0)"
// @Success 200 {object} APIResponse "Nodes are retrieved successfully"
// @Failure 400 {object} APIResponse "Invalid filter parameters"
// @Failure 500 {object} APIResponse "Internal server error"
// @Security UserMiddleware
// @Router /user/nodes [get]
func (h *NodeHandler) ListNodesHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "ListNodesHandler")

	rentedNodes, _, err := h.svc.GetRentedNodesForUser(c.Request.Context(), userID, true)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve rented nodes")
		InternalServerError(c)
		return
	}

	query := c.Request.URL.Query()

	limit := proxyTypes.DefaultLimit()
	limit.RetCount = true

	// prioritize nodes by uptime
	limit.SortBy = "uptime"
	limit.SortOrder = "desc"
	err = queryParamsToStruct(query, &limit)
	if err != nil {
		BadRequest(c, "Invalid limit params")
		return
	}

	filter := proxyTypes.NodeFilter{}
	err = queryParamsToStruct(query, &filter)
	if err != nil {
		BadRequest(c, "Invalid filter params")
		return
	}

	twinID, err := h.svc.GetTwinIDFromUserID(c.Request.Context(), userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve twin ID")
		InternalServerError(c)
		return
	}

	healthy := true
	filter.Healthy = &healthy
	filter.AvailableFor = &twinID

	availableNodes, _, err := h.svc.GetZos3Nodes(c.Request.Context(), filter, limit)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve available nodes")
		InternalServerError(c)
		return
	}

	rand.Shuffle(len(availableNodes), func(i, j int) {
		availableNodes[i], availableNodes[j] = availableNodes[j], availableNodes[i]
	})

	// Combine all nodes without duplicates
	var allNodes []proxyTypes.Node
	seen := make(map[int]bool)

	for _, node := range rentedNodes {
		if !seen[node.NodeID] {
			seen[node.NodeID] = true
			allNodes = append(allNodes, node)
		}
	}

	for _, node := range availableNodes {
		if !seen[node.NodeID] {
			seen[node.NodeID] = true
			allNodes = append(allNodes, node)
		}
	}

	unlockedNodes, err := h.svc.FilterLockedNodes(c.Request.Context(), allNodes)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to filter locked nodes")
		InternalServerError(c)
		return
	}

	OK(c, "Nodes retrieved successfully", ListNodesResponse{
		Total: len(unlockedNodes),
		Nodes: unlockedNodes,
	})
}

// @Summary Reserve node
// @Description Reserves a node for a user
// @Tags nodes
// @ID reserve-node
// @Accept json
// @Produce json
// @Param node_id path string true "Node ID"
// @Success 202 {object} APIResponse{data=ReserveNodeResponse} "Node reservation in progress"
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 404 {object} APIResponse "No nodes are available for rent."
// @Failure 500 {object} APIResponse
// @Security UserMiddleware
// @Router /user/nodes/{node_id} [post]
// ReserveNodeHandler reserves node for user
func (h *NodeHandler) ReserveNodeHandler(c *gin.Context) {
	nodeIDParam := c.Param("node_id")
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "ReserveNodeHandler")
	if nodeIDParam == "" {
		BadRequest(c, "Node ID is required")
		return
	}

	nodeID64, err := strconv.ParseUint(nodeIDParam, 10, 32)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to parse node ID")
		InternalServerError(c)
		return
	}
	nodeID := uint32(nodeID64)

	user, err := h.svc.GetUserByID(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve user")
		if errors.Is(err, models.ErrUserNotFound) {
			NotFound(c, "User is not found")
			return
		}
		InternalServerError(c)
		return
	}

	filter := proxyTypes.NodeFilter{NodeID: &nodeID64}

	nodes, _, err := h.svc.GetZos3Nodes(c.Request.Context(), filter, proxyTypes.Limit{})
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve nodes")
		InternalServerError(c)
		return
	}

	if len(nodes) == 0 {
		reqLog.Error().Msg("no nodes are available for rent")
		NotFound(c, "No nodes are available for rent.")
		return
	}

	node := nodes[0]
	if node.Rented {
		BadRequest(c, "Node is already reserved.")
		return
	}

	userNode, err := h.svc.GetUserNodeByNodeID(nodeID)
	if err != nil && !errors.Is(err, models.ErrUserNodeNotFound) {
		reqLog.Error().Err(err).Msg("failed to check node reservation state")
		InternalServerError(c)
		return
	}
	if err == nil && userNode.NodeID != 0 {
		BadRequest(c, "Node is already reserved.")
		return
	}

	if err := h.svc.CheckUserBalanceForOneHour(c.Request.Context(), user.Mnemonic, user.Debt, node.PriceUsd); err != nil {
		reqLog.Error().Err(err).Msg("failed to check user balance")
		BadRequest(c, "You should at least have enough balance for one hour")
		return
	}

	if err = h.locker.AcquireNodesLocks(c.Request.Context(), []uint32{nodeID}); err != nil {
		reqLog.Error().Err(err).Msg("failed to acquire nodes locks")
		if errors.Is(err, distributedlocks.ErrNodeLocked) {
			Conflict(c, err.Error())
			return
		}
		InternalServerError(c)
		return
	}

	wfUUID, err := h.svc.AsyncReserveNode(userID, user.Mnemonic, nodeID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to start workflow to reserve node")
		err = h.locker.ReleaseLock(c.Request.Context(), []uint32{nodeID}, wfUUID)
		if err != nil {
			reqLog.Error().Err(err).Msg("failed to release nodes locks")
		}
		InternalServerError(c)
		return
	}

	Accepted(c, "Node reservation in progress. You can check its status using the workflow id.", ReserveNodeResponse{
		WorkflowID: wfUUID,
		NodeID:     nodeID,
		Email:      user.Email,
	})
}

// @Summary List rentable nodes
// @Description Retrieves a list of rentable nodes from the grid proxy. These are healthy nodes that are available for rent.
// @Tags nodes
// @ID list-rentable-nodes
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=ListNodesWithDiscountResponse} "Rentable nodes retrieved successfully"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /user/nodes/rentable [get]
func (h *NodeHandler) ListRentableNodesHandler(c *gin.Context) {
	reqLog := requestLogger(c, "ListRentableNodesHandler")
	healthy := true
	rentable := true
	filter := proxyTypes.NodeFilter{
		Healthy:  &healthy,
		Rentable: &rentable,
	}

	limit := proxyTypes.DefaultLimit()
	limit.Randomize = true

	nodes, _, err := h.svc.GetZos3Nodes(c.Request.Context(), filter, limit)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve nodes")
		InternalServerError(c)
		return
	}

	unlockedNodes, err := h.svc.FilterLockedNodes(c.Request.Context(), nodes)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to filter locked nodes")
		InternalServerError(c)
		return
	}

	var nodesWithDiscount []NodesWithDiscount
	for _, node := range unlockedNodes {
		nodesWithDiscount = append(nodesWithDiscount, NodesWithDiscount{
			Node:          node,
			DiscountPrice: node.PriceUsd * 0.5,
		})
	}

	OK(c, "Nodes are retrieved successfully", ListNodesWithDiscountResponse{
		Total: len(nodesWithDiscount),
		Nodes: nodesWithDiscount,
	})
}

// @Summary List reserved nodes
// @Description Returns a list of reserved nodes for a user
// @Tags nodes
// @ID list-reserved-nodes
// @Accept json
// @Produce json
// @Success 200 {object} APIResponse{data=ListNodesWithDiscountResponse}
// @Failure 500 {object} APIResponse
// @Security UserMiddleware
// @Router /user/nodes/rented [get]
// ListReservedNodeHandler list reserved nodes for user on tfchain
func (h *NodeHandler) ListRentedNodesHandler(c *gin.Context) {
	userID := c.GetInt("user_id")

	nodes, count, err := h.svc.GetRentedNodesForUser(c.Request.Context(), userID, false)
	if err != nil {
		InternalServerError(c)
		return
	}

	var nodesWithDiscount []NodesWithDiscount
	for _, node := range nodes {
		nodesWithDiscount = append(nodesWithDiscount, NodesWithDiscount{
			Node:          node,
			DiscountPrice: node.PriceUsd * 0.5,
		})
	}

	OK(c, "Nodes are retrieved successfully", ListNodesWithDiscountResponse{
		Total: count,
		Nodes: nodesWithDiscount,
	})
}

// @Summary Unreserve node
// @Description Unreserve a node for a user
// @Tags nodes
// @ID unreserve-node
// @Accept json
// @Produce json
// @Param contract_id path string true "Contract ID"
// @Success 202 {object} APIResponse{data=UnreserveNodeResponse}
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 404 {object} APIResponse "User is not found"
// @Failure 500 {object} APIResponse
// @Security UserMiddleware
// @Router /user/nodes/unreserve/{contract_id} [delete]
// UnreserveNodeHandler unreserve node for user
func (h *NodeHandler) UnreserveNodeHandler(c *gin.Context) {
	contractIDParam := c.Param("contract_id")
	if contractIDParam == "" {
		BadRequest(c, "Contract ID is required")
		return
	}

	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "UnreserveNodeHandler")

	user, err := h.svc.GetUserByID(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to retrieve user")
		if errors.Is(err, models.ErrUserNotFound) {
			NotFound(c, "User is not found")
			return
		}
		InternalServerError(c)
		return
	}

	contractID, err := strconv.ParseUint(contractIDParam, 10, 32)
	if err != nil {
		reqLog.Error().Msg("Invalid contract ID or type")
		InternalServerError(c)
		return
	}

	userNode, err := h.svc.GetUserNodeByContractID(contractID)
	if err != nil {
		if errors.Is(err, models.ErrUserNodeNotFound) {
			NotFound(c, "Contract ID not found for user")
			return
		}
		reqLog.Error().Err(err).Msg("failed to get user node by contract id")
		InternalServerError(c)
		return
	}

	wfUUID, err := h.svc.AsyncUnreserveNode(userID, user.Mnemonic, contractID, userNode.NodeID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to unreserve node")
		InternalServerError(c)
		return
	}

	Accepted(c, "Node unreservation in progress. You can check its status using the workflow id.", UnreserveNodeResponse{
		WorkflowID: wfUUID,
		ContractID: contractID,
		Email:      user.Email,
	})
}

// @Summary Get account ID by twin ID
// @Description Retrieve the account ID associated with a specific twin ID
// @Tags twins
// @Accept json
// @Produce json
// @Param twin_id path int true "Twin ID"
// @Param limit query int false "Pagination limit"
// @Param offset query int false "Pagination offset"
// @Param filterParam  query string false "Other optional filter params"
// @Success 200 {object} APIResponse{data=TwinResponse} "Account ID is retrieved successfully"
// @Failure 400 {object} APIResponse "Bad Request or Invalid params"
// @Failure 404 {object} APIResponse "Twin ID not found"
// @Failure 500 {object} APIResponse "Internal Server Error"
// @Router /twins/{twin_id}/account [get]
func (h *NodeHandler) GetAccountIDHandler(c *gin.Context) {
	reqLog := requestLogger(c, "GetAccountIDHandler")
	twinIDParam := c.Param("twin_id")
	if twinIDParam == "" {
		BadRequest(c, "Twin ID is required")
		return
	}

	query := c.Request.URL.Query()

	limit := proxyTypes.DefaultLimit()
	err := queryParamsToStruct(query, &limit)
	if err != nil {
		BadRequest(c, "Invalid limit params")
		return
	}

	twinID64, err := strconv.ParseUint(twinIDParam, 10, 64)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to parse twin id")
		BadRequest(c, "Error parsing twin id")
		return
	}

	filter := proxyTypes.TwinFilter{}
	filter.TwinID = &twinID64
	err = queryParamsToStruct(query, &filter)
	if err != nil {
		BadRequest(c, "Invalid filter params")
		return
	}

	twins, _, err := h.svc.GetTwins(c.Request.Context(), filter, limit)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get twins")
		InternalServerError(c)
		return
	}

	if len(twins) == 0 {
		NotFound(c, "Twin ID not found")
		return
	}

	OK(c, "Twin Details are retrieved successfully", TwinResponse{
		AccountID: twins[0].AccountID,
		TwinID:    twins[0].TwinID,
		Relay:     twins[0].Relay,
		PublicKey: twins[0].PublicKey,
	})
}

// @Summary Get node storage pool
// @Description Returns node storage pool
// @Tags nodes
// @ID get-node-storage-pool
// @Accept json
// @Produce json
// @Param node_id path string true "Node ID"
// @Success 200 {object} APIResponse{data=NodeStoragePoolResponse} "Node storage pool is retrieved successfully"
// @Failure 400 {object} APIResponse "Bad Request or Invalid params"
// @Failure 404 {object} APIResponse "Node not found"
// @Failure 500 {object} APIResponse "Internal Server Error"
// @Router /nodes/{node_id}/storage-pool [get]
func (h *NodeHandler) GetNodeStoragePoolHandler(c *gin.Context) {
	reqLog := requestLogger(c, "GetNodeStoragePoolHandler")
	nodeIDParam := c.Param("node_id")
	if nodeIDParam == "" {
		BadRequest(c, "Node ID is required")
		return
	}

	nodeID, err := strconv.ParseUint(nodeIDParam, 10, 32)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to parse node id")
		BadRequest(c, "Error parsing node id")
		return
	}

	res, _, err := h.svc.GetNodes(c.Request.Context(), proxyTypes.NodeFilter{NodeID: &nodeID}, proxyTypes.DefaultLimit())
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get node from proxy")
		InternalServerError(c)
		return
	}

	if len(res) == 0 {
		NotFound(c, "Node not found")
		return
	}

	pools, err := h.svc.GetNodePools(c.Request.Context(), uint32(nodeID))
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get node pools")
		InternalServerError(c)
		return
	}

	OK(c, "Node storage pool is retrieved successfully", NodeStoragePoolResponse{Pools: pools})
}

// used to extend the built-in filters with queries from the request
func queryParamsToStruct(query url.Values, result interface{}) error {
	v := reflect.ValueOf(result).Elem()
	t := v.Type()

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		value := v.Field(i)

		paramName := field.Tag.Get("schema")
		if paramName == "" {
			paramName = field.Name
		}
		paramName = strings.Split(paramName, ",")[0]

		paramValues, ok := query[paramName]
		if !ok || len(paramValues) == 0 {
			continue
		}

		switch value.Kind() {
		case reflect.Slice:
			elemType := value.Type().Elem()
			slice := reflect.MakeSlice(value.Type(), 0, len(paramValues))
			for _, pv := range paramValues {
				elem := reflect.New(elemType).Elem()
				if err := setValueFromString(elem, pv); err != nil {
					return err
				}
				slice = reflect.Append(slice, elem)
			}
			value.Set(slice)

		case reflect.Ptr:
			ptr := reflect.New(value.Type().Elem())
			if err := setValueFromString(ptr.Elem(), paramValues[0]); err != nil {
				return err
			}
			value.Set(ptr)

		default:
			if err := setValueFromString(value, paramValues[0]); err != nil {
				return err
			}
		}
	}
	return nil
}

func setValueFromString(v reflect.Value, s string) error {
	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Int, reflect.Int64, reflect.Int32:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint64, reflect.Uint32:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)
	default:
		return fmt.Errorf("unsupported kind: %s", v.Kind())
	}
	return nil
}
