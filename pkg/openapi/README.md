# OpenAPI Package

Complete OpenAPI 3.x and Swagger 2.0 parser with full x-cli-* extension support for CliForge v0.9.0.

## Features

- **Full Spec Support**: Parse OpenAPI 3.0, 3.1, and Swagger 2.0 specifications
- **Auto-Conversion**: Automatically converts Swagger 2.0 to OpenAPI 3.0
- **Complete Extension Support**: Parses all x-cli-* extensions defined in the spec
- **Validation**: Validates specs and extensions with detailed error reporting
- **Change Detection**: Detects API changes between spec versions
- **Caching**: XDG-compliant caching with ETag support

## Components

### Parser (`parser.go`)
Parses OpenAPI specs and extracts CLI extensions.

```go
parser := openapi.NewParser()
parsed, err := parser.Parse(ctx, specData)
```

### Loader (`loader.go`)
Loads specs from URLs with caching.

```go
cache, _ := cache.NewSpecCache("myapp")
loader := openapi.NewLoader(cache)
parsed, err := loader.LoadFromURL(ctx, "https://api.example.com/openapi.yaml", nil)
```

### Extensions (`extensions.go`)
Parses all x-cli-* extensions:

- `x-cli-config` - Global CLI configuration
- `x-cli-command` - Command name mapping
- `x-cli-flags` - Flag definitions
- `x-cli-interactive` - Interactive prompts
- `x-cli-preflight` - Pre-execution checks
- `x-cli-confirmation` - Confirmation prompts
- `x-cli-async` - Async operation polling
- `x-cli-output` - Output formatting
- `x-cli-workflow` - Multi-step workflows
- `x-cli-plugin` - Plugin integration
- `x-cli-file-input` - File input handling
- `x-cli-watch` - Watch mode
- `x-cli-deprecation` - Deprecation notices
- `x-cli-secret` - Secret handling
- `x-auth-config` - Auth configuration

### Validator (`validator.go`)
Validates specs and extensions.

```go
validator := openapi.NewValidator()
result, err := validator.Validate(ctx, parsed)
```

### Change Detector (`changelog.go`)
Detects changes between spec versions.

```go
detector := openapi.NewChangeDetector()
changes, err := detector.DetectChanges(oldSpec, newSpec)
```

## Cache Package

XDG-compliant spec caching with ETag support.

```go
cache, err := cache.NewSpecCache("myapp")
```

**Features**:
- XDG Base Directory compliance (Linux/macOS/Windows)
- 5-minute default TTL
- ETag/Last-Modified support
- Cache invalidation and pruning
- Statistics tracking

## Testing

All components have comprehensive test coverage:

```bash
go test ./pkg/openapi/... -v
go test ./pkg/cache/... -v
```

## Examples

See `examples/openapi/`:
- `complete-example.yaml` - Full OpenAPI 3.0 spec with all extensions
- `swagger2-example.json` - Swagger 2.0 compatibility example

## Extension Specification

See `hack/rosa-cli/openapi-cli-extensions-spec.md` for complete extension documentation.
