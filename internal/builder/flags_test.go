package builder

import (
	"context"
	"testing"

	"github.com/CliForge/cliforge/pkg/openapi"
	"github.com/getkin/kin-openapi/openapi3"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func TestFlagBuilder_AddOperationFlags(t *testing.T) {
	parser := openapi.NewParser()
	spec, err := parser.ParseFile(context.Background(), "../../examples/openapi/swagger2-example.json")
	if err != nil {
		t.Fatalf("Failed to parse spec: %v", err)
	}

	// Get listUsers operation
	operations, err := spec.GetOperations()
	if err != nil {
		t.Fatalf("Failed to get operations: %v", err)
	}

	var listOp *openapi.Operation
	for _, op := range operations {
		if op.OperationID == "listUsers" {
			listOp = op
			break
		}
	}

	if listOp == nil {
		t.Fatal("listUsers operation not found")
	}

	// Create flag builder
	flagBuilder := NewFlagBuilder(spec.Extensions.Config)

	// Create test command
	cmd := &cobra.Command{
		Use: "list",
	}

	// Add operation flags
	if err := flagBuilder.AddOperationFlags(cmd, listOp); err != nil {
		t.Fatalf("Failed to add operation flags: %v", err)
	}

	// Verify limit flag was added (query parameter)
	limitFlag := cmd.Flags().Lookup("limit")
	if limitFlag == nil {
		t.Error("Expected 'limit' flag not found")
	}
}

func TestFlagBuilder_AddGlobalFlags(t *testing.T) {
	flagBuilder := NewFlagBuilder(nil)
	cmd := &cobra.Command{Use: "root"}

	flagBuilder.AddGlobalFlags(cmd)

	// Verify global flags exist
	expectedFlags := []string{"output", "verbose", "no-color", "config", "profile", "dry-run", "debug", "interactive"}
	for _, flagName := range expectedFlags {
		flag := cmd.PersistentFlags().Lookup(flagName)
		if flag == nil {
			t.Errorf("Expected global flag '%s' not found", flagName)
		}
	}
}

func TestBuildRequestParams(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Annotations = make(map[string]string)

	// Add a parameter flag
	cmd.Flags().String("limit", "100", "Limit")
	cmd.Annotations["param:limit"] = "limit"
	cmd.Annotations["param:limit:in"] = "query"

	// Set the flag
	cmd.Flags().Set("limit", "50")

	// Build params
	params, err := BuildRequestParams(cmd)
	if err != nil {
		t.Fatalf("Failed to build request params: %v", err)
	}

	// Verify param was extracted
	if val, ok := params["limit"]; !ok {
		t.Error("Expected 'limit' parameter not found")
	} else if val != "50" {
		t.Errorf("Expected limit value '50', got '%v'", val)
	}
}

func TestBuildRequestBody(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}
	cmd.Annotations = make(map[string]string)

	// Add body field flags
	cmd.Flags().String("cluster-name", "", "Cluster name")
	cmd.Flags().String("region", "", "Region")
	cmd.Annotations["body:cluster-name"] = "name"
	cmd.Annotations["body:region"] = "region"

	// Set the flags
	cmd.Flags().Set("cluster-name", "my-cluster")
	cmd.Flags().Set("region", "us-east-1")

	// Build body
	body, err := BuildRequestBody(cmd)
	if err != nil {
		t.Fatalf("Failed to build request body: %v", err)
	}

	// Verify body fields
	if val, ok := body["name"]; !ok {
		t.Error("Expected 'name' field not found")
	} else if val != "my-cluster" {
		t.Errorf("Expected name value 'my-cluster', got '%v'", val)
	}

	if val, ok := body["region"]; !ok {
		t.Error("Expected 'region' field not found")
	} else if val != "us-east-1" {
		t.Errorf("Expected region value 'us-east-1', got '%v'", val)
	}
}

