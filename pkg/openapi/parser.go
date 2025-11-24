// Package openapi provides OpenAPI 3.x and Swagger 2.0 specification parsing with CLI extensions support.
//
// The openapi package parses and validates OpenAPI/Swagger specifications,
// automatically converting Swagger 2.0 to OpenAPI 3.0, and extracts all
// x-cli-* extensions for CLI generation. It supports loading specs from files,
// URLs, and embedded resources.
//
// # Supported Formats
//
//   - OpenAPI 3.0.x (JSON/YAML)
//   - OpenAPI 3.1.x (JSON/YAML)
//   - Swagger 2.0 (auto-converted to OpenAPI 3.0)
//
// # CLI Extensions
//
// The parser extracts all x-cli-* extensions:
//
//   - x-cli-command: Override command name
//   - x-cli-aliases: Command aliases
//   - x-cli-hidden: Hide from help output
//   - x-cli-examples: Usage examples
//   - x-cli-auth: Authentication requirements
//   - x-cli-workflow: Multi-step workflows
//   - x-cli-async: Async operation polling
//   - x-cli-output: Output formatting
//   - x-cli-confirmation: Confirm before execution
//   - x-cli-preflight: Pre-execution checks
//   - x-cli-secret: Mark field as sensitive
//   - x-cli-deprecation: Deprecation metadata
//
// # Example Usage
//
//	// Create parser
//	parser := openapi.NewParser()
//
//	// Parse from file
//	spec, err := parser.ParseFile(ctx, "openapi.yaml")
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get API info
//	info := spec.GetInfo()
//	fmt.Printf("API: %s v%s\n", info.Title, info.Version)
//
//	// Get operations
//	operations, _ := spec.GetOperations()
//	for _, op := range operations {
//	    fmt.Printf("%s %s - %s\n", op.Method, op.Path, op.Summary)
//	}
//
// # Swagger 2.0 Conversion
//
// Swagger 2.0 specs are automatically detected and converted:
//
//	spec, _ := parser.Parse(ctx, swaggerData)
//	// Returns OpenAPI 3.0 ParsedSpec
//	fmt.Println(spec.OriginalVersion) // "2.0"
//
// # Validation
//
// By default, specs are validated during parsing. Disable with:
//
//	parser := openapi.NewParser()
//	parser.DisableValidation = true
//
// The parser supports remote $ref resolution when enabled:
//
//	parser.AllowRemoteRefs = true
//
// All extensions are fully documented at:
// https://github.com/CliForge/cliforge/blob/main/docs/openapi-extensions.md
package openapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/getkin/kin-openapi/openapi2"
	"github.com/getkin/kin-openapi/openapi2conv"
	"github.com/getkin/kin-openapi/openapi3"
)

// Parser handles parsing of OpenAPI 3.x and Swagger 2.0 specifications.
type Parser struct {
	// DisableValidation skips OpenAPI spec validation
	DisableValidation bool
	// AllowRemoteRefs enables loading remote $ref references
	AllowRemoteRefs bool
}

// NewParser creates a new Parser instance with default settings.
func NewParser() *Parser {
	return &Parser{
		DisableValidation: false,
		AllowRemoteRefs:   false,
	}
}

// ParsedSpec represents a parsed OpenAPI specification with metadata.
type ParsedSpec struct {
	// Spec is the OpenAPI 3.x specification
	Spec *openapi3.T
	// OriginalVersion indicates the original spec format
	OriginalVersion SpecVersion
	// Extensions contains all parsed x-cli-* extensions
	Extensions *SpecExtensions
}

// SpecVersion indicates the OpenAPI specification version.
type SpecVersion string

const (
	// SpecVersionSwagger2 represents Swagger 2.0 / OpenAPI 2.0
	SpecVersionSwagger2 SpecVersion = "2.0"
	// SpecVersionOpenAPI3 represents OpenAPI 3.0.x
	SpecVersionOpenAPI3 SpecVersion = "3.0"
	// SpecVersionOpenAPI31 represents OpenAPI 3.1.x
	SpecVersionOpenAPI31 SpecVersion = "3.1"
)

