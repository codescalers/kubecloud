# Architecture Overview

This document provides a comprehensive overview of Mycelium Cloud's architecture, components, and design principles.

## System Architecture

Mycelium Cloud is built as a cloud-native platform with multiple interconnected components:

```text
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend UI   │    │   Backend APIs  │    │   ThreeFold     │
│    (Vue.js)     │───▶│      (Go)       │◀──▶│     Grid        │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                       │                       │
         │                       │                       │
         ▼                       ▼                       ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Monitoring    │    │   CRDs &        │    │   Networking    │
│(Prometheus/Graf)│    │ Controllers     │    │  (Mycelium)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Core Components

### 1. Frontend Application (Vue.js)

**Technology Stack:**

- Vue.js 3 with Composition API
- TypeScript for type safety
- Vuetify for UI components
- Pinia for state management
- Vue Router for navigation

**Key Features:**

- Responsive web interface
- Real-time cluster monitoring
- Interactive deployment wizard
- User dashboard and admin panel
- Billing and payment integration

**Architecture:**

```text
src/
├── components/          # Reusable UI components
│   ├── dashboard/      # Dashboard-specific components
│   ├── deploy/         # Deployment wizard components
│   └── ui/             # Generic UI components
├── views/              # Page-level components
├── stores/             # Pinia state stores
├── router/             # Vue Router configuration
└── services/           # API service layer
```

### 2. Backend Services (Go)

**Technology Stack:**

- Go 1.19+ with Gin web framework
- SQLite/PostgreSQL for data persistence
- Redis for caching and sessions
- JWT for authentication
- Swagger for API documentation

**Core Services:**

- **Authentication Service**: User management and JWT tokens
- **Cluster Service**: Kubernetes cluster lifecycle management
- **Billing Service**: Payment processing and usage tracking
- **Node Service**: ThreeFold Grid node management
- **Notification Service**: Real-time alerts and messaging

**Architecture:**

```text
backend/
├── app/                # HTTP handlers and business logic
├── internal/           # Internal packages and utilities
├── models/             # Database models and schemas
├── middlewares/        # HTTP middlewares
├── docs/               # API documentation (Swagger)
└── cmd/                # CLI commands and entry points
```

### 3. Kubernetes Integration (K3s)

**K3s Features:**

- Lightweight Kubernetes distribution
- High availability support
- Dual-stack networking (IPv4/IPv6)
- Embedded etcd or external datastore
- Simplified installation and management

**Custom Enhancements:**

- ThreeFold Grid integration
- Mycelium networking support
- Automated cluster provisioning
- Health monitoring and auto-healing

**Container Architecture:**

```text
K3s Container:
├── zinit/              # Process manager
├── scripts/            # Setup and configuration scripts
├── rootfs/             # Custom root filesystem
└── manifests/          # Kubernetes manifests
```

### 4. Networking Layer (Mycelium CNI)

**Mycelium Network:**

- Peer-to-peer IPv6 overlay network
- End-to-end encryption
- Automatic routing and discovery
- NAT traversal capabilities

**CNI Plugin Features:**

- IPv6 address assignment
- Virtual ethernet pair creation
- Bridge network integration
- Route configuration

**Network Flow:**

```text
Pod ←→ veth pair ←→ mycelium-br ←→ Mycelium daemon ←→ Peer network
```

**How Workloads Connect Over Mycelium:**

Mycelium Cloud uses a sophisticated peer-to-peer networking layer that enables secure, encrypted communication between workloads across the ThreeFold Grid. Here's how it works:

1. **Pod Network Assignment**: Each pod gets a unique IPv6 address from the Mycelium network range
2. **Virtual Network Interface**: The Mycelium CNI creates a virtual ethernet pair (veth) for each pod
3. **Bridge Integration**: Pods are connected to a mycelium bridge that handles routing
4. **Peer Discovery**: The Mycelium daemon automatically discovers and connects to other nodes
5. **Encrypted Tunnels**: All communication between pods is encrypted using Mycelium's security protocols

**Workload Communication Flow:**

```text
Workload A (Pod) → Mycelium CNI → IPv6 Address → Mycelium Bridge → Encrypted Tunnel → Mycelium Bridge → IPv6 Address → Mycelium CNI → Workload B (Pod)
```

**Key Benefits:**

- **Decentralized**: No central routing authority required
- **Secure**: End-to-end encryption for all communications
- **Scalable**: Automatic peer discovery and connection management
- **Resilient**: Direct peer-to-peer connections reduce single points of failure

**Technical Implementation Details:**

The Mycelium CNI plugin performs the following operations when a pod is created:

1. **Network Seed Reading**: Reads the network seed from `/etc/netseed` to ensure consistent addressing
2. **Network Configuration**: Inspects the Mycelium network configuration using the ZOS base library
3. **Address Generation**: Generates a random seed for the pod's unique IPv6 address
4. **Interface Creation**: Creates a virtual ethernet (veth) pair with one end in the pod's network namespace
5. **Bridge Attachment**: Attaches the other end of the veth pair to the `mycelium-br` bridge
6. **Address Assignment**: Configures the pod's network interface with the assigned IPv6 address
7. **Route Setup**: Sets up a default route for IPv6 traffic (`400::/7`) via the Mycelium gateway
8. **CNI Response**: Returns the result to Kubernetes CNI for pod initialization

**Prerequisites for Mycelium CNI:**

- Kubernetes cluster running on ThreeFold Grid nodes
- `mycelium-br` bridge interface configured (via `network_setup.sh`)
- `/etc/netseed` file containing the network seed for consistent addressing
- CNI binary installed in `/opt/cni/bin/`
- CNI configuration file in `/etc/cni/net.d/10-mycelium.conflist`

**Network Architecture Diagram:**

```text
┌─────────────┐    ┌─────────────┐    ┌─────────────┐
│    Pod A    │    │    Pod B    │    │    Pod C    │
│             │    │             │    │             │
│  ┌───────┐  │    │  ┌───────┐  │    │  ┌───────┐  │
│  │ veth0 │  │    │  │ veth0 │  │    │  │ veth0 │  │
│  └───┬───┘  │    │  └───┬───┘  │    │  └───┬───┘  │
└──────┼──────┘    └──────┼──────┘    └──────┼──────┘
       │                  │                  │
       └──────────────────┼──────────────────┘
                          │
                    ┌─────▼─────┐
                    │mycelium-br│
                    │  Bridge   │
                    └─────┬─────┘
                          │
                    ┌─────▼─────┐
                    │ Mycelium  │
                    │  Daemon   │
                    └─────┬─────┘
                          │
                    ┌─────▼─────┐
                    │   Peer    │
                    │  Network  │
                    └───────────┘
