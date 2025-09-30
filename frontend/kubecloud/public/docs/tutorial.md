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

### 5.1 Deploy a Hello World Application

Let's deploy a simple "Hello World" application that demonstrates Mycelium Cloud's capabilities:

```yaml
# hello-world-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-world
  labels:
    app: hello-world
spec:
  replicas: 1
  selector:
    matchLabels:
      app: hello-world
  template:
    metadata:
      labels:
        app: hello-world
    spec:
      containers:
      - name: hello-world
        image: nginx:1.21
        ports:
        - containerPort: 80
        volumeMounts:
        - name: html-content
          mountPath: /usr/share/nginx/html
      volumes:
      - name: html-content
        configMap:
          name: hello-world-content
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: hello-world-content
data:
  index.html: |
    <!DOCTYPE html>
    <html>
    <head>
        <title>Hello from Mycelium Cloud!</title>
        <style>
            body { font-family: Arial, sans-serif; text-align: center; margin-top: 50px; }
            .container { max-width: 600px; margin: 0 auto; }
            .success { color: #28a745; }
            .info { color: #17a2b8; }
        </style>
    </head>
    <body>
        <div class="container">
            <h1 class="success">🚀 Hello from Mycelium Cloud!</h1>
            <p class="info">Your Kubernetes cluster is running on the decentralized ThreeFold Grid</p>
            <p>This application is connected via Mycelium's peer-to-peer IPv6 networking</p>
            <p><strong>Pod Name:</strong> <span id="pod-name">Loading...</span></p>
            <p><strong>Node:</strong> <span id="node-name">Loading...</span></p>
            <p><strong>Timestamp:</strong> <span id="timestamp"></span></p>
        </div>
        <script>
            // Display pod and node information
            document.getElementById('pod-name').textContent = window.location.hostname;
            document.getElementById('node-name').textContent = 'Mycelium Node';
            document.getElementById('timestamp').textContent = new Date().toLocaleString();
        </script>
    </body>
    </html>
---
apiVersion: v1
kind: Service
metadata:
  name: hello-world-service
  labels:
    app: hello-world
spec:
  selector:
    app: hello-world
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
  type: ClusterIP
```

Deploy the Hello World application:

```bash
kubectl apply -f hello-world-deployment.yaml
```

### 5.2 Verify Deployment

```bash
# Check deployment status
kubectl get deployments

# Check pods
kubectl get pods -l app=hello-world

# Check service
kubectl get services

# Check configmap
kubectl get configmaps
```

### 5.3 Access Your Application from Your Machine

#### Method 1: Port Forwarding (Recommended for Testing)

Create a port-forward to access the application from your local machine:

```bash
kubectl port-forward service/hello-world-service 8080:80
```

Open your browser to `http://localhost:8080` to see your Hello World application running.

#### Method 2: Using NodePort Service

If you want to access the application directly through a node IP:

```yaml
# hello-world-nodeport.yaml
apiVersion: v1
kind: Service
metadata:
  name: hello-world-nodeport
  labels:
    app: hello-world
spec:
  selector:
    app: hello-world
  ports:
  - port: 80
    targetPort: 80
    nodePort: 30080
    protocol: TCP
  type: NodePort
```

Apply the NodePort service:

```bash
kubectl apply -f hello-world-nodeport.yaml
```

Get the node IP and access the application:

```bash
# Get node IP
kubectl get nodes -o wide

# Access via node IP (replace <NODE_IP> with actual IP)
curl http://<NODE_IP>:30080
```

#### Method 3: Using LoadBalancer (Production)

For production access, you can use a LoadBalancer service:

```yaml
# hello-world-loadbalancer.yaml
apiVersion: v1
kind: Service
metadata:
  name: hello-world-lb
  labels:
    app: hello-world
spec:
  selector:
    app: hello-world
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
  type: LoadBalancer
```

### 5.4 Understanding the Connection

When you access your Hello World application, here's what happens:

1. **Your Browser** → Makes HTTP request to the service
2. **Kubernetes Service** → Routes traffic to the pod using Mycelium networking
3. **Mycelium CNI** → Handles IPv6 communication between nodes
4. **Pod** → Serves the HTML content via nginx
5. **Response** → Returns through the same encrypted Mycelium tunnel

The application demonstrates that your workload is running on the decentralized ThreeFold Grid and communicating through Mycelium's peer-to-peer network infrastructure.

## Step 6: Advanced Example - 3 Python Web Servers with Load Balancing

This example demonstrates how to deploy multiple Python web servers and use Mycelium Cloud's gateway functionality for load balancing.

### 6.1 Deploy Three Python Web Servers

Let's create three Python web servers, each identifying themselves:

```yaml
# python-servers-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: python-server-1
  labels:
    app: python-server
    server-id: "1"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: python-server
      server-id: "1"
  template:
    metadata:
      labels:
        app: python-server
        server-id: "1"
    spec:
      containers:
      - name: python-server
        image: python:3.9-slim
        command: ["python", "-c"]
        args:
        - |
          import http.server
          import socketserver
          import os
          import json
          from datetime import datetime

          class MyHandler(http.server.SimpleHTTPRequestHandler):
              def do_GET(self):
                  response = {
                      "server_id": "Python Server 1",
                      "message": "Hello from Python Server 1!",
                      "timestamp": datetime.now().isoformat(),
                      "pod_name": os.environ.get('HOSTNAME', 'unknown'),
                      "path": self.path
                  }

                  self.send_response(200)
                  self.send_header('Content-type', 'application/json')
                  self.send_header('Access-Control-Allow-Origin', '*')
                  self.end_headers()
                  self.wfile.write(json.dumps(response, indent=2).encode())

          PORT = 8000
          with socketserver.TCPServer(("", PORT), MyHandler) as httpd:
              print(f"Server 1 running on port {PORT}")
              httpd.serve_forever()
        ports:
        - containerPort: 8000
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: python-server-2
  labels:
    app: python-server
    server-id: "2"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: python-server
      server-id: "2"
  template:
    metadata:
      labels:
        app: python-server
        server-id: "2"
    spec:
      containers:
      - name: python-server
        image: python:3.9-slim
        command: ["python", "-c"]
        args:
        - |
          import http.server
          import socketserver
          import os
          import json
          from datetime import datetime

          class MyHandler(http.server.SimpleHTTPRequestHandler):
              def do_GET(self):
                  response = {
                      "server_id": "Python Server 2",
                      "message": "Hello from Python Server 2!",
                      "timestamp": datetime.now().isoformat(),
                      "pod_name": os.environ.get('HOSTNAME', 'unknown'),
                      "path": self.path
                  }

                  self.send_response(200)
                  self.send_header('Content-type', 'application/json')
                  self.send_header('Access-Control-Allow-Origin', '*')
                  self.end_headers()
                  self.wfile.write(json.dumps(response, indent=2).encode())

          PORT = 8000
          with socketserver.TCPServer(("", PORT), MyHandler) as httpd:
              print(f"Server 2 running on port {PORT}")
              httpd.serve_forever()
        ports:
        - containerPort: 8000
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: python-server-3
  labels:
    app: python-server
    server-id: "3"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: python-server
      server-id: "3"
  template:
    metadata:
      labels:
        app: python-server
        server-id: "3"
    spec:
      containers:
      - name: python-server
        image: python:3.9-slim
        command: ["python", "-c"]
        args:
        - |
          import http.server
          import socketserver
          import os
          import json
          from datetime import datetime

          class MyHandler(http.server.SimpleHTTPRequestHandler):
              def do_GET(self):
                  response = {
                      "server_id": "Python Server 3",
                      "message": "Hello from Python Server 3!",
                      "timestamp": datetime.now().isoformat(),
                      "pod_name": os.environ.get('HOSTNAME', 'unknown'),
                      "path": self.path
                  }

                  self.send_response(200)
                  self.send_header('Content-type', 'application/json')
                  self.send_header('Access-Control-Allow-Origin', '*')
                  self.end_headers()
                  self.wfile.write(json.dumps(response, indent=2).encode())

          PORT = 8000
          with socketserver.TCPServer(("", PORT), MyHandler) as httpd:
              print(f"Server 3 running on port {PORT}")
              httpd.serve_forever()
        ports:
        - containerPort: 8000
```