func TestToFlagName(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"clusterName", "clustername"},
		{"cluster_name", "cluster-name"},
		{"CLUSTER NAME", "cluster-name"},
		{"multi-AZ", "multi-az"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := toFlagName(tt.input)
			if result != tt.expected {
				t.Errorf("toFlagName(%s) = %s; want %s", tt.input, result, tt.expected)
			}
		})
	}
}

func TestValidateEnumFlags(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	// Add enum flag
	cmd.Flags().String("state", "", "State")
	cmd.Flags().SetAnnotation("state", "enum", []string{"pending", "ready", "error"})

	// Test valid value
	cmd.Flags().Set("state", "ready")
	if err := ValidateEnumFlags(cmd); err != nil {
		t.Errorf("Expected validation to pass for valid enum value, got error: %v", err)
	}

	// Test invalid value
	cmd.Flags().Set("state", "invalid")
	if err := ValidateEnumFlags(cmd); err == nil {
		t.Error("Expected validation to fail for invalid enum value")
	}
}

func TestAddRequestBodyFlags(t *testing.T) {
	// Create a test request body with properties
	requestBody := &openapi3.RequestBody{
		Content: openapi3.Content{
			"application/json": &openapi3.MediaType{
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"object"},
						Properties: openapi3.Schemas{
							"name": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type:        &openapi3.Types{"string"},
									Description: "Name of the resource",
								},
							},
							"enabled": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type:        &openapi3.Types{"boolean"},
									Description: "Whether resource is enabled",
								},
							},
							"count": &openapi3.SchemaRef{
								Value: &openapi3.Schema{
									Type:        &openapi3.Types{"integer"},
									Description: "Resource count",
								},
							},
						},
						Required: []string{"name"},
					},
				},
			},
		},
	}

	flagBuilder := NewFlagBuilder(nil)
	cmd := &cobra.Command{Use: "create"}

	// Add request body flags
	err := flagBuilder.addRequestBodyFlags(cmd, requestBody, nil)
	if err != nil {
		t.Fatalf("Failed to add request body flags: %v", err)
	}

	// Verify flags were added
	flagCount := 0
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		flagCount++
	})

	if flagCount != 3 {
		t.Errorf("Expected 3 request body flags, got %d", flagCount)
	}

	// Verify specific flags exist
	if cmd.Flags().Lookup("name") == nil {
		t.Error("Expected 'name' flag to be added")
	}
	if cmd.Flags().Lookup("enabled") == nil {
		t.Error("Expected 'enabled' flag to be added")
	}
	if cmd.Flags().Lookup("count") == nil {
		t.Error("Expected 'count' flag to be added")
	}

	// Verify body annotations were set
	if cmd.Annotations == nil {
		t.Error("Expected annotations to be set")
	} else {
		if _, ok := cmd.Annotations["body:name"]; !ok {
			t.Error("Expected body:name annotation")
		}
	}
}

