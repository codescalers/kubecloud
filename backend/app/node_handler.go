package app

import (
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
	substrate "github.com/threefoldtech/tfchain/clients/tfchain-client-go"
	proxyTypes "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
)

// ListNodesInput struct holds data required for listing nodes
type ListNodesInput struct {
	Filter *proxyTypes.NodeFilter `json:"filter"`
	Limit  *proxyTypes.Limit      `json:"limit"`
}

// ListNodesHandler requests all nodes from gridproxy
func (h *Handler) ListNodesHandler(c *gin.Context) {
	var request ListNodesInput

	if err := c.ShouldBindJSON(&request); err != nil {
		Error(c, http.StatusBadRequest, "Bad Request", "Invalid filter/limit payload")
		return
	}

	filter := proxyTypes.NodeFilter{}
	if request.Filter != nil {
		filter = *request.Filter
	}

	healthy := true
	rentable := true

	filter.Healthy = &healthy
	filter.Rentable = &rentable

	limit := proxyTypes.DefaultLimit()
	if request.Limit != nil {
		limit = *request.Limit
	}

	nodes, count, err := h.proxyClient.Nodes(c.Request.Context(), filter, limit)
	if err != nil {
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	Success(c, http.StatusOK, "Nodes retrieved successfully", gin.H{
		"total": count,
		"nodes": nodes,
	})
}

// ReserveNodeHandler reserves node for user
func (h *Handler) ReserveNodeHandler(c *gin.Context) {
	nodeIDParam := c.Param("node_id")
	if nodeIDParam == "" {
		Error(c, http.StatusBadRequest, "Node ID is required", "")
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		log.Error().Msg("user ID not found in context")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	nodeID64, err := strconv.ParseUint(nodeIDParam, 10, 32)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}
	nodeID := uint32(nodeID64)

	userID, ok := userIDVal.(int)
	if !ok {
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	filter := proxyTypes.NodeFilter{
		NodeID: &nodeID64,
	}

	// validate user has enough balance for reserving node
	nodes, _, err := h.proxyClient.Nodes(c.Request.Context(), filter, proxyTypes.Limit{})
	if err != nil {
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	node := nodes[0]
	userBalance := user.CreditCardBalance + user.CreditedBalance
	if node.PriceUsd > userBalance {
		Error(c, http.StatusPaymentRequired, "Insufficient balance", fmt.Sprintf("Node price is %.2f USD/Month, your balance is %.2f USD.  Please charge your balance first to proceed.", node.PriceUsd, userBalance))
		return
	}

	// Create identity from mnemonic
	identity, err := substrate.NewIdentityFromSr25519Phrase(user.Mnemonic)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	contractID, err := h.substrateClient.CreateRentContract(identity, nodeID, nil)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	Success(c, http.StatusOK, "Node rented successfully", gin.H{
		"contract_id": contractID,
		"node_id":     nodeID,
	})

}

// ListReservedNodeHandler list reserved nodes for user on tfchain
func (h *Handler) ListReservedNodeHandler(c *gin.Context) {
	userIDVal, exists := c.Get("user_id")
	if !exists {
		Error(c, http.StatusUnauthorized, "User ID not found in context", "")
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		log.Error().Msg("Invalid user ID or type")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User not found", "")
		return
	}

	identity, err := substrate.NewIdentityFromSr25519Phrase(user.Mnemonic)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	twinID, err := h.substrateClient.GetTwinByPubKey(identity.PublicKey())
	fmt.Println(twinID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	twinID64 := uint64(twinID)
	filter := proxyTypes.NodeFilter{
		RentedBy: &twinID64,
	}

	limit := proxyTypes.DefaultLimit()

	nodes, count, err := h.proxyClient.Nodes(c.Request.Context(), filter, limit)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	Success(c, http.StatusOK, "Nodes retrieved successfully", map[string]interface{}{
		"total": count,
		"nodes": nodes,
	})

}

// UnreserveNodeHandler unreserve node for user
func (h *Handler) UnreserveNodeHandler(c *gin.Context) {
	contractIDParam := c.Param("contract_id")
	if contractIDParam == "" {
		Error(c, http.StatusBadRequest, "Contract ID is required", "")
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		Error(c, http.StatusUnauthorized, "User ID not found in context", "")
		return
	}

	userID, ok := userIDVal.(int)
	if !ok {
		log.Error().Msg("Invalid user ID or type")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusNotFound, "User not found", "")
		return
	}

	identity, err := substrate.NewIdentityFromSr25519Phrase(user.Mnemonic)
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}

	contractID64, err := strconv.ParseUint(contractIDParam, 10, 32)
	if err != nil {
		log.Error().Msg("Invalid contract ID or type")
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}
	contractID := uint32(contractID64)

	err = h.substrateClient.CancelContract(identity, uint64(contractID))
	if err != nil {
		log.Error().Err(err).Send()
		Error(c, http.StatusInternalServerError, "internal server error", "")
		return
	}
	Success(c, http.StatusOK, "Node unreserved successfully", nil)

}
