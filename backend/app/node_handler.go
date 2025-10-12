package app

import (
	"context"
	"errors"
	"fmt"
	"kubecloud/internal"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"kubecloud/internal/constants"
	"kubecloud/internal/logger"

	"github.com/gin-gonic/gin"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	proxyTypes "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
	"gorm.io/gorm"
)

var (
	zos3NodeFeatures = []string{
		"zmachine",
		"network",
	}
)

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
	ContractID uint32 `json:"contract_id"`
	Email      string `json:"email"`
}

type TwinResponse struct {
	PublicKey string `json:"public_key"`
	AccountID string `json:"account_id"`
	Relay     string `json:"relay"`
	TwinID    uint   `json:"twin_id"`
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
func (h *Handler) ListAllGridNodesHandler(c *gin.Context) {
	query := c.Request.URL.Query()

	limit := proxyTypes.DefaultLimit()
	limit.RetCount = true
	err := queryParamsToStruct(query, &limit)
	if err != nil {
		Error(c, http.StatusBadRequest, "Bad Request", "Invalid limit params")
		return
	}

	filter := proxyTypes.NodeFilter{}
	err = queryParamsToStruct(query, &filter)
	if err != nil {
		Error(c, http.StatusBadRequest, "Bad Request", "Invalid filter params")
		return
	}

	nodes, count, err := h.proxyClient.Nodes(c.Request.Context(), filter, limit)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	Success(c, http.StatusOK, "All grid nodes retrieved successfully", ListNodesResponse{
		Total: count,
		Nodes: nodes,
	})
}

// @Summary List nodes
// @Description List nodes from proxy [rented nodes first + randomized shared nodes]
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
func (h *Handler) ListNodesHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	rentedNodes, rentedNodesCount, err := h.getRentedNodesForUser(c.Request.Context(), userID, true)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	query := c.Request.URL.Query()

	limit := proxyTypes.DefaultLimit()
	limit.RetCount = true
	limit.Randomize = true
	err = queryParamsToStruct(query, &limit)
	if err != nil {
		Error(c, http.StatusBadRequest, "Bad Request", "Invalid limit params")
		return
	}

	filter := proxyTypes.NodeFilter{}
	err = queryParamsToStruct(query, &filter)
	if err != nil {
		Error(c, http.StatusBadRequest, "Bad Request", "Invalid filter params")
		return
	}

	twinID, err := h.getTwinIDFromUserID(userID)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	healthy := true
	filter.Healthy = &healthy
	filter.AvailableFor = &twinID
	filter.Features = zos3NodeFeatures
	availableNodes, availableNodesCount, err := h.proxyClient.Nodes(c.Request.Context(), filter, limit)
	if err != nil {
		InternalServerError(c)
		return
	}

	// Combine all nodes without duplicates
	var allNodes []proxyTypes.Node
	duplicatesCount := 0
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
		} else {
			duplicatesCount++
		}
	}

	Success(c, http.StatusOK, "Nodes retrieved successfully", ListNodesResponse{
		Total: rentedNodesCount + availableNodesCount - duplicatesCount,
		Nodes: allNodes,
	})
}

