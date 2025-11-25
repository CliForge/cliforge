package builder

import (
	"fmt"
	"strings"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FlagBuilder builds flags for Cobra commands from OpenAPI parameters and schemas.
type FlagBuilder struct {
	globalFlags *GlobalFlags
}

// GlobalFlags contains global flags from config.
type GlobalFlags struct {
	Output      string
	Verbose     bool
	NoColor     bool
	ConfigFile  string
	Profile     string
	DryRun      bool
	Debug       bool
	Interactive bool
}

// NewFlagBuilder creates a new flag builder.
func NewFlagBuilder(config *openapi.CLIConfig) *FlagBuilder {
	return &FlagBuilder{
		globalFlags: &GlobalFlags{},
	}
}

// AddOperationFlags adds flags for an operation to a command.
func (fb *FlagBuilder) AddOperationFlags(cmd *cobra.Command, op *openapi.Operation) error {
	// Add flags from parameters
	if err := fb.addParameterFlags(cmd, op.Operation.Parameters); err != nil {
		return fmt.Errorf("failed to add parameter flags: %w", err)
	}

	// Add flags from request body
	if op.Operation.RequestBody != nil && op.Operation.RequestBody.Value != nil {
		if err := fb.addRequestBodyFlags(cmd, op.Operation.RequestBody.Value, op.CLIFlags); err != nil {
			return fmt.Errorf("failed to add request body flags: %w", err)
		}
	}

	// Add custom CLI flags from x-cli-flags
	if err := fb.addCustomFlags(cmd, op.CLIFlags); err != nil {
		return fmt.Errorf("failed to add custom flags: %w", err)
	}

	return nil
}

// addParameterFlags adds flags from OpenAPI parameters.
func (fb *FlagBuilder) addParameterFlags(cmd *cobra.Command, parameters openapi3.Parameters) error {
	for _, paramRef := range parameters {
		if paramRef.Value == nil {
			continue
		}

		param := paramRef.Value

		// Skip path parameters (they're positional arguments)
		if param.In == "path" {
			continue
		}

		// Skip header parameters unless explicitly flagged
		if param.In == "header" {
			continue
		}

		// Determine flag name
		flagName := toFlagName(param.Name)

		// Determine flag type and add
		if err := fb.addFlagFromSchema(cmd, flagName, param.Schema, param.Description, param.Required); err != nil {
			return fmt.Errorf("failed to add flag for parameter %s: %w", param.Name, err)
		}

		// Store parameter mapping in annotations
		if cmd.Annotations == nil {
			cmd.Annotations = make(map[string]string)
		}
		cmd.Annotations[fmt.Sprintf("param:%s", flagName)] = param.Name
		cmd.Annotations[fmt.Sprintf("param:%s:in", flagName)] = param.In
	}

	return nil
}

// addRequestBodyFlags adds flags from request body schema.
func (fb *FlagBuilder) addRequestBodyFlags(cmd *cobra.Command, requestBody *openapi3.RequestBody, cliFlags []*openapi.CLIFlag) error {
	// Get JSON content schema
	content := requestBody.Content.Get("application/json")
	if content == nil {
		// No JSON content, skip
		return nil
	}

	if content.Schema == nil || content.Schema.Value == nil {
		return nil
	}

	schema := content.Schema.Value

	// Build flag mapping from x-cli-flags
	flagMap := make(map[string]*openapi.CLIFlag)
	for _, cliFlag := range cliFlags {
		if cliFlag.Source == "body" {
			flagMap[cliFlag.Name] = cliFlag
		}
	}

	// Add flags for each property
	for propName, propSchema := range schema.Properties {
		if propSchema.Value == nil {
			continue
		}

		// Check if there's a custom CLI flag mapping
		var flagName string
		var description string
		var required bool

		if cliFlag, ok := flagMap[propName]; ok {
			flagName = cliFlag.Flag
			description = cliFlag.Description
			required = cliFlag.Required
		} else {
			flagName = toFlagName(propName)
			description = propSchema.Value.Description
			// Check if required in schema
			for _, req := range schema.Required {
				if req == propName {
					required = true
					break
				}
			}
		}

		// Add flag
		if err := fb.addFlagFromSchema(cmd, flagName, propSchema, description, required); err != nil {
			return fmt.Errorf("failed to add flag for property %s: %w", propName, err)
		}

		// Store body field mapping in annotations
		if cmd.Annotations == nil {
			cmd.Annotations = make(map[string]string)
		}
		cmd.Annotations[fmt.Sprintf("body:%s", flagName)] = propName
	}

	return nil
}

// addCustomFlags adds custom flags from x-cli-flags that aren't from params/body.
func (fb *FlagBuilder) addCustomFlags(cmd *cobra.Command, cliFlags []*openapi.CLIFlag) error {
	for _, cliFlag := range cliFlags {
		// Skip if already processed as param or body
		if cliFlag.Source == "param" || cliFlag.Source == "body" {
			continue
		}

		// Add custom flag
		flagName := cliFlag.Flag

		// Determine type and add flag
		switch cliFlag.Type {
		case "string":
			defaultVal := ""
			if cliFlag.Default != nil {
				if str, ok := cliFlag.Default.(string); ok {
					defaultVal = str
				}
			}
			cmd.Flags().String(flagName, defaultVal, cliFlag.Description)

		case "int", "integer":
			defaultVal := 0
			if cliFlag.Default != nil {
				if num, ok := cliFlag.Default.(float64); ok {
					defaultVal = int(num)
				}
			}
			cmd.Flags().Int(flagName, defaultVal, cliFlag.Description)

		case "bool", "boolean":
			defaultVal := false
			if cliFlag.Default != nil {
				if b, ok := cliFlag.Default.(bool); ok {
					defaultVal = b
				}
			}
			cmd.Flags().Bool(flagName, defaultVal, cliFlag.Description)

		case "array", "stringArray":
			cmd.Flags().StringArray(flagName, nil, cliFlag.Description)

		default:
			cmd.Flags().String(flagName, "", cliFlag.Description)
		}

		// Mark as required if specified
		if cliFlag.Required {
			if err := cmd.MarkFlagRequired(flagName); err != nil {
				return fmt.Errorf("failed to mark flag %s as required: %w", flagName, err)
			}
		}

		// Add aliases - store in custom annotation
		if len(cliFlag.Aliases) > 0 {
			cmd.Flags().SetAnnotation(flagName, "aliases", cliFlag.Aliases)
		}
	}

	return nil
}

// addFlagFromSchema adds a flag based on OpenAPI schema.
func (fb *FlagBuilder) addFlagFromSchema(cmd *cobra.Command, flagName string, schemaRef *openapi3.SchemaRef, description string, required bool) error {
	if schemaRef == nil || schemaRef.Value == nil {
		// Default to string
		cmd.Flags().String(flagName, "", description)
		return nil
	}

	schema := schemaRef.Value

	// Handle enum as string with validation
	if len(schema.Enum) > 0 {
		defaultVal := ""
		if schema.Default != nil {
			if str, ok := schema.Default.(string); ok {
				defaultVal = str
			}
		}
		cmd.Flags().String(flagName, defaultVal, description)

		// Store enum values in annotation for validation
		enumStrs := make([]string, len(schema.Enum))
		for i, e := range schema.Enum {
			enumStrs[i] = fmt.Sprintf("%v", e)
		}
		cmd.Flags().SetAnnotation(flagName, "enum", enumStrs)

		if required {
			if err := cmd.MarkFlagRequired(flagName); err != nil {
				return err
			}
		}
		return nil
	}

	// Handle based on type
	typeSlice := schema.Type.Slice()
	if len(typeSlice) == 0 {
		// Default to string for unknown types
		cmd.Flags().String(flagName, "", description)
		if required {
			if err := cmd.MarkFlagRequired(flagName); err != nil {
				return err
			}
		}
		return nil
	}

	schemaType := typeSlice[0]
	switch schemaType {
	case "string":
		defaultVal := ""
		if schema.Default != nil {
			if str, ok := schema.Default.(string); ok {
				defaultVal = str
			}
		}

		// Check for specific formats
		switch schema.Format {
		case "date", "date-time":
			cmd.Flags().String(flagName, defaultVal, description)
		case "password":
			cmd.Flags().String(flagName, defaultVal, description)
			// Mark as sensitive in annotations
			cmd.Flags().SetAnnotation(flagName, "sensitive", []string{"true"})
		default:
			cmd.Flags().String(flagName, defaultVal, description)
		}

	case "integer", "int32", "int64":
		defaultVal := 0
		if schema.Default != nil {
			if num, ok := schema.Default.(float64); ok {
				defaultVal = int(num)
			}
		}
		cmd.Flags().Int(flagName, defaultVal, description)

	case "number", "float", "double":
		defaultVal := 0.0
		if schema.Default != nil {
			if num, ok := schema.Default.(float64); ok {
				defaultVal = num
			}
		}
		cmd.Flags().Float64(flagName, defaultVal, description)

	case "boolean":
		defaultVal := false
		if schema.Default != nil {
			if b, ok := schema.Default.(bool); ok {
				defaultVal = b
			}
		}
		cmd.Flags().Bool(flagName, defaultVal, description)

	case "array":
		// Handle array types
		cmd.Flags().StringArray(flagName, nil, description)

	default:
		// Default to string for unknown types
		cmd.Flags().String(flagName, "", description)
	}

	// Mark as required if needed
	if required {
		if err := cmd.MarkFlagRequired(flagName); err != nil {
			return err
		}
	}

	return nil
}

// AddGlobalFlags adds global flags to a command.
func (fb *FlagBuilder) AddGlobalFlags(cmd *cobra.Command) {
	// Output format
	cmd.PersistentFlags().StringP("output", "o", "json", "Output format (json, yaml, table)")

	// Verbosity
	cmd.PersistentFlags().BoolP("verbose", "v", false, "Enable verbose output")

	// Color
	cmd.PersistentFlags().Bool("no-color", false, "Disable colored output")

	// Config file
	cmd.PersistentFlags().String("config", "", "Config file path")

	// Profile
	cmd.PersistentFlags().String("profile", "", "Configuration profile to use")

	// Dry run
	cmd.PersistentFlags().Bool("dry-run", false, "Print what would be done without executing")

	// Debug
	cmd.PersistentFlags().Bool("debug", false, "Enable debug mode")

	// Interactive
	cmd.PersistentFlags().BoolP("interactive", "i", false, "Enable interactive mode")
}

// GetFlagValue retrieves a flag value with type conversion.
func GetFlagValue(flags *pflag.FlagSet, name string) (interface{}, error) {
	flag := flags.Lookup(name)
	if flag == nil {
		return nil, fmt.Errorf("flag %s not found", name)
	}

	// Return based on type
	switch flag.Value.Type() {
	case "string":
		return flags.GetString(name)
	case "int", "int32", "int64":
		return flags.GetInt(name)
	case "float64", "float32":
		return flags.GetFloat64(name)
	case "bool":
		return flags.GetBool(name)
	case "stringArray":
		return flags.GetStringArray(name)
	case "stringSlice":
		return flags.GetStringSlice(name)
	default:
		return flag.Value.String(), nil
	}
}

// BuildRequestParams builds request parameters from flags.
func BuildRequestParams(cmd *cobra.Command) (map[string]interface{}, error) {
	params := make(map[string]interface{})

	// Iterate through all flags
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if !flag.Changed {
			return
		}

		// Get the parameter mapping from annotations
		paramKey := fmt.Sprintf("param:%s", flag.Name)
		if cmd.Annotations != nil {
			if paramName, ok := cmd.Annotations[paramKey]; ok {
				val, _ := GetFlagValue(cmd.Flags(), flag.Name)
				params[paramName] = val
			}
		}
	})

	return params, nil
}

