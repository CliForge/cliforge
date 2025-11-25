# Tutorial: Integrating CliForge CLI in CI/CD Pipelines

**Difficulty**: Intermediate
**Time**: 60-90 minutes
**Version**: 1.0.0

---

## Overview

In this tutorial, you'll learn how to integrate CliForge-generated CLIs into automated CI/CD pipelines. You'll discover best practices for:

- Running CLIs in non-interactive/headless mode
- Managing secrets and credentials securely
- Implementing robust error handling and retries
- Parsing and validating CLI output
- Integrating with popular CI/CD platforms
- Creating reusable workflow templates

By the end, you'll have working examples for GitHub Actions, GitLab CI, and Jenkins.

### What You'll Build

CI/CD pipelines that use your CLI for:

```yaml
# GitHub Actions workflow
- Deploy infrastructure on pull request
- Run integration tests using the CLI
- Clean up resources after testing
- Deploy to production on merge

# GitLab CI pipeline
- Multi-stage deployment with approval gates
- Parallel test execution
- Automatic rollback on failure

# Jenkins pipeline
- Parameterized builds
- Post-deployment validation
- Slack notifications
```

---

## Prerequisites

### Required Knowledge

- Understanding of CI/CD concepts
- Familiarity with YAML syntax
- Basic shell scripting
- Git and version control
- Experience with at least one CI/CD platform

### Required Software

- **Git**: Version control
- **Docker**: For local testing (optional)
- **GitHub/GitLab/Jenkins**: Access to at least one platform
- **CliForge CLI**: From previous tutorials

### Recommended Setup

Complete these tutorials first:
- [Building a REST API CLI](tutorial-rest-api.md)
- [Cloud Management CLI](tutorial-cloud-management.md)

---

## Architecture Overview

### CI/CD Integration Pattern

```
┌─────────────────────────────────────────────────────────┐
│                   Git Repository                        │
│  ├── .github/workflows/  (GitHub Actions)              │
│  ├── .gitlab-ci.yml      (GitLab CI)                   │
│  ├── Jenkinsfile         (Jenkins)                     │
│  └── scripts/            (Reusable scripts)            │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│                  CI/CD Platform                         │
│                                                         │
│  1. Install CLI                                         │
│  2. Configure authentication                            │
│  3. Run operations (non-interactive)                    │
│  4. Validate output                                     │
│  5. Handle errors                                       │
│  6. Cleanup resources                                   │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│                  Target API/Service                     │
└─────────────────────────────────────────────────────────┘
```

### Key Principles

1. **Non-Interactive Mode**: No prompts or user input required
2. **Idempotency**: Safe to retry without side effects
3. **Fail Fast**: Detect and report errors immediately
4. **Structured Output**: JSON/YAML for parsing
5. **Comprehensive Logging**: Debug information for troubleshooting
6. **Secret Management**: Never expose credentials in logs

---

## Step 1: Preparing Your CLI for CI/CD

### Enable Non-Interactive Mode

Update your CLI configuration to support headless operation.

Create `config/ci-config.yaml`:

```yaml
metadata:
  name: cloud-cli
  version: 2.0.0

# CI/CD specific settings
ci_mode:
  enabled: true

  # Disable interactive features
  interactive:
    confirmations: false
    prompts: false
    progress_bars: false
    colors: false

  # Use JSON for machine-readable output
  output:
    default_format: json
    pretty_print: false
    timestamps: true

  # Fail fast on errors
  error_handling:
    exit_on_error: true
    show_stack_trace: false
    structured_errors: true

  # Verbose logging for debugging
  logging:
    level: info
    format: json
    output: stderr

# Authentication
authentication:
  methods:
    api_key:
      # Read from environment
      env_var: "CLOUD_API_KEY"
      # Fail if not set
      required: true

# Retries for flaky operations
errors:
  retry:
    enabled: true
    max_attempts: 3
    backoff: exponential
    initial_delay: 2s

# Timeouts
timeouts:
  default: 300s
  long_running: 1800s
```

### Create CI-Specific Wrapper Script

Create `scripts/ci-wrapper.sh`:

```bash
#!/bin/bash
set -euo pipefail

# CI/CD wrapper for cloud-cli
# Provides consistent error handling and logging

# Configuration
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
CLI_BINARY="${CLI_BINARY:-cloud-cli}"
LOG_FILE="${LOG_FILE:-/tmp/cloud-cli.log}"
EXIT_CODE=0

# Colors (only if terminal)
if [ -t 1 ]; then
  RED='\033[0;31m'
  GREEN='\033[0;32m'
  YELLOW='\033[1;33m'
  NC='\033[0m'
else
  RED=''
  GREEN=''
  YELLOW=''
  NC=''
fi

# Logging functions
log_info() {
  echo -e "${GREEN}[INFO]${NC} $*" | tee -a "$LOG_FILE"
}

log_warn() {
  echo -e "${YELLOW}[WARN]${NC} $*" | tee -a "$LOG_FILE"
}

log_error() {
  echo -e "${RED}[ERROR]${NC} $*" | tee -a "$LOG_FILE" >&2
}

# Validate prerequisites
check_prerequisites() {
  log_info "Checking prerequisites..."

  # Check CLI is installed
  if ! command -v "$CLI_BINARY" &> /dev/null; then
    log_error "CLI binary not found: $CLI_BINARY"
    return 1
  fi

  # Check version
  local version
  version=$("$CLI_BINARY" --version 2>/dev/null | grep -oE '[0-9]+\.[0-9]+\.[0-9]+' || echo "unknown")
  log_info "CLI version: $version"

  # Check authentication
  if [ -z "${CLOUD_API_KEY:-}" ]; then
    log_error "CLOUD_API_KEY environment variable not set"
    return 1
  fi

  log_info "Prerequisites check passed"
  return 0
}

# Execute CLI command with error handling
run_cli() {
  local cmd="$*"
  log_info "Executing: $CLI_BINARY $cmd"

  local output
  local exit_code

  # Run command and capture output
  set +e
  output=$("$CLI_BINARY" $cmd 2>&1)
  exit_code=$?
  set -e

  # Log output
  echo "$output" >> "$LOG_FILE"

  # Handle errors
  if [ $exit_code -ne 0 ]; then
    log_error "Command failed with exit code $exit_code"
    log_error "Output: $output"
    return $exit_code
  fi

  # Return output
  echo "$output"
  return 0
}

# Parse JSON output
parse_json() {
  local json="$1"
  local field="$2"
  echo "$json" | jq -r "$field"
}

# Retry with exponential backoff
retry_with_backoff() {
  local max_attempts="${1:-3}"
  local delay="${2:-2}"
  local cmd="${@:3}"

  local attempt=1

  while [ $attempt -le $max_attempts ]; do
    log_info "Attempt $attempt/$max_attempts: $cmd"

    if eval "$cmd"; then
      return 0
    fi

    if [ $attempt -lt $max_attempts ]; then
      log_warn "Command failed, retrying in ${delay}s..."
      sleep "$delay"
      delay=$((delay * 2))  # Exponential backoff
    fi

    attempt=$((attempt + 1))
  done

  log_error "Command failed after $max_attempts attempts"
  return 1
}

# Cleanup function
cleanup() {
  local exit_code=$?

  if [ $exit_code -eq 0 ]; then
    log_info "Execution completed successfully"
  else
    log_error "Execution failed with exit code $exit_code"
  fi

  # Archive logs if in CI
  if [ -n "${CI:-}" ]; then
    log_info "Archiving logs to artifacts..."
    cp "$LOG_FILE" "${CI_ARTIFACTS_DIR:-./artifacts}/cloud-cli.log" 2>/dev/null || true
  fi

  exit $exit_code
}

trap cleanup EXIT

# Main execution
main() {
  # Check prerequisites
  if ! check_prerequisites; then
    return 1
  fi

  # Execute command passed as arguments
  if [ $# -eq 0 ]; then
    log_error "No command specified"
    echo "Usage: $0 <cli-command> [args...]"
    return 1
  fi

  run_cli "$@"
}

# Run main function with all arguments
main "$@"
```

Make it executable:

```bash
chmod +x scripts/ci-wrapper.sh
```

---

## Step 2: GitHub Actions Integration

### Basic Workflow

Create `.github/workflows/deploy.yml`:

```yaml
name: Deploy Infrastructure

on:
  push:
    branches:
      - main
  pull_request:
    branches:
      - main
  workflow_dispatch:
    inputs:
      environment:
        description: 'Deployment environment'
        required: true
        default: 'staging'
        type: choice
        options:
          - staging
          - production

env:
  CLI_VERSION: '2.0.0'

jobs:
  install-cli:
    name: Install CLI
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Setup Go
        uses: actions/setup-go@v4
        with:
          go-version: '1.21'

      - name: Cache CLI binary
        uses: actions/cache@v3
        id: cache-cli
        with:
          path: ~/bin/cloud-cli
          key: cloud-cli-${{ env.CLI_VERSION }}-${{ runner.os }}

      - name: Install CLI
        if: steps.cache-cli.outputs.cache-hit != 'true'
        run: |
          mkdir -p ~/bin
          # Download or build CLI
          curl -L "https://github.com/yourorg/cloud-cli/releases/download/v${CLI_VERSION}/cloud-cli-linux-amd64" \
            -o ~/bin/cloud-cli
          chmod +x ~/bin/cloud-cli

      - name: Verify installation
        run: |
          export PATH=$HOME/bin:$PATH
          cloud-cli --version

  validate:
    name: Validate Configuration
    needs: install-cli
    runs-on: ubuntu-latest

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Restore CLI cache
        uses: actions/cache@v3
        with:
          path: ~/bin/cloud-cli
          key: cloud-cli-${{ env.CLI_VERSION }}-${{ runner.os }}

      - name: Setup PATH
        run: echo "$HOME/bin" >> $GITHUB_PATH

      - name: Validate API connectivity
        env:
          CLOUD_API_KEY: ${{ secrets.CLOUD_API_KEY }}
        run: |
          cloud-cli --version
          cloud-cli compute instances list --limit 1 --format json

  deploy-staging:
    name: Deploy to Staging
    needs: validate
    runs-on: ubuntu-latest
    if: github.event_name == 'pull_request'

    environment:
      name: staging
      url: https://staging.example.com

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Restore CLI cache
        uses: actions/cache@v3
        with:
          path: ~/bin/cloud-cli
          key: cloud-cli-${{ env.CLI_VERSION }}-${{ runner.os }}

      - name: Setup PATH
        run: echo "$HOME/bin" >> $GITHUB_PATH

      - name: Deploy infrastructure
        env:
          CLOUD_API_KEY: ${{ secrets.CLOUD_API_KEY }}
        run: |
          ./scripts/ci-wrapper.sh deploy full-stack \
            --template workflows/full-stack-deploy.yaml \
            --env staging \
            --instance-count 2 \
            --region us-west \
            --format json > deploy-output.json

      - name: Parse deployment output
        id: deployment
        run: |
          DEPLOYMENT_ID=$(jq -r '.deployment.id' deploy-output.json)
          echo "deployment_id=$DEPLOYMENT_ID" >> $GITHUB_OUTPUT

          APP_URL=$(jq -r '.deployment.load_balancer.public_ip' deploy-output.json)
          echo "app_url=http://$APP_URL" >> $GITHUB_OUTPUT

      - name: Run smoke tests
        run: |
          # Wait for deployment to be ready
          sleep 30

          # Test application endpoint
          curl -f "${{ steps.deployment.outputs.app_url }}/health" || exit 1

      - name: Comment on PR
        uses: actions/github-script@v7
        with:
          script: |
            const output = require('./deploy-output.json');

            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: `## 🚀 Staging Deployment

              **Status**: ✅ Successful
              **Deployment ID**: \`${output.deployment.id}\`
              **URL**: ${output.deployment.load_balancer.public_ip}
              **Instances**: ${output.deployment.instance_count}

              <details>
              <summary>Deployment Details</summary>

              \`\`\`json
              ${JSON.stringify(output, null, 2)}
              \`\`\`

              </details>
              `
            });

      - name: Upload deployment artifacts
        uses: actions/upload-artifact@v3
        if: always()
        with:
          name: deployment-artifacts
          path: |
            deploy-output.json
            /tmp/cloud-cli.log

  deploy-production:
    name: Deploy to Production
    needs: validate
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'

    environment:
      name: production
      url: https://production.example.com

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Restore CLI cache
        uses: actions/cache@v3
        with:
          path: ~/bin/cloud-cli
          key: cloud-cli-${{ env.CLI_VERSION }}-${{ runner.os }}

      - name: Setup PATH
        run: echo "$HOME/bin" >> $GITHUB_PATH

      - name: Deploy infrastructure
        env:
          CLOUD_API_KEY: ${{ secrets.CLOUD_API_KEY_PROD }}
        run: |
          ./scripts/ci-wrapper.sh deploy full-stack \
            --template workflows/full-stack-deploy.yaml \
            --env production \
            --instance-count 5 \
            --region us-west \
            --enable-monitoring \
            --wait \
            --format json > deploy-output.json

      - name: Verify deployment
        env:
          CLOUD_API_KEY: ${{ secrets.CLOUD_API_KEY_PROD }}
        run: |
          DEPLOYMENT_ID=$(jq -r '.deployment.id' deploy-output.json)

          # Check all instances are running
          INSTANCES=$(jq -r '.deployment.instances[].id' deploy-output.json)

          for instance_id in $INSTANCES; do
            echo "Checking instance $instance_id..."
            STATE=$(cloud-cli compute instances get --id "$instance_id" --format json | jq -r '.state')

            if [ "$STATE" != "running" ]; then
              echo "Instance $instance_id is not running: $STATE"
              exit 1
            fi
          done

          echo "All instances are running"

      - name: Create GitHub release
        uses: actions/create-release@v1
        env:
          GITHUB_TOKEN: ${{ secrets.GITHUB_TOKEN }}
        with:
          tag_name: deploy-${{ github.run_number }}
          release_name: Production Deployment ${{ github.run_number }}
          body: |
            Automated production deployment

            Deployment ID: ${{ steps.deployment.outputs.deployment_id }}
            Commit: ${{ github.sha }}

  cleanup:
    name: Cleanup Staging
    needs: [deploy-staging]
    runs-on: ubuntu-latest
    if: always() && github.event_name == 'pull_request'

    steps:
      - name: Checkout code
        uses: actions/checkout@v4

      - name: Restore CLI cache
        uses: actions/cache@v3
        with:
          path: ~/bin/cloud-cli
          key: cloud-cli-${{ env.CLI_VERSION }}-${{ runner.os }}

      - name: Setup PATH
        run: echo "$HOME/bin" >> $GITHUB_PATH

      - name: Cleanup resources
        env:
          CLOUD_API_KEY: ${{ secrets.CLOUD_API_KEY }}
        continue-on-error: true
        run: |
          # List all resources with PR tag
          INSTANCES=$(cloud-cli compute instances list \
            --tag "PR=${{ github.event.pull_request.number }}" \
            --format json | jq -r '.[].id')

          # Delete instances
          for instance_id in $INSTANCES; do
            echo "Deleting instance $instance_id..."
            cloud-cli compute instances delete \
              --id "$instance_id" \
              --force \
              --wait || true
          done
```

### Matrix Strategy for Multi-Environment Testing

Create `.github/workflows/test-matrix.yml`:

```yaml
name: Multi-Environment Testing

on:
  pull_request:
    branches: [main]

