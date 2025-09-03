# KubeCloud

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8.svg)](https://golang.org/)
[![Node.js Version](https://img.shields.io/badge/Node.js-18+-339933.svg)](https://nodejs.org/)

KubeCloud is a comprehensive cloud-native platform designed for deploying, managing, and monitoring Kubernetes clusters on the decentralized TFGrid infrastructure. It provides a complete solution with backend APIs, frontend UI, custom resource definitions (CRDs), monitoring dashboards, and networking components to streamline cloud operations in a decentralized environment.

## Table of Contents

- [Features](#features)
- [Architecture Overview](#architecture-overview)
- [Prerequisites](#prerequisites)
- [Getting Started](#getting-started)
- [Repository Structure](#repository-structure)
- [Documentation](#documentation)
- [API Documentation](#api-documentation)
- [Contributing](#contributing)
- [Troubleshooting](#troubleshooting)
- [Support](#support)
- [License](#license)

## Features

- **Decentralized Kubernetes Management**: Deploy and manage Kubernetes clusters on TFGrid
- **Multi-Component Architecture**: Backend (Go), Frontend (Vue.js), CRDs, and networking plugins
- **Monitoring & Observability**: Integrated Prometheus and Grafana for metrics and dashboards
- **Custom Networking**: Mycelium CNI plugin for peer-to-peer networking
- **Ingress Management**: Custom ingress controller for traffic routing
- **RESTful APIs**: Comprehensive backend APIs for cluster operations
- **Web UI**: Modern Vue.js frontend for cluster management
- **Docker Integration**: Containerized deployment with Docker Compose
- **Configuration Management**: Flexible configuration via files, environment variables, and CLI flags

## Architecture Overview

```text
+-------------------+    +-------------------+    +-------------------+
|   Frontend UI     |    |   Backend APIs    |    |      TFGrid       |
|    (Vue.js)       |--->|      (Go)         |<-->|  (Decentralized)  |
+-------------------+    +-------------------+    +-------------------+
         |                      |                        |
         |                      |                        |
         |                 +----v----+                   |
         |                 | Metrics |                   |
         |                 | /metrics|                   |
         |                 +----+----+                   |
         |                      |                        |
         v                      v                        v
+-------------------+    +-------------------+    +-------------------+
|   Monitoring      |    |   CRDs &          |    |   Networking      |
| (Prometheus/Grafana)|    | Controllers      |    |   (Mycelium)      |
+-------------------+    +-------------------+    +-------------------+
```

## Prerequisites

Before getting started, ensure you have the following installed:

- **Go**: Version 1.19 or later ([Download](https://golang.org/dl/))
- **Node.js**: Version 18 or later ([Download](https://nodejs.org/))
- **Docker**: Latest stable version ([Download](https://docker.com/))
- **Docker Compose**: Latest version ([Install](https://docs.docker.com/compose/install/))
- **Git**: For cloning the repository

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/codescalers/kubecloud.git
cd kubecloud
```

### 2. Backend Setup

Navigate to the backend directory and follow the setup instructions:

```bash
cd backend
# Copy example configuration
cp config-example.json config.json
# Edit config.json with your settings (see backend/README.md for details)
# Build and run
make build
make run .
```

The backend will start on the configured port (default: 8080).

### 3. Frontend Setup

Navigate to the frontend directory and install dependencies:

```bash
cd ../frontend/kubecloud
npm install
npm run dev
```

The frontend development server will start on `http://localhost:5173` (default Vite port).

### 4. Docker Compose (Full Stack)

For a complete local development environment:

```bash
# From the root directory
docker-compose up
```

This will start all services including the backend, frontend, monitoring stack, and databases.

### 5. Access the Application

- **Frontend UI**: <http://localhost:5173>
- **Backend API**: <http://localhost:8080>
- **Grafana**: <http://localhost:3000> (admin/admin)
- **Prometheus**: <http://localhost:9090>

## Repository Structure

```bash
kubecloud/
├── backend/                 # Go backend services and APIs
│   ├── app/                # Application handlers and logic
│   ├── internal/           # Internal packages and utilities
│   ├── models/             # Database models
│   ├── docs/               # API documentation
│   └── cmd/                # CLI commands
├── frontend/               # Vue.js frontend application
│   └── kubecloud/          # Main frontend app
├── crd/                    # Custom Resource Definitions
├── grafana/                # Grafana monitoring dashboards
├── grafana-gen/            # Grafana dashboard generator
├── ingress-controller/     # Custom ingress controller
├── k3s/                    # K3s-related manifests and scripts
├── mycelium-cni/           # Mycelium CNI plugin
├── mycelium-peer/          # Mycelium peer networking
├── clean-routes-cni/       # CNI cleanup utilities
├── docker-compose.yml      # Multi-service orchestration
├── prometheus.yml          # Prometheus configuration
└── README.md               # This file
```

## Documentation

- **Backend Configuration**: See `backend/README.md` for detailed configuration options
- **Frontend Development**: See `frontend/kubecloud/README.md` for frontend setup
- **CRD Documentation**: See `crd/README.md` for custom resource definitions
- **K3s Integration**: See `k3s/README.md` for K3s deployment guides
- **Networking**: See `mycelium-cni/README.md` for networking setup

## API Documentation

The backend provides comprehensive REST APIs for:

- Cluster management
- Deployment operations
- User authentication
- Billing and invoicing
- Monitoring and metrics
- Network configuration

API documentation is available at `backend/docs/` and includes Swagger/OpenAPI specifications.

## Contributing

We welcome contributions! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go coding standards for backend contributions
- Use ESLint and Prettier for frontend code
- Write tests for new features
- Update documentation for API changes
- Ensure Docker compatibility

## Troubleshooting

### Common Issues

**Backend won't start:**

- Check Go version (`go version`)
- Verify configuration file exists and is valid JSON
- Ensure required ports are available

**Frontend build fails:**

- Clear node_modules: `rm -rf node_modules && npm install`
- Check Node.js version (`node --version`)

**Docker Compose issues:**

- Ensure Docker daemon is running
- Check port conflicts
- Try `docker-compose down` then `docker-compose up --build`

**Database connection errors:**

- Verify database credentials in configuration
- Check database service is running
- Review connection logs

### Getting Help

- Check existing issues on GitHub
- Review component-specific README files
- Enable debug logging for detailed error information

## Support

- **Issues**: [GitHub Issues](https://github.com/codescalers/kubecloud/issues)
- **Documentation**: Component-specific README files
- **Community**: Join our community channels for support

## License

This project is licensed under the Apache License 2.0 - see the [LICENSE](LICENSE) file for details.

---

**KubeCloud** - Empowering decentralized cloud operations with Kubernetes.
