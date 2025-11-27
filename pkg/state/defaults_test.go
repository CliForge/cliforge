package state

import (
	"os"
	"testing"
)

func TestNewDefaultsProvider(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")
	provider := NewDefaultsProvider(mgr)

	if provider.stateMgr != mgr {
		t.Error("Expected state manager to be set")
	}

	if provider.configDefaults == nil {
		t.Error("Expected config defaults to be initialized")
	}

	if provider.builtinDefaults == nil {
		t.Error("Expected builtin defaults to be initialized")
	}
}

func TestDefaultsProviderBuiltinDefaults(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")
	provider := NewDefaultsProvider(mgr)

	builtin := map[string]string{
		"region": "us-east-1",
		"format": "json",
	}
	provider.SetBuiltinDefaults(builtin)

	value, exists := provider.Get("region")
	if !exists {
		t.Error("Expected to get builtin default")
	}

	if value != "us-east-1" {
		t.Errorf("Expected 'us-east-1', got %s", value)
	}
}

func TestDefaultsProviderConfigOverridesBuiltin(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")
	provider := NewDefaultsProvider(mgr)

	builtin := map[string]string{
		"region": "us-east-1",
	}
	config := map[string]string{
		"region": "us-west-2",
	}

	provider.SetBuiltinDefaults(builtin)
	provider.SetConfigDefaults(config)

	value, _ := provider.Get("region")
	if value != "us-west-2" {
		t.Errorf("Expected config to override builtin, got %s", value)
	}
}

func TestDefaultsProviderContextOverridesConfig(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")
	provider := NewDefaultsProvider(mgr)

	builtin := map[string]string{"region": "us-east-1"}
	config := map[string]string{"region": "us-west-2"}

	provider.SetBuiltinDefaults(builtin)
	provider.SetConfigDefaults(config)

	// Set context value
	ctx := mgr.GetCurrentContext()
	ctx.Set("region", "eu-west-1")

	value, _ := provider.Get("region")
	if value != "eu-west-1" {
		t.Errorf("Expected context to override config, got %s", value)
	}
}

func TestDefaultsProviderRecentOverridesContext(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")
	provider := NewDefaultsProvider(mgr)

	config := map[string]string{"region": "us-west-2"}
	provider.SetConfigDefaults(config)

	ctx := mgr.GetCurrentContext()
	ctx.Set("region", "eu-west-1")

	// Add recent value
	mgr.AddRecentValue("region", "ap-south-1")

	value, _ := provider.Get("region")
	if value != "ap-south-1" {
		t.Errorf("Expected recent to override context, got %s", value)
	}
}

func TestDefaultsProviderGetWithPriority(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")
	provider := NewDefaultsProvider(mgr)

	// Test builtin priority
	provider.SetBuiltinDefaults(map[string]string{"key": "builtin-value"})
	value, priority, exists := provider.GetWithPriority("key")
	if !exists {
		t.Error("Expected key to exist")
	}
	if priority != PriorityBuiltin {
		t.Errorf("Expected builtin priority, got %s", priority.String())
	}
	if value != "builtin-value" {
		t.Error("Expected correct value")
	}

	// Test config priority
	provider.SetConfigDefaults(map[string]string{"key": "config-value"})
	value, priority, _ = provider.GetWithPriority("key")
	if priority != PriorityConfig {
		t.Errorf("Expected config priority, got %s", priority.String())
	}
	if value != "config-value" {
		t.Error("Expected config value")
	}

	// Test context priority
	ctx := mgr.GetCurrentContext()
	ctx.Set("key", "context-value")
	value, priority, _ = provider.GetWithPriority("key")
	if priority != PriorityContext {
		t.Errorf("Expected context priority, got %s", priority.String())
	}
	if value != "context-value" {
		t.Error("Expected context value")
	}

	// Test recent priority
	mgr.AddRecentValue("key", "recent-value")
	value, priority, _ = provider.GetWithPriority("key")
	if priority != PriorityRecent {
		t.Errorf("Expected recent priority, got %s", priority.String())
	}
	if value != "recent-value" {
		t.Error("Expected recent value")
	}
}

