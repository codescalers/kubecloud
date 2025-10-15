# Frequently Asked Questions

Find answers to common questions about Mycelium Cloud platform.

## General Questions

### What is Mycelium Cloud?

Mycelium Cloud is a comprehensive platform for deploying and managing Kubernetes clusters on the decentralized ThreeFold Grid infrastructure. It provides a complete solution with backend APIs, frontend UI, custom networking, and monitoring capabilities.

### What makes Mycelium Cloud different from other Kubernetes platforms?

- **Decentralized Infrastructure**: Runs on ThreeFold Grid's distributed network
- **IPv6 Networking**: Built-in Mycelium peer-to-peer networking
- **Cost Effective**: Competitive pricing on decentralized infrastructure
- **Geographic Distribution**: Deploy across global node locations
- **Simplified Management**: Easy-to-use web interface and APIs

### Is Mycelium Cloud suitable for production workloads?

Yes, Mycelium Cloud supports production workloads with:

- High availability cluster configurations
- Multi-master setups
- Persistent storage options
- Monitoring and alerting
- Backup and disaster recovery capabilities

## Account and Billing

### How do I create an account?

1. Visit [Mycelium Cloud](https://staging.vdc.grid.tf/sign-up)
2. Fill in your registration details
3. Verify your email address
4. Complete your profile setup

### What payment methods are accepted?

We accept:

- Credit cards (Visa, MasterCard, American Express)
- Stripe payment processing
- Account credits and vouchers

### How is billing calculated?

Billing is based on:

- **Compute Resources**: CPU cores and memory usage
- **Storage**: Disk space allocated
- **Network**: Data transfer (minimal charges)
- **Time**: Pay for actual usage time

### Can I get a refund?

Refunds are handled case-by-case. Contact our support team with your request and we'll review it based on our refund policy.

## Cluster Management

### What Kubernetes versions are supported?

We currently support K3s v1.26.0, which provides:

- Full Kubernetes API compatibility
- Lightweight resource usage
- High availability features
- Dual-stack networking (IPv4/IPv6)

### How many clusters can I deploy?

There's no hard limit on the number of clusters. Limits depend on:

- Your account credits
- Available ThreeFold Grid capacity
- Resource quotas (if any)

### Can I scale my cluster after deployment?

Yes, you can:

- Add or remove worker nodes
- Resize existing nodes (with redeployment)
- Scale applications independently
- Modify cluster configuration

### What happens if a node fails?

- **Worker Node Failure**: Kubernetes automatically reschedules pods to healthy nodes
- **Master Node Failure**: In HA setups, other masters take over
- **Complete Failure**: We provide backup and recovery options

### How do I backup my cluster?

Backup strategies include:

- **etcd Snapshots**: Automated cluster state backups
- **Persistent Volume Backups**: Application data backups
- **Configuration Backups**: YAML manifests and configurations

## Networking

### What networking is used?

Mycelium Cloud uses:

- **Mycelium CNI**: IPv6 peer-to-peer networking
- **Dual Stack**: Support for both IPv4 and IPv6
- **Secure Tunnels**: Encrypted communication between nodes
- **Load Balancing**: Built-in traffic distribution

### How do I expose applications to the internet?

Options include:

- **NodePort Services**: Direct node access
- **LoadBalancer Services**: Automatic load balancer creation
- **Ingress Controllers**: HTTP/HTTPS routing
- **ThreeFold Gateway**: Custom gateway solutions

### Can I use custom domains?

Yes, you can:

- Configure custom domains through ingress controllers
- Use ThreeFold Gateway for domain routing
- Set up DNS records pointing to your cluster
- Implement SSL/TLS certificates

## Security

### How secure is Mycelium Cloud?

Security features include:

- **Encrypted Communication**: All traffic encrypted in transit
- **Network Isolation**: Secure pod-to-pod communication
- **RBAC**: Role-based access control
- **SSH Key Authentication**: Secure node access
- **Regular Updates**: Automated security patches

### Can I use my own SSH keys?

Yes, SSH key management includes:

- Upload multiple SSH keys
- Assign keys to specific nodes
- Rotate keys as needed
- Secure shell access to nodes

### How is data protected?

Data protection measures:

- **Encryption at Rest**: Storage encryption
- **Encryption in Transit**: Network encryption
- **Access Controls**: User and application permissions
- **Audit Logs**: Activity monitoring and logging

## Technical Support

### What support is available?

Support options include:

- **Documentation**: Comprehensive guides and tutorials
- **Community**: Community forums and discussions
- **Email Support**: Direct support for technical issues
- **GitHub Issues**: Bug reports and feature requests

### How do I report a bug?

To report bugs:

1. Check existing [GitHub Issues](https://github.com/codescalers/kubecloud/issues)
2. Create a new issue with detailed information
3. Include logs, error messages, and reproduction steps
4. Our team will investigate and respond

### What information should I include in support requests?

Include:

- **Account Information**: Username/email (not password)
- **Cluster Details**: Cluster ID, name, configuration
- **Error Messages**: Full error text and logs
- **Steps to Reproduce**: What you were trying to do
- **Environment**: Browser, OS, kubectl version

## Platform Limitations

### Are there any resource limits?

Current limitations:

- **Node Resources**: Depends on available ThreeFold Grid nodes
- **Storage**: Limited by node storage capacity
- **Network**: Bandwidth depends on node connectivity
- **Regions**: Limited to ThreeFold Grid node locations

### What regions are available?

ThreeFold Grid nodes are available globally, including:

- North America
- Europe
- Asia Pacific
- Africa
- Middle East

Specific availability depends on active farms and nodes.

### Can I deploy in specific countries?

Yes, you can:

- Filter nodes by country during deployment
- Choose specific farms or regions
- Consider data sovereignty requirements
- Balance latency and compliance needs

## Troubleshooting

### My cluster deployment failed. What should I do?

Troubleshooting steps:

1. **Check Logs**: Review deployment logs in the dashboard
2. **Verify Resources**: Ensure sufficient account credits
3. **Node Availability**: Confirm selected nodes are available
4. **Configuration**: Validate cluster configuration
5. **Contact Support**: If issues persist, contact our team

### I can't connect to my cluster. How do I fix this?

Connection troubleshooting:

1. **Kubeconfig**: Ensure you've downloaded the correct kubeconfig
2. **kubectl**: Verify kubectl is installed and configured
3. **Network**: Check your internet connection
4. **Firewall**: Ensure no firewall blocking connections
5. **Cluster Status**: Verify cluster is running in dashboard

### My application pods are not starting. What's wrong?

Pod troubleshooting:

```bash
# Check pod status
kubectl get pods

# Describe pod for events
kubectl describe pod <pod-name>

# Check pod logs
kubectl logs <pod-name>

# Check node resources
kubectl top nodes
```

Common issues:

- **Resource Limits**: Insufficient CPU/memory
- **Image Issues**: Cannot pull container images
- **Configuration**: Invalid pod specifications
- **Storage**: Persistent volume issues

### How do I check cluster health?

Health monitoring:

```bash
# Check node status
kubectl get nodes

# Check system pods
kubectl get pods -n kube-system

# Check cluster info
kubectl cluster-info

# Check events
kubectl get events --sort-by=.metadata.creationTimestamp
```

Dashboard monitoring:

- Cluster status indicators
- Resource usage metrics
- Alert notifications
- Performance graphs

## Advanced Topics

### Can I use custom CNI plugins?

Currently, Mycelium Cloud uses the Mycelium CNI plugin for IPv6 networking. Custom CNI plugins are not supported, but the Mycelium CNI provides:

- Peer-to-peer connectivity
- IPv6 addressing
- Secure tunneling
- Integration with ThreeFold Grid

### How do I integrate with CI/CD pipelines?

Integration options:

- **API Access**: Use REST APIs for automation
- **kubectl**: Direct cluster access from pipelines
- **Webhooks**: Receive deployment notifications
- **GitOps**: Implement GitOps workflows

### Can I run Windows containers?

Currently, Mycelium Cloud focuses on Linux containers and nodes. Windows container support is not available at this time.

### How do I implement disaster recovery?

Disaster recovery strategies:

- **Multi-Region Deployment**: Deploy across multiple regions
- **Regular Backups**: Automated etcd and data backups
- **Infrastructure as Code**: Version-controlled configurations
- **Monitoring**: Proactive health monitoring
- **Recovery Procedures**: Documented recovery processes

## Getting More Help

Still have questions? Here's how to get additional help:

### Documentation

- [Getting Started Guide](./getting-started.md)
- [Platform Tutorial](./tutorial.md)
- [API Reference](./api-reference.md)
- [Architecture Overview](./architecture.md)

### Community

- GitHub Discussions
- Community Forums
- Discord/Slack Channels

### Direct Support

- Email: [support@myceliumcloud.com](mailto:support@myceliumcloud.com)
- GitHub Issues: [Report Issues](https://github.com/codescalers/kubecloud/issues)
- In-app Support: Use the help widget in your dashboard

### Emergency Support

For critical production issues:

- Mark support requests as "Critical"
- Include "URGENT" in subject lines
- Provide detailed impact assessment
- Include contact information for immediate response
