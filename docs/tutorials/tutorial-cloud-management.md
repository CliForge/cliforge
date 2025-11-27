# Tutorial: Cloud Infrastructure Management CLI

**Difficulty**: Intermediate
**Time**: 90-120 minutes
**Version**: 1.0.0

---

## Overview

In this tutorial, you'll build a production-ready CLI for managing cloud infrastructure, similar to AWS CLI or Azure CLI. You'll learn how to:

- Handle long-running asynchronous operations
- Implement workflow orchestration for complex deployments
- Manage state across multiple resources
- Provide progress feedback and streaming output
- Handle resource dependencies and rollback scenarios
- Implement advanced error handling and retries

By the end, you'll have a working `cloud-cli` tool for managing virtual machines, storage, and networking.

### What You'll Build

A comprehensive cloud management CLI with these capabilities:

```bash
# Resource Management
cloud-cli compute instances create --name web-server --size large --region us-west
cloud-cli compute instances list --region us-west
cloud-cli compute instances start --id i-abc123
cloud-cli storage buckets create --name my-data --encryption aes256

# Async Operations
cloud-cli compute instances delete --id i-abc123 --wait
cloud-cli storage sync --source ./local --destination s3://bucket --watch

# Workflows
cloud-cli deploy full-stack --template ./infra.yaml --env production
cloud-cli rollback deployment --id dep-123 --reason "High error rate"

# State Management
cloud-cli state list
cloud-cli state refresh
cloud-cli state export --format terraform
```

---

## Prerequisites

### Required Knowledge

- Intermediate command-line proficiency
- Understanding of cloud infrastructure concepts (VMs, storage, networking)
- Familiarity with asynchronous operations and state management
- Experience with infrastructure-as-code (helpful but not required)
- Basic understanding of YAML and JSON

### Required Software

- **Go**: Version 1.21 or later
- **CliForge**: Version 0.9.0 or later
- **Docker**: For local testing (optional)
- **Git**: For version control
- **jq**: For JSON processing

### Recommended Reading

- [CliForge Workflows Guide](../user-guide-workflows.md)

> **Note**: For detailed architecture information on state management, progress tracking, and streaming operations, please refer to the internal design documentation or contact the development team.

---

## Architecture Overview

### System Design

```
┌─────────────────────────────────────────────────────────┐
│                     Cloud CLI                           │
├─────────────────────────────────────────────────────────┤
│  Command Layer                                          │
│  ├── Compute (instances, images, snapshots)            │
│  ├── Storage (buckets, volumes)                        │
│  ├── Network (vpcs, subnets, security groups)          │
│  └── Workflows (deploy, rollback, scale)               │
├─────────────────────────────────────────────────────────┤
│  Orchestration Layer                                    │
│  ├── Workflow Engine                                    │
│  ├── State Manager                                      │
│  ├── Dependency Resolver                                │
│  └── Rollback Handler                                   │
├─────────────────────────────────────────────────────────┤
│  API Layer                                              │
│  ├── Async Operation Tracker                           │
│  ├── Progress Reporter                                  │
│  ├── Retry Logic                                        │
│  └── Rate Limiter                                       │
└─────────────────────────────────────────────────────────┘
                         ↓
┌─────────────────────────────────────────────────────────┐
│              Cloud Provider API                         │
│  (Simulated for this tutorial)                          │
└─────────────────────────────────────────────────────────┘
```

### Key Concepts

**Async Operations**: Long-running operations that return immediately with an operation ID, which can be polled for status.

**Workflows**: Multi-step processes that orchestrate multiple API calls with dependency management.

**State Management**: Tracking the current state of infrastructure and detecting drift from desired state.

**Idempotency**: Operations that can be safely retried without causing duplicate effects.

---

## Step 1: Project Setup

### Create Project Structure

```bash
# Create project directory
mkdir cloud-cli-tutorial
cd cloud-cli-tutorial

# Create comprehensive directory structure
mkdir -p {specs,config,workflows,state,scripts,tests}

# Initialize git
git init
echo "*.log" >> .gitignore
echo ".cloud-cli/" >> .gitignore
echo "state/*.state" >> .gitignore
```

Your structure:

```
cloud-cli-tutorial/
├── specs/              # OpenAPI specifications
├── config/             # CLI configuration
├── workflows/          # Workflow definitions
├── state/              # State management
├── scripts/            # Helper scripts
└── tests/              # Test files
```

### Initialize Go Module

```bash
# Initialize Go module
go mod init github.com/yourorg/cloud-cli

# Install dependencies
go get github.com/spf13/cobra@latest
go get github.com/spf13/viper@latest
```

---

## Step 2: OpenAPI Specification for Compute Service

Create `specs/compute-api.yaml`:

