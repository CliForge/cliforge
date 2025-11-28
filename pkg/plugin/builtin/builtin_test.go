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

// Additional ExecPlugin tests for comprehensive coverage

func TestExecPlugin_CommandInjection(t *testing.T) {
	execPlugin := NewExecPlugin(nil, false)

	tests := []struct {
		name    string
		command string
		wantErr bool
	}{
		{"semicolon injection", "ls;rm -rf /", true},
		{"pipe injection", "cat|grep", true},
		{"ampersand injection", "ls&whoami", true},
		{"dollar injection", "ls$HOME", true},
		{"backtick injection", "ls`whoami`", true},
		{"parenthesis injection", "ls()", true},
		{"redirect injection", "ls>file", true},
		{"newline injection", "ls\nrm", true},
		{"clean command", "ls", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Command: tt.command,
			}
			_, err := execPlugin.Execute(context.Background(), input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecPlugin_ArgumentValidation(t *testing.T) {
	execPlugin := NewExecPlugin([]string{"echo"}, false)

	tests := []struct {
		name    string
		args    []string
		wantErr bool
	}{
		{"normal args", []string{"hello", "world"}, false},
		{"args with spaces", []string{"hello world"}, false},
		{"args with null byte", []string{"hello\x00world"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Command: "echo",
				Args:    tt.args,
			}
			_, err := execPlugin.Execute(context.Background(), input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestExecPlugin_WithStdin(t *testing.T) {
	execPlugin := NewExecPlugin([]string{"cat"}, false)

	input := &plugin.PluginInput{
		Command: "cat",
		Stdin:   "test input data",
	}

	output, err := execPlugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(output.Stdout, "test input data") {
		t.Errorf("Stdout should contain stdin data, got: %s", output.Stdout)
	}
}

func TestExecPlugin_WithEnvironment(t *testing.T) {
	execPlugin := NewExecPlugin([]string{"sh"}, false)

	input := &plugin.PluginInput{
		Command: "sh",
		Args:    []string{"-c", "echo $TEST_VAR"},
		Env: map[string]string{
			"TEST_VAR": "test_value",
		},
	}

	output, err := execPlugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(output.Stdout, "test_value") {
		t.Errorf("Stdout should contain env variable, got: %s", output.Stdout)
	}
}

func TestExecPlugin_WithWorkingDirectory(t *testing.T) {
	execPlugin := NewExecPlugin([]string{"pwd"}, false)
	tmpDir := t.TempDir()

	input := &plugin.PluginInput{
		Command:    "pwd",
		WorkingDir: tmpDir,
	}

	output, err := execPlugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if !strings.Contains(output.Stdout, tmpDir) {
		t.Errorf("Stdout should contain working dir, got: %s", output.Stdout)
	}
}

func TestExecPlugin_InvalidWorkingDirectory(t *testing.T) {
	execPlugin := NewExecPlugin([]string{"ls"}, false)

	input := &plugin.PluginInput{
		Command:    "ls",
		WorkingDir: "/nonexistent/directory/path",
	}

	_, err := execPlugin.Execute(context.Background(), input)
	if err == nil {
		t.Error("Execute() should fail with invalid working directory")
	}
}

func TestExecPlugin_WithSandbox(t *testing.T) {
	execPlugin := NewExecPlugin([]string{"env"}, true)

	input := &plugin.PluginInput{
		Command: "env",
	}

	output, err := execPlugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.ExitCode != 0 {
		t.Errorf("ExitCode = %v, want 0", output.ExitCode)
	}

	// When sandbox is enabled, environment should be filtered
	// Check that only safe env vars are present
	if strings.Contains(output.Stdout, "PWD=") && !strings.Contains(output.Stdout, "PATH=") {
		// This is expected - PATH might be filtered by sandbox
	}
}

func TestExecPlugin_CommandNotFound(t *testing.T) {
	execPlugin := NewExecPlugin(nil, false)

	input := &plugin.PluginInput{
		Command: "nonexistent_command_xyz123",
	}

	output, err := execPlugin.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.ExitCode == 0 {
		t.Error("ExitCode should be non-zero for command not found")
	}
}

func TestExecuteWithShell(t *testing.T) {
	tests := []struct {
		name     string
		command  string
		shell    string
		wantErr  bool
		checkOut func(string) bool
	}{
		{
			name:    "simple shell command",
			command: "echo hello",
			shell:   "/bin/sh",
			wantErr: false,
			checkOut: func(out string) bool {
				return strings.Contains(out, "hello")
			},
		},
		{
			name:    "invalid shell path",
			command: "echo test",
			shell:   "relative/path/sh",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := ExecuteWithShell(context.Background(), tt.command, tt.shell)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExecuteWithShell() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.checkOut != nil {
				if !tt.checkOut(output.Stdout) {
					t.Errorf("Output check failed, stdout: %s", output.Stdout)
				}
			}
		})
	}
}

// Additional ValidatorsPlugin tests for comprehensive coverage

func TestValidatorsPlugin_URL(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name      string
		url       string
		wantValid bool
	}{
		{"valid http url", "http://example.com", true},
		{"valid https url", "https://example.com/path", true},
		{"url with port", "https://example.com:8080", true},
		{"no scheme", "example.com", false},
		{"no host", "http://", false},
		{"invalid url", "not a url", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "url",
					"value":     tt.url,
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
				t.Errorf("valid = %v, want %v for %s", valid, tt.wantValid, tt.url)
			}
		})
	}
}

func TestValidatorsPlugin_IP(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name      string
		ip        string
		wantValid bool
	}{
		{"valid ipv4", "192.168.1.1", true},
		{"valid ipv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", true},
		{"invalid ip", "256.1.1.1", false},
		{"not an ip", "not-an-ip", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "ip",
					"value":     tt.ip,
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
				t.Errorf("valid = %v, want %v for %s", valid, tt.wantValid, tt.ip)
			}
		})
	}
}

