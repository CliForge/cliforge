# CliForge Plugin System

The CliForge plugin system extends CLI capabilities beyond pure OpenAPI specifications, enabling integration with external tools, local file operations, and custom validation logic.

## Architecture

The plugin system consists of several key components:

### Core Components

1. **Plugin Interface** (`plugin.go`)
   - Defines the core `Plugin` interface
   - Types: `PluginInput`, `PluginOutput`, `PluginManifest`
   - Permission model and error handling

2. **Registry** (`registry.go`)
   - Manages plugin registration and discovery
   - Loads external plugins from `~/.config/{cli}/plugins/`
   - Validates plugin manifests

3. **Permission Manager** (`permissions.go`)
   - Handles permission approval workflow
   - Stores approved permissions in `~/.config/{cli}/plugin-permissions.yaml`
   - Validates permission requests

4. **Executor** (`executor.go`)
   - Executes plugins with timeout handling
   - Supports retry logic and parallel execution
   - Collects execution statistics

### Built-in Plugins

Located in `builtin/` directory:

#### 1. Exec Plugin (`exec.go`)
Executes external command-line tools with sandboxing.

**Capabilities:**
- Execute allowed commands
- Capture stdout/stderr
- Set environment variables
- Custom working directory

**Example:**
```go
execPlugin := builtin.NewExecPlugin([]string{"aws", "kubectl"}, true)
input := &plugin.PluginInput{
    Command: "aws",
    Args:    []string{"sts", "get-caller-identity"},
}
output, err := execPlugin.Execute(ctx, input)
```

#### 2. File Operations Plugin (`file_ops.go`)
Read, parse, validate, and transform files.

**Supported Formats:**
- JSON
- YAML
- PEM (X.509 certificates)
- HTPasswd

**Operations:**
- `read`: Read file contents
- `parse`: Parse structured files
- `validate`: Validate file format
- `transform`: Apply transformations

**Example:**
```go
fileOps := builtin.NewFileOpsPlugin([]string{"/allowed/path"}, 10*1024*1024)
input := &plugin.PluginInput{
    Data: map[string]interface{}{
        "operation": "parse",
        "file":      "/path/to/cert.pem",
        "format":    "pem",
    },
}
output, err := fileOps.Execute(ctx, input)
```

#### 3. Validators Plugin (`validators.go`)
Custom validation logic for various data types.

**Validators:**
- `regex`: Regular expression matching
- `email`: Email address validation
- `url`: URL validation
- `ip`: IP address validation
- `cidr`: CIDR notation validation
- `cluster-name`: Kubernetes-style cluster names
- `dns-label`: DNS label (RFC 1123)
- `length`: String length validation
- `range`: Numeric range validation
- `enum`: Allowed values validation
- `format`: Common formats (UUID, date, semver, etc.)

**Example:**
```go
validators := builtin.NewValidatorsPlugin()
input := &plugin.PluginInput{
    Data: map[string]interface{}{
        "validator": "email",
        "value":     "user@example.com",
    },
}
output, err := validators.Execute(ctx, input)
```

#### 4. Transformers Plugin (`transformers.go`)
Data transformation utilities.

**Transformations:**
- `json-to-yaml`: Convert JSON to YAML
- `yaml-to-json`: Convert YAML to JSON
- `base64-encode`: Base64 encoding
- `base64-decode`: Base64 decoding
- `htpasswd-to-users`: Parse htpasswd file
- `users-to-htpasswd`: Generate htpasswd file
- `extract-field`: Extract field from structured data
- `merge`: Merge multiple objects
- `filter`: Filter data by criteria
- `template`: Simple template substitution

**Example:**
```go
transformers := builtin.NewTransformersPlugin()
input := &plugin.PluginInput{
    Data: map[string]interface{}{
        "transformation": "json-to-yaml",
        "input":          `{"key": "value"}`,
    },
}
output, err := transformers.Execute(ctx, input)
```

## Permission System

### Permission Types

- `execute:<command>` - Execute external commands
- `read:env:<pattern>` - Read environment variables
- `write:env:<pattern>` - Write environment variables
- `read:file:<path>` - Read files
- `write:file:<path>` - Write files
- `network:<domain>` - Make network requests
- `credential` - Access stored credentials

### Permission Workflow

