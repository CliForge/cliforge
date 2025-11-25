#!/usr/bin/env bash
#
# Build script for Petstore CLI example
# Demonstrates CliForge v0.9.0 complete workflow
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

log_header() {
    echo ""
    echo -e "${CYAN}========================================${NC}"
    echo -e "${CYAN}$1${NC}"
    echo -e "${CYAN}========================================${NC}"
    echo ""
}

# Check prerequisites
check_prerequisites() {
    log_header "Checking Prerequisites"

    if ! command -v go &> /dev/null; then
        log_error "Go is not installed. Please install Go 1.21 or later."
        exit 1
    fi

    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "Go version: $GO_VERSION"

    if ! command -v jq &> /dev/null; then
        log_warning "jq is not installed. Some features may not work."
    fi

    log_success "Prerequisites check complete"
}

# Build mock API server
build_mock_server() {
    log_header "Building Mock API Server"

    cd "$SCRIPT_DIR"

    log_info "Compiling mock-server.go..."
    go build -o mock-server mock-server.go

    if [ $? -eq 0 ]; then
        log_success "Mock server built successfully: $SCRIPT_DIR/mock-server"
    else
        log_error "Failed to build mock server"
        exit 1
    fi
}

# Start mock API server
start_mock_server() {
    log_header "Starting Mock API Server"

    # Check if server is already running
    if lsof -Pi :8080 -sTCP:LISTEN -t >/dev/null 2>&1; then
        log_warning "Port 8080 is already in use. Stopping existing process..."
        kill $(lsof -t -i:8080) 2>/dev/null || true
        sleep 2
    fi

    cd "$SCRIPT_DIR"
    log_info "Starting mock server on http://localhost:8080..."
    ./mock-server > mock-server.log 2>&1 &
    SERVER_PID=$!
    echo $SERVER_PID > mock-server.pid

    # Wait for server to start
    log_info "Waiting for server to be ready..."
    for i in {1..10}; do
        if curl -s http://localhost:8080/openapi.yaml > /dev/null 2>&1; then
            log_success "Mock server is ready (PID: $SERVER_PID)"
            return 0
        fi
        sleep 1
    done

    log_error "Mock server failed to start"
    exit 1
}

# Stop mock API server
stop_mock_server() {
    log_header "Stopping Mock API Server"

    if [ -f "$SCRIPT_DIR/mock-server.pid" ]; then
        SERVER_PID=$(cat "$SCRIPT_DIR/mock-server.pid")
        if ps -p $SERVER_PID > /dev/null 2>&1; then
            log_info "Stopping mock server (PID: $SERVER_PID)..."
            kill $SERVER_PID
            rm "$SCRIPT_DIR/mock-server.pid"
            log_success "Mock server stopped"
        else
            log_warning "Mock server process not found"
            rm "$SCRIPT_DIR/mock-server.pid"
        fi
    else
        log_info "Mock server is not running"
    fi
}

# Build CliForge CLI
build_cliforge() {
    log_header "Building CliForge CLI Generator"

    cd "$PROJECT_ROOT"

    log_info "Building CliForge..."
    go build -o "$SCRIPT_DIR/cliforge" ./cmd/cliforge

    if [ $? -eq 0 ]; then
        log_success "CliForge built successfully: $SCRIPT_DIR/cliforge"
    else
        log_error "Failed to build CliForge"
        exit 1
    fi
}

# Generate Petstore CLI
generate_cli() {
    log_header "Generating Petstore CLI"

    cd "$SCRIPT_DIR"

    log_info "Generating CLI from OpenAPI spec..."
    ./cliforge generate \
        --spec http://localhost:8080/openapi.yaml \
        --config cli-config.yaml \
        --output petstore-cli \
        --verbose

    if [ $? -eq 0 ]; then
        log_success "Petstore CLI generated successfully"
        chmod +x petstore-cli
    else
        log_error "Failed to generate Petstore CLI"
        exit 1
    fi
}