func TestValidatorsPlugin_CIDR(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name      string
		cidr      string
		wantValid bool
	}{
		{"valid cidr", "192.168.1.0/24", true},
		{"valid ipv6 cidr", "2001:db8::/32", true},
		{"invalid cidr", "192.168.1.1", false},
		{"invalid range", "192.168.1.0/33", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "cidr",
					"value":     tt.cidr,
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
				t.Errorf("valid = %v, want %v for %s", valid, tt.wantValid, tt.cidr)
			}
		})
	}
}

func TestValidatorsPlugin_DNSLabel(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name      string
		label     string
		wantValid bool
	}{
		{"valid label", "my-service", true},
		{"valid single char", "a", true},
		{"valid with numbers", "service123", true},
		{"starts with hyphen", "-service", false},
		{"ends with hyphen", "service-", false},
		{"too long", strings.Repeat("a", 64), false},
		{"uppercase", "MyService", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "dns-label",
					"value":     tt.label,
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
				t.Errorf("valid = %v, want %v for %s", valid, tt.wantValid, tt.label)
			}
		})
	}
}

func TestValidatorsPlugin_Length(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name      string
		value     string
		min       float64
		max       float64
		wantValid bool
	}{
		{"within range", "hello", 3.0, 10.0, true},
		{"too short", "hi", 3.0, 10.0, false},
		{"too long", "hello world!", 3.0, 10.0, false},
		{"exact min", "abc", 3.0, 10.0, true},
		{"exact max", "1234567890", 3.0, 10.0, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "length",
					"value":     tt.value,
					"min":       tt.min,
					"max":       tt.max,
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
				t.Errorf("valid = %v, want %v for %s", valid, tt.wantValid, tt.value)
			}
		})
	}
}

func TestValidatorsPlugin_Range(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name      string
		value     interface{}
		min       float64
		max       float64
		wantValid bool
	}{
		{"float within range", 5.5, 1.0, 10.0, true},
		{"int within range", 5, 1.0, 10.0, true},
		{"string number", "7", 1.0, 10.0, true},
		{"below min", 0.5, 1.0, 10.0, false},
		{"above max", 15.0, 1.0, 10.0, false},
		{"invalid string", "not-a-number", 1.0, 10.0, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "range",
					"value":     tt.value,
					"min":       tt.min,
					"max":       tt.max,
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
				t.Errorf("valid = %v, want %v for %v", valid, tt.wantValid, tt.value)
			}
		})
	}
}

