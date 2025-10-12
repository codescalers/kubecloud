#!/bin/bash
set -euo pipefail

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Check if .env file exists
if [ ! -f "$PROJECT_DIR/.env" ]; then
    echo "❌ Error: .env file not found in $PROJECT_DIR"
    echo "Please create a .env file with required configuration."
    echo "See native-k3s.md for configuration examples."
    exit 1
fi

# Source environment variables
echo "📋 Loading configuration from .env..."
source "$PROJECT_DIR/.env"

# Validate required variables
if [ -z "${K3S_TOKEN:-}" ]; then
    echo "❌ Error: K3S_TOKEN is not set in .env"
    exit 1
fi

if [ -z "${MASTER:-}" ]; then
    echo "❌ Error: MASTER is not set in .env"
    exit 1
fi

echo "🚀 Starting K3s native installation..."
echo "   Node Name: ${K3S_NODE_NAME:-$(hostname)}"
echo "   Master: ${MASTER}"
echo "   HA: ${HA:-false}"
echo "   Dual Stack: ${DUAL_STACK:-false}"

# Step 1: Prepare system
echo ""
echo "=== Step 1: System Preparation ==="
if ! command -v k3s &> /dev/null; then
    "$SCRIPT_DIR/prepare.sh"
else
    echo "✅ K3s already installed, skipping preparation"
fi

# Step 2: Create flannel bridge for dual-stack
if [ "${DUAL_STACK:-false}" = "true" ]; then
    echo ""
    echo "=== Step 2: Configure Dual-Stack Networking ==="
    "$SCRIPT_DIR/create_flannel_iface.sh"
else
    echo ""
    echo "=== Step 2: Skipping dual-stack configuration ==="
fi

# Step 3: Start NTP daemon for time synchronization
echo ""
echo "=== Step 3: Starting NTP Service ==="
if command -v ntpd &> /dev/null; then
    # Kill existing ntpd if running
    pkill ntpd || true
    # Start ntpd in background
    ntpd -n &
    echo "✅ NTP daemon started"
else
    echo "⚠️  NTP not available, time synchronization may be affected"
fi

# Step 4: Start K3s
echo ""
echo "=== Step 4: Starting K3s ==="
"$SCRIPT_DIR/entrypoint.sh" &
K3S_PID=$!
echo "✅ K3s started with PID: $K3S_PID"

# Step 5: Wait for K3s to be ready and install CRDs
if [ -z "${K3S_URL:-}" ] && [ "${MASTER}" = "true" ]; then
    echo ""
    echo "=== Step 5: Waiting for K3s API Server ==="
    
    # Set KUBECONFIG if not already set
    export KUBECONFIG="${KUBECONFIG:-/etc/rancher/k3s/k3s.yaml}"
    
    # Wait up to 60 seconds for API server
    for i in {1..12}; do
        if kubectl get --raw='/readyz' &> /dev/null; then
            echo "✅ K3s API server is ready!"
            break
        fi
        echo "   Waiting for API server... ($i/12)"
        sleep 5
    done
    
    # Install CRDs
    echo ""
    echo "=== Step 6: Installing CRDs ==="
    "$SCRIPT_DIR/install_crd.sh" || echo "⚠️  CRD installation failed (this is optional)"
else
    echo ""
    echo "=== Step 5: Skipping CRD installation (not first master) ==="
fi

echo ""
echo "=========================================="
echo "✅ K3s installation complete!"
echo "=========================================="
echo ""
echo "Cluster Information:"
if [ "${MASTER}" = "true" ]; then
    echo "  Role: Server/Master"
    if [ -z "${K3S_URL:-}" ]; then
        echo "  Type: First Master (Leader)"
        echo "  Kubeconfig: ${KUBECONFIG:-/etc/rancher/k3s/k3s.yaml}"
        echo ""
        echo "To access the cluster:"
        echo "  export KUBECONFIG=${KUBECONFIG:-/etc/rancher/k3s/k3s.yaml}"
        echo "  kubectl get nodes"
    else
        echo "  Type: Additional Master"
    fi
else
    echo "  Role: Agent/Worker"
fi
echo ""
echo "To view logs:"
echo "  tail -f /var/log/k3s.log"
echo ""
echo "K3s is running in the background (PID: $K3S_PID)"