```yaml
openapi: 3.0.3
info:
  title: Cloud Compute API
  description: Virtual machine and compute resource management
  version: 2.0.0

servers:
  - url: https://api.cloudprovider.com/v2
    description: Production API

security:
  - ApiKeyAuth: []

components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key
      x-cli-auth:
        env_var: "CLOUD_API_KEY"
        config_key: "cloud.api_key"

  schemas:
    Instance:
      type: object
      properties:
        id:
          type: string
          description: Instance ID
        name:
          type: string
          description: Instance name
        size:
          type: string
          description: Instance size (small, medium, large, xlarge)
        region:
          type: string
          description: Deployment region
        state:
          type: string
          enum: [pending, running, stopping, stopped, terminated]
          description: Current instance state
        public_ip:
          type: string
          description: Public IP address
        private_ip:
          type: string
          description: Private IP address
        created_at:
          type: string
          format: date-time
        tags:
          type: object
          additionalProperties:
            type: string

    CreateInstanceRequest:
      type: object
      required:
        - name
        - size
        - region
      properties:
        name:
          type: string
          pattern: "^[a-z0-9-]+$"
          minLength: 3
          maxLength: 63
        size:
          type: string
          enum: [small, medium, large, xlarge]
        region:
          type: string
          enum: [us-east, us-west, eu-central, ap-southeast]
        image_id:
          type: string
          default: "ubuntu-22.04"
        ssh_keys:
          type: array
          items:
            type: string
        tags:
          type: object
          additionalProperties:
            type: string
        user_data:
          type: string
          description: Cloud-init script

    Operation:
      type: object
      properties:
        id:
          type: string
          description: Operation ID
        type:
          type: string
          description: Operation type
        status:
          type: string
          enum: [pending, running, succeeded, failed, cancelled]
        resource_id:
          type: string
          description: Target resource ID
        progress:
          type: integer
          minimum: 0
          maximum: 100
        started_at:
          type: string
          format: date-time
        completed_at:
          type: string
          format: date-time
        error:
          type: object
          properties:
            code:
              type: string
            message:
              type: string

    Error:
      type: object
      properties:
        code:
          type: string
        message:
          type: string
        details:
          type: object

paths:
  /compute/instances:
    get:
      summary: List instances
      operationId: listInstances
      x-cli-command: "compute instances list"
      x-cli-aliases: ["list instances", "instances ls"]
      x-cli-description: "List compute instances"
      x-cli-output:
        table:
          columns:
            - field: id
              header: ID
              width: 15
            - field: name
              header: NAME
              width: 25
            - field: size
              header: SIZE
              width: 10
            - field: region
              header: REGION
              width: 12
            - field: state
              header: STATE
              width: 12
              transform: |
                value == 'running' ? '🟢 RUNNING' :
                value == 'stopped' ? '⚪ STOPPED' :
                value == 'pending' ? '🟡 PENDING' :
                '🔴 ' + value.toUpperCase()
            - field: public_ip
              header: PUBLIC IP
              width: 15
      parameters:
        - name: region
          in: query
          schema:
            type: string
          x-cli-flag:
            name: "--region"
            description: "Filter by region"
        - name: state
          in: query
          schema:
            type: string
          x-cli-flag:
            name: "--state"
            description: "Filter by state"
        - name: tag
          in: query
          schema:
            type: string
          x-cli-flag:
            name: "--tag"
            description: "Filter by tag (key=value)"
      responses:
        '200':
          description: List of instances
          content:
            application/json:
              schema:
                type: object
                properties:
                  instances:
                    type: array
                    items:
                      $ref: '#/components/schemas/Instance'
                  total:
                    type: integer

    post:
      summary: Create instance
      operationId: createInstance
      x-cli-command: "compute instances create"
      x-cli-aliases: ["create instance"]
      x-cli-description: "Create a new compute instance"
      x-cli-async:
        enabled: true
        poll_interval: 5
        timeout: 600
        operation_field: "operation_id"
      x-cli-flags:
        - name: name
          source: name
          flag: "--name"
          required: true
        - name: size
          source: size
          flag: "--size"
          required: true
        - name: region
          source: region
          flag: "--region"
          required: true
        - name: image_id
          source: image_id
          flag: "--image"
        - name: ssh_keys
          source: ssh_keys
          flag: "--ssh-key"
          type: array
        - name: tags
          source: tags
          flag: "--tag"
          type: map
        - name: wait
          flag: "--wait"
          type: boolean
          description: "Wait for operation to complete"
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateInstanceRequest'
      responses:
        '202':
          description: Instance creation started
          content:
            application/json:
              schema:
                type: object
                properties:
                  instance:
                    $ref: '#/components/schemas/Instance'
                  operation_id:
                    type: string
          x-cli-output:
            success_message: "Instance creation started"
            show_operation_id: true

  /compute/instances/{instance_id}:
    parameters:
      - name: instance_id
        in: path
        required: true
        schema:
          type: string
        x-cli-flag:
          name: "--id"
          description: "Instance ID"

    get:
      summary: Get instance
      operationId: getInstance
      x-cli-command: "compute instances get"
      x-cli-description: "Get instance details"
      responses:
        '200':
          description: Instance details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Instance'

    delete:
      summary: Delete instance
      operationId: deleteInstance
      x-cli-command: "compute instances delete"
      x-cli-aliases: ["terminate instance"]
      x-cli-description: "Delete a compute instance"
      x-cli-async:
        enabled: true
        poll_interval: 5
        timeout: 300
      x-cli-confirmation:
        enabled: true
        message: "Delete instance {instance_id}?"
        destructive: true
      x-cli-flags:
        - name: wait
          flag: "--wait"
          type: boolean
          description: "Wait for deletion to complete"
        - name: force
          flag: "--force"
          type: boolean
          description: "Force delete without confirmation"
      responses:
        '202':
          description: Deletion started
          content:
            application/json:
              schema:
                type: object
                properties:
                  operation_id:
                    type: string

  /compute/instances/{instance_id}/start:
    parameters:
      - name: instance_id
        in: path
        required: true
        schema:
          type: string
        x-cli-flag:
          name: "--id"

    post:
      summary: Start instance
      operationId: startInstance
      x-cli-command: "compute instances start"
      x-cli-description: "Start a stopped instance"
      x-cli-async:
        enabled: true
        poll_interval: 3
        timeout: 180
      responses:
        '202':
          description: Instance starting
          content:
            application/json:
              schema:
                type: object
                properties:
                  operation_id:
                    type: string

  /compute/instances/{instance_id}/stop:
    parameters:
      - name: instance_id
        in: path
        required: true
        schema:
          type: string
        x-cli-flag:
          name: "--id"

    post:
      summary: Stop instance
      operationId: stopInstance
      x-cli-command: "compute instances stop"
      x-cli-description: "Stop a running instance"
      x-cli-async:
        enabled: true
        poll_interval: 3
        timeout: 180
      x-cli-flags:
        - name: force
          flag: "--force"
          type: boolean
          description: "Force stop (power off)"
      responses:
        '202':
          description: Instance stopping
          content:
            application/json:
              schema:
                type: object
                properties:
                  operation_id:
                    type: string

  /operations/{operation_id}:
    parameters:
      - name: operation_id
        in: path
        required: true
        schema:
          type: string

    get:
      summary: Get operation status
      operationId: getOperation
      x-cli-command: "operations get"
      x-cli-description: "Get operation status"
      x-cli-output:
        show_progress: true
      responses:
        '200':
          description: Operation status
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Operation'
```

