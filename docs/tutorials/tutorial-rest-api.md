# Tutorial: Building a Simple REST API CLI

**Difficulty**: Beginner
**Time**: 45-60 minutes
**Version**: 1.0.0

---

## Overview

In this tutorial, you'll build a fully functional CLI for the GitHub API, learning how to:

- Generate a CLI from an OpenAPI specification
- Implement API authentication
- Create CRUD operations for repositories and issues
- Handle errors gracefully
- Format output for different use cases

By the end, you'll have a working `github-cli` tool that can manage repositories and issues.

### What You'll Build

A CLI tool with these commands:

```bash
github-cli repos list --user octocat
github-cli repos create --name my-repo --description "My new repository"
github-cli repos get --owner octocat --repo hello-world
github-cli issues list --owner octocat --repo hello-world
github-cli issues create --owner octocat --repo hello-world --title "Bug report"
```

---

## Prerequisites

### Required Knowledge

- Basic command-line experience
- Understanding of REST APIs and HTTP methods
- Familiarity with YAML syntax
- Basic understanding of OpenAPI/Swagger (helpful but not required)

### Required Software

- **Go**: Version 1.21 or later
- **Git**: For version control
- **CliForge**: Latest version installed
- **Text Editor**: VS Code, Vim, or your preferred editor
- **GitHub Account**: For API testing (free tier is fine)

### Installation Check

Verify your environment:

```bash
# Check Go version
go version
# Expected: go version go1.21.0 or higher

# Check CliForge installation
cliforge version
# Expected: CliForge version 0.9.0 or higher

# Verify git
git --version
# Expected: git version 2.x or higher
```

---

## Step 1: Understanding the GitHub API

Before we build our CLI, let's understand what we're working with.

### GitHub API Basics

The GitHub REST API provides programmatic access to:

- **Repositories**: Create, read, update, delete repositories
- **Issues**: Manage issues and pull requests
- **Users**: Access user profiles and organizations
- **Authentication**: Token-based authentication

### API Endpoints We'll Use

```
GET    /user/repos              - List user's repositories
POST   /user/repos              - Create a repository
GET    /repos/{owner}/{repo}    - Get repository details
PATCH  /repos/{owner}/{repo}    - Update repository
DELETE /repos/{owner}/{repo}    - Delete repository
GET    /repos/{owner}/{repo}/issues        - List issues
POST   /repos/{owner}/{repo}/issues        - Create issue
GET    /repos/{owner}/{repo}/issues/{number} - Get issue
PATCH  /repos/{owner}/{repo}/issues/{number} - Update issue
```

### Authentication

GitHub requires authentication for most operations. We'll use:

- **Personal Access Token (PAT)**: Classic token with repo scope
- **Header-based auth**: `Authorization: token YOUR_TOKEN`

---

## Step 2: Project Setup

### Create Your Project Directory

```bash
# Create project directory
mkdir github-cli-tutorial
cd github-cli-tutorial

# Initialize git repository
git init

# Create directory structure
mkdir -p {specs,config,examples}
```

Your directory structure should look like:

```
github-cli-tutorial/
├── specs/          # OpenAPI specifications
├── config/         # CLI configuration files
└── examples/       # Example usage scripts
```

### Create a GitHub Personal Access Token

1. Go to GitHub Settings > Developer settings > Personal access tokens
2. Click "Generate new token (classic)"
3. Give it a descriptive name: "CliForge GitHub CLI Tutorial"
4. Select scopes:
   - `repo` (full control of private repositories)
   - `read:user` (read user profile data)
5. Click "Generate token"
6. **Save the token securely** - you won't see it again

Store your token in an environment variable:

```bash
export GITHUB_TOKEN="ghp_your_token_here"

# Add to your shell profile to persist
echo 'export GITHUB_TOKEN="ghp_your_token_here"' >> ~/.bashrc  # or ~/.zshrc
```

---

## Step 3: Creating the OpenAPI Specification

We'll create a simplified OpenAPI spec for GitHub API operations.

Create `specs/github-api.yaml`:

