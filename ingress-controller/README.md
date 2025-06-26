# ThreeFold Gateway Ingress Controller

A custom Kubernetes ingress controller that automatically creates ThreeFold Gateway workloads as reverse proxies for your services.

## What it does

- Watches for changes in Kubernetes Ingress, Service, and EndpointSlice resources
- Automatically creates ThreeFold Gateway workloads that act as reverse proxies
- Routes external traffic to your cluster services through the ThreeFold network

## How it works

The controller monitors your cluster and when it detects ingress configurations, it deploys a ThreeFold Gateway workload that serves as a reverse proxy, routing traffic from the ThreeFold network to your backend services.

## Usage

### 1. Configure credentials

Update the mnemonic and network in `ingress.yaml`:

```yaml
stringData:
  mnemonic: "your-threefold-mnemonic-phrase-here"
```

```yaml
env:
- name: THREEFOLD_NETWORK
  value: "dev"  # or "mainnet"
```

### 2. Deploy to your cluster

```bash
# Deploy the controller
kubectl apply -f ingress.yaml

# Deploy example application (optional)
kubectl apply -f service.yaml
```

### 3. Access your service

Your service will be accessible through the ThreeFold Gateway at the configured hostname.

## Files

- `ingress.yaml` - Controller deployment and ingress configuration
- `service.yaml` - Example application for testing
- `main.go` - Controller source code