---

## Step 3: Storage API Specification

Create `specs/storage-api.yaml`:

```yaml
openapi: 3.0.3
info:
  title: Cloud Storage API
  description: Object storage and volume management
  version: 2.0.0

servers:
  - url: https://api.cloudprovider.com/v2

security:
  - ApiKeyAuth: []

components:
  securitySchemes:
    ApiKeyAuth:
      type: apiKey
      in: header
      name: X-API-Key

  schemas:
    Bucket:
      type: object
      properties:
        id:
          type: string
        name:
          type: string
        region:
          type: string
        encryption:
          type: string
          enum: [none, aes256, kms]
        versioning:
          type: boolean
        size_bytes:
          type: integer
        object_count:
          type: integer
        created_at:
          type: string
          format: date-time

    CreateBucketRequest:
      type: object
      required:
        - name
      properties:
        name:
          type: string
          pattern: "^[a-z0-9.-]+$"
        region:
          type: string
        encryption:
          type: string
          enum: [none, aes256, kms]
          default: aes256
        versioning:
          type: boolean
          default: false

    Object:
      type: object
      properties:
        key:
          type: string
        size_bytes:
          type: integer
        etag:
          type: string
        last_modified:
          type: string
          format: date-time
        storage_class:
          type: string

    SyncOperation:
      type: object
      properties:
        id:
          type: string
        source:
          type: string
        destination:
          type: string
        status:
          type: string
        files_synced:
          type: integer
        bytes_transferred:
          type: integer
        errors:
          type: array
          items:
            type: object

paths:
  /storage/buckets:
    get:
      summary: List buckets
      operationId: listBuckets
      x-cli-command: "storage buckets list"
      x-cli-description: "List storage buckets"
      x-cli-output:
        table:
          columns:
            - field: name
              header: NAME
              width: 30
            - field: region
              header: REGION
              width: 15
            - field: encryption
              header: ENCRYPTION
              width: 12
            - field: object_count
              header: OBJECTS
              width: 10
            - field: size_bytes
              header: SIZE
              width: 12
              transform: "formatBytes(value)"
      responses:
        '200':
          description: List of buckets
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Bucket'

    post:
      summary: Create bucket
      operationId: createBucket
      x-cli-command: "storage buckets create"
      x-cli-description: "Create a storage bucket"
      x-cli-flags:
        - name: name
          source: name
          flag: "--name"
          required: true
        - name: region
          source: region
          flag: "--region"
        - name: encryption
          source: encryption
          flag: "--encryption"
        - name: versioning
          source: versioning
          flag: "--versioning"
          type: boolean
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateBucketRequest'
      responses:
        '201':
          description: Bucket created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Bucket'

  /storage/buckets/{bucket_name}/objects:
    parameters:
      - name: bucket_name
        in: path
        required: true
        schema:
          type: string

    get:
      summary: List objects
      operationId: listObjects
      x-cli-command: "storage objects list"
      x-cli-description: "List objects in bucket"
      x-cli-flags:
        - name: bucket
          source: bucket_name
          flag: "--bucket"
          required: true
        - name: prefix
          flag: "--prefix"
          description: "Filter by prefix"
      parameters:
        - name: prefix
          in: query
          schema:
            type: string
        - name: max_keys
          in: query
          schema:
            type: integer
            default: 1000
      responses:
        '200':
          description: List of objects
          content:
            application/json:
              schema:
                type: object
                properties:
                  objects:
                    type: array
                    items:
                      $ref: '#/components/schemas/Object'

  /storage/sync:
    post:
      summary: Sync files
      operationId: syncFiles
      x-cli-command: "storage sync"
      x-cli-description: "Sync files between local and cloud storage"
      x-cli-streaming:
        enabled: true
        event_types:
          - file_started
          - file_completed
          - error
      x-cli-flags:
        - name: source
          flag: "--source"
          required: true
          description: "Source path (local or s3://)"
        - name: destination
          flag: "--destination"
          required: true
          description: "Destination path"
        - name: delete
          flag: "--delete"
          type: boolean
          description: "Delete files not in source"
        - name: watch
          flag: "--watch"
          type: boolean
          description: "Continuous sync mode"
      requestBody:
        required: true
        content:
          application/json:
            schema:
              type: object
              properties:
                source:
                  type: string
                destination:
                  type: string
                delete:
                  type: boolean
                watch:
                  type: boolean
      responses:
        '200':
          description: Sync operation started
          content:
            text/event-stream:
              schema:
                $ref: '#/components/schemas/SyncOperation'
```