jobs:
  test:
    name: Test on ${{ matrix.environment }}
    runs-on: ubuntu-latest

    strategy:
      matrix:
        environment:
          - dev
          - staging
        region:
          - us-west
          - eu-central
        include:
          - environment: dev
            instance_count: 1
          - environment: staging
            instance_count: 2
      fail-fast: false

    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Install CLI
        run: |
          curl -L https://github.com/yourorg/cloud-cli/releases/latest/download/cloud-cli-linux-amd64 \
            -o cloud-cli
          chmod +x cloud-cli
          sudo mv cloud-cli /usr/local/bin/

      - name: Deploy to ${{ matrix.environment }}
        env:
          CLOUD_API_KEY: ${{ secrets.CLOUD_API_KEY }}
        run: |
          cloud-cli deploy full-stack \
            --env ${{ matrix.environment }} \
            --region ${{ matrix.region }} \
            --instance-count ${{ matrix.instance_count }} \
            --tag "CI=true" \
            --tag "PR=${{ github.event.pull_request.number }}" \
            --tag "Environment=${{ matrix.environment }}" \
            --tag "Region=${{ matrix.region }}" \
            --format json > deployment-${{ matrix.environment }}-${{ matrix.region }}.json

      - name: Test deployment
        run: |
          DEPLOYMENT_URL=$(jq -r '.deployment.url' deployment-${{ matrix.environment }}-${{ matrix.region }}.json)

          # Run tests
          curl -f "$DEPLOYMENT_URL/health"
          curl -f "$DEPLOYMENT_URL/api/status"

      - name: Cleanup
        if: always()
        env:
          CLOUD_API_KEY: ${{ secrets.CLOUD_API_KEY }}
        run: |
          DEPLOYMENT_ID=$(jq -r '.deployment.id' deployment-${{ matrix.environment }}-${{ matrix.region }}.json)
          cloud-cli rollback deployment --id "$DEPLOYMENT_ID" --force || true
```

---

## Step 3: GitLab CI Integration

### Complete Pipeline

Create `.gitlab-ci.yml`:

```yaml
variables:
  CLI_VERSION: "2.0.0"
  DOCKER_IMAGE: "golang:1.21"

stages:
  - prepare
  - validate
  - deploy
  - test
  - cleanup

# Reusable template
.cli_template:
  image: $DOCKER_IMAGE
  before_script:
    - mkdir -p $CI_PROJECT_DIR/bin
    - export PATH=$CI_PROJECT_DIR/bin:$PATH
    - |
      if [ ! -f "$CI_PROJECT_DIR/bin/cloud-cli" ]; then
        curl -L "https://github.com/yourorg/cloud-cli/releases/download/v${CLI_VERSION}/cloud-cli-linux-amd64" \
          -o $CI_PROJECT_DIR/bin/cloud-cli
        chmod +x $CI_PROJECT_DIR/bin/cloud-cli
      fi
    - cloud-cli --version
  cache:
    key: ${CI_COMMIT_REF_SLUG}
    paths:
      - bin/

install:
  extends: .cli_template
  stage: prepare
  script:
    - cloud-cli --version
  artifacts:
    paths:
      - bin/
    expire_in: 1 hour

validate-api:
  extends: .cli_template
  stage: validate
  dependencies:
    - install
  script:
    - cloud-cli compute instances list --limit 1 --format json
  only:
    - merge_requests
    - main

