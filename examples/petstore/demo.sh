#!/usr/bin/env bash
#
# Demo script for Petstore CLI example
# Demonstrates all CliForge v0.9.0 features interactively
#

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
MAGENTA='\033[0;35m'
NC='\033[0m' # No Color
BOLD='\033[1m'

# Demo control
PAUSE_BETWEEN_COMMANDS=2
INTERACTIVE=${INTERACTIVE:-true}

pause() {
    if [ "$INTERACTIVE" = "true" ]; then
        echo ""
        echo -ne "${CYAN}Press ENTER to continue...${NC}"
        read
        echo ""
    else
        sleep $PAUSE_BETWEEN_COMMANDS
    fi
}

print_header() {
    echo ""
    echo -e "${BOLD}${MAGENTA}╔══════════════════════════════════════════════════════════════╗${NC}"
    printf "${BOLD}${MAGENTA}║${NC} ${BOLD}%-60s${NC} ${BOLD}${MAGENTA}║${NC}\n" "$1"
    echo -e "${BOLD}${MAGENTA}╚══════════════════════════════════════════════════════════════╝${NC}"
    echo ""
}

print_section() {
    echo ""
    echo -e "${BOLD}${CYAN}▶ $1${NC}"
    echo -e "${CYAN}────────────────────────────────────────────────────────${NC}"
    echo ""
}

print_command() {
    echo -e "${YELLOW}\$ $1${NC}"
}

run_demo_command() {
    local cmd="$1"
    local description="$2"

    if [ -n "$description" ]; then
        echo -e "${BLUE}# $description${NC}"
    fi

    print_command "$cmd"

    # Note: These commands won't actually work until CliForge generates the CLI
    # For now, we'll show what the output would look like
    echo -e "${GREEN}(Command execution simulated - CLI generation not yet implemented)${NC}"
    echo ""
}

show_banner() {
    clear
    cat << 'EOF'
╔══════════════════════════════════════════════════════════════════════╗
║                                                                      ║
║   ██████╗ ███████╗████████╗███████╗████████╗ ██████╗ ██████╗ ███████╗║
║   ██╔══██╗██╔════╝╚══██╔══╝██╔════╝╚══██╔══╝██╔═══██╗██╔══██╗██╔════╝║
║   ██████╔╝█████╗     ██║   ███████╗   ██║   ██║   ██║██████╔╝█████╗  ║
║   ██╔═══╝ ██╔══╝     ██║   ╚════██║   ██║   ██║   ██║██╔══██╗██╔══╝  ║
║   ██║     ███████╗   ██║   ███████║   ██║   ╚██████╔╝██║  ██║███████╗║
║   ╚═╝     ╚══════╝   ╚═╝   ╚══════╝   ╚═╝    ╚═════╝ ╚═╝  ╚═╝╚══════╝║
║                                                                      ║
║              CliForge v0.9.0 - Complete Feature Demo                ║
║                                                                      ║
╚══════════════════════════════════════════════════════════════════════╝

EOF

    echo -e "${BOLD}Welcome to the Petstore CLI demonstration!${NC}"
    echo ""
    echo "This demo showcases ALL features of CliForge v0.9.0:"
    echo "  • Complete OpenAPI spec with all x-cli-* extensions"
    echo "  • Branding and customization"
    echo "  • Interactive mode and prompts"
    echo "  • Multiple output formats (table, JSON, YAML, CSV)"
    echo "  • Watch mode and streaming (SSE)"
    echo "  • Workflow orchestration"
    echo "  • Async operations with polling"
    echo "  • Pre-flight validation"
    echo "  • Context switching"
    echo "  • Plugin integration"
    echo "  • Deprecation warnings"
    echo "  • And more..."
    echo ""

    pause
}

demo_basic_operations() {
    print_header "1. BASIC OPERATIONS"

    print_section "List all pets"
    run_demo_command \
        "./petstore-cli list pets" \
        "Default table output with colors and formatting"
    pause

    print_section "List with filters"
    run_demo_command \
        "./petstore-cli list pets --status available --category dog" \
        "Filter pets by status and category"
    pause

    print_section "Get specific pet"
    run_demo_command \
        "./petstore-cli get pet --pet-id 1" \
        "Detailed YAML output for single pet"
    pause
}

