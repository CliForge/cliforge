# CliForge FAQ

**Version**: 0.9.0
**Last Updated**: 2025-11-23

Frequently asked questions about CliForge.

---

## Table of Contents

1. [General Questions](#general-questions)
2. [Comparison with Alternatives](#comparison-with-alternatives)
3. [Configuration Questions](#configuration-questions)
4. [Authentication Questions](#authentication-questions)
5. [Workflow Questions](#workflow-questions)
6. [Plugin Questions](#plugin-questions)
7. [Deployment Questions](#deployment-questions)
8. [Performance Questions](#performance-questions)
9. [Security Questions](#security-questions)
10. [Licensing Questions](#licensing-questions)

---

## General Questions

### What is CliForge?

CliForge is a hybrid CLI generation system that creates branded, self-updating command-line tools from OpenAPI specifications. Unlike static generators that require rebuilding for API changes, or dynamic loaders that can't be customized, CliForge provides the best of both worlds:

- **Static**: Branded binaries with embedded configuration
- **Dynamic**: Runtime loading of OpenAPI specs for API changes
- **Self-updating**: Automatic updates for security patches and features

**Key Features**:
- Generate branded CLIs from OpenAPI specs
- Self-updating binaries
- OAuth2, API key, and basic authentication
- Multi-step workflows
- Plugin integration
- Interactive mode
- Streaming and watch mode
- Custom output formats

---

### When should I use CliForge?

**Use CliForge when you want to**:
- Provide a CLI for your REST API
- Distribute a branded CLI tool
- Support multiple environments (dev/staging/prod)
- Keep CLIs updated automatically
- Integrate with other tools via plugins
- Support complex workflows
- Provide interactive user experience

**Don't use CliForge if**:
- You need a one-off API call (use `curl` instead)
- You need to generate client libraries for programming languages
- Your API doesn't have an OpenAPI specification
- You need real-time performance (CliForge adds ~10ms overhead)

---

### What makes CliForge different from other tools?

| Feature | CliForge | OpenAPI Generator | Restish | AWS CLI |
|---------|----------|-------------------|---------|---------|
| **Branded binaries** | ✅ | ❌ | ❌ | ✅ |
| **Dynamic spec loading** | ✅ | ❌ | ✅ | N/A |
| **Self-updating** | ✅ | ❌ | ❌ | ✅ |
| **Custom branding** | ✅ | ❌ | ❌ | N/A |
| **Multi-step workflows** | ✅ | ❌ | ❌ | ✅ |
| **Plugin system** | ✅ | ❌ | ❌ | ✅ |
| **OAuth2 support** | ✅ | ❌ | ✅ | ✅ |
| **Interactive mode** | ✅ | ❌ | ❌ | ✅ |
| **OpenAPI-driven** | ✅ | ✅ | ✅ | ❌ |
| **No rebuild for API changes** | ✅ | ❌ | ✅ | N/A |

**Key Innovation**: CliForge separates binary-level concerns (branding, URLs) from API-level concerns (endpoints, operations), allowing you to update your API without rebuilding the CLI.

---

### Does CliForge support Swagger 2.0?

**Yes!** CliForge supports both:
- **Swagger 2.0** (OpenAPI 2.0)
- **OpenAPI 3.0**
- **OpenAPI 3.1** (in progress)

Swagger 2.0 specs are automatically converted to OpenAPI 3.0 internally, so you don't need to worry about compatibility.

**Example**:
```yaml
# Swagger 2.0 spec
swagger: "2.0"
info:
  title: My API
  version: 1.0.0
host: api.example.com
basePath: /v1
schemes:
  - https
```

CliForge automatically converts this to:
```yaml
# OpenAPI 3.0 (internal)
openapi: 3.0.0
info:
  title: My API
  version: 1.0.0
servers:
  - url: https://api.example.com/v1
```

---

### How does CliForge compare to code generators?

**OpenAPI Generator** (static generation):
- **Pros**: Type-safe client libraries, works offline
- **Cons**: Requires rebuild for API changes, no CLI-specific features
- **Use case**: Generate client libraries for apps

**CliForge** (hybrid generation):
- **Pros**: Branded CLIs, self-updating, no rebuild for API changes
- **Cons**: Requires network for spec updates, CLI-only (not for libraries)
- **Use case**: Distribute CLI tools to users

**When to use both**:
- Use **OpenAPI Generator** for client libraries in your application code
- Use **CliForge** for CLI tools for your users/operators

---

### Can I use CliForge for internal tools?

**Absolutely!** CliForge is great for:
- **Internal APIs**: DevOps tools, admin panels, automation scripts
- **Enterprise tools**: Database management, infrastructure provisioning
- **Developer tools**: Testing, debugging, data migration

**Example**: Internal cloud management CLI
```yaml
metadata:
  name: internal-cloud-cli
  version: 1.0.0
  description: Internal Cloud Management CLI

api:
  openapi_url: https://internal-api.company.internal/openapi.yaml
  base_url: https://internal-api.company.internal

behaviors:
  auth:
    type: api_key
    api_key:
      header: X-API-Key
      env_var: COMPANY_API_KEY
```

---

### What platforms does CliForge support?

**Supported platforms**:
- **Linux**: x86_64, ARM64
- **macOS**: x86_64 (Intel), ARM64 (Apple Silicon)
- **Windows**: x86_64

**Distribution methods**:
- Direct download (curl install script)
- Homebrew (macOS/Linux)
- APT (Debian/Ubuntu)
- Scoop (Windows)
- Docker containers

---

### Does CliForge require Go knowledge?

**No!** CliForge is designed for API developers who may not know Go:

1. **Configuration**: YAML-based, no code required
2. **OpenAPI**: Standard OpenAPI spec with custom extensions
3. **Generation**: `cliforge build --config cli-config.yaml`
4. **Distribution**: Pre-built binaries

You only need Go knowledge if you want to:
- Contribute to CliForge itself
- Build custom plugins
- Extend the core functionality

---

### Can I customize the CLI behavior?

**Yes!** Customization options:

**1. Branding**:
- ASCII art banner
- Custom colors
- Custom prompts

**2. Commands**:
- Command aliases
- Hidden commands
- Custom output formats

**3. Authentication**:
- API key, OAuth2, Basic
- Custom headers
- Environment variables

**4. Behaviors**:
- Caching (spec and response TTLs)
- Retry logic
- Rate limiting
- Output formats

**5. Features**:
- Interactive mode
- Watch mode
- Streaming
- Workflows
- Plugins

See [Configuration DSL](configuration-dsl.md) for all options.

---

## Comparison with Alternatives

### CliForge vs OpenAPI Generator

**OpenAPI Generator**:
- **Purpose**: Generate client libraries for programming languages
- **Output**: Source code (Go, Java, Python, etc.)
- **Usage**: Import as library in application code
- **Updates**: Regenerate and recompile for changes
- **Customization**: Limited to templates

**CliForge**:
- **Purpose**: Generate CLI tools for end users
- **Output**: Compiled binaries
- **Usage**: Standalone command-line tool
- **Updates**: Self-updating, no regeneration needed
- **Customization**: Full branding, behaviors, workflows

**When to use OpenAPI Generator**:
- Need client libraries for application code
- Building SDKs for your API
- Need type-safe code generation

**When to use CliForge**:
- Building CLI tools for users
- Need branded, distributable binaries
- Want self-updating capabilities

---

### CliForge vs Restish

**Restish**:
- **Approach**: Dynamic spec loader
- **Pros**: No build step, instant API changes
- **Cons**: Cannot be branded or customized, no self-update

**CliForge**:
- **Approach**: Hybrid (static binary + dynamic spec)
- **Pros**: Branded binaries, self-updating, full customization
- **Cons**: Requires build step for branding changes

**When to use Restish**:
- Quick API testing and exploration
- Personal use, no branding needed
- One-off API interactions

**When to use CliForge**:
- Production CLI tools for users
- Branded, professional tools
- Long-term support and updates

---

### CliForge vs AWS CLI

**AWS CLI**:
- **Scope**: AWS-specific services
- **Approach**: Hand-crafted CLI with service models
- **Pros**: Rich features, mature ecosystem
- **Cons**: Not OpenAPI-driven, AWS-only

**CliForge**:
- **Scope**: Any REST API with OpenAPI spec
- **Approach**: OpenAPI-driven generation
- **Pros**: Works with any API, automated generation
- **Cons**: Less mature than AWS CLI

**Similarities**:
- Both support profiles, regions, output formats
- Both support plugins
- Both self-update

**When to use AWS CLI**:
- Managing AWS resources
- Need AWS-specific features

**When to use CliForge**:
- Building CLIs for your own API
- Need OpenAPI-driven approach
- Want automated CLI generation

---

### CliForge vs curl + jq

**curl + jq**:
- **Approach**: Manual API calls with shell scripts
- **Pros**: Simple, no dependencies, full control
- **Cons**: Verbose, error-prone, no auth management

**CliForge**:
- **Approach**: Generated CLI with built-in features
- **Pros**: User-friendly, auth management, output formatting
- **Cons**: Requires setup, adds overhead

**Migration example**:

**Before (curl + jq)**:
```bash
API_KEY="sk_live_abc123"
curl -H "X-API-Key: $API_KEY" \
     "https://api.example.com/users" \
     | jq '.data[] | {id, name, email}'
```

**After (CliForge)**:
```bash
mycli users list --output table
```

---

## Configuration Questions

### How do I configure authentication?

CliForge supports three authentication types:

**1. API Key**:
```yaml
behaviors:
  auth:
    type: api_key
    api_key:
      header: X-API-Key
      env_var: MYCLI_API_KEY
```

Usage:
```bash
export MYCLI_API_KEY=sk_live_abc123
mycli users list
```

**2. OAuth2**:
```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      client_id: mycli-client
      auth_url: https://auth.example.com/authorize
      token_url: https://auth.example.com/token
      scopes: [api:read, api:write]
```

Usage:
```bash
mycli auth login  # Opens browser
mycli users list
```

**3. Basic Auth**:
```yaml
behaviors:
  auth:
    type: basic
    basic:
      username_env: MYCLI_USERNAME
      password_env: MYCLI_PASSWORD
```

---

### Can I override configuration at runtime?

**Yes!** Configuration priority (highest to lowest):

1. **Environment variables**: `MYCLI_TIMEOUT=60s`
2. **Command-line flags**: `--timeout 60s`
3. **User config file**: `~/.config/mycli/config.yaml`
4. **Embedded config**: Built into binary
5. **Built-in defaults**: Hardcoded

**Example**:
```bash
# Override timeout via environment
MYCLI_TIMEOUT=120s mycli users list

# Override via flag
mycli users list --timeout 120s

# Override in user config
# ~/.config/mycli/config.yaml
preferences:
  http:
    timeout: 120s
```

---

### How do I support multiple environments?

**Option 1: Multiple environments in config**:
```yaml
api:
  environments:
    - name: production
      base_url: https://api.example.com
      default: true

    - name: staging
      base_url: https://staging-api.example.com

    - name: development
      base_url: http://localhost:8080
```

Usage:
```bash
mycli users list                        # Uses production (default)
mycli --env staging users list          # Uses staging
mycli --env development users list      # Uses development
```

**Option 2: User config profiles**:
```yaml
# ~/.config/mycli/config.yaml
profiles:
  production:
    base_url: https://api.example.com
    api_key: sk_prod_xxx

  staging:
    base_url: https://staging-api.example.com
    api_key: sk_stag_xxx
```

Usage:
```bash
mycli --profile production users list
mycli --profile staging users list
```

---

### What can users override in their config?

**User-overridable settings** (via `preferences` section):
- `http.timeout` - Request timeout
- `http.proxy` - HTTP proxy settings
- `caching.enabled` - Enable/disable caching
- `pagination.limit` - Default page size
- `output.*` - Output format, colors, paging
- `deprecations.*` - Warning preferences
- `retry.max_attempts` - Number of retries
- `telemetry.enabled` - Telemetry opt-in
- `updates.auto_install` - Auto-update opt-in

**Locked settings** (embedded only):
- `api.*` - API URLs and endpoints
- `metadata.*` - CLI name, version
- `branding.*` - Colors, ASCII art
- `behaviors.auth.*` - Auth configuration
- `updates.update_url` - Update server URL

See [Configuration DSL](configuration-dsl.md) for complete list.

---

## Authentication Questions

### How is OAuth2 implemented?

**OAuth2 Flow**:
1. User runs `mycli auth login`
2. CLI opens browser to authorization URL
3. User logs in and grants permissions
4. Browser redirects to `http://localhost:8085/callback`
5. CLI exchanges authorization code for access token
6. Token stored in OS-specific secure storage:
   - **macOS**: Keychain
   - **Linux**: Secret Service API (gnome-keyring, kwallet)
   - **Windows**: Credential Manager

**Token Refresh**:
- CLI automatically refreshes tokens before expiry
- Refresh threshold: 5 minutes (configurable)
- Silent refresh (no user interaction)

**Token Storage**:
```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      auto_refresh: true
      refresh_threshold: 300  # 5 minutes

      # Primary: OS keyring
      storage:
        primary: keyring
        fallback: file

      # Fallback: Encrypted file
      file:
        path: "~/.mycli/credentials.json"
        permissions: "0600"
```

---

### Where are credentials stored?

**OS-Specific Secure Storage**:

**macOS** (Keychain):
```
Service: mycli
Account: default
Password: {access_token}
```

**Linux** (Secret Service):
```
Collection: Login
Label: mycli-token
Secret: {access_token}
```

**Windows** (Credential Manager):
```
Target: mycli
Username: default
Password: {access_token}
```

**Fallback** (Encrypted file):
```
~/.mycli/credentials.json (chmod 0600)
```

**Check credential location**:
```bash
mycli auth status

# Output:
# Authenticated: Yes
# Token expires: 2025-11-24 14:30:00
# Storage: keychain (macOS)
```

---

### Can I use multiple API keys?

**Yes!** Use profiles:

```yaml
# ~/.config/mycli/config.yaml
profiles:
  personal:
    api_key: sk_personal_abc123

  work:
    api_key: sk_work_xyz789

  staging:
    api_key: sk_stag_def456
```

Usage:
```bash
# Use personal API key
mycli --profile personal users list

# Use work API key
mycli --profile work users list

# Use staging API key
mycli --profile staging users list
```

---

## Workflow Questions

### What are workflows?

Workflows enable **multi-step operations** that call multiple APIs in sequence, with conditional logic and data transformation.

**Use cases**:
- Deploy application (validate → create → wait → verify)
- Migrate data (export → validate → import → verify)
- Backup database (backup → upload to S3 → verify)

**Example**:
```yaml
x-cli-workflow:
  steps:
    - id: check-readiness
      request:
        method: GET
        url: "{base_url}/readiness"

    - id: create-deployment
      request:
        method: POST
        url: "{base_url}/deployments"
      condition: "check-readiness.body.ready == true"

    - id: wait-complete
      request:
        method: GET
        url: "{base_url}/deployments/{create-deployment.body.id}"
      polling:
        interval: 10
        terminal-condition: "response.body.status == 'complete'"
```

---

### How do I reference previous step results?

Use the step `id` to reference results:

**Syntax**:
- `{step-id.body.field}` - Response body field
- `{step-id.status}` - HTTP status code
- `{step-id.headers.Name}` - Response header

**Example**:
```yaml
steps:
  - id: create-user
    request:
      method: POST
      url: "{base_url}/users"
      body:
        name: "John"

  - id: get-user
    request:
      method: GET
      url: "{base_url}/users/{create-user.body.id}"

  - id: assign-role
    request:
      method: POST
      url: "{base_url}/users/{create-user.body.id}/roles"
      body:
        role: "admin"
```

---

### Can workflows call external APIs?

**Yes!** Workflows can call any HTTP endpoint:

```yaml
steps:
  # Step 1: Get data from your API
  - id: get-data
    request:
      method: GET
      url: "{base_url}/data"

  # Step 2: Call external AWS API
  - id: upload-s3
    request:
      method: PUT
      url: "https://s3.amazonaws.com/bucket/file.json"
      headers:
        Authorization: "AWS {env.AWS_ACCESS_KEY}"
      body: "{get-data.body}"

  # Step 3: Call GitHub API
  - id: create-issue
    request:
      method: POST
      url: "https://api.github.com/repos/owner/repo/issues"
      headers:
        Authorization: "token {env.GITHUB_TOKEN}"
      body:
        title: "Data uploaded"
```

---

### How do I handle errors in workflows?

**Error Handling**:

**1. Conditional steps** (skip if error):
```yaml
steps:
  - id: optional-step
    request:
      method: GET
      url: "{base_url}/optional"
    condition: "true"  # Always run, but don't fail workflow

  - id: next-step
    request:
      method: POST
      url: "{base_url}/next"
      body:
        data: "{optional-step.body?.data ?? 'default'}"
```

**2. Rollback on failure**:
```yaml
x-cli-workflow:
  steps:
    - id: create-resource
      ...

  rollback:
    enabled: true
    steps:
      - description: "Rolling back..."
        request:
          method: DELETE
          url: "{base_url}/resources/{create-resource.body.id}"
```

---

## Plugin Questions

### What are plugins?

Plugins allow CLI to **execute external commands** as part of workflows.

**Use cases**:
- Integrate with AWS CLI, kubectl, docker
- Run custom scripts
- Call system utilities
- Upload files to S3, GCS

**Types**:
- **External**: Call existing CLI tools (aws, kubectl, docker)
- **Built-in**: Pre-packaged plugins (future)

---

### How do I use AWS CLI plugin?

**OpenAPI Extension**:
```yaml
paths:
  /databases/{dbId}/backup:
    post:
      operationId: backupDatabase
      x-cli-command: "backup database"

      x-cli-plugin:
        type: external
        command: aws
        required: true
        min-version: "2.0.0"

        operations:
          # Call API
          - api-call:
              endpoint: "/databases/{dbId}/backup"
              method: POST
              output-var: backup

          # Upload to S3
          - plugin-call:
              command: "aws"
              args:
                - "s3"
                - "cp"
                - "{backup.body.file}"
                - "s3://bucket/backups/{dbId}.sql"
              env:
                AWS_REGION: "us-east-1"
```

**Usage**:
```bash
# Backup database (calls API + AWS CLI)
mycli backup database --db-id prod-db

# Output:
# ✓ Creating backup...
# ✓ Uploading to S3... (aws s3 cp)
# ✓ Backup complete: s3://bucket/backups/prod-db.sql
```

---

### Can I create custom plugins?

**Not yet in v0.9.0**, but planned for future:

**Future plugin types**:
- **Binary plugins**: Standalone executables
- **WASM plugins**: WebAssembly modules (sandboxed)
- **Script plugins**: Shell scripts

**Example (future)**:
```yaml
x-cli-plugin:
  type: binary
  path: ~/.mycli/plugins/custom-processor
  permissions:
    network: false
    filesystem: read
```

---

## Deployment Questions

### How do I distribute my CLI?

**Distribution methods**:

**1. Direct download**:
```bash
curl -L https://releases.example.com/install.sh | sh
```

**2. Homebrew** (macOS/Linux):
```ruby
# Formula: mycli.rb
class Mycli < Formula
  desc "My CLI tool"
  homepage "https://example.com/cli"
  url "https://releases.example.com/mycli-v1.0.0.tar.gz"
  sha256 "abc123..."

  def install
    bin.install "mycli"
  end
end
```

**3. APT** (Debian/Ubuntu):
```bash
echo "deb [trusted=yes] https://apt.example.com stable main" | \
  sudo tee /etc/apt/sources.list.d/mycli.list
sudo apt update && sudo apt install mycli
```

**4. Scoop** (Windows):
```json
{
  "version": "1.0.0",
  "url": "https://releases.example.com/mycli-v1.0.0.zip",
  "bin": "mycli.exe"
}
```

---

### How do I set up auto-updates?

**1. Configure update server in embedded config**:
```yaml
updates:
  enabled: true
  update_url: https://releases.example.com/cli
  check_interval: 24h

  public_key: |
    -----BEGIN PUBLIC KEY-----
    MCowBQYDK2VwAyEA...
    -----END PUBLIC KEY-----
```

**2. User opts in to auto-install** (in `~/.config/mycli/config.yaml`):
```yaml
preferences:
  updates:
    auto_install: true  # User explicitly opts in
```

**3. Create release server**:
```
https://releases.example.com/
├── latest/
│   └── version.json
├── binaries/
│   ├── mycli-v1.2.3-darwin-amd64
│   ├── mycli-v1.2.3-linux-amd64
│   └── mycli-v1.2.3-windows-amd64.exe
└── checksums/
    └── v1.2.3-checksums.txt
```

**version.json**:
```json
{
  "version": "1.2.3",
  "released_at": "2025-11-23T12:00:00Z",
  "changelog": [
    {
      "type": "feature",
      "description": "Added new command"
    }
  ],
  "binaries": {
    "darwin-amd64": {
      "url": "https://releases.example.com/binaries/mycli-v1.2.3-darwin-amd64",
      "sha256": "abc123..."
    }
  }
}
```

---

### How do I sign binaries?

**1. Generate Ed25519 keypair**:
```bash
# Generate private key
openssl genpkey -algorithm ED25519 -out private.pem

# Extract public key
openssl pkey -in private.pem -pubout -out public.pem
```

**2. Sign binary**:
```bash
openssl dgst -sha256 -sign private.pem \
  -out mycli-v1.0.0.sig \
  mycli-v1.0.0
```

**3. Embed public key in config**:
```yaml
updates:
  public_key: |
    -----BEGIN PUBLIC KEY-----
    MCowBQYDK2VwAyEA...
    -----END PUBLIC KEY-----
```

**4. CLI verifies signature before updating**:
```
Checking for updates...
✓ New version available: 1.0.1
✓ Downloading binary...
✓ Verifying signature...
✓ Update complete
```

---

## Performance Questions

### What's the startup time?

**Target**: < 50ms (cold start without network)
**Maximum**: < 200ms (including update check)

**Breakdown**:
- Parse embedded config: ~5ms
- Check for updates (background): ~100ms
- Load cached OpenAPI spec: ~10ms
- Build command tree: ~20ms
- Execute command: varies (network-bound)

**Optimization tips**:
- Enable caching (`spec_ttl: 5m`)
- Use `--no-update` flag to skip update check
- Cache OpenAPI spec locally

---

### Does CliForge cache responses?

**Yes!** CliForge caches:

**1. OpenAPI spec**:
```yaml
behaviors:
  caching:
    spec_ttl: 5m  # Cache spec for 5 minutes
```

**2. GET responses**:
```yaml
behaviors:
  caching:
    response_ttl: 1m  # Cache GET responses for 1 minute
```

**Note**: POST/PUT/DELETE are never cached.

**Cache location**:
- **macOS/Linux**: `~/.cache/mycli/`
- **Windows**: `%LOCALAPPDATA%\mycli\cache\`

**Clear cache**:
```bash
mycli cache clear
```

---

### How much memory does CliForge use?

**Typical usage**:
- **Idle**: < 20MB
- **Active** (API calls): < 50MB
- **Peak** (large responses): < 100MB

**Memory breakdown**:
- Binary: ~10MB
- OpenAPI spec: ~1-5MB
- Command tree: ~5MB
- Response buffer: varies

---

## Security Questions

### How secure is CliForge?

**Security features**:

**1. Binary signature verification**:
- Ed25519 signatures
- Public key embedded in binary
- Verified before update

**2. TLS certificate pinning** (optional):
- Pin expected certificate fingerprints
- Prevent MITM attacks

**3. Secrets management**:
- OS-specific secure storage (Keychain, Credential Manager)
- Environment variable support
- File permissions (0600)

**4. Secret masking**:
- Automatic detection of secrets in output
- Masking in logs and terminal
- Pattern-based detection

**5. Audit logging**:
- Track all API calls
- Local and remote logging
- Request ID tracking

---

### Where are secrets stored?

**OAuth2 tokens**:
- **macOS**: Keychain
- **Linux**: Secret Service API (gnome-keyring, kwallet)
- **Windows**: Credential Manager
- **Fallback**: Encrypted file (`~/.mycli/credentials.json`, chmod 0600)

**API keys**:
- **Preferred**: Environment variables (`MYCLI_API_KEY`)
- **Alternative**: User config file (chmod 0600)
- **Not recommended**: Command-line flags (visible in process list)

**Passwords**:
- **Never stored**: Prompted on-demand
- **OAuth2**: Use tokens, not passwords

---

### Can users override security settings?

**No!** Security-critical settings are locked in embedded config:

**Locked settings**:
- `api.base_url` - API endpoint
- `api.openapi_url` - OpenAPI spec URL
- `updates.update_url` - Update server URL
- `updates.public_key` - Signature verification key
- `behaviors.auth.*` - Authentication configuration

**Exception**: Debug mode (`metadata.debug: true`) allows overrides, but:
- Shows security warning on every command
- Only for development/testing builds
- Production builds should have `debug: false`

---

### How do I report security issues?

**Security contact**:
- Email: security@cliforge.org (if project has dedicated email)
- GitHub: Create private security advisory
- Responsible disclosure: 90-day disclosure policy

**What to include**:
- Description of vulnerability
- Steps to reproduce
- Affected versions
- Suggested fix (if any)

---

## Licensing Questions

### What license does CliForge use?

CliForge itself is open-source (MIT License), but generated CLIs can use any license you choose.

**CliForge generator**:
- **License**: MIT
- **Usage**: Free for commercial and non-commercial use

**Generated CLIs**:
- **Your choice**: MIT, Apache-2.0, proprietary, etc.
- **Specify in config**:
  ```yaml
  metadata:
    license: MIT
  ```

---

### Can I use CliForge for commercial products?

**Yes!** CliForge is MIT licensed, allowing:
- Commercial use
- Distribution
- Modification
- Private use

**No restrictions on**:
- Charging for your CLI
- Using in proprietary products
- Distributing binaries

**Only requirement**:
- Include MIT license notice in documentation

---

### Do I need to attribute CliForge?

**Required**:
- Include CliForge license in your documentation
- Credit CliForge in README or docs (optional but appreciated)

**Optional**:
- Add "Built with CliForge" badge
- Link to CliForge website

**Not required**:
- Mention CliForge in CLI output
- Include CliForge in binary name

---

### Can I modify CliForge?

**Yes!** MIT license allows modifications:
- Fork the repository
- Make changes
- Distribute modified versions

**Contribution**:
- PRs welcome for improvements
- Follow contribution guidelines
- Sign CLA (if required)

---

## Next Steps

- **[Examples and Recipes](examples-and-recipes.md)** - Practical examples
- **[Troubleshooting](troubleshooting.md)** - Common issues
- **[Configuration DSL](configuration-dsl.md)** - Full configuration reference
- **[Technical Specification](technical-specification.md)** - Architecture details

---

**Version**: 0.9.0
**Last Updated**: 2025-11-23
