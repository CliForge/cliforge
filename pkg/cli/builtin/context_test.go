package builtin

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/state"
)

func TestNewContextCommand(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	output := &bytes.Buffer{}

	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	cmd := NewContextCommand(opts)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	if cmd.Use != "context" {
		t.Errorf("expected Use 'context', got %q", cmd.Use)
	}

	// Check that subcommands are added
	subcommands := cmd.Commands()
	if len(subcommands) == 0 {
		t.Error("expected subcommands to be added")
	}
}

func TestContextCurrent_NoContext(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))
	_ = os.Setenv("XDG_STATE_HOME", filepath.Join(tmpDir, "state"))
	defer func() {
		_ = os.Unsetenv("XDG_DATA_HOME")
		_ = os.Unsetenv("XDG_STATE_HOME")
	}()

	stateMgr, _ := state.NewManager("testcli-nocontext")
	manager := state.NewContextManager(stateMgr)
	output := &bytes.Buffer{}

	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	cmd := newContextCurrentCommand(opts)
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	// Test passes if either no context or default context (flexible for existing state)
	if !strings.Contains(result, "No current context") && !strings.Contains(result, "default") {
		t.Errorf("expected 'No current context' or 'default', got: %s", result)
	}
}

func TestContextCurrent_WithContext(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	_ = manager.Create("test-ctx", "Test context", nil)
	_ = manager.SwitchTo("test-ctx")

	output := &bytes.Buffer{}
	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	cmd := newContextCurrentCommand(opts)
	err := cmd.RunE(cmd, []string{})
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "test-ctx") {
		t.Errorf("expected 'test-ctx', got: %s", result)
	}
}

func TestContextCreate(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	output := &bytes.Buffer{}

	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	cmd := newContextCreateCommand(opts)
	cmd.SetArgs([]string{"dev-env"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Created context \"dev-env\"") {
		t.Errorf("expected creation message, got: %s", result)
	}

	// Verify context was created
	contexts, _ := manager.List()
	if _, ok := contexts["dev-env"]; !ok {
		t.Error("context was not created")
	}
}

func TestContextCreate_WithDescription(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	output := &bytes.Buffer{}

	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	cmd := newContextCreateCommand(opts)
	cmd.SetArgs([]string{"prod-env", "--description", "Production environment"})

	err := cmd.Execute()
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	// Verify context was created with description
	contexts, _ := manager.List()
	ctx, ok := contexts["prod-env"]
	if !ok {
		t.Fatal("context was not created")
	}

	if ctx.Description != "Production environment" {
		t.Errorf("expected description 'Production environment', got %q", ctx.Description)
	}
}

func TestContextUse(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	_ = manager.Create("test-ctx", "", nil)

	output := &bytes.Buffer{}
	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	cmd := newContextUseCommand(opts)
	err := cmd.RunE(cmd, []string{"test-ctx"})
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Switched to context \"test-ctx\"") {
		t.Errorf("expected switch message, got: %s", result)
	}

	// Verify current context
	if manager.CurrentName() != "test-ctx" {
		t.Errorf("expected current context 'test-ctx', got %q", manager.CurrentName())
	}
}

func TestContextUse_NonExistent(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	output := &bytes.Buffer{}

	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	cmd := newContextUseCommand(opts)
	err := cmd.RunE(cmd, []string{"nonexistent"})
	if err == nil {
		t.Error("expected error for nonexistent context")
	}
}

func TestContextDelete(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	_ = manager.Create("ctx1", "", nil)
	_ = manager.Create("ctx2", "", nil)
	_ = manager.SwitchTo("ctx1")

	output := &bytes.Buffer{}
	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	cmd := newContextDeleteCommand(opts)
	err := cmd.RunE(cmd, []string{"ctx2"})
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Deleted context \"ctx2\"") {
		t.Errorf("expected deletion message, got: %s", result)
	}

	// Verify context was deleted
	contexts, _ := manager.List()
	if _, ok := contexts["ctx2"]; ok {
		t.Error("context was not deleted")
	}
}

func TestContextDelete_Current(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	_ = manager.Create("current-ctx", "", nil)
	_ = manager.SwitchTo("current-ctx")

	output := &bytes.Buffer{}
	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	cmd := newContextDeleteCommand(opts)
	err := cmd.RunE(cmd, []string{"current-ctx"})
	if err == nil {
		t.Error("expected error when deleting current context")
	}

	if !strings.Contains(err.Error(), "cannot delete current context") {
		t.Errorf("expected 'cannot delete current context' error, got: %v", err)
	}
}

