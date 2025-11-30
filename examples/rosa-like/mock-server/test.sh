#!/bin/bash

# Simple test script for the mock API server
# Assumes server is running on localhost:8080

set -e

API_BASE="http://localhost:8080"
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}=== Testing Mock ROSA API Server ===${NC}\n"

# 1. Test health check
echo -e "${GREEN}1. Health check${NC}"
curl -s "$API_BASE/health" | jq .
echo ""

# 2. Test API metadata
echo -e "${GREEN}2. API metadata${NC}"
curl -s "$API_BASE/api/v1" | jq .
echo ""

# 3. Get OAuth token
echo -e "${GREEN}3. Get OAuth2 token${NC}"
TOKEN_RESPONSE=$(curl -s -X POST "$API_BASE/auth/token" -d 'grant_type=authorization_code&code=test-code')
echo "$TOKEN_RESPONSE" | jq .
TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.access_token')
REFRESH_TOKEN=$(echo "$TOKEN_RESPONSE" | jq -r '.refresh_token')
echo ""

# 4. Test refresh token
echo -e "${GREEN}4. Refresh token${NC}"
curl -s -X POST "$API_BASE/auth/token" \
  -d "grant_type=refresh_token&refresh_token=$REFRESH_TOKEN&client_id=test&client_secret=test" | jq .
echo ""

# 5. Create a cluster
echo -e "${GREEN}5. Create cluster${NC}"
CLUSTER=$(curl -s -X POST "$API_BASE/api/v1/clusters" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "production-cluster",
    "region": "us-east-1",
    "multi_az": true,
    "version": "4.14.0"
  }')
echo "$CLUSTER" | jq .
CLUSTER_ID=$(echo "$CLUSTER" | jq -r '.id')
echo ""

# 6. List clusters
echo -e "${GREEN}6. List clusters${NC}"
curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/api/v1/clusters" | jq .
echo ""

# 7. Get cluster
echo -e "${GREEN}7. Get cluster details${NC}"
curl -s -H "Authorization: Bearer $TOKEN" "$API_BASE/api/v1/clusters/$CLUSTER_ID" | jq .
echo ""

# 8. Update cluster
echo -e "${GREEN}8. Update cluster${NC}"
curl -s -X PATCH "$API_BASE/api/v1/clusters/$CLUSTER_ID" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"version": "4.15.0"}' | jq .
echo ""

# 9. Create machine pool
echo -e "${GREEN}9. Create machine pool${NC}"
POOL=$(curl -s -X POST "$API_BASE/api/v1/clusters/$CLUSTER_ID/machine_pools" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "instance_type": "m5.2xlarge",
    "replicas": 5,
    "autoscaling_enabled": false
  }')
echo "$POOL" | jq .
POOL_ID=$(echo "$POOL" | jq -r '.id')
echo ""

# 10. List machine pools
echo -e "${GREEN}10. List machine pools${NC}"
curl -s -H "Authorization: Bearer $TOKEN" \
  "$API_BASE/api/v1/clusters/$CLUSTER_ID/machine_pools" | jq .
echo ""

# 11. Create identity provider
echo -e "${GREEN}11. Create identity provider${NC}"
IDP=$(curl -s -X POST "$API_BASE/api/v1/clusters/$CLUSTER_ID/identity_providers" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "github",
    "name": "github-corporate",
    "mapping_method": "claim",
    "config": {"organizations": ["my-org"]}
  }')
echo "$IDP" | jq .
IDP_ID=$(echo "$IDP" | jq -r '.id')
echo ""

# 12. List identity providers
echo -e "${GREEN}12. List identity providers${NC}"
curl -s -H "Authorization: Bearer $TOKEN" \
  "$API_BASE/api/v1/clusters/$CLUSTER_ID/identity_providers" | jq .
echo ""

# 13. Install addon
echo -e "${GREEN}13. Install addon${NC}"
ADDON=$(curl -s -X POST "$API_BASE/api/v1/clusters/$CLUSTER_ID/addons" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "addon_id": "cluster-logging-operator",
    "parameters": {"retention_days": 7}
  }')
echo "$ADDON" | jq .
ADDON_ID=$(echo "$ADDON" | jq -r '.id')
echo ""

# 14. List addons
echo -e "${GREEN}14. List addons${NC}"
curl -s -H "Authorization: Bearer $TOKEN" \
  "$API_BASE/api/v1/clusters/$CLUSTER_ID/addons" | jq .
echo ""

# 15. Create upgrade policy
echo -e "${GREEN}15. Create upgrade policy${NC}"
UPGRADE=$(curl -s -X POST "$API_BASE/api/v1/clusters/$CLUSTER_ID/upgrade_policies" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "version": "4.15.0",
    "schedule_type": "automatic",
    "enable_minor_version_upgrades": true
  }')
echo "$UPGRADE" | jq .
UPGRADE_ID=$(echo "$UPGRADE" | jq -r '.id')
echo ""

# 16. List upgrade policies
echo -e "${GREEN}16. List upgrade policies${NC}"
curl -s -H "Authorization: Bearer $TOKEN" \
  "$API_BASE/api/v1/clusters/$CLUSTER_ID/upgrade_policies" | jq .
echo ""

# 17. Delete machine pool
echo -e "${GREEN}17. Delete machine pool${NC}"
curl -s -X DELETE -H "Authorization: Bearer $TOKEN" \
  "$API_BASE/api/v1/clusters/$CLUSTER_ID/machine_pools/$POOL_ID" \
  -w "HTTP Status: %{http_code}\n"
echo ""

# 18. Delete cluster
echo -e "${GREEN}18. Delete cluster${NC}"
curl -s -X DELETE -H "Authorization: Bearer $TOKEN" \
  "$API_BASE/api/v1/clusters/$CLUSTER_ID" \
  -w "HTTP Status: %{http_code}\n"
echo ""

# 19. Test authentication failure
echo -e "${GREEN}19. Test authentication failure${NC}"
curl -s "$API_BASE/api/v1/clusters" | jq .
echo ""

echo -e "${BLUE}=== All tests completed ===${NC}"