// @Summary Reserve node
// @Description Reserves a node for a user
// @Tags nodes
// @ID reserve-node
// @Accept json
// @Produce json
// @Param node_id path string true "Node ID"
// @Success 202 {object} ReserveNodeResponse
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 404 {object} APIResponse "No nodes are available for rent."
// @Failure 500 {object} APIResponse
// @Security UserMiddleware
// @Router /user/nodes/{node_id} [post]
// ReserveNodeHandler reserves node for user
func (h *Handler) ReserveNodeHandler(c *gin.Context) {
	nodeIDParam := c.Param("node_id")
	if nodeIDParam == "" {
		Error(c, http.StatusBadRequest, "Node ID is required", "")
		return
	}

	nodeID64, err := strconv.ParseUint(nodeIDParam, 10, 32)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		InternalServerError(c)
		return
	}
	nodeID := uint32(nodeID64)

	userID := c.GetInt("user_id")

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	filter := proxyTypes.NodeFilter{
		NodeID:   &nodeID64,
		Features: zos3NodeFeatures,
	}

	nodes, _, err := h.proxyClient.Nodes(c.Request.Context(), filter, proxyTypes.Limit{})
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		InternalServerError(c)
		return
	}
	if len(nodes) == 0 {
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusNotFound, "No nodes are available for rent.", "")
		return
	}
	node := nodes[0]

	if node.Rented {
		Error(c, http.StatusBadRequest, "Node is already reserved.", "")
		return
	}

	userNode, err := h.db.GetUserNodeByNodeID(uint64(nodeID))
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.GetLogger().Error().Err(err).Uint32("node_id", nodeID).Msg("failed to check node reservation state")
		InternalServerError(c)
		return
	}
	if err == nil && userNode.NodeID != 0 {
		Error(c, http.StatusBadRequest, "Node is already reserved.", "")
		return
	}
	// validate user has enough balance for reserving node
	usdMillicentBalance, err := internal.GetUserBalanceUSDMillicent(h.substrateClient, user.Mnemonic)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	//TODO: check price in month constant
	if usdMillicentBalance-user.Debt < internal.FromUSDToUSDMillicent(node.PriceUsd)/24/30 {
		Error(c, http.StatusBadRequest, "You should at least have enough balance for one hour", "")
		return
	}

	wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowReserveNode)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		InternalServerError(c)
		return
	}

	wf.State = map[string]interface{}{
		"user_id":       userID,
		"mnemonic":      user.Mnemonic,
		"node_id":       nodeID,
		"target_status": constants.NodeRented,
	}

	h.ewfEngine.RunAsync(c, wf)

	Success(c, http.StatusAccepted, "Node reservation in progress. You can check its status using the workflow id.", ReserveNodeResponse{
		WorkflowID: wf.UUID,
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
func (h *Handler) ListRentableNodesHandler(c *gin.Context) {
	healthy := true
	rentable := true
	filter := proxyTypes.NodeFilter{
		Healthy:  &healthy,
		Rentable: &rentable,
		Features: zos3NodeFeatures,
	}

	limit := proxyTypes.DefaultLimit()
	limit.Randomize = true

	nodes, count, err := h.proxyClient.Nodes(c.Request.Context(), filter, limit)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
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
	Success(c, http.StatusOK, "Nodes are retrieved successfully", ListNodesWithDiscountResponse{
		Total: count,
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
func (h *Handler) ListRentedNodesHandler(c *gin.Context) {
	userID := c.GetInt("user_id")
	nodes, count, err := h.getRentedNodesForUser(c.Request.Context(), userID, false)
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
	Success(c, http.StatusOK, "Nodes are retrieved successfully", ListNodesWithDiscountResponse{
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
// @Success 202 {object} UnreserveNodeResponse
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 404 {object} APIResponse "User is not found"
// @Failure 500 {object} APIResponse
// @Security UserMiddleware
// @Router /user/nodes/unreserve/{contract_id} [delete]
// UnreserveNodeHandler unreserve node for user
func (h *Handler) UnreserveNodeHandler(c *gin.Context) {
	contractIDParam := c.Param("contract_id")
	if contractIDParam == "" {
		Error(c, http.StatusBadRequest, "Contract ID is required", "")
		return
	}

	userID := c.GetInt("user_id")

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User is not found", "")
		return
	}

	contractID64, err := strconv.ParseUint(contractIDParam, 10, 32)
	if err != nil {
		logger.GetLogger().Error().Msg("Invalid contract ID or type")
		InternalServerError(c)
		return
	}
	contractID := uint32(contractID64)
	userNode, err := h.db.GetUserNodeByContractID(contractID64)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			Error(c, http.StatusNotFound, "Contract ID not found", "Could not find contract ID in user nodes")
			return
		}
		logger.GetLogger().Error().Err(err)
		InternalServerError(c)
		return
	}

	wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowUnreserveNode)
	if err != nil {
		logger.GetLogger().Error().Err(err)
		InternalServerError(c)
		return
	}

	wf.State = map[string]interface{}{
		"user_id":       userID,
		"mnemonic":      user.Mnemonic,
		"contract_id":   contractID,
		"node_id":       userNode.NodeID,
		"target_status": constants.NodeRentable,
	}

	h.ewfEngine.RunAsync(c, wf)

	Success(c, http.StatusAccepted, "Node unreservation in progress. You can check its status using the workflow id.", UnreserveNodeResponse{
		WorkflowID: wf.UUID,
		ContractID: contractID,
		Email:      user.Email,
	})
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

func (h *Handler) getTwinIDFromUserID(userID int) (uint64, error) {
	user, err := h.db.GetUserByID(userID)
	if err != nil {
		return 0, err
	}

	identity, err := substrate.NewIdentityFromSr25519Phrase(user.Mnemonic)
	if err != nil {
		return 0, err
	}

	twinID, err := h.substrateClient.GetTwinByPubKey(identity.PublicKey())
	if err != nil {
		return 0, err
	}

	return uint64(twinID), nil
}

func (h *Handler) getRentedNodesForUser(ctx context.Context, userID int, healthy bool) ([]proxyTypes.Node, int, error) {
	twinID, err := h.getTwinIDFromUserID(userID)
	if err != nil {
		return nil, 0, err
	}

	filter := proxyTypes.NodeFilter{
		RentedBy: &twinID,
		Features: zos3NodeFeatures,
	}

	if healthy {
		filter.Healthy = &healthy
	}

	limit := proxyTypes.DefaultLimit()

	nodes, count, err := h.proxyClient.Nodes(ctx, filter, limit)
	if err != nil {
		return nil, 0, err
	}

	return nodes, count, nil
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
// @Success 200 {object} TwinResponse "Account ID is retrieved successfully"
// @Failure 400 {object} APIResponse "Bad Request or Invalid params"
// @Failure 404 {object} APIResponse "Twin ID not found"
// @Failure 500 {object} APIResponse "Internal Server Error"
// @Router /twins/{twin_id}/account [get]
func (h *Handler) GetAccountIDHandler(c *gin.Context) {
	twinIDParam := c.Param("twin_id")
	if twinIDParam == "" {
		Error(c, http.StatusBadRequest, "Twin ID is required", "")
		return
	}

	query := c.Request.URL.Query()

	limit := proxyTypes.DefaultLimit()
	err := queryParamsToStruct(query, &limit)
	if err != nil {
		Error(c, http.StatusBadRequest, "Bad Request", "Invalid limit params")
		return
	}

	twinID64, err := strconv.ParseUint(twinIDParam, 10, 64)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Bad Request", "Error parsing twin id")
		return
	}

	filter := proxyTypes.TwinFilter{}
	filter.TwinID = &twinID64
	err = queryParamsToStruct(query, &filter)
	if err != nil {
		Error(c, http.StatusBadRequest, "Bad Request", "Invalid filter params")
		return
	}

	twins, _, err := h.proxyClient.Twins(c.Request.Context(), filter, limit)
	if err != nil {
		InternalServerError(c)
		return
	}

	if len(twins) == 0 {
		Error(c, http.StatusNotFound, "Twin ID not found", "")
		return
	}
	Success(c, http.StatusOK, "Twin Details are retrieved successfully", TwinResponse{
		AccountID: twins[0].AccountID,
		TwinID:    twins[0].TwinID,
		Relay:     twins[0].Relay,
		PublicKey: twins[0].PublicKey,
	})

}

type Pool struct {
	Name string `json:"name"`
	// free space in bytes
	Free uint64 `json:"free"`
	// type of the disk wither ssd or hdd
	Type string `json:"type"`
}

type NodeStoragePoolResponse struct {
	Pools []Pool `json:"pools"`
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
func (h *Handler) GetNodeStoragePoolHandler(c *gin.Context) {
	nodeIDParam := c.Param("node_id")
	if nodeIDParam == "" {
		Error(c, http.StatusBadRequest, "Node ID is required", "")
		return
	}

	nodeID, err := strconv.ParseUint(nodeIDParam, 10, 32)
	if err != nil {
		logger.GetLogger().Error().Err(err).Send()
		Error(c, http.StatusBadRequest, "Bad Request", "Error parsing node id")
		return
	}

	res, _, err := h.proxyClient.Nodes(c.Request.Context(), proxyTypes.NodeFilter{NodeID: &nodeID}, proxyTypes.DefaultLimit())
	if err != nil {
		Error(c, http.StatusNotFound, "failed to get node", "")
		return
	}
	if len(res) == 0 {
		Error(c, http.StatusNotFound, "Node not found", "")
		return
	}

	nc, err := h.gridClient.NcPool.GetNodeClient(h.gridClient.SubstrateConn, uint32(nodeID))
	if err != nil {
		InternalServerError(c)
		return
	}

	storagePool, err := nc.Pools(c.Request.Context())
	if err != nil {
		InternalServerError(c)
		return
	}

	var pools []Pool
	for _, pool := range storagePool {
		pools = append(pools, Pool{
			Name: pool.Name,
			Free: uint64(pool.Size - pool.Used),
			Type: string(pool.Type),
		})
	}

	Success(c, http.StatusOK, "Node storage pool is retrieved successfully", NodeStoragePoolResponse{Pools: pools})
}
