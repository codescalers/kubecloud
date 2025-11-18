package kubedeployer

import (
	"context"
	"fmt"
	"kubecloud/internal/infrastructure/logger"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

const (
	sshPort           = "22"
	sshUser           = "root"
	sshTimeout        = 30 * time.Second
	sshRetryBackoff   = 2 * time.Second
	sshMaxRetries     = 3
	k3sKubeconfigPath = "/etc/rancher/k3s/k3s.yaml"
)

// GetKubeconfig retrieves the kubeconfig from a leader or master node in the cluster.
func (c *Cluster) GetKubeconfig(ctx context.Context, privateKey string) (string, error) {
	log := logger.ForOperation("kubedeployer", "get_kubeconfig")

	if privateKey == "" {
		return "", fmt.Errorf("private key cannot be empty")
	}

	if len(c.Nodes) == 0 {
		return "", fmt.Errorf("cluster %s has no nodes", c.Name)
	}

	for _, node := range c.Nodes {
		if node.Type == NodeTypeLeader || node.Type == NodeTypeMaster {
			log.Debug().
				Str("cluster", c.Name).
				Str("node_name", node.Name).
				Str("node_type", string(node.Type)).
				Str("node_ip", node.MyceliumIP).
				Msg("Attempting to retrieve kubeconfig from node")

			kubeconfig, err := getKubeconfigViaSSH(ctx, privateKey, &node)
			if err != nil {
				log.Debug().
					Err(err).
					Str("cluster", c.Name).
					Str("node_name", node.Name).
					Str("node_type", string(node.Type)).
					Msg("Failed to retrieve kubeconfig from node")
				continue // try the next node
			}

			log.Info().
				Str("cluster", c.Name).
				Str("node_name", node.Name).
				Msg("Successfully retrieved kubeconfig")
			return kubeconfig, nil
		}
	}

	return "", fmt.Errorf("no leader or master node found in cluster %s (checked %d nodes)", c.Name, len(c.Nodes))
}

func getKubeconfigViaSSH(ctx context.Context, privateKey string, node *Node) (string, error) {
	log := logger.ForOperation("kubedeployer", "ssh_kubeconfig_retrieval")

	ip := node.MyceliumIP
	if ip == "" {
		return "", fmt.Errorf("no mycelium IP address found for node %s", node.Name)
	}

	log.Debug().
		Str("ip", ip).
		Str("node", node.Name).
		Msg("Attempting SSH connection to retrieve kubeconfig")

	command := fmt.Sprintf("cat %s", k3sKubeconfigPath)
	kubeconfig, err := executeSSHCommand(ctx, privateKey, ip, command)
	if err != nil {
		log.Debug().
			Err(err).
			Str("ip", ip).
			Str("node", node.Name).
			Msg("Failed to execute SSH command")
		return "", fmt.Errorf("failed to execute SSH command on node %s (%s): %w", node.Name, ip, err)
	}

	if !isValidKubeconfig(kubeconfig) {
		return "", fmt.Errorf("invalid kubeconfig content retrieved from node %s (%s): missing required fields", node.Name, ip)
	}

	processedKubeconfig, processErr := processKubeconfig(kubeconfig, ip)
	if processErr != nil {
		log.Warn().
			Err(processErr).
			Str("ip", ip).
			Str("node", node.Name).
			Msg("Failed to process kubeconfig, returning original")
		return kubeconfig, nil
	}

	return processedKubeconfig, nil
}

func executeSSHCommand(ctx context.Context, privateKey, address, command string) (string, error) {
	log := logger.ForOperation("kubedeployer", "execute_ssh_command")

	key, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return "", fmt.Errorf("failed to parse SSH private key: %w", err)
	}

	config := &ssh.ClientConfig{
		User:            sshUser,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(key),
		},
		Timeout: sshTimeout,
	}

	var client *ssh.Client
	var lastErr error

	for attempt := 1; attempt <= sshMaxRetries; attempt++ {
		client, err = ssh.Dial("tcp", net.JoinHostPort(address, sshPort), config)
		if err == nil {
			break
		}

		lastErr = err
		if attempt < sshMaxRetries {
			backoffDuration := sshRetryBackoff * time.Duration(attempt)
			log.Debug().
				Err(err).
				Str("address", address).
				Int("attempt", attempt).
				Int("max_retries", sshMaxRetries).
				Dur("backoff", backoffDuration).
				Msg("SSH connection attempt failed, retrying")

			select {
			case <-time.After(backoffDuration):
				continue
			case <-ctx.Done():
				return "", fmt.Errorf("operation cancelled during retry backoff: %w", ctx.Err())
			}
		}
	}

	if client == nil {
		return "", fmt.Errorf("failed to establish SSH connection to %s after %d attempts: %w", address, sshMaxRetries, lastErr)
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create SSH session for %s: %w", address, err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(command)
	if err != nil {
		return "", fmt.Errorf("failed to execute command '%s' on %s: %w (output: %s)", command, address, err, string(output))
	}

	return string(output), nil
}

func isValidKubeconfig(kubeconfig string) bool {
	kubeconfigAPIVersionKey := "apiVersion"
	kubeconfigClustersKey := "clusters"
	if kubeconfig == "" {
		return false
	}
	return strings.Contains(kubeconfig, kubeconfigAPIVersionKey) &&
		strings.Contains(kubeconfig, kubeconfigClustersKey)
}

// processKubeconfig replaces localhost references in the kubeconfig with the external IP.
func processKubeconfig(kubeconfigYAML, externalIP string) (string, error) {
	log := logger.ForOperation("kubedeployer", "process_kubeconfig")

	kubeconfigServerPattern := "server: https://127.0.0.1:"

	if kubeconfigYAML == "" {
		return "", fmt.Errorf("kubeconfig content cannot be empty")
	}
	if externalIP == "" {
		return "", fmt.Errorf("external IP cannot be empty")
	}

	var newPattern string
	// Handle IPv6 addresses by wrapping them in brackets
	if strings.Contains(externalIP, ":") {
		newPattern = fmt.Sprintf("server: https://[%s]:", externalIP)
	} else {
		newPattern = fmt.Sprintf("server: https://%s:", externalIP)
	}

	updatedConfig := strings.ReplaceAll(kubeconfigYAML, kubeconfigServerPattern, newPattern)

	configChanged := updatedConfig != kubeconfigYAML
	if !configChanged {
		log.Warn().
			Str("target_ip", externalIP).
			Msg("No server URL replacement made in kubeconfig - pattern may not match")
	}

	log.Debug().
		Str("target_ip", externalIP).
		Bool("config_changed", configChanged).
		Msg("Processed kubeconfig for external IP")

	return updatedConfig, nil
}