---

## Step 4: CLI Configuration with Advanced Features

Create `config/cli-config.yaml`:

```yaml
metadata:
  name: cloud-cli
  version: 2.0.0
  description: Cloud Infrastructure Management CLI
  author: DevOps Team
  license: Apache-2.0

api:
  specs:
    - path: ../specs/compute-api.yaml
      prefix: compute
    - path: ../specs/storage-api.yaml
      prefix: storage

  base_url: https://api.cloudprovider.com/v2

  default_headers:
    User-Agent: "cloud-cli/2.0.0"
    Accept: "application/json"

  timeout:
    default: 30s
    long_running: 600s

authentication:
  default_method: api_key
  methods:
    api_key:
      type: apiKey
      header: "X-API-Key"
      env_var: "CLOUD_API_KEY"
      config_key: "cloud.api_key"

# Async Operations Configuration
async_operations:
  enabled: true
  polling:
    default_interval: 5s
    max_interval: 30s
    backoff: exponential
    jitter: true
  timeout:
    default: 600s
    operations:
      createInstance: 900s
      deleteInstance: 600s
      startInstance: 300s
  progress:
    show_spinner: true
    show_percentage: true
    show_elapsed: true
    show_eta: true

# Workflow Configuration
workflows:
  enabled: true
  directory: ../workflows
  state_file: ../state/workflows.state

  features:
    parallel_execution: true
    dependency_resolution: true
    rollback_on_failure: true
    dry_run: true

  limits:
    max_concurrent_operations: 5
    max_retry_attempts: 3
    operation_timeout: 3600s

# State Management
state:
  enabled: true
  backend: local
  path: ../state/terraform.state

  features:
    auto_refresh: true
    lock_timeout: 300s
    backup: true
    backup_count: 5

  drift_detection:
    enabled: true
    check_interval: 300s

# Error Handling
errors:
  retry:
    enabled: true
    max_attempts: 3
    initial_delay: 1s
    max_delay: 30s
    backoff: exponential
    jitter: true

    retryable_codes:
      - 408  # Request Timeout
      - 429  # Too Many Requests
      - 500  # Internal Server Error
      - 502  # Bad Gateway
      - 503  # Service Unavailable
      - 504  # Gateway Timeout

    retryable_operations:
      - listInstances
      - getInstance
      - getOperation

  circuit_breaker:
    enabled: true
    failure_threshold: 5
    timeout: 60s
    half_open_requests: 3

# Rate Limiting
rate_limit:
  enabled: true
  requests_per_second: 10
  burst: 20
  show_remaining: true

# Output Configuration
output:
  default_format: table
  formats:
    - table
    - json
    - yaml
    - terraform

  table:
    borders: true
    colors: true
    max_width: 120
    truncate: true

  streaming:
    enabled: true
    buffer_size: 4096
    flush_interval: 100ms

  progress:
    style: bar  # bar, spinner, dots
    width: 40
    show_percentage: true
    show_count: true

# Logging
logging:
  level: info
  format: text  # text, json
  output: stderr

  file:
    enabled: true
    path: ~/.cloud-cli/logs/cli.log
    max_size: 50MB
    max_age: 30d
    max_backups: 5
    compress: true

  debug:
    show_http_requests: true
    show_http_responses: false
    show_headers: false

# Cache
cache:
  enabled: true
  directory: ~/.cloud-cli/cache

  ttl:
    default: 300s
    operations:
      listInstances: 60s
      listBuckets: 300s
      getInstance: 30s

  size_limit: 100MB
  eviction_policy: lru

# Branding
branding:
  tagline: "Enterprise Cloud Infrastructure Management"

  colors:
    primary: "#0066cc"
    secondary: "#00aaff"
    success: "#00cc66"
    error: "#cc0000"
    warning: "#ff9900"

  help:
    examples:
      - description: "Create a compute instance"
        command: "cloud-cli compute instances create --name web-1 --size large --region us-west"
      - description: "Deploy full infrastructure"
        command: "cloud-cli deploy full-stack --template ./infra.yaml --env production"
      - description: "Sync files to cloud storage"
        command: "cloud-cli storage sync --source ./data --destination s3://my-bucket --watch"

# Telemetry
telemetry:
  enabled: false
  endpoint: https://telemetry.cloudprovider.com
  sample_rate: 0.1
```

---

## Step 5: Workflow Definitions

