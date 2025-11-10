package app

import (
	"errors"
	"fmt"
	"kubecloud/internal/constants"
	"kubecloud/internal/statemanager"
	"kubecloud/kubedeployer"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/xmonader/ewf"
	"gorm.io/gorm"
)

// Response represents the response structure for deployment requests
type DeploymentWorkflowResponse struct {
	WorkflowID string `json:"task_id"`
	Status     string `json:"status"`
}

// DeploymentResponse represents the response for deployment operations
type DeploymentResponse struct {
	ID          int         `json:"id"`
	ProjectName string      `json:"project_name"`
	Cluster     interface{} `json:"cluster"`
	CreatedAt   string      `json:"created_at"`
	UpdatedAt   string      `json:"updated_at"`
}

// DeploymentListResponse represents the response for listing deployments
type DeploymentListResponse struct {
	Deployments []DeploymentResponse `json:"deployments"`
	Count       int                  `json:"count"`
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
	DiskSize   uint64            `json:"disk_size"`                    // Storage in MB
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
func (h *Handler) HandleListDeployments(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "HandleListDeployments")
	if userID == 0 {
		Unauthorized(c, "user not authenticated")
		return
	}

	clusters, err := h.db.ListUserClusters(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to list user clusters")
		InternalServerError(c)
		return
	}

	deployments := make([]gin.H, 0, len(clusters))
	for _, cluster := range clusters {
		clusterResult, err := cluster.GetClusterResult()
		if err != nil {
			reqLog.Error().Err(err).Int("cluster_id", cluster.ID).Msg("failed to deserialize cluster result")
			continue
		}

		deployments = append(deployments, gin.H{
			"id":           cluster.ID,
			"project_name": cluster.ProjectName,
			"cluster":      clusterResult,
			"created_at":   cluster.CreatedAt,
			"updated_at":   cluster.UpdatedAt,
		})
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
// @Success 200 {object} APIResponse{data=DeploymentResponse} "Deployment details retrieved successfully"
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 404 {object} APIResponse "Deployment not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments/{name} [get]
func (h *Handler) HandleGetDeployment(c *gin.Context) {
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
	cluster, err := h.db.GetClusterByName(userID, projectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			reqLog.Error().Err(err).Msg("Deployment not found")
			NotFound(c, "Deployment not found")
		} else {
			reqLog.Error().Err(err).Msg("Database error when looking up deployment")
			InternalServerError(c)
		}
		return
	}

	clusterResult, err := cluster.GetClusterResult()
	if err != nil {
		reqLog.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
		InternalServerError(c)
		return
	}

	response := gin.H{
		"id":           cluster.ID,
		"project_name": cluster.ProjectName,
		"cluster":      clusterResult,
		"created_at":   cluster.CreatedAt,
		"updated_at":   cluster.UpdatedAt,
	}

	OK(c, "Deployment details retrieved successfully", response)
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
func (h *Handler) HandleGetKubeconfig(c *gin.Context) {
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
	cluster, err := h.db.GetClusterByName(userID, projectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			reqLog.Error().Err(err).Msg("Deployment not found")
			NotFound(c, "Deployment not found")
		} else {
			reqLog.Error().Err(err).Msg("Database error when looking up deployment for kubeconfig")
			InternalServerError(c)
		}
		return
	}

	if cluster.Kubeconfig != "" {
		OK(c, "Kubeconfig retrieved successfully", gin.H{"kubeconfig": cluster.Kubeconfig})
		return
	}

	clusterResult, err := cluster.GetClusterResult()
	if err != nil {
		reqLog.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
		InternalServerError(c)
		return
	}

	privateKeyBytes, err := os.ReadFile(h.config.SSH.PrivateKeyPath)
	if err != nil {
		reqLog.Error().Err(err).Str("key_path", h.config.SSH.PrivateKeyPath).Msg("Failed to read SSH private key")
		InternalServerError(c)
		return
	}

	kubeconfig, err := clusterResult.GetKubeconfig(c.Request.Context(), string(privateKeyBytes))
	if err != nil {
		reqLog.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to retrieve kubeconfig via SSH")
		InternalServerError(c)
		return
	}

	cluster.Kubeconfig = kubeconfig
	if err := h.db.UpdateCluster(&cluster); err != nil {
		reqLog.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to save kubeconfig to database")
	}

	OK(c, "Kubeconfig retrieved successfully", gin.H{"kubeconfig": kubeconfig})
}

func (h *Handler) getClientConfig(c *gin.Context) (statemanager.ClientConfig, error) {
	userID := c.GetInt("user_id")

	user, err := h.db.GetUserByID(userID)
	if err != nil {
		return statemanager.ClientConfig{}, fmt.Errorf("failed to get user: %v", err)
	}

	return statemanager.ClientConfig{
		SSHPublicKey: h.sshPublicKey,
		Mnemonic:     user.Mnemonic,
		UserID:       userID,
		Network:      h.config.SystemAccount.Network,
		Debug:        h.config.Debug,
	}, nil
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
func (h *Handler) HandleDeployCluster(c *gin.Context) {
	config, err := h.getClientConfig(c)
	reqLog := requestLogger(c, "HandleDeployCluster")
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
	_, err = h.db.GetClusterByName(config.UserID, projectName)
	if err == nil {
		Conflict(c, "Deployment already exists")
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		reqLog.Error().Err(err).Msg("Database error when checking for existing deployment")
		InternalServerError(c)
		return
	}

	wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowDeployCluster)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to create workflow for cluster deployment")
		InternalServerError(c)
		return
	}

	wf.State = ewf.State{
		"config":  config,
		"cluster": cluster,
	}

	h.ewfEngine.RunAsync(h.appContext, wf)
	Accepted(c, "Deployment workflow started successfully", DeploymentWorkflowResponse{WorkflowID: wf.UUID, Status: string(wf.Status)})
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
func (h *Handler) HandleDeleteCluster(c *gin.Context) {
	config, err := h.getClientConfig(c)
	reqLog := requestLogger(c, "HandleDeleteCluster")
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
	_, err = h.db.GetClusterByName(config.UserID, projectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NotFound(c, "Deployment not found")
		} else {
			reqLog.Error().Err(err).Msg("Database error when looking up deployment for deletion")
			InternalServerError(c)
		}
		return
	}

	wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowDeleteCluster)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to create workflow for cluster deletion")
		InternalServerError(c)
		return
	}

	wf.State = ewf.State{
		"config":       config,
		"project_name": projectName,
	}

	h.ewfEngine.RunAsync(h.appContext, wf)

	Accepted(c, "Deployment deletion workflow started successfully", DeploymentWorkflowResponse{WorkflowID: wf.UUID, Status: string(wf.Status)})
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
func (h *Handler) HandleDeleteAllDeployments(c *gin.Context) {
	config, err := h.getClientConfig(c)
	reqLog := requestLogger(c, "HandleDeleteAllDeployments")
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to get client config")
		InternalServerError(c)
		return
	}

	clusters, err := h.db.ListUserClusters(config.UserID)
	if err != nil {
		reqLog.Error().Err(err).Msg("Failed to list user clusters for deletion")
		InternalServerError(c)
		return
	}

	if len(clusters) == 0 {
		OK(c, "No deployments found to delete", nil)
		return
	}

	wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowDeleteAllClusters)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to create workflow for deleting all deployments")
		InternalServerError(c)
		return
	}

	wf.State = ewf.State{
		"config": config,
	}

	h.ewfEngine.RunAsync(h.appContext, wf)

	Accepted(c, "Delete all deployments workflow started successfully", DeploymentWorkflowResponse{WorkflowID: wf.UUID, Status: string(wf.Status)})
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
func (h *Handler) HandleAddNode(c *gin.Context) {
	config, err := h.getClientConfig(c)
	reqLog := requestLogger(c, "HandleAddNode")
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
	existingCluster, err := h.db.GetClusterByName(config.UserID, projectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			NotFound(c, "Deployment not found")
		} else {
			reqLog.Error().Err(err).Msg("Database error when looking up deployment for adding node")
			InternalServerError(c)
		}
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

	wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowAddNode)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to create workflow for adding node")
		InternalServerError(c)
		return
	}

	wf.State = ewf.State{
		"config":  config,
		"cluster": cl,
		"node":    cluster.Nodes[0],
	}

	h.ewfEngine.RunAsync(h.appContext, wf)
	Accepted(c, "Node addition workflow started successfully", DeploymentWorkflowResponse{WorkflowID: wf.UUID, Status: string(wf.Status)})
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
func (h *Handler) HandleRemoveNode(c *gin.Context) {
	config, err := h.getClientConfig(c)
	reqLog := requestLogger(c, "HandleRemoveNode")
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
	cluster, err := h.db.GetClusterByName(config.UserID, projectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			reqLog.Error().Err(err).Msg("Deployment not found")
			NotFound(c, "Deployment not found")
		} else {
			reqLog.Error().Err(err).Msg("Database error when looking up deployment for node removal")
			InternalServerError(c)
		}
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

	wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowRemoveNode)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to create workflow for removing node")
		InternalServerError(c)
		return
	}

	wf.State = ewf.State{
		"config":    config,
		"cluster":   cl,
		"node_name": nodeName,
	}

	h.ewfEngine.RunAsync(h.appContext, wf)

	Accepted(c, "Node removal workflow started successfully", DeploymentWorkflowResponse{WorkflowID: wf.UUID, Status: string(wf.Status)})
}