func TestAddCustomFlags(t *testing.T) {
	tests := []struct {
		name     string
		cliFlags []*openapi.CLIFlag
		verify   func(*testing.T, *cobra.Command)
	}{
		{
			name: "string flag with default",
			cliFlags: []*openapi.CLIFlag{
				{
					Flag:        "custom-string",
					Type:        "string",
					Description: "A custom string flag",
					Default:     "default-value",
					Source:      "custom",
				},
			},
			verify: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("custom-string")
				if flag == nil {
					t.Error("Expected custom-string flag to be added")
					return
				}
				if flag.DefValue != "default-value" {
					t.Errorf("Expected default value 'default-value', got '%s'", flag.DefValue)
				}
			},
		},
		{
			name: "int flag with default",
			cliFlags: []*openapi.CLIFlag{
				{
					Flag:        "custom-int",
					Type:        "int",
					Description: "A custom int flag",
					Default:     float64(42),
					Source:      "custom",
				},
			},
			verify: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("custom-int")
				if flag == nil {
					t.Error("Expected custom-int flag to be added")
					return
				}
			},
		},
		{
			name: "bool flag with default",
			cliFlags: []*openapi.CLIFlag{
				{
					Flag:        "custom-bool",
					Type:        "bool",
					Description: "A custom bool flag",
					Default:     true,
					Source:      "custom",
				},
			},
			verify: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("custom-bool")
				if flag == nil {
					t.Error("Expected custom-bool flag to be added")
				}
			},
		},
		{
			name: "array flag",
			cliFlags: []*openapi.CLIFlag{
				{
					Flag:        "custom-array",
					Type:        "array",
					Description: "A custom array flag",
					Source:      "custom",
				},
			},
			verify: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("custom-array")
				if flag == nil {
					t.Error("Expected custom-array flag to be added")
				}
			},
		},
		{
			name: "required flag",
			cliFlags: []*openapi.CLIFlag{
				{
					Flag:        "required-flag",
					Type:        "string",
					Description: "A required flag",
					Required:    true,
					Source:      "custom",
				},
			},
			verify: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("required-flag")
				if flag == nil {
					t.Error("Expected required-flag to be added")
				}
			},
		},
		{
			name: "flag with aliases",
			cliFlags: []*openapi.CLIFlag{
				{
					Flag:        "aliased-flag",
					Type:        "string",
					Description: "A flag with aliases",
					Aliases:     []string{"af", "alias"},
					Source:      "custom",
				},
			},
			verify: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("aliased-flag")
				if flag == nil {
					t.Error("Expected aliased-flag to be added")
					return
				}
				if flag.Annotations == nil {
					t.Error("Expected annotations to be set")
					return
				}
				if aliases, ok := flag.Annotations["aliases"]; !ok {
					t.Error("Expected aliases annotation")
				} else if len(aliases) != 2 {
					t.Errorf("Expected 2 aliases, got %d", len(aliases))
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagBuilder := NewFlagBuilder(nil)
			cmd := &cobra.Command{Use: "test"}

			err := flagBuilder.addCustomFlags(cmd, tt.cliFlags)
			if err != nil {
				t.Fatalf("Failed to add custom flags: %v", err)
			}

			tt.verify(t, cmd)
		})
	}
}

