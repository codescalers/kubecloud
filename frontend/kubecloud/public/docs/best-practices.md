# Best Practices

This guide covers best practices for deploying and managing production workloads on Mycelium Cloud.

## Cluster Design

### High Availability Setup

For production environments, always deploy with high availability:

```yaml
# Recommended HA configuration
Masters: 3 nodes (odd number for etcd quorum)
Workers: 3+ nodes (for workload distribution)
Regions: Multiple regions for disaster recovery
```

### Resource Planning

**CPU and Memory:**

- Reserve 20% capacity for system overhead
- Use resource requests and limits for all pods
- Monitor actual usage vs. allocated resources

**Storage:**

- Plan for data growth (3x current usage)
- Use appropriate storage classes
- Implement backup strategies

### Network Design

**Security:**

- Implement network policies
- Use service mesh for complex applications
- Secure ingress with TLS certificates

**Performance:**

- Co-locate related services
- Use node affinity for latency-sensitive workloads
- Implement proper load balancing

## Security Best Practices

### Access Control

**RBAC Implementation:**

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: Role
metadata:
  name: developer
rules:
- apiGroups: [""]
  resources: ["pods", "services"]
  verbs: ["get", "list", "create", "update", "patch", "delete"]
```

**Service Accounts:**

- Create dedicated service accounts for applications
- Use least privilege principle
- Rotate service account tokens regularly

### Secrets Management

**Best Practices:**

- Never store secrets in container images
- Use Kubernetes secrets or external secret management
- Encrypt secrets at rest
- Implement secret rotation

**Example Secret Usage:**

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
type: Opaque
data:
  database-password: <base64-encoded-password>
---
apiVersion: apps/v1
kind: Deployment
metadata:
  name: app
spec:
  template:
    spec:
      containers:
      - name: app
        env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: database-password
```

### Pod Security

**Security Contexts:**

```yaml
apiVersion: v1
kind: Pod
spec:
  securityContext:
    runAsNonRoot: true
    runAsUser: 1000
    fsGroup: 2000
  containers:
  - name: app
    securityContext:
      allowPrivilegeEscalation: false
      readOnlyRootFilesystem: true
      capabilities:
        drop:
        - ALL
```

## Resource Management

### Resource Requests and Limits

**CPU and Memory:**

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

**Guidelines:**

- Set requests based on actual usage
- Set limits to prevent resource starvation
- Use vertical pod autoscaler for optimization

### Quality of Service Classes

**Guaranteed (highest priority):**

```yaml
resources:
  requests:
    memory: "256Mi"
    cpu: "250m"
  limits:
    memory: "256Mi"
    cpu: "250m"
```

**Burstable (medium priority):**

```yaml
resources:
  requests:
    memory: "128Mi"
    cpu: "125m"
  limits:
    memory: "512Mi"
    cpu: "500m"
```

### Horizontal Pod Autoscaling

```yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: app-hpa
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: app
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
```

## Application Deployment

### Deployment Strategies

**Rolling Updates:**

```yaml
spec:
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
```

**Blue-Green Deployment:**

- Deploy new version alongside old
- Switch traffic after validation
- Keep old version for quick rollback

### Health Checks

**Liveness Probe:**

```yaml
livenessProbe:
  httpGet:
    path: /health
    port: 8080
  initialDelaySeconds: 30
  periodSeconds: 10
  timeoutSeconds: 5
  failureThreshold: 3
```

**Readiness Probe:**

```yaml
readinessProbe:
  httpGet:
    path: /ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 3
  failureThreshold: 3
```

### Configuration Management

**ConfigMaps:**

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
data:
  database-url: "postgres://db:5432/myapp"
  log-level: "info"
```

**Environment-specific Configurations:**

- Use separate ConfigMaps per environment
- Implement configuration validation
- Version control all configurations

## Monitoring and Observability

### Metrics Collection

**Application Metrics:**

- Expose metrics in Prometheus format
- Include business metrics
- Monitor SLIs (Service Level Indicators)

**Custom Metrics Example:**

```go
// Go application metrics
var (
    requestsTotal = prometheus.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )
)
```

### Logging Best Practices

**Structured Logging:**

```json
{
  "timestamp": "2023-12-01T10:00:00Z",
  "level": "info",
  "service": "user-service",
  "trace_id": "abc123",
  "message": "User created successfully",
  "user_id": "12345"
}
```

**Log Levels:**

- ERROR: System errors requiring attention
- WARN: Potential issues or degraded performance
- INFO: General application flow
- DEBUG: Detailed diagnostic information

### Alerting Strategy

**SLO-based Alerting:**

- Define Service Level Objectives (SLOs)
- Alert on SLO violations
- Implement error budgets

**Alert Examples:**

```yaml
# High error rate
- alert: HighErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) > 0.1
  for: 5m
  annotations:
    summary: High error rate detected

# Pod crash looping
- alert: PodCrashLooping
  expr: rate(kube_pod_container_status_restarts_total[15m]) > 0
  for: 5m
  annotations:
    summary: Pod is crash looping
```

## Storage Best Practices

### Persistent Volumes

**Storage Classes:**

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
  name: fast-ssd
provisioner: kubernetes.io/no-provisioner
parameters:
  type: ssd
  replication-type: none
```

**Volume Claims:**

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
  name: app-storage
