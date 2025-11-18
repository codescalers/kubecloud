# Mycelium Cloud Documentation

Welcome to the Mycelium Cloud documentation! This directory contains comprehensive guides and references for using, deploying, and contributing to the project.

## Documentation Structure

### Getting Started
- **[Getting Started Guide](./GETTING_STARTED.md)** - Quick setup and running locally
  - Prerequisites and installation
  - Local development setup
  - Docker Compose quick start
  - Mycelium networking setup

### Core Documentation

#### Architecture
- **[Architecture Overview](./architecture/OVERVIEW.md)** - System design and components
  - System architecture diagram
  - Component descriptions
  - Data flow
  - Technology stack

#### API
- **[API Documentation](./api/README.md)** - REST API reference
  - Endpoints and resources
  - Authentication
  - Request/response formats
  - Error codes
  - Rate limiting
  - Examples and SDKs

#### Deployment
- **[Deployment Guide](./deployment/README.md)** - Production deployment
  - Local development
  - Docker Compose
  - Kubernetes deployment
  - TFGrid deployment
  - Environment configuration

#### Contributing
- **[Contributing Guide](./contributing/CONTRIBUTING.md)** - How to contribute
  - Code of conduct
  - Development setup
  - Code style guidelines
  - Testing requirements
  - Pull request process
  - Areas for contribution

### Development Guides

- **[Backend README](../backend/README.md)** - Go API server
  - Configuration options
  - API endpoints
  - Database setup
  - Running tests

- **[Frontend README](../frontend/kubecloud/README.md)** - Vue.js application
  - Development environment
  - Building and testing
  - Component structure
  - Configuration

- **[CRD Documentation](../crd/README.md)** - Kubernetes Custom Resources
  - CRD definitions
  - Controller setup
  - Usage examples

- **[Mycelium CNI](../mycelium-cni/README.md)** - P2P networking
  - Installation
  - Configuration
  - Troubleshooting

## Quick Navigation

| Need | Go To |
|------|-------|
| Set up locally | [Getting Started](./GETTING_STARTED.md) |
| Understand architecture | [Architecture Overview](./architecture/OVERVIEW.md) |
| Use the API | [API Documentation](./api/README.md) |
| Deploy to production | [Deployment Guide](./deployment/README.md) |
| Contribute code | [Contributing Guide](./contributing/CONTRIBUTING.md) |
| Configure backend | [Backend README](../backend/README.md) |
| Build frontend | [Frontend README](../frontend/kubecloud/README.md) |

## By Role

### For Users
1. Start with [Getting Started](./GETTING_STARTED.md)
2. Read [Architecture Overview](./architecture/OVERVIEW.md) to understand the system
3. Check [API Documentation](./api/README.md) for integrations

### For Developers
1. Clone and follow [Getting Started](./GETTING_STARTED.md)
2. Read [Contributing Guide](./contributing/CONTRIBUTING.md)
3. Explore component READMEs:
   - [Backend](../backend/README.md)
   - [Frontend](../frontend/kubecloud/README.md)
4. Check [Architecture](./architecture/OVERVIEW.md) for system design

### For DevOps/Deployment
1. Review [Deployment Guide](./deployment/README.md)
2. Check component-specific docs:
   - [Backend Configuration](../backend/README.md)
   - [Docker Compose](./GETTING_STARTED.md#option-a-docker-compose-recommended-for-full-stack)
   - [Kubernetes](./deployment/KUBERNETES.md)
   - [TFGrid](./deployment/TFGRID.md)

### For Contributors
1. Read [Contributing Guide](./contributing/CONTRIBUTING.md)
2. Follow [Getting Started](./GETTING_STARTED.md) for setup
3. Check [Code Style Guidelines](./contributing/CONTRIBUTING.md#code-style)
4. Review [Pull Request Process](./contributing/CONTRIBUTING.md#pull-request-process)

## Additional Resources

### Component Documentation
- **Backend**: `backend/README.md`
- **Frontend**: `frontend/kubecloud/README.md`
- **CRD**: `crd/README.md`
- **Ingress Controller**: `ingress-controller/README.md`
- **Mycelium CNI**: `mycelium-cni/README.md`
- **K3s**: `k3s/README.md`

### External Links
- [GitHub Repository](https://github.com/codescalers/kubecloud)
- [Issue Tracker](https://github.com/codescalers/kubecloud/issues)
- [Discussions](https://github.com/codescalers/kubecloud/discussions)
- [ThreeFold Grid](https://threefold.io/)

## Documentation Standards

All documentation in this project follows these standards:

- Clear, concise language
- Practical examples
- Proper formatting and structure
- Links to related docs
- Keep it up-to-date
- Include troubleshooting tips

## Contributing Documentation

Documentation contributions are welcome! See [Contributing Guide](./contributing/CONTRIBUTING.md#documentation) for details.

**To contribute docs:**
1. Fork the repository
2. Create a branch for your changes
3. Follow the [Documentation Standards](#documentation-standards)
4. Submit a pull request

---

**Last Updated**: 2025-11-16

For the latest information, visit the [GitHub Repository](https://github.com/codescalers/kubecloud).