```yaml
openapi: 3.0.3
info:
  title: GitHub API CLI
  description: Simplified GitHub API for CLI demonstration
  version: 1.0.0
  contact:
    name: GitHub Support
    url: https://support.github.com

servers:
  - url: https://api.github.com
    description: GitHub REST API v3

security:
  - BearerAuth: []

components:
  securitySchemes:
    BearerAuth:
      type: http
      scheme: bearer
      description: GitHub Personal Access Token
      x-cli-auth:
        type: token
        token_prefix: "token"
        header: "Authorization"
        env_var: "GITHUB_TOKEN"

  schemas:
    Repository:
      type: object
      properties:
        id:
          type: integer
          description: Repository ID
        name:
          type: string
          description: Repository name
        full_name:
          type: string
          description: Full repository name (owner/name)
        description:
          type: string
          description: Repository description
        private:
          type: boolean
          description: Whether the repository is private
        html_url:
          type: string
          description: Repository URL
        created_at:
          type: string
          format: date-time
          description: Creation timestamp
        updated_at:
          type: string
          format: date-time
          description: Last update timestamp
        stargazers_count:
          type: integer
          description: Number of stars
        language:
          type: string
          description: Primary programming language

    CreateRepository:
      type: object
      required:
        - name
      properties:
        name:
          type: string
          description: Repository name
          pattern: "^[a-zA-Z0-9_-]+$"
        description:
          type: string
          description: Repository description
        private:
          type: boolean
          default: false
          description: Create private repository
        auto_init:
          type: boolean
          default: false
          description: Initialize with README

    Issue:
      type: object
      properties:
        number:
          type: integer
          description: Issue number
        title:
          type: string
          description: Issue title
        body:
          type: string
          description: Issue body
        state:
          type: string
          enum: [open, closed]
          description: Issue state
        html_url:
          type: string
          description: Issue URL
        created_at:
          type: string
          format: date-time
        updated_at:
          type: string
          format: date-time
        user:
          type: object
          properties:
            login:
              type: string

    CreateIssue:
      type: object
      required:
        - title
      properties:
        title:
          type: string
          description: Issue title
        body:
          type: string
          description: Issue body (supports markdown)
        labels:
          type: array
          items:
            type: string
          description: Labels to apply

    Error:
      type: object
      properties:
        message:
          type: string
        documentation_url:
          type: string

paths:
  /user/repos:
    get:
      summary: List user repositories
      description: List repositories for the authenticated user
      operationId: listUserRepos
      x-cli-command: "repos list"
      x-cli-aliases: ["list repos", "repo ls"]
      x-cli-description: "List your repositories"
      x-cli-output:
        table:
          columns:
            - field: name
              header: NAME
              width: 30
            - field: description
              header: DESCRIPTION
              width: 50
            - field: language
              header: LANGUAGE
              width: 15
            - field: stargazers_count
              header: STARS
              width: 8
            - field: private
              header: PRIVATE
              width: 8
              transform: "value ? '🔒' : ''"
      parameters:
        - name: type
          in: query
          schema:
            type: string
            enum: [all, owner, public, private, member]
            default: owner
          x-cli-flag:
            name: "--type"
            description: "Repository type filter"
        - name: sort
          in: query
          schema:
            type: string
            enum: [created, updated, pushed, full_name]
            default: updated
          x-cli-flag:
            name: "--sort"
            description: "Sort field"
        - name: per_page
          in: query
          schema:
            type: integer
            default: 30
            maximum: 100
          x-cli-flag:
            name: "--limit"
            description: "Number of results"
      responses:
        '200':
          description: List of repositories
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Repository'
        '401':
          description: Authentication required
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

    post:
      summary: Create repository
      description: Create a new repository for the authenticated user
      operationId: createRepo
      x-cli-command: "repos create"
      x-cli-aliases: ["create repo"]
      x-cli-description: "Create a new repository"
      x-cli-confirmation:
        enabled: false
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateRepository'
      x-cli-flags:
        - name: name
          source: name
          flag: "--name"
          required: true
          description: "Repository name"
        - name: description
          source: description
          flag: "--description"
          description: "Repository description"
        - name: private
          source: private
          flag: "--private"
          type: boolean
          description: "Make repository private"
        - name: auto_init
          source: auto_init
          flag: "--init"
          type: boolean
          description: "Initialize with README"
      responses:
        '201':
          description: Repository created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Repository'
          x-cli-output:
            success_message: "Repository created successfully"
            show_url: true
        '422':
          description: Validation error
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /repos/{owner}/{repo}:
    parameters:
      - name: owner
        in: path
        required: true
        schema:
          type: string
        x-cli-flag:
          name: "--owner"
          description: "Repository owner"
      - name: repo
        in: path
        required: true
        schema:
          type: string
        x-cli-flag:
          name: "--repo"
          description: "Repository name"

    get:
      summary: Get repository
      description: Get details for a specific repository
      operationId: getRepo
      x-cli-command: "repos get"
      x-cli-aliases: ["get repo", "repo show"]
      x-cli-description: "Get repository details"
      x-cli-output:
        default_format: yaml
      responses:
        '200':
          description: Repository details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Repository'
        '404':
          description: Repository not found
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

    delete:
      summary: Delete repository
      description: Delete a repository (requires admin access)
      operationId: deleteRepo
      x-cli-command: "repos delete"
      x-cli-aliases: ["delete repo", "repo rm"]
      x-cli-description: "Delete a repository"
      x-cli-confirmation:
        enabled: true
        message: "Are you sure you want to delete {owner}/{repo}? This cannot be undone!"
        require_exact_match: true
        match_text: "{owner}/{repo}"
      responses:
        '204':
          description: Repository deleted
          x-cli-output:
            success_message: "Repository {owner}/{repo} deleted successfully"
        '403':
          description: Forbidden
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Error'

  /repos/{owner}/{repo}/issues:
    parameters:
      - name: owner
        in: path
        required: true
        schema:
          type: string
        x-cli-flag:
          name: "--owner"
          description: "Repository owner"
      - name: repo
        in: path
        required: true
        schema:
          type: string
        x-cli-flag:
          name: "--repo"
          description: "Repository name"

    get:
      summary: List issues
      description: List issues for a repository
      operationId: listIssues
      x-cli-command: "issues list"
      x-cli-aliases: ["list issues", "issue ls"]
      x-cli-description: "List repository issues"
      x-cli-output:
        table:
          columns:
            - field: number
              header: "#"
              width: 6
            - field: title
              header: TITLE
              width: 50
            - field: state
              header: STATE
              width: 10
              transform: "value == 'open' ? '🟢 OPEN' : '⚪ CLOSED'"
            - field: user.login
              header: AUTHOR
              width: 20
      parameters:
        - name: state
          in: query
          schema:
            type: string
            enum: [open, closed, all]
            default: open
          x-cli-flag:
            name: "--state"
            description: "Issue state filter"
        - name: per_page
          in: query
          schema:
            type: integer
            default: 30
          x-cli-flag:
            name: "--limit"
            description: "Number of results"
      responses:
        '200':
          description: List of issues
          content:
            application/json:
              schema:
                type: array
                items:
                  $ref: '#/components/schemas/Issue'

    post:
      summary: Create issue
      description: Create a new issue
      operationId: createIssue
      x-cli-command: "issues create"
      x-cli-aliases: ["create issue"]
      x-cli-description: "Create a new issue"
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreateIssue'
      x-cli-flags:
        - name: title
          source: title
          flag: "--title"
          required: true
          description: "Issue title"
        - name: body
          source: body
          flag: "--body"
          description: "Issue description"
        - name: labels
          source: labels
          flag: "--labels"
          type: array
          description: "Comma-separated labels"
      responses:
        '201':
          description: Issue created
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Issue'
          x-cli-output:
            success_message: "Issue #{{ .number }} created"
            show_url: true

  /repos/{owner}/{repo}/issues/{issue_number}:
    parameters:
      - name: owner
        in: path
        required: true
        schema:
          type: string
        x-cli-flag:
          name: "--owner"
      - name: repo
        in: path
        required: true
        schema:
          type: string
        x-cli-flag:
          name: "--repo"
      - name: issue_number
        in: path
        required: true
        schema:
          type: integer
        x-cli-flag:
          name: "--number"
          description: "Issue number"

    get:
      summary: Get issue
      description: Get details for a specific issue
      operationId: getIssue
      x-cli-command: "issues get"
      x-cli-aliases: ["get issue"]
      x-cli-description: "Get issue details"
      x-cli-output:
        default_format: yaml
      responses:
        '200':
          description: Issue details
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/Issue'
```