demo_output_formats() {
    print_header "2. OUTPUT FORMATS"

    print_section "JSON output"
    run_demo_command \
        "./petstore-cli list pets --output json --limit 3" \
        "Pretty-printed JSON with syntax highlighting"
    pause

    print_section "YAML output"
    run_demo_command \
        "./petstore-cli list pets --output yaml --limit 3" \
        "YAML format for easy readability"
    pause

    print_section "CSV output"
    run_demo_command \
        "./petstore-cli list pets --output csv --limit 5" \
        "CSV format for spreadsheets"
    pause

    print_section "Custom table columns"
    run_demo_command \
        "./petstore-cli list pets --columns id,name,status,price" \
        "Select specific columns to display"
    pause
}

demo_create_operations() {
    print_header "3. CREATE OPERATIONS"

    print_section "Create pet with flags"
    run_demo_command \
        "./petstore-cli create pet --name \"Max\" --category dog --age 3 --price 300" \
        "Create a new pet using command-line flags"
    pause

    print_section "Create pet interactively"
    echo -e "${BLUE}# Interactive mode with prompts and validation${NC}"
    print_command "./petstore-cli create pet --interactive"
    echo ""
    echo -e "${CYAN}? What is the pet's name?${NC} Bella"
    echo -e "${CYAN}? Select pet category:${NC}"
    echo "  > Dog 🐕"
    echo "    Cat 🐈"
    echo "    Bird 🐦"
    echo "    Fish 🐟"
    echo "    Reptile 🦎"
    echo -e "${CYAN}? How old is the pet (in years)?${NC} 2"
    echo -e "${CYAN}? What is the price?${NC} \$200.00"
    echo -e "${CYAN}? What is the pet's status?${NC} Available for adoption"
    echo ""
    echo -e "${GREEN}✓ Checking store capacity...${NC}"
    echo -e "${GREEN}✓ Checking for duplicate pet names...${NC}"
    echo -e "${GREEN}✓ Pet 'Bella' created successfully with ID 101${NC}"
    echo ""
    pause
}

demo_update_delete() {
    print_header "4. UPDATE & DELETE OPERATIONS"

    print_section "Update pet status"
    run_demo_command \
        "./petstore-cli update pet --pet-id 1 --status adopted" \
        "Update a pet's status"
    pause

    print_section "Delete with confirmation"
    echo -e "${BLUE}# Deletion requires confirmation${NC}"
    print_command "./petstore-cli delete pet --pet-id 5"
    echo ""
    echo -e "${YELLOW}⚠ Are you sure you want to delete pet '5'? This cannot be undone. [y/N]:${NC} y"
    echo -e "${GREEN}✓ Pet 5 deleted successfully${NC}"
    echo ""
    pause

    print_section "Delete without confirmation"
    run_demo_command \
        "./petstore-cli delete pet --pet-id 5 --yes" \
        "Skip confirmation with --yes flag"
    pause
}

demo_watch_mode() {
    print_header "5. WATCH MODE & STREAMING"

    print_section "Watch pet status changes"
    echo -e "${BLUE}# Real-time updates using watch mode${NC}"
    print_command "./petstore-cli get pet --pet-id 1 --watch --interval 5"
    echo ""
    echo -e "${CYAN}[00:00]${NC} Status: available, Updated: 2024-11-23 10:30:00"
    echo -e "${CYAN}[00:05]${NC} Status: pending, Updated: 2024-11-23 10:30:05 ${YELLOW}(CHANGED)${NC}"
    echo -e "${CYAN}[00:10]${NC} Status: pending, Updated: 2024-11-23 10:30:05"
    echo -e "${CYAN}[00:15]${NC} Status: adopted, Updated: 2024-11-23 10:30:15 ${YELLOW}(CHANGED)${NC}"
    echo -e "${CYAN}^C${NC}"
    echo ""
    pause

    print_section "Stream pet status (SSE)"
    echo -e "${BLUE}# Real-time streaming using Server-Sent Events${NC}"
    print_command "./petstore-cli watch pet --pet-id 1"
    echo ""
    echo -e "${GREEN}[2024-11-23 10:30:00] status-change: Current status: available${NC}"
    echo -e "${BLUE}[2024-11-23 10:30:05] location-update: Location updated${NC}"
    echo -e "${GREEN}[2024-11-23 10:30:10] health-check: Pet is doing well${NC}"
    echo -e "${GREEN}[2024-11-23 10:30:15] status-change: Feeding completed${NC}"
    echo -e "${BLUE}[2024-11-23 10:30:20] location-update: Exercise time${NC}"
    echo -e "${CYAN}^C${NC}"
    echo ""
    pause
}