### 6.2 Create Services for Each Server

```yaml
# python-servers-services.yaml
apiVersion: v1
kind: Service
metadata:
  name: python-server-1-service
  labels:
    app: python-server
    server-id: "1"
spec:
  selector:
    app: python-server
    server-id: "1"
  ports:
  - port: 8000
    targetPort: 8000
    protocol: TCP
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: python-server-2-service
  labels:
    app: python-server
    server-id: "2"
spec:
  selector:
    app: python-server
    server-id: "2"
  ports:
  - port: 8000
    targetPort: 8000
    protocol: TCP
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: python-server-3-service
  labels:
    app: python-server
    server-id: "3"
spec:
  selector:
    app: python-server
    server-id: "3"
  ports:
  - port: 8000
    targetPort: 8000
    protocol: TCP
  type: ClusterIP
```

### 6.3 Deploy the Python Servers

```bash
# Deploy the servers
kubectl apply -f python-servers-deployment.yaml

# Deploy the services
kubectl apply -f python-servers-services.yaml

# Verify deployment
kubectl get pods -l app=python-server
kubectl get services -l app=python-server
```

### 6.4 Test Individual Servers

Test each server individually to ensure they're working:

```bash
# Test server 1
kubectl port-forward service/python-server-1-service 8001:8000 &
curl http://localhost:8001

# Test server 2
kubectl port-forward service/python-server-2-service 8002:8000 &
curl http://localhost:8002

# Test server 3
kubectl port-forward service/python-server-3-service 8003:8000 &
curl http://localhost:8003
```

### 6.5 Create a Load Balancer Service (Single WebGW)

Create a single load balancer that distributes traffic across all three servers:

```yaml
# python-servers-loadbalancer.yaml
apiVersion: v1
kind: Service
metadata:
  name: python-servers-lb
  labels:
    app: python-servers
spec:
  selector:
    app: python-server
  ports:
  - port: 80
    targetPort: 8000
    protocol: TCP
  type: LoadBalancer
```

Deploy and test:

```bash
kubectl apply -f python-servers-loadbalancer.yaml

# Get the external IP
kubectl get service python-servers-lb

# Test the load balancer (replace <EXTERNAL_IP> with actual IP)
curl http://<EXTERNAL_IP>
```

### 6.6 Advanced: Two WebGW Setup

For high availability and advanced load balancing, you can deploy two separate web gateways:

```yaml
# dual-webgw-setup.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webgw-1
  labels:
    app: webgw
    gateway-id: "1"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webgw
      gateway-id: "1"
  template:
    metadata:
      labels:
        app: webgw
        gateway-id: "1"
    spec:
      containers:
      - name: nginx-proxy
        image: nginx:1.21
        ports:
        - containerPort: 80
        volumeMounts:
        - name: nginx-config
          mountPath: /etc/nginx/conf.d
      volumes:
      - name: nginx-config
        configMap:
          name: webgw-1-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: webgw-1-config
data:
  default.conf: |
    upstream python_servers {
        server python-server-1-service:8000;
        server python-server-2-service:8000;
    }

    server {
        listen 80;
        location / {
            proxy_pass http://python_servers;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }
    }
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: webgw-2
  labels:
    app: webgw
    gateway-id: "2"
spec:
  replicas: 1
  selector:
    matchLabels:
      app: webgw
      gateway-id: "2"
  template:
    metadata:
      labels:
        app: webgw
        gateway-id: "2"
    spec:
      containers:
      - name: nginx-proxy
        image: nginx:1.21
        ports:
        - containerPort: 80
        volumeMounts:
        - name: nginx-config
          mountPath: /etc/nginx/conf.d
      volumes:
      - name: nginx-config
        configMap:
          name: webgw-2-config
---
apiVersion: v1
kind: ConfigMap
metadata:
  name: webgw-2-config
data:
  default.conf: |
    upstream python_servers {
        server python-server-2-service:8000;
        server python-server-3-service:8000;
    }

    server {
        listen 80;
        location / {
            proxy_pass http://python_servers;
            proxy_set_header Host $host;
            proxy_set_header X-Real-IP $remote_addr;
            proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        }
    }
---
apiVersion: v1
kind: Service
metadata:
  name: webgw-1-service
  labels:
    app: webgw
    gateway-id: "1"
spec:
  selector:
    app: webgw
    gateway-id: "1"
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
  type: LoadBalancer
---
apiVersion: v1
kind: Service
metadata:
  name: webgw-2-service
  labels:
    app: webgw
    gateway-id: "2"
spec:
  selector:
    app: webgw
    gateway-id: "2"
  ports:
  - port: 80
    targetPort: 80
    protocol: TCP
  type: LoadBalancer
```