func TestAddFlagFromSchema_Types(t *testing.T) {
	tests := []struct {
		name        string
		schemaType  string
		schemaValue *openapi3.Schema
		flagName    string
		verify      func(*testing.T, *cobra.Command)
	}{
		{
			name:       "string type",
			schemaType: "string",
			schemaValue: &openapi3.Schema{
				Type:        &openapi3.Types{"string"},
				Description: "A string parameter",
			},
			flagName: "test-string",
			verify: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("test-string")
				if flag == nil {
					t.Error("Expected test-string flag to be added")
				}
			},
		},
		{
			name:       "integer type",
			schemaType: "integer",
			schemaValue: &openapi3.Schema{
				Type:        &openapi3.Types{"integer"},
				Description: "An integer parameter",
			},
			flagName: "test-int",
			verify: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("test-int")
				if flag == nil {
					t.Error("Expected test-int flag to be added")
				}
			},
		},
		{
			name:       "boolean type",
			schemaType: "boolean",
			schemaValue: &openapi3.Schema{
				Type:        &openapi3.Types{"boolean"},
				Description: "A boolean parameter",
			},
			flagName: "test-bool",
			verify: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("test-bool")
				if flag == nil {
					t.Error("Expected test-bool flag to be added")
				}
			},
		},
		{
			name:       "number type",
			schemaType: "number",
			schemaValue: &openapi3.Schema{
				Type:        &openapi3.Types{"number"},
				Description: "A number parameter",
			},
			flagName: "test-float",
			verify: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("test-float")
				if flag == nil {
					t.Error("Expected test-float flag to be added")
				}
			},
		},
		{
			name:       "array type",
			schemaType: "array",
			schemaValue: &openapi3.Schema{
				Type:        &openapi3.Types{"array"},
				Description: "An array parameter",
			},
			flagName: "test-array",
			verify: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("test-array")
				if flag == nil {
					t.Error("Expected test-array flag to be added")
				}
			},
		},
		{
			name:       "enum type",
			schemaType: "string",
			schemaValue: &openapi3.Schema{
				Type:        &openapi3.Types{"string"},
				Enum:        []interface{}{"value1", "value2", "value3"},
				Description: "An enum parameter",
			},
			flagName: "test-enum",
			verify: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("test-enum")
				if flag == nil {
					t.Error("Expected test-enum flag to be added")
					return
				}
				if flag.Annotations == nil {
					t.Error("Expected annotations for enum")
					return
				}
				if enumVals, ok := flag.Annotations["enum"]; !ok {
					t.Error("Expected enum annotation")
				} else if len(enumVals) != 3 {
					t.Errorf("Expected 3 enum values, got %d", len(enumVals))
				}
			},
		},
		{
			name:       "password format",
			schemaType: "string",
			schemaValue: &openapi3.Schema{
				Type:        &openapi3.Types{"string"},
				Format:      "password",
				Description: "A password parameter",
			},
			flagName: "test-password",
			verify: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("test-password")
				if flag == nil {
					t.Error("Expected test-password flag to be added")
					return
				}
				if flag.Annotations == nil {
					t.Error("Expected sensitive annotation for password")
					return
				}
				if sensitive, ok := flag.Annotations["sensitive"]; !ok {
					t.Error("Expected sensitive annotation")
				} else if len(sensitive) == 0 || sensitive[0] != "true" {
					t.Error("Expected sensitive annotation to be true")
				}
			},
		},
		{
			name:       "date-time format",
			schemaType: "string",
			schemaValue: &openapi3.Schema{
				Type:        &openapi3.Types{"string"},
				Format:      "date-time",
				Description: "A date-time parameter",
			},
			flagName: "test-datetime",
			verify: func(t *testing.T, cmd *cobra.Command) {
				flag := cmd.Flags().Lookup("test-datetime")
				if flag == nil {
					t.Error("Expected test-datetime flag to be added")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagBuilder := NewFlagBuilder(nil)
			cmd := &cobra.Command{Use: "test"}

			schemaRef := &openapi3.SchemaRef{Value: tt.schemaValue}
			err := flagBuilder.addFlagFromSchema(cmd, tt.flagName, schemaRef, tt.schemaValue.Description, false)
			if err != nil {
				t.Fatalf("Failed to add flag from schema: %v", err)
			}

			tt.verify(t, cmd)
		})
	}
}

func TestAddFlagFromSchema_EdgeCases(t *testing.T) {
	tests := []struct {
		name      string
		schemaRef *openapi3.SchemaRef
		required  bool
		wantErr   bool
	}{
		{
			name:      "nil schema ref",
			schemaRef: nil,
			required:  false,
			wantErr:   false, // Should default to string
		},
		{
			name:      "nil schema value",
			schemaRef: &openapi3.SchemaRef{Value: nil},
			required:  false,
			wantErr:   false, // Should default to string
		},
		{
			name: "empty type slice",
			schemaRef: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{},
				},
			},
			required: false,
			wantErr:  false, // Should default to string
		},
		{
			name: "required flag",
			schemaRef: &openapi3.SchemaRef{
				Value: &openapi3.Schema{
					Type: &openapi3.Types{"string"},
				},
			},
			required: true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flagBuilder := NewFlagBuilder(nil)
			cmd := &cobra.Command{Use: "test"}

			err := flagBuilder.addFlagFromSchema(cmd, "test-flag", tt.schemaRef, "Test flag", tt.required)
			if (err != nil) != tt.wantErr {
				t.Errorf("addFlagFromSchema() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify flag was added
			flag := cmd.Flags().Lookup("test-flag")
			if flag == nil && !tt.wantErr {
				t.Error("Expected flag to be added")
			}
		})
	}
}

