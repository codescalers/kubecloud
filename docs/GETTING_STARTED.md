# Getting Started with Mycelium Cloud

This guide will help you set up Mycelium Cloud for local development or production deployment.

## Prerequisites

Before getting started, ensure you have the following installed:

- **Go**: Version 1.19 or later ([Download](https://golang.org/dl/))
- **Node.js**: Version 20 or later ([Download](https://nodejs.org/))
- **Docker**: Latest stable version ([Download](https://docker.com/))
- **Docker Compose**: Latest version ([Install](https://docs.docker.com/compose/install/))
- **Git**: For cloning the repository

## Quick Start

### 1. Clone the Repository

```bash
git clone https://github.com/codescalers/kubecloud.git
cd kubecloud
```

### 2. Choose Your Setup Method

#### Option A: Docker Compose (Recommended for Full Stack)

The easiest way to get everything running:

```bash
cp backend/config-example.json backend/config.json
docker-compose up
```

This starts all services:
- **Frontend UI**: http://localhost:8000
- **Backend API**: http://localhost:8080
- **Grafana**: http://localhost:3000 (admin/admin)
- **Prometheus**: http://localhost:9090

#### Option B: Local Development (Backend + Frontend)

##### Backend Setup

```bash
cd backend
cp config-example.json config.json
# Edit config.json with your settings
make run
```

The backend will start on http://localhost:8080.

##### Frontend Setup (in a new terminal)

```bash
cd frontend/kubecloud
npm install
npm run dev
```

The frontend will start on http://localhost:5173.

#### Option C: Make Commands (Simplified)

From the root directory:

```bash
# Run both backend and frontend
make run

# Or run them separately
make backend-run    # Terminal 1
make frontend-run   # Terminal 2
```

## Local Development with Mycelium

If running the backend locally (not in Docker), you need to run the Mycelium binary separately.

### Setup Mycelium

1. **Download Mycelium binary**:

```bash
wget https://github.com/threefoldtech/mycelium/releases/latest/download/mycelium-private-x86_64-unknown-linux-musl.tar.gz
tar -xzf mycelium-private-x86_64-unknown-linux-musl.tar.gz
sudo chmod +x mycelium-private
sudo mv mycelium-private /usr/local/bin/mycelium
```

2. **Start Mycelium** (in a separate terminal):

```bash
sudo mycelium --peers \
  tcp://188.40.132.242:9651 \
  tcp://136.243.47.186:9651 \
  tcp://185.69.166.7:9651 \
  tcp://185.69.166.8:9651 \
  tcp://65.21.231.58:9651 \
  tcp://65.109.18.113:9651 \
  tcp://209.159.146.190:9651 \
  tcp://5.78.122.16:9651 \
  tcp://5.223.43.251:9651 \
  tcp://142.93.217.194:9651
```

3. **Start Backend** (in another terminal):

```bash
cd backend
make run
```

4. **Start Frontend** (in another terminal):

```bash
cd frontend/kubecloud
npm install
npm run dev
```

## Configuration

### Backend Configuration

The backend uses a JSON configuration file. Copy and customize:

```bash
cd backend
cp config-example.json config.json
# Edit config.json with your TFGrid credentials and settings
```

Key configuration includes:
- Database connection (PostgreSQL in Docker Compose)
- Redis connection for caching
- JWT token secrets for authentication
- Mail service (SendGrid) credentials
- Stripe integration for billing
- TFGrid system account mnemonic for blockchain operations

See [Backend README](../../backend/README.md) for full configuration options.

### Frontend Configuration

The frontend uses environment variables. Copy and configure:

```bash
cd frontend/kubecloud
cp env.example .env
# Edit .env with your API base URL and feature flags
```

Key environment variables:
- `VITE_API_BASE_URL` - Backend API endpoint (default: http://localhost:8080/api)
- `VITE_ENABLE_DEBUG_MODE` - Enable/disable debug logging
- `VITE_AUTH_DOMAIN` - Auth0 domain (optional)
- `VITE_STRIPE_PUBLISHABLE_KEY` - Stripe key (optional)

## Troubleshooting

### Backend won't start

- Check Go version: `go version`
- Verify configuration file exists and is valid JSON
- Ensure required ports are available

### Frontend build fails

- Clear node_modules: `rm -rf node_modules && npm install`
- Check Node.js version: `node --version`

### Docker Compose issues

- Ensure Docker daemon is running
- Check port conflicts
- Try: `docker-compose down && docker-compose up --build`

### Database connection errors

- Verify database credentials in configuration
- Check database service is running
- Review connection logs

## Next Steps

- Read the [Architecture Overview](../architecture/OVERVIEW.md)
- Check the [API Documentation](../api/README.md)
- Explore [Contributing Guidelines](../contributing/CONTRIBUTING.md)
- Review [Deployment Guide](../deployment/README.md)

## Getting Help

- Check [component-specific README files](../../README.md#project-structure)
- Report issues on [GitHub Issues](https://github.com/codescalers/kubecloud/issues)
- Engage with the community

## See Also

- [Backend README](../../backend/README.md)
- [Frontend README](../../frontend/kubecloud/README.md)
- [CRD Documentation](../../crd/README.md)
