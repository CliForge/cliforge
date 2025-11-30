.PHONY: build mock-server demo test clean install run help validate quick

# Determine the alpha-omega repo path (where the main go.mod is)
ALPHA_OMEGA_ROOT := $(shell cd ../.. && pwd)

# Build the ROSA-like CLI
build:
	@echo "Building rosa CLI..."
	@cd $(ALPHA_OMEGA_ROOT) && go build -o examples/rosa-like/rosa ./examples/rosa-like/cmd/rosa
	@echo "✓ Built rosa binary at ./rosa"

# Install the CLI to GOPATH/bin
install:
	@echo "Installing rosa CLI..."
	@cd $(ALPHA_OMEGA_ROOT) && go install ./examples/rosa-like/cmd/rosa
	@echo "✓ Installed rosa to GOPATH/bin"

# Run the CLI (pass arguments with ARGS="...")
run: build
	@./rosa $(ARGS)

# Run the mock API server
mock-server:
	@echo "Starting mock API server on :8080..."
	@cd mock-server && go run .

# Start mock server in background
mock-server-bg:
	@echo "Starting mock API server in background..."
	@cd mock-server && go run . > /tmp/rosa-mock-server.log 2>&1 &
	@echo "✓ Mock server running (PID: $$!), logs: /tmp/rosa-mock-server.log"

# Stop background mock server
mock-server-stop:
	@echo "Stopping mock API server..."
	@pkill -f "go run.*mock-server" || true
	@echo "✓ Mock server stopped"

# Run the demo script
demo: build mock-server-bg
	@echo "Running demo..."
	@sleep 2  # Wait for server to start
	@./demo.sh || true
	@$(MAKE) mock-server-stop

# Run integration tests
test: build
	@echo "Running integration tests..."
	@cd tests && go test -v ./...

# Clean build artifacts
clean:
	@echo "Cleaning..."
	@rm -f rosa
	@rm -f cmd/rosa/rosa
	@rm -rf ~/.config/rosa  # Clean config/state
	@echo "✓ Cleaned"

# Verify OpenAPI spec
validate:
	@echo "Validating OpenAPI spec..."
	@echo "Checking api/rosa-api.yaml..."
	@test -f api/rosa-api.yaml || (echo "ERROR: api/rosa-api.yaml not found" && exit 1)
	@echo "✓ OpenAPI spec exists"

# Quick test - build and run version/help
quick: build
	@echo "Testing rosa binary..."
	@./rosa version
	@echo ""
	@./rosa --help

# Show help
help:
	@echo "ROSA-like CLI - Available targets:"
	@echo ""
	@echo "  build           - Build the rosa CLI binary"
	@echo "  install         - Install rosa to GOPATH/bin"
	@echo "  run            - Run the CLI (use ARGS='...' to pass arguments)"
	@echo "  mock-server    - Start mock API server (foreground)"
	@echo "  mock-server-bg - Start mock API server (background)"
	@echo "  mock-server-stop - Stop background mock server"
	@echo "  demo           - Run demo script"
	@echo "  test           - Run integration tests"
	@echo "  validate       - Validate OpenAPI spec"
	@echo "  quick          - Quick test (build + version + help)"
	@echo "  clean          - Remove build artifacts"
	@echo ""
	@echo "Examples:"
	@echo "  make build"
	@echo "  make run ARGS='--help'"
	@echo "  make run ARGS='list clusters'"
	@echo "  make mock-server"