Create comprehensive workflow definitions for complex operations.

Create `workflows/full-stack-deploy.yaml`:

```yaml
name: full-stack-deploy
description: Deploy complete application infrastructure
version: 1.0.0

parameters:
  - name: env
    type: string
    required: true
    description: "Environment (dev, staging, production)"
    validation:
      enum: [dev, staging, production]

  - name: region
    type: string
    default: "us-west"
    description: "Deployment region"

  - name: instance_count
    type: integer
    default: 3
    description: "Number of application instances"
    validation:
      min: 1
      max: 10

  - name: enable_monitoring
    type: boolean
    default: true
    description: "Enable monitoring and logging"

variables:
  app_name: "myapp-{{ .env }}"
  db_name: "{{ .app_name }}-db"
  storage_bucket: "{{ .app_name }}-data"

steps:
  # Step 1: Create Storage Bucket
  - name: create_storage
    operation: createBucket
    description: "Create application data bucket"
    params:
      name: "{{ .storage_bucket }}"
      region: "{{ .region }}"
      encryption: "aes256"
      versioning: true

    output:
      bucket_id: "{{ .response.id }}"

    on_error:
      action: fail
      rollback: true

  # Step 2: Create Database Instance
  - name: create_database
    operation: createInstance
    description: "Create database server"
    depends_on:
      - create_storage

    params:
      name: "{{ .db_name }}"
      size: "{{ .env == 'production' ? 'xlarge' : 'medium' }}"
      region: "{{ .region }}"
      image_id: "postgres-14"
      tags:
        Environment: "{{ .env }}"
        Role: "database"
        ManagedBy: "cloud-cli"

    async:
      wait: true
      timeout: 900s

    output:
      db_instance_id: "{{ .response.instance.id }}"
      db_private_ip: "{{ .response.instance.private_ip }}"

    validation:
      - field: "response.instance.state"
        equals: "running"

    on_error:
      action: retry
      max_attempts: 3
      rollback: true

  # Step 3: Create Application Instances
  - name: create_app_instances
    operation: createInstance
    description: "Create application servers"
    depends_on:
      - create_database

    loop:
      count: "{{ .instance_count }}"
      parallel: true
      max_parallel: 3

    params:
      name: "{{ .app_name }}-app-{{ .loop_index }}"
      size: "large"
      region: "{{ .region }}"
      image_id: "ubuntu-22.04"
      user_data: |
        #!/bin/bash
        export DB_HOST={{ .db_private_ip }}
        export STORAGE_BUCKET={{ .storage_bucket }}
        # Additional setup commands...
      tags:
        Environment: "{{ .env }}"
        Role: "application"
        Index: "{{ .loop_index }}"

    async:
      wait: true
      timeout: 600s

    output:
      app_instance_ids: "{{ append .app_instance_ids .response.instance.id }}"

    on_error:
      action: continue_on_error
      max_failures: 1

  # Step 4: Configure Load Balancer (conditional)
  - name: configure_lb
    operation: createLoadBalancer
    description: "Setup load balancer for app instances"
    depends_on:
      - create_app_instances

    condition: "{{ .instance_count > 1 }}"

    params:
      name: "{{ .app_name }}-lb"
      region: "{{ .region }}"
      targets: "{{ .app_instance_ids }}"
      health_check:
        path: "/health"
        interval: 30

    output:
      lb_public_ip: "{{ .response.public_ip }}"

  # Step 5: Setup Monitoring (conditional)
  - name: setup_monitoring
    operation: createMonitoring
    description: "Configure monitoring and alerts"
    depends_on:
      - create_app_instances
      - configure_lb

    condition: "{{ .enable_monitoring }}"

    params:
      targets: "{{ .app_instance_ids }}"
      metrics:
        - cpu_usage
        - memory_usage
        - disk_usage
        - network_io
      alerts:
        - name: "high_cpu"
          metric: "cpu_usage"
          threshold: 80
          duration: 300

# Post-deployment validation
validation:
  - name: verify_database
    check: "getInstance"
    params:
      instance_id: "{{ .db_instance_id }}"
    expect:
      state: "running"

  - name: verify_app_instances
    check: "getInstance"
    loop: "{{ .app_instance_ids }}"
    expect:
      state: "running"

# Rollback strategy
rollback:
  enabled: true
  on_failure: true

  steps:
    - name: cleanup_instances
      operation: deleteInstance
      params:
        instance_id: "{{ .app_instance_ids }}"
      continue_on_error: true

    - name: cleanup_database
      operation: deleteInstance
      params:
        instance_id: "{{ .db_instance_id }}"
      continue_on_error: true

    - name: cleanup_storage
      operation: deleteBucket
      params:
        bucket_id: "{{ .bucket_id }}"
      continue_on_error: true

# Output for users
output:
  format: yaml
  template: |
    Deployment: {{ .app_name }}
    Environment: {{ .env }}
    Region: {{ .region }}

    Database:
      Instance ID: {{ .db_instance_id }}
      Private IP: {{ .db_private_ip }}

    Application Instances: {{ .instance_count }}
    {{ range .app_instance_ids }}
      - {{ . }}
    {{ end }}

    {{ if .lb_public_ip }}
    Load Balancer: {{ .lb_public_ip }}
    Application URL: http://{{ .lb_public_ip }}
    {{ end }}

    Monitoring: {{ .enable_monitoring }}
```