demo_workflow() {
    print_header "6. WORKFLOW ORCHESTRATION"

    print_section "Multi-step pet adoption workflow"
    echo -e "${BLUE}# Automated workflow with multiple API calls${NC}"
    print_command "./petstore-cli adopt pet --pet-id 1 --user-id 1"
    echo ""
    echo -e "${CYAN}▶ Step 1/6: Checking pet availability...${NC}"
    echo -e "${GREEN}  ✓ Pet 'Buddy' is available for adoption${NC}"
    echo ""
    echo -e "${CYAN}▶ Step 2/6: Verifying user eligibility...${NC}"
    echo -e "${GREEN}  ✓ User 'John Doe' is eligible${NC}"
    echo ""
    echo -e "${CYAN}▶ Step 3/6: Creating adoption order...${NC}"
    echo -e "${GREEN}  ✓ Order #102 created${NC}"
    echo ""
    echo -e "${CYAN}▶ Step 4/6: Updating pet status...${NC}"
    echo -e "${GREEN}  ✓ Pet status updated to 'adopted'${NC}"
    echo ""
    echo -e "${CYAN}▶ Step 5/6: Sending adoption notification...${NC}"
    echo -e "${GREEN}  ✓ Email sent to john@example.com${NC}"
    echo ""
    echo -e "${CYAN}▶ Step 6/6: Waiting for admin approval...${NC}"
    echo -e "${YELLOW}  ⏳ Polling order status (0s elapsed)${NC}"
    echo -e "${YELLOW}  ⏳ Polling order status (10s elapsed)${NC}"
    echo -e "${GREEN}  ✓ Order approved!${NC}"
    echo ""
    echo -e "${GREEN}Workflow completed successfully:${NC}"
    cat << 'EOF'
{
  "adoption_id": 102,
  "pet": {
    "id": 1,
    "name": "Buddy",
    "category": "dog"
  },
  "user": {
    "id": 1,
    "name": "John Doe",
    "email": "john@example.com"
  },
  "status": "approved",
  "message": "Adoption approved"
}
EOF
    echo ""
    pause
}

demo_async_operations() {
    print_header "7. ASYNC OPERATIONS & POLLING"

    print_section "Create order with async polling"
    echo -e "${BLUE}# Automatic polling for long-running operations${NC}"
    print_command "./petstore-cli create order --pet-id 2"
    echo ""
    echo -e "${GREEN}✓ Order created (ID: 103)${NC}"
    echo ""
    echo -e "${CYAN}⏳ Waiting for order completion...${NC}"
    echo -e "${YELLOW}  [00:00] Status: placed${NC}"
    echo -e "${YELLOW}  [00:30] Status: placed${NC}"
    echo -e "${YELLOW}  [01:00] Status: approved (changed)${NC}"
    echo -e "${YELLOW}  [01:30] Status: approved${NC}"
    echo -e "${YELLOW}  [02:00] Status: approved${NC}"
    echo -e "${YELLOW}  [02:30] Status: delivered (changed)${NC}"
    echo ""
    echo -e "${GREEN}✓ Order delivered successfully (elapsed: 150s)${NC}"
    echo ""
    pause
}

demo_context_switching() {
    print_header "8. CONTEXT SWITCHING"

    print_section "List available contexts"
    echo -e "${BLUE}# Manage multiple environments${NC}"
    print_command "./petstore-cli list contexts"
    echo ""
    echo -e "  development    http://localhost:8080                         ${GREEN}(active)${NC}"
    echo -e "  staging        https://staging-api.petstore.example.com"
    echo -e "  production     https://api.petstore.example.com"
    echo ""
    pause

    print_section "Switch context"
    run_demo_command \
        "./petstore-cli use context production" \
        "Switch to production environment"
    echo -e "${GREEN}✓ Switched to context: production${NC}"
    echo ""
    pause

    print_section "Show current context"
    run_demo_command \
        "./petstore-cli current context" \
        "Display active context"
    echo -e "Current context: ${GREEN}production${NC}"
    echo -e "Base URL: https://api.petstore.example.com"
    echo ""
    pause
}

