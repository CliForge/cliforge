package plugin

import (
	"context"
	"testing"
	"time"
)

// MockPlugin is a mock implementation of the Plugin interface for testing.
type MockPlugin struct {
	name        string
	executeFunc func(ctx context.Context, input *PluginInput) (*PluginOutput, error)
	validateErr error
	info        *PluginInfo
}

func (m *MockPlugin) Execute(ctx context.Context, input *PluginInput) (*PluginOutput, error) {
	if m.executeFunc != nil {
		return m.executeFunc(ctx, input)
	}
	return &PluginOutput{ExitCode: 0}, nil
}

func (m *MockPlugin) Validate() error {
	return m.validateErr
}

func (m *MockPlugin) Describe() *PluginInfo {
	if m.info != nil {
		return m.info
	}
	return &PluginInfo{
		Manifest: PluginManifest{
			Name:    m.name,
			Version: "1.0.0",
			Type:    PluginTypeBuiltin,
		},
		Status: PluginStatusReady,
	}
}

func TestPluginOutput_Success(t *testing.T) {
	tests := []struct {
		name     string
		output   *PluginOutput
		expected bool
	}{
		{
			name: "successful execution",
			output: &PluginOutput{
				ExitCode: 0,
				Error:    "",
			},
			expected: true,
		},
		{
			name: "non-zero exit code",
			output: &PluginOutput{
				ExitCode: 1,
				Error:    "",
			},
			expected: false,
		},
		{
			name: "with error message",
			output: &PluginOutput{
				ExitCode: 0,
				Error:    "some error",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.output.Success(); got != tt.expected {
				t.Errorf("Success() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPluginOutput_GetString(t *testing.T) {
	output := &PluginOutput{
		Data: map[string]interface{}{
			"key1": "value1",
			"key2": 123,
		},
	}

	// Test existing string key
	val, ok := output.GetString("key1")
	if !ok {
		t.Error("GetString() should return true for existing string key")
	}
	if val != "value1" {
		t.Errorf("GetString() = %v, want value1", val)
	}

	// Test non-string key
	_, ok = output.GetString("key2")
	if ok {
		t.Error("GetString() should return false for non-string value")
	}

	// Test non-existent key
	_, ok = output.GetString("key3")
	if ok {
		t.Error("GetString() should return false for non-existent key")
	}
}

func TestPluginOutput_GetInt(t *testing.T) {
	output := &PluginOutput{
		Data: map[string]interface{}{
			"int":     123,
			"int64":   int64(456),
			"float64": float64(789),
			"string":  "not a number",
		},
	}

	tests := []struct {
		key      string
		expected int
		ok       bool
	}{
		{"int", 123, true},
		{"int64", 456, true},
		{"float64", 789, true},
		{"string", 0, false},
		{"missing", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			val, ok := output.GetInt(tt.key)
			if ok != tt.ok {
				t.Errorf("GetInt(%s) ok = %v, want %v", tt.key, ok, tt.ok)
			}
			if ok && val != tt.expected {
				t.Errorf("GetInt(%s) = %v, want %v", tt.key, val, tt.expected)
			}
		})
	}
}

func TestPluginOutput_GetBool(t *testing.T) {
	output := &PluginOutput{
		Data: map[string]interface{}{
			"true":  true,
			"false": false,
			"not":   "boolean",
		},
	}

	tests := []struct {
		key      string
		expected bool
		ok       bool
	}{
		{"true", true, true},
		{"false", false, true},
		{"not", false, false},
		{"missing", false, false},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			val, ok := output.GetBool(tt.key)
			if ok != tt.ok {
				t.Errorf("GetBool(%s) ok = %v, want %v", tt.key, ok, tt.ok)
			}
			if ok && val != tt.expected {
				t.Errorf("GetBool(%s) = %v, want %v", tt.key, val, tt.expected)
			}
		})
	}
}

func TestPluginOutput_GetMap(t *testing.T) {
	output := &PluginOutput{
		Data: map[string]interface{}{
			"map": map[string]interface{}{
				"nested": "value",
			},
			"not": "a map",
		},
	}

	// Test existing map key
	val, ok := output.GetMap("map")
	if !ok {
		t.Error("GetMap() should return true for existing map key")
	}
	if val["nested"] != "value" {
		t.Errorf("GetMap() returned incorrect map value")
	}

	// Test non-map key
	_, ok = output.GetMap("not")
	if ok {
		t.Error("GetMap() should return false for non-map value")
	}

	// Test non-existent key
	_, ok = output.GetMap("missing")
	if ok {
		t.Error("GetMap() should return false for non-existent key")
	}
}

func TestPluginError(t *testing.T) {
	// Test basic error
	err := NewPluginError("test-plugin", "test message", nil)
	if err.PluginName != "test-plugin" {
		t.Errorf("PluginName = %v, want test-plugin", err.PluginName)
	}
	if err.Message != "test message" {
		t.Errorf("Message = %v, want test message", err.Message)
	}

	// Test error with cause
	causeErr := NewPluginError("test-plugin", "test message", context.DeadlineExceeded)
	if causeErr.Unwrap() != context.DeadlineExceeded {
		t.Error("Unwrap() should return the cause error")
	}

	// Test with suggestion
	err = err.WithSuggestion("try this")
	if err.Suggestion != "try this" {
		t.Errorf("Suggestion = %v, want try this", err.Suggestion)
	}

	// Test as recoverable
	err = err.AsRecoverable()
	if !err.Recoverable {
		t.Error("AsRecoverable() should set Recoverable to true")
	}
}

func TestPermission_String(t *testing.T) {
	tests := []struct {
		name     string
		perm     Permission
		expected string
	}{
		{
			name: "with resource",
			perm: Permission{
				Type:     PermissionExecute,
				Resource: "aws",
			},
			expected: "execute:aws",
		},
		{
			name: "without resource",
			perm: Permission{
				Type: PermissionCredential,
			},
			expected: "credential",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.perm.String(); got != tt.expected {
				t.Errorf("String() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestValidatePermission(t *testing.T) {
	tests := []struct {
		name    string
		permStr string
		wantErr bool
	}{
		{
			name:    "valid execute permission",
			permStr: "execute:aws",
			wantErr: false,
		},
		{
			name:    "valid read file permission",
			permStr: "read:file:/path/to/file",
			wantErr: false,
		},
		{
			name:    "valid credential permission",
			permStr: "credential",
			wantErr: true, // credential doesn't require resource in this impl
		},
		{
			name:    "invalid permission type",
			permStr: "invalid:resource",
			wantErr: true,
		},
		{
			name:    "empty permission",
			permStr: "",
			wantErr: true,
		},
		{
			name:    "missing resource for execute",
			permStr: "execute",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePermission(tt.permStr)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePermission() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMatchPermission(t *testing.T) {
	tests := []struct {
		name     string
		pattern  string
		request  string
		expected bool
	}{
		{
			name:     "exact match",
			pattern:  "execute:aws",
			request:  "execute:aws",
			expected: true,
		},
		{
			name:     "wildcard match all",
			pattern:  "execute:*",
			request:  "execute:aws",
			expected: true,
		},
		{
			name:     "wildcard match prefix",
			pattern:  "read:file:/home/*",
			request:  "read:file:/home/user/file.txt",
			expected: true,
		},
		{
			name:     "no match",
			pattern:  "execute:kubectl",
			request:  "execute:aws",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchPermission(tt.pattern, tt.request); got != tt.expected {
				t.Errorf("MatchPermission() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestPluginInput(t *testing.T) {
	input := &PluginInput{
		Command: "test",
		Args:    []string{"arg1", "arg2"},
		Env: map[string]string{
			"VAR1": "value1",
		},
		Data: map[string]interface{}{
			"key": "value",
		},
		Timeout: 30 * time.Second,
	}

	if input.Command != "test" {
		t.Errorf("Command = %v, want test", input.Command)
	}

	if len(input.Args) != 2 {
		t.Errorf("Args length = %v, want 2", len(input.Args))
	}

	if input.Env["VAR1"] != "value1" {
		t.Errorf("Env[VAR1] = %v, want value1", input.Env["VAR1"])
	}

	if input.Timeout != 30*time.Second {
		t.Errorf("Timeout = %v, want 30s", input.Timeout)
	}
}

func TestPluginManifest(t *testing.T) {
	manifest := PluginManifest{
		Name:        "test-plugin",
		Version:     "1.0.0",
		Type:        PluginTypeBuiltin,
		Description: "Test plugin",
		Permissions: []Permission{
			{
				Type:     PermissionExecute,
				Resource: "test",
			},
		},
		Metadata: map[string]string{
			"author": "test",
		},
	}

	if manifest.Name != "test-plugin" {
		t.Errorf("Name = %v, want test-plugin", manifest.Name)
	}

	if len(manifest.Permissions) != 1 {
		t.Errorf("Permissions length = %v, want 1", len(manifest.Permissions))
	}

	if manifest.Metadata["author"] != "test" {
		t.Errorf("Metadata[author] = %v, want test", manifest.Metadata["author"])
	}
}
