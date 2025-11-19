package handlers

import (
	"errors"
	"fmt"

	"github.com/gin-gonic/gin"

	"kubecloud/internal/core/models"
	"kubecloud/internal/core/services"
	"kubecloud/internal/deployment/kubedeployer"
	"kubecloud/internal/infrastructure/logger"
)

type DeploymentHandler struct {
	svc services.DeploymentService
}

func NewDeploymentHandler(svc services.DeploymentService) DeploymentHandler {
	return DeploymentHandler{
		svc: svc,
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
func (h *DeploymentHandler) HandleListDeployments(c *gin.Context) {
	userID := c.GetInt("user_id")
	reqLog := requestLogger(c, "HandleListDeployments")
	if userID == 0 {
		auditLogFromContext(c, logger.AuditActionDeploymentList, logger.AuditSeverityWarning, map[string]any{
			"reason": "user_not_authenticated",
		})
		Unauthorized(c, "user not authenticated")
		return
	}

	deployments, err := h.svc.ListUserClustersData(userID)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to list user clusters")
		auditLogFromContext(c, logger.AuditActionDeploymentList, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	auditLogFromContext(c, logger.AuditActionDeploymentList, logger.AuditSeverityInfo, map[string]any{
		"count": len(deployments),
	})
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
		auditLogFromContext(c, logger.AuditActionDeploymentGet, logger.AuditSeverityWarning, map[string]any{
			"reason": "user_not_authenticated",
		})
		Unauthorized(c, "unauthorized")
		return
	}

	projectName := c.Param("name")
	if projectName == "" {
		auditLogFromContext(c, logger.AuditActionDeploymentGet, logger.AuditSeverityWarning, map[string]any{
			"reason": "project_name_required",
		})
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
			auditLogFromContext(c, logger.AuditActionDeploymentGet, logger.AuditSeverityWarning, map[string]any{
				"reason": "deployment_not_found",
			})
			NotFound(c, "Deployment not found")
			return
		}

		reqLog.Error().Err(err).Msg("Database error when looking up deployment")
		auditLogFromContext(c, logger.AuditActionDeploymentGet, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	auditLogFromContext(c, logger.AuditActionDeploymentGet, logger.AuditSeverityInfo, map[string]any{
		"project_name": projectName,
	})
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
		auditLogFromContext(c, logger.AuditActionDeploymentKubeconfig, logger.AuditSeverityWarning, map[string]any{
			"reason": "user_not_authenticated",
		})
		Unauthorized(c, "User not authenticated")
		return
	}

	projectName := c.Param("name")
	if projectName == "" {
		auditLogFromContext(c, logger.AuditActionDeploymentKubeconfig, logger.AuditSeverityWarning, map[string]any{
			"reason": "project_name_required",
		})
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
			auditLogFromContext(c, logger.AuditActionDeploymentKubeconfig, logger.AuditSeverityWarning, map[string]any{
				"reason": "deployment_not_found",
			})
			NotFound(c, "Deployment not found")
			return
		}
		reqLog.Error().Err(err).Msg("Database error when looking up deployment for kubeconfig")
		auditLogFromContext(c, logger.AuditActionDeploymentKubeconfig, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	kubeconfig, err := h.svc.GetClusterKubeconfig(c.Request.Context(), &cluster)
	if err != nil {
		reqLog.Error().Err(err).Msg("Failed to retrieve kubeconfig")
		auditLogFromContext(c, logger.AuditActionDeploymentKubeconfig, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	auditLogFromContext(c, logger.AuditActionDeploymentKubeconfig, logger.AuditSeverityInfo, map[string]any{
		"project_name": projectName,
	})
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
		auditLogFromContext(c, logger.AuditActionDeploymentDeploy, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	var cluster kubedeployer.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		auditLogFromContext(c, logger.AuditActionDeploymentDeploy, logger.AuditSeverityWarning, map[string]any{
			"reason": "invalid_request_format",
		})
		BadRequest(c, "Invalid request json format")
		return
	}

	if err := cluster.Validate(); err != nil {
		auditLogFromContext(c, logger.AuditActionDeploymentDeploy, logger.AuditSeverityWarning, map[string]any{
			"reason": "validation_failed",
		})
		BadRequest(c, "Validation failed: "+err.Error())
		return
	}

	projectName := kubedeployer.GetProjectName(config.UserID, cluster.Name)
	logWithProject := reqLog.With().Str("project_name", projectName).Logger()
	reqLog = &logWithProject

	// check if deployment already exists
	_, err = h.svc.GetClusterByName(config.UserID, projectName)
	if err == nil {
		auditLogFromContext(c, logger.AuditActionDeploymentDeploy, logger.AuditSeverityWarning, map[string]any{
			"reason": "deployment_exists",
		})
		Conflict(c, "Deployment already exists")
		return
	} else if !errors.Is(err, models.ErrClusterNotFound) {
		reqLog.Error().Err(err).Msg("Database error when checking for existing deployment")
		auditLogFromContext(c, logger.AuditActionDeploymentDeploy, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	wfUUID, wfStatus, err := h.svc.AsyncDeployCluster(config, cluster)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to start deployment workflow")
		auditLogFromContext(c, logger.AuditActionDeploymentDeploy, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	auditLogFromContext(c, logger.AuditActionDeploymentDeploy, logger.AuditSeverityInfo, map[string]any{
		"workflow_id":  wfUUID,
		"status":       string(wfStatus),
		"project_name": projectName,
	})
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
		auditLogFromContext(c, logger.AuditActionDeploymentDelete, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	deploymentName := c.Param("name")
	if deploymentName == "" {
		auditLogFromContext(c, logger.AuditActionDeploymentDelete, logger.AuditSeverityWarning, map[string]any{
			"reason": "deployment_name_required",
		})
		BadRequest(c, "Deployment name is required")
		return
	}

	projectName := kubedeployer.GetProjectName(config.UserID, deploymentName)
	logWithProject := reqLog.With().Str("project_name", projectName).Logger()
	reqLog = &logWithProject

	_, err = h.svc.GetClusterByName(config.UserID, projectName)
	if err != nil {
		if errors.Is(err, models.ErrClusterNotFound) {
			auditLogFromContext(c, logger.AuditActionDeploymentDelete, logger.AuditSeverityWarning, map[string]any{
				"reason": "deployment_not_found",
			})
			NotFound(c, "Deployment not found")
		} else {
			reqLog.Error().Err(err).Msg("Database error when looking up deployment for deletion")
			auditLogFromContext(c, logger.AuditActionDeploymentDelete, logger.AuditSeverityError, map[string]any{
				"reason": err.Error(),
			})
			InternalServerError(c)
		}
		return
	}

	wfUUID, wfStatus, err := h.svc.AsyncDeleteCluster(config, projectName)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to start deployment deletion workflow")
		auditLogFromContext(c, logger.AuditActionDeploymentDelete, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	auditLogFromContext(c, logger.AuditActionDeploymentDelete, logger.AuditSeverityInfo, map[string]any{
		"workflow_id":  wfUUID,
		"status":       string(wfStatus),
		"project_name": projectName,
	})
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
		auditLogFromContext(c, logger.AuditActionDeploymentDeleteAll, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	clusters, err := h.svc.ListUserClusters(config.UserID)
	if err != nil {
		reqLog.Error().Err(err).Msg("Failed to list user clusters for deletion")
		auditLogFromContext(c, logger.AuditActionDeploymentDeleteAll, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	if len(clusters) == 0 {
		auditLogFromContext(c, logger.AuditActionDeploymentDeleteAll, logger.AuditSeverityInfo, map[string]any{
			"result": "no_deployments",
		})
		OK(c, "No deployments found to delete", nil)
		return
	}

	wfUUID, wfStatus, err := h.svc.AsyncDeleteAllClusters(config)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to start delete all deployments workflow")
		auditLogFromContext(c, logger.AuditActionDeploymentDeleteAll, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	auditLogFromContext(c, logger.AuditActionDeploymentDeleteAll, logger.AuditSeverityInfo, map[string]any{
		"workflow_id": wfUUID,
		"status":      string(wfStatus),
		"count":       len(clusters),
	})
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
		auditLogFromContext(c, logger.AuditActionDeploymentAddNode, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	var cluster kubedeployer.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		auditLogFromContext(c, logger.AuditActionDeploymentAddNode, logger.AuditSeverityWarning, map[string]any{
			"reason": "invalid_request_format",
		})
		BadRequest(c, "Invalid request json format")
		return
	}

	projectName := kubedeployer.GetProjectName(config.UserID, cluster.Name)
	logWithProject := reqLog.With().Str("project_name", projectName).Logger()
	reqLog = &logWithProject
	existingCluster, err := h.svc.GetClusterByName(config.UserID, projectName)
	if err != nil {
		if errors.Is(err, models.ErrClusterNotFound) {
			auditLogFromContext(c, logger.AuditActionDeploymentAddNode, logger.AuditSeverityWarning, map[string]any{
				"reason": "deployment_not_found",
			})
			NotFound(c, "Deployment not found")
			return
		}
		reqLog.Error().Err(err).Msg("Database error when looking up deployment for adding node")
		auditLogFromContext(c, logger.AuditActionDeploymentAddNode, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	cl, err := existingCluster.GetClusterResult()
	if err != nil {
		reqLog.Error().Err(err).Int("cluster_id", existingCluster.ID).Msg("Failed to deserialize cluster result")
		auditLogFromContext(c, logger.AuditActionDeploymentAddNode, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	// TODO: find a better place for this
	cluster.Nodes[0].OriginalName = cluster.Nodes[0].Name

	for _, node := range cl.Nodes {
		if node.OriginalName == cluster.Nodes[0].OriginalName {
			auditLogFromContext(c, logger.AuditActionDeploymentAddNode, logger.AuditSeverityWarning, map[string]any{
				"reason": "node_name_exists",
				"name":   cluster.Nodes[0].OriginalName,
			})
			Conflict(c, "Node with the same name already exists")
			return
		}

		if node.NodeID == cluster.Nodes[0].NodeID {
			auditLogFromContext(c, logger.AuditActionDeploymentAddNode, logger.AuditSeverityWarning, map[string]any{
				"reason":  "node_id_exists",
				"node_id": node.NodeID,
			})
			Conflict(c, fmt.Sprintf("node id %d is already assigned to this cluster", node.NodeID))
			return
		}
	}

	wfUUID, wfStatus, err := h.svc.AsyncAddNode(config, cl, cluster.Nodes[0])
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to start add node workflow")
		auditLogFromContext(c, logger.AuditActionDeploymentAddNode, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	auditLogFromContext(c, logger.AuditActionDeploymentAddNode, logger.AuditSeverityInfo, map[string]any{
		"workflow_id":  wfUUID,
		"status":       string(wfStatus),
		"project_name": projectName,
		"node_name":    cluster.Nodes[0].OriginalName,
	})
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
		auditLogFromContext(c, logger.AuditActionDeploymentRemoveNode, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	deploymentName := c.Param("name")
	nodeName := c.Param("node_name")

	if deploymentName == "" {
		auditLogFromContext(c, logger.AuditActionDeploymentRemoveNode, logger.AuditSeverityWarning, map[string]any{
			"reason": "deployment_name_required",
		})
		BadRequest(c, "Deployment name is required")
		return
	}

	if nodeName == "" {
		auditLogFromContext(c, logger.AuditActionDeploymentRemoveNode, logger.AuditSeverityWarning, map[string]any{
			"reason": "node_name_required",
		})
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
			auditLogFromContext(c, logger.AuditActionDeploymentRemoveNode, logger.AuditSeverityWarning, map[string]any{
				"reason": "deployment_not_found",
			})
			NotFound(c, "Deployment not found")
			return
		}
		reqLog.Error().Err(err).Msg("Database error when looking up deployment for node removal")
		auditLogFromContext(c, logger.AuditActionDeploymentRemoveNode, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	cl, err := cluster.GetClusterResult()
	if err != nil {
		reqLog.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
		auditLogFromContext(c, logger.AuditActionDeploymentRemoveNode, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
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
		auditLogFromContext(c, logger.AuditActionDeploymentRemoveNode, logger.AuditSeverityWarning, map[string]any{
			"reason": "node_not_found",
			"node":   nodeName,
		})
		NotFound(c, fmt.Sprintf("node %q not found in cluster %q", nodeName, deploymentName))
		return
	}

	wfUUID, wfStatus, err := h.svc.AsyncRemoveNode(config, cl, nodeName)
	if err != nil {
		reqLog.Error().Err(err).Msg("failed to start remove node workflow")
		auditLogFromContext(c, logger.AuditActionDeploymentRemoveNode, logger.AuditSeverityError, map[string]any{
			"reason": err.Error(),
		})
		InternalServerError(c)
		return
	}

	auditLogFromContext(c, logger.AuditActionDeploymentRemoveNode, logger.AuditSeverityInfo, map[string]any{
		"workflow_id":  wfUUID,
		"status":       string(wfStatus),
		"project_name": projectName,
		"node_name":    nodeName,
	})
	Accepted(c, "Node removal workflow started successfully", DeploymentWorkflowResponse{WorkflowID: wfUUID, Status: string(wfStatus)})
}