validate-config:
  extends: .cli_template
  stage: validate
  dependencies:
    - install
  script:
    # Validate workflow files
    - |
      for workflow in workflows/*.yaml; do
        echo "Validating $workflow..."
        cloud-cli workflow validate --file "$workflow"
      done
  only:
    - merge_requests
    - main

# Deploy to development (automatic)
deploy:dev:
  extends: .cli_template
  stage: deploy
  dependencies:
    - install
  environment:
    name: development
    on_stop: cleanup:dev
  script:
    - |
      cloud-cli deploy full-stack \
        --template workflows/full-stack-deploy.yaml \
        --env development \
        --region ${AWS_REGION:-us-west} \
        --instance-count 1 \
        --tag "GitLab-CI=true" \
        --tag "Pipeline-ID=${CI_PIPELINE_ID}" \
        --tag "Commit=${CI_COMMIT_SHORT_SHA}" \
        --format json | tee deployment-dev.json

    # Save deployment info
    - |
      DEPLOYMENT_ID=$(jq -r '.deployment.id' deployment-dev.json)
      echo "DEPLOYMENT_ID=$DEPLOYMENT_ID" >> deploy-dev.env

      APP_URL=$(jq -r '.deployment.url' deployment-dev.json)
      echo "APP_URL=$APP_URL" >> deploy-dev.env
  artifacts:
    reports:
      dotenv: deploy-dev.env
    paths:
      - deployment-dev.json
    expire_in: 1 day
  only:
    - merge_requests

# Deploy to staging (manual)
deploy:staging:
  extends: .cli_template
  stage: deploy
  dependencies:
    - install
  environment:
    name: staging
    url: https://staging.example.com
    on_stop: cleanup:staging
  script:
    - |
      cloud-cli deploy full-stack \
        --template workflows/full-stack-deploy.yaml \
        --env staging \
        --region ${AWS_REGION:-us-west} \
        --instance-count 2 \
        --enable-monitoring \
        --wait \
        --format json | tee deployment-staging.json
  artifacts:
    paths:
      - deployment-staging.json
    expire_in: 7 days
  when: manual
  only:
    - main

# Deploy to production (manual with protection)
deploy:production:
  extends: .cli_template
  stage: deploy
  dependencies:
    - install
  environment:
    name: production
    url: https://production.example.com
  script:
    # Pre-deployment checks
    - |
      echo "Running pre-deployment checks..."
      cloud-cli preflight-check \
        --env production \
        --region us-west \
        --format json

    # Deploy with blue-green strategy
    - |
      cloud-cli deploy full-stack \
        --template workflows/full-stack-deploy.yaml \
        --env production \
        --region us-west \
        --instance-count 5 \
        --enable-monitoring \
        --enable-backup \
        --deployment-strategy blue-green \
        --wait \
        --format json | tee deployment-prod.json

    # Verify deployment
    - |
      DEPLOYMENT_ID=$(jq -r '.deployment.id' deployment-prod.json)
      echo "Verifying deployment $DEPLOYMENT_ID..."

      cloud-cli deployment verify \
        --id "$DEPLOYMENT_ID" \
        --health-check-url "/health" \
        --timeout 300s
  artifacts:
    paths:
      - deployment-prod.json
    expire_in: 30 days
  when: manual
  only:
    - main

# Integration tests
test:integration:
  extends: .cli_template
  stage: test
  dependencies:
    - install
    - deploy:dev
  script:
    - |
      echo "Running integration tests against $APP_URL..."

      # Test endpoints
      curl -f "$APP_URL/health" || exit 1
      curl -f "$APP_URL/api/status" || exit 1

      # Run API tests
      ./scripts/run-api-tests.sh "$APP_URL"
  only:
    - merge_requests

# Performance tests
test:performance:
  extends: .cli_template
  stage: test
  dependencies:
    - install
    - deploy:dev
  script:
    - |
      # Install performance testing tools
      apt-get update && apt-get install -y apache2-utils

      # Run load test
      ab -n 1000 -c 10 "$APP_URL/" > performance.txt

      # Parse results
      cat performance.txt
  artifacts:
    paths:
      - performance.txt
    expire_in: 7 days
  only:
    - merge_requests

# Cleanup development
cleanup:dev:
  extends: .cli_template
  stage: cleanup
  dependencies:
    - install
  environment:
    name: development
    action: stop
  script:
    - |
      # Find all resources from this pipeline
      echo "Cleaning up development resources..."

      cloud-cli compute instances list \
        --tag "Pipeline-ID=${CI_PIPELINE_ID}" \
        --format json > instances.json

      # Delete instances
      jq -r '.[].id' instances.json | while read instance_id; do
        echo "Deleting instance $instance_id..."
        cloud-cli compute instances delete \
          --id "$instance_id" \
          --force \
          --wait || true
      done
  when: manual
  only:
    - merge_requests

# Cleanup staging
cleanup:staging:
  extends: .cli_template
  stage: cleanup
  dependencies:
    - install
  environment:
    name: staging
    action: stop
  script:
    - |
      cloud-cli rollback deployment \
        --env staging \
        --force \
        --delete-resources
  when: manual
  only:
    - main
```

### Multi-Project Pipeline

For complex scenarios with multiple services:

```yaml
# .gitlab-ci.yml for orchestrator
stages:
  - trigger

trigger:microservices:
  stage: trigger
  trigger:
    project: team/microservices
    strategy: depend
  variables:
    DEPLOYMENT_ENV: staging

trigger:infrastructure:
  stage: trigger
  trigger:
    project: team/infrastructure
    strategy: depend
  variables:
    DEPLOYMENT_ENV: staging
```

---

## Step 4: Jenkins Integration

### Declarative Pipeline

Create `Jenkinsfile`:

```groovy
pipeline {
    agent any

    parameters {
        choice(
            name: 'ENVIRONMENT',
            choices: ['dev', 'staging', 'production'],
            description: 'Deployment environment'
        )
        choice(
            name: 'REGION',
            choices: ['us-west', 'us-east', 'eu-central'],
            description: 'Deployment region'
        )
        string(
            name: 'INSTANCE_COUNT',
            defaultValue: '3',
            description: 'Number of instances to deploy'
        )
        booleanParam(
            name: 'ENABLE_MONITORING',
            defaultValue: true,
            description: 'Enable monitoring and alerting'
        )
        booleanParam(
            name: 'DRY_RUN',
            defaultValue: false,
            description: 'Perform dry run without actual deployment'
        )
    }

    environment {
        CLI_VERSION = '2.0.0'
        PATH = "$HOME/bin:$PATH"
        CLOUD_API_KEY = credentials('cloud-api-key')
    }

    stages {
        stage('Setup') {
            steps {
                script {
                    echo "Setting up environment..."

                    // Create bin directory
                    sh 'mkdir -p $HOME/bin'

                    // Download CLI if not cached
                    def cliExists = fileExists("$HOME/bin/cloud-cli")
                    if (!cliExists) {
                        echo "Downloading cloud-cli ${CLI_VERSION}..."
                        sh """
                            curl -L "https://github.com/yourorg/cloud-cli/releases/download/v${CLI_VERSION}/cloud-cli-linux-amd64" \
                                -o \$HOME/bin/cloud-cli
                            chmod +x \$HOME/bin/cloud-cli
                        """
                    }

                    // Verify installation
                    sh 'cloud-cli --version'
                }
            }
        }

        stage('Validate') {
            parallel {
                stage('API Connectivity') {
                    steps {
                        echo "Validating API connectivity..."
                        sh 'cloud-cli compute instances list --limit 1 --format json'
                    }
                }

                stage('Configuration') {
                    steps {
                        echo "Validating workflow configuration..."
                        sh '''
                            for workflow in workflows/*.yaml; do
                                echo "Validating $workflow..."
                                cloud-cli workflow validate --file "$workflow"
                            done
                        '''
                    }
                }

                stage('Quota Check') {
                    steps {
                        script {
                            echo "Checking resource quotas..."
                            def quotaJson = sh(
                                script: 'cloud-cli quota check --region ${REGION} --format json',
                                returnStdout: true
                            ).trim()

                            def quota = readJSON text: quotaJson

                            if (quota.instances.available < params.INSTANCE_COUNT.toInteger()) {
                                error("Insufficient instance quota. Available: ${quota.instances.available}, Required: ${params.INSTANCE_COUNT}")
                            }
                        }
                    }
                }
            }
        }

        stage('Deploy') {
            when {
                expression { params.DRY_RUN == false }
            }
            steps {
                script {
                    echo "Deploying to ${params.ENVIRONMENT}..."

                    // Build deployment command
                    def deployCmd = """
                        cloud-cli deploy full-stack \
                            --template workflows/full-stack-deploy.yaml \
                            --env ${params.ENVIRONMENT} \
                            --region ${params.REGION} \
                            --instance-count ${params.INSTANCE_COUNT} \
                            --tag "Jenkins=true" \
                            --tag "Build=${BUILD_NUMBER}" \
                            --tag "Environment=${params.ENVIRONMENT}" \
                            --format json
                    """

                    if (params.ENABLE_MONITORING) {
                        deployCmd += " --enable-monitoring"
                    }

                    // Execute deployment
                    def deploymentJson = sh(
                        script: "${deployCmd} | tee deployment.json",
                        returnStdout: true
                    ).trim()

                    // Parse deployment output
                    writeFile file: 'deployment.json', text: deploymentJson
                    def deployment = readJSON file: 'deployment.json'

                    // Store deployment info
                    env.DEPLOYMENT_ID = deployment.deployment.id
                    env.DEPLOYMENT_URL = deployment.deployment.url

                    echo "Deployment ID: ${env.DEPLOYMENT_ID}"
                    echo "Deployment URL: ${env.DEPLOYMENT_URL}"
                }
            }
        }

        stage('Dry Run') {
            when {
                expression { params.DRY_RUN == true }
            }
            steps {
                echo "Performing dry run..."
                sh """
                    cloud-cli deploy full-stack \
                        --template workflows/full-stack-deploy.yaml \
                        --env ${params.ENVIRONMENT} \
                        --region ${params.REGION} \
                        --instance-count ${params.INSTANCE_COUNT} \
                        --dry-run \
                        --format yaml | tee dry-run-output.yaml
                """
                archiveArtifacts artifacts: 'dry-run-output.yaml'
            }
        }

        stage('Test') {
            when {
                expression { params.DRY_RUN == false }
            }
            parallel {
                stage('Smoke Tests') {
                    steps {
                        echo "Running smoke tests..."
                        sh '''
                            # Wait for deployment to stabilize
                            sleep 30

                            # Test health endpoint
                            curl -f "${DEPLOYMENT_URL}/health" || exit 1

                            # Test API endpoint
                            curl -f "${DEPLOYMENT_URL}/api/status" || exit 1
                        '''
                    }
                }

                stage('Security Scan') {
                    steps {
                        echo "Running security scan..."
                        sh '''
                            # Run security checks on deployed instances
                            INSTANCES=$(cloud-cli compute instances list \
                                --tag "Build=${BUILD_NUMBER}" \
                                --format json | jq -r '.[].id')

                            for instance in $INSTANCES; do
                                echo "Scanning instance $instance..."
                                cloud-cli security scan \
                                    --instance-id "$instance" \
                                    --format json > "security-scan-$instance.json"
                            done
                        '''
                        archiveArtifacts artifacts: 'security-scan-*.json'
                    }
                }
            }
        }

        stage('Approval') {
            when {
                expression { params.ENVIRONMENT == 'production' }
            }
            steps {
                script {
                    def userInput = input(
                        id: 'deployApproval',
                        message: 'Approve production deployment?',
                        parameters: [
                            booleanParam(
                                defaultValue: false,
                                description: 'Confirm deployment to production',
                                name: 'APPROVE'
                            )
                        ]
                    )

                    if (!userInput) {
                        error('Deployment not approved')
                    }
                }
            }
        }

        stage('Verify') {
            when {
                expression { params.DRY_RUN == false }
            }
            steps {
                echo "Verifying deployment..."
                sh """
                    cloud-cli deployment verify \
                        --id "${DEPLOYMENT_ID}" \
                        --health-check-url "/health" \
                        --timeout 300s \
                        --format json | tee verification.json
                """

                script {
                    def verification = readJSON file: 'verification.json'

                    if (verification.status != 'healthy') {
                        error("Deployment verification failed: ${verification.message}")
                    }

                    echo "Deployment verified successfully"
                }
            }
        }
    }

    post {
        success {
            script {
                if (!params.DRY_RUN) {
                    echo "Deployment successful!"

                    // Send notification
                    slackSend(
                        color: 'good',
                        message: """
                            Deployment Successful :rocket:
                            Environment: ${params.ENVIRONMENT}
                            Region: ${params.REGION}
                            Deployment ID: ${env.DEPLOYMENT_ID}
                            URL: ${env.DEPLOYMENT_URL}
                            Build: ${BUILD_URL}
                        """
                    )
                }
            }

            archiveArtifacts artifacts: 'deployment.json', allowEmptyArchive: true
        }

        failure {
            script {
                echo "Deployment failed!"

                // Rollback if not dry run
                if (!params.DRY_RUN && env.DEPLOYMENT_ID) {
                    echo "Initiating rollback..."
                    sh """
                        cloud-cli rollback deployment \
                            --id "${env.DEPLOYMENT_ID}" \
                            --reason "Build ${BUILD_NUMBER} failed" \
                            --force
                    """
                }

                // Send notification
                slackSend(
                    color: 'danger',
                    message: """
                        Deployment Failed :x:
                        Environment: ${params.ENVIRONMENT}
                        Build: ${BUILD_URL}
                        Console: ${BUILD_URL}console
                    """
                )
            }
        }

        always {
            // Archive logs
            sh 'cp /tmp/cloud-cli.log cloud-cli.log || true'
            archiveArtifacts artifacts: 'cloud-cli.log', allowEmptyArchive: true

            // Cleanup
            cleanWs()
        }
    }
}
```

### Shared Library for Reusability

Create `vars/cloudDeploy.groovy`:

```groovy
def call(Map config) {
    pipeline {
        agent any

        environment {
            CLOUD_API_KEY = credentials('cloud-api-key')
        }

        stages {
            stage('Deploy') {
                steps {
                    script {
                        echo "Deploying ${config.name}..."

                        sh """
                            cloud-cli deploy ${config.workflow} \
                                --env ${config.environment} \
                                --region ${config.region} \
                                ${config.extraArgs ?: ''} \
                                --format json > deployment.json
                        """

                        def deployment = readJSON file: 'deployment.json'
                        return deployment
                    }
                }
            }
        }
    }
}
```

Usage:

```groovy
@Library('cloud-pipeline-library') _

cloudDeploy(
    name: 'my-app',
    workflow: 'full-stack',
    environment: 'production',
    region: 'us-west',
    extraArgs: '--instance-count 5 --enable-monitoring'
)
```

---

## Step 5: Advanced Patterns

### Error Handling and Retries

Create `scripts/deploy-with-retry.sh`:

```bash
#!/bin/bash
set -euo pipefail

MAX_RETRIES=3
RETRY_DELAY=10

deploy() {
  local env="$1"
  local attempt=1

  while [ $attempt -le $MAX_RETRIES ]; do
    echo "Deployment attempt $attempt/$MAX_RETRIES..."

    if cloud-cli deploy full-stack \
        --env "$env" \
        --format json > "deployment-attempt-$attempt.json"; then

      # Verify deployment
      local deployment_id
      deployment_id=$(jq -r '.deployment.id' "deployment-attempt-$attempt.json")

      if cloud-cli deployment verify --id "$deployment_id"; then
        echo "Deployment successful"
        return 0
      else
        echo "Deployment verification failed"
      fi
    fi

    if [ $attempt -lt $MAX_RETRIES ]; then
      echo "Retrying in ${RETRY_DELAY}s..."
      sleep $RETRY_DELAY

      # Exponential backoff
      RETRY_DELAY=$((RETRY_DELAY * 2))
    fi

    attempt=$((attempt + 1))
  done

  echo "Deployment failed after $MAX_RETRIES attempts"
  return 1
}

deploy "$1"
```

### Parallel Deployments

For deploying to multiple regions:

```bash
#!/bin/bash

REGIONS=("us-west" "us-east" "eu-central")
PIDS=()

# Deploy to all regions in parallel
for region in "${REGIONS[@]}"; do
  (
    echo "Deploying to $region..."
    cloud-cli deploy full-stack \
      --env production \
      --region "$region" \
      --format json > "deployment-$region.json"

    echo "Deployment to $region complete"
  ) &

  PIDS+=($!)
done

# Wait for all deployments
FAILED=0
for pid in "${PIDS[@]}"; do
  if ! wait "$pid"; then
    FAILED=1
  fi
done

if [ $FAILED -eq 1 ]; then
  echo "One or more deployments failed"
  exit 1
fi

echo "All deployments successful"
```

### Blue-Green Deployment

```bash
#!/bin/bash
set -euo pipefail

ENVIRONMENT="production"
REGION="us-west"

echo "Starting blue-green deployment..."

# Deploy green environment
echo "Deploying green environment..."
cloud-cli deploy full-stack \
  --env "${ENVIRONMENT}-green" \
  --region "$REGION" \
  --format json > deployment-green.json

GREEN_URL=$(jq -r '.deployment.url' deployment-green.json)

# Test green environment
echo "Testing green environment..."
if ! curl -f "$GREEN_URL/health"; then
  echo "Green environment health check failed"
  cloud-cli rollback deployment --env "${ENVIRONMENT}-green" --force
  exit 1
fi

# Run smoke tests
./scripts/smoke-tests.sh "$GREEN_URL"

# Switch traffic to green
echo "Switching traffic to green..."
cloud-cli lb switch-traffic \
  --from "${ENVIRONMENT}-blue" \
  --to "${ENVIRONMENT}-green"

# Monitor for errors
sleep 60

# Check error rate
ERROR_RATE=$(cloud-cli metrics get \
  --env "${ENVIRONMENT}-green" \
  --metric error_rate \
  --format json | jq -r '.value')

if (( $(echo "$ERROR_RATE > 0.01" | bc -l) )); then
  echo "High error rate detected, rolling back..."
  cloud-cli lb switch-traffic \
    --from "${ENVIRONMENT}-green" \
    --to "${ENVIRONMENT}-blue"
  exit 1
fi

# Deployment successful, cleanup blue
echo "Cleaning up blue environment..."
cloud-cli cleanup deployment --env "${ENVIRONMENT}-blue"

echo "Blue-green deployment complete"
```

---

## Step 6: Monitoring and Notifications

### Slack Integration

Create `scripts/notify-slack.sh`:

```bash
#!/bin/bash

SLACK_WEBHOOK="${SLACK_WEBHOOK_URL}"
DEPLOYMENT_FILE="$1"
STATUS="$2"

if [ ! -f "$DEPLOYMENT_FILE" ]; then
  echo "Deployment file not found"
  exit 1
fi

# Parse deployment info
DEPLOYMENT_ID=$(jq -r '.deployment.id' "$DEPLOYMENT_FILE")
DEPLOYMENT_URL=$(jq -r '.deployment.url' "$DEPLOYMENT_FILE")
ENVIRONMENT=$(jq -r '.deployment.environment' "$DEPLOYMENT_FILE")
INSTANCE_COUNT=$(jq -r '.deployment.instance_count' "$DEPLOYMENT_FILE")

# Build message
if [ "$STATUS" = "success" ]; then
  COLOR="good"
  EMOJI=":rocket:"
else
  COLOR="danger"
  EMOJI=":x:"
fi

MESSAGE=$(cat <<EOF
{
  "attachments": [
    {
      "color": "$COLOR",
      "title": "$EMOJI Deployment $STATUS",
      "fields": [
        {
          "title": "Environment",
          "value": "$ENVIRONMENT",
          "short": true
        },
        {
          "title": "Deployment ID",
          "value": "$DEPLOYMENT_ID",
          "short": true
        },
        {
          "title": "URL",
          "value": "$DEPLOYMENT_URL",
          "short": false
        },
        {
          "title": "Instances",
          "value": "$INSTANCE_COUNT",
          "short": true
        },
        {
          "title": "Build",
          "value": "${CI_PIPELINE_URL:-N/A}",
          "short": true
        }
      ]
    }
  ]
}
EOF
)

curl -X POST -H 'Content-type: application/json' \
  --data "$MESSAGE" \
  "$SLACK_WEBHOOK"
```

### Datadog Metrics

```bash
#!/bin/bash

DATADOG_API_KEY="${DATADOG_API_KEY}"
DEPLOYMENT_FILE="$1"

# Extract metrics
DEPLOYMENT_TIME=$(jq -r '.deployment.duration' "$DEPLOYMENT_FILE")
INSTANCE_COUNT=$(jq -r '.deployment.instance_count' "$DEPLOYMENT_FILE")
ENVIRONMENT=$(jq -r '.deployment.environment' "$DEPLOYMENT_FILE")

# Send metrics to Datadog
curl -X POST "https://api.datadoghq.com/api/v1/series" \
  -H "DD-API-KEY: $DATADOG_API_KEY" \
  -H "Content-Type: application/json" \
  -d @- <<EOF
{
  "series": [
    {
      "metric": "deployment.duration",
      "points": [[$(date +%s), $DEPLOYMENT_TIME]],
      "type": "gauge",
      "tags": ["environment:$ENVIRONMENT"]
    },
    {
      "metric": "deployment.instances",
      "points": [[$(date +%s), $INSTANCE_COUNT]],
      "type": "gauge",
      "tags": ["environment:$ENVIRONMENT"]
    }
  ]
}
EOF
```

---

## Troubleshooting

### Issue: Authentication Failures in CI

**Problem**: CLI cannot authenticate in CI environment

**Solution**:
```yaml
# Verify secret is set
- name: Check authentication
  run: |
    if [ -z "$CLOUD_API_KEY" ]; then
      echo "CLOUD_API_KEY not set"
      exit 1
    fi

    # Test authentication
    cloud-cli --version
    cloud-cli compute instances list --limit 1
```

### Issue: Timeouts in CI

**Problem**: Operations timeout in CI but work locally

**Solution**:
```bash
# Increase timeouts in CI
cloud-cli deploy full-stack \
  --timeout 1800s \
  --async-poll-interval 10s

# Or set globally
export CLI_DEFAULT_TIMEOUT=1800s
```

### Issue: Rate Limiting

**Problem**: CI pipeline hits API rate limits

**Solution**:
```bash
# Implement backoff
cloud-cli --rate-limit 5/s deploy full-stack

# Or use exponential backoff
cloud-cli --retry-backoff exponential deploy full-stack
```

---

## Best Practices Summary

### 1. Always Use JSON Output in CI

```bash
# Good
cloud-cli deploy --format json > deployment.json

# Bad (not parseable)
cloud-cli deploy
```

### 2. Check Exit Codes

```bash
if cloud-cli deploy full-stack; then
  echo "Success"
else
  echo "Failed with code $?"
  exit 1
fi
```

### 3. Tag Resources

```bash
cloud-cli deploy \
  --tag "CI=true" \
  --tag "Build=$BUILD_NUMBER" \
  --tag "Environment=$ENV"
```

### 4. Archive Artifacts

```yaml
# GitHub Actions
- uses: actions/upload-artifact@v3
  with:
    path: |
      deployment.json
      logs/*.log
```

### 5. Use Secrets Management

```bash
# Never hardcode
API_KEY="secret123"  # Bad

# Use environment
CLOUD_API_KEY="${{ secrets.CLOUD_API_KEY }}"  # Good
```

---

## Summary

In this tutorial, you learned:

- ✓ Preparing CLIs for non-interactive CI/CD use
- ✓ GitHub Actions integration with matrix strategies
- ✓ GitLab CI with multi-stage pipelines
- ✓ Jenkins declarative and scripted pipelines
- ✓ Error handling and retry logic
- ✓ Parallel and blue-green deployments
- ✓ Monitoring and notifications
- ✓ Best practices for production CI/CD

---

**Tutorial Version**: 1.0.0
**Last Updated**: 2025-11-25
**CliForge Version**: 0.9.0+
