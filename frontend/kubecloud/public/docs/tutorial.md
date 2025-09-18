# Platform Tutorial

This comprehensive tutorial will guide you through deploying and managing your first Kubernetes cluster on Mycelium Cloud.

## Prerequisites

Before starting, ensure you have:
- A Mycelium Cloud account ([Sign up here](https://staging.vdc.grid.tf/sign-up))
- Verified email address
- Account credits for deployment costs
- Basic understanding of Kubernetes concepts

## Step 1: Account Setup and Preparation

### 1.1 Create Your Account

1. Navigate to [Mycelium Cloud](https://staging.vdc.grid.tf)
2. Click "Sign Up" and fill in your details
3. Verify your email address
4. Complete your profile setup

### 1.2 Add SSH Keys

SSH keys are essential for secure access to your cluster nodes:

1. Go to your **Dashboard**
2. Navigate to **SSH Keys** section
3. Click **Add SSH Key**
4. Provide a name and paste your public key
5. Save the key

**Generate SSH Key (if needed):**
```bash
ssh-keygen -t rsa -b 4096 -C "your-email@example.com"
cat ~/.ssh/id_rsa.pub  # Copy this content
```

### 1.3 Fund Your Account

1. Go to **Billing** section in your dashboard
2. Click **Add Funds**
3. Enter the amount and payment details
4. Complete the payment process

## Step 2: Planning Your Cluster

### 2.1 Cluster Architecture

For this tutorial, we'll deploy a simple but production-ready cluster:
- **1 Master Node**: 4 CPU, 8GB RAM, 100GB storage
- **2 Worker Nodes**: 2 CPU, 4GB RAM, 50GB storage each

### 2.2 Resource Requirements

Estimate your needs based on workload:
- **Development**: 1 master, 1 worker (minimal resources)
- **Staging**: 1 master, 2 workers (moderate resources)
- **Production**: 3 masters, 3+ workers (high availability)

## Step 3: Deploying Your Cluster

### 3.1 Start Deployment

1. From your dashboard, click **Deploy Cluster**
2. You'll enter the 3-step deployment wizard

### 3.2 Step 1: Define VMs

Configure your virtual machines:

**Master Node Configuration:**
- **Name**: `production-master-1`
- **CPU**: 4 cores
- **Memory**: 8192 MB (8GB)
- **Storage**: 100 GB
- **Type**: Master
- **SSH Key**: Select your uploaded key

**Worker Node 1:**
- **Name**: `production-worker-1`
- **CPU**: 2 cores
- **Memory**: 4096 MB (4GB)
- **Storage**: 50 GB
- **Type**: Worker
- **SSH Key**: Select your uploaded key

**Worker Node 2:**
- **Name**: `production-worker-2`
- **CPU**: 2 cores
- **Memory**: 4096 MB (4GB)
- **Storage**: 50 GB
- **Type**: Worker
- **SSH Key**: Select your uploaded key

Click **Next** to proceed.

### 3.3 Step 2: Assign Nodes

Select ThreeFold Grid nodes for deployment:

1. **Filter Options**: Use filters to find suitable nodes:
   - **Country**: Choose your preferred region
   - **Minimum Resources**: Ensure nodes meet your requirements
   - **Farm Rating**: Select highly-rated farms

2. **Node Selection**: 
   - Click on available nodes to assign them to your VMs
   - Ensure geographic distribution for better latency
   - Verify node specifications match your requirements

3. **Review Assignments**: Confirm each VM is assigned to an appropriate node

Click **Next** to continue.

### 3.4 Step 3: Review and Deploy

1. **Review Configuration**: 
   - Verify all VM specifications
   - Check node assignments
   - Review estimated costs

2. **Cluster Settings**:
   - **Cluster Name**: `my-production-cluster`
   - **Network Configuration**: IPv6 with Mycelium networking
   - **High Availability**: Enable if using multiple masters

3. **Deploy**: Click **Deploy Cluster** to start the deployment

### 3.5 Monitor Deployment

1. You'll be redirected to the cluster management page
2. Monitor deployment progress in real-time
3. Deployment typically takes 5-15 minutes
4. Status will change from "Deploying" to "Running"

## Step 4: Accessing Your Cluster

### 4.1 Download Kubeconfig

Once deployment is complete:

1. Go to your cluster management page
2. Click **Download Kubeconfig**
3. Save the file as `~/.kube/config` (or merge with existing config)

### 4.2 Verify Cluster Access

Test your cluster connection:

```bash
# Check cluster info
kubectl cluster-info

# List nodes
kubectl get nodes

# Check node status
kubectl get nodes -o wide

# Verify all pods are running
kubectl get pods --all-namespaces
```

### 4.3 Install kubectl (if needed)

**Linux/macOS:**
```bash
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"
chmod +x kubectl
sudo mv kubectl /usr/local/bin/
```

**Windows:**
```powershell
curl -LO "https://dl.k8s.io/release/v1.28.0/bin/windows/amd64/kubectl.exe"
```

## Step 5: Deploying Your First Application

### 5.1 Deploy a Sample Application

Let's deploy a simple nginx application:

```yaml
# nginx-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: nginx-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: nginx
  template:
    metadata:
      labels:
        app: nginx
    spec:
      containers:
      - name: nginx
        image: nginx:1.21
        ports:
        - containerPort: 80
---
apiVersion: v1
kind: Service
metadata:
  name: nginx-service
spec:
  selector:
    app: nginx
  ports:
  - port: 80
    targetPort: 80
  type: ClusterIP
```

Deploy the application:
```bash
kubectl apply -f nginx-deployment.yaml
```

### 5.2 Verify Deployment

```bash
# Check deployment status
kubectl get deployments

# Check pods
kubectl get pods -l app=nginx

# Check service
kubectl get services
```

### 5.3 Access Your Application

Create a port-forward to access the application:
```bash
kubectl port-forward service/nginx-service 8080:80
```

Open your browser to `http://localhost:8080` to see nginx running.

## Step 6: Monitoring and Management

### 6.1 Cluster Monitoring

Access built-in monitoring:

1. **Mycelium Cloud Dashboard**: View cluster metrics and status
2. **Grafana**: Access via the monitoring section (if enabled)
3. **Prometheus**: Query metrics directly

### 6.2 Resource Monitoring

Monitor resource usage:
```bash
# Node resource usage
kubectl top nodes

# Pod resource usage
kubectl top pods --all-namespaces

# Describe node details
kubectl describe node <node-name>
```

### 6.3 Scaling Your Cluster

#### Scale Applications
```bash
# Scale deployment
kubectl scale deployment nginx-deployment --replicas=5

# Verify scaling
kubectl get pods -l app=nginx
```

#### Add Worker Nodes
1. Go to cluster management page
2. Click **Edit Cluster**
3. Add new worker nodes
4. Deploy changes

## Step 7: Best Practices

### 7.1 Security

- **RBAC**: Implement Role-Based Access Control
- **Network Policies**: Control pod-to-pod communication
- **Secrets Management**: Use Kubernetes secrets for sensitive data
- **Regular Updates**: Keep cluster and applications updated

### 7.2 Resource Management

- **Resource Limits**: Set CPU and memory limits for pods
- **Namespaces**: Organize applications using namespaces
- **Storage Classes**: Use appropriate storage for different workloads

### 7.3 Backup and Disaster Recovery

- **etcd Backups**: Regular cluster state backups
- **Application Data**: Backup persistent volumes
- **Configuration**: Version control your YAML manifests

## Step 8: Troubleshooting

### 8.1 Common Issues

**Pods Not Starting:**
```bash
kubectl describe pod <pod-name>
kubectl logs <pod-name>
```

**Node Issues:**
```bash
kubectl describe node <node-name>
kubectl get events --sort-by=.metadata.creationTimestamp
```

**Network Issues:**
```bash
kubectl get endpoints
kubectl describe service <service-name>
kubectl get networkpolicy -A
kubectl describe networkpolicy -n <namespace> <name>
```

**Kubeconfig / Access Issues:**
```bash
# Show current context and clusters
kubectl config view --minify
kubectl config get-contexts
kubectl config use-context <context>

# Verify KUBECONFIG path (if using a custom location)
echo $KUBECONFIG   # (Linux/macOS)
$Env:KUBECONFIG    # (Windows PowerShell)

# Inspect certificate expiration (if applicable)
kubectl -n kube-system get secrets | grep kube
```

**Storage (PV/PVC) Issues:**
```bash
# Check PVC and PV status
kubectl get pvc -A
kubectl get pv

# Describe problematic claims
kubectl describe pvc -n <namespace> <pvc-name>

# Check StorageClass and provisioner
kubectl get storageclass
kubectl describe storageclass <name>
```

**Ingress / DNS Issues:**
```bash
# Verify ingress resources
kubectl get ingress -A
kubectl describe ingress -n <namespace> <name>

# Test service reachability inside the cluster
kubectl run tmp --image=busybox:1.36 --restart=Never -it --rm -- sh -c "wget -qO- http://<service>.<namespace>.svc.cluster.local:<port>"

# Validate DNS resolution from your workstation
nslookup <your-domain>
curl -v https://<your-domain>
```

**Collect Diagnostics Quickly:**
```bash
# Save a snapshot of cluster state (non-sensitive)
kubectl cluster-info dump --all-namespaces --output-directory=./cluster-dump

# Events sorted by time
kubectl get events -A --sort-by=.metadata.creationTimestamp | tail -n 200
```

### 8.2 Getting Help

- **Logs**: Check cluster and application logs
- **Events**: Monitor Kubernetes events
- **Support**: Contact Mycelium Cloud support
- **Community**: Join our community channels
- **Issues**: Open a ticket on GitHub with details: https://github.com/codescalers/kubecloud/issues

## Next Steps

Congratulations! You've successfully deployed and managed your first Kubernetes cluster on Mycelium Cloud. Here's what to explore next:

1. **Advanced Networking**: Configure ingress controllers and load balancers
2. **Storage Solutions**: Implement persistent storage for stateful applications
3. **CI/CD Integration**: Set up automated deployment pipelines
4. **Monitoring**: Advanced monitoring with custom metrics
5. **Security**: Implement advanced security policies

## Additional Resources

- [API Reference](./api-reference.md) - Complete API documentation
- [FAQ](./faq.md) - Frequently asked questions
- [Architecture Guide](./architecture.md) - Deep dive into platform architecture
- [Best Practices](./best-practices.md) - Production deployment guidelines