1. Plugin declares required permissions in manifest
2. On first use, user is prompted to approve permissions
3. Approved permissions are stored in `~/.config/{cli}/plugin-permissions.yaml`
4. Subsequent executions check against approved permissions

### Built-in Plugins

Built-in plugins (`exec`, `file-ops`, `validators`, `transformers`) are trusted by default and don't require user approval.

## Usage Examples

### Basic Plugin Registration

```go
import (
    "github.com/CliForge/cliforge/pkg/plugin"
    "github.com/CliForge/cliforge/pkg/plugin/builtin"
)

// Create permission manager
pm, err := plugin.NewPermissionManager(configDir, &plugin.DefaultApprover{})

// Create registry
registry := plugin.NewRegistry(pluginDir, pm)

// Register built-in plugins
registry.Register(builtin.NewExecPlugin(nil, false))
registry.Register(builtin.NewFileOpsPlugin(nil, 0))
registry.Register(builtin.NewValidatorsPlugin())
registry.Register(builtin.NewTransformersPlugin())

// Discover external plugins
registry.DiscoverPlugins()
```

### Executing a Plugin

```go
ctx := context.Background()
input := &plugin.PluginInput{
    Data: map[string]interface{}{
        "validator": "email",
        "value":     "test@example.com",
    },
}

output, err := registry.Execute(ctx, "validators", input)
if err != nil {
    log.Fatalf("Plugin execution failed: %v", err)
}

if output.Success() {
    fmt.Println("Validation passed!")
}
```

### Using the Executor

```go
executor := plugin.NewExecutor(30*time.Second, 5*time.Minute)

// Execute with timeout
output, err := executor.Execute(ctx, myPlugin, input)

// Execute with retry
output, err = executor.ExecuteWithRetry(ctx, myPlugin, input, 3)
```

## External Plugins

External plugins are loaded from `~/.config/{cli}/plugins/`. Each plugin must have a `plugin-manifest.yaml` file.

### Plugin Manifest Example

```yaml
name: my-custom-plugin
version: 1.0.0
type: binary
executable: ./plugin-binary
permissions:
  - type: execute
    resource: custom-tool
    description: Execute custom tool
  - type: read:file
    resource: /data/*
    description: Read data files
metadata:
  author: Your Name
  description: Custom plugin for special operations
```

### Binary Plugin Protocol

Binary plugins communicate via JSON-RPC over stdin/stdout:

**Request:**
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "method": "execute",
  "params": {
    "command": "aws",
    "args": ["sts", "get-caller-identity"],
    "env": {"AWS_PROFILE": "default"}
  }
}
```

**Response:**
```json
{
  "jsonrpc": "2.0",
  "id": "1",
  "result": {
    "stdout": "{\"UserId\": \"...\"}",
    "stderr": "",
    "exit_code": 0,
    "data": {}
  }
}
```

## Testing

Run all plugin tests:

```bash
go test -v ./pkg/plugin/...
```

Run specific test suite:

```bash
go test -v ./pkg/plugin/builtin/...
```

Run with coverage:

```bash
go test -cover ./pkg/plugin/...
```

## Security Considerations

1. **Built-in plugins** are trusted and compiled into the binary
2. **External plugins** require explicit user approval for permissions
3. **Sandboxing** limits resource access for exec plugin
4. **Validation** prevents command injection attacks
5. **File access** can be restricted to specific paths
6. **Wildcard matching** supports flexible permission patterns

## Future Enhancements

### v1.0.0 (Planned)

- **WASM plugins**: Secure, portable plugins using WebAssembly
- **Plugin signing**: Cryptographic verification of plugin authors
- **Advanced sandboxing**: chroot jails, namespaces, AppContainer
- **Plugin marketplace**: Centralized registry for discovering plugins
- **Plugin SDK**: Development tools and templates

### v2.0.0 (Planned)

- **Embedded scripting**: Lua or Starlark for simple plugins
- **gRPC plugins**: HashiCorp-style plugin architecture
- **Plugin versioning**: Compatibility checks and automatic updates
- **Performance monitoring**: Detailed execution metrics
- **Plugin dependencies**: Allow plugins to depend on other plugins

## References

- **Design Document**: `design/architecture/plugin-architecture.md`
- **OpenAPI Extensions**: Plugins integrate with `x-cli-workflow` steps
- **Configuration**: Uses existing config package for paths and settings