// Parse parses an OpenAPI specification from a byte slice.
// Automatically detects Swagger 2.0 and converts to OpenAPI 3.0.
func (p *Parser) Parse(ctx context.Context, data []byte) (*ParsedSpec, error) {
	// Try to detect version
	version, err := p.detectVersion(data)
	if err != nil {
		return nil, fmt.Errorf("failed to detect spec version: %w", err)
	}

	var spec *openapi3.T

	switch version {
	case SpecVersionSwagger2:
		spec, err = p.parseSwagger2(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse Swagger 2.0 spec: %w", err)
		}
	case SpecVersionOpenAPI3, SpecVersionOpenAPI31:
		spec, err = p.parseOpenAPI3(ctx, data)
		if err != nil {
			return nil, fmt.Errorf("failed to parse OpenAPI 3.x spec: %w", err)
		}
	default:
		return nil, fmt.Errorf("unsupported spec version: %s", version)
	}

	// Parse extensions
	extensions, err := ParseSpecExtensions(spec)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CLI extensions: %w", err)
	}

	return &ParsedSpec{
		Spec:            spec,
		OriginalVersion: version,
		Extensions:      extensions,
	}, nil
}

// ParseFile parses an OpenAPI specification from a file.
func (p *Parser) ParseFile(ctx context.Context, path string) (*ParsedSpec, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read file: %w", err)
	}
	return p.Parse(ctx, data)
}

// ParseReader parses an OpenAPI specification from an io.Reader.
func (p *Parser) ParseReader(ctx context.Context, r io.Reader) (*ParsedSpec, error) {
	data, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("failed to read spec: %w", err)
	}
	return p.Parse(ctx, data)
}

// detectVersion detects the OpenAPI specification version.
func (p *Parser) detectVersion(data []byte) (SpecVersion, error) {
	// Quick detection using basic JSON parsing
	var versionCheck struct {
		Swagger string `json:"swagger"`
		OpenAPI string `json:"openapi"`
	}

	if err := json.Unmarshal(data, &versionCheck); err != nil {
		return "", fmt.Errorf("failed to parse spec: %w", err)
	}

	if versionCheck.Swagger != "" {
		if strings.HasPrefix(versionCheck.Swagger, "2.") {
			return SpecVersionSwagger2, nil
		}
		return "", fmt.Errorf("unsupported swagger version: %s", versionCheck.Swagger)
	}

	if versionCheck.OpenAPI != "" {
		if strings.HasPrefix(versionCheck.OpenAPI, "3.0.") {
			return SpecVersionOpenAPI3, nil
		}
		if strings.HasPrefix(versionCheck.OpenAPI, "3.1.") {
			return SpecVersionOpenAPI31, nil
		}
		return "", fmt.Errorf("unsupported openapi version: %s", versionCheck.OpenAPI)
	}

	return "", fmt.Errorf("could not determine spec version (missing 'swagger' or 'openapi' field)")
}

// parseSwagger2 parses a Swagger 2.0 spec and converts it to OpenAPI 3.0.
func (p *Parser) parseSwagger2(ctx context.Context, data []byte) (*openapi3.T, error) {
	var spec2 openapi2.T
	if err := json.Unmarshal(data, &spec2); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Swagger 2.0: %w", err)
	}

	// Convert to OpenAPI 3.0
	spec3, err := openapi2conv.ToV3(&spec2)
	if err != nil {
		return nil, fmt.Errorf("failed to convert Swagger 2.0 to OpenAPI 3.0: %w", err)
	}

	// Validate if enabled
	if !p.DisableValidation {
		if err := spec3.Validate(ctx); err != nil {
			return nil, fmt.Errorf("converted spec validation failed: %w", err)
		}
	}

	return spec3, nil
}

// parseOpenAPI3 parses an OpenAPI 3.x specification.
func (p *Parser) parseOpenAPI3(ctx context.Context, data []byte) (*openapi3.T, error) {
	loader := openapi3.NewLoader()
	loader.IsExternalRefsAllowed = p.AllowRemoteRefs

	spec, err := loader.LoadFromData(data)
	if err != nil {
		return nil, fmt.Errorf("failed to load OpenAPI 3.x: %w", err)
	}

	// Validate if enabled
	if !p.DisableValidation {
		if err := spec.Validate(ctx); err != nil {
			return nil, fmt.Errorf("spec validation failed: %w", err)
		}
	}

	return spec, nil
}

// GetInfo returns basic information about the API from the spec.
func (ps *ParsedSpec) GetInfo() *SpecInfo {
	info := &SpecInfo{
		Title:       ps.Spec.Info.Title,
		Version:     ps.Spec.Info.Version,
		Description: ps.Spec.Info.Description,
	}

	// Extract x-cli-version if present
	if cliVersion, ok := ps.Spec.Info.Extensions["x-cli-version"].(string); ok {
		info.CLIVersion = cliVersion
	}

	// Extract x-cli-min-version if present
	if minVersion, ok := ps.Spec.Info.Extensions["x-cli-min-version"].(string); ok {
		info.MinCLIVersion = minVersion
	}

	return info
}

// SpecInfo contains basic information about the API specification.
type SpecInfo struct {
	Title         string
	Version       string
	Description   string
	CLIVersion    string
	MinCLIVersion string
}

