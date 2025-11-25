package builtin

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/CliForge/cliforge/pkg/plugin"
)

// TestExecPlugin tests the exec plugin functionality
func TestExecPlugin_Execute(t *testing.T) {
	execPlugin := NewExecPlugin([]string{"echo", "cat"}, false)

	tests := []struct {
		name      string
		input     *plugin.PluginInput
		wantErr   bool
		checkFunc func(*testing.T, *plugin.PluginOutput)
	}{
		{
			name: "simple echo command",
			input: &plugin.PluginInput{
				Command: "echo",
				Args:    []string{"hello", "world"},
			},
			wantErr: false,
			checkFunc: func(t *testing.T, output *plugin.PluginOutput) {
				if !strings.Contains(output.Stdout, "hello") {
					t.Errorf("Stdout should contain 'hello', got: %s", output.Stdout)
				}
				if output.ExitCode != 0 {
					t.Errorf("ExitCode = %v, want 0", output.ExitCode)
				}
			},
		},
		{
			name: "command not in allowed list",
			input: &plugin.PluginInput{
				Command: "rm",
				Args:    []string{"-rf", "/"},
			},
			wantErr: true,
		},
		{
			name: "missing command",
			input: &plugin.PluginInput{
				Command: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := execPlugin.Execute(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkFunc != nil {
				tt.checkFunc(t, output)
			}
		})
	}
}

func TestExecPlugin_Validate(t *testing.T) {
	execPlugin := NewExecPlugin(nil, false)
	if err := execPlugin.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestExecPlugin_Describe(t *testing.T) {
	execPlugin := NewExecPlugin(nil, false)
	info := execPlugin.Describe()

	if info.Manifest.Name != "exec" {
		t.Errorf("Name = %v, want exec", info.Manifest.Name)
	}

	if info.Manifest.Type != plugin.PluginTypeBuiltin {
		t.Errorf("Type = %v, want builtin", info.Manifest.Type)
	}

	if len(info.Capabilities) == 0 {
		t.Error("Capabilities should not be empty")
	}
}

// TestFileOpsPlugin tests the file operations plugin
func TestFileOpsPlugin_Read(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileOps := NewFileOpsPlugin([]string{tmpDir}, 10*1024*1024)

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"operation": "read",
			"file":      testFile,
		},
	}

	output, err := fileOps.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.ExitCode != 0 {
		t.Errorf("ExitCode = %v, want 0", output.ExitCode)
	}

	content, ok := output.GetString("content")
	if !ok {
		t.Fatal("Failed to get content from output")
	}

	if content != testContent {
		t.Errorf("Content = %v, want %v", content, testContent)
	}
}

func TestFileOpsPlugin_ParseJSON(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	testContent := `{"key": "value", "number": 42}`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileOps := NewFileOpsPlugin([]string{tmpDir}, 10*1024*1024)

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"operation": "parse",
			"file":      testFile,
			"format":    "json",
		},
	}

	output, err := fileOps.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.ExitCode != 0 {
		t.Errorf("ExitCode = %v, want 0", output.ExitCode)
	}

	parsed, ok := output.GetMap("parsed")
	if !ok {
		t.Fatal("Failed to get parsed data from output")
	}

	if parsed["key"] != "value" {
		t.Errorf("parsed[key] = %v, want value", parsed["key"])
	}
}

func TestFileOpsPlugin_ParseYAML(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.yaml")
	testContent := "key: value\nnumber: 42"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileOps := NewFileOpsPlugin([]string{tmpDir}, 10*1024*1024)

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"operation": "parse",
			"file":      testFile,
			"format":    "yaml",
		},
	}

	output, err := fileOps.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.ExitCode != 0 {
		t.Errorf("ExitCode = %v, want 0", output.ExitCode)
	}
}

func TestFileOpsPlugin_ParseHTPasswd(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "htpasswd")
	testContent := "user1:$apr1$abc123$xyz\nuser2:$apr1$def456$uvw"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileOps := NewFileOpsPlugin([]string{tmpDir}, 10*1024*1024)

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"operation": "parse",
			"file":      testFile,
			"format":    "htpasswd",
		},
	}

	output, err := fileOps.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.ExitCode != 0 {
		t.Errorf("ExitCode = %v, want 0", output.ExitCode)
	}

	count, ok := output.GetInt("count")
	if !ok {
		t.Fatal("Failed to get count from output")
	}

	if count != 2 {
		t.Errorf("count = %v, want 2", count)
	}
}