spec:
  accessModes:
    - ReadWriteOnce
  storageClassName: fast-ssd
  resources:
    requests:
      storage: 10Gi
```

### Backup Strategies

**Database Backups:**

- Automated daily backups
- Point-in-time recovery capability
- Cross-region backup replication
- Regular restore testing

**Application Data:**

- Volume snapshots
- Application-consistent backups
- Backup retention policies

## CI/CD Integration

### GitOps Workflow

**Repository Structure:**

```text
├── applications/
│   ├── staging/
│   └── production/
├── infrastructure/
│   ├── base/
│   └── overlays/
└── scripts/
    ├── deploy.sh
    └── rollback.sh
```

**Deployment Pipeline:**

1. Code commit triggers pipeline
2. Build and test application
3. Build container image
4. Update Kubernetes manifests
5. Deploy to staging environment
6. Run integration tests
7. Deploy to production (with approval)

### Testing Strategies

**Unit Tests:**

- Test individual components
- Mock external dependencies
- Achieve high code coverage

**Integration Tests:**

- Test service interactions
- Use test databases/services
- Validate API contracts

**End-to-End Tests:**

- Test complete user workflows
- Run against staging environment
- Automate critical user paths

## Performance Optimization

### Application Performance

**Resource Optimization:**

- Profile CPU and memory usage
- Optimize database queries
- Implement caching strategies
- Use connection pooling

**Scaling Strategies:**

- Horizontal scaling for stateless services
- Vertical scaling for resource-intensive tasks
- Implement circuit breakers
- Use async processing for heavy workloads

### Network Performance

**Service Mesh:**

- Implement Istio or Linkerd
- Use traffic splitting for deployments
- Monitor service-to-service latency
- Implement retry and timeout policies

**Load Balancing:**

- Use appropriate load balancing algorithms
- Implement health checks
- Configure session affinity when needed

## Disaster Recovery

### Backup and Recovery

**Cluster Backup:**

- Regular etcd backups
- Backup all custom resources
- Document recovery procedures
- Test recovery processes regularly

**Data Recovery:**

- Implement point-in-time recovery
- Cross-region data replication
- Automated failover mechanisms
- Recovery time objectives (RTO) planning

### Multi-Region Deployment

**Active-Active Setup:**

- Deploy across multiple regions
- Implement global load balancing
- Synchronize data between regions
- Handle split-brain scenarios

**Active-Passive Setup:**

- Primary region for active traffic
- Standby region for disaster recovery
- Automated failover triggers
- Data replication strategies

## Cost Optimization

### Resource Right-Sizing

**Monitoring Usage:**

- Track actual vs. requested resources
- Identify over-provisioned workloads
- Use cluster autoscaler
- Implement resource quotas

**Cost Analysis:**

- Monitor spending by namespace/team
- Implement chargeback mechanisms
- Use spot instances for batch workloads
- Schedule non-critical workloads

### Efficiency Improvements

**Pod Density:**

- Optimize node utilization
- Use smaller container images
- Implement resource sharing
- Consolidate similar workloads

## Compliance and Governance

### Policy Management

**Pod Security Policies:**

```yaml
apiVersion: policy/v1beta1
kind: PodSecurityPolicy
metadata:
  name: restricted
spec:
  privileged: false
  allowPrivilegeEscalation: false
  requiredDropCapabilities:
    - ALL
  runAsUser:
    rule: 'MustRunAsNonRoot'
```

**Network Policies:**

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: deny-all
spec:
  podSelector: {}
  policyTypes:
  - Ingress
  - Egress
```

### Audit and Compliance

**Audit Logging:**

- Enable Kubernetes audit logging
- Monitor privileged operations
- Track resource access patterns
- Implement log retention policies

**Compliance Frameworks:**

- Implement SOC 2 controls
- Follow CIS Kubernetes benchmarks
- Ensure GDPR compliance for data handling
- Regular security assessments

## Troubleshooting

### Common Issues

**Pod Startup Problems:**

```bash
# Check pod events
kubectl describe pod <pod-name>

# Check logs
kubectl logs <pod-name> --previous

# Check resource constraints
kubectl top pods
kubectl describe node <node-name>
```

**Network Issues:**

```bash
# Test connectivity
kubectl exec -it <pod-name> -- nslookup <service-name>

# Check endpoints
kubectl get endpoints <service-name>

# Verify network policies
kubectl describe networkpolicy
```

### Performance Debugging

**Resource Analysis:**

```bash
# Node resource usage
kubectl top nodes

# Pod resource usage
kubectl top pods --all-namespaces

# Detailed resource info
kubectl describe node <node-name>
```

**Application Debugging:**

```bash
# Port forwarding for local access
kubectl port-forward pod/<pod-name> 8080:80

# Execute commands in pod
kubectl exec -it <pod-name> -- /bin/bash

# Copy files from pod
kubectl cp <pod-name>:/path/to/file ./local-file
```

## Conclusion

Following these best practices will help ensure your Mycelium Cloud deployments are secure, reliable, and performant. Remember to:

- Start with security in mind
- Monitor everything
- Automate repetitive tasks
- Plan for failures
- Continuously optimize

For more specific guidance, refer to:

- [Architecture Overview](./architecture.md)
- [API Reference](./api-reference.md)
- [FAQ](./faq.md)
