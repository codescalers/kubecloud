# Getting Started with Mycelium Cloud

Welcome to Mycelium Cloud, a comprehensive platform for deploying and managing Kubernetes clusters on the decentralized ThreeFold Grid infrastructure.

## Overview

Mycelium Cloud provides a complete solution for cloud-native applications with:

- **Decentralized Infrastructure**: Deploy on ThreeFold Grid's distributed network
- **Kubernetes Management**: Full K3s cluster deployment and management
- **IPv6 Networking**: Mycelium peer-to-peer networking
- **High Availability**: Multi-master cluster support
- **Monitoring**: Integrated Prometheus and Grafana
- **Web Interface**: Modern Vue.js dashboard

## Quick Start

### 1. Account Setup

1. **Sign Up**: Create your account at [Mycelium Cloud](https://staging.vdc.grid.tf/sign-up)
2. **Verify Email**: Check your email and verify your account
3. **Add Funds**: Navigate to your dashboard and add credits to your account

### 2. Deploy Your First Cluster

1. **Access Deploy**: Click "Deploy Cluster" from your dashboard
2. **Configure VMs**: Define your virtual machines:
   - Choose CPU, memory, and storage requirements
   - Select the number of master and worker nodes
3. **Select Nodes**: Choose ThreeFold Grid nodes for deployment
4. **Review & Deploy**: Confirm your configuration and deploy

### 3. Access Your Cluster

Once deployed, you can:

- **Download Kubeconfig**: Get your cluster configuration file
- **Monitor Status**: View cluster health and metrics
- **Manage Resources**: Scale nodes up or down as needed

## Configuration

### Backend Configuration

Mycelium Cloud supports configuration through environment variables, CLI flags, and configuration files.

#### Configuration File

By default, Mycelium Cloud looks for a `config.json` file in the current directory. You can specify a custom configuration file path using the `--config` or `-c` flag:

```bash
myceliumcloud --config /path/to/config.json
```

The configuration file should be in JSON format. Check the [config example](https://github.com/codescalers/kubecloud/blob/master/backend/config-example.json) for reference.

#### Notification Configuration

Mycelium Cloud supports a separate notification configuration file to define how different types of notifications are handled:

```bash
kubecloud --notification_config_path /path/to/notification-config.json
```

Or set via environment variable:

```bash
export KUBECLOUD_NOTIFICATION_CONFIG_PATH=/path/to/notification-config.json
```

#### Default Behavior

If no notification configuration file is provided, Mycelium Cloud will use default settings:

- **All channels**: `["ui"]` (UI notifications only)
- **All severity levels**: `"info"`
- **All notification types**: Use the default settings unless specifically overridden

## Key Features

### Decentralized Deployment

Deploy Kubernetes clusters across the ThreeFold Grid's decentralized infrastructure for enhanced reliability and geographic distribution.

### IPv6 Networking

Built-in Mycelium networking provides secure peer-to-peer IPv6 connectivity between all cluster components.

### High Availability

Configure multi-master clusters for production workloads with automatic failover capabilities.

### Monitoring & Observability

Integrated Prometheus metrics collection and Grafana dashboards for comprehensive cluster monitoring.

## Next Steps

- [Platform Tutorial](https://github.com/codescalers/kubecloud/blob/master/frontend/kubecloud/public/docs/tutorial.md) - Complete walkthrough including Hello World, 3 Python servers, and service communication examples
- [Architecture Overview](https://github.com/codescalers/kubecloud/blob/master/frontend/kubecloud/public/docs/architecture.md) - Deep dive into Mycelium networking and system design
- [API Reference](https://github.com/codescalers/kubecloud/blob/master/frontend/kubecloud/public/docs/api-reference.md) - Complete API documentation
- [FAQ](https://github.com/codescalers/kubecloud/blob/master/frontend/kubecloud/public/docs/faq.md) - Frequently asked questions and troubleshooting

## Support

Need help? Contact our support team or check our community resources:

- GitHub Issues: [Report bugs and feature requests](https://github.com/codescalers/kubecloud/issues)
- Documentation: Browse our comprehensive guides
- Community: Join our community channels
