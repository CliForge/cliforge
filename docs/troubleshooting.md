# CliForge Troubleshooting Guide

**Version**: 0.9.0
**Last Updated**: 2025-11-23

Common issues and solutions when using CliForge.

---

## Table of Contents

1. [Common Errors](#common-errors)
2. [OpenAPI Spec Issues](#openapi-spec-issues)
3. [Configuration Errors](#configuration-errors)
4. [Authentication Failures](#authentication-failures)
5. [Plugin Errors](#plugin-errors)
6. [Workflow Failures](#workflow-failures)
7. [Build Errors](#build-errors)
8. [Runtime Errors](#runtime-errors)
9. [Performance Issues](#performance-issues)
10. [Debug Mode Usage](#debug-mode-usage)
11. [Getting Help](#getting-help)

---

## Common Errors

### Error: "command not found: mycli"

**Problem**: CLI binary not in PATH or not installed.

**Solutions**:

**1. Check if installed**:
```bash
which mycli
# If nothing, binary not installed
```

**2. Install CLI**:
```bash
# Direct download
curl -L https://releases.example.com/install.sh | sh

# Or Homebrew
brew install mycli

# Or from binary
sudo mv mycli /usr/local/bin/
sudo chmod +x /usr/local/bin/mycli
```

**3. Add to PATH**:
```bash
# Add to ~/.bashrc or ~/.zshrc
export PATH="$PATH:/path/to/mycli"

# Reload shell
source ~/.bashrc
```

**4. Verify installation**:
```bash
mycli --version
```

---

### Error: "Failed to load OpenAPI spec"

**Problem**: Cannot fetch or parse OpenAPI specification.

**Symptoms**:
```
Error: Failed to load OpenAPI spec
  URL: https://api.example.com/openapi.yaml
  Reason: connection timeout
```

**Solutions**:

**1. Check network connectivity**:
```bash
curl -I https://api.example.com/openapi.yaml
```

**2. Verify URL is correct**:
```bash
mycli config show | grep openapi_url
```

**3. Check if spec is valid**:
```bash
# Validate spec
curl https://api.example.com/openapi.yaml | \
  yq eval '.' -
```

**4. Use local file temporarily**:
```bash
# Download spec
curl https://api.example.com/openapi.yaml > openapi.yaml

# Use local file in user config
# ~/.config/mycli/config.yaml
debug_override:  # Only works in debug builds
  api:
    openapi_url: file://./openapi.yaml
```

**5. Check cache**:
```bash
# Clear cache and retry
mycli cache clear
mycli users list --refresh
```

---

### Error: "Authentication required"

**Problem**: No valid credentials found.

**Symptoms**:
```
Error: Authentication required
  Run 'mycli auth login' to authenticate
```

**Solutions**:

**1. Log in**:
```bash
mycli auth login
```

**2. Check auth status**:
```bash
mycli auth status

# Output:
# Authenticated: No
# Last login: Never
```

**3. Verify API key (if using API key auth)**:
```bash
# Check environment variable
echo $MYCLI_API_KEY

# Or set it
export MYCLI_API_KEY=sk_live_abc123
```

**4. Check stored credentials**:
```bash
# macOS
security find-generic-password -s mycli

# Linux
secret-tool lookup service mycli

# Windows
cmdkey /list | findstr mycli
```

**5. Re-authenticate**:
```bash
mycli auth logout
mycli auth login
```

---

### Error: "Rate limit exceeded"

**Problem**: Too many requests to API.

**Symptoms**:
```
Error: Rate limit exceeded
  Limit: 60 requests/minute
  Retry after: 45 seconds
```

**Solutions**:

**1. Wait and retry**:
```bash
# CLI automatically retries after backoff
# Or wait manually
sleep 45
mycli users list
```

**2. Check rate limit settings**:
```yaml
# ~/.config/mycli/config.yaml
preferences:
  retry:
    max_attempts: 5  # More retries with backoff
```

**3. Reduce concurrent requests**:
```bash
# Avoid parallel calls
mycli users list
mycli orders list

# Instead of
mycli users list & mycli orders list &
```

**4. Use caching**:
```yaml
# ~/.config/mycli/config.yaml
preferences:
  caching:
    enabled: true  # Cache responses
```

---

### Error: "Invalid JSON response"

**Problem**: API returned non-JSON response.

**Symptoms**:
```
Error: Invalid JSON response
  Status: 200 OK
  Content-Type: text/html
  Body: <html>...
```

**Solutions**:

**1. Check API endpoint**:
```bash
# Verify URL
curl -I https://api.example.com/users

# Check Content-Type
# Should be: application/json
```

**2. Check for maintenance page**:
```bash
curl https://api.example.com/users
# If HTML, API might be down
```

**3. Verify base URL**:
```bash
mycli config show | grep base_url
```

**4. Use debug mode**:
```bash
mycli users list --debug

# Shows full request/response
```

---

## OpenAPI Spec Issues

### Error: "OpenAPI spec validation failed"

**Problem**: Spec doesn't conform to OpenAPI standard.

**Symptoms**:
```
Error: OpenAPI spec validation failed
  Line 42: paths./users.get: missing required field 'responses'
  Line 58: components.schemas.User: invalid type 'strng' (did you mean 'string'?)
```

**Solutions**:

**1. Validate spec**:
```bash
# Use online validator
# https://editor.swagger.io/

# Or local validator
npm install -g @apidevtools/swagger-cli
swagger-cli validate openapi.yaml
```

**2. Check common issues**:

**Missing required fields**:
```yaml
# Bad
paths:
  /users:
    get:
      summary: List users
      # Missing 'responses'

# Good
paths:
  /users:
    get:
      summary: List users
      responses:
        '200':
          description: Success
```

**Invalid types**:
```yaml
# Bad
type: strng  # Typo

# Good
type: string
```

**Missing references**:
```yaml
# Bad
schema:
  $ref: '#/components/schemas/User'
# But User schema not defined

# Good
components:
  schemas:
    User:
      type: object
```

**3. Use Swagger 2.0 conversion**:
```bash
# If using Swagger 2.0, convert to OpenAPI 3.0
npm install -g swagger2openapi
swagger2openapi swagger.yaml -o openapi.yaml
```

---

### Error: "x-cli-* extension not supported"

**Problem**: Unknown or invalid CliForge extension.

**Symptoms**:
```
Warning: Unknown extension 'x-cli-custom-field' ignored
```

**Solutions**:

**1. Check extension name**:
```yaml
# Bad
x-cli-custom-field: value

# Good (supported extensions)
x-cli-command: "list users"
x-cli-aliases: ["users"]
x-cli-output: {...}
x-cli-workflow: {...}
```

**2. Verify extension syntax**:
```yaml
# Bad
x-cli-output: "table"

# Good
x-cli-output:
  table:
    columns: [...]
```

**3. Check version compatibility**:
```yaml
# Some extensions require minimum version
info:
  x-cli-min-version: "0.9.0"
```

**4. Refer to documentation**:
- See [Technical Specification](technical-specification.md) for all supported extensions

---

### Error: "Duplicate operationId"

**Problem**: Multiple operations with same operationId.

**Symptoms**:
```
Error: Duplicate operationId 'getUser'
  Found in: GET /users/{id} and GET /accounts/{id}
```

**Solutions**:

**1. Make operationIds unique**:
```yaml
# Bad
paths:
  /users/{id}:
    get:
      operationId: getUser
  /accounts/{id}:
    get:
      operationId: getUser

# Good
paths:
  /users/{id}:
    get:
      operationId: getUserById
  /accounts/{id}:
    get:
      operationId: getAccountById
```

**2. Use x-cli-command for custom names**:
```yaml
paths:
  /users/{id}:
    get:
      operationId: getUserById
      x-cli-command: "get user"

  /accounts/{id}:
    get:
      operationId: getAccountById
      x-cli-command: "get account"
```

---

## Configuration Errors

### Error: "Invalid configuration"

**Problem**: Syntax error in configuration file.

**Symptoms**:
```
Error: Invalid configuration
  File: cli-config.yaml
  Line 23: unexpected character '%'
```

**Solutions**:

**1. Validate YAML syntax**:
```bash
# Use YAML validator
yamllint cli-config.yaml

# Or parse with yq
yq eval '.' cli-config.yaml
```

**2. Check common YAML issues**:

**Indentation**:
```yaml
# Bad (inconsistent indentation)
metadata:
  name: mycli
 version: 1.0.0  # Wrong indentation

# Good
metadata:
  name: mycli
  version: 1.0.0
```

**Quotes**:
```yaml
# Bad (unescaped special chars)
description: CLI for API's

# Good
description: "CLI for API's"
```

**Multiline strings**:
```yaml
# Bad
ascii_art: Line 1
Line 2

# Good
ascii_art: |
  Line 1
  Line 2
```

**3. Validate against schema**:
```bash
cliforge validate cli-config.yaml
```

---

### Error: "Required field missing"

**Problem**: Configuration missing required fields.

**Symptoms**:
```
Error: Required field missing
  Field: metadata.name
  Required by: configuration schema
```

**Solutions**:

**1. Add required fields**:
```yaml
# Minimum required fields
metadata:
  name: mycli         # Required
  version: 1.0.0      # Required
  description: "..."  # Required

api:
  openapi_url: https://...  # Required
  base_url: https://...     # Required

behaviors:
  auth:
    type: none        # Required
```

**2. Check version compatibility**:
```yaml
# Some fields required in newer versions
metadata:
  version: 1.0.0  # Must be semver
```

**3. Refer to schema**:
- See [Configuration DSL](configuration-dsl.md) for all required fields

---

### Error: "Cannot override locked setting"

**Problem**: User config trying to override embedded setting.

**Symptoms**:
```
Warning: Cannot override locked setting 'api.base_url'
  This setting is locked to embedded config
  Ignored: api.base_url in user config
```

**Solutions**:

**1. Check what's overridable**:
```bash
mycli config show --overridable

# Lists all user-overridable settings
```

**2. Use allowed overrides**:
```yaml
# ~/.config/mycli/config.yaml

# ✅ Allowed overrides
preferences:
  http:
    timeout: 60s
  output:
    format: yaml
  caching:
    enabled: false

# ❌ Not allowed (locked to embedded)
# api:
#   base_url: http://localhost:8080
```

**3. Use debug mode for testing**:
```yaml
# Only works if binary built with debug: true
debug_override:
  api:
    base_url: http://localhost:8080
```

**Security note**: Debug mode shows warning on every command and should only be used in development builds.

---

## Authentication Failures

### Error: "OAuth2 callback timeout"

**Problem**: Browser didn't complete OAuth2 flow in time.

**Symptoms**:
```
Error: OAuth2 authentication timeout
  Waited: 120 seconds
  No callback received
```

**Solutions**:

**1. Check browser opened**:
```bash
# CLI should open browser automatically
# If not, copy URL manually
mycli auth login --no-browser

# Output:
# Visit: https://auth.example.com/authorize?...
```

**2. Verify callback URL**:
```yaml
# cli-config.yaml
behaviors:
  auth:
    oauth2:
      redirect_url: http://localhost:8085/callback
```

**3. Check firewall**:
```bash
# Ensure localhost:8085 not blocked
curl http://localhost:8085/health
```

**4. Use different port**:
```yaml
# ~/.config/mycli/config.yaml
preferences:
  auth:
    oauth2:
      callback_port: 9000  # If 8085 blocked
```

**5. Check browser cookies**:
- Clear cookies for auth domain
- Try incognito mode

---

### Error: "Token expired"

**Problem**: Stored access token expired.

**Symptoms**:
```
Error: Token expired
  Expired at: 2025-11-23 14:30:00
  Run 'mycli auth refresh' or 'mycli auth login'
```

**Solutions**:

**1. Refresh token**:
```bash
mycli auth refresh
```

**2. Re-authenticate**:
```bash
mycli auth logout
mycli auth login
```

**3. Enable auto-refresh**:
```yaml
# cli-config.yaml (embedded)
behaviors:
  auth:
    oauth2:
      auto_refresh: true
      refresh_threshold: 300  # 5 minutes
```

**4. Check token storage**:
```bash
mycli auth status

# Output:
# Authenticated: No
# Token expired: Yes
# Last refresh: 2025-11-23 10:00:00
```

---

### Error: "Invalid API key"

**Problem**: API key incorrect or malformed.

**Symptoms**:
```
Error: Invalid API key
  Status: 401 Unauthorized
  Message: API key format invalid
```

**Solutions**:

**1. Verify API key format**:
```bash
# Check key format (example: sk_live_...)
echo $MYCLI_API_KEY

# Should match pattern
# sk_live_[a-zA-Z0-9]{32}
```

**2. Regenerate API key**:
```bash
# In API dashboard, revoke and create new key
```

**3. Check environment variable**:
```bash
# Verify variable name
mycli config show | grep env_var

# Set correct variable
export MYCLI_API_KEY=sk_live_abc123xyz789
```

**4. Test API key directly**:
```bash
curl -H "X-API-Key: $MYCLI_API_KEY" \
     https://api.example.com/users
```

---

### Error: "Permission denied"

**Problem**: Insufficient permissions for operation.

**Symptoms**:
```
Error: Permission denied
  Status: 403 Forbidden
  Required scope: api:write
  Current scopes: api:read
```

**Solutions**:

**1. Check required scopes**:
```bash
mycli auth status

# Output:
# Scopes: api:read
# Required: api:read, api:write
```

**2. Re-authenticate with correct scopes**:
```bash
mycli auth logout
mycli auth login
# Grant additional permissions in browser
```

**3. Verify OAuth2 scopes in config**:
```yaml
behaviors:
  auth:
    oauth2:
      scopes:
        - api:read
        - api:write  # Add missing scope
```

---

## Plugin Errors

### Error: "Plugin not found"

**Problem**: Required external command not installed.

**Symptoms**:
```
Error: Required plugin 'aws' not found
  Install: https://aws.amazon.com/cli/
  Version: >= 2.0.0
```

**Solutions**:

**1. Install plugin**:
```bash
# AWS CLI
brew install awscli

# Or download from official site
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
unzip awscliv2.zip
sudo ./aws/install
```

**2. Verify installation**:
```bash
which aws
aws --version
```

**3. Add to PATH**:
```bash
export PATH="$PATH:/path/to/aws"
```

**4. Skip plugin (if optional)**:
```bash
# If plugin is optional, CLI continues
# Check x-cli-plugin.required: false
```

---

### Error: "Plugin version mismatch"

**Problem**: Installed plugin version too old.

**Symptoms**:
```
Error: Plugin version mismatch
  Plugin: aws
  Installed: 1.18.0
  Required: >= 2.0.0
```

**Solutions**:

**1. Upgrade plugin**:
```bash
# Homebrew
brew upgrade awscli

# Or download latest
curl "https://awscli.amazonaws.com/awscli-exe-linux-x86_64.zip" -o "awscliv2.zip"
```

**2. Check installed version**:
```bash
aws --version
```

**3. Verify requirement**:
```yaml
x-cli-plugin:
  command: aws
  min-version: "2.0.0"  # Check if this is correct
```

---

### Error: "Plugin execution failed"

**Problem**: External command returned error.

**Symptoms**:
```
Error: Plugin execution failed
  Command: aws s3 cp file.json s3://bucket/
  Exit code: 1
  Stderr: An error occurred (AccessDenied)
```

**Solutions**:

**1. Check plugin credentials**:
```bash
# AWS CLI
aws sts get-caller-identity

# Configure if needed
aws configure
```

**2. Run command manually**:
```bash
# Test plugin command directly
aws s3 cp file.json s3://bucket/

# Fix any errors
```

**3. Check environment variables**:
```yaml
x-cli-plugin:
  operations:
    - plugin-call:
        env:
          AWS_REGION: "us-east-1"  # Ensure set
          AWS_ACCESS_KEY_ID: "{env.AWS_ACCESS_KEY_ID}"
```

**4. Enable verbose output**:
```bash
mycli backup database --verbose

# Shows full plugin command and output
```

---

## Workflow Failures

### Error: "Workflow step failed"

**Problem**: Step in workflow returned error.

**Symptoms**:
```
Error: Workflow step 'create-deployment' failed
  Status: 400 Bad Request
  Message: Invalid parameter 'version'
```

**Solutions**:

**1. Check step request**:
```yaml
steps:
  - id: create-deployment
    request:
      method: POST
      url: "{base_url}/deployments"
      body:
        version: "{args.version}"  # Check parameter
```

**2. Verify previous step output**:
```bash
mycli deploy app --debug

# Shows each step's request/response
```

**3. Test step individually**:
```bash
# Call API directly
curl -X POST https://api.example.com/deployments \
  -H "Content-Type: application/json" \
  -d '{"version":"v1.0.0"}'
```

**4. Add error handling**:
```yaml
steps:
  - id: create-deployment
    request: {...}
    condition: "true"  # Make step optional

  - id: next-step
    request: {...}
    # Use optional chaining
    body:
      deployment_id: "{create-deployment.body?.id ?? 'default'}"
```

---

### Error: "Workflow timeout"

**Problem**: Workflow polling exceeded timeout.

**Symptoms**:
```
Error: Workflow timeout
  Step: wait-deployment
  Waited: 600 seconds
  Status: in_progress (not terminal)
```

**Solutions**:

**1. Increase timeout**:
```yaml
steps:
  - id: wait-deployment
    polling:
      interval: 10
      timeout: 1200  # Increase from 600 to 1200 seconds
```

**2. Check terminal condition**:
```yaml
polling:
  terminal-condition: "response.body.status in ['success', 'failed']"
  # Ensure this condition can be met
```

**3. Verify API status**:
```bash
# Check deployment status manually
curl https://api.example.com/deployments/dep-123

# Ensure 'status' field exists and updates
```

**4. Reduce polling interval**:
```yaml
polling:
  interval: 5  # Check more frequently
```

---

### Error: "Invalid workflow expression"

**Problem**: Syntax error in workflow expression.

**Symptoms**:
```
Error: Invalid workflow expression
  Expression: check-readiness.body.ready = true
  Error: unexpected token '=' (did you mean '=='?)
```

**Solutions**:

**1. Fix expression syntax**:
```yaml
# Bad
condition: "check-readiness.body.ready = true"

# Good
condition: "check-readiness.body.ready == true"
```

**2. Check expr syntax**:
```yaml
# Comparison operators
==  # Equal
!=  # Not equal
>   # Greater than
<   # Less than
>=  # Greater or equal
<=  # Less or equal

# Logical operators
&&  # And
||  # Or
!   # Not

# Functions
len(array)
filter(array, predicate)
map(array, transform)
```

**3. Test expression**:
```bash
# Use debug mode to see evaluated expressions
mycli deploy app --debug
```

**4. Reference documentation**:
- See [expr language reference](https://expr-lang.org/docs/language-definition)

---

## Build Errors

### Error: "cliforge: command not found"

**Problem**: CliForge generator not installed.

**Solutions**:

**1. Install CliForge**:
```bash
# From source
git clone https://github.com/cliforge/cliforge.git
cd cliforge
go install ./cmd/cliforge

# Or download binary
curl -L https://releases.cliforge.org/install.sh | sh
```

**2. Add to PATH**:
```bash
export PATH="$PATH:$GOPATH/bin"
```

**3. Verify installation**:
```bash
cliforge --version
```

---

### Error: "Build failed: validation error"

**Problem**: Configuration validation failed during build.

**Symptoms**:
```
Error: Build failed
  Validation errors:
    - metadata.version: invalid semver format
    - api.base_url: must be a valid URL
```

**Solutions**:

**1. Validate config first**:
```bash
cliforge validate cli-config.yaml
```

**2. Fix validation errors**:
```yaml
# Bad
metadata:
  version: 1.0  # Invalid semver

# Good
metadata:
  version: 1.0.0  # Valid semver
```

**3. Check all required fields**:
```bash
cliforge validate --verbose cli-config.yaml

# Shows all validation errors
```

---

### Error: "Embed failed: file not found"

**Problem**: Referenced file doesn't exist.

**Symptoms**:
```
Error: Embed failed
  File: logo.png not found
  Referenced in: branding.logo
```

**Solutions**:

**1. Check file path**:
```bash
ls -la logo.png
```

**2. Use absolute or relative path**:
```yaml
# Bad
branding:
  logo: logo.png  # Relative to what?

# Good
branding:
  logo: ./assets/logo.png  # Relative to config file
```

**3. Verify file exists during build**:
```bash
# Build from config directory
cd /path/to/config
cliforge build --config cli-config.yaml
```

---

## Runtime Errors

### Error: "Panic: runtime error"

**Problem**: Unexpected runtime error.

**Symptoms**:
```
panic: runtime error: invalid memory address or nil pointer dereference
```

**Solutions**:

**1. Capture panic details**:
```bash
mycli users list --debug > debug.log 2>&1
```

**2. Check for nil responses**:
```yaml
# Use optional chaining
transform: |
  {
    "user": response.body?.user ?? null
  }
```

**3. Report bug**:
- Include debug log
- Include command that triggered panic
- Include CLI version (`mycli --version`)

---

### Error: "Segmentation fault"

**Problem**: Critical memory error.

**Solutions**:

**1. Update CLI**:
```bash
mycli update
```

**2. Reinstall**:
```bash
# Remove old installation
sudo rm /usr/local/bin/mycli

# Reinstall
curl -L https://releases.example.com/install.sh | sh
```

**3. Report bug**:
- Critical error requiring investigation
- Include system info (`uname -a`)

---

## Performance Issues

### Issue: "Slow startup time"

**Problem**: CLI takes several seconds to start.

**Solutions**:

**1. Enable spec caching**:
```yaml
behaviors:
  caching:
    spec_ttl: 5m  # Cache spec for 5 minutes
```

**2. Disable update checks**:
```bash
# Temporarily
mycli users list --no-update

# Permanently (user config)
preferences:
  updates:
    check_interval: 168h  # Weekly instead of daily
```

**3. Use local spec file**:
```yaml
# For development
debug_override:
  api:
    openapi_url: file://./openapi.yaml
```

**4. Profile startup**:
```bash
time mycli users list --no-update --debug
```

---

### Issue: "High memory usage"

**Problem**: CLI using excessive memory.

**Solutions**:

**1. Limit response size**:
```bash
# Use pagination
mycli users list --limit 20

# Instead of
mycli users list --limit 1000
```

**2. Clear cache**:
```bash
mycli cache clear
```

**3. Reduce cache size**:
```yaml
behaviors:
  caching:
    max_size: 50MB  # Reduce from default 100MB
```

**4. Monitor memory**:
```bash
# macOS/Linux
ps aux | grep mycli

# Check RSS (resident set size)
```

---

## Debug Mode Usage

### Enable Debug Mode

**1. Via flag**:
```bash
mycli users list --debug
```

**2. Via environment**:
```bash
export MYCLI_DEBUG=true
mycli users list
```

**3. Via config**:
```yaml
# ~/.config/mycli/config.yaml
preferences:
  debug:
    enabled: true
```

---

### Debug Output

**Debug mode shows**:
- Full HTTP requests (method, URL, headers, body)
- Full HTTP responses (status, headers, body)
- Workflow step execution
- Expression evaluation
- Cache hits/misses
- Auth token refresh
- Plugin command execution

**Example**:
```bash
mycli users list --debug

# Output:
# DEBUG: Loading OpenAPI spec from cache
# DEBUG: Cache hit: ~/.cache/mycli/spec.yaml (age: 2m)
# DEBUG: Building command tree...
# DEBUG: Executing: GET /users
# DEBUG: Request:
#   GET https://api.example.com/users
#   Headers:
#     Authorization: Bearer eyJ...
#     User-Agent: mycli/1.0.0
# DEBUG: Response:
#   Status: 200 OK
#   Headers:
#     Content-Type: application/json
#   Body: {"data":[...]}
# DEBUG: Formatting output as table
```

---

### Verbose Mode

**Less detailed than debug, but useful**:

```bash
mycli users list --verbose

# Shows:
# → GET https://api.example.com/users
# ← 200 OK (245ms)
# Found 15 users
```

---

### Dry-Run Mode

**Preview request without executing**:

```bash
mycli create user --name John --dry-run

# Shows:
# Would execute:
#   POST https://api.example.com/users
#   Body: {"name":"John"}
#
# (Not executed - dry run mode)
```

---

## Getting Help

### Check Documentation

**1. Built-in help**:
```bash
# General help
mycli --help

# Command help
mycli users --help
mycli users list --help

# Command examples
mycli users list --help | grep -A 10 EXAMPLES
```

**2. Online documentation**:
- [Examples and Recipes](examples-and-recipes.md)
- [FAQ](faq.md)
- [Configuration DSL](configuration-dsl.md)
- [Technical Specification](technical-specification.md)

---

### Community Support

**1. GitHub Issues**:
```
https://github.com/cliforge/cliforge/issues
```

**Search existing issues**:
- Check if already reported
- Look for workarounds

**Create new issue**:
- Include CLI version (`mycli --version`)
- Include debug output (`--debug`)
- Include minimal reproduction steps

**2. Discussions**:
```
https://github.com/cliforge/cliforge/discussions
```

---

### Reporting Bugs

**Include in bug report**:

**1. System information**:
```bash
mycli --version
uname -a
```

**2. Configuration** (sanitized):
```bash
mycli config show
# Remove sensitive data (API keys, tokens)
```

**3. Debug output**:
```bash
mycli [command] --debug > debug.log 2>&1
```

**4. Steps to reproduce**:
```
1. Run mycli users list
2. See error: ...
```

**5. Expected vs actual behavior**:
```
Expected: List of users
Actual: Error "Invalid JSON response"
```

---

### Security Issues

**For security vulnerabilities**:

**DO**:
- Email security@cliforge.org (if available)
- Use private GitHub security advisory
- Wait for fix before public disclosure

**DON'T**:
- Post in public issues
- Share exploit details publicly
- Disclose before fix available

---

### Feature Requests

**Before requesting**:
1. Check existing issues
2. Search discussions
3. Review roadmap

**When requesting**:
- Describe use case
- Explain why current features insufficient
- Provide examples
- Consider contributing

---

## Common Solutions Checklist

When troubleshooting, try these steps in order:

**1. Basic checks**:
- [ ] CLI installed and in PATH?
- [ ] Latest version? (`mycli update`)
- [ ] Network connectivity?
- [ ] API endpoint accessible?

**2. Configuration**:
- [ ] Valid configuration? (`cliforge validate`)
- [ ] OpenAPI spec valid?
- [ ] Credentials configured?

**3. Cache**:
- [ ] Clear cache? (`mycli cache clear`)
- [ ] Refresh spec? (`mycli --refresh`)

**4. Authentication**:
- [ ] Logged in? (`mycli auth status`)
- [ ] Token expired?
- [ ] Correct permissions?

**5. Debug**:
- [ ] Enable debug mode? (`--debug`)
- [ ] Check logs?
- [ ] Test with verbose? (`--verbose`)

**6. Isolation**:
- [ ] Works with curl?
- [ ] Works in browser?
- [ ] Works with different CLI version?

**7. Report**:
- [ ] Search existing issues?
- [ ] Create bug report with details?

---

## Next Steps

- **[Examples and Recipes](examples-and-recipes.md)** - Practical examples
- **[FAQ](faq.md)** - Common questions
- **[Configuration DSL](configuration-dsl.md)** - Full configuration reference
- **[Technical Specification](technical-specification.md)** - Architecture details

---

**Version**: 0.9.0
**Last Updated**: 2025-11-23