func TestGetFlagValue_AllTypes(t *testing.T) {
	cmd := &cobra.Command{Use: "test"}

	// Add flags of different types
	cmd.Flags().String("string-flag", "default", "String flag")
	cmd.Flags().Int("int-flag", 42, "Int flag")
	cmd.Flags().Float64("float-flag", 3.14, "Float flag")
	cmd.Flags().Bool("bool-flag", true, "Bool flag")
	cmd.Flags().StringArray("array-flag", []string{"a", "b"}, "Array flag")
	cmd.Flags().StringSlice("slice-flag", []string{"x", "y"}, "Slice flag")

	tests := []struct {
		name     string
		flagName string
		wantType string
		wantErr  bool
	}{
		{"string type", "string-flag", "string", false},
		{"int type", "int-flag", "int", false},
		{"float type", "float-flag", "float64", false},
		{"bool type", "bool-flag", "bool", false},
		{"array type", "array-flag", "stringArray", false},
		{"slice type", "slice-flag", "stringSlice", false},
		{"non-existent flag", "missing-flag", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := GetFlagValue(cmd.Flags(), tt.flagName)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetFlagValue() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && value == nil {
				t.Error("Expected non-nil value")
			}
		})
	}
}

func TestValidateRequiredFlags(t *testing.T) {
	tests := []struct {
		name    string
		setup   func(*cobra.Command)
		wantErr bool
	}{
		{
			name: "all required flags set",
			setup: func(cmd *cobra.Command) {
				cmd.Flags().String("required-flag", "", "Required flag")
				cmd.MarkFlagRequired("required-flag")
				cmd.Flags().Set("required-flag", "value")
			},
			wantErr: false,
		},
		{
			name: "required flag not set",
			setup: func(cmd *cobra.Command) {
				cmd.Flags().String("required-flag", "", "Required flag")
				cmd.MarkFlagRequired("required-flag")
			},
			wantErr: true,
		},
		{
			name: "no required flags",
			setup: func(cmd *cobra.Command) {
				cmd.Flags().String("optional-flag", "", "Optional flag")
			},
			wantErr: false,
		},
		{
			name: "multiple required flags, one missing",
			setup: func(cmd *cobra.Command) {
				cmd.Flags().String("required-1", "", "Required 1")
				cmd.Flags().String("required-2", "", "Required 2")
				cmd.MarkFlagRequired("required-1")
				cmd.MarkFlagRequired("required-2")
				cmd.Flags().Set("required-1", "value")
				// required-2 not set
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd := &cobra.Command{Use: "test"}
			tt.setup(cmd)

			err := ValidateRequiredFlags(cmd)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateRequiredFlags() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestAddParameterFlags_SkipsHeaderAndPath(t *testing.T) {
	flagBuilder := NewFlagBuilder(nil)
	cmd := &cobra.Command{Use: "test"}

	// Create parameters with different "in" locations
	params := openapi3.Parameters{
		&openapi3.ParameterRef{
			Value: &openapi3.Parameter{
				Name: "path-param",
				In:   "path",
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"string"},
					},
				},
			},
		},
		&openapi3.ParameterRef{
			Value: &openapi3.Parameter{
				Name: "header-param",
				In:   "header",
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"string"},
					},
				},
			},
		},
		&openapi3.ParameterRef{
			Value: &openapi3.Parameter{
				Name: "query-param",
				In:   "query",
				Schema: &openapi3.SchemaRef{
					Value: &openapi3.Schema{
						Type: &openapi3.Types{"string"},
					},
				},
			},
		},
	}

	err := flagBuilder.addParameterFlags(cmd, params)
	if err != nil {
		t.Fatalf("Failed to add parameter flags: %v", err)
	}

	// Verify path and header params were skipped
	if cmd.Flags().Lookup("path-param") != nil {
		t.Error("Expected path parameter to be skipped")
	}
	if cmd.Flags().Lookup("header-param") != nil {
		t.Error("Expected header parameter to be skipped")
	}

	// Verify query param was added
	if cmd.Flags().Lookup("query-param") == nil {
		t.Error("Expected query parameter to be added")
	}
}