---

## Step 6: Building the CLI

### Build and Install

```bash
# Build the CLI
cliforge build \
  --config config/cli-config.yaml \
  --output ./cloud-cli \
  --platform $(go env GOOS)/$(go env GOARCH)

# Install globally
sudo mv cloud-cli /usr/local/bin/

# Verify installation
cloud-cli --version
# Output: cloud-cli version 2.0.0
```

### Initial Configuration

```bash
# Set API key
export CLOUD_API_KEY="your-api-key-here"

# Or configure permanently
cloud-cli config set cloud.api_key "your-api-key-here"

# Verify connection
cloud-cli compute instances list
```

---

## Step 7: Working with Async Operations

### Create Instance with Async Tracking

```bash
# Start instance creation (returns immediately)
cloud-cli compute instances create \
  --name web-server-1 \
  --size large \
  --region us-west

# Output:
# ✓ Instance creation started
# Operation ID: op-abc123
#
# Track progress:
#   cloud-cli operations get --id op-abc123
```

### Wait for Operation Completion

```bash
# Create and wait for completion
cloud-cli compute instances create \
  --name web-server-1 \
  --size large \
  --region us-west \
  --wait

# Output with progress:
# Creating instance web-server-1...
# [████████████████░░░░] 75% | 45s elapsed | ~15s remaining
#
# ✓ Instance web-server-1 created successfully
#
# ID: i-abc123
# Public IP: 203.0.113.42
# State: running
```

### Monitor Operation Status

```bash
# Check operation status
cloud-cli operations get --id op-abc123

# Output:
# Operation: op-abc123
# Type: create_instance
# Status: running
# Progress: 75%
# Started: 2025-11-25 10:30:00
# Elapsed: 45s
# Target: web-server-1
```

### Multiple Parallel Operations

```bash
# Start multiple instances (parallel)
for i in {1..5}; do
  cloud-cli compute instances create \
    --name "worker-$i" \
    --size medium \
    --region us-west &
done

# Wait for all background jobs
wait

# Check all operations
cloud-cli operations list --status running
```

---

## Step 8: Workflow Execution

### Run Full Stack Deployment

```bash
# Execute workflow
cloud-cli deploy full-stack \
  --template workflows/full-stack-deploy.yaml \
  --env production \
  --region us-west \
  --instance-count 5

# Output:
# Starting deployment: full-stack-deploy
# Environment: production
# Region: us-west
#
# Step 1/5: Create storage bucket...
# ✓ Bucket myapp-production-data created (2s)
#
# Step 2/5: Create database instance...
# Creating instance myapp-production-db...
# [████████████████████] 100% | 120s
# ✓ Database instance created (120s)
#
# Step 3/5: Create application instances (5)...
# Creating 5 instances in parallel...
#   ✓ myapp-production-app-1 (45s)
#   ✓ myapp-production-app-2 (47s)
#   ✓ myapp-production-app-3 (46s)
#   ✓ myapp-production-app-4 (48s)
#   ✓ myapp-production-app-5 (45s)
#
# Step 4/5: Configure load balancer...
# ✓ Load balancer configured (15s)
#
# Step 5/5: Setup monitoring...
# ✓ Monitoring configured (8s)
#
# Deployment Summary:
# Total time: 240s
# Resources created: 8
# Status: SUCCESS
```

### Dry Run Mode

```bash
# Test workflow without execution
cloud-cli deploy full-stack \
  --template workflows/full-stack-deploy.yaml \
  --env staging \
  --dry-run

# Output:
# DRY RUN: No resources will be created
#
# Workflow: full-stack-deploy
# Environment: staging
#
# Execution Plan:
#   1. create_storage
#      → Create bucket: myapp-staging-data
#      → Region: us-west
#      → Encryption: aes256
#
#   2. create_database (depends on: create_storage)
#      → Create instance: myapp-staging-db
#      → Size: medium
#      → Image: postgres-14
#
#   3. create_app_instances (depends on: create_database)
#      → Create 3 instances (parallel)
#      → Names: myapp-staging-app-{1..3}
#
#   4. configure_lb (depends on: create_app_instances)
#      → Create load balancer
#      → Targets: 3 instances
#
# Estimated cost: $245/month
# Estimated duration: 180s
```

---

## Step 9: State Management

### View Current State

```bash
# List all managed resources
cloud-cli state list

# Output:
# Resource Type        ID              Name                    State
# instance            i-abc123        myapp-prod-db           running
# instance            i-def456        myapp-prod-app-1        running
# instance            i-ghi789        myapp-prod-app-2        running
# bucket              bkt-123         myapp-prod-data         active
# load_balancer       lb-456          myapp-prod-lb           active
```

### Detect State Drift

```bash
# Check for drift
cloud-cli state drift-detect

# Output:
# Checking for configuration drift...
#
# ⚠️  Drift detected for 2 resources:
#
# instance/i-abc123 (myapp-prod-db):
#   - tags.Environment: "production" → "prod" (modified externally)
#   - size: "large" → "xlarge" (modified externally)
#
# instance/i-def456 (myapp-prod-app-1):
#   - state: "running" → "stopped" (unexpected state)
#
# Run 'cloud-cli state refresh' to update state
# Or 'cloud-cli state apply' to restore desired configuration
```

