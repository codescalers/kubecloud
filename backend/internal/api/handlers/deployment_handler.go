package handlers

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"

	"kubecloud/internal/core/models"
	"kubecloud/internal/core/services"
	"kubecloud/internal/deployment/kubedeployer"
)

type DeploymentHandler struct {
	svc            services.DeploymentService
	billingService services.BillingService
}

func NewDeploymentHandler(svc services.DeploymentService, billingService services.BillingService) DeploymentHandler {
	return DeploymentHandler{
		svc:            svc,
		billingService: billingService,
	}
}

// Response represents the response structure for deployment requests
type DeploymentWorkflowResponse struct {
	WorkflowID string `json:"task_id"`
	Status     string `json:"status"`
}

// DeploymentListResponse represents the response for listing deployments
type DeploymentListResponse struct {
	Deployments []services.ClusterData `json:"deployments"`
	Count       int                    `json:"count"`
}

// KubeconfigResponse represents the response for kubeconfig requests
type KubeconfigResponse struct {
	Kubeconfig string `json:"kubeconfig"`
}

// ClusterInput represents the simplified input structure for cluster creation
type ClusterInput struct {
	Name  string      `json:"name" binding:"required"`
	Token string      `json:"token"`
	Nodes []NodeInput `json:"nodes" binding:"required"`
}

// NodeInput represents the input structure for node configuration
type NodeInput struct {
	Name       string            `json:"name" binding:"required"`
	Type       string            `json:"type" binding:"required" enums:"worker,master,leader"`
	NodeID     uint32            `json:"node_id" binding:"required"`
	CPU        uint8             `json:"cpu" binding:"required"`
	Memory     uint64            `json:"memory" binding:"required"`    // Memory in MB
	RootSize   uint64            `json:"root_size" binding:"required"` // Storage in MB
	DataDisks  []uint64          `json:"data_disks"`                   // Storage in MB
	EnvVars    map[string]string `json:"env_vars"`                     // SSH_KEY, etc.
	GPUIDs     []string          `json:"gpu_ids,omitempty"`            // List of GPU IDs
	Flist      string            `json:"flist,omitempty"`
	Entrypoint string            `json:"entrypoint,omitempty"`
}

// @Summary List deployments
// @Description Retrieves a list of all deployments (clusters) for the authenticated user
// @Tags deployments
// @Security BearerAuth
// @Produce json
// @Success 200 {object} APIResponse{data=DeploymentListResponse} "Deployments retrieved successfully"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments [get]
func (h *DeploymentHandler) HandleListDeployments(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "HandleListDeployments")
	if userID == 0 {
		Unauthorized(c, "user not authenticated")
		return
	}

	deployments, err := h.svc.ListUserClustersData(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to list user clusters")
		InternalServerError(c)
		return
	}

	OK(c, "Deployments retrieved successfully", gin.H{
		"deployments": deployments,
		"count":       len(deployments),
	})
}

// @Summary Get deployment
// @Description Retrieves details of a specific deployment by name
// @Tags deployments
// @Security BearerAuth
// @Produce json
// @Param name path string true "Deployment name"
// @Success 200 {object} APIResponse{data=services.ClusterData} "Deployment details retrieved successfully"
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 404 {object} APIResponse "Deployment not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments/{name} [get]
func (h *DeploymentHandler) HandleGetDeployment(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "HandleGetDeployment")
	if userID == 0 {
		Unauthorized(c, "unauthorized")
		return
	}

	projectName := c.Param("name")
	if projectName == "" {
		BadRequest(c, "Project name is required")
		return
	}

	projectName = kubedeployer.GetProjectName(userID, projectName)
	logWithProject := reqLog.With().Str("project_name", projectName).Logger()
	reqLog = &logWithProject

	cluster, err := h.svc.GetClusterDataByProjectName(userID, projectName)
	if err != nil {
		if errors.Is(err, models.ErrClusterNotFound) {
			reqLog.Error().Err(err).Msg("Deployment not found")
			NotFound(c, "Deployment not found")
			return
		}

		reqLog.Error().Err(err).Msg("Database error when looking up deployment")
		InternalServerError(c)
		return
	}

	OK(c, "Deployment details retrieved successfully", cluster)
}

// @Summary Get kubeconfig
// @Description Retrieves the kubeconfig file for a specific deployment
// @Tags deployments
// @Security BearerAuth
// @Produce json
// @Param name path string true "Deployment name"
// @Success 200 {object} APIResponse{data=KubeconfigResponse} "Kubeconfig retrieved successfully"
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 404 {object} APIResponse "Deployment not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments/{name}/kubeconfig [get]
func (h *DeploymentHandler) HandleGetKubeconfig(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "HandleGetKubeconfig")
	if userID == 0 {
		Unauthorized(c, "User not authenticated")
		return
	}

	projectName := c.Param("name")
	if projectName == "" {
		BadRequest(c, "Project name is required")
		return
	}

	projectName = kubedeployer.GetProjectName(userID, projectName)
	logWithProject := reqLog.With().Str("project_name", projectName).Logger()
	reqLog = &logWithProject

	cluster, err := h.svc.GetClusterByName(userID, projectName)
	if err != nil {
		if errors.Is(err, models.ErrClusterNotFound) {
			reqLog.Error().Err(err).Msg("Deployment not found")
			NotFound(c, "Deployment not found")
			return
		}
		reqLog.Error().Err(err).Msg("Database error when looking up deployment for kubeconfig")
		InternalServerError(c)
		return
	}

	kubeconfig, err := h.svc.GetClusterKubeconfig(c.Request.Context(), &cluster)
	if err != nil {
		reqLog.Error().Err(err).Msg("Failed to retrieve kubeconfig")
		InternalServerError(c)
		return
	}

	OK(c, "Kubeconfig retrieved successfully", gin.H{"kubeconfig": kubeconfig})
}

