# Tutorials

This tutorial covers advanced deployment scenarios for Mycelium Cloud. For basic setup, see the [Getting Started Guide](#getting-started).

## Prerequisites

- Completed [Getting Started Guide](#getting-started)
- Deployed cluster with kubectl access
- Mycelium binary installed

### Setting up kubectl access

```bash
# Set the kubeconfig for this session
export KUBECONFIG=/path/to/your/config

# Verify cluster access
kubectl get nodes
```

> **Note**: All examples assume you have kubectl access configured. If not, run the commands above first.

## Example 1: Hello World

### Step 1 — Create the Deployment (save as `hello-world-deploy.yaml`)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-world
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
```

### Step 2 — Expose it with a Service (save as `hello-world-svc.yaml`)

```yaml
apiVersion: v1
kind: Service
metadata:
  name: hello-world-service
spec:
  selector:
    app: hello-world
  ports:
  - port: 80
    targetPort: 80
  type: ClusterIP
```

### Step 3 — Apply and test

```bash
kubectl apply -f hello-world-deploy.yaml
kubectl apply -f hello-world-svc.yaml
kubectl port-forward service/hello-world-service 8080:80
```

Open `http://localhost:8080` — you should see the Nginx welcome page.

## Example 2: 3 Python Servers with Load Balancing

### Step 1 — Deploy three simple Python HTTP servers

Save each as its own file and apply them (or combine into one file if you prefer).

```yaml
# python-server-1.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: python-server-1
spec:
  replicas: 2
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
          import http.server, socketserver, json
          class Handler(http.server.SimpleHTTPRequestHandler):
              def do_GET(self):
                  self.send_response(200)
                  self.send_header('Content-type', 'application/json')
                  self.end_headers()
                  response = {"server": "Python Server 1"}
                  self.wfile.write(json.dumps(response).encode())
          with socketserver.TCPServer(("", 8000), Handler) as httpd:
              httpd.serve_forever()
# python-server-2.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: python-server-2
spec:
  replicas: 2
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
          import http.server, socketserver, json
          class Handler(http.server.SimpleHTTPRequestHandler):
              def do_GET(self):
                  self.send_response(200)
                  self.send_header('Content-type', 'application/json')
                  self.end_headers()
                  response = {"server": "Python Server 2"}
                  self.wfile.write(json.dumps(response).encode())
          with socketserver.TCPServer(("", 8000), Handler) as httpd:
              httpd.serve_forever()
        ports:
        - containerPort: 8000

# python-server-3.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: python-server-3
spec:
  replicas: 2
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
          import http.server, socketserver, json
          class Handler(http.server.SimpleHTTPRequestHandler):
              def do_GET(self):
                  self.send_response(200)
                  self.send_header('Content-type', 'application/json')
                  self.end_headers()
                  response = {"server": "Python Server 3"}
                  self.wfile.write(json.dumps(response).encode())
          with socketserver.TCPServer(("", 8000), Handler) as httpd:
              httpd.serve_forever()
        ports:
        - containerPort: 8000
```

### Step 2 — Create a load balancer Service (save as `python-servers-lb.yaml`)

```yaml
apiVersion: v1
kind: Service
metadata:
  name: python-servers-lb
spec:
  selector:
    app: python-server
  ports:
  - port: 80
    targetPort: 8000
  type: LoadBalancer
```

### Step 3 — Apply and test (Python Servers)

```bash
kubectl apply -f python-server-1.yaml
kubectl apply -f python-server-2.yaml
kubectl apply -f python-server-3.yaml
kubectl apply -f python-servers-lb.yaml
```

**Test load balancing:**

```bash
# Option A: In-cluster test (recommended)
kubectl run tmp-curl --rm -it --image=curlimages/curl:8.7.1 --restart=Never -- \
  sh -lc 'for i in $(seq 1 6); do curl -s http://python-servers-lb; echo; done'

# Option B: Multiple port-forwards
kubectl port-forward svc/python-servers-lb 8082:80 &
kubectl port-forward svc/python-servers-lb 8083:80 &
# Open http://localhost:8082 and http://localhost:8083 in separate tabs
```

### Two WebGW Setup for High Availability

For high availability and load distribution, you can deploy two separate web gateways that route to different subsets of your Python servers:

```yaml
# webgw-1.yaml - Routes to Python Server 1
apiVersion: v1
kind: Service
metadata:
  name: webgw-1
spec:
  selector:
    app: python-server
    server-id: "1"  # Only routes to Python Server 1
  ports:
  - port: 80
    targetPort: 8000
  type: LoadBalancer
---
# webgw-2.yaml - Routes to Python Servers 2 & 3
apiVersion: v1
kind: Service
metadata:
  name: webgw-2
spec:
  selector:
    app: python-server
    server-id: "2"  # Routes to Python Server 2
  ports:
  - port: 80
    targetPort: 8000
  type: LoadBalancer
---
# webgw-3.yaml - Routes to Python Server 3
apiVersion: v1
kind: Service
metadata:
  name: webgw-3
spec:
  selector:
    app: python-server
    server-id: "3"  # Only routes to Python Server 3
  ports:
  - port: 80
    targetPort: 8000
  type: LoadBalancer
```

#### Deploy and Test Multiple WebGWs

```bash
kubectl apply -f webgw-1.yaml
kubectl apply -f webgw-2.yaml
kubectl apply -f webgw-3.yaml

# Test each gateway
kubectl port-forward svc/webgw-1 8084:80 &
kubectl port-forward svc/webgw-2 8085:80 &
kubectl port-forward svc/webgw-3 8086:80 &

curl -s http://localhost:8084   # Returns "Python Server 1"
curl -s http://localhost:8085   # Returns "Python Server 2"  
curl -s http://localhost:8086   # Returns "Python Server 3"
```

#### Benefits of Multiple WebGWs

- **High Availability**: If one gateway fails, others continue serving traffic
- **Load Distribution**: Traffic can be distributed across different gateways
- **Service Isolation**: Each gateway can route to specific services
- **Geographic Distribution**: Gateways can be deployed on different nodes for better performance

This setup allows for more granular control over traffic routing.

### What Happens When Servers Go Down?

**Current Behavior:**

- If one server goes down, the load balancer will automatically route traffic to the remaining healthy servers
- Kubernetes health checks will detect the failed pod and remove it from the service endpoints
- Traffic continues to flow to the remaining 2 servers without interruption

**Testing Server Failure:**

```bash
# Simulate a server failure by deleting one pod
kubectl delete pod -l server-id=1

# Check that traffic still works
kubectl port-forward svc/python-servers-lb 8080:80
curl http://localhost:8080  # Should still work with remaining servers
```

**Monitoring Server Health:**

```bash
# Check pod status
kubectl get pods -l app=python-server

# Check service endpoints
kubectl get endpoints python-servers-lb

# View pod logs if issues occur
kubectl logs -l server-id=1
```

**Resilience:**
With 2 replicas per server type, if one pod fails, there's still another pod of the same server type running, providing better fault tolerance.

## Example 3: Service-to-Service Communication Over Mycelium

This example shows how three services communicate across different nodes using Mycelium networking.

**Yes, services communicate over Mycelium!** Services on different nodes use Mycelium IPs for communication, while services on the same node use standard Kubernetes networking.

### Step 1 — Frontend service (save as `frontend.yaml`)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: frontend-service
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
          import http.server, socketserver, urllib.request, json
          class Handler(http.server.SimpleHTTPRequestHandler):
              def do_GET(self):
                  try:
                      backend = urllib.request.urlopen('http://backend-service:8001/api/data')
                      db = urllib.request.urlopen('http://database-service:8002/api/stats')
                      response = {
                          "service": "Frontend",
                          "backend": json.loads(backend.read().decode()),
                          "database": json.loads(db.read().decode())
                      }
                  except Exception as e:
                      response = {"error": str(e)}
                  self.send_response(200)
                  self.send_header('Content-type', 'application/json')
                  self.end_headers()
                  self.wfile.write(json.dumps(response).encode())
          with socketserver.TCPServer(("", 8000), Handler) as httpd:
              httpd.serve_forever()
        ports:
        - containerPort: 8000
```

### Step 2 — Backend service (save as `backend.yaml`)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: backend-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: microservice
      service: backend
  template:
    metadata:
      labels:
        app: microservice
        service: backend
    spec:
      containers:
      - name: backend
        image: python:3.9-slim
        command: ["python", "-c"]
        args:
        - |
          import http.server, socketserver, json
          class Handler(http.server.SimpleHTTPRequestHandler):
              def do_GET(self):
                  self.send_response(200)
                  self.send_header('Content-type', 'application/json')
                  self.end_headers()
                  response = {"service": "Backend", "data": "Hello from backend!"}
                  self.wfile.write(json.dumps(response).encode())
          with socketserver.TCPServer(("", 8001), Handler) as httpd:
              httpd.serve_forever()
        ports:
        - containerPort: 8001
```

### Step 3 — Database service (save as `database.yaml`)

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: database-service
spec:
  replicas: 1
  selector:
    matchLabels:
      app: microservice
      service: database
  template:
    metadata:
      labels:
        app: microservice
        service: database
    spec:
      containers:
      - name: database
        image: python:3.9-slim
        command: ["python", "-c"]
        args:
        - |
          import http.server, socketserver, json
          class Handler(http.server.SimpleHTTPRequestHandler):
              def do_GET(self):
                  self.send_response(200)
                  self.send_header('Content-type', 'application/json')
                  self.end_headers()
                  response = {"service": "Database", "stats": "Connected!"}
                  self.wfile.write(json.dumps(response).encode())
          with socketserver.TCPServer(("", 8002), Handler) as httpd:
              httpd.serve_forever()
        ports:
        - containerPort: 8002
```

### Step 4 — Services (save as `microservices-svc.yaml`)

```yaml
apiVersion: v1
kind: Service
metadata:
  name: frontend-service
spec:
  selector:
    app: microservice
    service: frontend
  ports:
  - port: 8000
    targetPort: 8000
  type: LoadBalancer
---
apiVersion: v1
kind: Service
metadata:
  name: backend-service
spec:
  selector:
    app: microservice
    service: backend
  ports:
  - port: 8001
    targetPort: 8001
  type: ClusterIP
---
apiVersion: v1
kind: Service
metadata:
  name: database-service
spec:
  selector:
    app: microservice
    service: database
  ports:
  - port: 8002
    targetPort: 8002
  type: ClusterIP
```

### Step 5 — Apply and test

```bash
kubectl apply -f frontend.yaml
kubectl apply -f backend.yaml
kubectl apply -f database.yaml
kubectl apply -f microservices-svc.yaml
kubectl port-forward service/frontend-service 8080:8000
```

Open `http://localhost:8080` — the frontend should call the backend and database over Mycelium networking.

### Step 6 — Verify Communication

```bash
# Check which nodes pods are running on
kubectl get pods -o wide

# Test service communication
kubectl exec -it deployment/frontend-service -- curl http://backend-service:8001/api/data
kubectl exec -it deployment/backend-service -- curl http://database-service:8002/api/stats
```

> Note: If you see an error like `exec: "curl": not found`, the base image (e.g., `python:3.9-slim`) doesn't include curl. Install it inside the running container, then rerun the curl command command:
>
> ```bash
> kubectl exec -it deployment/frontend-service -- sh -lc 'apt-get update && apt-get install -y curl'
> kubectl exec -it deployment/backend-service -- sh -lc 'apt-get update && apt-get install -y curl'
> ```

This demonstrates how services communicate across nodes using Mycelium networking.

## Example 4: Working with TF Gateway as CRDs

Custom Resource Definitions (CRDs) allow you to extend Kubernetes with your own custom resources. In this example, we'll work with the TFGW (ThreeFold Gateway) CRD, which manages load balancing and proxying by taking a hostname and backend services.

### Understanding the TFGW CRD

The TFGW CRD has the following structure:

**Spec:**
- `hostname`: Name of your subdomain (should be alphanumeric. example `example`) this generates the domain name `example.gent01.grid.tf`
- `backends`: List of backend service full URLs to route traffic to. (should be accessible from mycelium net)

**Status:**
- `fqdn`: The fully qualified domain name assigned
- `message`: Status message

### Step 1 — Create Backend Services for the Gateway

Let's create some backend services that our TFGW will route traffic to:

```yaml
# server-example.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: hello-world
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
---
apiVersion: v1
kind: Service
metadata:
  name: my-service
spec:
  selector:
    app: hello-world
  ports:
    - port: 80           # Service port
      targetPort: 80     # Container port
```

### Step 2 — Create a TFGW Custom Resource

Now let's create a TFGW resource that routes traffic to our backend services:

```yaml
# tfgw-example.yaml
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
    - "http://[5ce:20f3:1d33:d235:ff0f:b265:334f:e240]:80" # http://<node_id>:<service-port>
```

### Step 3 — Apply and Test the CRD

```bash
# Apply the backend services
kubectl apply -f backend-services.yaml

# Apply the TFGW custom resource
kubectl apply -f tfgw-example.yaml

# Check the TFGW resource
kubectl get tfgw

# Get detailed information about the TFGW
kubectl describe tfgw my-web-gateway
```

### Step 4 - Apply ingress

update the example with the generated FQDN from CRD

```yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: my-ingress
spec:
  rules:
  - host: omar.gent02.dev.grid.tf
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: my-service
            port:
              number: 80
```

Now you can access your service https://omar.gent02.dev.grid.tf

### Monitoring TFGW Status and Deleting

```bash
# Watch for status changes
kubectl get tfgw -w

# Check status conditions
kubectl get tfgw my-web-gateway -o jsonpath='{.status.conditions[*]}'

# Get the assigned FQDN
kubectl get tfgw my-web-gateway -o jsonpath='{.status.fqdn}'

# Delete a specific TFGW
kubectl delete tfgw my-web-gateway

# Delete all TFGW resources
kubectl delete tfgw --all
```
