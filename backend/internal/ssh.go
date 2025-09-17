package internal

import (
	"fmt"
	"kubecloud/internal/logger"
	"kubecloud/kubedeployer"
	"net"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"
)

func GetKubeconfigViaSSH(privateKey string, node *kubedeployer.Node) (string, error) {
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
		kubeconfig, err := executeSSHCommand(privateKey, ip, cmd)
		if err == nil && strings.Contains(kubeconfig, "apiVersion") && strings.Contains(kubeconfig, "clusters") {
			processedKubeconfig, processErr := processKubeconfig(kubeconfig, ip)
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

func executeSSHCommand(privateKey, address, command string) (string, error) {
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

func processKubeconfig(kubeconfigYAML, externalIP string) (string, error) {
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
