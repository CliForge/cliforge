# CliForge Quick Reference

Quick lookup guide for common CliForge commands, configuration patterns, and OpenAPI extensions.

---

## Generator CLI Commands

### Initialize New CLI
```bash
cliforge init --name my-api-cli --spec https://api.example.com/openapi.yaml
```

### Build CLI Binary
```bash
cliforge build --config cli-config.yaml --output dist/
cliforge build --all  # Build for all platforms
cliforge build --platform linux/amd64,darwin/arm64
```

### Validate Configuration
```bash
cliforge validate --config cli-config.yaml
```

### Generate Shell Completion
```bash
cliforge completion bash > /etc/bash_completion.d/cliforge
cliforge completion zsh > /usr/local/share/zsh/site-functions/_cliforge
```

---

## Configuration Quick Reference

### Minimal Configuration
```yaml
metadata:
  name: my-cli
  version: 1.0.0
  description: My API CLI

api:
  openapi_url: https://api.example.com/openapi.yaml
  base_url: https://api.example.com
```

### With API Key Auth
```yaml
behaviors:
  auth:
    type: api_key
    api_key:
      header: X-API-Key
      env_var: MY_CLI_API_KEY
```

### With OAuth2
```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      auth_url: https://auth.example.com/authorize
      token_url: https://auth.example.com/token
      client_id: your-client-id
      scopes: ["read", "write"]
```

### User Overridable Defaults
```yaml
defaults:
  http:
    timeout: 30s          # User can override
  output:
    format: json          # User can override
    color: auto           # User can override
    pretty_print: true
  caching:
    enabled: true
    ttl: 5m
  pagination:
    limit: 50             # User can override (max 100)
  retry:
    max_attempts: 3       # User can override
```

---

## OpenAPI Extension Quick Reference

### CLI Configuration
```yaml
info:
  x-cli-config:
    colors:
      primary: "#0066CC"
      success: "#00AA00"
      error: "#CC0000"
```

### Command Aliases
```yaml
paths:
  /users:
    get:
      x-cli-aliases: [list-users, ls-users]
```

### Hide Commands
```yaml
paths:
  /internal:
    get:
      x-cli-hidden: true
```

### Custom Flag Names
```yaml
paths:
  /search:
    get:
      parameters:
        - name: q
          in: query
          x-cli-flag: query  # Use --query instead of --q
```

### Interactive Prompts
```yaml
paths:
  /users:
    post:
      x-cli-interactive:
        - name: email
          prompt: "Enter user email:"
          required: true
```

### Confirmation Prompts
```yaml
paths:
  /users/{id}:
    delete:
      x-cli-confirmation:
        message: "Delete user {id}?"
        flag: force
```

### Async Operations
```yaml
paths:
  /deploy:
    post:
      x-cli-async:
        enabled: true
        status_endpoint: /deployments/{id}/status
        status_field: status
        complete_statuses: [completed, success]
        failed_statuses: [failed, error]
        poll_interval: 2s
        timeout: 5m
```

### Workflow Orchestration
```yaml
paths:
  /complex-operation:
    post:
      x-cli-workflow:
        steps:
          - id: step1
            request:
              method: GET
              url: "{base_url}/check"
          - id: step2
            request:
              method: POST
              url: "{base_url}/action"
            condition: "step1.body.ready == true"
```

---

## Common Flag Patterns

### Global Flags
```bash
mycli --output json          # Output format
mycli --no-color             # Disable colors
mycli --verbose              # Verbose output
mycli -vv                    # Very verbose
mycli --timeout 60s          # Request timeout
mycli --config /path/config  # Custom config
mycli --profile production   # Use profile
```

### Output Formatting
```bash
mycli users list --output json
mycli users list --output yaml
mycli users list --output table
mycli users list -o csv
```

### Pagination
```bash
mycli users list --limit 100
mycli users list --offset 50
mycli users list --auto-page  # Auto-paginate all results
```

### Filtering and Selection
```bash
mycli users list --filter 'status=="active"'
mycli users list --select '.items[].email'
```

---

## Environment Variables

### Configuration
```bash
MY_CLI_CONFIG=/path/to/config.yaml
MY_CLI_PROFILE=production
```

### Authentication
```bash
MY_CLI_API_KEY=your-api-key
MY_CLI_TOKEN=your-bearer-token
```

### Behavior Overrides
```bash
MY_CLI_OUTPUT_FORMAT=json
MY_CLI_TIMEOUT=60s
MY_CLI_NO_COLOR=1
MY_CLI_NO_CACHE=1
```

---

## User Configuration

### Location
```bash
~/.config/mycli/config.yaml     # Linux/macOS
%APPDATA%\mycli\config.yaml     # Windows
```