Save this file as `specs/github-api.yaml`.

---

## Step 4: Creating the CLI Configuration

Now create the CliForge configuration file.

Create `config/cli-config.yaml`:

```yaml
# CLI Metadata
metadata:
  name: github-cli
  version: 1.0.0
  description: GitHub API Command Line Interface
  author: Your Name
  license: MIT

# API Configuration
api:
  openapi_spec: ../specs/github-api.yaml
  base_url: https://api.github.com

  # GitHub API requires specific headers
  default_headers:
    Accept: "application/vnd.github.v3+json"
    User-Agent: "github-cli/1.0.0"

# Authentication
authentication:
  default_method: token
  methods:
    token:
      type: bearer
      header: "Authorization"
      prefix: "token"
      env_var: "GITHUB_TOKEN"
      config_key: "github.token"

# Output Configuration
output:
  default_format: table
  formats:
    - table
    - json
    - yaml

  table:
    borders: true
    header_style: bold

  colors:
    enabled: true
    success: green
    error: red
    warning: yellow
    info: blue

# Error Handling
errors:
  show_http_codes: true
  show_response_body: true
  retry:
    enabled: true
    max_attempts: 3
    backoff: exponential
    retryable_codes:
      - 408  # Request Timeout
      - 429  # Too Many Requests
      - 500  # Internal Server Error
      - 502  # Bad Gateway
      - 503  # Service Unavailable
      - 504  # Gateway Timeout

# Rate Limiting
rate_limit:
  enabled: true
  show_remaining: true

# Caching
cache:
  enabled: true
  ttl: 300  # 5 minutes
  directory: ~/.cache/github-cli

# Branding
branding:
  tagline: "Manage GitHub repositories and issues from the command line"
  colors:
    primary: "#6e40c9"
    secondary: "#8957e5"

  help:
    examples:
      - description: "List your repositories"
        command: "github-cli repos list"
      - description: "Create a new repository"
        command: "github-cli repos create --name my-repo --description 'My project'"
      - description: "List issues in a repository"
        command: "github-cli issues list --owner octocat --repo hello-world"

# Logging
logging:
  level: info
  file: ~/.github-cli/logs/cli.log
  max_size_mb: 10
  max_backups: 3
```

