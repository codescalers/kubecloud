# API Reference

Mycelium Cloud provides a comprehensive REST API for managing clusters, users, billing, and monitoring.

## Base URL

```text
https://staging.vdc.grid.tf/api/v1
```

## Authentication

All API requests require authentication using JWT tokens. Include the token in the Authorization header:

```bash
Authorization: Bearer <your-jwt-token>
```

### Obtaining a Token

```bash
POST /auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "your-password"
}
```

## Core Endpoints

### Authentication

#### Login

```bash
POST /auth/login
```

Authenticate user and receive JWT tokens.

#### Refresh Token

```bash
POST /auth/refresh
```

Refresh expired access token using refresh token.

#### Logout

```bash
POST /auth/logout
```

Invalidate current session tokens.

### User Management

#### Get User Profile

```bash
GET /user/profile
```

Retrieve current user information and account details.

#### Update Profile

```bash
PUT /user/profile
```

Update user profile information.

#### Change Password

```bash
POST /user/change-password
```

Change user password with current password verification.

### Cluster Management

#### List Clusters

```bash
GET /clusters
```

Retrieve all clusters for the authenticated user.

#### Create Cluster

```bash
POST /clusters
```

Deploy a new Kubernetes cluster.

**Request Body:**

```json
{
  "name": "my-cluster",
  "nodes": [
    {
      "name": "master-1",
      "cpu": 2,
      "memory": 4096,
      "storage": 50,
      "type": "master",
      "node_id": 123
    },
    {
      "name": "worker-1", 
      "cpu": 4,
      "memory": 8192,
      "storage": 100,
      "type": "worker",
      "node_id": 456
    }
  ],
  "token": "cluster-join-token"
}
```

#### Get Cluster Details

```bash
GET /clusters/{id}
```

Retrieve detailed information about a specific cluster.

#### Update Cluster

```bash
PUT /clusters/{id}
```

Modify cluster configuration (add/remove nodes).

#### Delete Cluster

```bash
DELETE /clusters/{id}
```

Permanently delete a cluster and all associated resources.

#### Get Kubeconfig

```bash
GET /clusters/{id}/kubeconfig
```

Download the kubeconfig file for cluster access.

### Node Management

#### List Available Nodes

```bash
GET /nodes
```

Retrieve available ThreeFold Grid nodes for deployment.

**Query Parameters:**

- `cpu_min`: Minimum CPU cores
- `memory_min`: Minimum memory (MB)
- `storage_min`: Minimum storage (GB)
- `country`: Filter by country code
- `farm_id`: Filter by farm ID

#### Get Node Details

```bash
GET /nodes/{id}
```

Retrieve detailed information about a specific node.

### Billing & Payments

#### Get Account Balance

```bash
GET /billing/balance
```

Retrieve current account balance and credit information.

#### Charge Balance

```bash
POST /billing/charge
```

Add funds to account using payment method.

**Request Body:**

```json
{
  "amount": 100.00,
  "payment_method_id": "pm_1234567890",
  "card_type": "visa"
}
```

#### Get Invoices

```bash
GET /billing/invoices
```

Retrieve billing history and invoices.

#### Get Usage Statistics

```bash
GET /billing/usage
```

Retrieve resource usage statistics and costs.

### SSH Key Management

#### List SSH Keys

```bash
GET /ssh-keys
```

Retrieve all SSH keys for the user.

#### Add SSH Key

```bash
POST /ssh-keys
```

Add a new SSH public key.

**Request Body:**

```json
{
  "name": "my-key",
  "public_key": "ssh-rsa AAAAB3NzaC1yc2E..."
}
```

#### Delete SSH Key

```bash
DELETE /ssh-keys/{id}
```

Remove an SSH key from the account.

### Notifications

#### Get Notifications

```bash
GET /notifications
```

Retrieve user notifications and alerts.

#### Mark as Read

```bash
PUT /notifications/{id}/read
```

Mark a notification as read.

## Admin Endpoints

Admin-only endpoints for platform management:

### User Administration

- `GET /admin/users` - List all users
- `GET /admin/users/{id}` - Get user details
- `PUT /admin/users/{id}` - Update user
- `DELETE /admin/users/{id}` - Delete user

### System Management

- `GET /admin/stats` - Get platform statistics
- `GET /admin/clusters` - List all clusters
- `GET /admin/nodes` - List all nodes

### Billing Administration

- `POST /admin/credit` - Add manual credit to user
- `GET /admin/invoices` - List all invoices
- `GET /admin/vouchers` - Manage vouchers

## Error Handling

The API uses standard HTTP status codes and returns errors in JSON format:

```json
{
  "status": 400,
  "error": "Bad Request",
  "message": "Invalid cluster configuration",
  "data": null
}
```

### Common Status Codes

- `200 OK` - Request successful
- `201 Created` - Resource created successfully
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Resource not found
- `500 Internal Server Error` - Server error

## Rate Limiting

API requests are rate limited to prevent abuse:

- **Authenticated users**: 1000 requests per hour
- **Unauthenticated**: 100 requests per hour

Rate limit headers are included in responses:

```json
X-RateLimit-Limit: 1000
X-RateLimit-Remaining: 999
X-RateLimit-Reset: 1640995200
```

## SDKs and Examples

### cURL Examples

**Create a cluster:**

```bash
curl -X POST https://staging.vdc.grid.tf/api/v1/clusters \
  -H "Authorization: Bearer <token>" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production-cluster",
    "nodes": [
      {
        "name": "master-1",
        "cpu": 4,
        "memory": 8192,
        "storage": 100,
        "type": "master",
        "node_id": 123
      }
    ]
  }'
```

**Get cluster status:**

```bash
curl -X GET https://staging.vdc.grid.tf/api/v1/clusters/1 \
  -H "Authorization: Bearer <token>"
```

### JavaScript/Node.js Example

```javascript
const axios = require('axios');

const api = axios.create({
  baseURL: 'https://staging.vdc.grid.tf/api/v1',
  headers: {
    'Authorization': `Bearer ${token}`,
    'Content-Type': 'application/json'
  }
});

// Create cluster
const cluster = await api.post('/clusters', {
  name: 'my-cluster',
  nodes: [
    {
      name: 'master-1',
      cpu: 2,
      memory: 4096,
      storage: 50,
      type: 'master',
      node_id: 123
    }
  ]
});

console.log('Cluster created:', cluster.data);
```

## Webhooks

Configure webhooks to receive real-time notifications about cluster events:

### Webhook Events

- `cluster.created` - New cluster deployed
- `cluster.updated` - Cluster configuration changed
- `cluster.deleted` - Cluster removed
- `cluster.failed` - Deployment failed
- `billing.charged` - Payment processed
- `billing.failed` - Payment failed

### Webhook Payload

```json
{
  "event": "cluster.created",
  "timestamp": "2023-12-01T10:00:00Z",
  "data": {
    "cluster_id": 123,
    "name": "production-cluster",
    "status": "running"
  }
}
```
