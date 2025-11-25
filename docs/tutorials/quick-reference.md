# Tutorial Quick Reference

Quick reference guide for commands and patterns from the CliForge tutorial series.

---

## REST API CLI Tutorial

### Setup Commands

```bash
# Install CliForge
curl -L https://github.com/cliforge/cliforge/releases/latest/download/cliforge-$(uname -s)-$(uname -m) -o cliforge
chmod +x cliforge
sudo mv cliforge /usr/local/bin/

# Initialize CLI project
cliforge init my-api-cli

# Build CLI
cliforge build --config config/cli-config.yaml --output ./my-cli
```

### Authentication

```bash
# Set token via environment
export GITHUB_TOKEN="ghp_your_token_here"

# Set token via config
./github-cli config set github.token "ghp_your_token_here"

# Verify authentication
./github-cli repos list --limit 1
```

### Basic Operations

```bash
# List resources
./github-cli repos list
./github-cli repos list --type public --sort created

# Create resource
./github-cli repos create \
  --name "my-repo" \
  --description "My project" \
  --private \
  --init

# Get resource
./github-cli repos get --owner octocat --repo hello-world

# Delete resource (with confirmation)
./github-cli repos delete --owner user --repo test-repo

# Force delete (skip confirmation)
./github-cli repos delete --owner user --repo test-repo --force
```

### Output Formats

```bash
# Table format (default)
./github-cli repos list

# JSON format
./github-cli repos list --format json

# YAML format
./github-cli repos get --owner user --repo name --format yaml

# JSON with jq filtering
./github-cli repos list --format json | jq '.[] | select(.private == true)'
```

### Issue Management

```bash
# List issues
./github-cli issues list --owner user --repo repo
./github-cli issues list --owner user --repo repo --state closed

# Create issue
./github-cli issues create \
  --owner user \
  --repo repo \
  --title "Bug report" \
  --body "Description here" \
  --labels "bug,urgent"

# Get issue details
./github-cli issues get --owner user --repo repo --number 123
```

---

## Cloud Management CLI Tutorial

### Instance Management

```bash
# List instances
cloud-cli compute instances list
cloud-cli compute instances list --region us-west --state running

# Create instance (async)
cloud-cli compute instances create \
  --name web-server-1 \
  --size large \
  --region us-west

# Create and wait for completion
cloud-cli compute instances create \
  --name web-server-1 \
  --size large \
  --region us-west \
  --wait

# Start/stop instances
cloud-cli compute instances start --id i-abc123 --wait
cloud-cli compute instances stop --id i-abc123 --wait
cloud-cli compute instances stop --id i-abc123 --force  # Force stop

# Delete instance
cloud-cli compute instances delete --id i-abc123 --wait
```

### Async Operations

```bash
# Check operation status
cloud-cli operations get --id op-abc123

# List running operations
cloud-cli operations list --status running

# Cancel operation
cloud-cli operations cancel --id op-abc123
```

### Storage Operations

```bash
# List buckets
cloud-cli storage buckets list

# Create bucket
cloud-cli storage buckets create \
  --name my-data \
  --region us-west \
  --encryption aes256 \
  --versioning

# List objects in bucket
cloud-cli storage objects list --bucket my-data --prefix logs/

# Sync files
cloud-cli storage sync \
  --source ./data \
  --destination s3://my-bucket/backup

# Sync with watch mode
cloud-cli storage sync \
  --source ./app-data \
  --destination s3://my-bucket \
  --watch \
  --delete
```

### Workflow Deployment

```bash
# Deploy infrastructure
cloud-cli deploy full-stack \
  --template workflows/full-stack-deploy.yaml \
  --env production \
  --region us-west \
  --instance-count 5

# Dry run (test without deploying)
cloud-cli deploy full-stack \
  --template workflows/full-stack-deploy.yaml \
  --env staging \
  --dry-run

# Deploy with monitoring
cloud-cli deploy full-stack \
  --template workflows/full-stack-deploy.yaml \
  --env production \
  --enable-monitoring
```

### State Management

```bash
# List managed resources
cloud-cli state list

# Detect configuration drift
cloud-cli state drift-detect

# Refresh state from actual infrastructure
cloud-cli state refresh

# Export state
cloud-cli state export --format terraform > infrastructure.tf
cloud-cli state export --format json > state.json
```

### Rollback