### Refresh State

```bash
# Update state from actual infrastructure
cloud-cli state refresh

# Output:
# Refreshing state...
# ✓ instance/i-abc123 (refreshed)
# ✓ instance/i-def456 (refreshed)
# ✓ instance/i-ghi789 (refreshed)
# ✓ bucket/bkt-123 (refreshed)
# ✓ load_balancer/lb-456 (refreshed)
#
# State updated successfully
```

### Export State

```bash
# Export as Terraform
cloud-cli state export --format terraform > infrastructure.tf

# Export as JSON
cloud-cli state export --format json > state.json

# Export as YAML
cloud-cli state export --format yaml > state.yaml
```

---

## Step 10: Streaming Operations

### File Sync with Progress

```bash
# Sync files to cloud storage
cloud-cli storage sync \
  --source ./data \
  --destination s3://my-bucket/backup \
  --watch

# Output (streaming):
# Starting sync: ./data → s3://my-bucket/backup
#
# [FILE STARTED] data/file1.txt (1.2 MB)
# [PROGRESS] data/file1.txt: [████████████░░░░] 75% | 900 KB/s
# [FILE COMPLETED] data/file1.txt (1.2 MB in 1.5s)
#
# [FILE STARTED] data/large-file.zip (250 MB)
# [PROGRESS] data/large-file.zip: [███░░░░░░░░░░░░░] 18% | 15 MB/s | ETA: 14s
# [FILE COMPLETED] data/large-file.zip (250 MB in 17s)
#
# [FILE STARTED] data/document.pdf (3.4 MB)
# [ERROR] data/document.pdf: permission denied
#
# Sync Summary:
#   Files synced: 124/125
#   Bytes transferred: 1.2 GB
#   Duration: 145s
#   Average speed: 8.5 MB/s
#   Errors: 1
#
# Watching for changes... (Ctrl+C to stop)
```

### Watch Mode for Continuous Sync

```bash
# Continuous sync with watching
cloud-cli storage sync \
  --source ./app-data \
  --destination s3://my-bucket \
  --watch \
  --delete

# Output:
# Watching ./app-data for changes...
#
# [11:23:45] FILE CREATED: app-data/new-file.txt
# [11:23:45] Syncing new-file.txt... ✓ (245 KB in 0.3s)
#
# [11:24:12] FILE MODIFIED: app-data/config.yaml
# [11:24:12] Syncing config.yaml... ✓ (2.1 KB in 0.1s)
#
# [11:25:33] FILE DELETED: app-data/old-file.txt
# [11:25:33] Deleting from s3://my-bucket/old-file.txt... ✓
```

---

## Step 11: Advanced Error Handling

### Automatic Retries

```bash
# Retry configuration is automatic
cloud-cli compute instances list --region us-west

# With debug output:
# → GET /compute/instances?region=us-west
# ← 503 Service Unavailable
# ℹ️  Retrying in 1s (attempt 1/3)...
#
# → GET /compute/instances?region=us-west
# ← 503 Service Unavailable
# ℹ️  Retrying in 2s (attempt 2/3)...
#
# → GET /compute/instances?region=us-west
# ← 200 OK
# ✓ Success
```

### Circuit Breaker

```bash
# When service is unhealthy
cloud-cli compute instances create --name test

# Output:
# ✗ Error: Circuit breaker is OPEN
#
# The Cloud Compute API is currently experiencing issues.
# Circuit opened after 5 consecutive failures.
# Will retry in 60s.
#
# Use --force-circuit-close to override (not recommended)
```

### Rollback on Workflow Failure

```bash
# Deploy with automatic rollback
cloud-cli deploy full-stack \
  --template workflows/full-stack-deploy.yaml \
  --env staging

# If step fails:
# Step 3/5: Create application instances...
# ✓ myapp-staging-app-1 created
# ✓ myapp-staging-app-2 created
# ✗ myapp-staging-app-3 failed: quota exceeded
#
# ⚠️  Workflow failed at step 3/5
# Initiating automatic rollback...
#
# Rollback 1/3: Delete app instances...
#   ✓ Deleted myapp-staging-app-1
#   ✓ Deleted myapp-staging-app-2
#
# Rollback 2/3: Delete database...
#   ✓ Deleted myapp-staging-db
#
# Rollback 3/3: Delete storage bucket...
#   ✓ Deleted myapp-staging-data
#
# Rollback completed successfully
# All resources cleaned up
```

---

## Step 12: Production Best Practices

### Use Configuration Profiles

Create `~/.cloud-cli/profiles.yaml`:

```yaml
profiles:
  development:
    cloud:
      api_key: "dev-key"
    output:
      colors: true
      format: table
    logging:
      level: debug

  production:
    cloud:
      api_key: "prod-key"
    output:
      colors: false
      format: json
    logging:
      level: warn
      file:
        enabled: true
    async_operations:
      polling:
        default_interval: 10s
```

Use profiles:

```bash
# Use development profile
cloud-cli --profile development compute instances list

# Use production profile
cloud-cli --profile production deploy full-stack --env production
```

### Implement Health Checks