// TestValidatorsPlugin tests the validators plugin
func TestValidatorsPlugin_Email(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name      string
		email     string
		wantValid bool
	}{
		{"valid email", "test@example.com", true},
		{"invalid email", "not-an-email", false},
		{"missing @", "test.example.com", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "email",
					"value":     tt.email,
				},
			}

			output, err := validators.Execute(context.Background(), input)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			valid, ok := output.GetBool("valid")
			if !ok {
				t.Fatal("Failed to get valid from output")
			}

			if valid != tt.wantValid {
				t.Errorf("valid = %v, want %v", valid, tt.wantValid)
			}
		})
	}
}

func TestValidatorsPlugin_ClusterName(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name        string
		clusterName string
		wantValid   bool
	}{
		{"valid name", "my-cluster", true},
		{"valid single char", "a", false}, // Needs at least 2 chars per regex
		{"starts with number", "1cluster", false},
		{"uppercase letters", "My-Cluster", false},
		{"ends with hyphen", "cluster-", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "cluster-name",
					"value":     tt.clusterName,
				},
			}

			output, err := validators.Execute(context.Background(), input)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			valid, ok := output.GetBool("valid")
			if !ok {
				t.Fatal("Failed to get valid from output")
			}

			if valid != tt.wantValid {
				t.Errorf("valid = %v, want %v for %s", valid, tt.wantValid, tt.clusterName)
			}
		})
	}
}

func TestValidatorsPlugin_Regex(t *testing.T) {
	validators := NewValidatorsPlugin()

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"validator": "regex",
			"value":     "test123",
			"pattern":   "^[a-z]+[0-9]+$",
		},
	}

	output, err := validators.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	valid, ok := output.GetBool("valid")
	if !ok {
		t.Fatal("Failed to get valid from output")
	}

	if !valid {
		t.Error("Regex validation should pass")
	}
}

// TestTransformersPlugin tests the transformers plugin
func TestTransformersPlugin_Base64Encode(t *testing.T) {
	transformers := NewTransformersPlugin()

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"transformation": "base64-encode",
			"input":          "Hello, World!",
		},
	}

	output, err := transformers.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.ExitCode != 0 {
		t.Errorf("ExitCode = %v, want 0", output.ExitCode)
	}

	encoded, ok := output.GetString("output")
	if !ok {
		t.Fatal("Failed to get output from result")
	}

	if encoded != "SGVsbG8sIFdvcmxkIQ==" {
		t.Errorf("Encoded = %v, want SGVsbG8sIFdvcmxkIQ==", encoded)
	}
}

func TestTransformersPlugin_Base64Decode(t *testing.T) {
	transformers := NewTransformersPlugin()

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"transformation": "base64-decode",
			"input":          "SGVsbG8sIFdvcmxkIQ==",
		},
	}

	output, err := transformers.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	decoded, ok := output.GetString("output")
	if !ok {
		t.Fatal("Failed to get output from result")
	}

	if decoded != "Hello, World!" {
		t.Errorf("Decoded = %v, want Hello, World!", decoded)
	}
}

func TestTransformersPlugin_JSONToYAML(t *testing.T) {
	transformers := NewTransformersPlugin()

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"transformation": "json-to-yaml",
			"input":          `{"key": "value", "number": 42}`,
		},
	}

	output, err := transformers.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	yamlOutput, ok := output.GetString("output")
	if !ok {
		t.Fatal("Failed to get output from result")
	}

	if !strings.Contains(yamlOutput, "key: value") {
		t.Errorf("YAML output should contain 'key: value', got: %s", yamlOutput)
	}
}

func TestTransformersPlugin_HTPasswdToUsers(t *testing.T) {
	transformers := NewTransformersPlugin()

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"transformation": "htpasswd-to-users",
			"input":          "user1:hash1\nuser2:hash2",
		},
	}

	output, err := transformers.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	count, ok := output.GetInt("count")
	if !ok {
		t.Fatal("Failed to get count from result")
	}

	if count != 2 {
		t.Errorf("count = %v, want 2", count)
	}
}

func TestTransformersPlugin_Template(t *testing.T) {
	transformers := NewTransformersPlugin()

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"transformation": "template",
			"template":       "Hello, {{name}}! You are {{age}} years old.",
			"values": map[string]interface{}{
				"name": "Alice",
				"age":  30,
			},
		},
	}

	output, err := transformers.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	result, ok := output.GetString("output")
	if !ok {
		t.Fatal("Failed to get output from result")
	}

	expected := "Hello, Alice! You are 30 years old."
	if result != expected {
		t.Errorf("result = %v, want %v", result, expected)
	}
}
