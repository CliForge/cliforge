package plugin

import (
	"context"
	"testing"
)

func TestNewRegistry(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})

	registry := NewRegistry(tmpDir, pm)
	if registry == nil {
		t.Fatal("NewRegistry() returned nil")
	}

	if registry.pluginDir != tmpDir {
		t.Errorf("pluginDir = %v, want %v", registry.pluginDir, tmpDir)
	}
}

func TestRegistry_Register(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})
	registry := NewRegistry(tmpDir, pm)

	mockPlugin := &MockPlugin{
		name: "test-plugin",
		info: &PluginInfo{
			Manifest: PluginManifest{
				Name:    "test-plugin",
				Version: "1.0.0",
				Type:    PluginTypeBuiltin,
			},
			Status: PluginStatusReady,
		},
	}

	err := registry.Register(mockPlugin)
	if err != nil {
		t.Fatalf("Register() error = %v", err)
	}

	// Verify plugin was registered
	plugin, err := registry.Get("test-plugin")
	if err != nil {
		t.Fatalf("Get() error = %v", err)
	}

	if plugin == nil {
		t.Fatal("Get() returned nil plugin")
	}
}

func TestRegistry_RegisterDuplicate(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})
	registry := NewRegistry(tmpDir, pm)

	mockPlugin := &MockPlugin{name: "test-plugin"}

	// First registration should succeed
	err := registry.Register(mockPlugin)
	if err != nil {
		t.Fatalf("First Register() error = %v", err)
	}

	// Second registration should fail
	err = registry.Register(mockPlugin)
	if err == nil {
		t.Error("Register() should fail for duplicate plugin")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})
	registry := NewRegistry(tmpDir, pm)

	mockPlugin := &MockPlugin{name: "test-plugin"}
	registry.Register(mockPlugin)

	err := registry.Unregister("test-plugin")
	if err != nil {
		t.Fatalf("Unregister() error = %v", err)
	}

	// Verify plugin was unregistered
	_, err = registry.Get("test-plugin")
	if err == nil {
		t.Error("Get() should fail after unregister")
	}
}

func TestRegistry_List(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})
	registry := NewRegistry(tmpDir, pm)

	// Register multiple plugins
	registry.Register(&MockPlugin{name: "plugin1"})
	registry.Register(&MockPlugin{name: "plugin2"})
	registry.Register(&MockPlugin{name: "plugin3"})

	names := registry.List()
	if len(names) != 3 {
		t.Errorf("List() count = %v, want 3", len(names))
	}

	// Verify all plugin names are present
	nameMap := make(map[string]bool)
	for _, name := range names {
		nameMap[name] = true
	}

	for _, expected := range []string{"plugin1", "plugin2", "plugin3"} {
		if !nameMap[expected] {
			t.Errorf("List() missing plugin: %s", expected)
		}
	}
}

func TestRegistry_GetManifest(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})
	registry := NewRegistry(tmpDir, pm)

	mockPlugin := &MockPlugin{
		name: "test-plugin",
		info: &PluginInfo{
			Manifest: PluginManifest{
				Name:        "test-plugin",
				Version:     "1.0.0",
				Type:        PluginTypeBuiltin,
				Description: "Test plugin",
			},
			Status: PluginStatusReady,
		},
	}
	registry.Register(mockPlugin)

	manifest, err := registry.GetManifest("test-plugin")
	if err != nil {
		t.Fatalf("GetManifest() error = %v", err)
	}

	if manifest.Name != "test-plugin" {
		t.Errorf("Manifest.Name = %v, want test-plugin", manifest.Name)
	}

	if manifest.Version != "1.0.0" {
		t.Errorf("Manifest.Version = %v, want 1.0.0", manifest.Version)
	}
}

