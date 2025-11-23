package runtime

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/spf13/cobra"
)

// CommandBuilder builds Cobra commands from OpenAPI operations.
type CommandBuilder struct {
	runtime *Runtime
}

// NewCommandBuilder creates a new CommandBuilder.
func NewCommandBuilder(rt *Runtime) *CommandBuilder {
	return &CommandBuilder{runtime: rt}
}

// BuildCommand builds a Cobra command from an OpenAPI operation.
func (cb *CommandBuilder) BuildCommand(op *openapi.Operation) (*cobra.Command, error) {
	// Determine command name
	cmdName := op.CLICommand
	if cmdName == "" {
		cmdName = generateCommandName(op.OperationID, op.Method, op.Path)
	}

	// Parse command structure (handle nested commands)
	cmdParts := strings.Fields(cmdName)
	if len(cmdParts) == 0 {
		return nil, fmt.Errorf("empty command name")
	}

	// Create command
	cmd := &cobra.Command{
		Use:   cmdParts[len(cmdParts)-1],
		Short: op.Summary,
		Long:  op.Description,
		RunE: func(cmd *cobra.Command, args []string) error {
			return cb.executeOperation(cmd.Context(), op, cmd, args)
		},
	}

	// Add aliases
	if len(op.CLIAliases) > 0 {
		cmd.Aliases = op.CLIAliases
	}

	// Add flags from operation
	cb.addOperationFlags(cmd, op)

	// Handle nested commands (e.g., "create cluster" -> create > cluster)
	if len(cmdParts) > 1 {
		// For now, return as flat command
		// TODO: Implement nested command structure
		cmd.Use = strings.Join(cmdParts, " ")
	}

	return cmd, nil
}

// addOperationFlags adds flags to a command based on the operation.
func (cb *CommandBuilder) addOperationFlags(cmd *cobra.Command, op *openapi.Operation) {
	// Add flags from x-cli-flags extension
	for _, flag := range op.CLIFlags {
		flagName := strings.TrimPrefix(flag.Flag, "--")

		switch flag.Type {
		case "string":
			var defaultVal string
			if flag.Default != nil {
				if s, ok := flag.Default.(string); ok {
					defaultVal = s
				}
			}
			cmd.Flags().String(flagName, defaultVal, flag.Description)
			if flag.Required {
				cmd.MarkFlagRequired(flagName)
			}

		case "int", "integer":
			var defaultVal int
			if flag.Default != nil {
				if i, ok := flag.Default.(int); ok {
					defaultVal = i
				}
			}
			cmd.Flags().Int(flagName, defaultVal, flag.Description)
			if flag.Required {
				cmd.MarkFlagRequired(flagName)
			}

		case "bool", "boolean":
			var defaultVal bool
			if flag.Default != nil {
				if b, ok := flag.Default.(bool); ok {
					defaultVal = b
				}
			}
			cmd.Flags().Bool(flagName, defaultVal, flag.Description)
		}
	}

	// Parse parameters from operation
	if op.Operation != nil && op.Operation.Parameters != nil {
		for _, paramRef := range op.Operation.Parameters {
			if paramRef.Value == nil {
				continue
			}
			param := paramRef.Value

			// Generate flag name
			flagName := strings.ReplaceAll(param.Name, "_", "-")

			// Get CLI flag from extension
			var cliFlag string
			if ext, ok := param.Extensions["x-cli-flag"].(string); ok {
				cliFlag = strings.TrimPrefix(ext, "--")
			} else {
				cliFlag = flagName
			}

			// Get description
			desc := param.Description

			// Add flag based on schema type
			if param.Schema != nil && param.Schema.Value != nil {
				schema := param.Schema.Value
				switch schema.Type.Slice()[0] {
				case "string":
					defaultVal := ""
					if schema.Default != nil {
						if s, ok := schema.Default.(string); ok {
							defaultVal = s
						}
					}
					cmd.Flags().String(cliFlag, defaultVal, desc)
					if param.Required {
						cmd.MarkFlagRequired(cliFlag)
					}

				case "integer":
					defaultVal := 0
					if schema.Default != nil {
						if f, ok := schema.Default.(float64); ok {
							defaultVal = int(f)
						}
					}
					cmd.Flags().Int(cliFlag, defaultVal, desc)
					if param.Required {
						cmd.MarkFlagRequired(cliFlag)
					}

				case "boolean":
					defaultVal := false
					if schema.Default != nil {
						if b, ok := schema.Default.(bool); ok {
							defaultVal = b
						}
					}
					cmd.Flags().Bool(cliFlag, defaultVal, desc)
				}
			}
		}
	}
}