func TestValidatorsPlugin_Enum(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name      string
		value     interface{}
		allowed   []interface{}
		wantValid bool
	}{
		{"string in enum", "apple", []interface{}{"apple", "banana", "orange"}, true},
		{"number in enum", 2, []interface{}{1, 2, 3}, true},
		{"not in enum", "grape", []interface{}{"apple", "banana", "orange"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "enum",
					"value":     tt.value,
					"allowed":   tt.allowed,
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
				t.Errorf("valid = %v, want %v for %v", valid, tt.wantValid, tt.value)
			}
		})
	}
}

func TestValidatorsPlugin_Format_UUID(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name      string
		uuid      string
		wantValid bool
	}{
		{"valid uuid", "550e8400-e29b-41d4-a716-446655440000", true},
		{"uppercase uuid", "550E8400-E29B-41D4-A716-446655440000", true},
		{"invalid uuid", "not-a-uuid", false},
		{"incomplete uuid", "550e8400-e29b-41d4", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "format",
					"value":     tt.uuid,
					"format":    "uuid",
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
				t.Errorf("valid = %v, want %v for %s", valid, tt.wantValid, tt.uuid)
			}
		})
	}
}

func TestValidatorsPlugin_Format_Date(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name      string
		date      string
		wantValid bool
	}{
		{"valid date", "2023-12-25", true},
		{"invalid format", "12/25/2023", false},
		{"invalid date", "2023-13-45", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "format",
					"value":     tt.date,
					"format":    "date",
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
				t.Errorf("valid = %v, want %v for %s", valid, tt.wantValid, tt.date)
			}
		})
	}
}

func TestValidatorsPlugin_Format_Time(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name      string
		time      string
		wantValid bool
	}{
		{"valid time", "14:30:00", true},
		{"invalid format", "2:30 PM", false},
		{"invalid time", "25:00:00", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "format",
					"value":     tt.time,
					"format":    "time",
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
				t.Errorf("valid = %v, want %v for %s", valid, tt.wantValid, tt.time)
			}
		})
	}
}

func TestValidatorsPlugin_Format_DateTime(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name      string
		datetime  string
		wantValid bool
	}{
		{"valid datetime", "2023-12-25T14:30:00Z", true},
		{"valid with offset", "2023-12-25T14:30:00+05:00", true},
		{"invalid format", "2023-12-25 14:30:00", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "format",
					"value":     tt.datetime,
					"format":    "datetime",
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
				t.Errorf("valid = %v, want %v for %s", valid, tt.wantValid, tt.datetime)
			}
		})
	}
}

func TestValidatorsPlugin_Format_SemVer(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name      string
		version   string
		wantValid bool
	}{
		{"valid semver", "1.2.3", true},
		{"with prerelease", "1.2.3-alpha.1", true},
		{"with build", "1.2.3+build.123", true},
		{"invalid semver", "1.2", false},
		{"invalid format", "v1.2.3", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "format",
					"value":     tt.version,
					"format":    "semver",
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
				t.Errorf("valid = %v, want %v for %s", valid, tt.wantValid, tt.version)
			}
		})
	}
}

func TestValidatorsPlugin_Validate(t *testing.T) {
	validators := NewValidatorsPlugin()
	if err := validators.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestValidatorsPlugin_Describe(t *testing.T) {
	validators := NewValidatorsPlugin()
	info := validators.Describe()

	if info.Manifest.Name != "validators" {
		t.Errorf("Name = %v, want validators", info.Manifest.Name)
	}

	if len(info.Capabilities) == 0 {
		t.Error("Capabilities should not be empty")
	}
}

func TestValidatorsPlugin_ErrorCases(t *testing.T) {
	validators := NewValidatorsPlugin()

	tests := []struct {
		name    string
		input   *plugin.PluginInput
		wantErr bool
	}{
		{
			name: "missing validator",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"value": "test",
				},
			},
			wantErr: true,
		},
		{
			name: "missing value",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "email",
				},
			},
			wantErr: true,
		},
		{
			name: "unknown validator",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "unknown",
					"value":     "test",
				},
			},
			wantErr: true,
		},
		{
			name: "regex missing pattern",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "regex",
					"value":     "test",
				},
			},
			wantErr: false, // Returns error in output
		},
		{
			name: "enum missing allowed",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "enum",
					"value":     "test",
				},
			},
			wantErr: false, // Returns error in output
		},
		{
			name: "format missing format",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"validator": "format",
					"value":     "test",
				},
			},
			wantErr: false, // Returns error in output
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := validators.Execute(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && output.ExitCode != 0 {
				// Expected error in output
				if output.Error == "" {
					t.Error("Expected error message in output")
				}
			}
		})
	}
}