func TestRegistry_Execute(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})
	registry := NewRegistry(tmpDir, pm)

	// Create a mock plugin that returns specific output
	mockPlugin := &MockPlugin{
		name: "test-plugin",
		executeFunc: func(ctx context.Context, input *PluginInput) (*PluginOutput, error) {
			return &PluginOutput{
				ExitCode: 0,
				Data: map[string]interface{}{
					"result": "success",
				},
			}, nil
		},
	}
	registry.Register(mockPlugin)

	input := &PluginInput{
		Data: map[string]interface{}{
			"test": "value",
		},
	}

	output, err := registry.Execute(context.Background(), "test-plugin", input)
	if err != nil {
		t.Fatalf("Execute() error = %v", err)
	}

	if output.ExitCode != 0 {
		t.Errorf("ExitCode = %v, want 0", output.ExitCode)
	}

	if result, ok := output.GetString("result"); !ok || result != "success" {
		t.Errorf("Result = %v, want success", result)
	}
}

func TestRegistry_GetPluginInfo(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})
	registry := NewRegistry(tmpDir, pm)

	mockPlugin := &MockPlugin{
		name: "test-plugin",
		info: &PluginInfo{
			Manifest: PluginManifest{
				Name:    "test-plugin",
				Version: "1.0.0",
				Type:    PluginTypeBuiltin,
			},
			Capabilities: []string{"test"},
			Status:       PluginStatusReady,
		},
	}
	registry.Register(mockPlugin)

	info, err := registry.GetPluginInfo("test-plugin")
	if err != nil {
		t.Fatalf("GetPluginInfo() error = %v", err)
	}

	if info.Manifest.Name != "test-plugin" {
		t.Errorf("Manifest.Name = %v, want test-plugin", info.Manifest.Name)
	}

	if len(info.Capabilities) != 1 {
		t.Errorf("Capabilities count = %v, want 1", len(info.Capabilities))
	}
}

func TestRegistry_ListPluginInfo(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})
	registry := NewRegistry(tmpDir, pm)

	// Register multiple plugins
	registry.Register(&MockPlugin{name: "plugin1"})
	registry.Register(&MockPlugin{name: "plugin2"})

	infos, err := registry.ListPluginInfo()
	if err != nil {
		t.Fatalf("ListPluginInfo() error = %v", err)
	}

	if len(infos) != 2 {
		t.Errorf("ListPluginInfo() count = %v, want 2", len(infos))
	}
}

func TestRegistry_ValidateManifest(t *testing.T) {
	tmpDir := t.TempDir()
	pm, _ := NewPermissionManager(tmpDir, &AutoApprover{})
	registry := NewRegistry(tmpDir, pm)

	tests := []struct {
		name     string
		manifest PluginManifest
		wantErr  bool
	}{
		{
			name: "valid builtin manifest",
			manifest: PluginManifest{
				Name:    "test",
				Version: "1.0.0",
				Type:    PluginTypeBuiltin,
			},
			wantErr: false,
		},
		{
			name: "valid binary manifest",
			manifest: PluginManifest{
				Name:       "test",
				Version:    "1.0.0",
				Type:       PluginTypeBinary,
				Executable: "/path/to/exec",
			},
			wantErr: false,
		},
		{
			name: "missing name",
			manifest: PluginManifest{
				Version: "1.0.0",
				Type:    PluginTypeBuiltin,
			},
			wantErr: true,
		},
		{
			name: "missing version",
			manifest: PluginManifest{
				Name: "test",
				Type: PluginTypeBuiltin,
			},
			wantErr: true,
		},
		{
			name: "missing executable for binary",
			manifest: PluginManifest{
				Name:    "test",
				Version: "1.0.0",
				Type:    PluginTypeBinary,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := registry.validateManifest(&tt.manifest)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateManifest() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestBinaryPlugin_Describe(t *testing.T) {
	manifest := PluginManifest{
		Name:       "test-binary",
		Version:    "1.0.0",
		Type:       PluginTypeBinary,
		Executable: "/usr/bin/test",
	}

	plugin := NewBinaryPlugin(manifest, "/usr/bin/test")
	info := plugin.Describe()

	if info.Manifest.Name != "test-binary" {
		t.Errorf("Name = %v, want test-binary", info.Manifest.Name)
	}

	if info.Manifest.Type != PluginTypeBinary {
		t.Errorf("Type = %v, want binary", info.Manifest.Type)
	}

	if len(info.Capabilities) == 0 {
		t.Error("Capabilities should not be empty")
	}
}