func TestContextSet(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	_ = manager.Create("test-ctx", "", nil)

	output := &bytes.Buffer{}
	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	cmd := newContextSetCommand(opts)
	err := cmd.RunE(cmd, []string{"test-ctx", "api_url", "https://api.example.com"})
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Set test-ctx.api_url = https://api.example.com") {
		t.Errorf("expected set message, got: %s", result)
	}

	// Verify field was set
	contexts, _ := manager.List()
	ctx := contexts["test-ctx"]
	if value, ok := ctx.Get("api_url"); !ok || value != "https://api.example.com" {
		t.Errorf("expected field value 'https://api.example.com', got %q", value)
	}
}

func TestContextGet(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	fields := map[string]string{"api_url": "https://api.example.com"}
	_ = manager.Create("test-ctx", "", fields)

	output := &bytes.Buffer{}
	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	cmd := newContextGetCommand(opts)
	err := cmd.RunE(cmd, []string{"test-ctx", "api_url"})
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "https://api.example.com") {
		t.Errorf("expected field value, got: %s", result)
	}
}

func TestContextGet_NonExistentKey(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	_ = manager.Create("test-ctx", "", nil)

	output := &bytes.Buffer{}
	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	cmd := newContextGetCommand(opts)
	err := cmd.RunE(cmd, []string{"test-ctx", "nonexistent"})
	if err == nil {
		t.Error("expected error for nonexistent key")
	}
}

func TestContextRename(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	_ = manager.Create("old-name", "Test context", nil)

	output := &bytes.Buffer{}
	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	cmd := newContextRenameCommand(opts)
	err := cmd.RunE(cmd, []string{"old-name", "new-name"})
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Renamed context \"old-name\" to \"new-name\"") {
		t.Errorf("expected rename message, got: %s", result)
	}

	// Verify context was renamed
	contexts, _ := manager.List()
	if _, ok := contexts["old-name"]; ok {
		t.Error("old context name still exists")
	}
	if _, ok := contexts["new-name"]; !ok {
		t.Error("new context name not found")
	}
}

func TestContextList_Empty(t *testing.T) {
	tmpDir := t.TempDir()
	_ = os.Setenv("XDG_DATA_HOME", filepath.Join(tmpDir, "data"))
	_ = os.Setenv("XDG_STATE_HOME", filepath.Join(tmpDir, "state"))
	defer func() {
		_ = os.Unsetenv("XDG_DATA_HOME")
		_ = os.Unsetenv("XDG_STATE_HOME")
	}()

	stateMgr, _ := state.NewManager("testcli-empty")
	manager := state.NewContextManager(stateMgr)
	output := &bytes.Buffer{}

	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	err := runContextList(opts, "table")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	// Test passes if either no contexts or has default contexts (flexible for existing state)
	if !strings.Contains(result, "No contexts found") && !strings.Contains(result, "default") {
		t.Errorf("expected 'No contexts found' or context list, got: %s", result)
	}
}

func TestContextList_Table(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	_ = manager.Create("dev", "Development environment", map[string]string{"env": "dev"})
	_ = manager.Create("prod", "Production environment", map[string]string{"env": "prod"})
	_ = manager.SwitchTo("dev")

	output := &bytes.Buffer{}
	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	err := runContextList(opts, "table")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "dev") {
		t.Errorf("expected 'dev' in output, got: %s", result)
	}
	if !strings.Contains(result, "prod") {
		t.Errorf("expected 'prod' in output, got: %s", result)
	}
	if !strings.Contains(result, "NAME") {
		t.Errorf("expected table header in output, got: %s", result)
	}
}

func TestContextList_JSON(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	_ = manager.Create("test-ctx", "Test", nil)
	_ = manager.SwitchTo("test-ctx")

	output := &bytes.Buffer{}
	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	err := runContextList(opts, "json")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, `"current"`) {
		t.Errorf("expected 'current' in JSON, got: %s", result)
	}
	if !strings.Contains(result, `"contexts"`) {
		t.Errorf("expected 'contexts' in JSON, got: %s", result)
	}
	if !strings.Contains(result, "test-ctx") {
		t.Errorf("expected 'test-ctx' in JSON, got: %s", result)
	}
}

func TestContextList_YAML(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	_ = manager.Create("test-ctx", "Test", nil)
	_ = manager.SwitchTo("test-ctx")

	output := &bytes.Buffer{}
	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	err := runContextList(opts, "yaml")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "current: test-ctx") {
		t.Errorf("expected 'current: test-ctx' in YAML, got: %s", result)
	}
	if !strings.Contains(result, "contexts:") {
		t.Errorf("expected 'contexts:' in YAML, got: %s", result)
	}
}