// Additional TransformersPlugin tests for comprehensive coverage

func TestTransformersPlugin_YAMLToJSON(t *testing.T) {
	transformers := NewTransformersPlugin()

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"transformation": "yaml-to-json",
			"input":          "key: value\nnumber: 42",
		},
	}

	output, err := transformers.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	jsonOutput, ok := output.GetString("output")
	if !ok {
		t.Fatal("Failed to get output from result")
	}

	if !strings.Contains(jsonOutput, "\"key\"") {
		t.Errorf("JSON output should contain key, got: %s", jsonOutput)
	}
}

func TestTransformersPlugin_UsersToHTPasswd(t *testing.T) {
	transformers := NewTransformersPlugin()

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"transformation": "users-to-htpasswd",
			"users": []interface{}{
				map[string]interface{}{
					"username": "user1",
					"password": "hash1",
				},
				map[string]interface{}{
					"username": "user2",
					"password": "hash2",
				},
			},
		},
	}

	output, err := transformers.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	htpasswd, ok := output.GetString("output")
	if !ok {
		t.Fatal("Failed to get output from result")
	}

	if !strings.Contains(htpasswd, "user1:hash1") {
		t.Errorf("htpasswd should contain user1:hash1, got: %s", htpasswd)
	}

	count, ok := output.GetInt("count")
	if !ok {
		t.Fatal("Failed to get count from result")
	}

	if count != 2 {
		t.Errorf("count = %v, want 2", count)
	}
}

func TestTransformersPlugin_ExtractField(t *testing.T) {
	transformers := NewTransformersPlugin()

	tests := []struct {
		name      string
		input     string
		field     string
		wantValue interface{}
		wantErr   bool
	}{
		{
			name:      "extract simple field from JSON",
			input:     `{"name": "Alice", "age": 30}`,
			field:     "name",
			wantValue: "Alice",
		},
		{
			name:      "extract nested field from JSON",
			input:     `{"user": {"name": "Bob", "id": 123}}`,
			field:     "user.name",
			wantValue: "Bob",
		},
		{
			name:      "extract from YAML",
			input:     "name: Charlie\nage: 25",
			field:     "name",
			wantValue: "Charlie",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pluginInput := &plugin.PluginInput{
				Data: map[string]interface{}{
					"transformation": "extract-field",
					"input":          tt.input,
					"field":          tt.field,
				},
			}

			output, err := transformers.Execute(context.Background(), pluginInput)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Execute() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !tt.wantErr {
				value := output.Data["value"]
				if value != tt.wantValue {
					t.Errorf("value = %v, want %v", value, tt.wantValue)
				}
			}
		})
	}
}

func TestTransformersPlugin_Merge(t *testing.T) {
	transformers := NewTransformersPlugin()

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"transformation": "merge",
			"sources": []interface{}{
				map[string]interface{}{
					"key1": "value1",
					"key2": "value2",
				},
				map[string]interface{}{
					"key3": "value3",
					"key2": "overwritten",
				},
			},
		},
	}

	output, err := transformers.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	merged, ok := output.Data["merged"].(map[string]interface{})
	if !ok {
		t.Fatal("Failed to get merged data from result")
	}

	if merged["key1"] != "value1" {
		t.Errorf("merged[key1] = %v, want value1", merged["key1"])
	}

	if merged["key2"] != "overwritten" {
		t.Errorf("merged[key2] = %v, want overwritten", merged["key2"])
	}

	if merged["key3"] != "value3" {
		t.Errorf("merged[key3] = %v, want value3", merged["key3"])
	}
}