```bash
# Rollback deployment
cloud-cli rollback deployment --id dep-123 --reason "High error rate"

# Force rollback
cloud-cli rollback deployment --id dep-123 --force
```

---

## CI/CD Integration Tutorial

### GitHub Actions Patterns

```yaml
# Install CLI in workflow
- name: Install CLI
  run: |
    curl -L https://github.com/org/cli/releases/latest/download/cli-linux-amd64 -o cli
    chmod +x cli
    sudo mv cli /usr/local/bin/

# Use CLI with secrets
- name: Deploy
  env:
    CLOUD_API_KEY: ${{ secrets.CLOUD_API_KEY }}
  run: |
    cloud-cli deploy --env production --format json > deployment.json

# Parse JSON output
- name: Get deployment info
  id: deploy
  run: |
    DEPLOYMENT_ID=$(jq -r '.deployment.id' deployment.json)
    echo "deployment_id=$DEPLOYMENT_ID" >> $GITHUB_OUTPUT
```

### GitLab CI Patterns

```yaml
# Reusable template
.cli_template:
  before_script:
    - curl -L https://releases.example.com/cli -o /usr/local/bin/cli
    - chmod +x /usr/local/bin/cli
    - cli --version

# Use template
deploy:
  extends: .cli_template
  script:
    - cli deploy --env $CI_ENVIRONMENT_NAME --format json
```

### Jenkins Patterns

```groovy
// Install and verify CLI
stage('Setup') {
    steps {
        sh '''
            curl -L https://releases.example.com/cli -o $HOME/bin/cli
            chmod +x $HOME/bin/cli
            cli --version
        '''
    }
}

// Run deployment
stage('Deploy') {
    steps {
        script {
            def output = sh(
                script: 'cli deploy --env production --format json',
                returnStdout: true
            ).trim()

            def deployment = readJSON text: output
            env.DEPLOYMENT_ID = deployment.deployment.id
        }
    }
}
```

### CI Helper Scripts

```bash
# CI wrapper script (scripts/ci-wrapper.sh)
#!/bin/bash
set -euo pipefail

# Validate prerequisites
if [ -z "${CLOUD_API_KEY:-}" ]; then
  echo "Error: CLOUD_API_KEY not set"
  exit 1
fi

# Run CLI with error handling
cloud-cli "$@" 2>&1 | tee cli.log

# Check exit code
if [ ${PIPESTATUS[0]} -ne 0 ]; then
  echo "CLI command failed"
  exit 1
fi

# Usage
./scripts/ci-wrapper.sh deploy --env production
```

### Retry Logic

```bash
# Retry with exponential backoff
retry_with_backoff() {
  local max_attempts=3
  local delay=2
  local attempt=1

  while [ $attempt -le $max_attempts ]; do
    if cloud-cli "$@"; then
      return 0
    fi

    if [ $attempt -lt $max_attempts ]; then
      sleep $delay
      delay=$((delay * 2))
    fi
    attempt=$((attempt + 1))
  done

  return 1
}

# Usage
retry_with_backoff deploy --env production
```

### Blue-Green Deployment

```bash
# Deploy green environment
cloud-cli deploy full-stack --env production-green

# Test green environment
./scripts/smoke-tests.sh "$GREEN_URL"

# Switch traffic
cloud-cli lb switch-traffic --from production-blue --to production-green

# Cleanup blue environment
cloud-cli cleanup deployment --env production-blue
```

### Parallel Deployments

```bash
# Deploy to multiple regions in parallel
REGIONS=("us-west" "us-east" "eu-central")

for region in "${REGIONS[@]}"; do
  (
    cloud-cli deploy --region "$region" --format json > "deploy-$region.json"
  ) &
done

wait  # Wait for all deployments
```

---

## Common Patterns

### Error Handling

```bash
# Check exit code
if cloud-cli deploy --env production; then
  echo "Success"
else
  echo "Failed with code $?"
  exit 1
fi

# Capture output and exit code
output=$(cloud-cli deploy --env production 2>&1)
exit_code=$?

if [ $exit_code -ne 0 ]; then
  echo "Error: $output"
  exit $exit_code
fi
```

### JSON Parsing