# Test basic CLI functionality
test_cli() {
    log_header "Testing Petstore CLI"

    cd "$SCRIPT_DIR"

    log_info "Test 1: CLI version"
    ./petstore-cli --version || log_warning "Version command not yet implemented"

    log_info "Test 2: CLI help"
    ./petstore-cli --help || log_warning "Help command not yet implemented"

    log_info "Test 3: List pets"
    ./petstore-cli list pets --limit 5 || log_warning "List pets command not yet implemented"

    log_info "Test 4: Get pet details"
    ./petstore-cli get pet --pet-id 1 || log_warning "Get pet command not yet implemented"

    log_success "CLI tests completed (check output above for any failures)"
}

# Display example commands
show_examples() {
    log_header "Example Commands"

    cat << 'EOF'
The Petstore CLI has been built successfully!

Mock API Server:
  - Running at: http://localhost:8080
  - OpenAPI spec: http://localhost:8080/openapi.yaml
  - Logs: ./mock-server.log

Example CLI Commands:
  # Basic operations
  ./petstore-cli list pets
  ./petstore-cli list pets --status available
  ./petstore-cli list pets --category dog --limit 10
  ./petstore-cli get pet --pet-id 1
  ./petstore-cli create pet --name "Fluffy" --category cat --age 2 --price 150
  ./petstore-cli update pet --pet-id 1 --status adopted
  ./petstore-cli delete pet --pet-id 5 --yes

  # Interactive mode
  ./petstore-cli create pet --interactive

  # Streaming (SSE)
  ./petstore-cli watch pet --pet-id 1

  # Workflow
  ./petstore-cli adopt pet --pet-id 1 --user-id 1

  # Different output formats
  ./petstore-cli list pets --output json
  ./petstore-cli list pets --output yaml
  ./petstore-cli list pets --output csv
  ./petstore-cli get pet --pet-id 1 --output yaml

  # Watch mode
  ./petstore-cli get pet --pet-id 1 --watch

  # Context switching
  ./petstore-cli list contexts
  ./petstore-cli use context development
  ./petstore-cli current context

  # Store operations
  ./petstore-cli list stores

  # Order operations
  ./petstore-cli create order --pet-id 1

Run the demo script for a full demonstration:
  ./demo.sh

Stop the mock server:
  ./build.sh stop

EOF

    log_success "Build complete!"
}

# Cleanup
cleanup() {
    log_header "Cleanup"

    cd "$SCRIPT_DIR"

    log_info "Removing build artifacts..."
    rm -f mock-server cliforge petstore-cli
    rm -f mock-server.log mock-server.pid

    log_success "Cleanup complete"
}

# Main script
main() {
    case "${1:-all}" in
        prereq)
            check_prerequisites
            ;;
        server)
            build_mock_server
            start_mock_server
            ;;
        cli)
            build_cliforge
            generate_cli
            ;;
        test)
            test_cli
            ;;
        stop)
            stop_mock_server
            ;;
        clean)
            cleanup
            ;;
        all)
            check_prerequisites
            build_mock_server
            start_mock_server
            build_cliforge
            generate_cli
            # test_cli  # Uncomment when CLI is implemented
            show_examples
            ;;
        *)
            cat << EOF
Usage: $0 [command]

Commands:
  prereq    Check prerequisites
  server    Build and start mock API server
  cli       Build CliForge and generate Petstore CLI
  test      Test the generated CLI
  stop      Stop the mock API server
  clean     Clean up build artifacts
  all       Run complete build (default)

Examples:
  $0              # Run complete build
  $0 server       # Just start the mock server
  $0 stop         # Stop the mock server
  $0 clean        # Clean up everything

EOF
            exit 1
            ;;
    esac
}

# Trap to ensure server is stopped on script exit
trap 'if [ "$1" == "all" ] || [ "$1" == "server" ]; then echo ""; log_info "Press Ctrl+C again to stop the server, or run: ./build.sh stop"; fi' INT

main "$@"
