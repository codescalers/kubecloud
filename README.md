# Mycelium Cloud

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Version](https://img.shields.io/badge/Go-1.19+-00ADD8.svg)](https://golang.org/)
[![Node.js Version](https://img.shields.io/badge/Node.js-20+-339933.svg)](https://nodejs.org/)
[![Docker](https://img.shields.io/badge/Docker-Latest-2496ED.svg)](https://docker.com/)

Cloud-native platform for deploying and managing Kubernetes clusters on decentralized infrastructure (TFGrid).

**Decentralized Kubernetes** | **Modern UI** | **Observability** | **P2P Networking** | **Docker Ready**

## Quick Start

```bash
git clone https://github.com/codescalers/kubecloud.git
cd kubecloud
cp backend/config-example.json backend/config.json
docker-compose up
```

**Access:** Frontend http://localhost:8000 · API http://localhost:8080 · Grafana http://localhost:3000

👉 [Full setup guide →](docs/GETTING_STARTED.md)

## Documentation

| | |
|---|---|
| **Getting Started** | [Setup and run locally](docs/GETTING_STARTED.md) |
| **Architecture** | [System design](docs/architecture/OVERVIEW.md) |
| **API Docs** | [REST API reference](docs/api/README.md) |
| **Deployment** | [Production options](docs/deployment/README.md) |
| **Contributing** | [How to contribute](docs/contributing/CONTRIBUTING.md) |

## Project Structure

- **backend/** - Go API server
- **frontend/kubecloud/** - Vue.js web UI
- **crd/** - Kubernetes operators
- **mycelium-cni/** - P2P networking
- **docs/** - Full documentation

[See component READMEs →](docs/README.md#-documentation-structure)

## License

Apache License 2.0 - [LICENSE](LICENSE)
