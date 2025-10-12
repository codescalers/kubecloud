#!/bin/bash
set -euo pipefail

echo "🔧 Preparing system for K3s installation..."

# Update package list
echo "[1/4] Updating package list..."
apt-get update -qq

# Install required dependencies
echo "[2/4] Installing dependencies (wget, iproute2, ntp)..."
apt-get install -y wget iproute2 ntp

# Download K3s binary
echo "[3/4] Downloading K3s v1.33.1+k3s1..."
wget -q --show-progress -O /usr/local/bin/k3s \
    https://github.com/k3s-io/k3s/releases/download/v1.33.1+k3s1/k3s
chmod +x /usr/local/bin/k3s

# Download kubectl binary
echo "[3/4] Downloading kubectl v1.33.1..."
wget -q --show-progress -O /usr/local/bin/kubectl \
    https://dl.k8s.io/release/v1.33.1/bin/linux/amd64/kubectl
chmod +x /usr/local/bin/kubectl

# Create K3s configuration directory
echo "[4/4] Creating K3s configuration directory..."
mkdir -p /etc/rancher/k3s

# Set containerd snapshotter to native (more efficient)
echo 'snapshotter: "native"' > /etc/rancher/k3s/config.yaml

echo "✅ System preparation complete!"
echo "   - K3s binary: /usr/local/bin/k3s"
echo "   - kubectl binary: /usr/local/bin/kubectl"
echo "   - Config directory: /etc/rancher/k3s"