```

### 5. Custom Resource Definitions (CRDs)

**TFGW (ThreeFold Gateway) CRD:**

- Load balancing and proxying
- Hostname to backend mapping
- Automatic FQDN assignment
- Status reporting and health checks

**Controller Architecture:**

```text
Kubernetes API ←→ TFGW Controller ←→ Gateway Resources
                        ↓
                  Backend Services
```

### 6. Monitoring and Observability

**Prometheus Stack:**

- Metrics collection from all components
- Custom metrics for business logic
- Alerting rules and notifications
- Long-term storage and retention

**Grafana Dashboards:**

- Cluster health and performance
- Resource utilization metrics
- Business metrics and KPIs
- Custom alerting and notifications

**Logging:**

- Structured logging with Loki
- Centralized log aggregation
- Log rotation and retention
- Debug and audit trails

## Data Flow Architecture

### 1. User Authentication Flow

```text
User → Frontend → Backend API → JWT Service → Database
                     ↓
              Token Generation → Redis Cache → Response
```

### 2. Cluster Deployment Flow

```text
User Request → Frontend → Backend API → Validation
                             ↓
                    ThreeFold Grid API → Node Selection
                             ↓
                    Deployment Service → K3s Provisioning
                             ↓
                    Monitoring Setup → Status Updates
```

### 3. Network Communication Flow

```text
Pod A → Mycelium CNI → mycelium-br → Mycelium Daemon
                                          ↓
                                    Peer Network
                                          ↓
                                    Mycelium Daemon → mycelium-br → Mycelium CNI → Pod B
