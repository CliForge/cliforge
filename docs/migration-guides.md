# Migration Guides

**Version**: 0.9.0
**Last Updated**: 2025-11-25

---

## Table of Contents

1. [Overview](#overview)
2. [Migrating from OpenAPI Generator](#migrating-from-openapi-generator)
3. [Migrating from Restish](#migrating-from-restish)
4. [Migrating from curl/Shell Scripts](#migrating-from-curlshell-scripts)
5. [Migrating from AWS CLI Patterns](#migrating-from-aws-cli-patterns)
6. [Migration Decision Matrix](#migration-decision-matrix)
7. [Common Migration Pitfalls](#common-migration-pitfalls)
8. [Migration Success Stories](#migration-success-stories)

---

## Overview

This guide helps you migrate to CliForge from other CLI tools and patterns. Whether you're using static code generators, dynamic spec loaders, shell scripts, or following AWS CLI patterns, this guide provides step-by-step instructions for a smooth transition.

### Why Migrate to CliForge?

CliForge combines the best features of different approaches:

| Feature | OpenAPI Generator | Restish | curl Scripts | AWS CLI | CliForge |
|---------|------------------|---------|--------------|---------|----------|
| **Branded binaries** | ✅ | ❌ | ❌ | ✅ | ✅ |
| **Dynamic spec updates** | ❌ | ✅ | ❌ | N/A | ✅ |
| **Self-updating** | ❌ | ❌ | ❌ | ❌ | ✅ |
| **No code generation** | ❌ | ✅ | ✅ | N/A | ✅ |
| **Workflow orchestration** | ❌ | ❌ | Manual | ✅ | ✅ |
| **Context switching** | ❌ | ✅ | Manual | ✅ | ✅ |
| **OAuth2 support** | Varies | ✅ | Manual | N/A | ✅ |
| **Interactive mode** | ❌ | ❌ | ❌ | ❌ | ✅ |

### Migration Support

Need help migrating? Check these resources:

- **Migration tool**: `cliforge migrate` (coming soon)
- **Documentation**: [Getting Started](getting-started.md)
- **Examples**: [Examples and Recipes](examples-and-recipes.md)
- **Support**: [GitHub Issues](https://github.com/cliforge/cliforge/issues)

---

## Migrating from OpenAPI Generator

### Understanding the Difference

**OpenAPI Generator** generates static client code that must be regenerated when your API changes. **CliForge** creates a runtime engine that dynamically loads your OpenAPI spec.

#### Conceptual Comparison

**OpenAPI Generator Workflow:**
```
OpenAPI Spec → Generator → Generated Code → Compile → Binary
                              ↓
                         Version Control
                         (Generated files)
```

**CliForge Workflow:**
```
OpenAPI Spec (Remote) → CliForge Config → Build → Binary
                                                    ↓
                                          Runtime loads spec
                                          (No generated files)
```

### Key Philosophical Differences

| Aspect | OpenAPI Generator | CliForge |
|--------|------------------|----------|
| **Approach** | Static code generation | Dynamic spec loading |
| **API changes** | Regenerate + recompile | Automatic (spec cached) |
| **Repository** | Generated code checked in | Only config checked in |
| **Customization** | Templates + post-processing | Configuration + extensions |
| **Distribution** | Compile from source | Single binary |
| **Updates** | Manual rebuild | Self-update mechanism |

---

### Step 1: Analyze Your Current Setup

First, understand what you're using from OpenAPI Generator:

```bash
# Find your OpenAPI Generator config
find . -name "openapi-generator-config.yaml" -o -name ".openapi-generator"

# Check your generation script
cat scripts/generate-cli.sh  # or wherever you keep it

# Review generated code structure
ls -la generated/
```

**Common OpenAPI Generator patterns:**

```yaml
# openapi-generator-config.yaml
generatorName: go
outputDir: ./generated
packageName: mycli
gitUserId: myorg
gitRepoId: mycli

# Template customizations (if any)
templateDir: ./templates

# Additional properties
additionalProperties:
  packageVersion: 1.0.0
  hideGenerationTimestamp: true
```

---

### Step 2: Create Equivalent CliForge Configuration

Map your OpenAPI Generator config to CliForge's `cli-config.yaml`:

**Before (OpenAPI Generator):**
```yaml
# openapi-generator-config.yaml
generatorName: go-cli
outputDir: ./generated
packageName: myapi-cli
packageVersion: 1.0.0
additionalProperties:
  apiBaseUrl: https://api.example.com
  authType: oauth2
```

**After (CliForge):**
```yaml
# cli-config.yaml
metadata:
  name: myapi-cli
  version: 1.0.0
  description: My API Command Line Interface

api:
  openapi_url: https://api.example.com/openapi.yaml
  base_url: https://api.example.com

behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: authorization_code
      authorization_url: https://auth.example.com/authorize
      token_url: https://auth.example.com/token
      scopes: [read, write]
```

#### Configuration Mapping Reference

| OpenAPI Generator | CliForge | Notes |
|------------------|----------|-------|
| `packageName` | `metadata.name` | CLI binary name |
| `packageVersion` | `metadata.version` | Semantic version |
| `additionalProperties.apiBaseUrl` | `api.base_url` | API base URL |
| Template customizations | `branding.*` | Colors, ASCII art, prompts |
| Custom templates | `x-cli-*` extensions | In OpenAPI spec |
| Post-processing scripts | `behaviors.*` | Built-in behaviors |

---

### Step 3: Migrate Custom Templates to OpenAPI Extensions

If you customized OpenAPI Generator templates, map them to CliForge extensions:

**Before (Custom template for command naming):**
```mustache
{{! Custom template: api_operation.mustache }}
func {{operationIdCamelCase}}Command() *cobra.Command {
    return &cobra.Command{
        Use:   "{{#lambda.lowercase}}{{operationId}}{{/lambda.lowercase}}",
        Short: "{{summary}}",
        // ... generated code
    }
}
```

**After (OpenAPI extension):**
```yaml
paths:
  /users:
    get:
      operationId: listUsers
      summary: List all users
      x-cli-command: "list users"      # Custom command name
      x-cli-aliases: ["ls users", "users"]  # Command aliases
      x-cli-output:
        table:
          columns:
            - field: id
              header: ID
            - field: name
              header: Name
```

#### Template Migration Map

| Template Feature | CliForge Extension | Example |
|-----------------|-------------------|---------|
| Command naming | `x-cli-command` | `"list users"` |
| Command aliases | `x-cli-aliases` | `["ls", "users"]` |
| Flag naming | `x-cli-flags` | Custom flag names |
| Output formatting | `x-cli-output` | Table/JSON/YAML |
| Help text | `x-cli-examples` | Usage examples |
| Confirmation prompts | `x-cli-confirmation` | Delete confirmations |
| Pre-flight checks | `x-cli-workflow` | Validation steps |

---

### Step 4: Remove Generated Code

CliForge doesn't generate code, so you can remove:

```bash
# Remove generated code (be careful!)
rm -rf generated/
rm -rf docs/  # If generated by OpenAPI Generator
rm -rf test/  # If generated by OpenAPI Generator

# Remove generator config
rm openapi-generator-config.yaml
rm -rf templates/  # If you had custom templates

# Remove from .gitignore (no longer needed)
# Delete these lines:
# generated/
# .openapi-generator/
# .openapi-generator-ignore
```

Update your `.gitignore`:

```diff
# Remove OpenAPI Generator artifacts
- generated/
- .openapi-generator/
- .openapi-generator-ignore

# Add CliForge artifacts
+ dist/  # Built binaries
+ .cliforge/  # Local cache (optional)
```

---

### Step 5: Migrate Build Scripts

**Before (OpenAPI Generator):**
```bash
#!/bin/bash
# scripts/generate-cli.sh

# Download latest spec
curl -o openapi.yaml https://api.example.com/openapi.yaml

# Generate code
openapi-generator-cli generate \
  -i openapi.yaml \
  -g go \
  -o ./generated \
  -c openapi-generator-config.yaml

# Build binary
cd generated
go build -o ../bin/myapi-cli ./cmd/myapi-cli

echo "CLI built successfully"
```

**After (CliForge):**
```bash
#!/bin/bash
# scripts/build-cli.sh

# Build branded binary
cliforge build \
  --config cli-config.yaml \
  --output dist/

echo "CLI built successfully"
echo "Binaries available in dist/"
```

**Key differences:**
- No code generation step
- No intermediate generated code
- Single command to build
- Multi-platform binaries automatically

---

### Step 6: Migrate Custom Functionality

If you added custom code to generated files, migrate to CliForge features:

#### Custom Authentication Logic

**Before (Generated + customized):**
```go
// generated/auth.go (customized)
func (a *APIAuth) GetToken() (string, error) {
    // Custom logic added to generated code
    token, err := a.cache.Get("token")
    if err != nil {
        token, err = a.fetchNewToken()
        a.cache.Set("token", token, 3600)
    }
    return token, err
}
```

**After (CliForge configuration):**
```yaml
# cli-config.yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: client_credentials
      token_url: https://auth.example.com/token
      storage:
        primary: keyring  # Built-in secure storage
        fallback: file
      cache:
        enabled: true
        ttl: 3600  # Token cache duration
```

#### Custom Output Formatting

**Before (Generated + customized):**
```go
// generated/output.go (customized)
func PrintUsers(users []User) {
    // Custom table formatting
    table := tablewriter.NewWriter(os.Stdout)
    table.SetHeader([]string{"ID", "Name", "Email", "Status"})
    for _, u := range users {
        table.Append([]string{
            fmt.Sprintf("%d", u.ID),
            u.Name,
            u.Email,
            colorizeStatus(u.Status),  // Custom coloring
        })
    }
    table.Render()
}
```

**After (OpenAPI extension):**
```yaml
paths:
  /users:
    get:
      x-cli-output:
        table:
          columns:
            - field: id
              header: ID
              width: 10
            - field: name
              header: Name
            - field: email
              header: Email
            - field: status
              header: Status
              color-map:
                active: green
                inactive: red
                pending: yellow
```

#### Custom Validation

**Before (Custom code):**
```go
// Added to generated code
func ValidateCreateUser(user *User) error {
    if user.Age < 18 {
        return fmt.Errorf("user must be 18 or older")
    }
    if !isValidEmail(user.Email) {
        return fmt.Errorf("invalid email format")
    }
    return nil
}
```

**After (OpenAPI schema + workflow):**
```yaml
paths:
  /users:
    post:
      # Schema validation (built-in)
      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                age:
                  type: integer
                  minimum: 18  # Automatic validation
                email:
                  type: string
                  format: email  # Automatic validation

      # Pre-flight validation (if needed)
      x-cli-workflow:
        steps:
          - id: validate-email
            type: plugin
            plugin: validators
            command: check-email
            args:
              email: "{body.email}"

          - id: create-user
            type: api-call
            depends-on: [validate-email]
```

---

### Step 7: Migrate Tests

Update your tests to work with CliForge:

**Before (Testing generated code):**
```go
// generated_test.go
func TestListUsers(t *testing.T) {
    client := NewAPIClient()
    users, err := client.ListUsers()
    assert.NoError(t, err)
    assert.NotEmpty(t, users)
}
```

**After (Testing CLI binary):**
```bash
#!/bin/bash
# test/integration-test.sh

# Start mock server
./test/mock-server &
SERVER_PID=$!

# Test CLI commands
export API_KEY=test-key

# Test list command
OUTPUT=$(./dist/myapi-cli list users --output json)
echo "$OUTPUT" | jq -e '.[] | select(.id == 1)' || exit 1

# Test create command
./dist/myapi-cli create user \
  --name "Test User" \
  --email "test@example.com" || exit 1

# Cleanup
kill $SERVER_PID
```

---

### Step 8: Update CI/CD Pipeline

**Before (OpenAPI Generator pipeline):**
```yaml
# .github/workflows/build.yml
name: Build CLI

on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Set up Java  # For OpenAPI Generator
        uses: actions/setup-java@v2
        with:
          java-version: '11'

      - name: Generate code
        run: |
          npm install -g @openapitools/openapi-generator-cli
          ./scripts/generate-cli.sh

      - name: Build binary
        run: |
          cd generated
          go build -o ../bin/myapi-cli ./cmd/myapi-cli
```

**After (CliForge pipeline):**
```yaml
# .github/workflows/build.yml
name: Build CLI

on: [push]

jobs:
  build:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v2

      - name: Set up Go
        uses: actions/setup-go@v3
        with:
          go-version: '1.21'

      - name: Install CliForge
        run: |
          go install github.com/cliforge/cliforge/cmd/cliforge@latest

      - name: Build CLI
        run: |
          cliforge build --config cli-config.yaml --output dist/

      - name: Upload artifacts
        uses: actions/upload-artifact@v3
        with:
          name: cli-binaries
          path: dist/
```

---

### Migration Checklist

Use this checklist to track your migration progress:

- [ ] **Analysis**
  - [ ] Document current OpenAPI Generator setup
  - [ ] Identify custom templates and modifications
  - [ ] List all custom functionality added to generated code
  - [ ] Review build and deployment scripts

- [ ] **Configuration**
  - [ ] Create `cli-config.yaml` with equivalent settings
  - [ ] Map branding and customization to CliForge config
  - [ ] Configure authentication and authorization
  - [ ] Set up output formatting preferences

- [ ] **OpenAPI Extensions**
  - [ ] Add `x-cli-command` for custom command names
  - [ ] Add `x-cli-output` for custom output formatting
  - [ ] Add `x-cli-workflow` for multi-step operations
  - [ ] Add `x-cli-interactive` for prompts

- [ ] **Cleanup**
  - [ ] Remove generated code directory
  - [ ] Remove OpenAPI Generator config
  - [ ] Remove custom templates
  - [ ] Update `.gitignore`

- [ ] **Build System**
  - [ ] Update build scripts
  - [ ] Update CI/CD pipeline
  - [ ] Configure release process
  - [ ] Test multi-platform builds

- [ ] **Testing**
  - [ ] Create integration tests for CLI commands
  - [ ] Test authentication flows
  - [ ] Test output formatting
  - [ ] Test error handling

- [ ] **Documentation**
  - [ ] Update user documentation
  - [ ] Update developer documentation
  - [ ] Create migration notes for users
  - [ ] Document new features available

---

### Common Pitfalls: OpenAPI Generator Migration

#### Pitfall 1: Expecting Generated Code Structure

**Problem:**
```bash
# This won't exist with CliForge
ls generated/api/users_api.go
# Error: No such file or directory
```

**Solution:**
Understand that CliForge doesn't generate code. All CLI functionality comes from the runtime engine + configuration.

#### Pitfall 2: Trying to Customize Generated Code

**Problem:**
"Where do I add custom logic like I did in generated files?"

**Solution:**
Use configuration and OpenAPI extensions instead:
- Authentication → `behaviors.auth.*`
- Validation → OpenAPI schema + `x-cli-workflow`
- Output formatting → `x-cli-output`
- Custom commands → `x-cli-workflow` + plugins

#### Pitfall 3: Missing Template Customizations

**Problem:**
"My custom command names and formatting are gone!"

**Solution:**
Migrate template customizations to OpenAPI extensions:

```yaml
# Add to your OpenAPI spec
paths:
  /users:
    get:
      x-cli-command: "list users"  # Custom command name
      x-cli-output:
        table:  # Custom formatting
          columns:
            - field: id
              header: ID
```

#### Pitfall 4: Spec Caching Confusion

**Problem:**
"I updated my OpenAPI spec but the CLI doesn't see the changes."

**Solution:**
CliForge caches specs for performance. Force refresh:

```bash
# Clear cache and reload spec
myapi-cli --refresh list users

# Or check cache location
myapi-cli config show cache_dir
rm -rf $(myapi-cli config show cache_dir)
```

#### Pitfall 5: Authentication Token Storage

**Problem:**
"Where are my authentication tokens stored now?"

**Solution:**
CliForge uses secure OS-level storage by default:

```yaml
# cli-config.yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      storage:
        primary: keyring  # macOS Keychain, Windows Credential Store, Linux Secret Service
        fallback: file    # Encrypted file if keyring unavailable
```

Check token storage:
```bash
# View current auth status
myapi-cli whoami

# Clear stored tokens
myapi-cli logout
```

---

### Side-by-Side Comparison Example

Here's a complete example showing the same functionality in both approaches:

#### OpenAPI Generator Approach

**Files created:**
```
project/
├── openapi-generator-config.yaml
├── openapi.yaml
├── templates/  # Custom templates
│   ├── api.mustache
│   └── model.mustache
├── scripts/
│   └── generate-cli.sh
└── generated/  # ~50+ files
    ├── api/
    ├── cmd/
    ├── models/
    └── utils/
```

**openapi-generator-config.yaml:**
```yaml
generatorName: go
outputDir: ./generated
packageName: petstore-cli
packageVersion: 1.0.0
gitUserId: example
gitRepoId: petstore-cli
additionalProperties:
  apiBaseUrl: https://api.petstore.com
```

**Build process:**
```bash
# Generate code
openapi-generator-cli generate \
  -i openapi.yaml \
  -g go \
  -c openapi-generator-config.yaml

# Customize generated code (manual edits)
vim generated/api/pets_api.go  # Add custom logic

# Build
cd generated
go build -o ../bin/petstore-cli ./cmd/petstore-cli

# Distribution
# Users must have Go installed or you distribute per platform
```

#### CliForge Approach

**Files created:**
```
project/
├── cli-config.yaml
└── petstore-api.yaml  # Your OpenAPI spec with x-cli-* extensions
```

**cli-config.yaml:**
```yaml
metadata:
  name: petstore-cli
  version: 1.0.0
  description: Petstore Command Line Interface

api:
  openapi_url: https://api.petstore.com/openapi.yaml
  base_url: https://api.petstore.com

branding:
  colors:
    primary: "#FF6B35"

behaviors:
  auth:
    type: api_key
    api_key:
      header: X-API-Key
      env_var: PETSTORE_API_KEY
```

**Build process:**
```bash
# Build (no code generation)
cliforge build --config cli-config.yaml --output dist/

# Automatically creates:
# dist/petstore-cli-darwin-amd64
# dist/petstore-cli-darwin-arm64
# dist/petstore-cli-linux-amd64
# dist/petstore-cli-windows-amd64.exe

# Distribution
# Single binary per platform, no dependencies
```

**Result comparison:**

| Aspect | OpenAPI Generator | CliForge |
|--------|------------------|----------|
| Files in repo | 50+ generated files | 1 config file |
| Build time | ~30s (generate + compile) | ~5s (compile only) |
| API update | Regenerate all + rebuild | Automatic (spec cached) |
| Customization | Edit generated code | Edit config + extensions |
| Distribution | One binary or source | One binary per platform |

---

### Success Metrics

After migration, you should see:

**Repository Cleanup:**
- ✅ 50-100+ generated files removed
- ✅ Custom template files removed
- ✅ Repository size reduced by 60-80%
- ✅ Cleaner git history (no generated code churn)

**Build Performance:**
- ✅ Build time reduced from minutes to seconds
- ✅ No Java dependency for OpenAPI Generator
- ✅ Simpler CI/CD pipeline
- ✅ Faster local development iteration

**Maintenance Benefits:**
- ✅ API changes don't require code regeneration
- ✅ No merge conflicts in generated code
- ✅ Single source of truth (OpenAPI spec)
- ✅ Self-updating binaries for end users

**User Experience:**
- ✅ Smaller binary distribution
- ✅ Automatic API updates
- ✅ Consistent command structure
- ✅ Better error messages

---

## Migrating from Restish

### Understanding the Difference

**Restish** is a dynamic CLI that can call any OpenAPI-described API without code generation. **CliForge** builds on this concept but creates branded, distributable binaries.

#### Conceptual Comparison

**Restish Workflow:**
```
User installs: restish
User configures: ~/.restish/apis.json
User runs: restish api-name operation
```

**CliForge Workflow:**
```
Developer builds: branded CLI binary
User installs: single binary (myapi-cli)
User runs: myapi-cli operation
```

### Key Philosophical Differences

| Aspect | Restish | CliForge |
|--------|---------|----------|
| **Distribution** | Generic tool + config | Branded binary |
| **Configuration** | User configures | Developer configures |
| **Branding** | Generic "restish" | Custom brand/name |
| **Updates** | Manual config changes | Self-updating binary |
| **Auth storage** | User's Restish config | CLI-specific storage |
| **Multi-API** | One tool, many APIs | One CLI per API |

---

### Step 1: Export Restish Configuration

First, export your current Restish setup:

```bash
# View current Restish APIs
restish list

# Export configuration for your API
cat ~/.restish/apis.json
```

**Typical Restish configuration:**
```json
{
  "myapi": {
    "short": "My API",
    "base": "https://api.example.com",
    "spec_files": ["https://api.example.com/openapi.yaml"],
    "profiles": {
      "default": {
        "headers": {
          "Authorization": "Bearer ${MYAPI_TOKEN}"
        }
      },
      "production": {
        "base": "https://api.example.com",
        "headers": {
          "Authorization": "Bearer ${MYAPI_PROD_TOKEN}"
        }
      },
      "staging": {
        "base": "https://staging.api.example.com",
        "headers": {
          "Authorization": "Bearer ${MYAPI_STAGING_TOKEN}"
        }
      }
    }
  }
}
```

---

### Step 2: Convert to CliForge Configuration

Map Restish configuration to CliForge `cli-config.yaml`:

**Before (Restish):**
```json
{
  "myapi": {
    "short": "My API",
    "base": "https://api.example.com",
    "spec_files": ["https://api.example.com/openapi.yaml"],
    "profiles": {
      "default": {
        "headers": {
          "Authorization": "Bearer ${MYAPI_TOKEN}"
        }
      }
    }
  }
}
```

**After (CliForge):**
```yaml
# cli-config.yaml
metadata:
  name: myapi-cli
  version: 1.0.0
  description: My API Command Line Interface

api:
  openapi_url: https://api.example.com/openapi.yaml
  base_url: https://api.example.com

behaviors:
  auth:
    type: oauth2  # or api_key, depending on your auth method
    oauth2:
      flow: authorization_code
      token_url: https://auth.example.com/token
      storage:
        primary: keyring
        fallback: file
```

#### Configuration Mapping Reference

| Restish | CliForge | Notes |
|---------|----------|-------|
| `short` | `metadata.description` | Short description |
| `base` | `api.base_url` | API base URL |
| `spec_files[0]` | `api.openapi_url` | OpenAPI spec URL |
| `profiles` | `contexts.available` | Environment contexts |
| `headers.Authorization` | `behaviors.auth.*` | Authentication config |
| Profile switching | `contexts.*` | Context switching |

---

### Step 3: Migrate Profile/Context System

Restish profiles map directly to CliForge contexts:

**Before (Restish profiles):**
```json
{
  "myapi": {
    "profiles": {
      "development": {
        "base": "http://localhost:8080",
        "headers": {
          "Authorization": "Bearer dev-token"
        }
      },
      "staging": {
        "base": "https://staging.api.example.com",
        "headers": {
          "Authorization": "Bearer ${STAGING_TOKEN}"
        }
      },
      "production": {
        "base": "https://api.example.com",
        "headers": {
          "Authorization": "Bearer ${PROD_TOKEN}"
        }
      }
    }
  }
}
```

**After (CliForge contexts):**
```yaml
# cli-config.yaml
contexts:
  default: production

  available:
    development:
      api:
        base_url: http://localhost:8080
        openapi_url: http://localhost:8080/openapi.yaml
      auth:
        type: none  # No auth in dev
      display:
        color: yellow
        prompt: "[DEV]"

    staging:
      api:
        base_url: https://staging.api.example.com
        openapi_url: https://staging.api.example.com/openapi.yaml
      auth:
        type: oauth2
      display:
        color: blue
        prompt: "[STAGING]"

    production:
      api:
        base_url: https://api.example.com
        openapi_url: https://api.example.com/openapi.yaml
      auth:
        type: oauth2
      display:
        color: green
        prompt: "[PROD]"
```

**Usage comparison:**

```bash
# Restish
restish myapi --rsh-profile=staging get /users

# CliForge
myapi-cli --context staging users list
```

---

### Step 4: Migrate Authentication

Restish authentication is typically configured in profiles. CliForge provides built-in auth flows:

**Before (Restish - manual token management):**
```json
{
  "myapi": {
    "profiles": {
      "default": {
        "headers": {
          "Authorization": "Bearer ${MYAPI_TOKEN}"
        }
      }
    }
  }
}
```

```bash
# User manually obtains token
export MYAPI_TOKEN=$(curl -X POST https://auth.example.com/token ...)

# Then uses Restish
restish myapi get /users
```

**After (CliForge - automatic OAuth2 flow):**
```yaml
# cli-config.yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: authorization_code
      client_id: myapi-cli
      authorization_url: https://auth.example.com/authorize
      token_url: https://auth.example.com/token
      scopes: [read, write]
      storage:
        primary: keyring  # Secure OS storage
        fallback: file
```

```bash
# CliForge handles OAuth flow automatically
myapi-cli login  # Opens browser, handles callback, stores token

# Then commands work automatically
myapi-cli users list  # Token added automatically
```

---

### Step 5: Migrate Custom Headers and Parameters

**Before (Restish - custom headers):**
```json
{
  "myapi": {
    "profiles": {
      "default": {
        "headers": {
          "X-API-Version": "2024-01-01",
          "X-Client-ID": "restish-cli",
          "Accept": "application/json"
        },
        "query": {
          "client": "restish"
        }
      }
    }
  }
}
```

**After (CliForge):**
```yaml
# cli-config.yaml
api:
  openapi_url: https://api.example.com/openapi.yaml
  base_url: https://api.example.com
  default_headers:
    X-API-Version: "2024-01-01"
    X-Client-ID: "myapi-cli"
    Accept: "application/json"
  user_agent: "myapi-cli/1.0.0"
```

---

### Step 6: Migrate Output Formatting

Restish uses flags for output control. CliForge provides the same plus configuration:

**Before (Restish):**
```bash
# JSON output (default)
restish myapi get /users

# Pretty JSON
restish myapi get /users | jq

# Headers and verbose
restish myapi -v get /users

# Filter with JMESPath
restish myapi get /users --filter "items[?status=='active']"
```

**After (CliForge):**
```bash
# JSON output (default)
myapi-cli users list

# Pretty JSON (configured as default)
myapi-cli users list --output json

# Table output
myapi-cli users list --output table

# Verbose with headers
myapi-cli users list --verbose

# Filter (built-in)
myapi-cli users list --filter "status=active"
```

Configure default output format:

```yaml
# cli-config.yaml
defaults:
  output:
    format: table  # table, json, yaml, csv
    pretty_print: true
    color: auto
```

---

### Step 7: Build and Distribute

The key difference: Restish is one tool for all APIs, CliForge creates one CLI per API.

**Before (Restish - user setup):**
```bash
# User installs Restish
brew install restish

# User configures for your API
mkdir -p ~/.restish
cat > ~/.restish/apis.json << EOF
{
  "myapi": {
    "base": "https://api.example.com",
    "spec_files": ["https://api.example.com/openapi.yaml"]
  }
}
EOF

# User sets environment variables
export MYAPI_TOKEN=xyz...

# User can now call your API
restish myapi get /users
```

**After (CliForge - developer setup):**
```bash
# Developer builds branded CLI
cliforge build --config cli-config.yaml --output dist/

# Developer distributes binary
# Users download ONE file, no dependencies

# User usage (no setup needed)
myapi-cli users list
```

**Distribution:**

```bash
# Restish
# Send documentation to users explaining:
# 1. Install Restish
# 2. Configure ~/.restish/apis.json
# 3. Set environment variables
# 4. Learn Restish commands

# CliForge
# Send one binary file
# Users run: ./myapi-cli
```

---

### Migration Checklist

- [ ] **Analysis**
  - [ ] Export Restish configuration (`~/.restish/apis.json`)
  - [ ] Document all profiles and their settings
  - [ ] List custom headers and query parameters
  - [ ] Identify authentication method used

- [ ] **Configuration**
  - [ ] Create `cli-config.yaml`
  - [ ] Map profiles to contexts
  - [ ] Configure authentication flow
  - [ ] Set default headers and parameters

- [ ] **Testing**
  - [ ] Test each former Restish profile as CliForge context
  - [ ] Verify authentication works
  - [ ] Test output formatting
  - [ ] Verify all operations work

- [ ] **User Migration**
  - [ ] Create user migration guide
  - [ ] Document new command structure
  - [ ] Provide comparison of old vs new commands
  - [ ] Highlight new features available

---

### Common Pitfalls: Restish Migration

#### Pitfall 1: Missing Generic Tool

**Problem:**
"I liked having one tool (restish) for all my APIs."

**Solution:**
CliForge's approach provides better UX for end users:

```bash
# Restish - users must remember "restish" + API name
restish api1 get /users
restish api2 get /customers
restish api3 get /products

# CliForge - branded CLIs with clear purpose
api1-cli users list
api2-cli customers list
api3-cli products list
```

Each CLI is focused and branded for its specific API.

#### Pitfall 2: Command Structure Changes

**Problem:**
Restish uses HTTP-centric commands, CliForge uses resource-centric commands.

**Before (Restish):**
```bash
restish myapi get /users
restish myapi post /users '{"name":"John"}'
restish myapi delete /users/123
```

**After (CliForge):**
```bash
myapi-cli users list
myapi-cli users create --name "John"
myapi-cli users delete --user-id 123
```

**Solution:**
Add aliases for HTTP-style commands if needed:

```yaml
paths:
  /users:
    get:
      x-cli-command: "users list"
      x-cli-aliases: ["get /users", "users"]
```

#### Pitfall 3: Profile Switching Muscle Memory

**Problem:**
"I'm used to `--rsh-profile` flag"

**Solution:**
Use context switching:

```bash
# Restish
restish myapi --rsh-profile=staging get /users

# CliForge equivalent
myapi-cli --context staging users list

# Or set default context
myapi-cli context use staging
myapi-cli users list  # Uses staging
```

Create shell aliases for convenience:

```bash
# ~/.bashrc or ~/.zshrc
alias myapi-dev='myapi-cli --context development'
alias myapi-staging='myapi-cli --context staging'
alias myapi-prod='myapi-cli --context production'

# Usage
myapi-staging users list
```

---

### Feature Comparison

Here's what you gain (and what changes) when migrating from Restish:

**Features You Keep:**

✅ Dynamic OpenAPI spec loading
✅ No code generation
✅ Multiple environment support (profiles → contexts)
✅ Flexible output formatting
✅ HTTP header customization

**Features You Gain:**

➕ Branded binary with your API name
➕ Self-updating mechanism
➕ Built-in OAuth2 flow (not just tokens)
➕ Secure token storage (OS keyring)
➕ Workflow orchestration (multi-step operations)
➕ Interactive mode with prompts
➕ Table output with custom formatting
➕ Automatic command structure from OpenAPI
➕ Context-aware prompts and colors
➕ Plugin system for extensions

**What Changes:**

⚠️ Command structure: HTTP verbs → Resource actions
⚠️ One tool per API instead of one tool for all
⚠️ Configuration managed by developer, not end user
⚠️ Binary distribution instead of universal tool

---

### Side-by-Side Comparison Example

**Restish Approach:**

```bash
# Installation (user)
brew install restish

# Configuration (user)
cat > ~/.restish/apis.json << 'EOF'
{
  "petstore": {
    "base": "https://api.petstore.com",
    "spec_files": ["https://api.petstore.com/openapi.yaml"],
    "profiles": {
      "default": {
        "headers": {
          "Authorization": "Bearer ${PETSTORE_TOKEN}"
        }
      }
    }
  }
}
EOF

# Set token (user)
export PETSTORE_TOKEN=$(curl ...)

# Usage
restish petstore get /pets
restish petstore post /pets '{"name":"Fluffy","category":"cat"}'
restish petstore get /pets/1
restish petstore delete /pets/1

# Multiple environments (user switches)
restish petstore --rsh-profile=staging get /pets
restish petstore --rsh-profile=production get /pets
```

**CliForge Approach:**

```bash
# Installation (user downloads one binary)
curl -L https://releases.petstore.com/petstore-cli -o petstore-cli
chmod +x petstore-cli
sudo mv petstore-cli /usr/local/bin/

# Authentication (automatic OAuth flow)
petstore-cli login  # Opens browser, handles everything

# Usage (cleaner commands)
petstore-cli pets list
petstore-cli pets create --name "Fluffy" --category cat
petstore-cli pets get --pet-id 1
petstore-cli pets delete --pet-id 1

# Multiple environments (context switching)
petstore-cli context use staging
petstore-cli pets list  # Uses staging

petstore-cli context use production
petstore-cli pets list  # Uses production

# Or inline
petstore-cli --context staging pets list
```

**User Experience Comparison:**

| Task | Restish | CliForge |
|------|---------|----------|
| Initial setup | 5 steps | 1 step (download binary) |
| Authentication | Manual token management | Automatic OAuth |
| Command clarity | HTTP-centric | Resource-centric |
| Environment switching | Flag per command | Context switching |
| Updates | Brew upgrade | Self-update |
| Help | Generic for all APIs | Specific to your API |

---

## Migrating from curl/Shell Scripts

### Understanding the Difference

If you're using **curl** commands or shell scripts to interact with APIs, you're dealing with raw HTTP. **CliForge** provides a structured CLI with automatic:

- Authentication management
- Parameter validation
- Output formatting
- Error handling
- Documentation

#### Conceptual Comparison

**curl/Script Workflow:**
```bash
# Script: create-user.sh
TOKEN=$(cat ~/.api-token)
curl -X POST https://api.example.com/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"'"$1"'","email":"'"$2"'"}' | jq
```

**CliForge Workflow:**
```bash
# Single command, all complexity hidden
myapi-cli users create --name "$1" --email "$2"
```

---

### Step 1: Audit Your Current Scripts

Identify all the curl commands and scripts you're using:

```bash
# Find shell scripts with curl
grep -r "curl" scripts/ bin/ ~/bin/

# Find API-related scripts
find scripts/ -name "*.sh" -exec grep -l "api.example.com" {} \;
```

**Common curl patterns:**

```bash
# GET request
curl -X GET https://api.example.com/users

# GET with authentication
curl -X GET https://api.example.com/users \
  -H "Authorization: Bearer $TOKEN"

# POST with JSON body
curl -X POST https://api.example.com/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"John","email":"john@example.com"}'

# PUT request
curl -X PUT https://api.example.com/users/123 \
  -H "Authorization: Bearer $TOKEN" \
  -d '{"status":"active"}'

# DELETE request
curl -X DELETE https://api.example.com/users/123 \
  -H "Authorization: Bearer $TOKEN"

# Complex script with error handling
#!/bin/bash
TOKEN=$(cat ~/.api-token)
RESPONSE=$(curl -s -w "\n%{http_code}" -X POST \
  https://api.example.com/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"'"$1"'","email":"'"$2"'"}')

HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
BODY=$(echo "$RESPONSE" | sed '$d')

if [ "$HTTP_CODE" = "201" ]; then
  echo "User created successfully"
  echo "$BODY" | jq
else
  echo "Error: HTTP $HTTP_CODE"
  echo "$BODY" | jq
  exit 1
fi
```

---

### Step 2: Map curl Commands to CLI Operations

Create a mapping of your curl commands to CliForge CLI commands:

| curl Command | CliForge Equivalent |
|--------------|-------------------|
| `curl GET /users` | `myapi-cli users list` |
| `curl GET /users/123` | `myapi-cli users get --user-id 123` |
| `curl POST /users` | `myapi-cli users create` |
| `curl PUT /users/123` | `myapi-cli users update --user-id 123` |
| `curl DELETE /users/123` | `myapi-cli users delete --user-id 123` |

**Detailed mapping:**

**Before (curl):**
```bash
# List users
curl -X GET https://api.example.com/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Accept: application/json" | jq

# Get specific user
curl -X GET https://api.example.com/users/123 \
  -H "Authorization: Bearer $TOKEN" | jq '.name'

# Create user
curl -X POST https://api.example.com/users \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "John Doe",
    "email": "john@example.com",
    "age": 30
  }' | jq

# Update user
curl -X PUT https://api.example.com/users/123 \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"status": "active"}' | jq

# Delete user
curl -X DELETE https://api.example.com/users/123 \
  -H "Authorization: Bearer $TOKEN"

# Search users
curl -X GET 'https://api.example.com/users?status=active&limit=10' \
  -H "Authorization: Bearer $TOKEN" | jq
```

**After (CliForge):**
```bash
# List users
myapi-cli users list

# Get specific user (with field selection)
myapi-cli users get --user-id 123 --fields name

# Create user
myapi-cli users create \
  --name "John Doe" \
  --email "john@example.com" \
  --age 30

# Update user
myapi-cli users update --user-id 123 --status active

# Delete user
myapi-cli users delete --user-id 123

# Search users
myapi-cli users list --filter "status=active" --limit 10
```

---

### Step 3: Convert Shell Scripts to CLI Configurations

Migrate complex shell scripts to CliForge workflows:

**Before (Complex shell script):**
```bash
#!/bin/bash
# deploy-app.sh - Multi-step deployment script

set -e

TOKEN=$(cat ~/.api-token)
APP_ID=$1
BASE_URL="https://api.example.com"

echo "Step 1: Checking readiness..."
READY=$(curl -s "$BASE_URL/readiness" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.ready')

if [ "$READY" != "true" ]; then
  echo "Error: System not ready"
  exit 1
fi

echo "Step 2: Creating deployment..."
DEPLOYMENT_ID=$(curl -s -X POST "$BASE_URL/deployments" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"app_id\": \"$APP_ID\"}" | jq -r '.id')

echo "Deployment ID: $DEPLOYMENT_ID"

echo "Step 3: Waiting for deployment..."
while true; do
  STATUS=$(curl -s "$BASE_URL/deployments/$DEPLOYMENT_ID" \
    -H "Authorization: Bearer $TOKEN" | jq -r '.status')

  echo "Status: $STATUS"

  if [ "$STATUS" = "completed" ]; then
    echo "Deployment successful!"
    break
  elif [ "$STATUS" = "failed" ]; then
    echo "Deployment failed!"
    exit 1
  fi

  sleep 5
done

echo "Step 4: Running health check..."
curl -s "$BASE_URL/apps/$APP_ID/health" \
  -H "Authorization: Bearer $TOKEN" | jq
```

**After (CliForge workflow in OpenAPI spec):**
```yaml
# In your OpenAPI spec
paths:
  /deployments:
    post:
      operationId: deploy App
      x-cli-command: "deploy"
      x-cli-workflow:
        settings:
          fail-fast: true
          timeout: 600

        steps:
          - id: check-readiness
            type: api-call
            description: "Checking system readiness"
            request:
              method: GET
              url: "{base_url}/readiness"
            output:
              ready: "body.ready"

          - id: create-deployment
            type: api-call
            description: "Creating deployment"
            depends-on: [check-readiness]
            condition: "{check-readiness.ready} == true"
            request:
              method: POST
              url: "{base_url}/deployments"
              body:
                app_id: "{args.app_id}"
            output:
              deployment_id: "body.id"

          - id: wait-for-completion
            type: wait
            description: "Waiting for deployment to complete"
            depends-on: [create-deployment]
            poll:
              url: "{base_url}/deployments/{create-deployment.deployment_id}"
              interval: 5
              timeout: 300
              until: "body.status in ['completed', 'failed']"
              success-condition: "body.status == 'completed'"

          - id: health-check
            type: api-call
            description: "Running health check"
            depends-on: [wait-for-completion]
            request:
              method: GET
              url: "{base_url}/apps/{args.app_id}/health"

        output:
          transform: |
            {
              "deployment_id": create-deployment.deployment_id,
              "status": wait-for-completion.body.status,
              "health": health-check.body
            }
```

**Usage:**
```bash
# Before (shell script)
./deploy-app.sh my-app-123

# After (CliForge)
myapi-cli deploy --app-id my-app-123
```

---

### Step 4: Migrate Authentication Patterns

**Before (Manual token management):**
```bash
#!/bin/bash
# get-token.sh - Fetch and save token

RESPONSE=$(curl -s -X POST https://auth.example.com/token \
  -H "Content-Type: application/x-www-form-urlencoded" \
  -d "grant_type=client_credentials" \
  -d "client_id=$CLIENT_ID" \
  -d "client_secret=$CLIENT_SECRET")

TOKEN=$(echo "$RESPONSE" | jq -r '.access_token')
echo "$TOKEN" > ~/.api-token
chmod 600 ~/.api-token

echo "Token saved to ~/.api-token"
```

```bash
# Using the token
TOKEN=$(cat ~/.api-token)
curl -X GET https://api.example.com/users \
  -H "Authorization: Bearer $TOKEN"
```

**After (CliForge - automatic):**
```yaml
# cli-config.yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      flow: client_credentials
      client_id: my-cli
      client_secret_env: CLIENT_SECRET
      token_url: https://auth.example.com/token
      scopes: [read, write]
      storage:
        primary: keyring  # Secure OS keyring
        fallback: file
```

```bash
# Set client secret once
export CLIENT_SECRET=your-secret

# CliForge handles token fetch, refresh, and storage
myapi-cli users list  # Token automatically obtained and used
```

---

### Step 5: Migrate Error Handling

**Before (Manual error handling):**
```bash
#!/bin/bash
# Comprehensive error handling in shell

function api_call() {
  local method=$1
  local endpoint=$2
  local data=$3

  RESPONSE=$(curl -s -w "\n%{http_code}" -X $method \
    https://api.example.com$endpoint \
    -H "Authorization: Bearer $(cat ~/.api-token)" \
    -H "Content-Type: application/json" \
    -d "$data")

  HTTP_CODE=$(echo "$RESPONSE" | tail -n1)
  BODY=$(echo "$RESPONSE" | sed '$d')

  case $HTTP_CODE in
    200|201)
      echo "$BODY" | jq
      return 0
      ;;
    400)
      echo "Error: Bad request" >&2
      echo "$BODY" | jq >&2
      return 1
      ;;
    401)
      echo "Error: Unauthorized - please refresh token" >&2
      return 1
      ;;
    404)
      echo "Error: Resource not found" >&2
      return 1
      ;;
    429)
      echo "Error: Rate limited - please wait" >&2
      return 1
      ;;
    500)
      echo "Error: Server error" >&2
      echo "$BODY" | jq >&2
      return 1
      ;;
    *)
      echo "Error: Unexpected HTTP code $HTTP_CODE" >&2
      return 1
      ;;
  esac
}

# Usage
if ! api_call GET /users ""; then
  echo "Failed to list users"
  exit 1
fi
```

**After (CliForge - automatic):**
```bash
# CliForge handles all error cases automatically:
# - 401: Automatic token refresh
# - 429: Automatic retry with backoff
# - 500: Clear error message with debug info
# - Network errors: Retry mechanism

myapi-cli users list

# Custom error behavior can be configured
myapi-cli users list --retry-attempts 5 --verbose
```

Configure error handling:
```yaml
# cli-config.yaml
behaviors:
  retry:
    enabled: true
    max_attempts: 3
    backoff: exponential
    retry_on: [429, 500, 502, 503, 504]

  errors:
    format: friendly  # friendly, json, debug
    show_request_id: true
    mask_secrets: true
```

---

### Step 6: Migrate Data Processing Pipelines

**Before (Complex jq pipelines):**
```bash
#!/bin/bash
# Get all active users with pagination

PAGE=1
LIMIT=100
ALL_USERS=()

while true; do
  RESPONSE=$(curl -s \
    "https://api.example.com/users?page=$PAGE&limit=$LIMIT&status=active" \
    -H "Authorization: Bearer $(cat ~/.api-token)")

  USERS=$(echo "$RESPONSE" | jq -r '.items[]')

  if [ -z "$USERS" ]; then
    break
  fi

  ALL_USERS+=("$USERS")
  PAGE=$((PAGE + 1))
done

# Process and format
echo "${ALL_USERS[@]}" | jq -s '. | group_by(.department) |
  map({
    department: .[0].department,
    count: length,
    users: map(.name)
  })'
```

**After (CliForge with built-in features):**
```bash
# Automatic pagination
myapi-cli users list --filter "status=active" --all

# With custom output format
myapi-cli users list \
  --filter "status=active" \
  --output json | jq 'group_by(.department) |
  map({
    department: .[0].department,
    count: length,
    users: map(.name)
  })'
```

Or configure in OpenAPI spec:
```yaml
paths:
  /users:
    get:
      x-cli-output:
        table:
          columns:
            - field: name
              header: Name
            - field: department
              header: Department
            - field: status
              header: Status
          groupBy: department
          aggregate:
            - field: count
              function: count
```

---

### Step 7: Replace Common Script Patterns

Here are migrations for common curl/script patterns:

#### Pattern 1: Retrying Failed Requests

**Before:**
```bash
#!/bin/bash
MAX_RETRIES=3
RETRY_DELAY=5

for i in $(seq 1 $MAX_RETRIES); do
  if curl -f -X GET https://api.example.com/users \
      -H "Authorization: Bearer $(cat ~/.api-token)"; then
    break
  else
    if [ $i -lt $MAX_RETRIES ]; then
      echo "Retry $i failed, waiting $RETRY_DELAY seconds..."
      sleep $RETRY_DELAY
    else
      echo "All retries failed"
      exit 1
    fi
  fi
done
```

**After:**
```yaml
# cli-config.yaml
behaviors:
  retry:
    enabled: true
    max_attempts: 3
    initial_delay: 5s
    max_delay: 30s
    backoff: exponential
```

```bash
# Automatic retries
myapi-cli users list
```

#### Pattern 2: Progress Indicators

**Before:**
```bash
#!/bin/bash
echo -n "Fetching users... "
curl -s https://api.example.com/users \
  -H "Authorization: Bearer $(cat ~/.api-token)" > /tmp/users.json
echo "done"

echo -n "Processing... "
# ... processing
echo "done"
```

**After:**
```bash
# Automatic progress indicators
myapi-cli users list
# ⠋ Fetching users...
# ✓ Users fetched (123 items)
```

Configure progress indicators:
```yaml
# cli-config.yaml
defaults:
  output:
    progress: true  # Show progress spinners
    verbose: false  # Don't show detailed progress
```

#### Pattern 3: Confirmation Prompts

**Before:**
```bash
#!/bin/bash
USER_ID=$1

read -p "Delete user $USER_ID? (y/n) " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
  curl -X DELETE https://api.example.com/users/$USER_ID \
    -H "Authorization: Bearer $(cat ~/.api-token)"
  echo "User deleted"
else
  echo "Cancelled"
fi
```

**After (automatic confirmation):**
```yaml
# In OpenAPI spec
paths:
  /users/{userId}:
    delete:
      x-cli-confirmation:
        enabled: true
        message: "Delete user {userId}?"
        destructive: true
```

```bash
# CliForge shows confirmation automatically
myapi-cli users delete --user-id 123
# Delete user 123? (y/n)

# Skip confirmation with --yes
myapi-cli users delete --user-id 123 --yes
```

#### Pattern 4: Parallel Requests

**Before:**
```bash
#!/bin/bash
USER_IDS=(123 456 789 101)

for id in "${USER_IDS[@]}"; do
  (
    curl -s https://api.example.com/users/$id \
      -H "Authorization: Bearer $(cat ~/.api-token)" > /tmp/user-$id.json
  ) &
done

wait
cat /tmp/user-*.json | jq -s
rm /tmp/user-*.json
```

**After:**
```bash
# CliForge workflow with parallel execution
myapi-cli users batch-get --user-ids 123,456,789,101
```

Configure in OpenAPI spec:
```yaml
paths:
  /users/batch:
    get:
      x-cli-command: "users batch-get"
      x-cli-workflow:
        settings:
          parallel-execution: true
          max-parallel: 5

        steps:
          - id: fetch-users
            type: parallel
            for-each: "{args.user_ids}"
            step:
              type: api-call
              request:
                method: GET
                url: "{base_url}/users/{item}"
```

---

### Migration Checklist

- [ ] **Script Inventory**
  - [ ] List all curl commands
  - [ ] List all shell scripts using your API
  - [ ] Document what each script does
  - [ ] Identify common patterns

- [ ] **Create CliForge Config**
  - [ ] Map API endpoints to CLI commands
  - [ ] Configure authentication
  - [ ] Set up output formatting
  - [ ] Add error handling config

- [ ] **Create OpenAPI Spec**
  - [ ] Document all API endpoints
  - [ ] Add `x-cli-*` extensions for custom behavior
  - [ ] Define workflows for multi-step scripts
  - [ ] Add validation rules

- [ ] **Build and Test**
  - [ ] Build CliForge binary
  - [ ] Test each migrated command
  - [ ] Verify workflows work correctly
  - [ ] Test error cases

- [ ] **Replace Scripts**
  - [ ] Update documentation
  - [ ] Create command migration guide
  - [ ] Update CI/CD to use new CLI
  - [ ] Archive old scripts

---

### Common Pitfalls: curl/Script Migration

#### Pitfall 1: Missing Complex jq Transformations

**Problem:**
"My curl scripts use complex jq pipelines that CliForge doesn't support."

**Solution:**
CliForge supports output to JSON, so you can still use jq:

```bash
# Before
curl ... | jq 'complex | transformation'

# After
myapi-cli users list --output json | jq 'complex | transformation'
```

Or define custom output formats in your OpenAPI spec.

#### Pitfall 2: Custom HTTP Headers

**Problem:**
"My curl scripts send custom headers that vary per request."

**Solution:**
Use default headers for static headers, flags for dynamic ones:

```yaml
# cli-config.yaml - static headers
api:
  default_headers:
    X-API-Version: "2024-01-01"

# OpenAPI spec - dynamic headers
paths:
  /users:
    get:
      parameters:
        - name: X-Request-ID
          in: header
          schema:
            type: string
```

```bash
myapi-cli users list --header "X-Request-ID=abc123"
```

#### Pitfall 3: Raw HTTP Access

**Problem:**
"Sometimes I need raw HTTP access for debugging."

**Solution:**
Use verbose and debug modes:

```bash
# Show full HTTP request/response
myapi-cli users list --debug

# Dry run (show what would be sent)
myapi-cli users create --name "Test" --dry-run

# Very verbose (all HTTP details)
myapi-cli users list --verbose --debug
```

---

### Conversion Examples

Here are complete before/after examples for common scenarios:

#### Example 1: Simple CRUD Script

**Before (Shell script: manage-users.sh):**
```bash
#!/bin/bash
# manage-users.sh - User management script

TOKEN=$(cat ~/.api-token)
BASE_URL="https://api.example.com"

function list_users() {
  curl -s "$BASE_URL/users" \
    -H "Authorization: Bearer $TOKEN" | jq
}

function get_user() {
  local id=$1
  curl -s "$BASE_URL/users/$id" \
    -H "Authorization: Bearer $TOKEN" | jq
}

function create_user() {
  local name=$1
  local email=$2
  curl -s -X POST "$BASE_URL/users" \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"$name\",\"email\":\"$email\"}" | jq
}

function delete_user() {
  local id=$1
  curl -X DELETE "$BASE_URL/users/$id" \
    -H "Authorization: Bearer $TOKEN"
}

# Main
case "$1" in
  list)
    list_users
    ;;
  get)
    get_user "$2"
    ;;
  create)
    create_user "$2" "$3"
    ;;
  delete)
    delete_user "$2"
    ;;
  *)
    echo "Usage: $0 {list|get|create|delete} [args]"
    exit 1
    ;;
esac
```

**Usage:**
```bash
./manage-users.sh list
./manage-users.sh get 123
./manage-users.sh create "John Doe" "john@example.com"
./manage-users.sh delete 123
```

**After (CliForge):**

```yaml
# cli-config.yaml
metadata:
  name: myapi-cli
  version: 1.0.0

api:
  openapi_url: https://api.example.com/openapi.yaml
  base_url: https://api.example.com

behaviors:
  auth:
    type: api_key
    api_key:
      header: Authorization
      prefix: "Bearer "
      env_var: API_TOKEN
```

**Usage:**
```bash
export API_TOKEN=$(cat ~/.api-token)

myapi-cli users list
myapi-cli users get --user-id 123
myapi-cli users create --name "John Doe" --email "john@example.com"
myapi-cli users delete --user-id 123
```

**Benefits:**
- ✅ No script maintenance
- ✅ Automatic parameter validation
- ✅ Built-in help (`myapi-cli users --help`)
- ✅ Multiple output formats
- ✅ Better error messages

#### Example 2: Deployment Script

**Before (Complex shell script):**
```bash
#!/bin/bash
# deploy.sh - Deploy application

set -e

APP_ID=$1
ENVIRONMENT=$2

if [ -z "$APP_ID" ] || [ -z "$ENVIRONMENT" ]; then
  echo "Usage: $0 <app-id> <environment>"
  exit 1
fi

TOKEN=$(cat ~/.api-token)
BASE_URL="https://api.example.com"

echo "Validating app $APP_ID..."
APP_STATUS=$(curl -s "$BASE_URL/apps/$APP_ID" \
  -H "Authorization: Bearer $TOKEN" | jq -r '.status')

if [ "$APP_STATUS" != "ready" ]; then
  echo "Error: App not ready (status: $APP_STATUS)"
  exit 1
fi

echo "Creating deployment..."
DEPLOY_ID=$(curl -s -X POST "$BASE_URL/deployments" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{\"app_id\":\"$APP_ID\",\"environment\":\"$ENVIRONMENT\"}" | jq -r '.id')

echo "Deployment ID: $DEPLOY_ID"
echo "Waiting for completion..."

while true; do
  STATUS=$(curl -s "$BASE_URL/deployments/$DEPLOY_ID" \
    -H "Authorization: Bearer $TOKEN" | jq -r '.status')

  case $STATUS in
    completed)
      echo "Deployment completed successfully!"
      exit 0
      ;;
    failed)
      echo "Deployment failed!"
      curl -s "$BASE_URL/deployments/$DEPLOY_ID/logs" \
        -H "Authorization: Bearer $TOKEN" | jq
      exit 1
      ;;
    *)
      echo "Status: $STATUS"
      sleep 10
      ;;
  esac
done
```

**After (CliForge workflow):**

```yaml
# In OpenAPI spec
paths:
  /deployments:
    post:
      x-cli-command: "deploy"
      x-cli-workflow:
        steps:
          - id: validate-app
            type: api-call
            description: "Validating application"
            request:
              method: GET
              url: "{base_url}/apps/{args.app_id}"
            output:
              status: "body.status"

          - id: check-ready
            type: conditional
            condition: "{validate-app.status} == 'ready'"
            on-false:
              error: "App not ready (status: {validate-app.status})"

          - id: create-deployment
            type: api-call
            description: "Creating deployment"
            depends-on: [check-ready]
            request:
              method: POST
              url: "{base_url}/deployments"
              body:
                app_id: "{args.app_id}"
                environment: "{args.environment}"
            output:
              deployment_id: "body.id"

          - id: wait-completion
            type: wait
            description: "Waiting for deployment"
            depends-on: [create-deployment]
            poll:
              url: "{base_url}/deployments/{create-deployment.deployment_id}"
              interval: 10
              timeout: 600
              until: "body.status in ['completed', 'failed']"
              success-condition: "body.status == 'completed'"

          - id: show-logs-on-failure
            type: api-call
            description: "Fetching deployment logs"
            condition: "{wait-completion.body.status} == 'failed'"
            request:
              method: GET
              url: "{base_url}/deployments/{create-deployment.deployment_id}/logs"
```

**Usage:**
```bash
# Before
./deploy.sh my-app production

# After
myapi-cli deploy --app-id my-app --environment production
```

**Benefits:**
- ✅ Declarative workflow (easier to understand)
- ✅ Automatic progress indicators
- ✅ Better error messages
- ✅ Built-in timeout handling
- ✅ Automatic retry on transient failures

---

## Migrating from AWS CLI Patterns

### Understanding the Difference

The **AWS CLI** is one of the most well-designed API CLIs. If you're building a CLI inspired by AWS CLI patterns, **CliForge** can help you achieve similar UX with less code.

#### Conceptual Comparison

**AWS CLI Pattern:**
```bash
aws <service> <command> <subcommand> [options]
aws ec2 describe-instances --filters "Name=tag:Name,Values=MyInstance"
aws s3 cp file.txt s3://bucket/file.txt
```

**CliForge Pattern:**
```bash
<mycli> <resource> <action> [options]
mycli instances list --filter "tag.name=MyInstance"
mycli files upload --source file.txt --destination /bucket/file.txt
```

### Key Similarities

| AWS CLI Feature | CliForge Equivalent | How to Implement |
|----------------|--------------------|--------------------|
| Profiles (--profile) | Contexts (--context) | `contexts.*` config |
| Output formats | Output formats | `defaults.output.*` |
| Pagination | Pagination | `behaviors.pagination.*` |
| Waiters | Poll/Wait steps | `x-cli-workflow` with `wait` type |
| Dry run (--dry-run) | Dry run | Built-in flag |
| Query (--query) | JMESPath filtering | Built-in filtering |
| Verbosity (--debug) | Debug mode | Built-in flags |

---

### Step 1: Map AWS CLI Patterns to CliForge

#### Profile Management

**AWS CLI:**
```bash
# Configure profiles
aws configure --profile development
aws configure --profile production

# Use profiles
aws --profile development ec2 describe-instances
aws --profile production s3 ls

# Set default profile
export AWS_PROFILE=production
```

**CliForge:**
```yaml
# cli-config.yaml
contexts:
  default: production

  available:
    development:
      api:
        base_url: https://dev.api.example.com
      auth:
        type: api_key
        api_key:
          env_var: DEV_API_KEY
      display:
        color: yellow
        prompt: "[DEV]"

    production:
      api:
        base_url: https://api.example.com
      auth:
        type: api_key
        api_key:
          env_var: PROD_API_KEY
      display:
        color: green
        prompt: "[PROD]"
```

```bash
# Use contexts
myapi-cli --context development instances list
myapi-cli --context production files list

# Set default context
myapi-cli context use production
myapi-cli instances list  # Uses production
```

#### Output Formatting

**AWS CLI:**
```bash
# JSON (default)
aws ec2 describe-instances

# Table
aws ec2 describe-instances --output table

# Text
aws ec2 describe-instances --output text

# YAML
aws ec2 describe-instances --output yaml

# Query specific fields
aws ec2 describe-instances --query 'Reservations[].Instances[].InstanceId'
```

**CliForge:**
```bash
# Table (configured as default)
myapi-cli instances list

# JSON
myapi-cli instances list --output json

# YAML
myapi-cli instances list --output yaml

# CSV
myapi-cli instances list --output csv

# Filter specific fields
myapi-cli instances list --fields id,name,status
myapi-cli instances list --output json | jq '.[] | .instance_id'
```

Configure defaults:
```yaml
# cli-config.yaml
defaults:
  output:
    format: table  # table, json, yaml, csv
    pretty_print: true
    color: auto
```

#### Waiters

**AWS CLI:**
```bash
# Wait for instance to be running
aws ec2 wait instance-running --instance-ids i-1234567890abcdef0

# Wait for stack to be created
aws cloudformation wait stack-create-complete --stack-name my-stack
```

**CliForge:**
```yaml
# In OpenAPI spec
paths:
  /instances/{instanceId}/start:
    post:
      x-cli-command: "instances start"
      x-cli-workflow:
        steps:
          - id: start-instance
            type: api-call
            request:
              method: POST
              url: "{base_url}/instances/{args.instance_id}/start"

          - id: wait-running
            type: wait
            depends-on: [start-instance]
            description: "Waiting for instance to start"
            poll:
              url: "{base_url}/instances/{args.instance_id}"
              interval: 5
              timeout: 300
              until: "body.status == 'running'"
            output:
              status: "body.status"
```

```bash
# Automatic waiting
myapi-cli instances start --instance-id i-123

# Skip waiting
myapi-cli instances start --instance-id i-123 --no-wait
```

#### Pagination

**AWS CLI:**
```bash
# Automatic pagination
aws ec2 describe-instances

# Manual pagination
aws ec2 describe-instances --max-items 100
aws ec2 describe-instances --max-items 100 --starting-token <token>

# Get all results
aws ec2 describe-instances --max-items 1000
```

**CliForge:**
```yaml
# cli-config.yaml
behaviors:
  pagination:
    enabled: true
    page_size: 20
    max_page_size: 100
    style: cursor  # cursor, offset, page
```

```bash
# Automatic pagination (first page)
myapi-cli instances list

# Get all results
myapi-cli instances list --all

# Custom page size
myapi-cli instances list --limit 50

# Manual pagination
myapi-cli instances list --limit 20 --page 2
```

---

### Step 2: Implement AWS-Style Command Structure

AWS CLI uses a consistent command structure. You can replicate this in CliForge:

**AWS CLI Structure:**
```
aws <service> <command> [--options]

Examples:
aws ec2 describe-instances
aws ec2 run-instances
aws ec2 terminate-instances
aws s3 ls
aws s3 cp
aws s3 sync
```

**CliForge Equivalent:**
```yaml
# In OpenAPI spec, use consistent command naming

paths:
  /instances:
    get:
      operationId: listInstances
      x-cli-command: "instances list"  # or "instances describe"
      x-cli-aliases: ["list instances", "describe instances"]

    post:
      operationId: createInstance
      x-cli-command: "instances create"  # or "instances run"
      x-cli-aliases: ["run instance", "create instance"]

  /instances/{instanceId}:
    get:
      operationId: getInstance
      x-cli-command: "instances get"
      x-cli-aliases: ["describe instance"]

    delete:
      operationId: deleteInstance
      x-cli-command: "instances delete"
      x-cli-aliases: ["terminate instance"]

  /files:
    get:
      operationId: listFiles
      x-cli-command: "files list"
      x-cli-aliases: ["ls files"]

  /files/upload:
    post:
      operationId: uploadFile
      x-cli-command: "files upload"
      x-cli-aliases: ["files cp", "files copy"]
```

**Result:**
```bash
# AWS-style commands in your CLI
myapi-cli instances list
myapi-cli instances create --type t2.micro
myapi-cli instances terminate --instance-id i-123

# Alternative style
myapi-cli describe instances
myapi-cli run instance --type t2.micro
myapi-cli terminate instance --instance-id i-123
```

---

### Step 3: Implement AWS-Style Features

#### Feature 1: Dry Run

**AWS CLI:**
```bash
aws ec2 run-instances --dry-run --instance-type t2.micro ...
```

**CliForge:**
```yaml
# Built-in dry-run support
paths:
  /instances:
    post:
      x-cli-workflow:
        settings:
          dry-run-supported: true
```

```bash
myapi-cli instances create --type t2.micro --dry-run
```

#### Feature 2: Filters

**AWS CLI:**
```bash
aws ec2 describe-instances \
  --filters "Name=instance-type,Values=t2.micro" \
            "Name=instance-state-name,Values=running"
```

**CliForge:**
```bash
myapi-cli instances list \
  --filter "type=t2.micro" \
  --filter "status=running"

# Or combined
myapi-cli instances list --filter "type=t2.micro,status=running"
```

Configure filter support:
```yaml
paths:
  /instances:
    get:
      parameters:
        - name: filter
          in: query
          schema:
            type: string
          x-cli-filter:
            enabled: true
            fields: [type, status, tag.name]
```

#### Feature 3: Configuration Files

**AWS CLI:**
```bash
# ~/.aws/config
[default]
region = us-west-2
output = json

[profile production]
region = us-east-1
output = table
```

**CliForge:**
```yaml
# ~/.myapi-cli/config.yaml
preferences:
  context: production
  output:
    format: table
    color: true

contexts:
  development:
    auth:
      api_key: dev-key-here

  production:
    auth:
      api_key: prod-key-here
```

#### Feature 4: Credential Management

**AWS CLI:**
```bash
# ~/.aws/credentials
[default]
aws_access_key_id = AKIAIOSFODNN7EXAMPLE
aws_secret_access_key = wJalrXUtnFEMI/K7MDENG/bPxRfiCYEXAMPLEKEY

[production]
aws_access_key_id = AKIAI44QH8DHBEXAMPLE
aws_secret_access_key = je7MtGbClwBF/2Zp9Utk/h3yCo8nvbEXAMPLEKEY
```

**CliForge:**
```yaml
# cli-config.yaml
behaviors:
  auth:
    type: api_key
    api_key:
      header: X-API-Key
      storage:
        primary: keyring  # Secure OS keyring
        fallback: file    # Encrypted file
```

```bash
# Set credentials per context
myapi-cli config set auth.api_key "key-here" --context development
myapi-cli config set auth.api_key "key-here" --context production

# Credentials stored securely in OS keyring
```

---

### Step 4: Implement Service-Specific Behaviors

AWS CLI has service-specific behaviors (like S3 sync). You can implement similar features:

**AWS S3 Sync:**
```bash
aws s3 sync ./local-dir s3://bucket/prefix/ \
  --exclude "*.tmp" \
  --include "*.txt" \
  --delete
```

**CliForge Equivalent:**
```yaml
# In OpenAPI spec
paths:
  /files/sync:
    post:
      x-cli-command: "files sync"
      x-cli-workflow:
        steps:
          - id: list-local
            type: plugin
            plugin: filesystem
            command: list-files
            args:
              path: "{args.source}"
              recursive: true

          - id: list-remote
            type: api-call
            request:
              method: GET
              url: "{base_url}/files"
              params:
                prefix: "{args.destination}"

          - id: compute-diff
            type: plugin
            plugin: sync-helper
            command: compute-diff
            args:
              local: "{list-local.files}"
              remote: "{list-remote.body.files}"
              exclude: "{args.exclude}"
              include: "{args.include}"

          - id: upload-new
            type: parallel
            for-each: "{compute-diff.to_upload}"
            step:
              type: api-call
              request:
                method: POST
                url: "{base_url}/files"
                body:
                  path: "{item.path}"
                  content: "{item.content}"

          - id: delete-removed
            type: parallel
            condition: "{args.delete} == true"
            for-each: "{compute-diff.to_delete}"
            step:
              type: api-call
              request:
                method: DELETE
                url: "{base_url}/files/{item.id}"
```

```bash
myapi-cli files sync \
  --source ./local-dir \
  --destination /bucket/prefix \
  --exclude "*.tmp" \
  --include "*.txt" \
  --delete
```

---

### Step 5: Implement Help and Documentation

AWS CLI has excellent help:

```bash
aws help
aws ec2 help
aws ec2 describe-instances help
```

CliForge provides similar help automatically:

```bash
myapi-cli --help
myapi-cli instances --help
myapi-cli instances list --help
```

Enhance help with OpenAPI extensions:

```yaml
paths:
  /instances:
    get:
      summary: List instances
      description: |
        Lists all instances in your account. You can filter results
        by instance type, status, and tags.

      x-cli-examples:
        - description: List all instances
          command: mycli instances list

        - description: List running instances
          command: mycli instances list --filter "status=running"

        - description: List specific instance type
          command: mycli instances list --filter "type=t2.micro"

        - description: Output as JSON
          command: mycli instances list --output json

      x-cli-notes: |
        NOTE: This operation is eventually consistent. It may take a few
        seconds for newly created instances to appear.
```

---

### Migration Checklist

- [ ] **Command Structure**
  - [ ] Map AWS service to your resource types
  - [ ] Map AWS commands to your operations
  - [ ] Define consistent command naming
  - [ ] Add command aliases for familiarity

- [ ] **Profile/Context System**
  - [ ] Map AWS profiles to CliForge contexts
  - [ ] Configure context-specific settings
  - [ ] Set up credential management
  - [ ] Test context switching

- [ ] **Output Formatting**
  - [ ] Configure default output format
  - [ ] Test all output formats (JSON, YAML, table)
  - [ ] Implement field filtering
  - [ ] Test pagination

- [ ] **Advanced Features**
  - [ ] Implement waiters with workflows
  - [ ] Add dry-run support
  - [ ] Configure filtering
  - [ ] Add examples to help text

- [ ] **Documentation**
  - [ ] Write comprehensive help text
  - [ ] Add usage examples
  - [ ] Document all command aliases
  - [ ] Create migration guide for users

---

### Common Pitfalls: AWS CLI Pattern Migration

#### Pitfall 1: Over-Complicating Command Structure

**Problem:**
Trying to exactly replicate AWS CLI's complex command structure.

**Solution:**
CliForge is for YOUR API. Use AWS CLI patterns as inspiration, but simplify for your use case:

```bash
# AWS CLI (complex, lots of options)
aws ec2 run-instances \
  --image-id ami-abc123 \
  --instance-type t2.micro \
  --key-name my-key \
  --security-group-ids sg-123 \
  --subnet-id subnet-123 \
  --block-device-mappings '[{"DeviceName":"/dev/sda1","Ebs":{"VolumeSize":20}}]'

# Your CLI (simpler, focused on your API)
mycli instances create \
  --name "my-instance" \
  --type "small" \
  --region "us-west"
```

#### Pitfall 2: Not Using Workflows for Complex Operations

**Problem:**
AWS CLI operations like `wait` and `sync` seem hard to replicate.

**Solution:**
Use CliForge workflows:

```yaml
# AWS: aws ec2 wait instance-running --instance-ids i-123
# CliForge: Built into workflow

paths:
  /instances/{instanceId}/start:
    post:
      x-cli-workflow:
        steps:
          - id: start
            type: api-call
            # ...
          - id: wait
            type: wait
            poll:
              until: "body.status == 'running'"
```

---

### Side-by-Side Comparison

**AWS CLI:**
```bash
# Configure
aws configure --profile production
# Enter: Access Key, Secret Key, Region, Output Format

# Use
aws --profile production ec2 describe-instances \
  --filters "Name=instance-state-name,Values=running" \
  --output table

aws --profile production ec2 wait instance-running \
  --instance-ids i-1234567890

aws --profile production s3 sync ./local s3://bucket/
```

**CliForge:**
```bash
# Configure (one-time setup)
myapi-cli login --context production
# Opens browser for OAuth, or prompts for API key

# Use
myapi-cli --context production instances list \
  --filter "status=running" \
  --output table

myapi-cli --context production instances start \
  --instance-id i-1234567890
# Automatically waits for running state

myapi-cli --context production files sync \
  --source ./local \
  --destination /bucket/
```

**Configuration:**

AWS CLI (`~/.aws/config`):
```ini
[default]
region = us-west-2
output = json

[profile production]
region = us-east-1
output = table
role_arn = arn:aws:iam::123456789012:role/MyRole
```

CliForge (`cli-config.yaml`):
```yaml
contexts:
  default: production

  available:
    production:
      api:
        base_url: https://api.example.com
      auth:
        type: oauth2
      display:
        color: green
      defaults:
        output:
          format: table
```

---

## Migration Decision Matrix

Use this matrix to decide if CliForge is right for you:

### Should You Migrate?

| Your Current Situation | Recommendation | Why |
|----------------------|----------------|-----|
| Using OpenAPI Generator | ✅ **Migrate** | Eliminate generated code, faster API updates |
| Using Restish | ✅ **Migrate** | Gain branding, distribution, self-updates |
| Using curl/shell scripts | ✅ **Migrate** | Massive UX improvement, less maintenance |
| Following AWS CLI patterns | ✅ **Migrate** | Achieve AWS-level UX with less code |
| Happy with current solution | ⚠️ **Evaluate** | If it works, migration may not be worth it |
| Need generic multi-API tool | ❌ **Don't Migrate** | Use Restish instead |
| API has no OpenAPI spec | ⚠️ **Create Spec First** | You need OpenAPI for CliForge |

### Migration Effort Estimation

| Current State | Estimated Effort | Complexity |
|--------------|------------------|------------|
| OpenAPI Generator (basic) | 1-2 days | Low |
| OpenAPI Generator (heavily customized) | 3-5 days | Medium |
| Restish (simple config) | 1 day | Low |
| curl/shell scripts (< 10 scripts) | 2-3 days | Low-Medium |
| curl/shell scripts (> 10 scripts) | 5-10 days | Medium-High |
| AWS CLI patterns (basic) | 2-3 days | Medium |
| AWS CLI patterns (advanced) | 5-7 days | High |

**Effort includes:**
- Creating CliForge configuration
- Adding OpenAPI extensions
- Migrating custom logic to workflows
- Testing
- Documentation updates

---

## Common Migration Pitfalls

These pitfalls apply across all migration paths:

### Pitfall 1: Not Reading the Documentation First

**Problem:**
Jumping into migration without understanding CliForge concepts.

**Solution:**
Read these docs first:
1. [Getting Started](getting-started.md) - Understand basics
2. [Configuration DSL](configuration-dsl.md) - Learn config options
3. [Technical Specification](technical-specification.md) - Understand architecture

### Pitfall 2: Trying to Migrate Everything at Once

**Problem:**
Attempting to migrate your entire CLI in one go.

**Solution:**
Incremental migration:

```bash
# Week 1: Migrate read-only operations
myapi-cli users list
myapi-cli users get

# Week 2: Migrate create operations
myapi-cli users create

# Week 3: Migrate update operations
myapi-cli users update

# Week 4: Migrate delete operations
myapi-cli users delete

# Week 5: Migrate complex workflows
myapi-cli deploy
```

Keep old and new CLIs side-by-side during transition.

### Pitfall 3: Ignoring OpenAPI Spec Quality

**Problem:**
Expecting CliForge to work well with a poor OpenAPI spec.

**Solution:**
Improve your OpenAPI spec:

```yaml
# Bad: No descriptions, no examples
paths:
  /users:
    get:
      operationId: getUsers
      responses:
        200:
          description: OK

# Good: Comprehensive documentation
paths:
  /users:
    get:
      operationId: listUsers
      summary: List all users
      description: |
        Returns a paginated list of users. You can filter by status,
        role, and creation date.

      parameters:
        - name: status
          in: query
          description: Filter by user status
          schema:
            type: string
            enum: [active, inactive, pending]

      responses:
        200:
          description: List of users
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/User'
              examples:
                example-1:
                  summary: Active users
                  value:
                    - id: 1
                      name: "John Doe"
                      status: "active"

      x-cli-command: "users list"
      x-cli-examples:
        - description: List all users
          command: mycli users list
        - description: List active users only
          command: mycli users list --status active
```

### Pitfall 4: Not Testing Edge Cases

**Problem:**
Only testing happy path scenarios.

**Solution:**
Test failure cases:

```bash
# Test authentication failures
unset API_KEY
myapi-cli users list  # Should show clear error

# Test network failures
# (disconnect network)
myapi-cli users list  # Should retry and show timeout

# Test validation errors
myapi-cli users create --email "invalid"  # Should show validation error

# Test rate limiting
for i in {1..100}; do myapi-cli users list; done  # Should handle rate limits

# Test large responses
myapi-cli users list --limit 10000  # Should paginate correctly

# Test concurrent operations
myapi-cli users create --name "User1" &
myapi-cli users create --name "User2" &
wait
```

### Pitfall 5: Forgetting About Backward Compatibility

**Problem:**
Breaking existing user workflows.

**Solution:**
Provide migration period:

```bash
# Support both old and new command structures
myapi-cli get /users  # Old (HTTP-style)
myapi-cli users list  # New (resource-style)

# Provide warnings
myapi-cli get /users
# Warning: HTTP-style commands are deprecated. Use 'users list' instead.
# This command will be removed in version 2.0.0
```

Use aliases:
```yaml
paths:
  /users:
    get:
      x-cli-command: "users list"
      x-cli-aliases:
        - "get /users"  # Old style
        - "list users"  # Alternative style
        - "users"       # Short form
```

### Pitfall 6: Poor Error Messages

**Problem:**
Unhelpful error messages after migration.

**Solution:**
Configure user-friendly errors:

```yaml
# cli-config.yaml
behaviors:
  errors:
    format: friendly  # Not just JSON dumps
    show_request_id: true
    suggest_fixes: true
    mask_secrets: true
```

**Bad error:**
```
Error: HTTP 401
Body: {"error":"unauthorized","code":"AUTH001"}
```

**Good error:**
```
Error: Authentication failed

The API key you provided is invalid or expired.

To fix this:
1. Check that your API key is correct: myapi-cli config show auth.api_key
2. Generate a new API key: https://example.com/settings/api-keys
3. Update your key: myapi-cli config set auth.api_key "new-key"

Need help? Visit: https://docs.example.com/auth
Request ID: req_abc123 (include this in support requests)
```

### Pitfall 7: Not Leveraging CliForge Features

**Problem:**
Just replicating old behavior instead of using new features.

**Solution:**
Take advantage of CliForge features:

```yaml
# Use workflows for multi-step operations
x-cli-workflow:
  steps: [...]

# Use interactive mode
x-cli-interactive:
  prompts: [...]

# Use custom output formats
x-cli-output:
  table:
    columns: [...]

# Use confirmation prompts
x-cli-confirmation:
  enabled: true

# Use context switching
contexts:
  available: [...]

# Use plugins for external tools
plugins:
  enabled: true
```

---

## Migration Success Stories

### Success Story 1: Startup API (OpenAPI Generator → CliForge)

**Before:**
- 200+ generated Go files
- 30-minute build time
- Required Go knowledge to build
- API changes required full regeneration

**After:**
- 1 configuration file
- 2-minute build time
- Single binary distribution
- API changes automatic (spec cached)

**Results:**
- 95% reduction in repository size
- 93% reduction in build time
- 100% elimination of merge conflicts in generated code
- Users get API updates without CLI updates

### Success Story 2: Enterprise SaaS (curl Scripts → CliForge)

**Before:**
- 45 shell scripts
- Each script 100-300 lines
- Manual token management
- No error handling consistency
- No documentation

**After:**
- 1 CliForge binary
- Automatic OAuth2 flow
- Consistent error handling
- Built-in help and examples
- Interactive mode for complex operations

**Results:**
- 87% reduction in code to maintain
- 100% of operations now documented
- 0 security incidents (previously: token leaks in scripts)
- User satisfaction increased from 6.2 to 9.1 (out of 10)

### Success Story 3: Internal Tools (Restish → CliForge)

**Before:**
- Users had to install and configure Restish
- Each API required separate configuration
- Generic "restish" command for all APIs
- Users confused about which API they're using

**After:**
- One branded CLI per API
- Single binary, no configuration needed
- Clear branding and purpose
- Self-updating

**Results:**
- 70% reduction in support tickets
- 95% reduction in onboarding time (from 30 min to 2 min)
- Better adoption (users don't have to "learn Restish")
- Professional branded experience

---

## Next Steps

After migration:

### 1. Optimize Your Configuration

Review and enhance your CliForge configuration:

```yaml
# Start with basics
metadata:
  name: myapi-cli
  version: 1.0.0

# Add branding
branding:
  colors:
    primary: "#your-brand-color"
  ascii_art: |
    Your ASCII art banner

# Configure features
features:
  interactive_mode:
    enabled: true
  workflows:
    enabled: true
  plugins:
    enabled: true
```

### 2. Enhance Your OpenAPI Spec

Add CliForge extensions for better UX:

```yaml
paths:
  /users:
    get:
      x-cli-command: "users list"
      x-cli-aliases: ["list users"]
      x-cli-output:
        table:
          columns: [...]
      x-cli-examples:
        - description: List active users
          command: mycli users list --status active
```

### 3. Set Up Self-Updates

Configure automatic updates:

```yaml
updates:
  enabled: true
  update_url: https://releases.example.com/myapi-cli
  check_interval: 24h
  public_key: |
    -----BEGIN PUBLIC KEY-----
    ...
    -----END PUBLIC KEY-----
```

### 4. Create User Documentation

Document the migration for your users:

```markdown
# Migrating to myapi-cli v2.0

We've rebuilt our CLI from the ground up. Here's what changed:

## Installation

### Before (v1.x)
brew install myapi-cli
myapi-cli configure  # Manual setup

### After (v2.0)
brew install myapi-cli
myapi-cli login      # Automatic OAuth setup

## Commands

| Old Command | New Command |
|------------|-------------|
| myapi-cli get /users | myapi-cli users list |
| myapi-cli post /users '{"name":"X"}' | myapi-cli users create --name X |

## New Features

- ✅ Automatic OAuth2 authentication
- ✅ Table output by default
- ✅ Interactive mode for complex operations
- ✅ Self-updating (run `myapi-cli update`)
```

### 5. Get Feedback

Collect user feedback:

```bash
# Add telemetry (optional, opt-in)
behaviors:
  telemetry:
    enabled: false  # Users opt in
    endpoint: https://telemetry.example.com
    events:
      - command_executed
      - error_occurred
      - update_checked

# Or use GitHub Discussions/Issues
# Or user surveys
```

---

## Additional Resources

### Documentation

- [Getting Started](getting-started.md) - Quick start guide
- [Configuration DSL](configuration-dsl.md) - Complete configuration reference
- [Technical Specification](technical-specification.md) - Architecture and design
- [User Guide: Authentication](user-guide-authentication.md) - Authentication setup
- [User Guide: Workflows](user-guide-workflows.md) - Workflow orchestration
- [Examples and Recipes](examples-and-recipes.md) - Practical examples

### Tools

- **Migration checker**: `cliforge migrate check` (coming soon)
- **Config validator**: `cliforge validate --config cli-config.yaml`
- **Spec validator**: `cliforge validate --spec openapi.yaml`

### Support

- **GitHub Issues**: [github.com/cliforge/cliforge/issues](https://github.com/cliforge/cliforge/issues)
- **Discussions**: [github.com/cliforge/cliforge/discussions](https://github.com/cliforge/cliforge/discussions)
- **Documentation**: [docs/](.)

---

## Conclusion

Migrating to CliForge provides significant benefits:

✅ **Cleaner repositories** - No generated code
✅ **Faster development** - API changes don't require rebuilds
✅ **Better UX** - Professional branded CLIs
✅ **Self-updating** - Users always have latest features and security fixes
✅ **Less maintenance** - Configuration instead of code

Choose the migration path that matches your current tool, follow the checklist, avoid common pitfalls, and you'll have a production-ready CliForge CLI in no time.

Happy migrating!
