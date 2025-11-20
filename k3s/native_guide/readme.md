# Native K3s Single-stack Installation over Mycelium

K3s cluster deployment on Ubuntu VMs, using Mycelium for single-stack IPv6 networking. All cluster traffic is routed through Mycelium.

## Quick Start

```bash
# 1. Configure environment
cp .env.example .env
# Edit .env with your configuration

# 2. Run as root
sudo ./k3s.sh .env
```

**Note:** K3s stores its kubeconfig at `/etc/rancher/k3s/k3s.yaml` on server nodes.

## How It Works

The `k3s.sh` script is an all-in-one installer that:

1. **Installs dependencies** - K3s binary (v1.33.1), kubectl, curl, iproute2
2. **Detects Mycelium IPv6** - Automatically finds your Mycelium IPv6 address from the interface
3. **Configures K3s** - Sets up IPv6-only cluster with proper CIDRs
4. **Patches CRD manifests** - Injects environment variables (MNEMONIC, NETWORK) into manifests
5. **Starts K3s** - Runs as server or agent based on your configuration
6. **Auto-applies manifests** - Resources in `$K3S_DATA_DIR/server/manifests/` are automatically deployed

## Configuration

### Environment Variables

Configure your cluster by setting these variables in `.env`:

#### Network Configuration
- **K3S_FLANNEL_IFACE** - Network interface for CNI
  - Required: (Mycelium interface)
  - All traffic routes through this interface

#### Cluster Configuration
- **K3S_TOKEN** - Secret for cluster authentication (must match on all nodes)
  - Generate: `openssl rand -base64 32`
- **K3S_DATA_DIR** - K3s data directory (default: `/var/lib/rancher/k3s`)
  - Requires 20GB+ storage
- **K3S_NODE_NAME** - Unique node identifier

#### Node Role Configuration
- **MASTER** - Node type
  - `true` = Server/Control Plane
  - `false` = Agent/Worker
- **HA** - High availability mode
  - `true` = Enable HA (first master only)
  - `false` = Single master
- **K3S_URL** - Cluster join URL
  - Empty = First master node
  - `https://[mycelium-ipv6]:6443` = Join existing cluster

#### CRD Configuration (Optional)
- **MNEMONIC** - TF grid mnemonic for CRD controller
  - Raw mnemonic (12 words) or encrypted mnemonic
- **NETWORK** - ThreeFold network identifier (`dev`, `test`, `main`)
- **TOKEN** - Encryption password for mnemonic
  - **Only required if MNEMONIC is encrypted**
  - Leave empty for raw mnemonic

### Node Configuration Examples

#### First Master (Initial Server)
```bash
K3S_FLANNEL_IFACE=mycelium
K3S_TOKEN="your-secure-token"
K3S_DATA_DIR=/mnt/data
K3S_NODE_NAME=master-1
MASTER=true
HA=true
K3S_URL=
MNEMONIC="word1 word2 ... word12"
NETWORK="mainnet"
TOKEN=
```

#### Additional Master (HA)
```bash
K3S_FLANNEL_IFACE=eth1
K3S_TOKEN="your-secure-token"
K3S_DATA_DIR=/mnt/data
K3S_NODE_NAME=master-2
MASTER=true
HA=false
K3S_URL="https://[400:abcd:ef01:2345::1]:6443"
```

#### Worker Node
```bash
K3S_FLANNEL_IFACE=eth0
K3S_TOKEN="your-secure-token"
K3S_DATA_DIR=/mnt/data
K3S_NODE_NAME=worker-1
MASTER=false
HA=false
K3S_URL="https://[400:abcd:ef01:2345::1]:6443"
```

## Custom Resource Definitions (CRD)

CRDs are automatically deployed via K3s manifest auto-apply feature.

### How It Works

1. Manifests in `manifests/` directory are copied to `$K3S_DATA_DIR/server/manifests/`
2. K3s automatically applies manifests on server startup
3. `patch-crd-manifest.sh` replaces placeholders in manifests:
   - `${MNEMONIC}` → Your wallet mnemonic
   - `${NETWORK}` → Network identifier
   - `${TOKEN}` → Encryption password (if mnemonic is encrypted)

### On Agent Nodes

The script detects agent nodes (K3S_URL is set) and automatically removes CRD manifests since only server nodes should deploy cluster-wide resources.

## Side Effects & Cleanup

### Running k3s.sh

The script makes these system changes:
- Moves K3s data from `/var/lib/rancher/k3s` to `$K3S_DATA_DIR` (if configured)
- Creates CNI network interfaces (cni0, flannel.1, flannel-v6.1, etc.)
- Adds iptables/ip6tables rules for cluster networking
- Creates kubelet mounts in `/var/lib/kubelet`
- Starts K3s process (exec, replaces script process)

### Cleanup

Check this script which will stop all services and unmount affected directories. 

```bash
sudo ./k3s_killall.sh
```

**Note:** Data in `$K3S_DATA_DIR` is preserved. Delete manually if needed.

## Testing

### Test 1: Mycelium Connectivity

Verify cross-node communication over Mycelium.

**On Master Node:**
```bash
# Deploy nginx pod
kubectl run test-nginx --image=nginx --port=80

# Get pod IP
kubectl get pod test-nginx -o wide
```

**On Worker Node:**
```bash
# Test connectivity from worker node
curl http://[pod-ip]

# Expected output:
# <!DOCTYPE html>
# <html>
# <head>
# <title>Welcome to nginx!</title>
# ...
```

**Cleanup:**
```bash
kubectl delete pod test-nginx
```

### Test 2: CRD Installation

Verify TFGW CRD is installed and functional.

**Check CRD exists:**
```bash
kubectl get crd tfgws.ingress.grid.tf

# Expected output:
# NAME                      CREATED AT
# tfgws.ingress.grid.tf     2025-10-15T10:30:00Z
```

**Check CRD controller:**
```bash
kubectl get pods -n crd-system

# Expected output:
# NAME                                     READY   STATUS    RESTARTS   AGE
# crd-controller-manager-xxxxxxxxxx-xxxxx  1/1     Running   0          5m
```

**Create test TFGW resource:**
```bash
cat <<EOF | kubectl apply -f -
apiVersion: ingress.grid.tf/v1
kind: TFGW
metadata:
  labels:
    app.kubernetes.io/name: crd
    app.kubernetes.io/managed-by: kustomize
  name: my-tfgw
spec:
  hostname: "omar"
  backends:
    - "http://[5ce:20f3:1d33:d235:ff0f:b265:334f:e240]:80"
EOF
```

**Verify TFGW status:**
```bash
kubectl get tfgw my-tfgw

# Expected output:
# NAME           HOST   BACKENDS            FQDN
# my-tfgw   test   ["http://..."]      omar.gent02.dev.grid.tf
```

**Cleanup:**
```bash
kubectl delete tfgw my-tfgw
```
