# Getting Started with Mycelium Cloud

Welcome to Mycelium Cloud, a comprehensive platform for deploying and managing Kubernetes clusters on the decentralized ThreeFold Grid infrastructure.

## Overview

Mycelium Cloud provides a complete solution for cloud-native applications with:

- **Decentralized Infrastructure**: Deploy on ThreeFold Grid's distributed network
- **Kubernetes Management**: Full K3s cluster deployment and management
- **IPv6 Networking**: Mycelium peer-to-peer networking
- **High Availability**: Multi-master cluster support

## Architecture

Mycelium Cloud uses peer-to-peer networking that enables:

- **Direct Node Access**: Each node gets a unique Mycelium IP address
- **Cross-Node Communication**: Services communicate across nodes using Mycelium networking
- **Secure Communication**: All traffic is encrypted through the Mycelium network
- **No Public IPs Required**: Services accessible via Mycelium IPs

**Network Flow**: `User Machine → Mycelium Network → Cluster Node → Service`

## Quick Start

### 1. Account Setup

1. **Sign Up**: Create your account from signup page
2. **Verify Email**: Check your email and verify your account
3. **Add Funds**: Navigate to your dashboard and add credits to your account
4. **Add SSH Key**: Navigate to Add SSH card and upload your public SSH key

### 2. Deploy Your First Cluster

1. **Access Deploy**: Click "Deploy Cluster" from your dashboard
2. **Configure VMs**: Define your virtual machines:
   - Choose CPU, memory, and storage requirements
   - Select the number of master and worker nodes
3. **Select Nodes**: Choose ThreeFold Grid nodes for deployment
4. **Review & Deploy**: Confirm your configuration and deploy

### 3. Access Your Cluster

#### Download Kubeconfig

1. Go to dashboard → Clusters → Click download icon (⬇️)
2. Set kubeconfig: `export KUBECONFIG=/path/to/config`
3. Test: `kubectl get nodes`

#### SSH Access

1. **Find Mycelium IPs**: Check cluster details page for node IPs
2. **Download Mycelium Binary**:

   ```bash
   wget https://github.com/threefoldtech/mycelium/releases/latest/download/mycelium-private-x86_64-unknown-linux-musl.tar.gz
   tar -xzf mycelium-private-x86_64-unknown-linux-musl.tar.gz
   sudo chmod +x mycelium-private
   sudo mv mycelium-private /usr/local/bin/mycelium
   ```

3. **Start Mycelium**:

   ```bash
   sudo mycelium --peers tcp://188.40.132.242:9651 tcp://136.243.47.186:9651 tcp://185.69.166.7:9651 tcp://185.69.166.8:9651 tcp://65.21.231.58:9651 tcp://65.109.18.113:9651 tcp://209.159.146.190:9651 tcp://5.78.122.16:9651 tcp://5.223.43.251:9651 tcp://142.93.217.194:9651
   ```

4. **SSH to nodes**: `ssh root@<mycelium-ip>`
