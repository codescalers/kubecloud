package controller

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"

	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/types"
	"github.com/threefoldtech/zosbase/pkg/gridtypes/zos"
	"k8s.io/klog"
)

type GWRequest struct {
	Hostname string   `json:"hostname"`
	Backends []string `json:"backends"`
}

func deployGateway(pluginClient deployer.TFPluginClient, gw GWRequest) (workloads.GatewayNameProxy, error) {
	var zosBackends []zos.Backend
	for _, backend := range gw.Backends {
		zosBackends = append(zosBackends, zos.Backend(backend))
	}

	gateway := workloads.GatewayNameProxy{
		Name:         gw.Hostname,
		Backends:     zosBackends,
		SolutionType: gw.Hostname,
	}

	node, err := selectNode(pluginClient)
	if err != nil {
		return workloads.GatewayNameProxy{}, fmt.Errorf("failed to select node: %w", err)
	}

	gateway.NodeID = node
	if err := pluginClient.GatewayNameDeployer.Deploy(context.TODO(), &gateway); err != nil {
		return workloads.GatewayNameProxy{}, fmt.Errorf("failed to deploy gateway on node %d: %w", node, err)
	}

	res, err := pluginClient.State.LoadGatewayNameFromGrid(context.TODO(), gateway.NodeID, gateway.Name, gateway.Name)
	if err != nil {
		return workloads.GatewayNameProxy{}, fmt.Errorf("failed to load gateway for name %s: %w", gateway.Name, err)
	}

	klog.Infof("Gateway deployed successfully on node %d", node)

	return res, nil
}

func selectNode(pluginClient deployer.TFPluginClient) (uint32, error) {
	trueVal := true
	nodes, err := deployer.FilterNodes(
		context.Background(),
		pluginClient,
		types.NodeFilter{Domain: &trueVal, Status: []string{"up"}},
		nil, nil, nil,
	)
	if err != nil {
		return 0, fmt.Errorf("failed to filter nodes: %w", err)
	}

	if len(nodes) == 0 {
		return 0, fmt.Errorf("no available nodes found")
	}

	return uint32(nodes[0].NodeID), nil
}

func generateSessionId() (string, error) {
	b := make([]byte, 6)
	_, err := rand.Read(b)
	if err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}
	sessionID := fmt.Sprintf("tfgwCRD-%s", base64.URLEncoding.EncodeToString(b))
	return sessionID, nil
}

// decrypt decrypts an encrypted base64 string with a string key
func decrypt(key, encryptedText string) (string, error) {
	hash := sha256.Sum256([]byte(key)) // valid 32 bytes for AES-256
	key = string(hash[:])

	data, err := base64.StdEncoding.DecodeString(encryptedText)
	if err != nil {
		return "", err
	}

	block, err := aes.NewCipher([]byte(key))
	if err != nil {
		return "", err
	}

	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}

	nonceSize := aesGCM.NonceSize()
	nonce, ciphertext := data[:nonceSize], data[nonceSize:]

	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		return "", err
	}

	return string(plaintext), nil
}