```

## Security Architecture

### 1. Authentication and Authorization

**Multi-layer Security:**

- JWT-based authentication
- Role-based access control (RBAC)
- API key management
- Session management with Redis

**User Roles:**

- **Admin**: Full platform access
- **User**: Cluster management within account
- **Viewer**: Read-only access

### 2. Network Security

**Encryption:**

- TLS/HTTPS for all web traffic
- Mycelium network encryption
- Inter-service communication encryption

**Network Isolation:**

- Kubernetes network policies
- Pod-to-pod communication controls
- Ingress/egress traffic filtering

### 3. Data Protection

**Data at Rest:**

- Database encryption
- Persistent volume encryption
- Backup encryption

**Data in Transit:**

- End-to-end encryption
- Certificate management
- Secure API communication

## Scalability and Performance

### 1. Horizontal Scaling

**Backend Services:**

- Stateless service design
- Load balancer distribution
- Auto-scaling based on metrics

**Database Scaling:**

- Read replicas for queries
- Connection pooling
- Query optimization

### 2. Caching Strategy

**Multi-level Caching:**

- Redis for session data
- Application-level caching
- CDN for static assets
- Database query caching

### 3. Performance Optimization

**Frontend:**

- Code splitting and lazy loading
- Asset optimization and compression
- Progressive web app features

**Backend:**

- Efficient database queries
- Background job processing
- Resource pooling

## Deployment Architecture

### 1. Development Environment

```text
Local Development:
├── Frontend (npm run dev)
├── Backend (go run main.go)
├── Database (SQLite)
└── Redis (local instance)
```

### 2. Staging Environment

```text
Staging Deployment:
├── Frontend (Docker container)
├── Backend (Docker container)
├── Database (PostgreSQL)
├── Redis (Docker container)
└── Monitoring (Prometheus/Grafana)
```

### 3. Production Environment

```text
Production Deployment:
├── Load Balancer
├── Frontend Cluster (multiple replicas)
├── Backend Cluster (multiple replicas)
├── Database Cluster (HA setup)
├── Redis Cluster
└── Monitoring Stack
```

## Integration Points

### 1. ThreeFold Grid Integration

**APIs Used:**

- GridProxy: Node discovery and information
- TFChain: Blockchain operations
- GraphQL: Grid data queries
- Activation Service: Account activation

### 2. External Services

**Payment Processing:**

- Stripe for credit card payments
- Webhook handling for payment events
- Invoice generation and management

**Email Services:**

- SendGrid for transactional emails
- Template management
- Delivery tracking

### 3. Monitoring Integration

**External Monitoring:**

- Health check endpoints
- Metrics export for external systems
- Alerting webhook integrations

## Disaster Recovery

### 1. Backup Strategy

**Data Backups:**

- Automated database backups
- etcd cluster state backups
- Configuration and secret backups

**Recovery Procedures:**

- Point-in-time recovery
- Cross-region backup replication
- Automated failover mechanisms

### 2. High Availability

**Service Redundancy:**

- Multiple backend instances
- Database clustering
- Load balancer redundancy

**Geographic Distribution:**

- Multi-region deployments
- Data replication
- Failover automation

## Future Architecture Considerations

### 1. Microservices Evolution

**Service Decomposition:**

- Breaking monolithic backend into microservices
- Event-driven architecture
- Service mesh implementation

### 2. Cloud-Native Enhancements

**Kubernetes Operators:**

- Custom operators for automation
- GitOps integration
- Advanced scheduling and placement

### 3. Edge Computing

**Edge Deployment:**

- Edge node support
- Distributed computing capabilities
- Latency optimization

## Performance Metrics

### 1. System Metrics

**Response Times:**

- API response time: < 200ms (95th percentile)
- UI load time: < 3 seconds
- Cluster deployment: 5-15 minutes

**Availability:**

- Service uptime: 99.9%
- Database availability: 99.95%
- Network connectivity: 99.8%

### 2. Business Metrics

**User Experience:**

- Deployment success rate: > 98%
- Support response time: < 4 hours
- User satisfaction: > 4.5/5

## Conclusion

Mycelium Cloud's architecture is designed for scalability, reliability, and ease of use. The modular design allows for independent scaling and updates of components while maintaining system integrity. The integration with ThreeFold Grid provides unique decentralized infrastructure capabilities, while standard cloud-native practices ensure reliability and performance.

For more detailed technical information, refer to:

- [API Reference](https://github.com/codescalers/kubecloud/blob/master_docs/frontend/kubecloud/public/docs/api-reference.md)
- [Deployment Guide](https://github.com/codescalers/kubecloud/blob/master_docs/frontend/kubecloud/public/docs/tutorial.md)
- [FAQ](https://github.com/codescalers/kubecloud/blob/master_docs/frontend/kubecloud/public/docs/faq.md)
