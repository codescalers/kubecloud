# Deployment Guide

This guide covers deploying Mycelium Cloud in different environments.

## Deployment Options

- [Local Development](./LOCAL_DEV.md) - Single machine setup
- [Docker Compose](./DOCKER_COMPOSE.md) - Multi-container orchestration
- [Kubernetes](./KUBERNETES.md) - Production-grade deployment
- [TFGrid](./TFGRID.md) - Decentralized infrastructure

## Quick Start

### Local Development
```bash
cd backend && make run
cd ../frontend/kubecloud && npm run dev
```

### Docker Compose
```bash
cp backend/config-example.json backend/config.json
docker-compose up
```

## Environment Configuration

Create configuration files in the root directory:

- `backend/config.json` - Backend configuration
- `frontend/kubecloud/.env` - Frontend environment variables
- `prod_config.json` - Production configuration

Copy from example files:
```bash
cp backend/config-example.json backend/config.json
cp frontend/kubecloud/env.example frontend/kubecloud/.env
```

## Service Access

| Service | Local Dev | Docker |
|---------|-----------|--------|
| Frontend | http://localhost:5173 | http://localhost:8000 |
| Backend API | http://localhost:8080 | http://localhost:8080 |
| Grafana | http://localhost:3000 | http://localhost:3000 |
| Prometheus | http://localhost:9090 | http://localhost:9090 |

## Prerequisites

- **Go** 1.19+
- **Node.js** 20+
- **Docker** & **Docker Compose**
- **Git**

## Documentation

- [Local Development Setup](./LOCAL_DEV.md)
- [Docker Compose Guide](./DOCKER_COMPOSE.md)
- [Kubernetes Deployment](./KUBERNETES.md)
- [TFGrid Deployment](./TFGRID.md)

## Troubleshooting

### Port Conflicts
```bash
# Find process using port 8080
lsof -i :8080
# Kill process
kill -9 <PID>
```

### Docker Issues
```bash
# Clear everything and rebuild
docker-compose down -v
docker-compose up --build
```

### Database Issues
- Verify credentials in config
- Check database is running
- Review logs for errors

## Next Steps

1. Choose your deployment method
2. Follow the specific deployment guide
3. Configure environment variables
4. Start the services
5. Access the application

## Support

- [Getting Started Guide](../GETTING_STARTED.md)
- [Architecture Overview](../architecture/OVERVIEW.md)
- [Report Issues](https://github.com/codescalers/kubecloud/issues)
