# K3s Image

[![K3s Version](https://img.shields.io/badge/K3s-v1.26.0-blue)](https://github.com/k3s-io/k3s/releases/tag/v1.26.0)
[![Docker Image](https://img.shields.io/badge/Docker-threefoldtech%2Fk3s-green)](https://hub.docker.com/r/threefoldtech/k3s)

This image provides a lightweight K3s Kubernetes distribution with enhanced support for:

- High availability (HA) cluster configuration
- Dual stack networking (IPv4/IPv6)
- Easy deployment on ThreeFold Grid

## Features

- **Lightweight**: Minimal resource requirements compared to full Kubernetes
- **High Availability**: Support for multi-master setup
- **Dual Stack**: Full IPv4/IPv6 networking support
- **Simplified Setup**: Easy configuration through environment variables
- **ThreeFold Integration**: Ready for deployment on ThreeFold Grid

## Prerequisites

- Docker installed (for local development)
- Network connectivity between nodes
- Sufficient privileges for container execution

## Building

To build the K3s image locally:

```bash
# Navigate to the k3s directory
cd k3s

# Build the Docker image
docker build -t threefoldtech/k3s:latest .
```

## Running

### Leader/Master Node

```bash
docker run -it --name master \
  -e K3S_URL="" \
  -e K3S_TOKEN="<YOUR_CLUSTER_TOKEN>" \
  -e MASTER="true" \
  --privileged \
  threefoldtech/k3s:latest
```

### Worker Node

```bash
docker run -it --name worker \
  -e K3S_URL="https://<MASTER_IP>:6443" \
  -e K3S_TOKEN="<YOUR_CLUSTER_TOKEN>" \
  --privileged \
  threefoldtech/k3s:latest
```

### High Availability Setup

```bash
# First master (leader)
docker run -it --name master1 \
  -e K3S_URL="" \
  -e K3S_TOKEN="<YOUR_CLUSTER_TOKEN>" \
  -e MASTER="true" \
  -e HA="true" \
  --privileged \
  threefoldtech/k3s:latest

# Additional masters
docker run -it --name master2 \
  -e K3S_URL="https://<LEADER_IP>:6443" \
  -e K3S_TOKEN="<YOUR_CLUSTER_TOKEN>" \
  -e MASTER="true" \
  --privileged \
  threefoldtech/k3s:latest
```

## ThreeFold Deployment

### Flist

The K3s image is available as a Flist for ThreeFold Grid deployment:

[https://hub.threefold.me/omarabdulaziz.3bot/omarabdul3ziz-k3s-opt_crypto.flist](https://hub.threefold.me/omarabdulaziz.3bot/omarabdul3ziz-k3s-opt_crypto.flist)

### Dual-Stack Networking with Mycelium

This k3s deployment provides secure dual-stack Kubernetes networking that integrates seamlessly with Mycelium on ThreeFold Grid.

#### Architecture Overview

The cluster implements a **layered security model** with private pod networking and controlled external access:

- **Private Pod Network**: Pods run in an isolated Flannel overlay network with private IPv4 (`10.42.0.0/16`) and IPv6 (`2001:cafe:42::/56`) address ranges
- **Mycelium Integration**: Each node has a Mycelium interface managed by ZOS or a DaemonSet, providing secure end-to-end encrypted connectivity
- **Controlled Exposure**: Pods remain secure by default; only explicitly exposed services are accessible via Mycelium

#### How It Works

1. **Secure by Default**: All pods communicate internally using private Flannel overlay networks (IPv4 and IPv6). These addresses are not directly accessible from outside the cluster.

2. **Mycelium Access**: The Mycelium interface on each node enables secure access to cluster services without exposing the underlying pod network. Access is controlled through standard Kubernetes primitives:
   - **Ingress Rules**: Route external traffic to internal services
   - **ClusterIP Services**: Expose pods through stable service endpoints accessible via Mycelium
   - **No Helm Chart Modifications**: Works out-of-the-box with existing applications

3. **Development Workflow**: For testing and development, use `kubectl expose` to quickly make pods/services accessible through the Mycelium network.

#### Network Configuration

The cluster uses **Flannel with VXLAN backend** for dual-stack support:

- **IPv4 Cluster CIDR**: `10.42.0.0/16` (private)
- **IPv6 Cluster CIDR**: `2001:cafe:42::/56` (private)
- **IPv4 Service CIDR**: `10.43.0.0/16`
- **IPv6 Service CIDR**: `2001:cafe:43::/112`

#### Required Network Interfaces

- **IPv4 Interface**: `eth0` - Primary network interface for IPv4 connectivity
- **IPv6 Interface**: `eth2` or Mycelium interface - Dedicated interface for IPv6 connectivity

#### Deployment Steps

1. **Enable Dual-Stack**: Set `DUAL_STACK="true"` environment variable
2. **Interface Configuration**: The entrypoint script automatically creates a bridge interface (`flannel-br`) that merges IPv4 (eth0) and IPv6 interfaces
3. **Flannel Interface**: The `K3S_FLANNEL_IFACE` defaults to `flannel-br` when dual-stack is enabled

#### Example Deployment

```bash
# Deploy with dual-stack enabled
docker run -it --name k3s-master \
  -e K3S_URL="" \
  -e K3S_TOKEN="<YOUR_CLUSTER_TOKEN>" \
  -e MASTER="true" \
  -e DUAL_STACK="true" \
  --privileged \
  threefoldtech/k3s:latest
```

**Note**: All Kubernetes services created on the cluster will automatically have `ipFamilyPolicy: RequireDualStack` set, ensuring both IPv4 and IPv6 addresses are assigned.

#### Security Benefits

- **Network Isolation**: Pods run in private networks, reducing attack surface
- **Selective Exposure**: Only services you explicitly expose are accessible
- **Zero Trust**: Mycelium provides end-to-end encryption for all external access
- **No Configuration Overhead**: Standard Kubernetes networking works without modification

### Entrypoint

The default entrypoint for the container is:

```bash
zinit init
```

## Configuration

### Environment Variables

| Variable | Description | Default | Required |
|----------|-------------|---------|----------|
| `K3S_URL` | URL of the leader node. Empty for leader, `https://<LEADER_IP>:6443` for workers/additional masters | - | Yes |
| `K3S_TOKEN` | Authentication token for the cluster (must be identical across all nodes) | - | Yes |
| `K3S_DATA_DIR` | Data directory for Kubernetes | `/var/lib/rancher/k3s/` | No |
| `K3S_FLANNEL_IFACE` | Network interface used by Flannel | `eth0` (single stack) or `flannel-br` (dual stack) | No |
| `K3S_DATASTORE_ENDPOINT` | External datastore endpoint (etcd, sqlite, postgres, mysql) | - | No |
| `K3S_NODE_NAME` | Custom node name | Hostname | No |
| `DUAL_STACK` | Enable dual stack (IPv4/IPv6) networking | `false` | No |
| `MASTER` | Configure node as a master | `false` | No |
| `HA` | Enable high availability mode on leader node | `false` | No |

## Persistent Storage

For production deployments, it's recommended to mount persistent storage to an external location and set K3S_DATA_DIR to point to this location:

```bash
docker run -it --name master \
  -e K3S_URL="" \
  -e K3S_TOKEN="<YOUR_CLUSTER_TOKEN>" \
  -e MASTER="true" \
  -e K3S_DATA_DIR="/mnt/data" \
  -v /path/to/storage:/mnt/data \
  --privileged \
  threefoldtech/k3s:latest
```

This approach ensures your Kubernetes data is stored on the mounted volume rather than inside the container.

## Troubleshooting

### Common Issues

- **Nodes not joining the cluster**: Verify network connectivity and that the correct K3S_TOKEN is being used
- **Dual stack not working**: Ensure the correct network interface is specified with K3S_FLANNEL_IFACE
- **Container fails to start**: Check for sufficient privileges (--privileged flag)

### Logs

To view container logs:

```bash
docker logs <container_name>
```

## Contributing

Contributions to improve this image are welcome. Please follow the standard GitHub workflow:

1. Fork the repository
2. Create a feature branch
3. Submit a pull request

## License

This project is licensed under the Apache License 2.0 - see the LICENSE file for details.