Deploy the dual webgw setup:

```bash
kubectl apply -f dual-webgw-setup.yaml

# Check the services
kubectl get services -l app=webgw

# Test both gateways
kubectl get service webgw-1-service
kubectl get service webgw-2-service

# Test each gateway (replace with actual IPs)
curl http://<WEBGW1_IP>
curl http://<WEBGW2_IP>
```

### 6.7 Understanding the Two WebGW Architecture

With two web gateways, you get:

1. **Load Distribution**: Each gateway handles different servers
   - WebGW-1: Routes to Server 1 and Server 2
   - WebGW-2: Routes to Server 2 and Server 3

2. **High Availability**: If one gateway fails, the other continues serving

3. **Geographic Distribution**: Deploy gateways on different nodes for better performance

4. **Traffic Splitting**: Use DNS or external load balancer to distribute traffic between gateways

### 6.8 Monitoring the Setup

Monitor your multi-server setup:

```bash
# Check all pods
kubectl get pods -l app=python-server
kubectl get pods -l app=webgw

# Check services
kubectl get services

# Check logs
kubectl logs -l app=python-server
kubectl logs -l app=webgw

# Monitor resource usage
kubectl top pods -l app=python-server
kubectl top pods -l app=webgw
```

## Step 7: Service-to-Service Communication Over Mycelium

This example demonstrates how three services running on different pods across remote machines can communicate with each other using Mycelium's peer-to-peer networking.

### 7.1 Deploy Three Microservices

Let's create three microservices that communicate with each other:

```yaml
# microservices-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend-service
  labels:
    app: microservice
    service: frontend
spec:
  replicas: 1
  selector:
    matchLabels:
      app: microservice
      service: frontend
  template:
    metadata:
      labels:
        app: microservice
        service: frontend
    spec:
      containers:
      - name: frontend
        image: python:3.9-slim
        command: ["python", "-c"]
        args:
        - |
          import http.server
          import socketserver
          import urllib.request
          import json
          import os
          from datetime import datetime

          class FrontendHandler(http.server.SimpleHTTPRequestHandler):
              def do_GET(self):
                  if self.path == '/':
                      # Call backend service
                      try:
                          backend_response = urllib.request.urlopen('http://backend-service:8001/api/data')
                          backend_data = json.loads(backend_response.read().decode())
                      except Exception as e:
                          backend_data = {"error": str(e)}

                      # Call database service
                      try:
                          db_response = urllib.request.urlopen('http://database-service:8002/api/stats')
                          db_data = json.loads(db_response.read().decode())
                      except Exception as e:
                          db_data = {"error": str(e)}

                      response = {
                          "service": "Frontend Service",
                          "message": "Hello from Frontend!",
                          "timestamp": datetime.now().isoformat(),
                          "pod_name": os.environ.get('HOSTNAME', 'unknown'),
                          "backend_data": backend_data,
                          "database_data": db_data,
                          "communication_method": "Mycelium P2P IPv6"
                      }

                      self.send_response(200)
                      self.send_header('Content-type', 'application/json')
                      self.send_header('Access-Control-Allow-Origin', '*')
                      self.end_headers()
                      self.wfile.write(json.dumps(response, indent=2).encode())
                  else:
                      self.send_response(404)
                      self.end_headers()

          PORT = 8000
          with socketserver.TCPServer(("", PORT), FrontendHandler) as httpd:
              print(f"Frontend service running on port {PORT}")
              httpd.serve_forever()
        ports:
        - containerPort: 8000
```