```bash
# Extract single field
DEPLOYMENT_ID=$(cloud-cli deploy --format json | jq -r '.deployment.id')

# Extract array of values
INSTANCE_IDS=$(cloud-cli instances list --format json | jq -r '.[].id')

# Filter results
RUNNING_INSTANCES=$(cloud-cli instances list --format json | \
  jq '.[] | select(.state == "running")')

# Multiple fields
cloud-cli deploy --format json | jq '{id: .deployment.id, url: .deployment.url}'
```

### Logging

```bash
# Log to file
cloud-cli deploy --env production 2>&1 | tee deployment.log

# Structured logging (JSON)
cloud-cli --log-format json deploy --env production 2>> app.log

# Verbose mode
cloud-cli --verbose deploy --env production
```

### Resource Tagging

```bash
# Tag resources for tracking
cloud-cli instances create \
  --tag "Environment=production" \
  --tag "Team=platform" \
  --tag "CostCenter=engineering" \
  --tag "ManagedBy=cli" \
  --tag "CreatedAt=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Query by tags
cloud-cli instances list \
  --tag "Environment=production" \
  --tag "Team=platform"

# Cleanup resources by tag
cloud-cli instances list --tag "CI=true" --format json | \
  jq -r '.[].id' | \
  xargs -I {} cloud-cli instances delete --id {} --force
```

---

## Configuration

### Profile Management

```bash
# Use specific profile
cloud-cli --profile production deploy

# Set default profile
cloud-cli config set default-profile production

# List profiles
cloud-cli config profiles list

# Create profile
cloud-cli config profile create staging \
  --api-key "staging-key" \
  --region "us-west"
```

### Configuration Values

```bash
# Get configuration value
cloud-cli config get output.format

# Set configuration value
cloud-cli config set output.format json
cloud-cli config set cache.enabled true
cloud-cli config set cache.ttl 600

# List all configuration
cloud-cli config list

# Edit configuration interactively
cloud-cli config edit
```

---

## Debugging

### Debug Mode

```bash
# Enable debug logging
cloud-cli --log-level debug deploy

# Show HTTP requests/responses
cloud-cli --debug-http deploy

# Verbose output
cloud-cli --verbose deploy
```

### Validation

```bash
# Validate OpenAPI spec
cliforge validate --spec specs/api.yaml

# Validate workflow
cloud-cli workflow validate --file workflows/deploy.yaml

# Validate configuration
cloud-cli config validate
```

### Dry Run

```bash
# Test without execution
cloud-cli deploy --dry-run

# Show what would be created
cloud-cli instances create \
  --name test \
  --size large \
  --dry-run
```

---

## Performance

### Caching

```bash
# Enable caching
cloud-cli config set cache.enabled true

# Set cache TTL
cloud-cli config set cache.ttl 300  # 5 minutes

# Clear cache
cloud-cli cache clear

# Show cache stats
cloud-cli cache stats
```

### Rate Limiting

```bash
# Limit request rate
cloud-cli --rate-limit 5/s instances list

# Disable rate limiting
cloud-cli --no-rate-limit instances list
```

### Parallel Execution

```bash
# Set max concurrent operations
cloud-cli config set workflows.max_concurrent_operations 10

# Deploy with parallelism
cloud-cli deploy --parallel 5
```

---

## Shell Integration

### Bash Completion

```bash
# Install completion
cloud-cli completion bash > /etc/bash_completion.d/cloud-cli
source ~/.bashrc

# Or user-level
cloud-cli completion bash > ~/.cloud-cli-completion
echo 'source ~/.cloud-cli-completion' >> ~/.bashrc
```

### Zsh Completion

```bash
# Install completion
cloud-cli completion zsh > "${fpath[1]}/_cloud-cli"
compinit
```

### Aliases

```bash
# Common aliases
alias ci='cloud-cli instances'
alias cil='cloud-cli instances list'
alias cic='cloud-cli instances create'

alias cd='cloud-cli deploy'
alias cs='cloud-cli state'
```

---

## Environment Variables

```bash
# Authentication
export CLOUD_API_KEY="your-api-key"
export GITHUB_TOKEN="your-github-token"

# Configuration
export CLI_DEFAULT_REGION="us-west"
export CLI_DEFAULT_FORMAT="json"
export CLI_LOG_LEVEL="debug"

# CI/CD specific
export CLI_CI_MODE="true"
export CLI_NO_COLOR="true"
export CLI_NO_INTERACTIVE="true"
```

---

**Version**: 1.0.0
**Last Updated**: 2025-11-25
**Related**: [Tutorial Series](README.md)
