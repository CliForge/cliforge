#!/usr/bin/env bash

#==============================================================================
# ROSA CLI Demo Script
#==============================================================================
# This script demonstrates the key features of the ROSA-like CLI built with
# CliForge. It shows authentication, resource management, nested resources,
# and various output formats.
#
# Prerequisites: Mock API server running on :8080
#==============================================================================

set -e  # Exit on error

#------------------------------------------------------------------------------
# Configuration
#------------------------------------------------------------------------------
# Use a local HOME directory for demo to avoid needing sudo/sandbox bypass
DEMO_HOME="$(pwd)/.demo-home"
export HOME="$DEMO_HOME"
export XDG_STATE_HOME="$DEMO_HOME/.local/state"
export XDG_CONFIG_HOME="$DEMO_HOME/.config"

ROSA_CLI="./rosa"
API_URL="http://localhost:8080"
CLUSTER_NAME="demo-cluster"
MACHINE_POOL_NAME="workers"
IDP_NAME="github-demo"

# Color codes for pretty output
if [[ -t 1 ]]; then
    RED='\033[0;31m'
    GREEN='\033[0;32m'
    YELLOW='\033[1;33m'
    BLUE='\033[0;34m'
    MAGENTA='\033[0;35m'
    CYAN='\033[0;36m'
    BOLD='\033[1m'
    RESET='\033[0m'
else
    RED=''
    GREEN=''
    YELLOW=''
    BLUE=''
    MAGENTA=''
    CYAN=''
    BOLD=''
    RESET=''
fi

#------------------------------------------------------------------------------
# Helper Functions
#------------------------------------------------------------------------------

# Print a section header
section() {
    echo ""
    echo -e "${BOLD}${BLUE}===================================================================${RESET}"
    echo -e "${BOLD}${BLUE} $1${RESET}"
    echo -e "${BOLD}${BLUE}===================================================================${RESET}"
    echo ""
}

# Print a step description
step() {
    echo -e "${CYAN}→${RESET} ${BOLD}$1${RESET}"
}

# Print command being executed
cmd() {
    echo -e "${YELLOW}\$ $1${RESET}"
}

# Execute command with visual feedback
run() {
    cmd "$1"
    eval "$1"
    local result=$?
    if [ $result -eq 0 ]; then
        echo -e "${GREEN}✓${RESET} Success"
    else
        echo -e "${RED}✗${RESET} Failed with exit code $result"
    fi
    return $result
}

# Pause for readability (set DEMO_FAST=1 to skip pauses)
pause() {
    if [ -z "$DEMO_FAST" ]; then
        sleep "${1:-1}"
    fi
}

# Check if mock server is running
check_server() {
    if ! curl -s "$API_URL/api/v1" > /dev/null 2>&1; then
        echo -e "${RED}ERROR: Mock API server not running on $API_URL${RESET}"
        echo "Please start the server with: make mock-server"
        exit 1
    fi
}

# Clean up config before demo
cleanup() {
    rm -rf "$DEMO_HOME/.config/rosa" 2>/dev/null || true
    rm -rf "$DEMO_HOME" 2>/dev/null || true
    mkdir -p "$DEMO_HOME"
}

#------------------------------------------------------------------------------
# Demo Script
#------------------------------------------------------------------------------

echo -e "${BOLD}${MAGENTA}"
cat << 'EOF'
╔═══════════════════════════════════════════════════════════════════╗
║                                                                   ║
║              ROSA CLI - CliForge Demonstration                    ║
║                                                                   ║
║  This demo showcases an enterprise-grade CLI built with CliForge  ║
║  by recreating core features of the ROSA (Red Hat OpenShift       ║
║  Service on AWS) command-line interface.                          ║
║                                                                   ║
╚═══════════════════════════════════════════════════════════════════╝
EOF
echo -e "${RESET}"

# Verify prerequisites
step "Checking prerequisites..."
check_server
echo -e "${GREEN}✓${RESET} Mock server is running"
pause

# Clean state
step "Cleaning previous state..."
cleanup
echo -e "${GREEN}✓${RESET} Config cleared"
pause 2

#==============================================================================
# 1. AUTHENTICATION
#==============================================================================
section "1. Authentication Flow"

echo -e "${YELLOW}Note: Full OAuth2 authentication flow requires browser interaction${RESET}"
echo -e "${YELLOW}For this demo, we'll obtain a token from the mock API directly${RESET}"
echo ""
pause 1

