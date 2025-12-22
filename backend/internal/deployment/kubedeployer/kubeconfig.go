package kubedeployer

import (
	"context"
	"fmt"
	"kubecloud/internal/infrastructure/logger"
	"kubecloud/internal/infrastructure/telemetry"
	"net"
	"strings"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
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
	ctx, span := getTracer().Start(ctx, "Cluster.GetKubeconfig",
		trace.WithAttributes(
			attribute.String("cluster.name", c.Name),
			attribute.Int("cluster.node_count", len(c.Nodes)),
		),
	)
	defer span.End()

	log := logger.ForOperation("kubedeployer", "get_kubeconfig")

	if privateKey == "" {
		err := fmt.Errorf("private key cannot be empty")
		telemetry.RecordError(span, err)
		return "", err
	}

	if len(c.Nodes) == 0 {
		err := fmt.Errorf("cluster %s has no nodes", c.Name)
		telemetry.RecordError(span, err)
		return "", err
	}

	for _, node := range c.Nodes {
		if node.Type == NodeTypeLeader || node.Type == NodeTypeMaster {
			span.AddEvent("Attempting to retrieve kubeconfig from node",
				trace.WithAttributes(
					attribute.String("node.name", node.Name),
					attribute.String("node.type", string(node.Type)),
					attribute.String("node.mycelium_ip", node.MyceliumIP),
				),
			)
			log.Debug().
				Str("cluster", c.Name).
				Str("node_name", node.Name).
				Str("node_type", string(node.Type)).
				Str("node_ip", node.MyceliumIP).
				Msg("Attempting to retrieve kubeconfig from node")

			kubeconfig, err := GetKubeconfigViaSSH(ctx, privateKey, &node)
			if err != nil {
				log.Debug().
					Err(err).
					Str("cluster", c.Name).
					Str("node_name", node.Name).
					Str("node_type", string(node.Type)).
					Msg("Failed to retrieve kubeconfig from node")
				continue // try the next node
			}

			span.SetAttributes(
				attribute.String("node.successful", node.Name),
				attribute.String("node.successful_type", string(node.Type)),
			)
			span.AddEvent("Successfully retrieved kubeconfig")
			log.Info().
				Str("cluster", c.Name).
				Str("node_name", node.Name).
				Msg("Successfully retrieved kubeconfig")
			return kubeconfig, nil
		}
	}

	err := fmt.Errorf("no leader or master node found in cluster %s (checked %d nodes)", c.Name, len(c.Nodes))
	telemetry.RecordError(span, err)
	return "", err
}

func GetKubeconfigViaSSH(ctx context.Context, privateKey string, node *Node) (string, error) {
	ctx, span := getTracer().Start(ctx, "getKubeconfigViaSSH",
		trace.WithAttributes(
			attribute.String("node.name", node.Name),
			attribute.String("node.mycelium_ip", node.MyceliumIP),
		),
	)
	defer span.End()

	log := logger.ForOperation("kubedeployer", "ssh_kubeconfig_retrieval")

	ip := node.MyceliumIP
	if ip == "" {
		err := fmt.Errorf("no mycelium IP address found for node %s (node_id: %d)", node.Name, node.NodeID)
		telemetry.RecordError(span, err)
		return "", err
	}

	span.AddEvent("Attempting SSH connection to retrieve kubeconfig")
	log.Debug().
		Str("ip", ip).
		Str("node", node.Name).
		Msg("Attempting SSH connection to retrieve kubeconfig")

	command := fmt.Sprintf("cat %s", k3sKubeconfigPath)
	span.SetAttributes(attribute.String("ssh.command", command))

	kubeconfig, err := executeSSHCommand(ctx, privateKey, ip, command)
	if err != nil {
		log.Debug().
			Err(err).
			Str("ip", ip).
			Str("node", node.Name).
			Msg("Failed to execute SSH command")
		telemetry.RecordError(span, err)
		return "", fmt.Errorf("failed to execute SSH command on node %s (node_id: %d, ip: %s): %w", node.Name, node.NodeID, ip, err)
	}

	if !isValidKubeconfig(kubeconfig) {
		err := fmt.Errorf("invalid kubeconfig content retrieved from node %s (node_id: %d, ip: %s): missing required fields", node.Name, node.NodeID, ip)
		telemetry.RecordError(span, err)
		return "", err
	}

	processedKubeconfig, processErr := processKubeconfig(ctx, kubeconfig, ip)
	if processErr != nil {
		log.Warn().
			Err(processErr).
			Str("ip", ip).
			Str("node", node.Name).
			Msg("Failed to process kubeconfig, returning original")
		span.AddEvent("Failed to process kubeconfig, returning original")
		return kubeconfig, nil
	}

	span.AddEvent("Kubeconfig retrieved and processed successfully")
	return processedKubeconfig, nil
}