// @Summary Deploy cluster
// @Description Creates and deploys a new Kubernetes cluster
// @Tags deployments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param cluster body ClusterInput true "Cluster configuration"
// @Success 202 {object} APIResponse{data=DeploymentWorkflowResponse} "Deployment workflow started successfully"
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments [post]
func (h *DeploymentHandler) HandleDeployCluster(c *gin.Context) {
	reqLog := requestLogger(c, "HandleDeployCluster")

	userID := c.GetInt("user_id")
	config, err := h.svc.GetClientConfig(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get client config")
		InternalServerError(c)
		return
	}

	var cluster kubedeployer.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		BadRequest(c, "Invalid request json format")
		return
	}

	if err := cluster.Validate(); err != nil {
		BadRequest(c, "Validation failed: "+err.Error())
		return
	}

	projectName := kubedeployer.GetProjectName(config.UserID, cluster.Name)
	logWithProject := reqLog.With().Str("project_name", projectName).Logger()
	reqLog = &logWithProject

	// check if deployment already exists
	_, err = h.svc.GetClusterByName(config.UserID, projectName)
	if err == nil {
		Conflict(c, "Deployment already exists")
		return
	} else if !errors.Is(err, models.ErrClusterNotFound) {
		reqLog.Error().Err(err).Msg("Database error when checking for existing deployment")
		InternalServerError(c)
		return
	}

	wfUUID, wfStatus, err := h.svc.AsyncDeployCluster(config, cluster)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to start deployment workflow")
		InternalServerError(c)
		return
	}

	Accepted(c, "Deployment workflow started successfully", DeploymentWorkflowResponse{WorkflowID: wfUUID, Status: string(wfStatus)})
}

// @Summary Delete deployment
// @Description Deletes a specific deployment and all its resources
// @Tags deployments
// @Security BearerAuth
// @Produce json
// @Param name path string true "Deployment name"
// @Success 202 {object} APIResponse{data=DeploymentWorkflowResponse} "Deployment deletion workflow started successfully"
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 404 {object} APIResponse "Deployment not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments/{name} [delete]
func (h *DeploymentHandler) HandleDeleteCluster(c *gin.Context) {
	reqLog := requestLogger(c, "HandleDeleteCluster")

	userID := c.GetInt("user_id")
	config, err := h.svc.GetClientConfig(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get client config")
		InternalServerError(c)
		return
	}

	deploymentName := c.Param("name")
	if deploymentName == "" {
		BadRequest(c, "Deployment name is required")
		return
	}

	projectName := kubedeployer.GetProjectName(config.UserID, deploymentName)
	logWithProject := reqLog.With().Str("project_name", projectName).Logger()
	reqLog = &logWithProject

	_, err = h.svc.GetClusterByName(config.UserID, projectName)
	if err != nil {
		if errors.Is(err, models.ErrClusterNotFound) {
			NotFound(c, "Deployment not found")
		} else {
			reqLog.Error().Err(err).Msg("Database error when looking up deployment for deletion")
			InternalServerError(c)
		}
		return
	}

	wfUUID, wfStatus, err := h.svc.AsyncDeleteCluster(config, projectName)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to start deployment deletion workflow")
		InternalServerError(c)
		return
	}

	Accepted(c, "Deployment deletion workflow started successfully", DeploymentWorkflowResponse{WorkflowID: wfUUID, Status: string(wfStatus)})
}

// @Summary Delete all deployments
// @Description Deletes all deployments and their resources for the authenticated user
// @Tags deployments
// @Security BearerAuth
// @Produce json
// @Success 202 {object} APIResponse{data=DeploymentWorkflowResponse} "Delete all deployments workflow started successfully"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments [delete]
func (h *DeploymentHandler) HandleDeleteAllDeployments(c *gin.Context) {
	reqLog := requestLogger(c, "HandleDeleteAllDeployments")

	userID := c.GetInt("user_id")
	config, err := h.svc.GetClientConfig(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get client config")
		InternalServerError(c)
		return
	}

	clusters, err := h.svc.ListUserClusters(config.UserID)
	if err != nil {
		reqLog.Error().Err(err).Msg("Failed to list user clusters for deletion")
		InternalServerError(c)
		return
	}

	if len(clusters) == 0 {
		OK(c, "No deployments found to delete", nil)
		return
	}

	wfUUID, wfStatus, err := h.svc.AsyncDeleteAllClusters(config)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to start delete all deployments workflow")
		InternalServerError(c)
		return
	}

	Accepted(c, "Delete all deployments workflow started successfully", DeploymentWorkflowResponse{WorkflowID: wfUUID, Status: string(wfStatus)})
}

