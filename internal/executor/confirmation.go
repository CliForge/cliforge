// Package executor provides confirmation prompt functionality for destructive operations.
//
// The confirmation system reads the x-cli-confirmation extension from OpenAPI specs
// and prompts users before executing potentially dangerous operations like deletions.
//
// # Features
//
//   - Automatic confirmation prompts for operations marked with x-cli-confirmation
//   - Bypass flag support (e.g., --yes, --force) to skip prompts in automation
//   - Parameter substitution in confirmation messages (e.g., "Delete {cluster_id}?")
//   - Support for kebab-case, camelCase, and snake_case parameter names
//   - Styled terminal output using pterm for clear visual feedback
//   - Optional explicit confirmation requiring typing "yes" for highly destructive operations
//
// # Example OpenAPI Configuration
//
//	delete:
//	  operationId: deleteCluster
//	  x-cli-confirmation:
//	    enabled: true
//	    message: "Are you sure you want to delete cluster '{cluster_id}'?"
//	    flag: "--yes"
//
// # Usage
//
// The executor automatically checks for confirmation before executing operations.
// Users can bypass prompts with:
//
//	mycli delete cluster my-cluster --yes
//
// Or they will be prompted interactively:
//
//	┌─ Confirmation Required ─┐
//	│ Delete cluster 'my-cluster'? │
//	└─────────────────────────┘
//	Do you want to continue? [y/N]:
package executor

import (
	"fmt"
	"strings"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/pterm/pterm"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// ConfirmationOptions contains options for confirmation prompts.
type ConfirmationOptions struct {
	// Config from the OpenAPI spec
	Config *openapi.CLIConfirmation
	// Parameters to use for template substitution in the message
	Parameters map[string]interface{}
}

// CheckConfirmation checks if confirmation is required and prompts the user.
// Returns true if the operation should proceed, false if it should be canceled.
func (e *Executor) CheckConfirmation(cmd *cobra.Command, op *openapi.Operation) (bool, error) {
	// If no confirmation config, proceed
	if op.CLIConfirmation == nil || !op.CLIConfirmation.Enabled {
		return true, nil
	}

	// Check for bypass flag
	flagName := op.CLIConfirmation.Flag
	if flagName == "" {
		flagName = "--yes"
	}

	// Remove leading dashes if present
	flagName = strings.TrimPrefix(flagName, "--")
	flagName = strings.TrimPrefix(flagName, "-")

	// Check if the bypass flag is set
	if cmd.Flags().Changed(flagName) {
		if bypass, _ := cmd.Flags().GetBool(flagName); bypass {
			return true, nil
		}
	}

	// Build the confirmation message
	message := op.CLIConfirmation.Message
	if message == "" {
		message = "Are you sure you want to proceed with this operation?"
	}

	// Substitute parameters in the message
	message = e.substituteParameters(cmd, message)

	// Show confirmation prompt
	return ShowConfirmationPrompt(message)
}

// ShowConfirmationPrompt displays a confirmation prompt to the user.
// Returns true if the user confirms, false otherwise.
func ShowConfirmationPrompt(message string) (bool, error) {
	// Add warning indicator for destructive operations
	styledMessage := pterm.DefaultBox.
		WithTitle("Confirmation Required").
		WithTitleTopCenter().
		WithBoxStyle(pterm.NewStyle(pterm.FgYellow)).
		Sprint(message)

	pterm.Println()
	pterm.Println(styledMessage)
	pterm.Println()

	// Use pterm's interactive confirmation
	confirmed, err := pterm.DefaultInteractiveConfirm.
		WithDefaultText("Do you want to continue?").
		WithDefaultValue(false).
		Show()

	if err != nil {
		return false, fmt.Errorf("confirmation prompt failed: %w", err)
	}

	if !confirmed {
		pterm.Info.Println("Operation canceled by user")
	}

	return confirmed, nil
}

// ShowExplicitConfirmationPrompt requires the user to type "yes" to confirm.
// Returns true if the user types "yes", false otherwise.
func ShowExplicitConfirmationPrompt(message string) (bool, error) {
	// Add warning indicator for highly destructive operations
	styledMessage := pterm.DefaultBox.
		WithTitle("⚠️  DESTRUCTIVE OPERATION").
		WithTitleTopCenter().
		WithBoxStyle(pterm.NewStyle(pterm.FgRed)).
		Sprint(message)

	pterm.Println()
	pterm.Println(styledMessage)
	pterm.Println()

	pterm.Warning.Println("Type 'yes' to confirm:")

	// Use pterm's interactive text input
	result, err := pterm.DefaultInteractiveTextInput.Show()
	if err != nil {
		return false, fmt.Errorf("confirmation input failed: %w", err)
	}

	confirmed := strings.TrimSpace(strings.ToLower(result)) == "yes"

	if !confirmed {
		pterm.Info.Println("Operation canceled (confirmation not received)")
	}

	return confirmed, nil
}

// substituteParameters replaces parameter placeholders in the message.
// Supports {paramName} syntax for parameter substitution.
func (e *Executor) substituteParameters(cmd *cobra.Command, message string) string {
	// Find all flags and their values
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if flag.Changed {
			placeholder := fmt.Sprintf("{%s}", flag.Name)
			message = strings.ReplaceAll(message, placeholder, flag.Value.String())

			// Also support camelCase and snake_case variations
			camelCase := toCamelCase(flag.Name)
			placeholder = fmt.Sprintf("{%s}", camelCase)
			message = strings.ReplaceAll(message, placeholder, flag.Value.String())

			snakeCase := toSnakeCase(flag.Name)
			placeholder = fmt.Sprintf("{%s}", snakeCase)
			message = strings.ReplaceAll(message, placeholder, flag.Value.String())
		}
	})

	return message
}

// toCamelCase converts a kebab-case or snake_case string to camelCase.
func toCamelCase(s string) string {
	s = strings.ReplaceAll(s, "-", "_")
	parts := strings.Split(s, "_")
	for i := 1; i < len(parts); i++ {
		if len(parts[i]) > 0 {
			parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
		}
	}
	return strings.Join(parts, "")
}

// toSnakeCase converts a kebab-case string to snake_case.
func toSnakeCase(s string) string {
	return strings.ReplaceAll(s, "-", "_")
}
