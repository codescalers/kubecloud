package app

import (
	"errors"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/activities"
	"kubecloud/internal/statemanager"
	"kubecloud/kubedeployer"
	"net"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/xmonader/ewf"
	"golang.org/x/crypto/ssh"
	"gorm.io/gorm"
	"kubecloud/internal/logger"
)

// Response represents the response structure for deployment requests
type Response struct {
	WorkflowID string `json:"task_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
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
// @Success 200 {object} DeploymentListResponse "Deployments retrieved successfully"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments [get]
func (h *Handler) HandleListDeployments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := fmt.Sprintf("%v", userID)
	clusters, err := h.db.ListUserClusters(id)
	if err != nil {
		logger.GetLogger().Error().Err(err).Str("user_id", id).Msg("Failed to list user clusters")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployments"})
		return
	}

	deployments := make([]gin.H, 0, len(clusters))
	for _, cluster := range clusters {
		clusterResult, err := cluster.GetClusterResult()
		if err != nil {
			logger.GetLogger().Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
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

	c.JSON(http.StatusOK, gin.H{
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
// @Success 200 {object} DeploymentResponse "Deployment details retrieved successfully"
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 404 {object} APIResponse "Deployment not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments/{name} [get]
func (h *Handler) HandleGetDeployment(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectName := c.Param("name")
	if projectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project name is required"})
		return
	}

	id := fmt.Sprintf("%v", userID)
	projectName = kubedeployer.GetProjectName(id, projectName)
	cluster, err := h.db.GetClusterByName(id, projectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.GetLogger().Error().Err(err).Str("user_id", id).Str("project_name", projectName).Msg("Deployment not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		} else {
			logger.GetLogger().Error().Err(err).Str("user_id", id).Str("project_name", projectName).Msg("Database error when looking up deployment")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to lookup deployment"})
		}
		return
	}

	clusterResult, err := cluster.GetClusterResult()
	if err != nil {
		logger.GetLogger().Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployment details"})
		return
	}

	response := gin.H{
		"id":           cluster.ID,
		"project_name": cluster.ProjectName,
		"cluster":      clusterResult,
		"created_at":   cluster.CreatedAt,
		"updated_at":   cluster.UpdatedAt,
	}

	c.JSON(http.StatusOK, response)
}

// @Summary Get kubeconfig
// @Description Retrieves the kubeconfig file for a specific deployment
// @Tags deployments
// @Security BearerAuth
// @Produce json
// @Param name path string true "Deployment name"
// @Success 200 {object} KubeconfigResponse "Kubeconfig retrieved successfully"
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 404 {object} APIResponse "Deployment not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments/{name}/kubeconfig [get]
func (h *Handler) HandleGetKubeconfig(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	projectName := c.Param("name")
	if projectName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "project name is required"})
		return
	}

	id := fmt.Sprintf("%v", userID)
	projectName = kubedeployer.GetProjectName(id, projectName)
	cluster, err := h.db.GetClusterByName(id, projectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.GetLogger().Error().Err(err).Str("user_id", id).Str("project_name", projectName).Msg("Deployment not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		} else {
			logger.GetLogger().Error().Err(err).Str("user_id", id).Str("project_name", projectName).Msg("Database error when looking up deployment for kubeconfig")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to lookup deployment"})
		}
		return
	}

	clusterResult, err := cluster.GetClusterResult()
	if err != nil {
		logger.GetLogger().Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployment details"})
		return
	}

	var targetNode *kubedeployer.Node
	for _, node := range clusterResult.Nodes {
		if node.Type == kubedeployer.NodeTypeLeader {
			targetNode = &node
			break
		}
	}

	if targetNode == nil {
		for _, node := range clusterResult.Nodes {
			if node.Type == kubedeployer.NodeTypeMaster {
				targetNode = &node
				break
			}
		}
	}

	if targetNode == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "No leader or master node found in deployment"})
		return
	}

	privateKeyBytes, err := os.ReadFile(h.config.SSH.PrivateKeyPath)
	if err != nil {
		logger.GetLogger().Error().Err(err).Str("key_path", h.config.SSH.PrivateKeyPath).Msg("Failed to read SSH private key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read SSH configuration"})
		return
	}

	kubeconfig, err := h.getKubeconfigViaSSH(string(privateKeyBytes), targetNode)
	if err != nil {
		logger.GetLogger().Error().Err(err).Str("node_name", targetNode.Name).Msg("Failed to retrieve kubeconfig via SSH")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve kubeconfig: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"kubeconfig": kubeconfig})
}

func (h *Handler) getKubeconfigViaSSH(privateKey string, node *kubedeployer.Node) (string, error) {
	ip := node.MyceliumIP
	if ip == "" {
		return "", fmt.Errorf("no valid IP address found for node %s", node.Name)
	}

	logger.GetLogger().Debug().Str("ip", ip).Str("node", node.Name).Msg("Attempting SSH connection")
	commands := []string{
		"kubectl config view --minify --raw",
		"cat /etc/rancher/k3s/k3s.yaml",
		"cat ~/.kube/config",
	}

	for _, cmd := range commands {
		kubeconfig, err := h.executeSSHCommand(privateKey, ip, cmd)
		if err == nil && strings.Contains(kubeconfig, "apiVersion") && strings.Contains(kubeconfig, "clusters") {
			processedKubeconfig, processErr := h.processKubeconfig(kubeconfig, ip)
			if processErr != nil {
				logger.GetLogger().Warn().Err(processErr).Str("ip", ip).Msg("Failed to process kubeconfig, returning original")
				return kubeconfig, nil
			}
			return processedKubeconfig, nil
		}
		if err != nil {
			logger.GetLogger().Debug().Err(err).Str("ip", ip).Str("command", cmd).Msg("Command failed, trying next")
		}
	}

	return "", fmt.Errorf("failed to retrieve kubeconfig from node %s at IP %s", node.Name, ip)
}

func (h *Handler) executeSSHCommand(privateKey, address, command string) (string, error) {
	key, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return "", fmt.Errorf("could not parse SSH private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            "root",
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		Timeout: 30 * time.Second,
	}

	port := "22"
	var client *ssh.Client
	for attempt := 1; attempt <= 3; attempt++ {
		client, err = ssh.Dial("tcp", net.JoinHostPort(address, port), config)
		if err == nil {
			break
		}
		if attempt < 3 {
			logger.GetLogger().Debug().Err(err).Str("address", address).Int("attempt", attempt).Msg("SSH connection attempt failed, retrying")
			time.Sleep(time.Duration(attempt) * time.Second)
		}
	}

	if err != nil {
		return "", fmt.Errorf("could not establish SSH connection to %s after 3 attempts: %w", address, err)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("could not create SSH session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("could not execute command '%s': %w, output: %s", command, err, string(output))
	}

	return string(output), nil
}

func (h *Handler) processKubeconfig(kubeconfigYAML, externalIP string) (string, error) {
	updatedConfig := kubeconfigYAML
	oldPattern := "server: https://127.0.0.1:"
	var newPattern string

	if strings.Contains(externalIP, ":") {
		newPattern = fmt.Sprintf("server: https://[%s]:", externalIP)
	} else {
		newPattern = fmt.Sprintf("server: https://%s:", externalIP)
	}

	updatedConfig = strings.ReplaceAll(updatedConfig, oldPattern, newPattern)

	logger.GetLogger().Debug().
		Str("target_ip", externalIP).
		Bool("config_changed", updatedConfig != kubeconfigYAML).
		Msg("Processed kubeconfig for external IP")

	return updatedConfig, nil
}

func (h *Handler) getClientConfig(c *gin.Context) (statemanager.ClientConfig, error) {
	userID, exists := c.Get("user_id")
	if !exists {
		return statemanager.ClientConfig{}, fmt.Errorf("user_id not found in context")
	}
	userIDStr := fmt.Sprintf("%v", userID)

	userIDInt, err := strconv.Atoi(userIDStr)
	if err != nil {
		return statemanager.ClientConfig{}, fmt.Errorf("failed to parse user ID: %v", err)
	}

	user, err := h.db.GetUserByID(userIDInt)
	if err != nil {
		return statemanager.ClientConfig{}, fmt.Errorf("failed to get user: %v", err)
	}

	return statemanager.ClientConfig{
		SSHPublicKey: h.sshPublicKey,
		Mnemonic:     user.Mnemonic,
		UserID:       userIDStr,
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
// @Success 202 {object} Response "Deployment workflow started successfully"
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments [post]
func (h *Handler) HandleDeployCluster(c *gin.Context) {
	config, err := h.getClientConfig(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var cluster kubedeployer.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request json format"})
		return
	}

	if err := internal.ValidateStruct(cluster); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	if err := cluster.Validate(); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	projectName := kubedeployer.GetProjectName(config.UserID, cluster.Name)
	_, err = h.db.GetClusterByName(config.UserID, projectName)
	if err == nil {
		c.JSON(http.StatusConflict, gin.H{"error": "deployment already exists"})
		return
	} else if !errors.Is(err, gorm.ErrRecordNotFound) {
		logger.GetLogger().Error().Err(err).Str("user_id", config.UserID).Str("project_name", projectName).Msg("Database error when checking for existing deployment")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to check existing deployments"})
		return
	}

	wfName := fmt.Sprintf("deploy-%d-nodes", len(cluster.Nodes))
	activities.NewDynamicDeployWorkflowTemplate(h.ewfEngine, h.metrics, wfName, len(cluster.Nodes), h.sseManager)

	// Get the workflow
	wf, err := h.ewfEngine.NewWorkflow(wfName)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow"})
		return
	}

	wf.State = ewf.State{
		"config":  config,
		"cluster": cluster,
	}

	h.ewfEngine.RunAsync(c, wf)

	c.JSON(http.StatusAccepted, Response{
		WorkflowID: wf.UUID,
		Status:     string(wf.Status),
		Message:    "Deployment workflow started successfully",
	})
}

// @Summary Delete deployment
// @Description Deletes a specific deployment and all its resources
// @Tags deployments
// @Security BearerAuth
// @Produce json
// @Param name path string true "Deployment name"
// @Success 200 {object} Response "Deployment deletion workflow started successfully"
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 404 {object} APIResponse "Deployment not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments/{name} [delete]
func (h *Handler) HandleDeleteCluster(c *gin.Context) {
	config, err := h.getClientConfig(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	deploymentName := c.Param("name")
	if deploymentName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deployment name is required"})
		return
	}
	projectName := kubedeployer.GetProjectName(config.UserID, deploymentName)
	_, err = h.db.GetClusterByName(config.UserID, projectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		} else {
			logger.GetLogger().Error().Err(err).Str("user_id", config.UserID).Str("project_name", projectName).Msg("Database error when looking up deployment for deletion")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to lookup deployment"})
		}
		return
	}

	wf, err := h.ewfEngine.NewWorkflow(activities.WorkflowDeleteCluster)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow"})
		return
	}

	wf.State = ewf.State{
		"config":       config,
		"project_name": projectName,
	}

	h.ewfEngine.RunAsync(c, wf)

	c.JSON(http.StatusOK, Response{
		WorkflowID: wf.UUID,
		Status:     string(wf.Status),
		Message:    "Deployment deletion workflow started successfully",
	})
}

// @Summary Delete all deployments
// @Description Deletes all deployments and their resources for the authenticated user
// @Tags deployments
// @Security BearerAuth
// @Produce json
// @Success 200 {object} Response "Delete all deployments workflow started successfully"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments [delete]
func (h *Handler) HandleDeleteAllDeployments(c *gin.Context) {
	config, err := h.getClientConfig(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	clusters, err := h.db.ListUserClusters(config.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployments"})
		return
	}

	if len(clusters) == 0 {
		c.JSON(http.StatusOK, gin.H{"message": "No deployments found to delete"})
		return
	}

	wf, err := h.ewfEngine.NewWorkflow(activities.WorkflowDeleteAllClusters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow"})
		return
	}

	wf.State = ewf.State{
		"config": config,
	}

	h.ewfEngine.RunAsync(c, wf)

	c.JSON(http.StatusAccepted, Response{
		WorkflowID: wf.UUID,
		Status:     string(wf.Status),
		Message:    "Delete all deployments workflow started successfully",
	})
}

// @Summary Add node to deployment
// @Description Adds a new node to an existing deployment
// @Tags deployments
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param cluster body ClusterInput true "Cluster configuration with new node"
// @Success 202 {object} Response "Node addition workflow started successfully"
// @Failure 400 {object} APIResponse "Invalid request format"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 404 {object} APIResponse "Deployment not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments/{name}/nodes [post]
func (h *Handler) HandleAddNode(c *gin.Context) {
	config, err := h.getClientConfig(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var cluster kubedeployer.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request json format"})
		return
	}

	if err := internal.ValidateStruct(cluster); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	projectName := kubedeployer.GetProjectName(config.UserID, cluster.Name)
	existingCluster, err := h.db.GetClusterByName(config.UserID, projectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		} else {
			logger.GetLogger().Error().Err(err).Str("user_id", config.UserID).Str("project_name", projectName).Msg("Database error when looking up deployment for adding node")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to lookup deployment"})
		}
		return
	}

	cl, err := existingCluster.GetClusterResult()
	if err != nil {
		logger.GetLogger().Error().Err(err).Int("cluster_id", existingCluster.ID).Msg("Failed to deserialize cluster result")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployment details"})
		return
	}

	// TODO: find a better place for this
	cluster.Nodes[0].OriginalName = cluster.Nodes[0].Name

	for _, node := range cl.Nodes {
		if node.OriginalName == cluster.Nodes[0].OriginalName {
			c.JSON(http.StatusConflict, gin.H{"error": "Node with the same name already exists"})
			return
		}
	}

	wf, err := h.ewfEngine.NewWorkflow(activities.WorkflowAddNode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow"})
		return
	}

	wf.State = ewf.State{
		"config":  config,
		"cluster": cl,
		"node":    cluster.Nodes[0],
	}

	h.ewfEngine.RunAsync(c, wf)

	c.JSON(http.StatusAccepted, Response{
		WorkflowID: wf.UUID,
		Status:     string(wf.Status),
		Message:    "Node addition workflow started successfully",
	})
}

// @Summary Remove node from deployment
// @Description Removes a specific node from an existing deployment
// @Tags deployments
// @Security BearerAuth
// @Produce json
// @Param name path string true "Deployment name"
// @Param node_name path string true "Node name to remove"
// @Success 200 {object} Response "Node removal workflow started successfully"
// @Failure 400 {object} APIResponse "Invalid request"
// @Failure 401 {object} APIResponse "Unauthorized"
// @Failure 404 {object} APIResponse "Deployment not found"
// @Failure 500 {object} APIResponse "Internal server error"
// @Router /deployments/{name}/nodes/{node_name} [delete]
func (h *Handler) HandleRemoveNode(c *gin.Context) {
	config, err := h.getClientConfig(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	deploymentName := c.Param("name")
	nodeName := c.Param("node_name")

	if deploymentName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "deployment name is required"})
		return
	}

	if nodeName == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "node name is required"})
		return
	}

	projectName := kubedeployer.GetProjectName(config.UserID, deploymentName)
	cluster, err := h.db.GetClusterByName(config.UserID, projectName)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			logger.GetLogger().Error().Err(err).Str("user_id", config.UserID).Str("deployment_name", deploymentName).Msg("Deployment not found")
			c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		} else {
			logger.GetLogger().Error().Err(err).Str("user_id", config.UserID).Str("deployment_name", deploymentName).Msg("Database error when looking up deployment for node removal")
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to lookup deployment"})
		}
		return
	}

	cl, err := cluster.GetClusterResult()
	if err != nil {
		logger.GetLogger().Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployment details"})
		return
	}

	nodeExists := false
	for _, node := range cl.Nodes {
		if node.OriginalName == nodeName {
			nodeExists = true
		}
	}

	if !nodeExists {
		c.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("node %q not found in cluster %q", nodeName, deploymentName)})
		return
	}

	wf, err := h.ewfEngine.NewWorkflow(activities.WorkflowRemoveNode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow"})
		return
	}

	wf.State = ewf.State{
		"config":    config,
		"cluster":   cl,
		"node_name": nodeName,
	}

	h.ewfEngine.RunAsync(c, wf)

	c.JSON(http.StatusOK, Response{
		WorkflowID: wf.UUID,
		Status:     string(wf.Status),
		Message:    "Node removal workflow started successfully",
	})
}
