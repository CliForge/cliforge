# CliForge Examples and Recipes

**Version**: 0.9.0
**Last Updated**: 2025-11-23

Practical examples and common patterns for building CLI tools with CliForge.

---

## Table of Contents

1. [Common Patterns](#common-patterns)
2. [REST API CLI (Simple CRUD)](#rest-api-cli-simple-crud)
3. [CLI with Authentication](#cli-with-authentication)
4. [CLI with Workflows](#cli-with-workflows)
5. [CLI with Plugins](#cli-with-plugins)
6. [Streaming and Watch Mode](#streaming-and-watch-mode)
7. [Interactive Mode](#interactive-mode)
8. [Context Switching](#context-switching)
9. [Custom Output Formats](#custom-output-formats)
10. [Migration Examples](#migration-examples)
11. [Real-World Use Cases](#real-world-use-cases)

---

## Common Patterns

### Pattern 1: Simple CRUD Operations

**Goal**: Provide basic create, read, update, delete operations for a resource.

**OpenAPI Extensions**:
```yaml
paths:
  /pets:
    get:
      x-cli-command: "list pets"
      x-cli-aliases: ["ls pets", "pets"]
      x-cli-output:
        table:
          columns:
            - field: id
              header: ID
            - field: name
              header: NAME
            - field: status
              header: STATUS

    post:
      x-cli-command: "create pet"
      x-cli-flags:
        - name: name
          source: name
          flag: "--name"
          required: true
        - name: category
          source: category.name
          flag: "--category"
          required: true

  /pets/{petId}:
    get:
      x-cli-command: "get pet"

    put:
      x-cli-command: "update pet"

    delete:
      x-cli-command: "delete pet"
      x-cli-confirmation:
        enabled: true
        message: "Delete pet {petId}?"
```

**Usage**:
```bash
# List resources
mycli list pets

# Create resource
mycli create pet --name "Fluffy" --category cat

# Get specific resource
mycli get pet --pet-id 123

# Update resource
mycli update pet --pet-id 123 --status adopted

# Delete resource (with confirmation)
mycli delete pet --pet-id 123
```

---

### Pattern 2: Filtered Listing

**Goal**: List resources with filtering, sorting, and pagination.

**OpenAPI Extensions**:
```yaml
paths:
  /users:
    get:
      operationId: listUsers
      x-cli-command: "list users"

      parameters:
        - name: status
          in: query
          schema:
            type: string
            enum: [active, inactive, suspended]
          x-cli-flag: "--status"
          x-cli-aliases: ["-s"]

        - name: role
          in: query
          schema:
            type: string
          x-cli-flag: "--role"

        - name: limit
          in: query
          schema:
            type: integer
            default: 20
          x-cli-flag: "--limit"
          x-cli-aliases: ["-l"]

        - name: sort
          in: query
          schema:
            type: string
            enum: [name, email, created_at]
          x-cli-flag: "--sort"

      x-cli-output:
        table:
          columns:
            - field: id
              header: ID
              width: 10
            - field: name
              header: NAME
              width: 25
            - field: email
              header: EMAIL
              width: 30
            - field: role
              header: ROLE
              width: 15
            - field: status
              header: STATUS
              width: 12
              color-map:
                active: green
                inactive: yellow
                suspended: red
          sort-by: "name"
          sort-order: "asc"
```

**Usage**:
```bash
# List all users
mycli list users

# Filter by status
mycli list users --status active

# Filter by role
mycli list users --role admin

# Combine filters
mycli list users --status active --role admin

# Limit results
mycli list users --limit 50

# Sort results
mycli list users --sort email
```

---

### Pattern 3: Batch Operations

**Goal**: Perform operations on multiple resources.

**OpenAPI Extensions**:
```yaml
paths:
  /users/batch-delete:
    delete:
      operationId: batchDeleteUsers
      x-cli-command: "delete users"

      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                user_ids:
                  type: array
                  items:
                    type: integer

      x-cli-flags:
        - name: ids
          source: user_ids
          flag: "--ids"
          type: array
          required: true
          description: "Comma-separated user IDs"

      x-cli-confirmation:
        enabled: true
        message: "Delete {count} users? This cannot be undone."
        flag: "--yes"
```

**Usage**:
```bash
# Delete multiple users
mycli delete users --ids 1,2,3,4,5

# Skip confirmation
mycli delete users --ids 1,2,3 --yes
```

---

## REST API CLI (Simple CRUD)

### Complete Example: Task Management API

**OpenAPI Specification** (`task-api.yaml`):
```yaml
openapi: 3.0.0
info:
  title: Task Management API
  version: 1.0.0
  x-cli-version: "2024.11.23.1"

x-cli-config:
  name: taskmgr
  description: "Task management CLI"
  branding:
    colors:
      primary: "#4A90E2"
      success: "#7ED321"
      error: "#D0021B"

paths:
  /tasks:
    get:
      operationId: listTasks
      summary: List all tasks
      x-cli-command: "list tasks"
      x-cli-aliases: ["ls tasks", "tasks"]

      parameters:
        - name: status
          in: query
          schema:
            type: string
            enum: [todo, in_progress, done]
          x-cli-flag: "--status"

        - name: priority
          in: query
          schema:
            type: string
            enum: [low, medium, high]
          x-cli-flag: "--priority"

      responses:
        '200':
          description: List of tasks

      x-cli-output:
        table:
          columns:
            - field: id
              header: ID
              width: 8
            - field: title
              header: TITLE
              width: 40
            - field: status
              header: STATUS
              width: 15
              color-map:
                todo: yellow
                in_progress: cyan
                done: green
            - field: priority
              header: PRIORITY
              width: 10
              color-map:
                high: red
                medium: yellow
                low: white
            - field: due_date
              header: DUE
              width: 12

    post:
      operationId: createTask
      summary: Create a new task
      x-cli-command: "create task"
      x-cli-aliases: ["add task", "new task"]

      requestBody:
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/Task'

      x-cli-flags:
        - name: title
          source: title
          flag: "--title"
          required: true
          type: string

        - name: description
          source: description
          flag: "--description"
          type: string

        - name: priority
          source: priority
          flag: "--priority"
          type: string
          default: "medium"

        - name: due-date
          source: due_date
          flag: "--due-date"
          type: string

      x-cli-interactive:
        enabled: true
        prompts:
          - parameter: title
            type: text
            message: "Task title:"
            validation: "^.{1,100}$"

          - parameter: priority
            type: select
            message: "Priority:"
            choices:
              - value: low
                label: "Low"
              - value: medium
                label: "Medium"
              - value: high
                label: "High"

      responses:
        '201':
          description: Task created

      x-cli-output:
        success-message: "Task '{title}' created with ID {id}"

  /tasks/{taskId}:
    get:
      operationId: getTask
      summary: Get task details
      x-cli-command: "get task"

      parameters:
        - name: taskId
          in: path
          required: true
          schema:
            type: integer
          x-cli-flag: "--task-id"
          x-cli-aliases: ["-t"]

      responses:
        '200':
          description: Task details

      x-cli-output:
        format: yaml
        template: |
          Task #{{.id}}:
            Title:       {{.title}}
            Status:      {{.status}}
            Priority:    {{.priority}}
            Description: {{.description}}
            Created:     {{.created_at}}
            Due:         {{.due_date}}

    put:
      operationId: updateTask
      summary: Update a task
      x-cli-command: "update task"

      parameters:
        - name: taskId
          in: path
          required: true
          schema:
            type: integer
          x-cli-flag: "--task-id"

      x-cli-flags:
        - name: title
          source: title
          flag: "--title"
        - name: status
          source: status
          flag: "--status"
        - name: priority
          source: priority
          flag: "--priority"

      responses:
        '200':
          description: Task updated

    delete:
      operationId: deleteTask
      summary: Delete a task
      x-cli-command: "delete task"

      parameters:
        - name: taskId
          in: path
          required: true
          schema:
            type: integer
          x-cli-flag: "--task-id"

      x-cli-confirmation:
        enabled: true
        message: "Delete task {taskId}?"
        flag: "--yes"

      responses:
        '204':
          description: Task deleted

components:
  schemas:
    Task:
      type: object
      required:
        - title
      properties:
        id:
          type: integer
        title:
          type: string
        description:
          type: string
        status:
          type: string
          enum: [todo, in_progress, done]
        priority:
          type: string
          enum: [low, medium, high]
        due_date:
          type: string
          format: date
        created_at:
          type: string
          format: date-time
```

**CLI Configuration** (`cli-config.yaml`):
```yaml
metadata:
  name: taskmgr
  version: 1.0.0
  description: Task Management CLI
  author:
    name: Example Corp
    email: support@example.com

branding:
  colors:
    primary: "#4A90E2"
    success: "#7ED321"
    warning: "#F5A623"
    error: "#D0021B"

  ascii_art: |
    ╔════════════════════════════╗
    ║   Task Manager CLI v1.0    ║
    ╚════════════════════════════╝

api:
  openapi_url: https://api.example.com/tasks/openapi.yaml
  base_url: https://api.example.com/tasks

defaults:
  output:
    format: table
    color: auto

updates:
  enabled: true
  update_url: https://releases.example.com/taskmgr

behaviors:
  auth:
    type: api_key
    api_key:
      header: X-API-Key
      env_var: TASKMGR_API_KEY

  caching:
    spec_ttl: 5m
    response_ttl: 1m

features:
  config_file: true
  interactive_mode: true
```

**Usage Examples**:
```bash
# List all tasks
taskmgr list tasks

# Filter tasks
taskmgr list tasks --status todo
taskmgr list tasks --priority high

# Create task (interactive)
taskmgr create task --interactive

# Create task (non-interactive)
taskmgr create task --title "Fix bug" --priority high --due-date 2025-12-01

# Get task details
taskmgr get task --task-id 42

# Update task
taskmgr update task --task-id 42 --status done

# Delete task
taskmgr delete task --task-id 42
```

---

## CLI with Authentication

### Example 1: OAuth2 Authentication

**CLI Configuration**:
```yaml
behaviors:
  auth:
    type: oauth2
    oauth2:
      client_id: mycli-client
      client_secret: ${OAUTH_CLIENT_SECRET}
      auth_url: https://auth.example.com/oauth/authorize
      token_url: https://auth.example.com/oauth/token
      scopes:
        - api:read
        - api:write
      redirect_url: http://localhost:8085/callback
```

**OpenAPI Specification**:
```yaml
components:
  securitySchemes:
    OAuth2:
      type: oauth2
      flows:
        authorizationCode:
          authorizationUrl: https://auth.example.com/oauth/authorize
          tokenUrl: https://auth.example.com/oauth/token
          scopes:
            api:read: Read access
            api:write: Write access

security:
  - OAuth2: [api:read, api:write]
```

**Usage**:
```bash
# First-time login (opens browser)
mycli auth login

# Check auth status
mycli auth status

# Refresh token
mycli auth refresh

# Logout
mycli auth logout

# Use CLI with auth
mycli users list
```

**Auth Flow**:
1. User runs `mycli auth login`
2. CLI opens browser to authorization URL
3. User logs in and grants permissions
4. Browser redirects to localhost callback
5. CLI exchanges code for access token
6. Token stored in OS keychain (macOS/Linux) or Credential Manager (Windows)
7. Subsequent commands use stored token

---

### Example 2: API Key Authentication

**CLI Configuration**:
```yaml
behaviors:
  auth:
    type: api_key
    api_key:
      header: X-API-Key
      env_var: MYCLI_API_KEY
```

**Usage**:
```bash
# Set API key via environment
export MYCLI_API_KEY=sk_live_abc123xyz789

# Or use flag
mycli users list --api-key sk_live_abc123xyz789

# Or store in config
mycli config set auth.api_key sk_live_abc123xyz789
```

---

### Example 3: Multi-Environment Auth

**CLI Configuration**:
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

behaviors:
  auth:
    type: api_key
    api_key:
      header: X-API-Key
      env_var: MYCLI_API_KEY
```

**User Config** (`~/.config/mycli/config.yaml`):
```yaml
profiles:
  production:
    api_key: sk_prod_xxx

  staging:
    api_key: sk_stag_xxx

  development:
    api_key: sk_dev_xxx
```

**Usage**:
```bash
# Use production (default)
mycli users list

# Use staging
mycli --profile staging users list

# Use development
mycli --profile development users list
```

---

## CLI with Workflows

### Example 1: Multi-Step Deployment

**OpenAPI Specification**:
```yaml
paths:
  /deploy:
    post:
      operationId: deployApplication
      summary: Deploy application with validation
      x-cli-command: "deploy app"

      parameters:
        - name: app_id
          in: query
          required: true
          schema:
            type: string
          x-cli-flag: "--app-id"

        - name: version
          in: query
          required: true
          schema:
            type: string
          x-cli-flag: "--version"

      x-cli-workflow:
        steps:
          # Step 1: Validate deployment readiness
          - id: check-readiness
            description: "Checking deployment readiness..."
            request:
              method: GET
              url: "{base_url}/deployments/readiness"
            validation:
              condition: "check-readiness.body.ready == true"
              error-message: "System not ready for deployment"

          # Step 2: Create deployment
          - id: create-deployment
            description: "Creating deployment..."
            request:
              method: POST
              url: "{base_url}/deployments"
              body:
                app_id: "{args.app_id}"
                version: "{args.version}"
            condition: "check-readiness.body.ready == true"

          # Step 3: Wait for deployment to complete
          - id: wait-deployment
            description: "Waiting for deployment..."
            request:
              method: GET
              url: "{base_url}/deployments/{create-deployment.body.id}"
            polling:
              interval: 10
              timeout: 600
              terminal-condition: "response.body.status in ['success', 'failed']"

          # Step 4: Run health check
          - id: health-check
            description: "Running health check..."
            request:
              method: GET
              url: "{base_url}/health"
            condition: "wait-deployment.body.status == 'success'"

        output:
          format: json
          transform: |
            {
              "deployment_id": create-deployment.body.id,
              "status": wait-deployment.body.status,
              "health": health-check.body.healthy,
              "url": create-deployment.body.url
            }
```

**Usage**:
```bash
# Deploy application
mycli deploy app --app-id myapp --version v1.2.3

# Output:
# ✓ Checking deployment readiness... (200 OK)
# ✓ Creating deployment... (201 Created)
# ⏳ Waiting for deployment... (polling every 10s)
# ✓ Deployment complete (success)
# ✓ Running health check... (healthy)
#
# Deployment successful:
# {
#   "deployment_id": "dep-abc123",
#   "status": "success",
#   "health": true,
#   "url": "https://myapp.example.com"
# }
```

---

### Example 2: Data Migration Workflow

**OpenAPI Specification**:
```yaml
paths:
  /migrate:
    post:
      operationId: migrateData
      summary: Migrate data between environments
      x-cli-command: "migrate data"

      parameters:
        - name: source_env
          in: query
          required: true
          schema:
            type: string
          x-cli-flag: "--source"

        - name: target_env
          in: query
          required: true
          schema:
            type: string
          x-cli-flag: "--target"

      x-cli-workflow:
        steps:
          # Step 1: Export from source
          - id: export-data
            description: "Exporting from {args.source_env}..."
            request:
              method: POST
              url: "{base_url}/export"
              body:
                environment: "{args.source_env}"

          # Step 2: Validate export
          - id: validate-export
            description: "Validating export..."
            request:
              method: GET
              url: "{base_url}/export/{export-data.body.export_id}/validate"
            validation:
              condition: "validate-export.body.valid == true"
              error-message: "Export validation failed"

          # Step 3: Import to target
          - id: import-data
            description: "Importing to {args.target_env}..."
            request:
              method: POST
              url: "{base_url}/import"
              body:
                environment: "{args.target_env}"
                export_id: "{export-data.body.export_id}"

          # Step 4: Verify import
          - id: verify-import
            description: "Verifying import..."
            request:
              method: GET
              url: "{base_url}/import/{import-data.body.import_id}/verify"

        rollback:
          enabled: true
          steps:
            - description: "Rolling back import..."
              request:
                method: DELETE
                url: "{base_url}/import/{import-data.body.import_id}"

        output:
          format: table
          transform: |
            {
              "export_id": export-data.body.export_id,
              "import_id": import-data.body.import_id,
              "records_migrated": verify-import.body.record_count,
              "status": verify-import.body.status
            }
```

---

## CLI with Plugins

### Example: AWS S3 Backup Plugin

**OpenAPI Specification**:
```yaml
paths:
  /databases/{dbId}/backup:
    post:
      operationId: backupDatabase
      summary: Backup database to S3
      x-cli-command: "backup database"

      parameters:
        - name: dbId
          in: path
          required: true
          schema:
            type: string
          x-cli-flag: "--db-id"

        - name: bucket
          in: query
          required: true
          schema:
            type: string
          x-cli-flag: "--bucket"
          x-cli-env-var: "MYCLI_BACKUP_BUCKET"

      x-cli-plugin:
        type: external
        command: aws
        required: true
        min-version: "2.0.0"
        install-hint: "Install AWS CLI: https://aws.amazon.com/cli/"

        operations:
          # Step 1: Create database backup
          - description: "Creating database backup..."
            api-call:
              endpoint: "/databases/{dbId}/backup"
              method: POST
              output-var: backup

          # Step 2: Upload to S3
          - description: "Uploading to S3..."
            plugin-call:
              command: "aws"
              args:
                - "s3"
                - "cp"
                - "{backup.body.backup_file}"
                - "s3://{args.bucket}/backups/{dbId}/{backup.body.backup_id}.sql"
              env:
                AWS_REGION: "us-east-1"

          # Step 3: Verify upload
          - description: "Verifying upload..."
            plugin-call:
              command: "aws"
              args:
                - "s3"
                - "ls"
                - "s3://{args.bucket}/backups/{dbId}/{backup.body.backup_id}.sql"
```

**Usage**:
```bash
# Backup database to S3
mycli backup database --db-id prod-db --bucket my-backups

# Output:
# ✓ Creating database backup... (backup-abc123)
# ⏳ Uploading to S3... (aws s3 cp)
# ✓ Upload complete
# ✓ Verifying upload...
#
# Backup successful:
# Location: s3://my-backups/backups/prod-db/backup-abc123.sql
# Size: 1.2 GB
```

---

## Streaming and Watch Mode

### Example 1: Server-Sent Events (SSE)

**OpenAPI Specification**:
```yaml
paths:
  /builds/{buildId}/logs:
    get:
      operationId: streamBuildLogs
      summary: Stream build logs in real-time
      x-cli-command: "logs build"

      parameters:
        - name: buildId
          in: path
          required: true
          schema:
            type: string
          x-cli-flag: "--build-id"

      responses:
        '200':
          description: SSE stream
          content:
            text/event-stream:
              schema:
                type: string

      x-cli-streaming:
        enabled: true
        type: sse
        event-types:
          - log
          - error
          - complete
        format:
          template: "[{timestamp}] {message}"
          colors:
            log: white
            error: red
            complete: green
        reconnect:
          enabled: true
          max-retries: 5
          backoff: exponential
```

**Usage**:
```bash
# Stream build logs
mycli logs build --build-id abc123

# Output (streaming):
# [14:32:01] Starting build...
# [14:32:02] Installing dependencies...
# [14:32:15] Running tests...
# [14:32:20] ✓ All tests passed
# [14:32:21] Building application...
# [14:32:35] ✓ Build complete
```

---

### Example 2: Watch Mode

**OpenAPI Specification**:
```yaml
paths:
  /deployments/{deploymentId}:
    get:
      operationId: getDeployment
      summary: Get deployment status
      x-cli-command: "get deployment"

      parameters:
        - name: deploymentId
          in: path
          required: true
          schema:
            type: string
          x-cli-flag: "--deployment-id"

      x-cli-watch:
        enabled: true
        interval: 5
        fields: [status, progress, message]
        alert-on-change:
          - field: status
            message: "Status changed to {new_value}"
          - field: progress
            condition: "new_value >= 100"
            message: "Deployment complete!"
```

**Usage**:
```bash
# Watch deployment status
mycli get deployment --deployment-id dep-123 --watch

# Output (updates every 5s):
# Status: deploying
# Progress: 25%
# Message: Building containers...
#
# Status: deploying
# Progress: 50%
# Message: Pushing to registry...
#
# Status: deploying
# Progress: 75%
# Message: Updating services...
#
# 🔔 Status changed to completed
# Status: completed
# Progress: 100%
# Message: Deployment successful
# 🔔 Deployment complete!
```

---

## Interactive Mode

### Example: Interactive Resource Creation

**OpenAPI Specification**:
```yaml
paths:
  /servers:
    post:
      operationId: createServer
      summary: Create a new server
      x-cli-command: "create server"

      requestBody:
        content:
          application/json:
            schema:
              type: object
              properties:
                name:
                  type: string
                region:
                  type: string
                size:
                  type: string
                image:
                  type: string

      x-cli-flags:
        - name: name
          source: name
          flag: "--name"
          type: string
        - name: region
          source: region
          flag: "--region"
          type: string
        - name: size
          source: size
          flag: "--size"
          type: string
        - name: image
          source: image
          flag: "--image"
          type: string

      x-cli-interactive:
        enabled: true
        prompts:
          # Text input
          - parameter: name
            type: text
            message: "Server name:"
            validation: "^[a-z][a-z0-9-]{0,63}$"
            validation-message: "Must start with letter, lowercase, max 64 chars"

          # Dynamic select (fetch from API)
          - parameter: region
            type: select
            message: "Select region:"
            source:
              endpoint: "/regions"
              value-field: "id"
              display-field: "name"

          # Static select
          - parameter: size
            type: select
            message: "Server size:"
            choices:
              - value: small
                label: "Small (1 CPU, 2GB RAM) - $10/month"
              - value: medium
                label: "Medium (2 CPU, 4GB RAM) - $20/month"
              - value: large
                label: "Large (4 CPU, 8GB RAM) - $40/month"

          # Multi-select
          - parameter: image
            type: select
            message: "Select OS image:"
            choices:
              - value: ubuntu-22.04
                label: "Ubuntu 22.04 LTS"
              - value: debian-12
                label: "Debian 12"
              - value: centos-9
                label: "CentOS 9"

          # Confirmation
          - parameter: confirm
            type: confirm
            message: "Create server with these settings?"
            default: false
```

**Usage**:
```bash
# Interactive mode
mycli create server --interactive

# Prompts:
# ? Server name: my-web-server
# ? Select region: us-east-1
# ? Server size: Medium (2 CPU, 4GB RAM) - $20/month
# ? Select OS image: Ubuntu 22.04 LTS
# ? Create server with these settings? Yes
#
# ✓ Server created: srv-abc123
```

---

## Context Switching

### Example: Multi-Environment CLI

**CLI Configuration**:
```yaml
api:
  environments:
    - name: production
      openapi_url: https://api.example.com/openapi.yaml
      base_url: https://api.example.com
      default: true

    - name: staging
      openapi_url: https://staging-api.example.com/openapi.yaml
      base_url: https://staging-api.example.com

    - name: development
      openapi_url: http://localhost:8080/openapi.yaml
      base_url: http://localhost:8080
```

**OpenAPI Specification**:
```yaml
paths:
  /contexts:
    get:
      operationId: listContexts
      summary: List available contexts
      x-cli-command: "list contexts"

      x-cli-context:
        enabled: true
        default: "production"
        contexts:
          production:
            base-url: "https://api.example.com"
            auth:
              type: oauth2
          staging:
            base-url: "https://staging-api.example.com"
            auth:
              type: oauth2
          development:
            base-url: "http://localhost:8080"
            auth:
              type: none

        switch-command: "use context"
        current-command: "current context"
```

**Usage**:
```bash
# List available contexts
mycli list contexts

# Output:
# NAME         BASE URL                           DEFAULT
# production   https://api.example.com            *
# staging      https://staging-api.example.com
# development  http://localhost:8080

# Switch context
mycli use context staging

# Check current context
mycli current context
# Output: staging

# Use context for single command
mycli --context development users list

# Switch back to default
mycli use context production
```

---

## Custom Output Formats

### Example 1: Table Output with Colors

**OpenAPI Specification**:
```yaml
paths:
  /servers:
    get:
      operationId: listServers
      x-cli-command: "list servers"

      x-cli-output:
        table:
          columns:
            - field: id
              header: ID
              width: 12
              align: left

            - field: name
              header: NAME
              width: 25

            - field: status
              header: STATUS
              width: 12
              transform: uppercase
              color-map:
                running: green
                stopped: red
                starting: yellow

            - field: cpu_usage
              header: CPU
              width: 8
              align: right
              suffix: "%"
              format: "%.1f"

            - field: memory_usage
              header: MEMORY
              width: 10
              align: right
              format: "%.0f"
              suffix: "MB"

            - field: ip_address
              header: IP
              width: 15

            - field: created_at
              header: CREATED
              width: 12
              transform: date

          sort-by: "name"
          sort-order: "asc"
```

**Usage**:
```bash
mycli list servers --output table

# Output:
# ID           NAME                      STATUS      CPU     MEMORY    IP              CREATED
# srv-abc123   web-server-1              RUNNING     45.2%   1024MB    192.168.1.10    2024-11-01
# srv-def456   api-server-1              RUNNING     23.1%   2048MB    192.168.1.11    2024-11-05
# srv-ghi789   worker-1                  STOPPED     0.0%    0MB       192.168.1.12    2024-11-10
```

---

### Example 2: Custom Template Output

**OpenAPI Specification**:
```yaml
paths:
  /servers/{serverId}:
    get:
      operationId: getServer
      x-cli-command: "get server"

      x-cli-output:
        format: custom
        template: |
          ╔════════════════════════════════════════════════════╗
          ║              Server: {{.name}}
          ╠════════════════════════════════════════════════════╣
          ║
          ║  ID:           {{.id}}
          ║  Status:       {{.status | upper}}
          ║  Region:       {{.region}}
          ║  IP Address:   {{.ip_address}}
          ║
          ║  Resources:
          ║    CPU:        {{.cpu_usage}}%
          ║    Memory:     {{.memory_usage}}MB / {{.memory_total}}MB
          ║    Disk:       {{.disk_usage}}GB / {{.disk_total}}GB
          ║
          ║  Network:
          ║    Bandwidth:  {{.bandwidth_in | bytes}} in / {{.bandwidth_out | bytes}} out
          ║    Requests:   {{.requests_per_second}} req/s
          ║
          ║  Created:      {{.created_at | date}}
          ║  Updated:      {{.updated_at | timeAgo}}
          ║
          ╚════════════════════════════════════════════════════╝
```

---

## Migration Examples

### Migrating from OpenAPI Generator

**Before** (OpenAPI Generator):
```bash
# Generate client
openapi-generator-cli generate \
  -i openapi.yaml \
  -g go \
  -o ./client

# Use in code
import "myclient"
client := myclient.NewAPIClient()
response, err := client.UsersApi.ListUsers(ctx)
```

**After** (CliForge):
```bash
# Generate CLI
cliforge build --config cli-config.yaml

# Use from command line
mycli users list
```

**Advantages**:
- No code compilation required
- Self-updating binary
- Built-in output formatting
- Authentication built-in
- User-friendly interface

---

### Migrating from cURL Scripts

**Before** (bash + cURL):
```bash
#!/bin/bash
API_KEY="sk_live_abc123"
BASE_URL="https://api.example.com"

# List users
curl -H "X-API-Key: $API_KEY" \
     "$BASE_URL/users" \
     | jq '.data[] | {id, name, email}'

# Create user
curl -X POST \
     -H "X-API-Key: $API_KEY" \
     -H "Content-Type: application/json" \
     -d '{"name":"John","email":"john@example.com"}' \
     "$BASE_URL/users"
```

**After** (CliForge):
```bash
# List users
mycli users list --output table

# Create user
mycli create user --name John --email john@example.com
```

---

## Real-World Use Cases

### Use Case 1: Cloud Infrastructure Management

**Scenario**: Manage cloud resources (servers, databases, networks) across multiple regions.

**Features Used**:
- Multi-environment support
- Interactive server creation
- Watch mode for deployment status
- Workflow for backup operations
- Plugin integration for cloud provider CLI

**Example Commands**:
```bash
# List servers in all regions
mycli list servers

# Create server (interactive)
mycli create server --interactive

# Deploy application
mycli deploy app --app-id myapp --version v1.2.3

# Watch deployment
mycli get deployment --deployment-id dep-123 --watch

# Backup database to S3
mycli backup database --db-id prod-db --bucket backups
```

---

### Use Case 2: CI/CD Pipeline Integration

**Scenario**: Trigger builds, deployments, and monitor status in CI/CD pipelines.

**Features Used**:
- Async polling for build status
- Streaming logs
- Exit codes for success/failure
- JSON output for parsing

**Example Pipeline** (GitHub Actions):
```yaml
name: Deploy

on:
  push:
    branches: [main]

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Install CLI
        run: |
          curl -L https://releases.example.com/install.sh | sh

      - name: Trigger deployment
        env:
          MYCLI_API_KEY: ${{ secrets.API_KEY }}
        run: |
          DEPLOYMENT_ID=$(mycli deploy app \
            --app-id myapp \
            --version ${{ github.sha }} \
            --output json | jq -r '.deployment_id')

          echo "deployment_id=$DEPLOYMENT_ID" >> $GITHUB_OUTPUT

      - name: Wait for deployment
        run: |
          mycli get deployment \
            --deployment-id ${{ steps.deploy.outputs.deployment_id }} \
            --watch \
            --timeout 600
```

---

### Use Case 3: Database Administration

**Scenario**: Manage databases, run migrations, create backups.

**Features Used**:
- Workflow for migration
- Plugin integration for backup tools
- Confirmation prompts for destructive operations
- Streaming logs for long-running operations

**Example Commands**:
```bash
# List databases
mycli list databases

# Create database
mycli create database --name mydb --region us-east-1

# Run migration
mycli migrate database \
  --db-id mydb \
  --migration-file schema.sql

# Create backup
mycli backup database --db-id mydb

# Restore from backup
mycli restore database \
  --db-id mydb \
  --backup-id backup-abc123
```

---

### Use Case 4: API Testing and Debugging

**Scenario**: Test API endpoints, inspect responses, debug issues.

**Features Used**:
- JSON/YAML output
- Verbose mode for request details
- Dry-run mode
- Custom headers

**Example Commands**:
```bash
# Test endpoint
mycli users list --verbose

# Dry-run (show request without executing)
mycli create user \
  --name John \
  --email john@example.com \
  --dry-run

# Debug mode (show full request/response)
mycli users list --debug

# Custom headers
mycli users list \
  --header "X-Request-ID: test-123" \
  --header "X-Custom: value"
```

---

## Best Practices

### 1. Use Aliases for Common Commands

```yaml
x-cli-command: "list users"
x-cli-aliases: ["users", "ls users", "get users"]
```

### 2. Provide Interactive Mode for Complex Operations

```yaml
x-cli-interactive:
  enabled: true
  prompts:
    - parameter: name
      type: text
    - parameter: region
      type: select
```

### 3. Use Confirmation for Destructive Operations

```yaml
x-cli-confirmation:
  enabled: true
  message: "Delete {count} resources?"
  flag: "--yes"
```

### 4. Enable Watch Mode for Long-Running Operations

```yaml
x-cli-watch:
  enabled: true
  interval: 5
  fields: [status, progress]
```

### 5. Use Workflows for Multi-Step Operations

```yaml
x-cli-workflow:
  steps:
    - id: validate
      ...
    - id: execute
      ...
    - id: verify
      ...
```

### 6. Provide Good Output Formatting

```yaml
x-cli-output:
  table:
    columns:
      - field: status
        color-map:
          success: green
          error: red
```

---

## Next Steps

- **[Configuration DSL Reference](configuration-dsl.md)** - Full configuration options
- **[Technical Specification](technical-specification.md)** - Architecture and design
- **[FAQ](faq.md)** - Common questions
- **[Troubleshooting](troubleshooting.md)** - Common issues and solutions

---

**Version**: 0.9.0
**Last Updated**: 2025-11-23