---

## Step 5: Building the CLI

Now let's build our CLI binary.

### Initialize the CLI Project

```bash
# Navigate to your project directory
cd github-cli-tutorial

# Initialize with CliForge
cliforge init github-cli --config config/cli-config.yaml
```

This creates the Go module and necessary files.

### Build the Binary

```bash
# Build the CLI
cliforge build \
  --config config/cli-config.yaml \
  --output ./github-cli \
  --platform linux,darwin,windows

# Or for your current platform only
go build -o github-cli ./cmd/github-cli
```

Expected output:

```
✓ Loaded OpenAPI specification (specs/github-api.yaml)
✓ Validated 8 API operations
✓ Generated command structure
✓ Building binary...
✓ CLI built successfully: ./github-cli

Build Summary:
  Commands: 8
  Size: 12.5 MB
  Platform: darwin/arm64
```

### Verify the Build

```bash
# Check the binary
./github-cli --version
# Output: github-cli version 1.0.0

# View available commands
./github-cli --help
```

You should see:

```
GitHub API Command Line Interface

Manage GitHub repositories and issues from the command line

Usage:
  github-cli [command]

Available Commands:
  repos       Repository operations
  issues      Issue operations
  completion  Generate completion script
  help        Help about any command

Flags:
  -h, --help            help for github-cli
      --format string   Output format (table, json, yaml) (default "table")
      --no-color        Disable colored output
  -v, --verbose         Enable verbose output
      --version         version for github-cli

Use "github-cli [command] --help" for more information about a command.
```

---

## Step 6: Testing Basic Operations

### Configure Authentication

```bash
# Set your GitHub token
export GITHUB_TOKEN="ghp_your_token_here"

# Or configure it permanently
./github-cli config set github.token "ghp_your_token_here"
```

### Test Repository Listing

```bash
# List your repositories
./github-cli repos list

# Output:
# NAME                DESCRIPTION                                   LANGUAGE    STARS  PRIVATE
# my-awesome-project  A really cool project                        Go          42
# test-repo           Testing repository                           Python      3      🔒
# dotfiles            My configuration files                       Shell       15
```

### Test with Different Formats

```bash
# JSON output
./github-cli repos list --format json | jq .

# YAML output
./github-cli repos list --format yaml

# Limit results
./github-cli repos list --limit 5

# Filter by type
./github-cli repos list --type public

# Sort by creation date
./github-cli repos list --sort created
```

---

## Step 7: Creating Resources

### Create a Repository