// BuildRequestBody builds request body from flags.
func BuildRequestBody(cmd *cobra.Command) (map[string]interface{}, error) {
	body := make(map[string]interface{})

	// Iterate through all flags
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if !flag.Changed {
			return
		}

		// Get the body field mapping from annotations
		bodyKey := fmt.Sprintf("body:%s", flag.Name)
		if cmd.Annotations != nil {
			if fieldName, ok := cmd.Annotations[bodyKey]; ok {
				val, _ := GetFlagValue(cmd.Flags(), flag.Name)
				body[fieldName] = val
			}
		}
	})

	return body, nil
}

// toFlagName converts a parameter name to a flag name.
func toFlagName(name string) string {
	// Convert to kebab-case
	name = strings.ReplaceAll(name, "_", "-")
	name = strings.ReplaceAll(name, " ", "-")
	name = strings.ToLower(name)
	return name
}

// ValidateRequiredFlags validates that all required flags are set.
func ValidateRequiredFlags(cmd *cobra.Command) error {
	var missingFlags []string

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		// Check if flag is required
		annotations := flag.Annotations
		if annotations != nil {
			if _, ok := annotations[cobra.BashCompOneRequiredFlag]; ok {
				if !flag.Changed {
					missingFlags = append(missingFlags, flag.Name)
				}
			}
		}
	})

	if len(missingFlags) > 0 {
		return fmt.Errorf("required flags not set: %s", strings.Join(missingFlags, ", "))
	}

	return nil
}

// ValidateEnumFlags validates enum flag values.
func ValidateEnumFlags(cmd *cobra.Command) error {
	var errors []string

	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		if !flag.Changed {
			return
		}

		// Check for enum annotation
		if flag.Annotations != nil {
			if enumValues, ok := flag.Annotations["enum"]; ok {
				value := flag.Value.String()

				// Check if value is in enum
				valid := false
				for _, enumVal := range enumValues {
					if value == enumVal {
						valid = true
						break
					}
				}

				if !valid {
					errors = append(errors, fmt.Sprintf("flag --%s: value '%s' not in allowed values: %s",
						flag.Name, value, strings.Join(enumValues, ", ")))
				}
			}
		}
	})

	if len(errors) > 0 {
		return fmt.Errorf("validation errors:\n  %s", strings.Join(errors, "\n  "))
	}

	return nil
}
