# Plugin Developer Guide

**Version**: 0.9.0
**Last Updated**: 2025-11-23

---

## Table of Contents

1. [Introduction](#introduction)
2. [Plugin System Overview](#plugin-system-overview)
3. [Built-in Plugins](#built-in-plugins)
4. [Creating External Plugins](#creating-external-plugins)
5. [Plugin Manifest Format](#plugin-manifest-format)
6. [Permission System](#permission-system)
7. [Security Considerations](#security-considerations)
8. [Testing Plugins](#testing-plugins)
9. [Plugin Examples](#plugin-examples)
10. [Distribution and Installation](#distribution-and-installation)
11. [Troubleshooting](#troubleshooting)

---

## Introduction

The CliForge plugin system extends CLI capabilities beyond pure OpenAPI specifications. Plugins enable:

- Executing external command-line tools (AWS CLI, kubectl, etc.)
- Performing local file operations and transformations
- Implementing custom validation logic
- Integrating with authentication providers
- Creating custom commands and workflows

This guide teaches you how to create, test, and distribute plugins for CliForge-generated CLIs.

### Prerequisites

- CliForge v0.9.0 or later
- Basic understanding of CliForge workflows
- Familiarity with Go (for built-in plugins) or any language (for external plugins)

---

## Plugin System Overview

### Architecture

```
┌─────────────────────────────────────────────────────────┐
│                    CliForge Binary                      │
├─────────────────────────────────────────────────────────┤
│                                                         │
│  ┌───────────────────────────────────────────────────┐ │
│  │              Plugin Manager                       │ │
│  ├───────────────────────────────────────────────────┤ │
│  │  • Discovery  • Validation  • Permission Control │ │
│  │  • Execution  • Sandboxing  • Error Handling     │ │
│  └───────────────────────────────────────────────────┘ │
│                          │                             │
│         ┌────────────────┴────────────────┐            │
│         │                                 │            │
│  ┌──────▼──────┐                  ┌──────▼──────┐     │
│  │   Built-in  │                  │   External  │     │
│  │   Plugins   │                  │   Plugins   │     │
│  └─────────────┘                  └─────────────┘     │
│  • exec         Compiled           Binary/WASM       │
│  • file-ops     into CLI           executables       │
│  • validators                                         │
│  • transformers                                       │
│                                                         │
└─────────────────────────────────────────────────────────┘
```

### Plugin Types

| Type | Location | Security | Use Case |
|------|----------|----------|----------|
| **Built-in** | Compiled into binary | Fully trusted | Common operations (file ops, exec) |
| **External Binary** | `~/.config/<cli>/plugins/` | Sandboxed | Custom tools, scripts |
| **WASM** (v1.0) | `~/.config/<cli>/plugins/` | Sandboxed by design | Portable, secure custom logic |

### Plugin Lifecycle

```
1. Discovery
   └─> Scan plugin directories
       Find plugin manifests

2. Registration
   └─> Load manifest
       Validate plugin
       Check permissions

3. Approval (first use)
   └─> Request user permission
       Save approval state

4. Execution
   └─> Create isolated context
       Execute plugin
       Capture output

5. Cleanup
   └─> Release resources
       Update usage statistics
```

---

## Built-in Plugins

CliForge ships with four built-in plugins that cover common use cases.

### 1. Exec Plugin

Execute external command-line tools.

**Plugin Name**: `exec`

**Capabilities**:
- Run any executable in PATH
- Capture stdout/stderr
- Set environment variables
- Configure working directory
- Apply sandboxing

**Usage in Workflow**:
```yaml
- id: run-aws-cli
  type: plugin
  plugin: exec
  input:
    command: aws
    args:
      - sts
      - get-caller-identity
      - --profile
      - "{flags.aws_profile}"
    env:
      AWS_REGION: "{flags.region}"
  output:
    account_id: "{result.Account}"
```

**Input Fields**:
- `command` (string, required): Command to execute
- `args` (array, optional): Command arguments
- `env` (map, optional): Environment variables
- `stdin` (string, optional): Standard input
- `working_dir` (string, optional): Working directory
- `timeout` (duration, optional): Execution timeout

**Output Fields**:
- `stdout` (string): Standard output
- `stderr` (string): Standard error
- `exit_code` (int): Process exit code
- `duration` (duration): Execution time

**Security**:
- Command injection protection
- Allowed command list (optional)
- Environment variable filtering
- Sandboxing support

### 2. File-Ops Plugin

Read, parse, validate, and transform files.

**Plugin Name**: `file-ops`

**Capabilities**:
- Read file contents
- Parse JSON, YAML, PEM, htpasswd
- Validate file formats
- Base64 encode/decode
- Parse X.509 certificates

**Usage in Workflow**:

**Reading a file**:
```yaml
- id: read-config
  type: plugin
  plugin: file-ops
  input:
    operation: read
    file: "{flags.config_file}"
  output:
    content: "{result.content}"
    size: "{result.size}"
```

**Parsing JSON**:
```yaml
- id: parse-json
  type: plugin
  plugin: file-ops
  input:
    operation: parse
    file: "{flags.json_file}"
    format: json
  output:
    data: "{result.parsed}"
```

**Parsing X.509 certificate**:
```yaml
- id: parse-cert
  type: plugin
  plugin: file-ops
  input:
    operation: parse
    file: "{flags.ca_cert}"
    format: pem
    type: x509-certificate
  output:
    cert_data: "{result.base64}"
    issuer: "{result.issuer}"
    not_after: "{result.not_after}"
```

**Parsing htpasswd file**:
```yaml
- id: parse-htpasswd
  type: plugin
  plugin: file-ops
  input:
    operation: parse
    file: "{flags.htpasswd_file}"
    format: htpasswd
  output:
    users: "{result.users}"
    user_count: "{result.count}"
```

**Input Fields**:
- `operation` (string, required): Operation type (`read`, `parse`, `validate`, `transform`)
- `file` (string, required): File path
- `format` (string, required for parse): Format type (`json`, `yaml`, `pem`, `htpasswd`)
- `transformation` (string, required for transform): Transformation type (`base64-encode`, `base64-decode`)

**Output Fields** (vary by operation):
- `content` (string): File contents
- `size` (int): File size in bytes
- `parsed` (any): Parsed data structure
- `valid` (bool): Validation result
- `base64` (string): Base64-encoded data

**Security**:
- Path validation
- File size limits (10MB default)
- Allowed path restrictions
- No write operations (read-only)

### 3. Validators Plugin

Execute complex validation logic.

**Plugin Name**: `validators`

**Capabilities**:
- Regex pattern matching
- Length validation
- Range validation
- Custom validation functions
- Format validators (email, URL, DNS)

**Usage in Workflow**:

**Regex validation**:
```yaml
- id: validate-cluster-name
  type: plugin
  plugin: validators
  input:
    validator: regex
    value: "{flags.cluster_name}"
    pattern: "^[a-z][a-z0-9-]{0,53}$"
    error_message: "Cluster name must start with a letter and contain only lowercase letters, numbers, and hyphens"
```

**Length validation**:
```yaml
- id: validate-password
  type: plugin
  plugin: validators
  input:
    validator: length
    value: "{flags.password}"
    min: 8
    max: 128
    error_message: "Password must be 8-128 characters"
```

**Custom validation**:
```yaml
- id: validate-dns
  type: plugin
  plugin: validators
  input:
    validator: dns-available
    value: "{flags.cluster_name}.example.com"
```

**Input Fields**:
- `validator` (string, required): Validator type
- `value` (any, required): Value to validate
- `pattern` (string, for regex): Regular expression
- `min`, `max` (int, for length/range): Bounds
- `error_message` (string, optional): Custom error message

**Output Fields**:
- `valid` (bool): Validation result
- `message` (string): Error or success message

### 4. Transformers Plugin

Transform data between formats.

**Plugin Name**: `transformers`

**Capabilities**:
- Format conversion (JSON ↔ YAML)
- Data restructuring
- Template rendering
- String transformations

**Usage in Workflow**:

**Convert JSON to YAML**:
```yaml
- id: json-to-yaml
  type: plugin
  plugin: transformers
  input:
    operation: json-to-yaml
    data: "{flags.json_data}"
  output:
    yaml: "{result.yaml}"
```

**Template rendering**:
```yaml
- id: render-template
  type: plugin
  plugin: transformers
  input:
    operation: template
    template: "Cluster {{.name}} in {{.region}}"
    data:
      name: "{flags.cluster_name}"
      region: "{flags.region}"
  output:
    rendered: "{result.output}"
```

**Input Fields**:
- `operation` (string, required): Operation type
- `data` (any, required): Input data
- `template` (string, for templating): Template string

**Output Fields**:
- Varies by operation

---

## Creating External Plugins

External plugins are standalone executables that communicate with CliForge via JSON-RPC over stdin/stdout.

### Plugin Protocol

**JSON-RPC 2.0** over stdin/stdout:

**Request** (from CliForge to plugin):
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "execute",
  "params": {
    "command": "validate-credentials",
    "input": {
      "profile": "default",
      "region": "us-east-1"
    }
  }
}
```

**Response** (from plugin to CliForge):
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "result": {
    "exit_code": 0,
    "data": {
      "account_id": "123456789012",
      "user_arn": "arn:aws:iam::123456789012:user/admin"
    }
  }
}
```

**Error Response**:
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "error": {
    "code": -32000,
    "message": "AWS credentials not found",
    "data": {
      "suggestion": "Run 'aws configure' to set up credentials"
    }
  }
}
```

### Creating a Plugin in Python

**Example**: AWS credential validator plugin

**Step 1**: Create plugin structure
```bash
mkdir -p ~/.config/mycli/plugins/aws-validator
cd ~/.config/mycli/plugins/aws-validator
```

**Step 2**: Create `plugin.py`
```python
#!/usr/bin/env python3
import json
import sys
import boto3
from botocore.exceptions import ClientError, NoCredentialsError

def execute(params):
    """Execute plugin command."""
    command = params.get('command')
    input_data = params.get('input', {})

    if command == 'validate-credentials':
        return validate_credentials(input_data)
    else:
        raise ValueError(f"Unknown command: {command}")

def validate_credentials(input_data):
    """Validate AWS credentials."""
    profile = input_data.get('profile', 'default')

    try:
        session = boto3.Session(profile_name=profile)
        sts = session.client('sts')
        identity = sts.get_caller_identity()

        return {
            "exit_code": 0,
            "data": {
                "account_id": identity['Account'],
                "user_arn": identity['Arn'],
                "user_id": identity['UserId']
            }
        }
    except NoCredentialsError:
        return {
            "exit_code": 1,
            "error": "AWS credentials not found"
        }
    except ClientError as e:
        return {
            "exit_code": 1,
            "error": str(e)
        }

def main():
    """Main entry point for JSON-RPC communication."""
    for line in sys.stdin:
        try:
            request = json.loads(line)

            # Validate JSON-RPC request
            if request.get('jsonrpc') != '2.0':
                raise ValueError("Invalid JSON-RPC version")

            # Execute command
            params = request.get('params', {})
            result = execute(params)

            # Send response
            response = {
                "jsonrpc": "2.0",
                "id": request.get('id'),
                "result": result
            }
            print(json.dumps(response))
            sys.stdout.flush()

        except Exception as e:
            # Send error response
            error_response = {
                "jsonrpc": "2.0",
                "id": request.get('id') if 'request' in locals() else None,
                "error": {
                    "code": -32000,
                    "message": str(e)
                }
            }
            print(json.dumps(error_response))
            sys.stdout.flush()

if __name__ == '__main__':
    main()
```

**Step 3**: Make executable
```bash
chmod +x plugin.py
```

**Step 4**: Create `manifest.yaml`
```yaml
name: aws-validator
version: 1.0.0
type: binary
description: Validate AWS credentials and permissions
author: Your Name
executable: ./plugin.py

permissions:
  - type: read:env
    resource: AWS_*
    description: Read AWS environment variables
  - type: execute
    resource: aws
    description: Execute AWS CLI commands
  - type: network
    resource: "*.amazonaws.com"
    description: Make requests to AWS APIs

metadata:
  requires:
    - aws-cli
    - boto3
```

**Step 5**: Test plugin
```bash
# Test JSON-RPC communication
echo '{"jsonrpc":"2.0","id":"1","method":"execute","params":{"command":"validate-credentials","input":{"profile":"default"}}}' | ./plugin.py
```

### Creating a Plugin in Go

**Example**: DNS availability checker

**Step 1**: Create Go module
```bash
mkdir -p ~/plugins/dns-checker
cd ~/plugins/dns-checker
go mod init github.com/yourname/dns-checker
```

**Step 2**: Create `main.go`
```go
package main

import (
    "bufio"
    "encoding/json"
    "fmt"
    "net"
    "os"
    "time"
)

type Request struct {
    JSONRPC string                 `json:"jsonrpc"`
    ID      string                 `json:"id"`
    Method  string                 `json:"method"`
    Params  map[string]interface{} `json:"params"`
}

type Response struct {
    JSONRPC string      `json:"jsonrpc"`
    ID      string      `json:"id"`
    Result  interface{} `json:"result,omitempty"`
    Error   *Error      `json:"error,omitempty"`
}

type Error struct {
    Code    int                    `json:"code"`
    Message string                 `json:"message"`
    Data    map[string]interface{} `json:"data,omitempty"`
}

func main() {
    scanner := bufio.NewScanner(os.Stdin)

    for scanner.Scan() {
        var req Request
        if err := json.Unmarshal(scanner.Bytes(), &req); err != nil {
            sendError("", -32700, "Parse error", nil)
            continue
        }

        result, err := execute(req.Params)
        if err != nil {
            sendError(req.ID, -32000, err.Error(), nil)
            continue
        }

        sendResponse(req.ID, result)
    }
}

func execute(params map[string]interface{}) (interface{}, error) {
    command, ok := params["command"].(string)
    if !ok {
        return nil, fmt.Errorf("command is required")
    }

    input, _ := params["input"].(map[string]interface{})

    switch command {
    case "check-dns":
        return checkDNS(input)
    default:
        return nil, fmt.Errorf("unknown command: %s", command)
    }
}

func checkDNS(input map[string]interface{}) (interface{}, error) {
    domain, ok := input["domain"].(string)
    if !ok {
        return nil, fmt.Errorf("domain is required")
    }

    // Check if DNS name is already in use
    _, err := net.LookupHost(domain)
    available := err != nil

    return map[string]interface{}{
        "exit_code": 0,
        "data": map[string]interface{}{
            "domain":    domain,
            "available": available,
            "checked_at": time.Now().Format(time.RFC3339),
        },
    }, nil
}

func sendResponse(id string, result interface{}) {
    resp := Response{
        JSONRPC: "2.0",
        ID:      id,
        Result:  result,
    }

    data, _ := json.Marshal(resp)
    fmt.Println(string(data))
}

func sendError(id string, code int, message string, data map[string]interface{}) {
    resp := Response{
        JSONRPC: "2.0",
        ID:      id,
        Error: &Error{
            Code:    code,
            Message: message,
            Data:    data,
        },
    }

    data, _ := json.Marshal(resp)
    fmt.Println(string(data))
}
```

**Step 3**: Build plugin
```bash
go build -o dns-checker
```

**Step 4**: Install plugin
```bash
mkdir -p ~/.config/mycli/plugins/dns-checker
cp dns-checker ~/.config/mycli/plugins/dns-checker/
cp manifest.yaml ~/.config/mycli/plugins/dns-checker/
```

### Creating a Plugin in Bash

**Example**: Simple configuration validator

**plugin.sh**:
```bash
#!/bin/bash

# JSON-RPC handler
while IFS= read -r line; do
    # Parse JSON request
    command=$(echo "$line" | jq -r '.params.command')

    case "$command" in
        validate-config)
            # Get input
            config_file=$(echo "$line" | jq -r '.params.input.file')

            # Validate
            if [[ -f "$config_file" ]] && [[ -r "$config_file" ]]; then
                # Success response
                cat <<EOF
{
  "jsonrpc": "2.0",
  "id": $(echo "$line" | jq '.id'),
  "result": {
    "exit_code": 0,
    "data": {
      "valid": true,
      "path": "$config_file"
    }
  }
}
EOF
            else
                # Error response
                cat <<EOF
{
  "jsonrpc": "2.0",
  "id": $(echo "$line" | jq '.id'),
  "result": {
    "exit_code": 1,
    "error": "Config file not found or not readable"
  }
}
EOF
            fi
            ;;
        *)
            # Unknown command
            cat <<EOF
{
  "jsonrpc": "2.0",
  "id": $(echo "$line" | jq '.id'),
  "error": {
    "code": -32000,
    "message": "Unknown command: $command"
  }
}
EOF
            ;;
    esac
done
```

---

## Plugin Manifest Format

Every plugin must have a `manifest.yaml` file.

### Manifest Schema

```yaml
# Required fields
name: string                    # Unique plugin identifier
version: string                 # Semantic version (e.g., "1.0.0")
type: string                    # "builtin" | "binary" | "wasm"

# Optional metadata
description: string             # Human-readable description
author: string                  # Author or organization name

# Executable configuration (for binary/wasm plugins)
executable: string              # Path to executable (relative to manifest)
entrypoint: string              # Entry function (WASM only)

# Security
permissions:                    # Required permissions
  - type: string                # Permission type
    resource: string            # Resource pattern
    description: string         # Why this permission is needed

# Additional metadata
metadata:                       # Custom key-value pairs
  key: value
  requires:                     # External dependencies
    - dependency-name
  homepage: url
  repository: url
```

### Complete Example

```yaml
name: aws-cli
version: 2.0.0
type: binary
description: Execute AWS CLI operations for cluster management
author: Example Corp

executable: ./aws-cli-plugin
interpreter: /usr/bin/env python3  # Optional: interpreter to use

permissions:
  - type: execute
    resource: aws
    description: Execute AWS CLI commands

  - type: read:env
    resource: AWS_*
    description: Read AWS configuration from environment

  - type: read:file
    resource: ~/.aws/*
    description: Read AWS credentials and configuration files

  - type: network
    resource: "*.amazonaws.com"
    description: Make requests to AWS APIs

metadata:
  homepage: https://github.com/example/aws-cli-plugin
  repository: https://github.com/example/aws-cli-plugin
  requires:
    - aws-cli >= 2.0.0
    - python >= 3.8
  supported_platforms:
    - linux
    - darwin
  tags:
    - aws
    - cloud
    - infrastructure
```

---

## Permission System

Plugins must declare required permissions in their manifest. Users approve permissions on first use.

### Permission Types

| Type | Resource Pattern | Description |
|------|-----------------|-------------|
| `execute` | Command name or `*` | Execute external commands |
| `read:env` | Env var pattern | Read environment variables |
| `write:env` | Env var pattern | Modify environment variables |
| `read:file` | Path pattern | Read files from filesystem |
| `write:file` | Path pattern | Write files to filesystem |
| `network` | Domain pattern | Make network requests |
| `credential` | N/A | Access stored credentials |

### Resource Patterns

**Exact match**:
```yaml
- type: execute
  resource: aws
```

**Wildcard**:
```yaml
- type: read:env
  resource: AWS_*
```

**Multiple paths**:
```yaml
- type: read:file
  resource: ~/.aws/*
```

### Permission Approval Flow

**First execution**:
```
$ mycli cluster create --cluster-name test

⚠️  Plugin 'aws-cli' requests the following permissions:

  Execute Commands:
    • aws (Execute AWS CLI commands)

  Read Files:
    • ~/.aws/* (Read AWS credentials and configuration)

  Read Environment:
    • AWS_* (Read AWS configuration from environment)

  Network Access:
    • *.amazonaws.com (Make requests to AWS APIs)

Grant these permissions? [y/N]: y

✓ Permissions granted and saved

Continuing with cluster creation...
```

**Subsequent executions**: Permissions already approved, no prompt.

### Permission Storage

Approved permissions are stored in:
```
~/.config/<cli-name>/plugin-permissions.yaml
```

Example:
```yaml
plugins:
  aws-cli:
    approved_permissions:
      - execute:aws
      - read:file:~/.aws/*
      - read:env:AWS_*
      - network:*.amazonaws.com
    approved_at: 2025-11-23T10:30:00Z
    approved_by: user@example.com

  dns-checker:
    approved_permissions:
      - network:*
    approved_at: 2025-11-23T11:00:00Z
```

### Revoking Permissions

**Via CLI**:
```bash
mycli plugin revoke aws-cli
```

**Manual**:
Edit or delete `~/.config/<cli-name>/plugin-permissions.yaml`

---

## Security Considerations

### Threat Model

**Threats**:
1. Malicious plugin steals credentials
2. Plugin modifies system files
3. Plugin makes unauthorized network calls
4. Plugin consumes excessive resources
5. Command injection via user input

### Security Controls

#### 1. Sandboxing

**Built-in plugins**: Run in process, trusted by default

**External plugins**: Sandboxed execution:
- Limited environment variables
- Restricted file system access
- Network access filtering
- Resource limits (CPU, memory, time)

**Example sandbox config**:
```yaml
x-cli-config:
  plugins:
    sandbox:
      enabled: true
      allow-network: false
      allowed-paths:
        - ~/.config/mycli
        - /tmp
      env-whitelist:
        - PATH
        - HOME
        - USER
      max-execution-time: 300
      max-memory-mb: 512
```

#### 2. Input Validation

Always validate plugin input to prevent injection:

```python
def validate_input(input_data):
    """Validate plugin input."""
    # Check for null bytes
    for key, value in input_data.items():
        if isinstance(value, str) and '\x00' in value:
            raise ValueError(f"Null byte in input: {key}")

    # Validate file paths
    if 'file' in input_data:
        file_path = input_data['file']
        if '..' in file_path:
            raise ValueError("Path traversal attempt detected")

    return True
```

#### 3. Command Injection Prevention

**Never use shell=True**:

```python
# BAD - vulnerable to injection
os.system(f"aws {user_input}")

# GOOD - safe from injection
subprocess.run(['aws', 'sts', 'get-caller-identity'],
               capture_output=True)
```

**Validate command arguments**:

```python
def safe_execute(command, args):
    """Execute command safely."""
    # Whitelist allowed commands
    allowed_commands = ['aws', 'kubectl', 'git']
    if command not in allowed_commands:
        raise ValueError(f"Command not allowed: {command}")

    # Validate arguments
    for arg in args:
        if any(char in arg for char in [';', '|', '&', '$', '`']):
            raise ValueError(f"Dangerous character in argument: {arg}")

    # Execute safely
    return subprocess.run([command] + args, capture_output=True)
```

#### 4. Path Traversal Prevention

**Validate file paths**:

```python
import os

def validate_path(file_path, allowed_dirs):
    """Validate file path is within allowed directories."""
    abs_path = os.path.abspath(file_path)

    for allowed_dir in allowed_dirs:
        abs_allowed = os.path.abspath(allowed_dir)
        if abs_path.startswith(abs_allowed):
            return abs_path

    raise ValueError(f"Path not allowed: {file_path}")
```

#### 5. Resource Limits

**Set execution timeout**:

```python
import signal

def timeout_handler(signum, frame):
    raise TimeoutError("Plugin execution timeout")

# Set 5-minute timeout
signal.signal(signal.SIGALRM, timeout_handler)
signal.alarm(300)

try:
    result = execute_plugin()
finally:
    signal.alarm(0)  # Cancel timeout
```

**Limit memory usage** (Linux):

```python
import resource

# Limit to 512MB
resource.setrlimit(resource.RLIMIT_AS, (512 * 1024 * 1024, 512 * 1024 * 1024))
```

### Best Practices

1. **Principle of Least Privilege**: Request only necessary permissions
2. **Input Validation**: Validate all user input
3. **Safe Execution**: Never use shell execution with user input
4. **Path Validation**: Prevent path traversal attacks
5. **Resource Limits**: Set timeouts and memory limits
6. **Error Handling**: Don't leak sensitive information in errors
7. **Logging**: Log security-relevant events
8. **Updates**: Keep dependencies up to date

---

## Testing Plugins

### Unit Testing

**Python example**:

```python
import unittest
import json
from plugin import execute

class TestAWSValidator(unittest.TestCase):

    def test_validate_credentials_success(self):
        """Test successful credential validation."""
        params = {
            'command': 'validate-credentials',
            'input': {'profile': 'default'}
        }

        result = execute(params)

        self.assertEqual(result['exit_code'], 0)
        self.assertIn('account_id', result['data'])

    def test_validate_credentials_no_creds(self):
        """Test credential validation with missing credentials."""
        params = {
            'command': 'validate-credentials',
            'input': {'profile': 'nonexistent'}
        }

        result = execute(params)

        self.assertEqual(result['exit_code'], 1)
        self.assertIn('error', result)

if __name__ == '__main__':
    unittest.main()
```

### Integration Testing

Test plugin with actual CliForge CLI:

**Test script**:
```bash
#!/bin/bash

# Install plugin
mycli plugin install ./my-plugin

# Test plugin execution
result=$(mycli test-command --use-plugin)

if [[ $? -eq 0 ]]; then
    echo "✓ Plugin integration test passed"
else
    echo "✗ Plugin integration test failed"
    exit 1
fi

# Cleanup
mycli plugin uninstall my-plugin
```

### JSON-RPC Testing

Test plugin protocol directly:

```bash
#!/bin/bash

# Test valid request
echo '{"jsonrpc":"2.0","id":"1","method":"execute","params":{"command":"test","input":{}}}' | ./plugin.py

# Test invalid JSON
echo 'invalid json' | ./plugin.py

# Test missing parameters
echo '{"jsonrpc":"2.0","id":"2","method":"execute","params":{}}' | ./plugin.py
```

### Manual Testing

```bash
# 1. Install plugin locally
mkdir -p ~/.config/mycli/plugins/test-plugin
cp manifest.yaml ~/.config/mycli/plugins/test-plugin/
cp plugin.py ~/.config/mycli/plugins/test-plugin/

# 2. List plugins
mycli plugin list

# 3. Test plugin execution
mycli command-that-uses-plugin --verbose

# 4. Check plugin logs
tail -f ~/.config/mycli/logs/plugins.log

# 5. Uninstall plugin
mycli plugin uninstall test-plugin
```

---

## Plugin Examples

### Example 1: AWS Integration Plugin

Complete AWS CLI integration for cluster management.

**manifest.yaml**:
```yaml
name: aws-cli
version: 1.0.0
type: binary
description: AWS CLI integration for cluster operations
author: Example Corp
executable: ./aws-plugin.py

permissions:
  - type: execute
    resource: aws
    description: Execute AWS CLI commands
  - type: read:env
    resource: AWS_*
    description: Read AWS configuration
  - type: read:file
    resource: ~/.aws/*
    description: Read AWS credentials

metadata:
  requires:
    - aws-cli >= 2.0.0
    - python >= 3.8
```

**aws-plugin.py**:
```python
#!/usr/bin/env python3
import json
import sys
import subprocess
import boto3

def execute_aws_command(command, args):
    """Execute AWS CLI command."""
    cmd = ['aws'] + command.split() + args
    result = subprocess.run(cmd, capture_output=True, text=True)

    return {
        'exit_code': result.returncode,
        'stdout': result.stdout,
        'stderr': result.stderr
    }

def validate_credentials(input_data):
    """Validate AWS credentials."""
    profile = input_data.get('profile', 'default')

    try:
        session = boto3.Session(profile_name=profile)
        sts = session.client('sts')
        identity = sts.get_caller_identity()

        return {
            'exit_code': 0,
            'data': {
                'account_id': identity['Account'],
                'user_arn': identity['Arn']
            }
        }
    except Exception as e:
        return {
            'exit_code': 1,
            'error': str(e)
        }

def check_quotas(input_data):
    """Check AWS service quotas."""
    service = input_data.get('service')
    region = input_data.get('region')

    # Implementation here
    return {
        'exit_code': 0,
        'data': {
            'quotas': {}
        }
    }

def main():
    for line in sys.stdin:
        try:
            request = json.loads(line)
            params = request.get('params', {})
            command = params.get('command')
            input_data = params.get('input', {})

            if command == 'validate-credentials':
                result = validate_credentials(input_data)
            elif command == 'check-quotas':
                result = check_quotas(input_data)
            elif command == 'execute':
                aws_cmd = input_data.get('aws_command')
                aws_args = input_data.get('args', [])
                result = execute_aws_command(aws_cmd, aws_args)
            else:
                raise ValueError(f"Unknown command: {command}")

            response = {
                'jsonrpc': '2.0',
                'id': request.get('id'),
                'result': result
            }
            print(json.dumps(response))
            sys.stdout.flush()

        except Exception as e:
            error_response = {
                'jsonrpc': '2.0',
                'id': request.get('id'),
                'error': {
                    'code': -32000,
                    'message': str(e)
                }
            }
            print(json.dumps(error_response))
            sys.stdout.flush()

if __name__ == '__main__':
    main()
```

### Example 2: Custom Validator Plugin

Validate cluster names against organization policies.

**cluster-validator.go**:
```go
package main

import (
    "regexp"
    "strings"
)

func validateClusterName(input map[string]interface{}) (interface{}, error) {
    name, ok := input["name"].(string)
    if !ok {
        return nil, fmt.Errorf("name is required")
    }

    // Organization rules:
    // 1. Must start with environment prefix (dev-, staging-, prod-)
    // 2. Must contain team name
    // 3. Must be lowercase with hyphens
    // 4. Max 54 characters

    validFormat := regexp.MustCompile(`^[a-z][a-z0-9-]{0,53}$`)
    if !validFormat.MatchString(name) {
        return map[string]interface{}{
            "exit_code": 1,
            "data": map[string]interface{}{
                "valid": false,
                "message": "Invalid format: must be lowercase with hyphens, max 54 chars",
            },
        }, nil
    }

    // Check environment prefix
    validPrefixes := []string{"dev-", "staging-", "prod-"}
    hasValidPrefix := false
    for _, prefix := range validPrefixes {
        if strings.HasPrefix(name, prefix) {
            hasValidPrefix = true
            break
        }
    }

    if !hasValidPrefix {
        return map[string]interface{}{
            "exit_code": 1,
            "data": map[string]interface{}{
                "valid": false,
                "message": "Must start with: dev-, staging-, or prod-",
            },
        }, nil
    }

    return map[string]interface{}{
        "exit_code": 0,
        "data": map[string]interface{}{
            "valid": true,
            "message": "Cluster name is valid",
        },
    }, nil
}
```

### Example 3: Multi-Cloud Plugin

Support for multiple cloud providers.

**Workflow usage**:
```yaml
- id: validate-cloud-credentials
  type: plugin
  plugin: multi-cloud
  input:
    provider: "{flags.cloud_provider}"  # aws, gcp, azure
    operation: validate-credentials
    credentials:
      aws:
        profile: "{flags.aws_profile}"
      gcp:
        project: "{flags.gcp_project}"
      azure:
        subscription: "{flags.azure_subscription}"
```

---

## Distribution and Installation

### Plugin Distribution

**Option 1: GitHub Releases**

```bash
# Create release
git tag v1.0.0
git push origin v1.0.0

# Users install with:
mycli plugin install github.com/yourname/plugin-name
```

**Option 2: Direct Download**

```bash
# Package plugin
tar -czf my-plugin-v1.0.0.tar.gz manifest.yaml plugin.py

# Users install with:
mycli plugin install https://example.com/my-plugin-v1.0.0.tar.gz
```

**Option 3: Plugin Registry** (future)

```bash
# Publish to registry
mycli plugin publish

# Users install with:
mycli plugin install my-plugin
```

### Installation Methods

**From local directory**:
```bash
mycli plugin install ./my-plugin/
```

**From URL**:
```bash
mycli plugin install https://example.com/plugin.tar.gz
```

**From GitHub**:
```bash
mycli plugin install github.com/user/repo
```

### Plugin Structure

**Required structure**:
```
my-plugin/
├── manifest.yaml          # Required: Plugin manifest
├── plugin.py             # Required: Executable
├── README.md             # Recommended: Documentation
├── LICENSE               # Recommended: License file
└── tests/                # Optional: Test suite
    └── test_plugin.py
```

### Versioning

Follow [Semantic Versioning](https://semver.org/):

- **MAJOR**: Incompatible API changes
- **MINOR**: New functionality (backward compatible)
- **PATCH**: Bug fixes (backward compatible)

**Example**: `1.2.3`
- Major: 1
- Minor: 2
- Patch: 3

---

## Troubleshooting

### Common Issues

#### Plugin Not Found

**Error**:
```
Error: Plugin 'my-plugin' not found
```

**Solutions**:
1. Check plugin is installed: `mycli plugin list`
2. Verify plugin directory: `ls ~/.config/mycli/plugins/`
3. Check manifest.yaml exists and is valid
4. Reinstall plugin

#### Permission Denied

**Error**:
```
Error: Plugin execution failed: permission denied
```

**Solutions**:
1. Make plugin executable: `chmod +x plugin.py`
2. Check file permissions: `ls -l ~/.config/mycli/plugins/my-plugin/`
3. Verify interpreter exists (for scripts): `which python3`

#### JSON-RPC Protocol Error

**Error**:
```
Error: Invalid JSON-RPC response from plugin
```

**Solutions**:
1. Test plugin manually: `echo '{"jsonrpc":"2.0",...}' | ./plugin.py`
2. Check plugin outputs valid JSON
3. Verify all responses include `jsonrpc` and `id` fields
4. Check for extra output (debug prints) on stdout

#### Plugin Timeout

**Error**:
```
Error: Plugin execution timeout (300s)
```

**Solutions**:
1. Optimize plugin performance
2. Increase timeout in workflow:
   ```yaml
   input:
     timeout: 600  # 10 minutes
   ```
3. Check for infinite loops or deadlocks

#### Missing Dependencies

**Error**:
```
Error: ModuleNotFoundError: No module named 'boto3'
```

**Solutions**:
1. Install dependencies: `pip install boto3`
2. Document dependencies in manifest.yaml:
   ```yaml
   metadata:
     requires:
       - boto3 >= 1.20.0
   ```
3. Include installation instructions in README

### Debug Mode

Enable debug logging:

```bash
mycli --debug command-with-plugin
```

**Output**:
```
[DEBUG] Loading plugin: my-plugin
[DEBUG] Plugin manifest: {...}
[DEBUG] Checking permissions...
[DEBUG] Executing plugin: my-plugin
[DEBUG] Plugin request: {"jsonrpc":"2.0",...}
[DEBUG] Plugin response: {"jsonrpc":"2.0",...}
[DEBUG] Plugin execution time: 1.2s
```

### Plugin Logs

Check plugin execution logs:

```bash
tail -f ~/.config/mycli/logs/plugins.log
```

### Testing Plugin Protocol

Test JSON-RPC communication manually:

```bash
# Valid request
cat <<EOF | ./plugin.py
{"jsonrpc":"2.0","id":"1","method":"execute","params":{"command":"test","input":{}}}
EOF

# Expected output:
{"jsonrpc":"2.0","id":"1","result":{"exit_code":0,...}}
```

---

## Summary

Plugins in CliForge v0.9.0 enable you to:

- Extend CLI capabilities beyond OpenAPI specifications
- Execute external tools and scripts securely
- Perform file operations and data transformations
- Implement custom validation and business logic
- Integrate with cloud providers and third-party services

**Key Takeaways**:

1. **Built-in plugins** (`exec`, `file-ops`, `validators`, `transformers`) cover common use cases
2. **External plugins** use JSON-RPC protocol over stdin/stdout
3. **Permission system** protects users from malicious plugins
4. **Sandboxing** provides isolation and security
5. **Manifest format** declares plugin metadata and requirements

**Next Steps**:

- Review built-in plugins for your use case
- Create a simple external plugin
- Test plugin with your CLI
- Document and distribute your plugin

For more information:
- **Workflow Guide**: See `user-guide-workflows.md`
- **Architecture**: See `design/architecture/plugin-architecture.md`
- **API Reference**: See `docs/technical-specification.md`

---

*Forged with CliForge v0.9.0*