// @Summary Add node to deployment
// @Description Adds a new node to an existing deployment
// @Tags deployments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param cluster body ClusterInput true "Cluster configuration with new node"
// @Success 202 {object} APIResponse{data=DeploymentWorkflowResponse} "Node addition workflow started successfully"
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 404 {object} APIResponse "Deployment not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments/{name}/nodes [post]
func (h *DeploymentHandler) HandleAddNode(c *gin.Context) {
	reqLog := requestLogger(c, "HandleAddNode")

	userID := c.GetInt("user_id")
	config, err := h.svc.GetClientConfig(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get client config")
		InternalServerError(c)
		return
	}

	var cluster kubedeployer.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		BadRequest(c, "Invalid request json format")
		return
	}

	projectName := kubedeployer.GetProjectName(config.UserID, cluster.Name)
	logWithProject := reqLog.With().Str("project_name", projectName).Logger()
	reqLog = &logWithProject
	existingCluster, err := h.svc.GetClusterByName(config.UserID, projectName)
	if err != nil {
		if errors.Is(err, models.ErrClusterNotFound) {
			NotFound(c, "Deployment not found")
			return
		}
		reqLog.Error().Err(err).Msg("Database error when looking up deployment for adding node")
		InternalServerError(c)
		return
	}

	cl, err := existingCluster.GetClusterResult()
	if err != nil {
		reqLog.Error().Err(err).Int("cluster_id", existingCluster.ID).Msg("Failed to deserialize cluster result")
		InternalServerError(c)
		return
	}

	// TODO: find a better place for this
	cluster.Nodes[0].OriginalName = cluster.Nodes[0].Name

	for _, node := range cl.Nodes {
		if node.OriginalName == cluster.Nodes[0].OriginalName {
			Conflict(c, "Node with the same name already exists")
			return
		}

		if node.NodeID == cluster.Nodes[0].NodeID {
			Conflict(c, fmt.Sprintf("node id %d is already assigned to this cluster", node.NodeID))
			return
		}
	}

	wfUUID, wfStatus, err := h.svc.AsyncAddNode(config, cl, cluster.Nodes[0])
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to start add node workflow")
		InternalServerError(c)
		return
	}

	Accepted(c, "Node addition workflow started successfully", DeploymentWorkflowResponse{WorkflowID: wfUUID, Status: string(wfStatus)})
}

// @Summary Remove node from deployment
// @Description Removes a specific node from an existing deployment
// @Tags deployments
// @Security BearerAuth
// @Produce json
// @Param name path string true "Deployment name"
// @Param node_name path string true "Node name to remove"
// @Success 202 {object} APIResponse{data=DeploymentWorkflowResponse} "Node removal workflow started successfully"
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 404 {object} APIResponse "Deployment not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments/{name}/nodes/{node_name} [delete]
func (h *DeploymentHandler) HandleRemoveNode(c *gin.Context) {
	reqLog := requestLogger(c, "HandleRemoveNode")

	userID := c.GetInt("user_id")
	config, err := h.svc.GetClientConfig(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get client config")
		InternalServerError(c)
		return
	}

	deploymentName := c.Param("name")
	nodeName := c.Param("node_name")

	if deploymentName == "" {
		BadRequest(c, "Deployment name is required")
		return
	}

	if nodeName == "" {
		BadRequest(c, "Node name is required")
		return
	}

	projectName := kubedeployer.GetProjectName(config.UserID, deploymentName)
	logWithFields := reqLog.With().
		Str("project_name", projectName).
		Str("deployment_name", deploymentName).
		Logger()
	reqLog = &logWithFields

	cluster, err := h.svc.GetClusterByName(config.UserID, projectName)
	if err != nil {
		if errors.Is(err, models.ErrClusterNotFound) {
			reqLog.Error().Err(err).Msg("Deployment not found")
			NotFound(c, "Deployment not found")
			return
		}
		reqLog.Error().Err(err).Msg("Database error when looking up deployment for node removal")
		InternalServerError(c)
		return
	}

	cl, err := cluster.GetClusterResult()
	if err != nil {
		reqLog.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
		InternalServerError(c)
		return
	}

	nodeExists := false
	for _, node := range cl.Nodes {
		if node.OriginalName == nodeName {
			nodeExists = true
		}
	}

	if !nodeExists {
		NotFound(c, fmt.Sprintf("node %q not found in cluster %q", nodeName, deploymentName))
		return
	}

	wfUUID, wfStatus, err := h.svc.AsyncRemoveNode(config, cl, nodeName)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to start remove node workflow")
		InternalServerError(c)
		return
	}

	Accepted(c, "Node removal workflow started successfully", DeploymentWorkflowResponse{WorkflowID: wfUUID, Status: string(wfStatus)})
}