// GetOperations returns all operations from the spec with their CLI extensions.
func (ps *ParsedSpec) GetOperations() ([]*Operation, error) {
	var operations []*Operation

	for path, pathItem := range ps.Spec.Paths.Map() {
		for method, operation := range pathItem.Operations() {
			if operation == nil {
				continue
			}

			// Check if operation is hidden
			if hidden, ok := operation.Extensions["x-cli-hidden"].(bool); ok && hidden {
				continue
			}

			op := &Operation{
				Method:      method,
				Path:        path,
				OperationID: operation.OperationID,
				Summary:     operation.Summary,
				Description: operation.Description,
				Tags:        operation.Tags,
				Operation:   operation,
			}

			// Parse operation extensions
			if err := parseOperationExtensions(operation, op); err != nil {
				return nil, fmt.Errorf("failed to parse extensions for %s %s: %w", method, path, err)
			}

			operations = append(operations, op)
		}
	}

	return operations, nil
}

// Operation represents an API operation with CLI extensions.
type Operation struct {
	Method      string
	Path        string
	OperationID string
	Summary     string
	Description string
	Tags        []string
	Operation   *openapi3.Operation

	// CLI Extensions
	CLICommand      string
	CLIAliases      []string
	CLIFlags        []*CLIFlag
	CLIInteractive  *CLIInteractive
	CLIPreflight    []*CLIPreflight
	CLIConfirmation *CLIConfirmation
	CLIAsync        *CLIAsync
	CLIOutput       *CLIOutput
	CLIWorkflow     *CLIWorkflow
	CLIParentRes    string
}

// parseOperationExtensions extracts CLI extensions from an operation.
func parseOperationExtensions(operation *openapi3.Operation, op *Operation) error {
	// x-cli-command
	if cmd, ok := operation.Extensions["x-cli-command"].(string); ok {
		op.CLICommand = cmd
	}

	// x-cli-aliases
	if aliases, ok := operation.Extensions["x-cli-aliases"].([]interface{}); ok {
		for _, alias := range aliases {
			if str, ok := alias.(string); ok {
				op.CLIAliases = append(op.CLIAliases, str)
			}
		}
	}

	// x-cli-parent-resource
	if parent, ok := operation.Extensions["x-cli-parent-resource"].(string); ok {
		op.CLIParentRes = parent
	}

	// x-cli-flags
	if flags, ok := operation.Extensions["x-cli-flags"].([]interface{}); ok {
		parsedFlags, err := parseCLIFlags(flags)
		if err != nil {
			return fmt.Errorf("failed to parse x-cli-flags: %w", err)
		}
		op.CLIFlags = parsedFlags
	}

	// x-cli-interactive
	if interactive, ok := operation.Extensions["x-cli-interactive"].(map[string]interface{}); ok {
		parsed, err := parseCLIInteractive(interactive)
		if err != nil {
			return fmt.Errorf("failed to parse x-cli-interactive: %w", err)
		}
		op.CLIInteractive = parsed
	}

	// x-cli-preflight
	if preflight, ok := operation.Extensions["x-cli-preflight"].([]interface{}); ok {
		parsed, err := parseCLIPreflight(preflight)
		if err != nil {
			return fmt.Errorf("failed to parse x-cli-preflight: %w", err)
		}
		op.CLIPreflight = parsed
	}

	// x-cli-confirmation
	if confirmation, ok := operation.Extensions["x-cli-confirmation"].(map[string]interface{}); ok {
		parsed, err := parseCLIConfirmation(confirmation)
		if err != nil {
			return fmt.Errorf("failed to parse x-cli-confirmation: %w", err)
		}
		op.CLIConfirmation = parsed
	}

	// x-cli-async
	if async, ok := operation.Extensions["x-cli-async"].(map[string]interface{}); ok {
		parsed, err := parseCLIAsync(async)
		if err != nil {
			return fmt.Errorf("failed to parse x-cli-async: %w", err)
		}
		op.CLIAsync = parsed
	}

	// x-cli-output
	if output, ok := operation.Extensions["x-cli-output"].(map[string]interface{}); ok {
		parsed, err := parseCLIOutput(output)
		if err != nil {
			return fmt.Errorf("failed to parse x-cli-output: %w", err)
		}
		op.CLIOutput = parsed
	}

	// x-cli-workflow
	if workflow, ok := operation.Extensions["x-cli-workflow"].(map[string]interface{}); ok {
		parsed, err := parseCLIWorkflow(workflow)
		if err != nil {
			return fmt.Errorf("failed to parse x-cli-workflow: %w", err)
		}
		op.CLIWorkflow = parsed
	}

	return nil
}