```bash
# Pre-flight checks before deployment
cloud-cli deploy full-stack \
  --template workflows/full-stack-deploy.yaml \
  --env production \
  --pre-flight-checks

# Output:
# Running pre-flight checks...
#
# ✓ API connectivity
# ✓ Authentication valid
# ✓ Quota available (instances: 15/100)
# ✓ Region available (us-west)
# ✗ Network configuration (VPC not found)
#
# Pre-flight checks failed. Fix errors before deploying.
```

### Use Tagging for Resource Management

```bash
# Create with comprehensive tags
cloud-cli compute instances create \
  --name prod-web-1 \
  --size large \
  --region us-west \
  --tag "Environment=production" \
  --tag "Team=platform" \
  --tag "CostCenter=engineering" \
  --tag "ManagedBy=cloud-cli" \
  --tag "CreatedAt=$(date -u +%Y-%m-%dT%H:%M:%SZ)"

# Query by tags
cloud-cli compute instances list \
  --tag "Environment=production" \
  --tag "Team=platform"
```

---

## Troubleshooting

### Long-Running Operations Timeout

**Problem**: Operation times out before completing

**Solution**:
```bash
# Increase timeout
cloud-cli compute instances create \
  --name large-instance \
  --size xlarge \
  --timeout 1800s \
  --wait

# Or configure globally
cloud-cli config set async_operations.timeout.default 1800s
```

### State File Conflicts

**Problem**: Multiple users modifying state simultaneously

**Solution**:
```bash
# Enable state locking
cloud-cli config set state.lock_enabled true
cloud-cli config set state.lock_timeout 300s

# If lock is stuck
cloud-cli state unlock --force
```

### Workflow Step Failures

**Problem**: Single step failure stops entire workflow

**Solution**:
```yaml
# In workflow definition, use continue_on_error
steps:
  - name: optional_step
    operation: createMonitoring
    on_error:
      action: continue_on_error
      max_failures: 1
```

### Memory Issues with Large Syncs

**Problem**: CLI crashes during large file sync

**Solution**:
```bash
# Use streaming mode with smaller buffer
cloud-cli storage sync \
  --source ./large-dataset \
  --destination s3://bucket \
  --chunk-size 5MB \
  --max-concurrent 3

# Or sync in batches
find ./large-dataset -type f | split -l 1000 - batch-
for batch in batch-*; do
  cloud-cli storage sync --source-list $batch --destination s3://bucket
done
```

---

## Testing

### Integration Test Script

Create `tests/integration-test.sh`:

```bash
#!/bin/bash
set -e

TEST_PREFIX="clitest-$(date +%s)"
REGION="us-west"

echo "Running integration tests..."

# Test 1: Instance lifecycle
echo "Test: Instance lifecycle"
INSTANCE_ID=$(cloud-cli compute instances create \
  --name "${TEST_PREFIX}-instance" \
  --size small \
  --region $REGION \
  --wait \
  --format json | jq -r '.instance.id')

cloud-cli compute instances stop --id $INSTANCE_ID --wait
cloud-cli compute instances start --id $INSTANCE_ID --wait
cloud-cli compute instances delete --id $INSTANCE_ID --force --wait
echo "✓ Instance lifecycle test passed"

# Test 2: Storage operations
echo "Test: Storage operations"
BUCKET_NAME="${TEST_PREFIX}-bucket"
cloud-cli storage buckets create --name $BUCKET_NAME --region $REGION

# Create test file
echo "test data" > /tmp/test-file.txt
cloud-cli storage objects upload \
  --bucket $BUCKET_NAME \
  --key test-file.txt \
  --file /tmp/test-file.txt

cloud-cli storage objects list --bucket $BUCKET_NAME
cloud-cli storage buckets delete --name $BUCKET_NAME --force
echo "✓ Storage operations test passed"

# Test 3: Workflow execution
echo "Test: Workflow execution"
cloud-cli deploy full-stack \
  --template workflows/full-stack-deploy.yaml \
  --env dev \
  --instance-count 1 \
  --dry-run
echo "✓ Workflow dry-run test passed"

echo "All tests passed!"
```

---

## Next Steps

### Advanced Features to Explore

1. **Custom Workflows**: Create domain-specific workflows
2. **Plugin Development**: Extend CLI with custom commands
3. **CI/CD Integration**: See [CI/CD Tutorial](tutorial-ci-cd-integration.md)
4. **Multi-Cloud**: Manage multiple cloud providers
5. **Disaster Recovery**: Implement backup and restore workflows

### Further Reading

- [CliForge Workflow Orchestration](../user-guide-workflows.md)

> **Note**: Additional architectural design documentation on state management, progress tracking, streaming operations, and plugin architecture is available in the internal design documentation.

---

## Summary

In this tutorial, you learned how to:

- ✓ Build a CLI for cloud infrastructure management
- ✓ Handle asynchronous operations with progress tracking
- ✓ Implement workflow orchestration for complex deployments
- ✓ Manage infrastructure state and detect drift
- ✓ Use streaming operations for file sync
- ✓ Implement comprehensive error handling and retries
- ✓ Configure circuit breakers for resilience
- ✓ Use rollback strategies for failed deployments
- ✓ Apply production best practices

You now have the skills to build enterprise-grade infrastructure management CLIs with CliForge!

---

**Tutorial Version**: 1.0.0
**Last Updated**: 2025-11-25
**CliForge Version**: 0.9.0+