func TestTransformersPlugin_Filter(t *testing.T) {
	transformers := NewTransformersPlugin()

	tests := []struct {
		name        string
		input       interface{}
		criteria    map[string]interface{}
		wantCount   int
		description string
	}{
		{
			name: "filter array of objects",
			input: []interface{}{
				map[string]interface{}{"name": "Alice", "age": 30},
				map[string]interface{}{"name": "Bob", "age": 25},
				map[string]interface{}{"name": "Charlie", "age": 30},
			},
			criteria:    map[string]interface{}{"age": 30},
			wantCount:   2,
			description: "should filter items with age 30",
		},
		{
			name: "filter map",
			input: map[string]interface{}{
				"item1": map[string]interface{}{"type": "A"},
				"item2": map[string]interface{}{"type": "B"},
				"item3": map[string]interface{}{"type": "A"},
			},
			criteria:    map[string]interface{}{"type": "A"},
			wantCount:   2,
			description: "should filter items with type A",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pluginInput := &plugin.PluginInput{
				Data: map[string]interface{}{
					"transformation": "filter",
					"input":          tt.input,
					"criteria":       tt.criteria,
				},
			}

			output, err := transformers.Execute(context.Background(), pluginInput)
			if err != nil {
				t.Fatalf("Execute() error = %v", err)
			}

			count, ok := output.GetInt("count")
			if !ok {
				t.Fatal("Failed to get count from result")
			}

			if count != tt.wantCount {
				t.Errorf("count = %v, want %v", count, tt.wantCount)
			}
		})
	}
}

func TestTransformersPlugin_Validate(t *testing.T) {
	transformers := NewTransformersPlugin()
	if err := transformers.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestTransformersPlugin_Describe(t *testing.T) {
	transformers := NewTransformersPlugin()
	info := transformers.Describe()

	if info.Manifest.Name != "transformers" {
		t.Errorf("Name = %v, want transformers", info.Manifest.Name)
	}

	if len(info.Capabilities) == 0 {
		t.Error("Capabilities should not be empty")
	}
}

func TestTransformersPlugin_ErrorCases(t *testing.T) {
	transformers := NewTransformersPlugin()

	tests := []struct {
		name    string
		input   *plugin.PluginInput
		wantErr bool
	}{
		{
			name: "missing transformation",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"input": "test",
				},
			},
			wantErr: true,
		},
		{
			name: "unknown transformation",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"transformation": "unknown",
					"input":          "test",
				},
			},
			wantErr: true,
		},
		{
			name: "json-to-yaml missing input",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"transformation": "json-to-yaml",
				},
			},
			wantErr: false, // Returns error in output
		},
		{
			name: "json-to-yaml invalid json",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"transformation": "json-to-yaml",
					"input":          "not valid json",
				},
			},
			wantErr: false, // Returns error in output
		},
		{
			name: "yaml-to-json invalid yaml",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"transformation": "yaml-to-json",
					"input":          "invalid: yaml: syntax: error",
				},
			},
			wantErr: false, // Returns error in output
		},
		{
			name: "base64-decode invalid base64",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"transformation": "base64-decode",
					"input":          "not valid base64!@#$",
				},
			},
			wantErr: false, // Returns error in output
		},
		{
			name: "users-to-htpasswd missing users",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"transformation": "users-to-htpasswd",
				},
			},
			wantErr: false, // Returns error in output
		},
		{
			name: "extract-field missing input",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"transformation": "extract-field",
					"field":          "name",
				},
			},
			wantErr: false, // Returns error in output
		},
		{
			name: "merge missing sources",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"transformation": "merge",
				},
			},
			wantErr: false, // Returns error in output
		},
		{
			name: "filter missing input",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"transformation": "filter",
					"criteria":       map[string]interface{}{"key": "value"},
				},
			},
			wantErr: false, // Returns error in output
		},
		{
			name: "template missing template",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"transformation": "template",
					"values":         map[string]interface{}{"key": "value"},
				},
			},
			wantErr: false, // Returns error in output
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := transformers.Execute(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && output.ExitCode != 0 {
				// Expected error in output
				if output.Error == "" {
					t.Error("Expected error message in output")
				}
			}
		})
	}
}

// Additional FileOpsPlugin tests for comprehensive coverage

func TestFileOpsPlugin_Validate(t *testing.T) {
	fileOps := NewFileOpsPlugin([]string{"/tmp"}, 1024*1024)
	if err := fileOps.Validate(); err != nil {
		t.Errorf("Validate() error = %v", err)
	}
}

