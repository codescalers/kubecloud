package app

import (
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
	"github.com/rs/zerolog/log"
	"github.com/xmonader/ewf"
	"golang.org/x/crypto/ssh"
)

// Response represents the response structure for deployment requests
type Response struct {
	WorkflowID string `json:"task_id"`
	Status     string `json:"status"`
	Message    string `json:"message"`
}

func (h *Handler) HandleListDeployments(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := fmt.Sprintf("%v", userID)
	clusters, err := h.db.ListUserClusters(id)
	if err != nil {
		log.Error().Err(err).Str("user_id", id).Msg("Failed to list user clusters")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployments"})
		return
	}

	deployments := make([]gin.H, 0, len(clusters))
	for _, cluster := range clusters {
		clusterResult, err := cluster.GetClusterResult()
		if err != nil {
			log.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
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
		log.Error().Err(err).Str("user_id", id).Str("project_name", projectName).Msg("Failed to get cluster")
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	clusterResult, err := cluster.GetClusterResult()
	if err != nil {
		log.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
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
		log.Error().Err(err).Str("user_id", id).Str("project_name", projectName).Msg("Failed to get cluster")
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	clusterResult, err := cluster.GetClusterResult()
	if err != nil {
		log.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
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
		log.Error().Err(err).Str("key_path", h.config.SSH.PrivateKeyPath).Msg("Failed to read SSH private key")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to read SSH configuration"})
		return
	}

	kubeconfig, err := h.getKubeconfigViaSSH(string(privateKeyBytes), targetNode)
	if err != nil {
		log.Error().Err(err).Str("node_name", targetNode.Name).Msg("Failed to retrieve kubeconfig via SSH")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve kubeconfig: " + err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"kubeconfig": kubeconfig})
}

func (h *Handler) getKubeconfigViaSSH(privateKey string, node *kubedeployer.Node) (string, error) {
	ips := []string{}
	if node.MyceliumIP != "" {
		ips = append(ips, node.MyceliumIP)
	}
	if node.PlanetaryIP != "" {
		ips = append(ips, node.PlanetaryIP)
	}

	if len(ips) == 0 {
		return "", fmt.Errorf("no valid IP addresses found for node %s", node.Name)
	}

	var lastErr error
	for _, ip := range ips {
		log.Debug().Str("ip", ip).Str("node", node.Name).Msg("Attempting SSH connection")

		commands := []string{
			"kubectl config view --minify --raw",
			"cat /etc/rancher/k3s/k3s.yaml",
			"cat ~/.kube/config",
		}

		var kubeconfig string
		var cmdErr error

		for _, cmd := range commands {
			kubeconfig, cmdErr = h.executeSSHCommand(privateKey, ip, cmd)
			if cmdErr == nil && strings.Contains(kubeconfig, "apiVersion") && strings.Contains(kubeconfig, "clusters") {
				processedKubeconfig, processErr := h.processKubeconfig(kubeconfig, ip)
				if processErr != nil {
					log.Warn().Err(processErr).Str("ip", ip).Msg("Failed to process kubeconfig, returning original")
					return kubeconfig, nil
				}
				return processedKubeconfig, nil
			}
			if cmdErr != nil {
				log.Debug().Err(cmdErr).Str("ip", ip).Str("command", cmd).Msg("Command failed, trying next")
			}
		}

		if cmdErr != nil {
			log.Warn().Err(cmdErr).Str("ip", ip).Str("node", node.Name).Msg("All commands failed on this IP, trying next IP")
			lastErr = cmdErr
		} else {
			lastErr = fmt.Errorf("no valid kubeconfig found on node %s at IP %s", node.Name, ip)
		}
	}

	return "", fmt.Errorf("failed to retrieve kubeconfig from any IP address: %v", lastErr)
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
			log.Debug().Err(err).Str("address", address).Int("attempt", attempt).Msg("SSH connection attempt failed, retrying")
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

	var targetIP string
	if strings.Contains(externalIP, ":") {
		parts := strings.Split(externalIP, ":")
		if len(parts) >= 4 {
			targetIP = strings.Join(parts[:4], ":") + "::1"
		} else {
			targetIP = externalIP
		}

		log.Debug().
			Str("original_ipv6", externalIP).
			Str("modified_ipv6", targetIP).
			Msg("Modified IPv6 address for kubeconfig")
	} else {
		targetIP = externalIP
	}

	oldPattern := "server: https://127.0.0.1:"
	var newPattern string

	if strings.Contains(targetIP, ":") {
		newPattern = fmt.Sprintf("server: https://[%s]:", targetIP)
	} else {
		newPattern = fmt.Sprintf("server: https://%s:", targetIP)
	}

	updatedConfig = strings.ReplaceAll(updatedConfig, oldPattern, newPattern)

	log.Debug().
		Str("target_ip", targetIP).
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
	}, nil
}

func (h *Handler) HandleDeployCluster(c *gin.Context) {
	config, err := h.getClientConfig(c)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// TODO: validate the cluster required fields/ pingable nodes
	var cluster kubedeployer.Cluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request json format"})
		return
	}

	if err := internal.ValidateStruct(cluster); err != nil {
		Error(c, http.StatusBadRequest, "Validation failed", err.Error())
		return
	}

	wfName := fmt.Sprintf("deploy_%d_nodes_%s", len(cluster.Nodes), config.UserID)
	activities.NewDynamicDeployWorkflowTemplate(h.ewfEngine, wfName, len(cluster.Nodes))

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
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
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
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	cl, err := existingCluster.GetClusterResult()
	if err != nil {
		log.Error().Err(err).Int("cluster_id", existingCluster.ID).Msg("Failed to deserialize cluster result")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployment details"})
		return
	}

	wf, err := h.ewfEngine.NewWorkflow(activities.WorkflowAddNode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create workflow"})
		return
	}

	// TODO: find a better place for this
	cluster.Nodes[0].OriginalName = cluster.Nodes[0].Name

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
		log.Error().Err(err).Str("user_id", config.UserID).Str("deployment_name", deploymentName).Msg("Failed to find deployment")
		c.JSON(http.StatusNotFound, gin.H{"error": "deployment not found"})
		return
	}

	cl, err := cluster.GetClusterResult()
	if err != nil {
		log.Error().Err(err).Int("cluster_id", cluster.ID).Msg("Failed to deserialize cluster result")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to retrieve deployment details"})
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