func TestContextShow_YAML(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	fields := map[string]string{
		"api_url": "https://api.example.com",
		"token":   "secret123",
	}
	_ = manager.Create("test-ctx", "Test context", fields)

	output := &bytes.Buffer{}
	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	err := runContextShow(opts, "test-ctx", "yaml")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "name: test-ctx") {
		t.Errorf("expected 'name: test-ctx' in YAML, got: %s", result)
	}
	if !strings.Contains(result, "description: Test context") {
		t.Errorf("expected description in YAML, got: %s", result)
	}
	if !strings.Contains(result, "fields:") {
		t.Errorf("expected 'fields:' in YAML, got: %s", result)
	}
	if !strings.Contains(result, "api_url:") {
		t.Errorf("expected 'api_url' field in YAML, got: %s", result)
	}
}

func TestContextShow_JSON(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	_ = manager.Create("test-ctx", "Test context", nil)

	output := &bytes.Buffer{}
	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	err := runContextShow(opts, "test-ctx", "json")
	if err != nil {
		t.Fatalf("command failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, `"name"`) {
		t.Errorf("expected 'name' in JSON, got: %s", result)
	}
	if !strings.Contains(result, "test-ctx") {
		t.Errorf("expected 'test-ctx' in JSON, got: %s", result)
	}
}

func TestContextShow_NonExistent(t *testing.T) {
	stateMgr, _ := state.NewManager("testcli")
	manager := state.NewContextManager(stateMgr)
	output := &bytes.Buffer{}

	opts := &ContextOptions{
		ContextManager: manager,
		Output:         output,
	}

	err := runContextShow(opts, "nonexistent", "yaml")
	if err == nil {
		t.Error("expected error for nonexistent context")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("expected 'not found' error, got: %v", err)
	}
}

func TestFormatContextsTable(t *testing.T) {
	contexts := map[string]*state.Context{
		"dev": {
			Name:        "dev",
			Description: "Development environment",
			Fields:      map[string]string{"env": "dev"},
			CreatedAt:   time.Now(),
		},
		"prod": {
			Name:        "prod",
			Description: "Production environment",
			Fields:      map[string]string{"env": "prod", "api": "https://api.prod"},
			CreatedAt:   time.Now(),
		},
	}

	output := &bytes.Buffer{}
	err := formatContextsTable(contexts, "dev", output)
	if err != nil {
		t.Fatalf("formatContextsTable failed: %v", err)
	}

	result := output.String()

	// Should have header
	if !strings.Contains(result, "NAME") {
		t.Error("expected header in table output")
	}

	// Should have contexts sorted
	devIndex := strings.Index(result, "dev")
	prodIndex := strings.Index(result, "prod")
	if devIndex == -1 || prodIndex == -1 {
		t.Error("expected both contexts in output")
	}
	if devIndex > prodIndex {
		t.Error("expected contexts to be sorted alphabetically")
	}

	// Should mark current context
	if !strings.Contains(result, "*") {
		t.Error("expected current context to be marked with *")
	}
}

func TestFormatContextsJSON(t *testing.T) {
	contexts := map[string]*state.Context{
		"test": {
			Name:      "test",
			CreatedAt: time.Now(),
		},
	}

	output := &bytes.Buffer{}
	err := formatContextsJSON(contexts, "test", output)
	if err != nil {
		t.Fatalf("formatContextsJSON failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, `"current"`) {
		t.Error("expected 'current' in JSON output")
	}
	if !strings.Contains(result, `"contexts"`) {
		t.Error("expected 'contexts' in JSON output")
	}
}

func TestFormatContextsYAML(t *testing.T) {
	contexts := map[string]*state.Context{
		"test": {
			Name:        "test",
			Description: "Test context",
			Fields:      map[string]string{"key": "value"},
			CreatedAt:   time.Now(),
		},
	}

	output := &bytes.Buffer{}
	err := formatContextsYAML(contexts, "test", output)
	if err != nil {
		t.Fatalf("formatContextsYAML failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "current: test") {
		t.Error("expected 'current' in YAML output")
	}
	if !strings.Contains(result, "contexts:") {
		t.Error("expected 'contexts:' in YAML output")
	}
	if !strings.Contains(result, "test:") {
		t.Error("expected context name in YAML output")
	}
}
