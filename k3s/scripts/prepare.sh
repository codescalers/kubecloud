#!/bin/bash
set -euo pipefail

readonly K3S_VERSION="v1.33.1+k3s1"
readonly KUBECTL_VERSION="v1.33.1"

echo "🔧 Preparing system for K3s installation..."

echo "[1/4] Updating package list..."
apt-get update -qq && apt upgrade -y

echo "[2/4] Installing dependencies..."
apt-get install -y wget iproute2

echo "[3/4] Downloading K3s ${K3S_VERSION}..."
wget -q --show-progress -O /usr/local/bin/k3s \
    "https://github.com/k3s-io/k3s/releases/download/${K3S_VERSION}/k3s"
chmod +x /usr/local/bin/k3s

echo "[4/4] Downloading kubectl ${KUBECTL_VERSION}..."
wget -q --show-progress -O /usr/local/bin/kubectl \
    "https://dl.k8s.io/release/${KUBECTL_VERSION}/bin/linux/amd64/kubectl"
chmod +x /usr/local/bin/kubectl

chmod +x ~/scripts/*