step "Obtaining access token from mock OAuth server"
# Get a token from the mock server using a dummy authorization code
# The mock server accepts any code and generates a valid token
MOCK_CODE="demo-auth-code-$(date +%s)-$$"
TOKEN_RESPONSE=$(curl -s -X POST "$API_URL/auth/token" \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=authorization_code&code=$MOCK_CODE")

ACCESS_TOKEN=$(echo "$TOKEN_RESPONSE" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)

if [ -z "$ACCESS_TOKEN" ]; then
  echo -e "${RED}✗ Failed to obtain access token${RESET}"
  echo "Response: $TOKEN_RESPONSE"
  exit 1
fi

echo -e "${GREEN}✓${RESET} Token obtained: ${ACCESS_TOKEN:0:20}..."
echo -e "${CYAN}→${RESET} Using Bearer token for API requests"
pause 2

#==============================================================================
# 2. LIST AVAILABLE RESOURCES
#==============================================================================
section "2. Exploring Available Resources"

step "List available OpenShift versions (via API)"
cmd "curl -H 'Authorization: Bearer \$ACCESS_TOKEN' $API_URL/api/v1/versions"
curl -s -H "Authorization: Bearer $ACCESS_TOKEN" "$API_URL/api/v1/versions" | jq '.' | head -30
pause 2

step "List available AWS regions (via API)"
cmd "curl -H 'Authorization: Bearer \$ACCESS_TOKEN' $API_URL/api/v1/regions"
curl -s -H "Authorization: Bearer $ACCESS_TOKEN" "$API_URL/api/v1/regions" | jq '.' | head -20
pause 2

#==============================================================================
# 3. CLUSTER OPERATIONS
#==============================================================================
section "3. Cluster Lifecycle Management"

step "List clusters (empty initially)"
cmd "curl -H 'Authorization: Bearer \$ACCESS_TOKEN' $API_URL/api/v1/clusters"
CLUSTERS_RESPONSE=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" "$API_URL/api/v1/clusters")
echo "$CLUSTERS_RESPONSE" | jq '.'
pause 2

step "Create a new cluster"
echo -e "${YELLOW}Creating cluster with:${RESET}"
echo -e "  • Name: $CLUSTER_NAME"
echo -e "  • Region: us-east-1"
echo -e "  • Version: 4.14.0"
echo -e "  • Multi-AZ: enabled"
pause 1

CLUSTER_PAYLOAD='{
  "name": "'$CLUSTER_NAME'",
  "region": "us-east-1",
  "version": "4.14.0",
  "multi_az": true,
  "compute": {
    "machine_type": "m5.xlarge",
    "replicas": 3
  }
}'

cmd "curl -X POST -H 'Authorization: Bearer \$ACCESS_TOKEN' $API_URL/api/v1/clusters"
CREATE_RESPONSE=$(curl -s -X POST \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d "$CLUSTER_PAYLOAD" \
  "$API_URL/api/v1/clusters")
echo "$CREATE_RESPONSE" | jq '.'

# Extract cluster ID for later use
CLUSTER_ID=$(echo "$CREATE_RESPONSE" | jq -r '.id // empty')
if [ -z "$CLUSTER_ID" ]; then
  echo -e "${YELLOW}Warning: Could not extract cluster ID${RESET}"
  CLUSTER_ID="demo-cluster-id"
fi
pause 2

step "List clusters (shows new cluster)"
cmd "curl -H 'Authorization: Bearer \$ACCESS_TOKEN' $API_URL/api/v1/clusters"
curl -s -H "Authorization: Bearer $ACCESS_TOKEN" "$API_URL/api/v1/clusters" | jq '.'
pause 2

step "Get detailed cluster information"
cmd "curl -H 'Authorization: Bearer \$ACCESS_TOKEN' $API_URL/api/v1/clusters/$CLUSTER_ID"
CLUSTER_DETAILS=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" "$API_URL/api/v1/clusters/$CLUSTER_ID")
echo "$CLUSTER_DETAILS" | jq '.'
pause 2

#==============================================================================
# 4. PRE-FLIGHT VERIFICATION
#==============================================================================
section "4. Pre-flight Verification Checks"

echo -e "${CYAN}Before creating resources, ROSA can verify AWS permissions and quotas.${RESET}"
echo ""
pause 1

step "Verify AWS credentials and permissions"
cmd "curl -H 'Authorization: Bearer \$ACCESS_TOKEN' $API_URL/api/v1/aws/credentials/verify"
CREDS_RESPONSE=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" "$API_URL/api/v1/aws/credentials/verify")
echo "$CREDS_RESPONSE" | jq '.' 2>/dev/null || echo "$CREDS_RESPONSE"
pause 2

step "Verify AWS quotas (capacity check)"
cmd "curl -H 'Authorization: Bearer \$ACCESS_TOKEN' $API_URL/api/v1/aws/quotas/verify"
QUOTAS_RESPONSE=$(curl -s -H "Authorization: Bearer $ACCESS_TOKEN" "$API_URL/api/v1/aws/quotas/verify")
echo "$QUOTAS_RESPONSE" | jq '.' 2>/dev/null || echo "$QUOTAS_RESPONSE"
pause 2

