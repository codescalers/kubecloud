# TFGW Kubernetes CRD

A Kubernetes Custom Resource Definition (CRD) built with Kubebuilder that manages ThreeFold Gateway resources.

## What is TFGW?

TFGW (ThreeFold Gateway) is a custom resource that manages load balancing and proxying. It takes a hostname and a list of backend services, then creates a gateway that routes traffic from the hostname to the specified backends.

**Spec:**
- `hostname`: The domain/hostname for the gateway
- `backends`: List of backend service URLs to route traffic to

**Status:**
- `fqdn`: The fully qualified domain name assigned
- `message`: Status message

## Development

1. **Update code** (types/controller):

    Edit api/v1/tfgw_types.go or internal/controller/tfgw_controller.go

2. **Build and push image**:
   ```bash
   make docker-build docker-push IMG=<registry>/crd:tag
   ```

3. **Generate manifests**:
   ```bash
   make generate manifests
   ```

4. **Install CRD to cluster**:
   ```bash
   make install # only crd install to see the changes
   ```

## Installation & Usage

### Install CRDs and Controller

Update controller environment variables in `config/manager/manager.yaml` before deploying.

```bash
make deploy IMG=<registry>/crd:tag # install crd and controller
```

This installs all CRDs and starts the controller.

### Apply Custom Resource

```bash
kubectl apply -f config/samples/ingress_v1_tfgw.yaml
```

### Check Resources

```bash
# Get TFGW resources
kubectl get tfgw

# Describe specific resource
kubectl describe tfgw tfgw-sample
```

## Cleanup

```bash
# Delete custom resources. wait untill it cleanup
kubectl delete -k config/samples/

# Uninstall CRDs
make uninstall

# Remove controller
make undeploy
```