func executeSSHCommand(ctx context.Context, privateKey, address, command string) (string, error) {
	ctx, span := getTracer().Start(ctx, "executeSSHCommand",
		trace.WithAttributes(
			attribute.String("ssh.address", address),
			attribute.String("ssh.command", command),
			attribute.Int("ssh.max_retries", sshMaxRetries),
		),
	)
	defer span.End()

	log := logger.ForOperation("kubedeployer", "execute_ssh_command")

	key, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		telemetry.RecordError(span, err)
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
		span.AddEvent("SSH connection attempt",
			trace.WithAttributes(
				attribute.Int("attempt", attempt),
				attribute.Int("max_retries", sshMaxRetries),
			),
		)
		client, err = ssh.Dial("tcp", net.JoinHostPort(address, sshPort), config)
		if err == nil {
			span.AddEvent("SSH connection established",
				trace.WithAttributes(
					attribute.Int("attempt", attempt),
					attribute.Int("max_retries", sshMaxRetries),
				),
			)
			break
		}

		lastErr = err
		if attempt < sshMaxRetries {
			backoffDuration := sshRetryBackoff * time.Duration(attempt)

			span.AddEvent("SSH connection attempt failed, retrying",
				trace.WithAttributes(
					attribute.Int("attempt", attempt),
					attribute.Int("max_retries", sshMaxRetries),
					attribute.String("backoff", backoffDuration.String()),
				),
			)

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
				span.AddEvent("Operation cancelled during retry backoff")
				return "", fmt.Errorf("operation cancelled during retry backoff: %w", ctx.Err())
			}
		}
	}

	if client == nil {
		err := fmt.Errorf("failed to establish SSH connection to %s after %d attempts: %w", address, sshMaxRetries, lastErr)
		telemetry.RecordError(span, err)
		span.SetAttributes(attribute.Int("ssh.attempts_failed", sshMaxRetries))
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		telemetry.RecordError(span, err)
		return "", fmt.Errorf("failed to create SSH session for %s: %w", address, err)
	}
	defer session.Close()

	span.AddEvent("Executing SSH command", trace.WithAttributes(
		attribute.String("ssh.command", command),
	))

	output, err := session.CombinedOutput(command)
	if err != nil {
		telemetry.RecordError(span, err)
		return "", fmt.Errorf("failed to execute command '%s' on %s: %w (output: %s)", command, address, err, string(output))
	}

	span.AddEvent("SSH command executed successfully")
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
func processKubeconfig(ctx context.Context, kubeconfigYAML, externalIP string) (string, error) {
	log := logger.ForOperation("kubedeployer", "process_kubeconfig")
	_, span := getTracer().Start(ctx, "processKubeconfig")
	defer span.End()

	if kubeconfigYAML == "" {
		span.RecordError(fmt.Errorf("kubeconfig content cannot be empty"))
		return "", fmt.Errorf("kubeconfig content cannot be empty")
	}
	if externalIP == "" {
		span.RecordError(fmt.Errorf("external IP cannot be empty"))
		return "", fmt.Errorf("external IP cannot be empty")
	}

	var newPattern string
	// Handle IPv6 addresses by wrapping them in brackets
	if strings.Contains(externalIP, ":") {
		newPattern = fmt.Sprintf("server: https://[%s]:", externalIP)
	} else {
		newPattern = fmt.Sprintf("server: https://%s:", externalIP)
	}

	kubeconfigServerPattern := "server: https://127.0.0.1:"

	span.AddEvent("Processing kubeconfig for external IP", trace.WithAttributes(
		attribute.String("kubeconfig.external_ip", externalIP),
		attribute.String("kubeconfig.server_pattern", kubeconfigServerPattern),
		attribute.String("kubeconfig.new_pattern", newPattern),
	))

	updatedConfig := strings.ReplaceAll(kubeconfigYAML, kubeconfigServerPattern, newPattern)

	configChanged := updatedConfig != kubeconfigYAML
	if !configChanged {
		span.AddEvent("No server URL replacement made in kubeconfig - pattern may not match")
		log.Warn().
			Str("target_ip", externalIP).
			Msg("No server URL replacement made in kubeconfig - pattern may not match")
	}

	span.AddEvent("Kubeconfig processed successfully", trace.WithAttributes(
		attribute.Bool("config_changed", configChanged),
	))

	log.Debug().
		Str("target_ip", externalIP).
		Bool("config_changed", configChanged).
		Msg("Processed kubeconfig for external IP")

	return updatedConfig, nil
}
