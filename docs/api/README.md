# API Documentation

Mycelium Cloud provides comprehensive REST APIs for all cluster management operations.

## API Overview

The Backend APIs are built with Go and provide endpoints for:

- **Authentication & Authorization**
- **Cluster Management**
- **Deployment Operations**
- **User Management**
- **Billing & Resource Tracking**
- **Monitoring & Metrics**

## Base URL

- **Development**: `http://localhost:8080`
- **Production**: Configure in your deployment

## API Documentation

### Available Resources

- [Cluster API](./CLUSTER.md) - Cluster operations
- [Deployment API](./DEPLOYMENT.md) - Deployment management
- [User API](./USER.md) - User operations
- [Billing API](./BILLING.md) - Billing information
- [Metrics API](./METRICS.md) - Monitoring metrics

## OpenAPI/Swagger

Swagger/OpenAPI specifications are available in `backend/docs/swagger/`

### Access Swagger UI

When running locally:
```
http://localhost:8080/swagger-ui.html
```

## Authentication

All API requests require authentication via:

- **Bearer Token**: Include in Authorization header
  ```
  Authorization: Bearer <token>
  ```
- **API Key**: Include in X-API-Key header
  ```
  X-API-Key: <api-key>
  ```

## Request Format

All requests use JSON format:

```bash
curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-cluster"
  }'
```

## Response Format

Successful responses (2xx):
```json
{
  "success": true,
  "data": { ... },
  "message": "Operation successful"
}
```

Error responses (4xx, 5xx):
```json
{
  "success": false,
  "error": "Error description",
  "code": "ERROR_CODE"
}
```

## Rate Limiting

API requests are rate-limited per authenticated user:

- **Rate Limit**: 1000 requests per hour
- **Headers**:
  ```
  X-RateLimit-Limit: 1000
  X-RateLimit-Remaining: 999
  X-RateLimit-Reset: 1234567890
  ```

## Pagination

List endpoints support pagination:

```bash
GET /api/v1/clusters?page=1&limit=20
```

Response:
```json
{
  "success": true,
  "data": [...],
  "pagination": {
    "page": 1,
    "limit": 20,
    "total": 100,
    "pages": 5
  }
}
```

## Error Codes

| Code | Status | Description |
|------|--------|-------------|
| 200 | OK | Request successful |
| 201 | Created | Resource created |
| 400 | Bad Request | Invalid request |
| 401 | Unauthorized | Auth required |
| 403 | Forbidden | Not authorized |
| 404 | Not Found | Resource not found |
| 429 | Too Many Requests | Rate limited |
| 500 | Server Error | Internal error |

## Common Endpoints

### Health Check
```bash
GET /health
```

### Metrics
```bash
GET /metrics
```

### API Version
```bash
GET /api/v1
```

## SDKs and Client Libraries

For integration with your applications, client libraries are available:

- [Go SDK](./SDK_GO.md)
- [JavaScript/TypeScript SDK](./SDK_JS.md)
- [Python SDK](./SDK_PYTHON.md)

## Examples

### Create a Cluster
```bash
curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production-cluster",
    "nodes": 3
  }'
```

### Get Cluster Status
```bash
curl http://localhost:8080/api/v1/clusters/cluster-id \
  -H "Authorization: Bearer $TOKEN"
```

### List All Clusters
```bash
curl http://localhost:8080/api/v1/clusters \
  -H "Authorization: Bearer $TOKEN"
```

## Webhooks

Setup webhooks to receive notifications about cluster events:

```bash
POST /api/v1/webhooks
{
  "url": "https://your-domain.com/webhook",
  "events": ["cluster.created", "cluster.deleted"]
}
```

## API Versioning

Current API version: **v1**

We maintain backward compatibility within major versions. Breaking changes are announced in release notes.

## Documentation

- Backend source code: `backend/docs/`
- API specifications: `backend/docs/swagger/`
- API tests: `backend/tests/`

## Support

- [Getting Started Guide](../GETTING_STARTED.md)
- [Architecture Overview](../architecture/OVERVIEW.md)
- [Report Issues](https://github.com/codescalers/kubecloud/issues)

## See Also

- [Backend README](../../backend/README.md)
- [Contributing Guide](../contributing/CONTRIBUTING.md)