```bash
# Create a public repository
./github-cli repos create \
  --name "test-cliforge" \
  --description "Testing CliForge CLI generation" \
  --init

# Output:
# ✓ Repository created successfully
#
# Repository: test-cliforge
# URL: https://github.com/your-username/test-cliforge
# Private: false
```

### Create with More Options

```bash
# Create a private repository
./github-cli repos create \
  --name "private-project" \
  --description "My secret project" \
  --private \
  --init

# Verify creation
./github-cli repos get --owner your-username --repo private-project
```

---

## Step 8: Working with Issues

### List Issues

```bash
# List open issues
./github-cli issues list \
  --owner octocat \
  --repo hello-world

# Output:
# #    TITLE                                    STATE      AUTHOR
# 123  Bug: Login not working                   🟢 OPEN    user1
# 122  Feature request: Add dark mode           🟢 OPEN    user2
# 121  Documentation needs update               ⚪ CLOSED  user3
```

### Create an Issue

```bash
# Create a new issue
./github-cli issues create \
  --owner your-username \
  --repo test-cliforge \
  --title "Add more features" \
  --body "We should add the following features:\n- Feature 1\n- Feature 2" \
  --labels "enhancement,help wanted"

# Output:
# ✓ Issue #1 created
# URL: https://github.com/your-username/test-cliforge/issues/1
```

### Get Issue Details

```bash
# Get specific issue
./github-cli issues get \
  --owner your-username \
  --repo test-cliforge \
  --number 1 \
  --format yaml
```

---

## Step 9: Error Handling

Let's test error handling scenarios.

### Test Authentication Error

```bash
# Temporarily unset token
unset GITHUB_TOKEN

# Try to list repos
./github-cli repos list

# Output:
# ✗ Error: Authentication required
#
# HTTP 401: Unauthorized
#
# To authenticate, set GITHUB_TOKEN environment variable:
#   export GITHUB_TOKEN="your_token_here"
#
# Or configure it:
#   github-cli config set github.token "your_token_here"
```

### Test Not Found Error

```bash
# Try to get non-existent repo
./github-cli repos get --owner octocat --repo does-not-exist

# Output:
# ✗ Error: Repository not found
#
# HTTP 404: Not Found
# Repository: octocat/does-not-exist
#
# Verify the owner and repository name are correct.
```

### Test Validation Error

```bash
# Try to create repo with invalid name
./github-cli repos create --name "invalid name with spaces"

# Output:
# ✗ Error: Validation failed
#
# HTTP 422: Unprocessable Entity
#
# Repository name must match pattern: ^[a-zA-Z0-9_-]+$
# Use only letters, numbers, hyphens, and underscores.
```

---

## Step 10: Advanced Features

### Using Confirmation Prompts

When deleting a repository, you'll be prompted for confirmation:

```bash
# Delete repository (with confirmation)
./github-cli repos delete --owner your-username --repo test-cliforge

# Output:
# ⚠️  WARNING: This action cannot be undone!
#
# Are you sure you want to delete your-username/test-cliforge?
# Type the repository name to confirm: test-cliforge
#
# ✓ Repository your-username/test-cliforge deleted successfully
```

### Using Verbose Mode

```bash
# Enable verbose output to see HTTP details
./github-cli repos list --verbose

# Output includes:
# → GET https://api.github.com/user/repos
# → Headers:
#     Authorization: token ghp_****
#     Accept: application/vnd.github.v3+json
#     User-Agent: github-cli/1.0.0
# ← 200 OK (245ms)
# ← Rate Limit: 4999/5000 remaining
#
# [normal output follows]
```

### Filtering with JQ

```bash
# Get just repository names
./github-cli repos list --format json | jq -r '.[].name'

# Get repos with more than 10 stars
./github-cli repos list --format json | jq '.[] | select(.stargazers_count > 10)'

# Get private repositories only
./github-cli repos list --format json | jq '.[] | select(.private == true)'
```

---

## Step 11: Configuration Management

### View Current Configuration

```bash
# Show all configuration
./github-cli config list

# Output:
# github.token=ghp_**** (from environment)
# output.format=table (default)
# cache.enabled=true (from config file)
```

### Set Configuration Values

```bash
# Set default output format
./github-cli config set output.format json

# Set cache TTL
./github-cli config set cache.ttl 600

# Disable colors
./github-cli config set output.colors.enabled false
```

### Configuration File Location

Configuration is stored at `~/.github-cli/config.yaml`:

```yaml
github:
  token: ghp_your_token_here

output:
  format: json
  colors:
    enabled: false

cache:
  enabled: true
  ttl: 600
```

---

## Step 12: Shell Completion

Enable shell completion for a better experience.

### Bash Completion

```bash
# Generate completion script
./github-cli completion bash > /etc/bash_completion.d/github-cli

# Or for user-level installation
./github-cli completion bash > ~/.github-cli-completion
echo 'source ~/.github-cli-completion' >> ~/.bashrc
source ~/.bashrc
```

### Zsh Completion

```bash
# Generate completion script
./github-cli completion zsh > "${fpath[1]}/_github-cli"

# Reload completions
compinit
```

### Test Completion

```bash
# Type and press TAB
./github-cli repos <TAB>

# Shows: create  delete  get  list

./github-cli repos create --<TAB>

# Shows: --name  --description  --private  --init
```

---

## Testing Your CLI

### Create a Test Script

Create `examples/test-workflow.sh`:

```bash
#!/bin/bash
set -e

echo "Testing GitHub CLI..."

# Test 1: List repositories
echo "Test 1: Listing repositories"
./github-cli repos list --limit 5
echo "✓ Test 1 passed"

# Test 2: Create test repository
echo "Test 2: Creating test repository"
REPO_NAME="cliforge-test-$(date +%s)"
./github-cli repos create \
  --name "$REPO_NAME" \
  --description "Automated test repository" \
  --init
echo "✓ Test 2 passed"

# Test 3: Get repository details
echo "Test 3: Getting repository details"
./github-cli repos get --owner "$GITHUB_USER" --repo "$REPO_NAME" --format yaml
echo "✓ Test 3 passed"

# Test 4: Create issue
echo "Test 4: Creating issue"
./github-cli issues create \
  --owner "$GITHUB_USER" \
  --repo "$REPO_NAME" \
  --title "Test issue" \
  --body "This is an automated test issue"
echo "✓ Test 4 passed"

# Test 5: List issues
echo "Test 5: Listing issues"
./github-cli issues list --owner "$GITHUB_USER" --repo "$REPO_NAME"
echo "✓ Test 5 passed"

# Test 6: Cleanup - Delete repository
echo "Test 6: Cleaning up"
./github-cli repos delete --owner "$GITHUB_USER" --repo "$REPO_NAME" --yes
echo "✓ Test 6 passed"

echo "All tests passed! ✓"
```

Run the test:

```bash
chmod +x examples/test-workflow.sh
export GITHUB_USER="your-username"
./examples/test-workflow.sh
```

---

## Troubleshooting

### Common Issues and Solutions

#### Issue: "Authentication required"

**Problem**: CLI cannot find GitHub token

**Solutions**:
```bash
# Check if token is set
echo $GITHUB_TOKEN

# Set token in environment
export GITHUB_TOKEN="ghp_your_token_here"

# Or set in config
./github-cli config set github.token "ghp_your_token_here"

# Verify token is valid
curl -H "Authorization: token $GITHUB_TOKEN" https://api.github.com/user
```

#### Issue: "Rate limit exceeded"

**Problem**: Too many API requests

**Solutions**:
```bash
# Check rate limit status
./github-cli --verbose repos list 2>&1 | grep "Rate Limit"

# Wait for rate limit reset
# Or use authenticated requests (higher limit)

# Enable caching to reduce requests
./github-cli config set cache.enabled true
./github-cli config set cache.ttl 600
```

#### Issue: "Command not found: repos"

**Problem**: OpenAPI spec not loaded correctly

**Solutions**:
```bash
# Verify spec file exists
ls -l specs/github-api.yaml

# Validate spec
cliforge validate --spec specs/github-api.yaml

# Rebuild CLI
cliforge build --config config/cli-config.yaml --output ./github-cli
```

#### Issue: "Invalid repository name"

**Problem**: Repository name doesn't match GitHub requirements

**Solutions**:
```bash
# Use only valid characters
# ✓ Valid: my-repo, my_repo, MyRepo123
# ✗ Invalid: my repo, my@repo, my.repo

# Create with valid name
./github-cli repos create --name my-awesome-repo
```

### Debug Mode

Enable debug logging:

```bash
# Set debug log level
./github-cli --log-level debug repos list

# Output shows detailed execution
# DEBUG: Loading config from ~/.github-cli/config.yaml
# DEBUG: OpenAPI spec loaded: 8 operations
# DEBUG: Authenticating with bearer token
# DEBUG: Sending GET https://api.github.com/user/repos
# DEBUG: Response: 200 OK (156ms)
```