demo_plugin_integration() {
    print_header "9. PLUGIN INTEGRATION"

    print_section "Backup pet data to S3 (using AWS CLI plugin)"
    echo -e "${BLUE}# External plugin integration (AWS-like)${NC}"
    print_command "./petstore-cli backup pet --pet-id 1 --bucket my-petstore-backups"
    echo ""
    echo -e "${CYAN}▶ Checking AWS CLI availability...${NC}"
    echo -e "${GREEN}  ✓ AWS CLI v2.13.0 found${NC}"
    echo ""
    echo -e "${CYAN}▶ Fetching pet data...${NC}"
    echo -e "${GREEN}  ✓ Retrieved pet data for ID 1${NC}"
    echo ""
    echo -e "${CYAN}▶ Uploading to S3...${NC}"
    echo -e "${GREEN}  ✓ Uploaded to s3://my-petstore-backups/pets/1.json${NC}"
    echo ""
    echo -e "${CYAN}▶ Verifying backup...${NC}"
    echo -e "${GREEN}  ✓ Backup verified successfully${NC}"
    echo ""
    echo -e "${GREEN}Backup completed!${NC}"
    echo ""
    pause
}

demo_preflight_checks() {
    print_header "10. PRE-FLIGHT VALIDATION"

    print_section "Create pet with pre-flight checks"
    echo -e "${BLUE}# Automatic validation before execution${NC}"
    print_command "./petstore-cli create pet --name \"Charlie\" --category cat"
    echo ""
    echo -e "${CYAN}Running pre-flight checks...${NC}"
    echo ""
    echo -e "${CYAN}▶ Checking store capacity...${NC}"
    echo -e "${GREEN}  ✓ Store capacity available (75/100)${NC}"
    echo ""
    echo -e "${CYAN}▶ Checking for duplicate pet names...${NC}"
    echo -e "${GREEN}  ✓ No duplicate found${NC}"
    echo ""
    echo -e "${GREEN}All pre-flight checks passed!${NC}"
    echo ""
    echo -e "${GREEN}✓ Pet 'Charlie' created successfully with ID 104${NC}"
    echo ""
    pause

    print_section "Skip optional pre-flight checks"
    run_demo_command \
        "./petstore-cli create pet --name \"Charlie\" --category cat --skip-quota-check" \
        "Skip optional validation checks"
    pause
}

demo_deprecation() {
    print_header "11. DEPRECATION WARNINGS"

    print_section "Using deprecated command"
    echo -e "${BLUE}# Automatic deprecation warnings${NC}"
    print_command "./petstore-cli delete pet --pet-id 3"
    echo ""
    echo -e "${YELLOW}╔═══════════════════════════════════════════════════════════╗${NC}"
    echo -e "${YELLOW}║  DEPRECATION WARNING                                      ║${NC}"
    echo -e "${YELLOW}╠═══════════════════════════════════════════════════════════╣${NC}"
    echo -e "${YELLOW}║  Command: delete pet                                      ║${NC}"
    echo -e "${YELLOW}║  Sunset Date: 2025-12-31                                  ║${NC}"
    echo -e "${YELLOW}║                                                           ║${NC}"
    echo -e "${YELLOW}║  Direct deletion is deprecated. Use 'archive pet'        ║${NC}"
    echo -e "${YELLOW}║  instead.                                                 ║${NC}"
    echo -e "${YELLOW}╚═══════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${YELLOW}⚠ Are you sure you want to delete pet '3'? This cannot be undone. [y/N]:${NC} y"
    echo -e "${GREEN}✓ Pet 3 deleted successfully${NC}"
    echo ""
    pause
}