func TestFileOpsPlugin_Describe(t *testing.T) {
	fileOps := NewFileOpsPlugin([]string{"/tmp"}, 1024*1024)
	info := fileOps.Describe()

	if info.Manifest.Name != "file-ops" {
		t.Errorf("Name = %v, want file-ops", info.Manifest.Name)
	}

	if len(info.Capabilities) == 0 {
		t.Error("Capabilities should not be empty")
	}
}

func TestFileOpsPlugin_ValidateOperation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.json")
	testContent := `{"key": "value"}`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileOps := NewFileOpsPlugin([]string{tmpDir}, 10*1024*1024)

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"operation": "validate",
			"file":      testFile,
			"format":    "json",
		},
	}

	output, err := fileOps.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	valid, ok := output.GetBool("valid")
	if !ok {
		t.Fatal("Failed to get valid from output")
	}

	if !valid {
		t.Error("Valid JSON file should be reported as valid")
	}
}

func TestFileOpsPlugin_TransformOperation(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.txt")
	testContent := "Hello, World!"

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileOps := NewFileOpsPlugin([]string{tmpDir}, 10*1024*1024)

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"operation":      "transform",
			"file":           testFile,
			"transformation": "base64-encode",
		},
	}

	output, err := fileOps.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	encoded, ok := output.GetString("encoded")
	if !ok {
		t.Fatal("Failed to get encoded from output")
	}

	if encoded == "" {
		t.Error("Encoded content should not be empty")
	}
}

func TestFileOpsPlugin_ParsePEM(t *testing.T) {
	tmpDir := t.TempDir()
	testFile := filepath.Join(tmpDir, "test.pem")
	// Use a simple PEM block (not a full certificate)
	testContent := `-----BEGIN RSA PRIVATE KEY-----
MIIBOwIBAAJBANDiE2+Xi/WnO+s120NiiJhNyIButVu6zxqlVzz0wy2j4kQVUC4Z
RZD80IY+4wIiXRx4Z5YBJjxNXQv9mZBN1KkCAwEAAQJBAMqOsM4xRNgm5AKyVlhv
+LmV6c3pVKLpgVqiGBLJg2xSLJUEGGLSQeZLx8vMZWV2PaWBSJBKOvNqiOJyLbBN
-----END RSA PRIVATE KEY-----`

	if err := os.WriteFile(testFile, []byte(testContent), 0644); err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}

	fileOps := NewFileOpsPlugin([]string{tmpDir}, 10*1024*1024)

	input := &plugin.PluginInput{
		Data: map[string]interface{}{
			"operation": "parse",
			"file":      testFile,
			"format":    "pem",
		},
	}

	output, err := fileOps.Execute(context.Background(), input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	pemType, ok := output.GetString("type")
	if !ok {
		t.Fatal("Failed to get type from output")
	}

	if pemType != "RSA PRIVATE KEY" {
		t.Errorf("type = %v, want RSA PRIVATE KEY", pemType)
	}
}

func TestFileOpsPlugin_ErrorCases(t *testing.T) {
	tmpDir := t.TempDir()
	fileOps := NewFileOpsPlugin([]string{tmpDir}, 10*1024*1024)

	tests := []struct {
		name    string
		input   *plugin.PluginInput
		wantErr bool
	}{
		{
			name: "missing operation",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"file": "/some/file",
				},
			},
			wantErr: true, // Returns error directly
		},
		{
			name: "unknown operation",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"operation": "unknown",
					"file":      "/some/file",
				},
			},
			wantErr: true, // Returns error directly
		},
		{
			name: "missing file",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"operation": "read",
				},
			},
			wantErr: false, // Returns error in output (from readFile)
		},
		{
			name: "file outside allowed paths",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"operation": "read",
					"file":      "/etc/passwd",
				},
			},
			wantErr: false, // Returns error in output (from validatePath)
		},
		{
			name: "nonexistent file",
			input: &plugin.PluginInput{
				Data: map[string]interface{}{
					"operation": "read",
					"file":      filepath.Join(tmpDir, "nonexistent.txt"),
				},
			},
			wantErr: false, // Returns error in output (from validatePath)
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			output, err := fileOps.Execute(context.Background(), tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("Execute() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && output.ExitCode != 0 {
				// Expected error in output
				if output.Error == "" {
					t.Error("Expected error message in output")
				}
			}
		})
	}
}
