# Native K3s Setup Guide

This guide explains how to run K3s natively on Ubuntu (without Docker) with support for cluster formation, dual-stack networking, and custom CRD installation.

## Overview

This setup allows you to:
- Run K3s server/agent directly on Ubuntu machines
- Support dual-stack (IPv4/IPv6) networking with Mycelium
- Create high-availability (HA) multi-master clusters
- Deploy custom CRDs with encrypted credentials

## Prerequisites

- **OS**: Ubuntu 20.04 or later
- **Privileges**: Root access (sudo)
- **Network**: Internet connectivity for downloading binaries
- **Storage**: At least 2GB free space for K3s data directory
- **Dual-Stack**: Two network interfaces (eth0 for IPv4, eth1+ for IPv6 with 400::/7 route)

## Quick Start

### 1. Clone and Navigate

```bash
cd /path/to/kubecloud/k3s
```

### 2. Configure Environment

Create a `.env` file in the k3s directory:

```bash
# .env

# Enable dual-stack networking (IPv4 + IPv6)
DUAL_STACK=true

# Secret token for cluster authentication (must be same on all nodes)
K3S_TOKEN="your-secure-token-here"

# Network interface for Flannel CNI (flannel-br for dual-stack, eth0 for IPv4-only)
K3S_FLANNEL_IFACE=flannel-br

# Data directory for K3s (ensure sufficient storage)
K3S_DATA_DIR=/var/lib/rancher/k3s

# Unique node name per cluster node
K3S_NODE_NAME=node1

# Node type: true for server/master, false for agent/worker
MASTER=true

# High availability: true only for the first master node
HA=true

# Cluster URL: empty for first master, https://<first-master-ip>:6443 for others
K3S_URL=

# Kubeconfig location (auto-generated)
KUBECONFIG=/etc/rancher/k3s/k3s.yaml

# Optional: Encrypted mnemonic for CRD deployment
MNEMONIC=""

# Optional: Network identifier for CRD
NETWORK=""
```

### 3. Run the Setup

```bash
sudo ./scripts/start-k3s.sh
```

## Configuration Examples

### First Master Node (HA Cluster)

```bash
# .env
DUAL_STACK=true
K3S_TOKEN="my-cluster-secret"
K3S_FLANNEL_IFACE=flannel-br
K3S_DATA_DIR=/var/lib/rancher/k3s
K3S_NODE_NAME=master1
MASTER=true
HA=true
K3S_URL=
KUBECONFIG=/etc/rancher/k3s/k3s.yaml
```

### Additional Master Nodes

```bash
# .env
DUAL_STACK=true
K3S_TOKEN="my-cluster-secret"
K3S_FLANNEL_IFACE=flannel-br
K3S_DATA_DIR=/var/lib/rancher/k3s
K3S_NODE_NAME=master2
MASTER=true
HA=false
K3S_URL=https://10.0.0.10:6443  # IP of first master
KUBECONFIG=/etc/rancher/k3s/k3s.yaml
```

### Worker/Agent Nodes

```bash
# .env
DUAL_STACK=true
K3S_TOKEN="my-cluster-secret"
K3S_FLANNEL_IFACE=flannel-br
K3S_DATA_DIR=/var/lib/rancher/k3s
K3S_NODE_NAME=worker1
MASTER=false
HA=false
K3S_URL=https://10.0.0.10:6443  # IP of first master
KUBECONFIG=/etc/rancher/k3s/k3s.yaml
```

### IPv4-Only Setup

```bash
# .env
DUAL_STACK=false
K3S_TOKEN="my-cluster-secret"
K3S_FLANNEL_IFACE=eth0
K3S_DATA_DIR=/var/lib/rancher/k3s
K3S_NODE_NAME=node1
MASTER=true
HA=false
K3S_URL=
KUBECONFIG=/etc/rancher/k3s/k3s.yaml
```

## Scripts

All scripts are located in the `scripts/` directory:

- **`prepare.sh`**: Installs system dependencies and K3s binaries
- **`start-k3s.sh`**: Main orchestration script (sources .env, runs all steps)
- **`create_flannel_iface.sh`**: Creates bridge interface for dual-stack networking
- **`entrypoint.sh`**: Launches K3s server or agent with proper configuration
- **`install_crd.sh`**: Applies custom CRDs after cluster is ready

## Manual Step-by-Step

If you prefer to run each step manually:

```bash
# 1. Install dependencies and binaries
sudo ./scripts/prepare.sh

# 2. Source environment variables
source .env

# 3. Create flannel bridge (only if DUAL_STACK=true)
if [ "$DUAL_STACK" = "true" ]; then
    sudo ./scripts/create_flannel_iface.sh
fi

# 4. Start K3s
sudo ./scripts/entrypoint.sh &

# 5. Wait for cluster to be ready, then install CRDs
sleep 30
sudo ./scripts/install_crd.sh
```

## Verification

After starting K3s, verify the installation:

```bash
# Check K3s service status
sudo systemctl status k3s || ps aux | grep k3s

# List nodes
kubectl get nodes

# Check pods
kubectl get pods -A

# Verify dual-stack (if enabled)
kubectl get nodes -o jsonpath='{.items[*].status.addresses[?(@.type=="InternalIP")].address}'
```

## Troubleshooting

### K3s fails to start

- Check logs: `journalctl -u k3s -f` or check process output
- Verify network interfaces: `ip addr show`
- Ensure K3S_TOKEN matches across all nodes

### Dual-stack networking issues

- Verify eth0 has IPv4 address: `ip -4 addr show eth0`
- Verify eth1+ has 400::/7 route: `ip -6 route | grep 400::/7`
- Check bridge creation: `ip link show flannel-br`

### Nodes not joining cluster

- Verify K3S_URL points to the correct master IP and port (6443)
- Check firewall rules allow traffic on port 6443
- Ensure K3S_TOKEN is identical on all nodes

### CRD installation fails

- Check if MNEMONIC and NETWORK are set (if required by your CRD)
- Verify manifest exists: `ls -la /var/lib/rancher/k3s/server/manifests/install-crd.yaml`
- Check API server is ready: `kubectl get --raw='/readyz'`

## Cleanup

To completely remove K3s:

```bash
# Stop K3s
sudo pkill k3s

# Remove data directory
sudo rm -rf /var/lib/rancher/k3s

# Remove configuration
sudo rm -rf /etc/rancher/k3s

# Remove bridge interface (if created)
sudo ip link delete flannel-br
```

## Notes

- The `K3S_TOKEN` must be the same across all nodes in the cluster
- For production, use a strong random token (e.g., `openssl rand -base64 32`)
- The first master node should have `HA=true` and empty `K3S_URL`
- All subsequent nodes should have `HA=false` and `K3S_URL` pointing to first master
- Ensure time synchronization across nodes (NTP is installed by prepare.sh)
- Custom CRD manifests should be placed in `/var/lib/rancher/k3s/server/manifests/`