demo_advanced_features() {
    print_header "12. ADVANCED FEATURES"

    print_section "Dry-run mode"
    echo -e "${BLUE}# Preview operations without executing${NC}"
    print_command "./petstore-cli create pet --name \"Test\" --category dog --dry-run"
    echo ""
    echo -e "${YELLOW}[DRY RUN] Would execute:${NC}"
    echo "  POST http://localhost:8080/pets"
    echo "  Body: {\"name\":\"Test\",\"category\":{\"name\":\"dog\"},\"status\":\"available\"}"
    echo ""
    echo -e "${YELLOW}[DRY RUN] Pre-flight checks would run:${NC}"
    echo "  1. validate-capacity"
    echo "  2. validate-duplicate"
    echo ""
    echo -e "${YELLOW}No changes made (dry-run mode)${NC}"
    echo ""
    pause

    print_section "Verbose output"
    echo -e "${BLUE}# Detailed logging and debugging${NC}"
    print_command "./petstore-cli list pets --verbose --limit 2"
    echo ""
    echo -e "${CYAN}[DEBUG] Loading config from ~/.config/petstore/config.yaml${NC}"
    echo -e "${CYAN}[DEBUG] Using context: development${NC}"
    echo -e "${CYAN}[DEBUG] Base URL: http://localhost:8080${NC}"
    echo -e "${CYAN}[DEBUG] GET http://localhost:8080/pets?limit=2${NC}"
    echo -e "${CYAN}[DEBUG] Response: 200 OK (145ms)${NC}"
    echo -e "${CYAN}[DEBUG] Cache hit: false${NC}"
    echo -e "${CYAN}[DEBUG] Caching response (TTL: 60s)${NC}"
    echo ""
    echo "ID    NAME      CATEGORY  STATUS     AGE  PRICE"
    echo "1     Buddy     DOG       AVAILABLE  3    \$250.00"
    echo "2     Whiskers  CAT       AVAILABLE  2    \$150.00"
    echo ""
    pause

    print_section "Shell completion"
    echo -e "${BLUE}# Auto-completion support${NC}"
    print_command "./petstore-cli completion bash > /etc/bash_completion.d/petstore"
    echo -e "${GREEN}✓ Bash completion installed${NC}"
    echo ""
    echo -e "${BLUE}# Now you can use tab completion:${NC}"
    echo -e "  \$ petstore-cli list <TAB><TAB>"
    echo -e "  pets    stores    users    contexts    regions"
    echo ""
    pause
}

demo_summary() {
    print_header "DEMONSTRATION COMPLETE"

    cat << 'EOF'
You've seen all the major features of CliForge v0.9.0:

✓ OpenAPI Extensions:
  • x-cli-command, x-cli-aliases, x-cli-description
  • x-cli-flags with validation and env vars
  • x-cli-interactive with various prompt types
  • x-cli-preflight for validation checks
  • x-cli-output for table/JSON/YAML/CSV formatting
  • x-cli-async for polling long-running operations
  • x-cli-workflow for multi-step orchestration
  • x-cli-plugin for external tool integration
  • x-cli-streaming for SSE support
  • x-cli-watch for real-time monitoring
  • x-cli-confirmation for destructive operations
  • x-cli-deprecation for sunset warnings
  • x-cli-context for environment switching
  • x-cli-cache for response caching

✓ Configuration Features:
  • Branding with ASCII art and colors
  • Multiple output formats
  • OAuth2 authentication with auto-refresh
  • Keyring and file-based credential storage
  • XDG-compliant directory structure
  • Self-update mechanism
  • Telemetry and logging
  • Secret detection and masking

✓ Developer Experience:
  • Interactive mode with smart prompts
  • Auto-completion for all commands
  • Watch mode for monitoring resources
  • Dry-run mode for safe testing
  • Verbose debugging output
  • Context switching for environments
  • Plugin system for extensibility

Next Steps:
  1. Review the generated OpenAPI spec: petstore-api.yaml
  2. Review the CLI configuration: cli-config.yaml
  3. Explore the mock API server: mock-server.go
  4. Try building the example: ./build.sh
  5. Read the documentation: README.md

For more information, visit: https://github.com/cliforge/cliforge

EOF

    echo -e "${GREEN}${BOLD}Thank you for exploring CliForge!${NC}"
    echo ""
}

# Main demo flow
main() {
    show_banner
    demo_basic_operations
    demo_output_formats
    demo_create_operations
    demo_update_delete
    demo_watch_mode
    demo_workflow
    demo_async_operations
    demo_context_switching
    demo_plugin_integration
    demo_preflight_checks
    demo_deprecation
    demo_advanced_features
    demo_summary
}

# Parse arguments
case "${1:-}" in
    --non-interactive|-n)
        INTERACTIVE=false
        PAUSE_BETWEEN_COMMANDS=1
        ;;
    --fast|-f)
        INTERACTIVE=false
        PAUSE_BETWEEN_COMMANDS=0.5
        ;;
    --help|-h)
        cat << EOF
Usage: $0 [options]

Options:
  --non-interactive, -n   Run without pausing
  --fast, -f              Run quickly without pausing
  --help, -h              Show this help message

EOF
        exit 0
        ;;
esac

main