### Example User Config
```yaml
preferences:
  http:
    timeout: 60s
    proxy: http://proxy.example.com:8080
  output:
    format: yaml
    color: always
  caching:
    enabled: false
```

### Debug Mode (Development Only)
```yaml
# Only works when binary built with metadata.debug: true
debug_override:
  api:
    base_url: http://localhost:8080
  behaviors:
    auth:
      type: none
```

---

## Built-in Commands

### Version Information
```bash
mycli version              # Show client and server versions
mycli version --client     # Client version only
mycli version --server     # Server version only
```

### Configuration Management
```bash
mycli config get output.format
mycli config set output.format yaml
mycli config unset output.format
mycli config edit          # Open in $EDITOR
mycli config path          # Show config file location
```

### Authentication
```bash
mycli auth login
mycli auth logout
mycli auth status
mycli auth refresh
```

### Context Management
```bash
mycli context create production
mycli context use production
mycli context list
mycli context delete staging
```

### History
```bash
mycli history
mycli history --limit 50
mycli history clear
```

### Updates
```bash
mycli update check
mycli update install
mycli update install --force
```

### Changelog
```bash
mycli changelog
mycli changelog --binary-only
mycli changelog --api-only
```

### Deprecations
```bash
mycli deprecations
mycli deprecations --severity urgent
mycli deprecations scan script.sh
```

### Cache Management
```bash
mycli cache clear
mycli cache show
```

---

## Troubleshooting Quick Checks

### Debug Mode
```bash
mycli --debug users list    # Enable debug output
mycli -vv users list        # Very verbose
```

### Check Configuration
```bash
mycli config path           # Find config file
mycli config get            # Show all settings
```

### Network Issues
```bash
mycli --timeout 120s users list
mycli --no-cache users list
MY_CLI_HTTP_PROXY=http://proxy:8080 mycli users list
```

### Authentication Issues
```bash
mycli auth status          # Check auth state
mycli auth logout          # Clear credentials
mycli auth login           # Re-authenticate
```

### Clear Cache
```bash
mycli cache clear          # Clear all caches
rm -rf ~/.cache/mycli      # Manual cache clear
```

---

## Common Workflows

### First-Time Setup
```bash
# 1. Install CLI
curl -sSL https://install.example.com/cli | bash

# 2. Authenticate
mycli auth login

# 3. Verify
mycli version
mycli users list --limit 5
```

### Switching Environments
```bash
# Create contexts for each environment
mycli context create dev --profile dev
mycli context create prod --profile prod

# Switch contexts
mycli context use dev
mycli users list              # Uses dev environment

mycli context use prod
mycli users list              # Uses prod environment
```

### Scripting
```bash
#!/bin/bash
set -e

# Non-interactive mode
mycli --yes delete-resource \
  --id 123 \
  --force \
  --output json

# Check exit code
if [ $? -eq 0 ]; then
  echo "Success"
fi

# Use output in pipeline
mycli users list --output json | jq '.items[].email'
```

---

## Performance Tips

1. **Enable Caching** (default: enabled)
   ```yaml
   defaults:
     caching:
       enabled: true
       ttl: 5m
   ```

2. **Adjust Timeout** for slow operations
   ```bash
   mycli --timeout 300s slow-operation
   ```

3. **Use Auto-Pagination** sparingly
   ```bash
   # Good for small datasets
   mycli users list --auto-page

   # Bad for large datasets (use manual pagination)
   mycli users list --limit 100 --offset 0
   ```

4. **Limit Output Fields**
   ```bash
   mycli users list --select '.items[] | {id, email}'
   ```

---

## Security Best Practices

1. **Never Commit Credentials**
   ```bash
   echo "config.yaml" >> .gitignore
   echo ".env" >> .gitignore
   ```

2. **Use Environment Variables**
   ```bash
   export MY_CLI_API_KEY=$(cat ~/.secrets/api-key)
   mycli users list
   ```

3. **Use System Keyring** (when available)
   ```yaml
   behaviors:
     auth:
       type: oauth2
       oauth2:
         credentials_store: keyring  # Secure storage
   ```

4. **Review Permissions**
   ```bash
   ls -la ~/.config/mycli/
   chmod 600 ~/.config/mycli/config.yaml
   ```

---

## Getting Help

### Command Help
```bash
mycli --help                   # Main help
mycli users --help             # Command help
mycli users create --help      # Subcommand help
```

### Show Examples
```bash
mycli users create --examples
```

### Online Resources
- Documentation: https://docs.cliforge.com
- GitHub: https://github.com/cliforge/cliforge
- Issues: https://github.com/cliforge/cliforge/issues

---

*For detailed documentation, see the [full user guide](README.md).*