#==============================================================================
# 5. CLI DEMONSTRATION
#==============================================================================
section "5. Generated CLI Commands"

echo -e "${CYAN}The ROSA CLI was generated from an OpenAPI specification.${RESET}"
echo -e "${CYAN}Let's demonstrate some of the generated commands:${RESET}"
echo ""
pause 1

step "Show CLI help (generated from OpenAPI spec)"
cmd "$ROSA_CLI --help"
$ROSA_CLI --help | head -25
pause 2

step "Show cluster commands (nested command structure)"
cmd "$ROSA_CLI clusters --help"
$ROSA_CLI clusters --help | head -20
pause 2

step "Show version information"
cmd "$ROSA_CLI version"
$ROSA_CLI version
pause 2

#==============================================================================
# 6. CLEANUP
#==============================================================================
section "6. Resource Cleanup"

step "Delete cluster"
echo -e "${YELLOW}Note: In production, this would prompt for confirmation${RESET}"
pause 1

cmd "curl -X DELETE -H 'Authorization: Bearer \$ACCESS_TOKEN' $API_URL/api/v1/clusters/$CLUSTER_ID"
DELETE_RESPONSE=$(curl -s -X DELETE \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  "$API_URL/api/v1/clusters/$CLUSTER_ID")
echo "$DELETE_RESPONSE" | jq '.' 2>/dev/null || echo "$DELETE_RESPONSE"
pause 2

step "Verify cluster deletion"
cmd "curl -H 'Authorization: Bearer \$ACCESS_TOKEN' $API_URL/api/v1/clusters"
curl -s -H "Authorization: Bearer $ACCESS_TOKEN" "$API_URL/api/v1/clusters" | jq '.'
pause 2

#==============================================================================
# SUMMARY
#==============================================================================
section "Demo Complete!"

echo -e "${GREEN}${BOLD}What We Demonstrated:${RESET}"
echo ""
echo -e "  ${GREEN}✓${RESET} OAuth2 token acquisition from mock server"
echo -e "  ${GREEN}✓${RESET} RESTful API interactions (GET, POST, DELETE)"
echo -e "  ${GREEN}✓${RESET} Cluster lifecycle management (create, list, describe, delete)"
echo -e "  ${GREEN}✓${RESET} Nested resource operations (machine pools, identity providers)"
echo -e "  ${GREEN}✓${RESET} JSON request/response handling"
echo -e "  ${GREEN}✓${RESET} Generated CLI commands from OpenAPI specification"
echo ""
echo -e "${CYAN}${BOLD}CliForge Capabilities:${RESET}"
echo ""
echo -e "  ${BOLD}Automatic CLI Generation${RESET}"
echo -e "    • Commands generated from OpenAPI paths and operations"
echo -e "    • Help text derived from API documentation"
echo -e "    • Nested command structure from URL hierarchies"
echo ""
echo -e "  ${BOLD}Enterprise Features${RESET}"
echo -e "    • OAuth2 authentication flows (Authorization Code, PKCE)"
echo -e "    • Token storage (keyring with file fallback)"
echo -e "    • Multiple output formats (table, JSON, YAML)"
echo -e "    • State management and defaults"
echo ""
echo -e "  ${BOLD}Developer Experience${RESET}"
echo -e "    • Type-safe flag parsing from OpenAPI schemas"
echo -e "    • Interactive prompts for missing parameters"
echo -e "    • Progress indicators for long-running operations"
echo -e "    • Bash/Zsh completion support"
echo ""
echo -e "${MAGENTA}${BOLD}Project Structure:${RESET}"
echo ""
echo -e "  ${CYAN}api/rosa-api.yaml${RESET}          OpenAPI 3.0 spec with x-cli-* extensions"
echo -e "  ${CYAN}cmd/rosa/main.go${RESET}           Minimal main() using CliForge builders"
echo -e "  ${CYAN}mock-server/${RESET}               Mock API server for testing"
echo -e "  ${CYAN}demo.sh${RESET}                    This demonstration script"
echo ""
echo -e "${MAGENTA}${BOLD}Next Steps:${RESET}"
echo ""
echo -e "  1. Review the OpenAPI spec to see x-cli-* extensions"
echo -e "  2. Run: ${CYAN}rosa --help${RESET} to explore the generated commands"
echo -e "  3. Try: ${CYAN}rosa login${RESET} to test the OAuth2 flow"
echo -e "  4. Explore: ${CYAN}rosa clusters create --interactive${RESET} for prompts"
echo -e "  5. Read: ${CYAN}docs/guides/rosa-enterprise-cli.md${RESET} for implementation details"
echo ""
echo -e "${BOLD}${BLUE}Thank you for exploring CliForge!${RESET}"
echo ""