### 7.2 Understanding Mycelium Communication

**Yes, this communication goes over Mycelium!** Here's how:

1. **Pod-to-Pod Communication**: When `frontend-service` calls `backend-service:8001`, the request goes through Mycelium's peer-to-peer network
2. **IPv6 Addressing**: Each pod gets a unique IPv6 address from the Mycelium network range
3. **Encrypted Tunnels**: All communication between pods is encrypted using Mycelium's security protocols
4. **Cross-Node Communication**: If pods are on different nodes, communication still works seamlessly through Mycelium

**Communication Flow:**

```text
Frontend Pod → Mycelium CNI → IPv6 Address → Mycelium Bridge → Encrypted Tunnel → Mycelium Bridge → IPv6 Address → Mycelium CNI → Backend Pod
```

## Step 8: Monitoring and Management

### 8.1 Cluster Monitoring

Access built-in monitoring:

1. **Mycelium Cloud Dashboard**: View cluster metrics and status
2. **Grafana**: Access via the monitoring section (if enabled)
3. **Prometheus**: Query metrics directly

### 8.2 Resource Monitoring

Monitor resource usage:

```bash
# Node resource usage
kubectl top nodes

# Pod resource usage
kubectl top pods --all-namespaces

# Describe node details
kubectl describe node <node-name>
```

### 8.3 Scaling Your Cluster

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

## Step 9: Best Practices

### 9.1 Security

- **RBAC**: Implement Role-Based Access Control
- **Network Policies**: Control pod-to-pod communication
- **Secrets Management**: Use Kubernetes secrets for sensitive data
- **Regular Updates**: Keep cluster and applications updated

### 9.2 Resource Management

- **Resource Limits**: Set CPU and memory limits for pods
- **Namespaces**: Organize applications using namespaces
- **Storage Classes**: Use appropriate storage for different workloads

### 9.3 Backup and Disaster Recovery

- **etcd Backups**: Regular cluster state backups
- **Application Data**: Backup persistent volumes
- **Configuration**: Version control your YAML manifests

## Step 10: Troubleshooting

### 10.1 Common Issues

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

### 10.2 Getting Help

- **Logs**: Check cluster and application logs
- **Events**: Monitor Kubernetes events
- **Support**: Contact Mycelium Cloud support
- **Community**: Join our community channels
- **Issues**: Open a ticket on GitHub with details: [https://github.com/codescalers/kubecloud/issues](https://github.com/codescalers/kubecloud/issues)

## Next Steps

Congratulations! You've successfully deployed and managed your first Kubernetes cluster on Mycelium Cloud. Here's what to explore next:

1. **Advanced Networking**: Configure ingress controllers and load balancers
2. **Storage Solutions**: Implement persistent storage for stateful applications
3. **CI/CD Integration**: Set up automated deployment pipelines
4. **Monitoring**: Advanced monitoring with custom metrics
5. **Security**: Implement advanced security policies

## Additional Resources

- [API Reference](https://github.com/codescalers/kubecloud/blob/master/frontend/kubecloud/public/docs/api-reference.md) - Complete API documentation
- [FAQ](https://github.com/codescalers/kubecloud/blob/master/frontend/kubecloud/public/docs/faq.md) - Frequently asked questions
- [Architecture Guide](https://github.com/codescalers/kubecloud/blob/master/frontend/kubecloud/public/docs/architecture.md) - Deep dive into platform architecture
- [Best Practices](https://github.com/codescalers/kubecloud/blob/master/frontend/kubecloud/public/docs/best-practices.md) - Production deployment guidelines
