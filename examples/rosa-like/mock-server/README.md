# ROSA Mock API Server

A complete mock implementation of the Red Hat OpenShift Service on AWS (ROSA) API server for testing and development.

## Features

- **OAuth2 Authentication**: Authorization code and refresh token flows
- **Thread-safe Data Store**: In-memory storage with sync.RWMutex
- **RESTful API**: Full CRUD operations for all resource types
- **Nested Resources**: Machine pools, identity providers, addons, and upgrade policies scoped to clusters
- **CORS Support**: Cross-origin requests enabled for web clients

## Quick Start

```bash
# Build and run
go run .

# Or build binary
go build -o mock-server .
./mock-server
```

The server starts on `http://localhost:8080` by default.

## API Endpoints

### Authentication

- `POST /auth/token` - Get OAuth2 access token

### API Metadata

- `GET /api/v1` - API version and metadata
- `GET /health` - Health check

### Clusters

- `GET /api/v1/clusters` - List all clusters
- `POST /api/v1/clusters` - Create a cluster
- `GET /api/v1/clusters/{id}` - Get cluster details
- `PATCH /api/v1/clusters/{id}` - Update cluster
- `DELETE /api/v1/clusters/{id}` - Delete cluster

### Machine Pools

- `GET /api/v1/clusters/{id}/machine_pools` - List machine pools
- `POST /api/v1/clusters/{id}/machine_pools` - Create machine pool
- `GET /api/v1/clusters/{id}/machine_pools/{pool_id}` - Get machine pool
- `PATCH /api/v1/clusters/{id}/machine_pools/{pool_id}` - Update machine pool
- `DELETE /api/v1/clusters/{id}/machine_pools/{pool_id}` - Delete machine pool

### Identity Providers

- `GET /api/v1/clusters/{id}/identity_providers` - List IDPs
- `POST /api/v1/clusters/{id}/identity_providers` - Create IDP
- `GET /api/v1/clusters/{id}/identity_providers/{idp_id}` - Get IDP
- `DELETE /api/v1/clusters/{id}/identity_providers/{idp_id}` - Delete IDP

### Addons

- `GET /api/v1/clusters/{id}/addons` - List addon installations
- `POST /api/v1/clusters/{id}/addons` - Install addon
- `GET /api/v1/clusters/{id}/addons/{addon_id}` - Get addon
- `DELETE /api/v1/clusters/{id}/addons/{addon_id}` - Uninstall addon

### Upgrade Policies

- `GET /api/v1/clusters/{id}/upgrade_policies` - List upgrade policies
- `POST /api/v1/clusters/{id}/upgrade_policies` - Create upgrade policy
- `GET /api/v1/clusters/{id}/upgrade_policies/{policy_id}` - Get policy
- `DELETE /api/v1/clusters/{id}/upgrade_policies/{policy_id}` - Delete policy

## Usage Examples

### Get OAuth Token

```bash
curl -X POST http://localhost:8080/auth/token \
  -d 'grant_type=authorization_code&code=any-code-value'
```

Response:
```json
{
  "access_token": "abc123...",
  "token_type": "Bearer",
  "expires_in": 3600,
  "refresh_token": "def456...",
  "scope": "openid"
}
```

### Create a Cluster

```bash
TOKEN="your-access-token"

curl -X POST http://localhost:8080/api/v1/clusters \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "my-cluster",
    "region": "us-east-1",
    "multi_az": true,
    "version": "4.14.0"
  }'
```

### Create a Machine Pool

```bash
CLUSTER_ID="cluster-uuid"

curl -X POST "http://localhost:8080/api/v1/clusters/$CLUSTER_ID/machine_pools" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "instance_type": "m5.xlarge",
    "replicas": 3
  }'
```

### Refresh Token

```bash
REFRESH_TOKEN="your-refresh-token"

curl -X POST http://localhost:8080/auth/token \
  -d "grant_type=refresh_token&refresh_token=$REFRESH_TOKEN&client_id=test&client_secret=test"
```

## Testing

Run the comprehensive test script:

```bash
./test.sh
```

The test script requires `jq` for JSON formatting:
```bash
# macOS
brew install jq

# Linux
apt-get install jq
```

## Architecture

### Files

- `main.go` - HTTP server and routing
- `types.go` - Data structures for all resource types
- `store.go` - Thread-safe in-memory data store with CRUD operations
- `auth.go` - OAuth2 token handling and authentication middleware
- `test.sh` - Comprehensive API test script

### Data Store

The store uses a read-write mutex for thread safety and maintains indexes for efficient nested resource lookups:

- Clusters indexed by ID
- Machine pools indexed by ID and cluster ID
- Identity providers indexed by ID and cluster ID
- Addons indexed by ID and cluster ID
- Upgrade policies indexed by ID and cluster ID

### Authentication

The auth server:
- Generates random tokens using crypto/rand
- Validates bearer tokens from Authorization headers
- Supports authorization_code and refresh_token grant types
- Tokens expire after 1 hour (configurable)
- For testing, any authorization code is accepted

## Configuration

Currently all configuration is hardcoded:

- Port: `8080`
- Token lifetime: `1 hour`
- Refresh token lifetime: `24 hours`

## Limitations

This is a mock server for development and testing:

- Data is stored in memory (lost on restart)
- No database persistence
- No rate limiting
- No request validation beyond basic requirements
- No AWS integration
- Simplified OAuth2 flow (no client ID/secret validation for auth codes)

## License

Same as parent project (MIT)