// executeOperation executes an API operation.
func (cb *CommandBuilder) executeOperation(ctx context.Context, op *openapi.Operation, cmd *cobra.Command, args []string) error {
	// Check if this is a workflow operation
	if op.CLIWorkflow != nil {
		return cb.executeWorkflow(ctx, op, cmd)
	}

	// Build HTTP request
	req, err := cb.buildRequest(ctx, op, cmd, args)
	if err != nil {
		return fmt.Errorf("failed to build request: %w", err)
	}

	// Add authentication
	if cb.runtime.authManager != nil {
		token, err := cb.runtime.authManager.GetToken(ctx, "")
		if err != nil {
			return fmt.Errorf("authentication failed: %w", err)
		}
		// Get authenticator and add headers
		auth, err := cb.runtime.authManager.GetAuthenticator("")
		if err != nil {
			return fmt.Errorf("failed to get authenticator: %w", err)
		}
		headers := auth.GetHeaders(token)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	// Execute request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	// Handle response
	return cb.handleResponse(ctx, op, resp, cmd)
}

// buildRequest builds an HTTP request from the operation and flags.
func (cb *CommandBuilder) buildRequest(ctx context.Context, op *openapi.Operation, cmd *cobra.Command, args []string) (*http.Request, error) {
	// Build URL
	url := cb.runtime.config.API.BaseURL + op.Path

	// Replace path parameters
	// TODO: Implement path parameter replacement from flags

	// Create request
	req, err := http.NewRequestWithContext(ctx, op.Method, url, nil)
	if err != nil {
		return nil, err
	}

	// Add default headers
	if cb.runtime.config.API.DefaultHeaders != nil {
		for k, v := range cb.runtime.config.API.DefaultHeaders {
			req.Header.Set(k, v)
		}
	}

	// Add user agent
	userAgent := cb.runtime.config.API.UserAgent
	if userAgent == "" {
		userAgent = fmt.Sprintf("%s/%s", cb.runtime.config.Metadata.Name, cb.runtime.version)
	}
	req.Header.Set("User-Agent", userAgent)

	// TODO: Add request body from flags for POST/PUT/PATCH
	// TODO: Add query parameters from flags

	return req, nil
}

// handleResponse handles the HTTP response and outputs the result.
func (cb *CommandBuilder) handleResponse(ctx context.Context, op *openapi.Operation, resp *http.Response, cmd *cobra.Command) error {
	// Check status code
	if resp.StatusCode >= 400 {
		return fmt.Errorf("HTTP %d: %s", resp.StatusCode, resp.Status)
	}

	// Output response - simple implementation for now
	// Future: use output manager for formatting
	defer resp.Body.Close()
	_, err := fmt.Fprintf(os.Stdout, "Success: HTTP %d\n", resp.StatusCode)
	return err
}

// executeWorkflow executes a workflow operation.
func (cb *CommandBuilder) executeWorkflow(ctx context.Context, op *openapi.Operation, cmd *cobra.Command) error {
	// Workflow execution not yet implemented
	return fmt.Errorf("workflow execution not yet implemented")
}

// generateCommandName generates a command name from operation metadata.
func generateCommandName(operationID, method, path string) string {
	if operationID != "" {
		// Convert camelCase to kebab-case
		return camelToKebab(operationID)
	}

	// Fallback to method + path
	parts := strings.Split(strings.Trim(path, "/"), "/")
	if len(parts) > 0 {
		resource := parts[len(parts)-1]
		resource = strings.ReplaceAll(resource, "{", "")
		resource = strings.ReplaceAll(resource, "}", "")

		switch strings.ToLower(method) {
		case "get":
			if strings.Contains(path, "{") {
				return "get " + resource
			}
			return "list " + resource
		case "post":
			return "create " + resource
		case "put", "patch":
			return "update " + resource
		case "delete":
			return "delete " + resource
		}
	}

	return strings.ToLower(method)
}

// camelToKebab converts camelCase to kebab-case.
func camelToKebab(s string) string {
	var result strings.Builder
	for i, r := range s {
		if i > 0 && r >= 'A' && r <= 'Z' {
			result.WriteRune('-')
		}
		result.WriteRune(r)
	}
	return strings.ToLower(result.String())
}
