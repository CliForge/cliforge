# Hybrid API-Driven CLI: Technical Specification

## Project Name: **CliForge**

**Tagline**: Forge CLIs from APIs
**Version**: 0.7.0
**Date**: 2025-01-11
**Status**: Design Phase

---

## Table of Contents

1. [Executive Summary](#executive-summary)
2. [System Overview](#system-overview)
3. [Architecture](#architecture)
4. [Components](#components)
5. [Data Flow](#data-flow)
6. [Configuration DSL](#configuration-dsl)
7. [OpenAPI Extensions](#openapi-extensions)
8. [Security](#security)
9. [Performance Requirements](#performance-requirements)
10. [Error Handling](#error-handling)
11. [Testing Strategy](#testing-strategy)
12. [Deployment & Distribution](#deployment--distribution)
13. [Future Enhancements](#future-enhancements)

---

## Executive Summary

### Problem Statement

Existing OpenAPI/Swagger CLI tools fall into two categories:
- **Static generators** (OpenAPI Generator, openapi-cli-generator): Require regeneration for API changes
- **Dynamic loaders** (Restish): Cannot be branded or pre-configured for distribution

**Note**: CliForge supports both **Swagger 2.0** and **OpenAPI 3.x** specifications, providing maximum compatibility with legacy and modern APIs.

Neither approach supports:
1. Branded binary distribution with embedded configuration
2. Dynamic API updates without binary rebuilds
3. Self-updating binaries for security patches
4. Behavioral change notifications to end users

### Solution

A hybrid CLI generation system that:
1. **Generates** branded, self-contained binaries at build time
2. **Loads** OpenAPI specifications dynamically at runtime
3. **Updates** itself for security patches and configuration changes
4. **Notifies** users of API behavioral changes

### Key Innovation

The system separates concerns:
- **Binary-level concerns**: Branding, URLs, security patches → handled via self-update
- **API-level concerns**: Endpoints, operations, schemas → handled via dynamic spec loading
- **Behavioral concerns**: Auth flows, rate limits, caching → configurable at both levels

---

## System Overview

### High-Level Architecture

```
┌──────────────────────────────────────────────────────────────────┐
│                        DEVELOPMENT TIME                          │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Developer writes:                                               │
│  ┌────────────────┐                                             │
│  │ cli-config.yaml│  (Branding, URLs, Behaviors)                │
│  └────────┬───────┘                                             │
│           │                                                      │
│           ▼                                                      │
│  ┌────────────────┐                                             │
│  │   Generator    │  → Compiles → ┌──────────────────┐          │
│  │   CLI Tool     │               │  Branded Binary  │          │
│  └────────────────┘               │  (my-api-cli)    │          │
│                                    └──────────────────┘          │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│                         DISTRIBUTION                             │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  Release Server:                                                 │
│  ┌────────────────────────────────────────────────────┐          │
│  │ https://releases.example.com/                      │          │
│  │ ├── latest/version.json                            │          │
│  │ ├── binaries/                                      │          │
│  │ │   ├── my-api-cli-v1.2.3-darwin-amd64            │          │
│  │ │   ├── my-api-cli-v1.2.3-darwin-arm64            │          │
│  │ │   ├── my-api-cli-v1.2.3-linux-amd64             │          │
│  │ │   └── my-api-cli-v1.2.3-windows-amd64.exe       │          │
│  │ └── checksums/                                     │          │
│  │     └── v1.2.3-checksums.txt                       │          │
│  └────────────────────────────────────────────────────┘          │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────────────────────┐
│                         RUNTIME                                  │
├──────────────────────────────────────────────────────────────────┤
│                                                                  │
│  User runs: my-api-cli users list                               │
│                                                                  │
│  Binary execution flow:                                          │
│  1. ┌─────────────────────────┐                                 │
│     │ Check for binary update │                                 │
│     └────────┬────────────────┘                                 │
│              │ (every 24h or --update flag)                     │
│              ▼                                                   │
│  2. ┌─────────────────────────┐                                 │
│     │ Fetch OpenAPI spec      │ ← https://api.example.com/      │
│     └────────┬────────────────┘    openapi.yaml                 │
│              │ (cached for 5min)                                │
│              ▼                                                   │
│  3. ┌─────────────────────────┐                                 │
│     │ Parse spec → commands   │                                 │
│     └────────┬────────────────┘                                 │
│              │                                                   │
│              ▼                                                   │
│  4. ┌─────────────────────────┐                                 │
│     │ Check for API changes   │                                 │
│     └────────┬────────────────┘                                 │
│              │ (compare versions)                               │
│              ▼                                                   │
│  5. ┌─────────────────────────┐                                 │
│     │ Show changelog if new   │                                 │
│     └────────┬────────────────┘                                 │
│              │                                                   │
│              ▼                                                   │
│  6. ┌─────────────────────────┐                                 │
│     │ Execute command         │                                 │
│     └────────┬────────────────┘                                 │
│              │                                                   │
│              ▼                                                   │
│  7. ┌─────────────────────────┐                                 │
│     │ Format & display result │                                 │
│     └─────────────────────────┘                                 │
│                                                                  │
└──────────────────────────────────────────────────────────────────┘
```

---

## Architecture

### Components Overview

```
┌─────────────────────────────────────────────────────────────┐
│                    Generated Binary                         │
├─────────────────────────────────────────────────────────────┤
│                                                             │
│  ┌──────────────────────────────────────────────────────┐  │
│  │              Embedded Assets Layer                    │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │ • Configuration (config.yaml)                         │  │
│  │ • Branding assets (ASCII art, colors)                │  │
│  │ • Version info                                        │  │
│  │ • API endpoint URLs                                   │  │
│  │ • Public keys (for signature verification)           │  │
│  └──────────────────────────────────────────────────────┘  │
│                           │                                 │
│  ┌────────────────────────▼─────────────────────────────┐  │
│  │              Runtime Engine Layer                     │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │                                                       │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────┐  │  │
│  │  │   Updater    │  │ Spec Loader  │  │  Command  │  │  │
│  │  │   Manager    │  │   & Parser   │  │  Builder  │  │  │
│  │  └──────────────┘  └──────────────┘  └───────────┘  │  │
│  │                                                       │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────┐  │  │
│  │  │    Cache     │  │     Auth     │  │  Output   │  │  │
│  │  │   Manager    │  │   Manager    │  │ Formatter │  │  │
│  │  └──────────────┘  └──────────────┘  └───────────┘  │  │
│  │                                                       │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌───────────┐  │  │
│  │  │  Changelog   │  │     HTTP     │  │   Config  │  │  │
│  │  │   Detector   │  │    Client    │  │   Reader  │  │  │
│  │  └──────────────┘  └──────────────┘  └───────────┘  │  │
│  │                                                       │  │
│  └───────────────────────────────────────────────────────┘  │
│                           │                                 │
│  ┌────────────────────────▼─────────────────────────────┐  │
│  │           CLI Framework Layer (Cobra)                 │  │
│  ├──────────────────────────────────────────────────────┤  │
│  │ • Command tree                                        │  │
│  │ • Flag parsing                                        │  │
│  │ • Help generation                                     │  │
│  │ • Shell completion                                    │  │
│  └──────────────────────────────────────────────────────┘  │
│                                                             │
└─────────────────────────────────────────────────────────────┘
```

### Core Subsystems

#### 1. Generator Subsystem
Converts configuration → branded binary

**Input**: `cli-config.yaml`
**Output**: Platform-specific binaries
**Technology**: Go compiler with embed directives

#### 2. Update Subsystem
Manages binary self-updates

**Components**:
- Version checker
- Binary downloader
- Signature verifier
- Binary replacer
- Process restarter

#### 3. Spec Loading Subsystem
Fetches and parses OpenAPI specifications

**Components**:
- HTTP fetcher
- Cache manager
- OpenAPI parser
- Version tracker

#### 4. Command Building Subsystem
Converts OpenAPI operations → CLI commands

**Components**:
- Operation mapper
- Parameter parser
- Command tree builder
- Help generator

#### 5. Execution Subsystem
Executes API operations

**Components**:
- HTTP client
- Auth provider
- Request builder
- Response handler
- Error mapper

#### 6. Notification Subsystem
Detects and displays changes

**Components**:
- Changelog detector
- Version comparator
- User notifier
- Deprecation warner

---

## Components

### 1. Generator CLI (`cliforge`)

**Purpose**: Developer-facing tool to generate branded binaries

**Usage**:
```bash
# Initialize new CLI project
cliforge init my-api-cli

# Generate binaries
cliforge build --config cli-config.yaml --output dist/

# Validate configuration
cliforge validate cli-config.yaml
```

**Responsibilities**:
- Parse configuration YAML
- Validate configuration
- Embed assets into binary
- Cross-compile for all platforms
- Generate checksums
- Create release artifacts

**Output Structure**:
```
dist/
├── my-api-cli-v1.0.0-darwin-amd64
├── my-api-cli-v1.0.0-darwin-arm64
├── my-api-cli-v1.0.0-linux-amd64
├── my-api-cli-v1.0.0-linux-arm64
├── my-api-cli-v1.0.0-windows-amd64.exe
└── checksums.txt
```

---

### 2. Updater Manager

**Purpose**: Handle binary self-updates

**State Machine**:
```
┌─────────┐
│ Startup │
└────┬────┘
     │
     ▼
┌──────────────┐      No      ┌──────────┐
│ Check Update │────────────>│ Continue │
│   Needed?    │              └──────────┘
└──────┬───────┘
       │ Yes
       ▼
┌──────────────┐
│ Fetch Latest │
│  Version     │
└──────┬───────┘
       │
       ▼
┌──────────────┐      Auto     ┌──────────┐
│ User Consent │─────────────>│ Download │
│   Required?  │               └─────┬────┘
└──────┬───────┘                     │
       │ Prompt                       │
       ▼                              │
┌──────────────┐      Yes             │
│ Ask User     │──────────────────────┘
└──────┬───────┘
       │ No
       ▼
┌──────────────┐
│   Skip       │
└──────────────┘

Download Flow:
┌──────────────┐
│   Download   │
│    Binary    │
└──────┬───────┘
       │
       ▼
┌──────────────┐      Fail    ┌──────────┐
│    Verify    │────────────>│  Abort   │
│  Signature   │              └──────────┘
└──────┬───────┘
       │ Pass
       ▼
┌──────────────┐
│   Replace    │
│    Binary    │
└──────┬───────┘
       │
       ▼
┌──────────────┐
│   Restart    │
│   Process    │
└──────────────┘
```

**API**:
```go
type UpdateManager interface {
    // Check if update is available
    CheckForUpdate(ctx context.Context) (*UpdateInfo, error)

    // Download and install update
    PerformUpdate(ctx context.Context, info *UpdateInfo) error

    // Verify update signature
    VerifySignature(binary []byte, signature []byte) error

    // Replace current binary
    ReplaceBinary(newBinary []byte) error

    // Restart process
    Restart() error
}

type UpdateInfo struct {
    Version      string
    DownloadURL  string
    Signature    []byte
    Changelog    []ChangelogEntry
    RequiredBy   *time.Time  // Mandatory update deadline
    CVEs         []string     // Security fixes
}
```

**Update Check Strategy**:
- Check on startup if last check > 24 hours
- Check with `--update` flag
- Skip with `--no-update` flag
- Background check (non-blocking)

**Security**:
- Ed25519 signature verification
- SHA-256 checksum validation
- HTTPS-only downloads
- Public key embedded in binary

---

### 3. Spec Loader & Parser

**Purpose**: Fetch, cache, and parse OpenAPI specifications

**Caching Strategy**:
```
┌────────────────────────────────────────┐
│         Cache Decision Tree            │
├────────────────────────────────────────┤
│                                        │
│  Spec requested                        │
│       │                                │
│       ▼                                │
│  ┌──────────┐     Yes    ┌──────────┐ │
│  │ In cache?│──────────>│  Valid?  │ │
│  └────┬─────┘            └────┬─────┘ │
│       │ No                    │ Yes   │
│       │                       ▼       │
│       │                  ┌──────────┐ │
│       │                  │ Use cache│ │
│       │                  └──────────┘ │
│       │                       │ No    │
│       ▼                       ▼       │
│  ┌──────────────────────────────┐    │
│  │      Fetch from URL          │    │
│  └──────────┬───────────────────┘    │
│             │                         │
│             ▼                         │
│  ┌──────────────────────────────┐    │
│  │      Parse & Validate        │    │
│  └──────────┬───────────────────┘    │
│             │                         │
│             ▼                         │
│  ┌──────────────────────────────┐    │
│  │      Store in cache          │    │
│  └──────────┬───────────────────┘    │
│             │                         │
│             ▼                         │
│        Return spec                    │
│                                        │
└────────────────────────────────────────┘
```

**Cache Invalidation**:
- TTL-based (configurable, default 5 minutes)
- ETag/If-None-Match support
- Manual refresh with `--refresh` flag
- Version change detection

**API**:
```go
type SpecLoader interface {
    // Load spec (from cache or network)
    LoadSpec(ctx context.Context) (*openapi3.T, error)

    // Refresh spec (bypass cache)
    RefreshSpec(ctx context.Context) (*openapi3.T, error)

    // Get spec metadata
    GetSpecVersion() string
    GetSpecChangelog() []ChangelogEntry
}

type SpecCache struct {
    Path       string
    TTL        time.Duration
    Spec       *openapi3.T
    Version    string
    FetchedAt  time.Time
    ETag       string
}
```

**Error Handling**:
- Network errors → use cached spec if available
- Parse errors → fail with helpful message
- No cache + network error → fail gracefully

---

### 4. Command Builder

**Purpose**: Convert OpenAPI operations to CLI commands

**Mapping Rules**:

| OpenAPI Element | CLI Element | Example |
|----------------|-------------|---------|
| `paths./users` | Command | `users` |
| `get /users` | Subcommand | `list` |
| `post /users` | Subcommand | `create` |
| `get /users/{id}` | Subcommand + arg | `get <id>` |
| `parameters` | Flags | `--limit`, `--filter` |
| `requestBody` | Flags or stdin | `--data` or `-` |
| `x-cli-aliases` | Aliases | `ls`, `list` |

**Command Tree Example**:
```
OpenAPI:
  /users:
    get: listUsers
    post: createUser
  /users/{id}:
    get: getUser
    put: updateUser
    delete: deleteUser

CLI:
my-api-cli
├── users
│   ├── list     [GET /users]
│   ├── create   [POST /users]
│   ├── get      [GET /users/{id}]
│   ├── update   [PUT /users/{id}]
│   └── delete   [DELETE /users/{id}]
└── --help
```

**API**:
```go
type CommandBuilder interface {
    // Build command tree from OpenAPI spec
    BuildCommands(spec *openapi3.T) (*cobra.Command, error)

    // Build single command from operation
    BuildOperation(path string, method string, op *openapi3.Operation) (*cobra.Command, error)

    // Apply behavioral configs
    ApplyBehaviors(cmd *cobra.Command, behaviors Behaviors) error
}
```

**Command Properties**:
- `Use`: operation name
- `Short`: operation summary
- `Long`: operation description
- `Example`: from `x-cli-examples`
- `Aliases`: from `x-cli-aliases`
- `Args`: path parameters
- `Flags`: query parameters + headers

---

### 5. Execution Subsystem

**Purpose**: Execute API operations

**Request Pipeline**:
```
┌─────────────────────────────────────────────┐
│         Request Execution Pipeline          │
├─────────────────────────────────────────────┤
│                                             │
│  1. Parse command flags                     │
│     └─> Extract parameters                  │
│                                             │
│  2. Build HTTP request                      │
│     ├─> Method (GET/POST/etc.)             │
│     ├─> URL (base + path + query)          │
│     ├─> Headers                             │
│     └─> Body                                │
│                                             │
│  3. Apply auth                              │
│     ├─> API key → header                    │
│     ├─> OAuth2 → bearer token               │
│     └─> Basic → base64 encode               │
│                                             │
│  4. Apply behaviors                         │
│     ├─> Rate limiting                       │
│     ├─> Retry logic                         │
│     ├─> Timeout                             │
│     └─> User-Agent                          │
│                                             │
│  5. Execute request                         │
│     └─> HTTP client                         │
│                                             │
│  6. Handle response                         │
│     ├─> Check status code                   │
│     ├─> Parse body                          │
│     ├─> Handle errors                       │
│     └─> Extract data                        │
│                                             │
│  7. Format output                           │
│     ├─> JSON                                │
│     ├─> YAML                                │
│     ├─> Table                               │
│     └─> Custom template                     │
│                                             │
│  8. Display to user                         │
│     └─> stdout/stderr                       │
│                                             │
└─────────────────────────────────────────────┘
```

**API**:
```go
type Executor interface {
    // Execute operation
    Execute(ctx context.Context, req *Request) (*Response, error)
}

type Request struct {
    Method      string
    Path        string
    Query       map[string]string
    Headers     map[string]string
    Body        interface{}
    Auth        AuthConfig
}

type Response struct {
    StatusCode  int
    Headers     map[string][]string
    Body        []byte
    Error       *APIError
}
```

---

### 6. Notification Subsystem

**Purpose**: Detect and notify users of API changes

**Change Detection**:
```go
type ChangeDetector interface {
    // Detect changes between versions
    DetectChanges(oldSpec, newSpec *openapi3.T) []Change

    // Check if changes are breaking
    IsBreaking(changes []Change) bool

    // Get user-facing changelog
    GetChangelog(spec *openapi3.T) []ChangelogEntry
}

type Change struct {
    Type        ChangeType  // Added, Removed, Modified
    Severity    Severity    // Breaking, Dangerous, Safe
    Path        string      // API path affected
    Description string
}

type ChangeType string
const (
    ChangeAdded    ChangeType = "added"
    ChangeRemoved  ChangeType = "removed"
    ChangeModified ChangeType = "modified"
)

type Severity string
const (
    SeverityBreaking  Severity = "breaking"
    SeverityDangerous Severity = "dangerous"
    SeveritySafe      Severity = "safe"
)
```

**Notification Display**:
```
┌─────────────────────────────────────────────────┐
│  🔔 API Changes Detected (v2.1.0 → v2.2.0)     │
├─────────────────────────────────────────────────┤
│                                                 │
│  ✨ New Features:                               │
│    • GET /v2/analytics - Analytics endpoint     │
│    • POST /v2/webhooks - Webhook management     │
│                                                 │
│  ⚠️  Deprecations:                              │
│    • GET /v1/users - Use /v2/users instead      │
│      (Will be removed on 2025-12-31)            │
│                                                 │
│  🔧 Changes:                                    │
│    • POST /users now requires 'email' field     │
│    • GET /users?limit max increased to 1000     │
│                                                 │
│  Run 'my-api-cli changelog' for full details    │
│                                                 │
└─────────────────────────────────────────────────┘
```

---

## Data Flow

### Startup Sequence

```
1. Parse embedded config
2. Initialize logging
3. Check for binary updates (if enabled && interval passed)
   ├─> Update available? → Prompt/Auto-update
   └─> No update → Continue
4. Load OpenAPI spec
   ├─> Check cache
   │   ├─> Valid? → Use cached
   │   └─> Invalid/Missing → Fetch from URL
   └─> Parse spec
5. Detect API changes
   ├─> Compare versions
   └─> Show changelog if new
6. Build command tree
7. Execute user command
8. Format & display output
```

### Command Execution Flow

```
User: my-api-cli users create --name "John" --email "john@example.com"

1. Command Router
   └─> Match: users → create

2. Parameter Parser
   ├─> Flags: name="John", email="john@example.com"
   └─> Validate against OpenAPI schema

3. Request Builder
   ├─> Method: POST
   ├─> URL: https://api.example.com/users
   ├─> Body: {"name": "John", "email": "john@example.com"}
   └─> Headers: Content-Type, Authorization, User-Agent

4. Auth Provider
   └─> Add bearer token (from cached OAuth2 flow)

5. HTTP Client
   ├─> Apply rate limiting
   ├─> Apply retry logic
   └─> Execute request

6. Response Handler
   ├─> Status: 201 Created
   ├─> Body: {"id": 123, "name": "John", ...}
   └─> Extract data

7. Output Formatter (JSON default)
   └─> Pretty print with colors

8. Display
   {
     "id": 123,
     "name": "John",
     "email": "john@example.com",
     "created_at": "2025-11-09T10:30:00Z"
   }
```

---

## Configuration DSL

See separate documents:
- **[`configuration-dsl.md`](configuration-dsl.md)** - Complete configuration schema and reference
- **[`builtin-commands-design.md`](https://github.com/wgordon17/cliforge/blob/main/design/architecture/builtin-commands-design.md)** - Built-in commands and global flags design

**Key configuration areas:**
- **Metadata**: CLI name, version, description, author
- **Branding**: Colors, ASCII art, prompts
- **API**: OpenAPI spec URL, base URL
- **Behaviors**: Auth, caching, output, rate limiting, retry, built-in commands, global flags
- **Features**: Telemetry, completion, validation
- **Commands**: Hidden operations, aliases, custom commands
- **Hooks**: Lifecycle event scripts
- **Updates**: Self-update configuration

---

## OpenAPI Extensions

### `x-cli-version`

**Purpose**: Track API version for changelog detection

**Location**: `info` object

**Example**:
```yaml
openapi: 3.1.0
info:
  title: My API
  version: 2.1.0
  x-cli-version: "2024.11.09.1"
  x-cli-min-version: "1.0.0"  # Minimum CLI version required
```

---

### `x-cli-changelog`

**Purpose**: Document user-facing changes

**Location**: `info` object

**Schema**:
```yaml
x-cli-changelog:
  - date: "2024-11-09"
    version: "2.1.0"
    changes:
      - type: added | removed | modified | deprecated | security
        severity: breaking | dangerous | safe
        description: "Human-readable description"
        path: "/users"  # Optional: affected path
        migration: "How to migrate"  # Optional
```

**Example**:
```yaml
x-cli-changelog:
  - date: "2024-11-09"
    version: "2.1.0"
    changes:
      - type: added
        severity: safe
        description: "New analytics endpoint"
        path: "/analytics"
      - type: deprecated
        severity: dangerous
        description: "GET /v1/users is deprecated"
        path: "/v1/users"
        migration: "Use GET /v2/users instead"
        sunset: "2025-12-31"
      - type: modified
        severity: breaking
        description: "POST /users now requires email field"
        path: "/users"
        migration: "Add 'email' to request body"
```

**Note**: For comprehensive deprecation handling (including the `x-cli-deprecation` extension, warning levels, user controls, and migration automation), see [`deprecation-strategy.md`](https://github.com/wgordon17/cliforge/blob/main/design/architecture/deprecation-strategy.md).

---

### `x-cli-aliases`

**Purpose**: Define command aliases

**Location**: `operation` object

**Example**:
```yaml
paths:
  /users:
    get:
      operationId: listUsers
      x-cli-aliases: ["ls", "list"]
      x-cli-description: "List all users"
```

**Result**:
```bash
my-api-cli users list    # Primary
my-api-cli users ls      # Alias
```

---

### `x-cli-examples`

**Purpose**: Provide usage examples

**Location**: `operation` object

**Example**:
```yaml
paths:
  /users:
    get:
      operationId: listUsers
      x-cli-examples:
        - description: "List first 10 users"
          command: "my-api-cli users list --limit 10"
        - description: "List admin users"
          command: "my-api-cli users list --filter 'role=admin'"
```

---

### `x-cli-hidden`

**Purpose**: Hide operations from CLI

**Location**: `operation` object

**Example**:
```yaml
paths:
  /internal/debug:
    get:
      operationId: debugEndpoint
      x-cli-hidden: true  # Not exposed in CLI
```

---

### `x-cli-auth`

**Purpose**: Specify auth requirements per operation

**Location**: `operation` object

**Example**:
```yaml
paths:
  /users:
    get:
      x-cli-auth:
        required: true
        scopes: ["users:read"]
```

---

### `x-cli-workflow`

**Purpose**: Define multi-step API workflows with chaining, templating, and transformation

**Location**: `operation` object

**Use Case**: Commands that need to call multiple APIs (including external services like AWS, GitHub) and combine results

**Expression Language**: Uses [expr](https://expr-lang.org) for conditions and transformations (Go-like syntax, sandboxed, type-safe)

#### Basic Structure

```yaml
x-cli-workflow:
  steps:
    # Regular API call step
    - id: string                    # Unique step identifier (used to reference results)
      request:
        method: string               # GET, POST, PUT, DELETE, PATCH
        url: string                  # URL template (can use {variables})
        headers: map                 # Optional headers
        body: object                 # Optional request body (for POST/PUT)
        query: map                   # Optional query parameters
      condition: string              # Optional: expr expression (e.g., "status == 200")

    # Iteration step (for calling API for each item)
    - id: string
      foreach: string                # expr expression returning array (e.g., "get-objects.body.data")
      as: string                     # Loop variable name (e.g., "item")
      request:
        method: string
        url: string                  # Can reference loop variable: {item.id}
      condition: string              # Optional: filter items (e.g., "item.status == 'active'")

  output:
    format: string                   # json, yaml, table, csv
    transform: string                # expr expression to transform combined results
```

**Key Changes from Initial Design**:
- ❌ Removed redundant `output` field - reference steps directly by `id` (e.g., `get-objects.body.field`)
- ✅ Added `foreach` step type for iteration (handled at YAML level, not expr)
- ✅ Using `expr` for all conditions and transformations (safe, fast, type-safe)
- ✅ Template interpolation `{variable}` handled by our parser, evaluated with expr

#### Example 1: Multi-Step API Orchestration

```yaml
paths:
  /objects:
    get:
      operationId: listObjects
      summary: List objects with AWS metadata
      x-cli-workflow:
        steps:
          # Step 1: Get objects from primary API
          - id: get-objects
            request:
              method: GET
              url: "{base_url}/api/objects"
              query:
                limit: "{args.limit}"

          # Step 2: Get AWS details using region from step 1
          - id: get-aws-metadata
            request:
              method: GET
              url: "https://ec2.{get-objects.body.region}.amazonaws.com"
              headers:
                Authorization: "Bearer {env.AWS_TOKEN}"
              query:
                filters: "instance-id={get-objects.body.instance_ids}"
            condition: "get-objects.status == 200 && get-objects.body.region != null"

          # Step 3: Get GitHub repo info
          - id: get-github-repo
            request:
              method: GET
              url: "https://api.github.com/repos/{get-objects.body.repo}"
              headers:
                Authorization: "token {env.GITHUB_TOKEN}"
                Accept: "application/vnd.github.v3+json"

        output:
          format: json
          transform: |
            {
              "objects": get-objects.body.data,
              "aws_instances": get-aws-metadata.body.instances,
              "github_repo": get-github-repo.body,
              "summary": {
                "total_objects": len(get-objects.body.data),
                "aws_region": get-objects.body.region,
                "repo_stars": get-github-repo.body.stargazers_count
              }
            }
```

#### Example 2: Conditional Workflow

```yaml
paths:
  /deploy:
    post:
      operationId: deployApplication
      x-cli-workflow:
        steps:
          # Step 1: Check deployment readiness
          - id: check-readiness
            request:
              method: GET
              url: "{base_url}/api/deployments/readiness"

          # Step 2: Create deployment (only if ready)
          - id: create-deployment
            request:
              method: POST
              url: "{base_url}/api/deployments"
              body:
                app_id: "{args.app_id}"
                version: "{args.version}"
            condition: "check-readiness.body.ready == true"

          # Step 3: Trigger AWS CodeDeploy
          - id: trigger-codedeploy
            request:
              method: POST
              url: "https://codedeploy.{create-deployment.body.region}.amazonaws.com"
              headers:
                X-Amz-Target: "CodeDeploy_20141006.CreateDeployment"
              body:
                applicationName: "{create-deployment.body.app_name}"
                deploymentGroupName: "{create-deployment.body.group}"
            condition: "create-deployment.status == 201"

        output:
          format: table
          transform: |
            {
              "deployment_id": create-deployment.body.id,
              "status": create-deployment.body.status,
              "codedeploy_id": trigger-codedeploy.body.deploymentId,
              "url": create-deployment.body.url
            }
```

#### Templating Variables

Templates use `{variable}` syntax, which are parsed and evaluated as `expr` expressions:

| Variable | Description | Example |
|----------|-------------|---------|
| `{base_url}` | Configured base URL | `{base_url}/api/users` |
| `{args.name}` | CLI argument/flag value | `{args.limit}`, `{args.user_id}` |
| `{env.VAR}` | Environment variable | `{env.AWS_TOKEN}` |
| `{step-id.body.field}` | Response body from previous step | `{get-objects.body.region}` |
| `{step-id.status}` | HTTP status code | `{create-deployment.status}` |
| `{step-id.headers.Name}` | Response header | `{create-deployment.headers.Location}` |

**Notes**:
- Step IDs with hyphens work in expr: `get-objects.body.field`
- Optional chaining supported: `{step-id.body?.data?[0]?.region}`
- Nil coalescing supported: `{user.email ?? "no-email"}`

#### Transformation Functions (expr)

The `transform` field uses `expr` expressions with built-in functions:

**Data Access**:
- `step-id.body.field` - Access response fields
- `step-id.status` - HTTP status code
- `step-id.headers.Name` - Response headers
- Optional chaining: `step-id.body?.data?[0]`

**Array Functions**:
- `filter(array, predicate)` - Filter with predicate: `filter(users, .age >= 18)`
- `map(array, transform)` - Transform elements: `map(users, .name)`
- `all(array, predicate)` - Check if all match: `all(items, .inStock)`
- `any(array, predicate)` - Check if any match: `any(items, .price < 100)`
- `one(array, predicate)` - Check if exactly one matches
- `none(array, predicate)` - Check if none match
- `count(array, predicate)` - Count matching elements
- `len(array)` - Array length

**String Functions**:
- `trim(str)`, `upper(str)`, `lower(str)`
- `split(str, sep)`, `join(array, sep)`
- `contains(str, substr)`, `startsWith(str, prefix)`, `endsWith(str, suffix)`
- `replace(str, old, new)`

**Type Conversion**:
- `int(value)`, `float(value)`, `string(value)`

**Math Functions**:
- `abs(n)`, `ceil(n)`, `floor(n)`, `round(n)`
- `max(a, b)`, `min(a, b)`

**Example Transformations**:

```yaml
# Combine and filter with expr predicates
transform: |
  {
    "active_objects": filter(get-objects.body.data, .status == "active"),
    "total": len(get-objects.body.data),
    "regions": map(get-objects.body.data, .region)
  }

# Nested filtering and transformation
transform: |
  {
    "users": filter(get-users.body.data, .age >= 18 && .verified == true),
    "names": map(
      filter(get-users.body.data, .active),
      .firstName + " " + .lastName
    )
  }

# Complex aggregation
transform: |
  {
    "summary": {
      "total": len(get-objects.body.data),
      "active_count": count(get-objects.body.data, .status == "active"),
      "has_errors": any(get-objects.body.data, .error != null),
      "all_verified": all(get-objects.body.data, .verified == true)
    }
  }
```

**See**: https://expr-lang.org/docs/language-definition for complete expr syntax reference

#### Example 3: Iteration with `foreach`

```yaml
paths:
  /users/{id}/details:
    get:
      operationId: getUsersWithDetails
      summary: Get all users and fetch details for each
      x-cli-workflow:
        steps:
          # Step 1: Get list of users
          - id: get-users
            request:
              method: GET
              url: "{base_url}/users"

          # Step 2: For each user, fetch detailed profile
          - id: get-user-details
            foreach: "filter(get-users.body.data, .status == 'active')"
            as: user
            request:
              method: GET
              url: "{base_url}/users/{user.id}/details"
              headers:
                X-User-Context: "{user.id}"
            condition: "user.verified == true"

        output:
          format: json
          transform: |
            {
              "users": get-users.body.data,
              "details": get-user-details,
              "summary": {
                "total_users": len(get-users.body.data),
                "active_users": count(get-users.body.data, .status == "active"),
                "details_fetched": len(get-user-details)
              }
            }
```

**How `foreach` Works**:
- The `foreach` field contains an expr expression that returns an array
- Can be a simple reference: `get-users.body.data`
- Or filtered: `filter(get-users.body.data, .active == true)`
- The `as` field names the loop variable (available in `request` and `condition`)
- The `condition` is evaluated per-item (optional - filters which items to process)
- Results are collected into an array named by the step `id`

#### Error Handling

**Step Failures**:
- If a step fails and has no `condition`, workflow stops and returns error
- If a step fails but has a `condition`, it's skipped and workflow continues
- Use `condition` to make steps optional

**Example**:
```yaml
steps:
  - id: optional-step
    request:
      method: GET
      url: "https://api.example.com/optional"
    condition: "true"  # Always run, but don't fail workflow if it errors

  - id: use-optional
    request:
      method: POST
      url: "{base_url}/api/process"
      body:
        data: "{optional-step.body?.data ?? 'default'}"
    # Using optional chaining and nil coalescing to handle missing optional-step
```

#### Performance Considerations

**Parallel Execution** (future enhancement):
```yaml
x-cli-workflow:
  steps:
    - id: get-objects
      request: {...}

    - id: get-aws-metadata
      request: {...}
      parallel: true  # Run in parallel with next step

    - id: get-github-data
      request: {...}
      parallel: true  # Run in parallel with previous step

    - id: combine
      depends_on: [get-aws-metadata, get-github-data]
      # Wait for both parallel steps before executing
      request: {...}
```

#### Caching

Workflow steps respect caching settings:
- GET requests with same URL are cached per `response_ttl`
- POST/PUT/DELETE requests are never cached
- Use `--refresh` flag to bypass cache

#### Security

**Authentication**:
- Steps inherit auth from main operation
- Can override per-step with custom headers
- Environment variables for external API tokens

**Secrets**:
- Use `{env.VAR}` for sensitive data
- Never hardcode tokens in OpenAPI spec
- Tokens stored in OS-specific secure storage

---

## Security

### Binary Signature Verification

**Algorithm**: Ed25519 (fast, secure, small signatures)

**Process**:
1. Developer signs binary with private key
2. Public key embedded in generated binary
3. Update process verifies signature before installation

**Implementation**:
```go
import "crypto/ed25519"

func VerifyBinary(binary, signature []byte, publicKey ed25519.PublicKey) error {
    if !ed25519.Verify(publicKey, binary, signature) {
        return errors.New("signature verification failed")
    }
    return nil
}
```

---

### TLS Certificate Pinning

**Purpose**: Prevent MITM attacks on update server

**Implementation**:
- Embed expected certificate fingerprints
- Verify during TLS handshake
- Fail update if mismatch

---

### Secrets Management

**OAuth2 Tokens**:
- Store in OS-specific secure storage
  - macOS: Keychain
  - Linux: Secret Service API (gnome-keyring, kwallet)
  - Windows: Credential Manager
- Fallback: Encrypted file with OS-specific key

**API Keys**:
- Read from environment variables (preferred)
- Read from config file (warn if world-readable)
- Prompt user (ephemeral, not stored)

---

### Audit Logging

**What to log**:
- Binary updates (old version, new version, timestamp)
- Auth events (login, token refresh, logout)
- API calls (method, path, timestamp, user)
- Errors (type, message, timestamp)

**Where to log**:
- Local: `~/.config/my-api-cli/audit.log`
- Remote (optional): Configurable audit endpoint

---

## Performance Requirements

### Startup Time
- **Target**: < 50ms (cold start without network)
- **Maximum**: < 200ms (including update check)

### Command Execution
- **Network-bound**: As fast as API response
- **Local processing**: < 10ms overhead

### Binary Size
- **Target**: < 15MB per platform
- **Maximum**: < 25MB

### Memory Usage
- **Idle**: < 20MB
- **Active**: < 50MB

---

## Error Handling

### Error Categories

1. **Network Errors**
   - DNS resolution failed
   - Connection timeout
   - TLS handshake failed

2. **API Errors**
   - 4xx client errors
   - 5xx server errors
   - Rate limiting

3. **Parsing Errors**
   - Invalid OpenAPI spec
   - Malformed JSON response

4. **Auth Errors**
   - Expired token
   - Invalid credentials
   - Insufficient permissions

5. **Update Errors**
   - Download failed
   - Signature verification failed
   - Binary replacement failed

### Error Display Format

```
Error: Failed to create user

  API Error: 400 Bad Request

  Validation failed:
    • email: must be a valid email address
    • age: must be at least 18

  Request ID: req_abc123xyz

  For help, run: my-api-cli users create --help

  Documentation: https://docs.example.com/api/users#create
```

---

## Testing Strategy

### Unit Tests
- Configuration parsing
- Command building
- Parameter validation
- Output formatting

### Integration Tests
- Full command execution against mock API
- Update flow with mock update server
- Auth flows

### End-to-End Tests
- Real API calls (test environment)
- Binary generation
- Cross-platform testing

### Test Coverage Target
- Core libraries: 80%+
- Critical paths (auth, updates): 95%+

---

## Deployment & Distribution

### Release Process

1. **Build**
   ```bash
   cliforge build --config cli-config.yaml --version 1.2.3
   ```

2. **Sign**
   ```bash
   for binary in dist/*; do
     openssl dgst -sha256 -sign private.pem -out $binary.sig $binary
   done
   ```

3. **Upload**
   ```bash
   aws s3 sync dist/ s3://releases.example.com/my-api-cli/v1.2.3/
   ```

4. **Update version.json**
   ```json
   {
     "version": "1.2.3",
     "released_at": "2025-11-09T12:00:00Z",
     "changelog": [...],
     "binaries": {
       "darwin-amd64": {
         "url": "https://releases.example.com/.../darwin-amd64",
         "sha256": "..."
       }
     }
   }
   ```

### Installation Methods

**Direct Download**:
```bash
curl -L https://releases.example.com/install.sh | sh
```

**Homebrew** (macOS/Linux):
```bash
brew tap example/tap
brew install my-api-cli
```

**APT** (Debian/Ubuntu):
```bash
echo "deb [trusted=yes] https://apt.example.com stable main" | sudo tee /etc/apt/sources.list.d/my-api-cli.list
sudo apt update && sudo apt install my-api-cli
```

**Scoop** (Windows):
```powershell
scoop bucket add example https://github.com/example/scoop-bucket
scoop install my-api-cli
```

---

## Future Enhancements

### Phase 2
- **Dynamic Custom Functions** (`x-cli-functions` extension)
  - Allow API owners to define custom functions in OpenAPI spec
  - Functions written as expr expressions, compiled at runtime
  - Enable domain-specific abstractions and reusable logic
  - Example: `isValidStatus(order)`, `calculateShipping(order)`, `ec2URL(region)`
  - Provides API-specific extensions without binary recompilation
- Plugin system for custom commands
- Interactive mode (REPL)
- Bash/Zsh/Fish completion generation
- Man page generation

### Phase 3
- Parallel workflow execution (`parallel: true`, `depends_on: [...]`)
- GraphQL support
- gRPC support
- Multi-API configuration (one CLI, many APIs)

### Phase 4
- GUI wrapper (Electron/Tauri)
- VS Code extension
- Docker container with CLI pre-installed

---

## Appendices

### A. Technology Stack

- **Language**: Go 1.21+
- **CLI Framework**: spf13/cobra
- **Config**: spf13/viper
- **OpenAPI Parser**: getkin/kin-openapi
  - Supports Swagger 2.0 (OpenAPI 2.0)
  - Supports OpenAPI 3.0
  - Supports OpenAPI 3.1 (in progress)
  - Includes `openapi2conv` for auto-converting 2.0 → 3.0
- **Expression Language**: expr-lang/expr (for workflows, conditions, transformations)
- **HTTP**: net/http (stdlib)
- **JSON**: encoding/json (stdlib)
- **YAML**: go-yaml/yaml
- **Crypto**: crypto/ed25519 (stdlib)
- **Testing**: testify

**Why kin-openapi?**
- Industry-standard OpenAPI parser
- Active maintenance and community support
- Built-in validation and conversion tools
- Native Go implementation (no CGO dependencies)

### B. OpenAPI & Swagger Compatibility

**CliForge supports all major API specification formats:**

| Format | Status | Notes |
|--------|--------|-------|
| **Swagger 2.0** | ✅ Fully supported | Auto-converts to OpenAPI 3.0 internally via `openapi2conv` |
| **OpenAPI 3.0** | ✅ Fully supported | Primary target format |
| **OpenAPI 3.1** | 🚧 In progress | Via kin-openapi library updates |

**Terminology Clarification:**
- **"Swagger"** historically refers to Swagger 2.0 specification (2014)
- **"OpenAPI"** is the modern name (since 2015) for versions 3.0+
- The specification was donated to the Linux Foundation in 2015 and renamed
- "Swagger" today refers to SmartBear's tooling (Swagger UI, Editor, Codegen)

**Why support both?**
- **Legacy APIs**: Many enterprise APIs still use Swagger 2.0
- **Modern APIs**: New APIs use OpenAPI 3.x
- **Seamless migration**: Auto-conversion allows supporting both without code duplication
- **Maximum adoption**: Developers can use CliForge regardless of their API's spec version

**How it works:**
```go
// CliForge auto-detects and normalizes
if spec.Swagger == "2.0" {
    // Auto-convert to OpenAPI 3.0
    spec3, err := openapi2conv.ToV3(spec)
    // Continue with unified OpenAPI 3.x handling
}
```

**Breaking changes between Swagger 2.0 and OpenAPI 3.0:**
- Server definitions: `host`/`basePath`/`schemes` → `servers[]` array
- Reusable components: `definitions` → `components/schemas`
- Request bodies: part of `parameters` → separate `requestBody`
- JSON Schema: limited → full Draft 5 support (`oneOf`, `anyOf`, `not`)
- Security: basic OAuth2 → OAuth2 fixes + bearer tokens + OpenID Connect

Users don't need to worry about these differences - CliForge handles conversion automatically.

### C. File Locations

**Follows XDG Base Directory Specification on Linux/macOS**

**macOS/Linux**:
- Config: `$XDG_CONFIG_HOME/my-api-cli/config.yaml` (default: `~/.config/my-api-cli/config.yaml`)
- Cache: `$XDG_CACHE_HOME/my-api-cli/` (default: `~/.cache/my-api-cli/`)
- Data: `$XDG_DATA_HOME/my-api-cli/` (default: `~/.local/share/my-api-cli/`)
- State: `$XDG_STATE_HOME/my-api-cli/` (default: `~/.local/state/my-api-cli/`)
- Auth: OS Keychain (macOS Keychain, Linux Secret Service API)

**Windows**:
- Config: `%APPDATA%\my-api-cli\config.yaml`
- Cache: `%LOCALAPPDATA%\my-api-cli\cache\`
- Data: `%LOCALAPPDATA%\my-api-cli\data\`
- Logs: `%LOCALAPPDATA%\my-api-cli\logs\`
- Auth: Credential Manager

**What Goes Where**:
- **Config**: User configuration overrides (`config.yaml`)
- **Cache**: OpenAPI specs, HTTP response cache (ephemeral, can be deleted)
- **Data**: Audit logs, telemetry data (persistent)
- **State**: Last update check timestamp, session state (persistent but less critical)
- **Auth**: OAuth2 tokens, API keys (OS secure storage)

**Implementation Example** (Go):
```go
package paths

import (
    "os"
    "path/filepath"
    "runtime"
)

// GetConfigDir returns XDG_CONFIG_HOME or platform default
func GetConfigDir(appName string) string {
    if runtime.GOOS == "windows" {
        return filepath.Join(os.Getenv("APPDATA"), appName)
    }

    if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
        return filepath.Join(xdg, appName)
    }
    return filepath.Join(os.Getenv("HOME"), ".config", appName)
}

// GetCacheDir returns XDG_CACHE_HOME or platform default
func GetCacheDir(appName string) string {
    if runtime.GOOS == "windows" {
        return filepath.Join(os.Getenv("LOCALAPPDATA"), appName, "cache")
    }

    if xdg := os.Getenv("XDG_CACHE_HOME"); xdg != "" {
        return filepath.Join(xdg, appName)
    }
    return filepath.Join(os.Getenv("HOME"), ".cache", appName)
}

// GetDataDir returns XDG_DATA_HOME or platform default
func GetDataDir(appName string) string {
    if runtime.GOOS == "windows" {
        return filepath.Join(os.Getenv("LOCALAPPDATA"), appName, "data")
    }

    if xdg := os.Getenv("XDG_DATA_HOME"); xdg != "" {
        return filepath.Join(xdg, appName)
    }
    return filepath.Join(os.Getenv("HOME"), ".local", "share", appName)
}

// GetStateDir returns XDG_STATE_HOME or platform default
func GetStateDir(appName string) string {
    if runtime.GOOS == "windows" {
        return filepath.Join(os.Getenv("LOCALAPPDATA"), appName, "state")
    }

    if xdg := os.Getenv("XDG_STATE_HOME"); xdg != "" {
        return filepath.Join(xdg, appName)
    }
    return filepath.Join(os.Getenv("HOME"), ".local", "state", appName)
}
```

### C. Environment Variables

- `MY_API_CLI_API_KEY`: API key (overrides config)
- `MY_API_CLI_BASE_URL`: Base URL (overrides config)
- `MY_API_CLI_CONFIG`: Config file path
- `MY_API_CLI_NO_UPDATE`: Disable update checks
- `MY_API_CLI_DEBUG`: Enable debug logging

### D. Exit Codes

- `0`: Success
- `1`: General error
- `2`: Configuration error
- `3`: Network error
- `4`: Authentication error
- `5`: API error
- `6`: Update error

---

**End of Technical Specification**

*Version 0.7.0*
*Last Updated: 2025-01-11*

**Changelog**:
- v0.7.0 (2025-01-11): **Built-in Commands & Global Flags**
  - Cross-referenced [`builtin-commands-design.md`](https://github.com/wgordon17/cliforge/blob/main/design/architecture/builtin-commands-design.md)
  - Updated configuration DSL reference
- v0.6.0 (2025-01-11): **Swagger/OpenAPI Compatibility**
  - Added OpenAPI & Swagger Compatibility section (Appendix B)
  - Documented auto-conversion from Swagger 2.0 to OpenAPI 3.0
- v0.5.0 (2025-01-11): **Deprecation Strategy**
  - Cross-referenced [`deprecation-strategy.md`](https://github.com/wgordon17/cliforge/blob/main/design/architecture/deprecation-strategy.md)
- v0.4.0 (2025-01-11): **Rebranded to CliForge**
  - Changed project name from "Alpha-Omega" to "CliForge"
  - Updated generator CLI name from `alpha-omega-gen` to `cliforge`
  - Added tagline: "Forge CLIs from APIs"
  - Added comprehensive deprecation strategy document ([`deprecation-strategy.md`](https://github.com/wgordon17/cliforge/blob/main/design/architecture/deprecation-strategy.md))
    - Handles both API deprecations and CLI deprecations
    - Time-based severity escalation
    - User-friendly migration assistance
- v0.3.0 (2025-11-10): **Security & XDG Compliance**
  - User configuration security boundaries
  - XDG Base Directory Specification compliance
- v0.2.0 (2025-11-10): **expr Expression Language**
  - Updated `x-cli-workflow` to use expr expression language
  - Removed redundant `output` field from steps (reference by step `id` directly)
  - Added `foreach` iteration support at YAML level
  - Documented expr syntax for conditions and transformations
  - Added `x-cli-functions` to Future Enhancements (Phase 2)
  - Added expr to Technology Stack
- v0.1.0 (2025-11-09): Initial specification