func TestDefaultsProviderGetAll(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")
	provider := NewDefaultsProvider(mgr)

	builtin := map[string]string{
		"region": "us-east-1",
		"format": "json",
	}
	config := map[string]string{
		"region": "us-west-2",
		"color":  "always",
	}

	provider.SetBuiltinDefaults(builtin)
	provider.SetConfigDefaults(config)

	all := provider.GetAll()

	if all["format"] != "json" {
		t.Error("Expected builtin format to be in merged defaults")
	}

	if all["region"] != "us-west-2" {
		t.Error("Expected config to override builtin for region")
	}

	if all["color"] != "always" {
		t.Error("Expected config-only key to be present")
	}
}

func TestDefaultsProviderGetWithFallback(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")
	provider := NewDefaultsProvider(mgr)

	value := provider.GetWithFallback("nonexistent", "fallback-value")
	if value != "fallback-value" {
		t.Errorf("Expected fallback value, got %s", value)
	}

	provider.SetBuiltinDefaults(map[string]string{"key": "value"})
	value = provider.GetWithFallback("key", "fallback-value")
	if value != "value" {
		t.Errorf("Expected actual value, got %s", value)
	}
}

func TestDefaultsProviderGetInt(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")
	provider := NewDefaultsProvider(mgr)

	provider.SetBuiltinDefaults(map[string]string{"limit": "100"})

	value := provider.GetInt("limit", 50)
	if value != 100 {
		t.Errorf("Expected 100, got %d", value)
	}

	// Test fallback
	value = provider.GetInt("nonexistent", 50)
	if value != 50 {
		t.Errorf("Expected fallback 50, got %d", value)
	}

	// Test invalid int
	provider.SetBuiltinDefaults(map[string]string{"invalid": "not-a-number"})
	value = provider.GetInt("invalid", 50)
	if value != 50 {
		t.Errorf("Expected fallback for invalid int, got %d", value)
	}
}

func TestDefaultsProviderGetBool(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")
	provider := NewDefaultsProvider(mgr)

	provider.SetBuiltinDefaults(map[string]string{"enabled": "true"})

	value := provider.GetBool("enabled", false)
	if !value {
		t.Error("Expected true")
	}

	// Test fallback
	value = provider.GetBool("nonexistent", false)
	if value {
		t.Error("Expected fallback false")
	}

	// Test invalid bool
	provider.SetBuiltinDefaults(map[string]string{"invalid": "not-a-bool"})
	value = provider.GetBool("invalid", false)
	if value {
		t.Error("Expected fallback for invalid bool")
	}
}

func TestDefaultsResolverEnvOverride(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")
	provider := NewDefaultsProvider(mgr)
	resolver := NewDefaultsResolver(provider)

	provider.SetBuiltinDefaults(map[string]string{"region": "us-east-1"})
	resolver.SetEnvVar("region", "TEST_REGION")

	// Without env var
	value, _ := resolver.Resolve("region")
	if value != "us-east-1" {
		t.Error("Expected builtin value without env var")
	}

	// With env var
	_ = os.Setenv("TEST_REGION", "eu-west-1")
	defer func() { _ = os.Unsetenv("TEST_REGION") }()

	value, _ = resolver.Resolve("region")
	if value != "eu-west-1" {
		t.Errorf("Expected env var to override, got %s", value)
	}
}

func TestDefaultsResolverWithPriority(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")
	provider := NewDefaultsProvider(mgr)
	resolver := NewDefaultsResolver(provider)

	provider.SetBuiltinDefaults(map[string]string{"key": "builtin"})
	resolver.SetEnvVar("key", "TEST_KEY")

	// Without env var
	_, priority, _ := resolver.ResolveWithPriority("key")
	if priority != PriorityBuiltin {
		t.Error("Expected builtin priority without env var")
	}

	// With env var
	_ = os.Setenv("TEST_KEY", "env-value")
	defer func() { _ = os.Unsetenv("TEST_KEY") }()

	value, priority, _ := resolver.ResolveWithPriority("key")
	if priority != PriorityEnv {
		t.Errorf("Expected env priority, got %s", priority.String())
	}
	if value != "env-value" {
		t.Error("Expected env value")
	}
}

