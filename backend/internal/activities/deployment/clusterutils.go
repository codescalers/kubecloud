package deployment

import (
	"context"
	"errors"
	"fmt"
	"kubecloud/internal"
	"kubecloud/internal/logger"
	"kubecloud/internal/statemanager"
	"kubecloud/kubedeployer"
	"kubecloud/models"
	"os"

	"github.com/xmonader/ewf"
	"gorm.io/gorm"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

func retrieveKubeconfig(state ewf.State, db models.DB, privateKeyPath string) (string, error) {
	if kc, ok := state["kubeconfig"].(string); ok && kc != "" {
		return kc, nil
	}

	config, err := GetFromState[statemanager.ClientConfig](state, "config")
	if err != nil {
		return "", fmt.Errorf("failed to get config from state: %w", err)
	}

	cluster, err := GetFromState[kubedeployer.Cluster](state, "cluster")
	if err != nil {
		return "", fmt.Errorf("failed to get cluster from state: %w", err)
	}

	// when updating existing cluster
	existingCluster, err := db.GetClusterByName(config.UserID, cluster.Name)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("failed to query cluster from database: %w", err)
	}

	if existingCluster.ID != 0 && existingCluster.Kubeconfig != "" {
		logger.GetLogger().Debug().Msgf("Using kubeconfig from DB for cluster %s", existingCluster.ProjectName)
		return existingCluster.Kubeconfig, nil
	}

	var master kubedeployer.Node
	if existingCluster.ID != 0 {
		existingClusterResult, err := existingCluster.GetClusterResult()
		if err != nil {
			return "", fmt.Errorf("failed to get cluster result: %w", err)
		}
		master, err = existingClusterResult.GetLeaderNode()
		if err != nil {
			return "", fmt.Errorf("failed to get leader node: %w", err)
		}
	} else {
		master, err = cluster.GetLeaderNode()
		if err != nil {
			return "", fmt.Errorf("failed to get leader node: %w", err)
		}
	}

	privateKeyBytes, err := os.ReadFile(privateKeyPath)
	if err != nil {
		return "", fmt.Errorf("failed to read SSH private key: %w", err)
	}

	logger.GetLogger().Debug().Msg("Fetching kubeconfig from leader node via SSH")
	return internal.GetKubeconfigViaSSH(string(privateKeyBytes), &master)
}

// createKubernetesClient creates a Kubernetes clientset from kubeconfig
func createKubernetesClient(kubeconfig string) (*kubernetes.Clientset, error) {
	restConfig, err := clientcmd.RESTConfigFromKubeConfig([]byte(kubeconfig))
	if err != nil {
		return nil, fmt.Errorf("failed to parse kubeconfig: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	return clientset, nil
}

// isNodeReady checks if a Kubernetes node is in Ready state
func isNodeReady(node v1.Node) bool {
	for _, cond := range node.Status.Conditions {
		if cond.Type == v1.NodeReady && cond.Status == v1.ConditionTrue {
			return true
		}
	}
	return false
}

// verifyAllNodesReady verifies that all nodes in the cluster are ready
func verifyAllNodesReady(ctx context.Context, clientset *kubernetes.Clientset) error {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to list nodes: %w", err)
	}

	for _, n := range nodes.Items {
		if !isNodeReady(n) {
			return fmt.Errorf("node %s is not ready", n.Name)
		}
	}

	return nil
}

// verifySpecificNodeReady verifies that a specific node is ready
func verifySpecificNodeReady(ctx context.Context, clientset *kubernetes.Clientset, nodeName string) error {
	n, err := clientset.CoreV1().Nodes().Get(ctx, nodeName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node %s from cluster: %w", nodeName, err)
	}

	if !isNodeReady(*n) {
		return fmt.Errorf("node %s is not ready", nodeName)
	}

	return nil
}
