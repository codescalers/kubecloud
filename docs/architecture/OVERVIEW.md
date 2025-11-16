# Mycelium Cloud Architecture

## System Overview

Mycelium Cloud is a cloud-native platform for deploying and managing Kubernetes clusters on decentralized infrastructure. The architecture consists of multiple integrated components working together to provide a complete solution.

```text
┌─────────────────────────────────────────────────────────────────┐
│                     Frontend UI (Vue.js)                        │
│              http://localhost:8000 (Development)                │
└────────────────────┬────────────────────────────────────────────┘
                     │
                     │ REST API
                     ▼
┌─────────────────────────────────────────────────────────────────┐
│                   Backend APIs (Go)                             │
│              http://localhost:8080 (Development)                │
│                                                                 │
│  ┌─────────────┐  ┌─────────────┐  ┌─────────────┐            │
│  │  Auth       │  │  Cluster    │  │  Billing    │            │
│  │  Management │  │  Management │  │  Service    │            │
│  └─────────────┘  └─────────────┘  └─────────────┘            │
└────────────────┬──────────────────────────┬────────────────────┘
                 │                          │
                 ▼                          ▼
        ┌─────────────────┐      ┌──────────────────┐
        │  Metrics/       │      │  TFGrid          │
        │  Monitoring     │      │  Infrastructure  │
        │  (/metrics)     │      │  (Decentralized) │
        └─────────────────┘      └──────────────────┘
                 │
    ┌────────────┴────────────┐
    ▼                         ▼
┌────────────────┐    ┌──────────────────┐
│  Prometheus    │    │  Grafana         │
│  Metrics Store │    │  Dashboards      │
└────────────────┘    └──────────────────┘
```

## Core Components

### Frontend (Vue.js)
- Modern, responsive web interface for cluster management
- Real-time updates via WebSockets/SSE
- User authentication and dashboards
- Location: `frontend/kubecloud/`

### Backend (Go)
- RESTful API server for all cluster operations
- User authentication and authorization
- Cluster management and deployment orchestration
- Billing and resource tracking
- Integration with TFGrid infrastructure
- Location: `backend/`

### Database & Storage
- PostgreSQL for persistent data storage
- Redis for caching and session management
- SQLite option for local development

### Networking & Infrastructure

#### Mycelium Peer
- Decentralized peer networking integration
- Runs in Docker or as standalone binary for connectivity
- Location: `mycelium-peer/`

#### Mycelium CNI
- Custom Container Network Interface plugin for P2P networking
- Location: `mycelium-cni/`

#### Custom Ingress Controller
- Traffic routing and ingress management
- Location: `ingress-controller/`

#### CRD (Custom Resource Definitions)
- Kubernetes operators for custom resources
- Cluster provisioning and lifecycle management
- Location: `crd/`

### Monitoring & Observability

#### Prometheus
- Metrics collection and storage
- Scrapes application and infrastructure metrics
- Data retention: 30 days (configurable)

#### Grafana
- Dashboard visualization and alerting
- Pre-configured dashboards for system monitoring
- Admin credentials: admin/admin (change in production)

#### Loki
- Log aggregation system
- Centralized log storage and querying
- Promtail log shipping agent

## Data Flow

1. **User Interaction**
   - User accesses Frontend UI (http://localhost:8000 or deployment URL)
   - Frontend sends REST API requests to Backend

2. **API Processing**
   - Backend receives and validates requests
   - Performs authentication and authorization
   - Executes business logic and database operations
   - Returns JSON responses to Frontend

3. **Data Persistence**
   - Backend stores data in PostgreSQL database
   - Uses Redis for session caching and temporary data
   - Maintains cluster state and deployment records

4. **Cluster Operations**
   - Backend communicates with TFGrid infrastructure
   - Manages Kubernetes clusters via CRDs
   - Uses Mycelium peer networking for decentralized connectivity

5. **Monitoring & Observability**
   - Backend exposes metrics endpoint (`/metrics`)
   - Prometheus scrapes metrics at configured intervals
   - Grafana queries Prometheus for visualization
   - Loki aggregates logs from backend and services

## Communication Patterns

### API Communication
- **Frontend ↔ Backend**: REST APIs with JSON payloads
- **Backend ↔ TFGrid**: gRPC or REST (depending on component)

### Service-to-Service
- Networking via Mycelium for decentralized connectivity
- Direct networking within Docker Compose environment

### Metrics & Monitoring
- Backend exposes Prometheus metrics
- Prometheus scrapes at configured intervals
- Grafana queries Prometheus for visualization

## Deployment Models

### Local Development
- All services on single machine
- Docker Compose for orchestration
- Direct file-based configuration

### Production on TFGrid
- Distributed across TFGrid nodes
- Decentralized networking via Mycelium
- Kubernetes-native deployment

## Technology Stack

| Component | Technology | Version |
|-----------|-----------|---------|
| Backend | Go | 1.19+ |
| Frontend | Vue.js / TypeScript | Vue 3 |
| Frontend Build | Vite | Latest |
| Node.js | Runtime | 20+ |
| Container Runtime | Docker | Latest |
| Orchestration | Docker Compose / Kubernetes | Latest |
| Metrics | Prometheus | Latest |
| Visualization | Grafana | Latest |
| Logs | Loki | Latest |

## Key Features

- **Multi-tenant Support**: Isolated cluster management per tenant
- **Decentralized Infrastructure**: Leverages TFGrid for compute
- **Monitoring & Observability**: Built-in prometheus/Grafana
- **Custom Networking**: Mycelium for peer-to-peer communication
- **Scalability**: Kubernetes-native architecture
- **API-first**: Complete REST API for all operations

## See Also

- [Getting Started Guide](../GETTING_STARTED.md)
- [Deployment Guide](../deployment/README.md)
- [Backend Documentation](../../backend/README.md)
- [Frontend Documentation](../../frontend/kubecloud/README.md)