func TestDefaultsBuilder(t *testing.T) {
	builder := NewDefaultsBuilder()

	defaults := builder.
		Set("region", "us-east-1").
		SetInt("limit", 100).
		SetBool("enabled", true).
		SetMultiple(map[string]string{
			"format": "json",
			"color":  "always",
		}).
		Build()

	if defaults["region"] != "us-east-1" {
		t.Error("Expected region to be set")
	}

	if defaults["limit"] != "100" {
		t.Error("Expected limit to be set as string")
	}

	if defaults["enabled"] != "true" {
		t.Error("Expected enabled to be set as string")
	}

	if defaults["format"] != "json" {
		t.Error("Expected format from SetMultiple")
	}
}

func TestMergeDefaults(t *testing.T) {
	map1 := map[string]string{
		"region": "us-east-1",
		"format": "json",
	}

	map2 := map[string]string{
		"region": "us-west-2",
		"color":  "always",
	}

	merged := MergeDefaults(map1, map2)

	if merged["region"] != "us-west-2" {
		t.Error("Expected later map to override region")
	}

	if merged["format"] != "json" {
		t.Error("Expected format from first map")
	}

	if merged["color"] != "always" {
		t.Error("Expected color from second map")
	}
}

func TestFilterDefaults(t *testing.T) {
	defaults := map[string]string{
		"http.timeout":  "30s",
		"http.retry":    "3",
		"output.format": "json",
		"output.color":  "always",
	}

	filtered := FilterDefaults(defaults, "http.")

	if len(filtered) != 2 {
		t.Errorf("Expected 2 filtered items, got %d", len(filtered))
	}

	if filtered["http.timeout"] != "30s" {
		t.Error("Expected http.timeout to be included")
	}

	if filtered["output.format"] != "" {
		t.Error("Expected output.format to be excluded")
	}
}

func TestTransformDefaults(t *testing.T) {
	defaults := map[string]string{
		"region": "us-east-1",
		"format": "json",
	}

	transformed := TransformDefaults(defaults, func(key string) string {
		return "prefix." + key
	})

	if transformed["prefix.region"] != "us-east-1" {
		t.Error("Expected transformed key")
	}

	if transformed["region"] != "" {
		t.Error("Expected original key to be replaced")
	}
}

func TestValidateDefaults(t *testing.T) {
	defaults := map[string]string{
		"region": "us-east-1",
		"format": "json",
	}

	allowed := []string{"region", "format", "color"}

	err := ValidateDefaults(defaults, allowed)
	if err != nil {
		t.Errorf("Expected validation to pass, got error: %v", err)
	}

	defaults["invalid"] = "value"
	err = ValidateDefaults(defaults, allowed)
	if err == nil {
		t.Error("Expected validation to fail for invalid key")
	}
}

func TestDefaultsSnapshot(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_STATE_HOME", tmpDir)
	defer func() { _ = os.Unsetenv("XDG_STATE_HOME") }()

	mgr, _ := NewManager("testcli")
	provider := NewDefaultsProvider(mgr)

	provider.SetBuiltinDefaults(map[string]string{
		"region": "us-east-1",
	})

	provider.SetConfigDefaults(map[string]string{
		"format": "json",
	})

	ctx := mgr.GetCurrentContext()
	ctx.Set("color", "always")

	snapshot := provider.Snapshot()

	if snapshot.BuiltinDefaults["region"] != "us-east-1" {
		t.Error("Expected builtin defaults in snapshot")
	}

	if snapshot.ConfigDefaults["format"] != "json" {
		t.Error("Expected config defaults in snapshot")
	}

	if snapshot.ContextDefaults["color"] != "always" {
		t.Error("Expected context defaults in snapshot")
	}

	if snapshot.Context != "default" {
		t.Errorf("Expected context name to be 'default', got %s", snapshot.Context)
	}
}

func TestPriorityString(t *testing.T) {
	tests := []struct {
		priority DefaultPriority
		expected string
	}{
		{PriorityNone, "none"},
		{PriorityBuiltin, "builtin"},
		{PriorityConfig, "config"},
		{PriorityContext, "context"},
		{PriorityRecent, "recent"},
	}

	for _, test := range tests {
		if test.priority.String() != test.expected {
			t.Errorf("Expected %s, got %s", test.expected, test.priority.String())
		}
	}
}
