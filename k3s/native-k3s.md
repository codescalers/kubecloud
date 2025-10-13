# Native K3s Setup Guide

This guide explains how to run K3s natively on Ubuntu VMs on the ThreeFold Grid, with support for cluster formation, dual-stack networking, and custom CRD installation.

## Prerequisites

- **Two Ubuntu VMs** on the ThreeFold Grid connected via WireGuard through the grid network
    - 2GB RAM
    - 20GB disk mounted at `$K3S_DATA_DIR`
- Clone the directories to both VMs
    - `scripts/` => `~/scripts/`
    - `manifests/` => `$K3S_DATA_DIR/server/manifests/` those manifests will be applied when k3s is starting

## Setup Steps

### 1. Install Dependencies

Run the prepare script to install all the required dependencies:

```bash
./prepare
```

This script will install:
- K3s binary
- Required system packages
- Network utilities
- Other dependencies needed for the cluster

### 2. Configure Environment Variables

Modify the `.env` file according to your setup. Reference `.env-example` for documentation on each configuration field. and export

```bash
set -a && source .env && set +a
```

### 3. Create Flannel Interface

Run the Flannel interface creation script:

```bash
./create_flannel_iface.sh
```

**⚠️ Important:** This script should only be run once and is not reversible.

### 4. Update CRD Manifests

Run the CRD manifest update script, only needed on the k3s server node.

```bash
./update-crd-manifest.sh
```

This will modify the manifest files to match your environment configuration.

### 5. Start K3s

Run the entrypoint script to start K3s:

```bash
./entrypoint
```

The script will automatically start K3s as either a server or agent based on your environment variables.

## Cluster Formation

- **Server Node:** On the first VM (master), K3s will start as a server and create the cluster
- **Agent Node:** On the second VM (worker), K3s will join the existing cluster as an agent

## Verification

After starting K3s, verify the cluster status:

```bash
# On the server node
kubectl get nodes

# Check cluster info
kubectl cluster-info
```

## Troubleshooting

- Ensure WireGuard connectivity between VMs before starting K3s
- Check K3s logs: `journalctl -u k3s` (if running as a service)
- Verify the data directory has sufficient space
- Confirm all environment variables are correctly set