### Getting Help

```bash
# General help
./github-cli --help

# Command-specific help
./github-cli repos --help
./github-cli repos create --help

# Show examples
./github-cli repos create --examples
```

---

## Best Practices

### 1. Secure Token Management

**DO**:
```bash
# Use environment variables
export GITHUB_TOKEN="ghp_token"

# Use config file with restricted permissions
chmod 600 ~/.github-cli/config.yaml

# Use secret management tools
./github-cli config set github.token "$(vault read -field=token secret/github)"
```

**DON'T**:
```bash
# Don't commit tokens to git
echo "GITHUB_TOKEN=ghp_token" >> .env  # Add to .gitignore!

# Don't use tokens in command history
./github-cli --token ghp_token repos list  # Token visible in history!
```

### 2. Error Handling

**DO**:
```bash
# Check exit codes in scripts
if ./github-cli repos create --name test-repo; then
  echo "Success"
else
  echo "Failed with code $?"
  exit 1
fi

# Use JSON output for parsing
RESULT=$(./github-cli repos get --owner user --repo repo --format json)
if [ $? -eq 0 ]; then
  STARS=$(echo "$RESULT" | jq -r '.stargazers_count')
fi
```

### 3. Performance Optimization

```bash
# Enable caching for repeated queries
./github-cli config set cache.enabled true

# Use specific fields in JSON queries
./github-cli repos list --format json | jq '{name, stars: .stargazers_count}'

# Limit results when you don't need everything
./github-cli repos list --limit 10
```

### 4. Output Formatting

```bash
# Use appropriate format for context
./github-cli repos list --format table  # Human-readable
./github-cli repos list --format json   # For scripts/jq
./github-cli repos list --format yaml   # Configuration/review

# Disable colors in CI/CD
./github-cli repos list --no-color
```

---

## Next Steps

Congratulations! You've built a functional GitHub CLI. Here are some ways to extend it:

### Add More Endpoints

Extend `specs/github-api.yaml` with:
- **Pull Requests**: List, create, merge PRs
- **Releases**: Create and manage releases
- **Workflows**: Trigger GitHub Actions
- **Gists**: Manage code snippets

### Add Workflows

Create a workflow file `config/workflows.yaml`:

```yaml
workflows:
  - name: create-project
    description: Create repository and setup project
    steps:
      - operation: createRepo
        params:
          name: "{{ .project_name }}"
          description: "{{ .description }}"
          private: "{{ .private }}"
          auto_init: true

      - operation: createIssue
        params:
          owner: "{{ .github_user }}"
          repo: "{{ .project_name }}"
          title: "Initial setup"
          body: "Project created via CLI workflow"
          labels: ["setup"]
```

### Add Custom Commands

Extend with custom commands in `config/custom-commands.yaml`:

```yaml
custom_commands:
  - name: clone-all
    description: Clone all your repositories
    script: |
      repos=$(github-cli repos list --format json)
      echo "$repos" | jq -r '.[].clone_url' | xargs -I {} git clone {}
```

### Integration Examples

See these tutorials for integration:
- [CI/CD Integration Tutorial](tutorial-ci-cd-integration.md)
- [Cloud Management Tutorial](tutorial-cloud-management.md)

### Further Reading

- [CliForge User Guide - Authentication](../user-guide-authentication.md)
- [CliForge OpenAPI Extensions Reference](../openapi-extensions-reference.md)
- [CliForge Configuration DSL](../configuration-dsl.md)
- [GitHub API Documentation](https://docs.github.com/en/rest)

---

## Summary

In this tutorial, you learned how to:

- ✓ Create an OpenAPI specification for a REST API
- ✓ Configure CliForge to generate a branded CLI
- ✓ Implement authentication with bearer tokens
- ✓ Create CRUD operations for resources
- ✓ Handle errors gracefully with retries
- ✓ Format output for different use cases (table, JSON, YAML)
- ✓ Add confirmation prompts for destructive actions
- ✓ Enable shell completion
- ✓ Test and debug your CLI

You now have a solid foundation for building CLI tools with CliForge!

---

**Tutorial Version**: 1.0.0
**Last Updated**: 2025-11-25
**CliForge Version**: 0.9.0+
