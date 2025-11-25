// Package state manages CLI state, contexts, history, and recent values.
//
// The state package provides persistent storage for CLI runtime state including
// named contexts (like kubectl contexts), command history, recently used values,
// resource preferences, and session data. All state is stored in XDG-compliant
// directories and uses YAML format for human readability.
//
// # State Components
//
//   - Contexts: Named configuration sets (API endpoints, auth, preferences)
//   - Recent Values: Recently used IDs/names for autocomplete suggestions
//   - Resource Preferences: Per-resource metadata (last used, favorites, counts)
//   - Session Data: Current session info (last command, working directory)
//   - History: Command execution history with timestamps
//
// # Example Usage
//
//	// Create state manager
//	mgr, _ := state.NewManager("mycli")
//
//	// Work with contexts
//	mgr.CreateContext("prod", &state.Context{
//	    APIEndpoint: "https://api.prod.example.com",
//	})
//	mgr.SetCurrentContext("prod")
//
//	// Track recent values for autocomplete
//	mgr.AddRecentValue("cluster-ids", "cluster-abc-123")
//	suggestions := mgr.GetRecentValues("cluster-ids")
//
//	// Mark resources as used
//	mgr.MarkResourceUsed("clusters", "cluster-abc-123")
//
//	// Save state
//	mgr.Save()
//
// # Context Management
//
// Contexts allow users to switch between different API environments:
//
//	mycli context create staging --endpoint https://staging.api.example.com
//	mycli context use staging
//	mycli context list
//
// # State Location
//
//   - Linux: ~/.local/state/mycli/state.yaml
//   - macOS: ~/Library/Application Support/mycli/state.yaml
//   - Windows: %LOCALAPPDATA%\mycli\state\state.yaml
//
// The Manager is thread-safe and uses atomic file writes to prevent
// corruption during concurrent access.
package state

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"
)

// Manager handles loading, saving, and managing CLI state.
type Manager struct {
	cliName   string
	statePath string
	state     *State
	mu        sync.RWMutex
}

// State represents the complete CLI state.
type State struct {
	// Current context
	CurrentContext string `yaml:"current_context,omitempty" json:"current_context,omitempty"`

	// Named contexts
	Contexts map[string]*Context `yaml:"contexts,omitempty" json:"contexts,omitempty"`

	// Recent items for autocomplete
	Recent *Recent `yaml:"recent,omitempty" json:"recent,omitempty"`

	// Per-resource preferences
	Preferences map[string]map[string]*ResourcePreference `yaml:"preferences,omitempty" json:"preferences,omitempty"`

	// Session data
	Session *Session `yaml:"session,omitempty" json:"session,omitempty"`

	// Last modified timestamp
	LastModified time.Time `yaml:"last_modified,omitempty" json:"last_modified,omitempty"`
}

// ResourcePreference represents preferences for a specific resource instance.
type ResourcePreference struct {
	LastUsed  time.Time         `yaml:"last_used" json:"last_used"`
	Favorite  bool              `yaml:"favorite,omitempty" json:"favorite,omitempty"`
	Metadata  map[string]string `yaml:"metadata,omitempty" json:"metadata,omitempty"`
	UseCount  int               `yaml:"use_count,omitempty" json:"use_count,omitempty"`
}

// Session represents current session data.
type Session struct {
	LastCommand     string    `yaml:"last_command,omitempty" json:"last_command,omitempty"`
	LastCommandTime time.Time `yaml:"last_command_time,omitempty" json:"last_command_time,omitempty"`
	WorkingDir      string    `yaml:"working_dir,omitempty" json:"working_dir,omitempty"`
}

// NewManager creates a new state manager.
func NewManager(cliName string) (*Manager, error) {
	statePath, err := getStatePath(cliName)
	if err != nil {
		return nil, fmt.Errorf("failed to get state path: %w", err)
	}

	m := &Manager{
		cliName:   cliName,
		statePath: statePath,
		state:     newDefaultState(),
	}

	// Try to load existing state
	if err := m.Load(); err != nil {
		// If file doesn't exist, that's okay - we'll create it on first save
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("failed to load state: %w", err)
		}
	}

	return m, nil
}

// newDefaultState creates a new default state.
func newDefaultState() *State {
	return &State{
		CurrentContext: "default",
		Contexts: map[string]*Context{
			"default": NewContext("default"),
		},
		Recent: NewRecent(),
		Preferences: make(map[string]map[string]*ResourcePreference),
		Session: &Session{
			WorkingDir: ".",
		},
		LastModified: time.Now(),
	}
}

// Load loads state from disk.
func (m *Manager) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	data, err := os.ReadFile(m.statePath)
	if err != nil {
		return err
	}

	var state State
	if err := yaml.Unmarshal(data, &state); err != nil {
		return fmt.Errorf("failed to parse state file: %w", err)
	}

	// Ensure we have default structures
	if state.Contexts == nil {
		state.Contexts = make(map[string]*Context)
	}
	if state.Recent == nil {
		state.Recent = NewRecent()
	} else {
		// Ensure Recent has initialized maps even if loaded from YAML
		if state.Recent.Lists == nil {
			state.Recent.Lists = make(map[string]*RecentList)
		}
		if state.Recent.MaxPerList == 0 {
			state.Recent.MaxPerList = DefaultMaxRecentEntries
		}
		// Ensure all loaded lists have initialized entries slices
		for _, list := range state.Recent.Lists {
			if list.Entries == nil {
				list.Entries = make([]*RecentItem, 0)
			}
			if list.Max == 0 {
				list.Max = DefaultMaxRecentEntries
			}
		}
	}
	if state.Preferences == nil {
		state.Preferences = make(map[string]map[string]*ResourcePreference)
	}
	if state.Session == nil {
		state.Session = &Session{WorkingDir: "."}
	}

	// Ensure default context exists
	if state.CurrentContext == "" {
		state.CurrentContext = "default"
	}
	if _, exists := state.Contexts[state.CurrentContext]; !exists {
		state.Contexts[state.CurrentContext] = NewContext(state.CurrentContext)
	}

	m.state = &state
	return nil
}

// Save saves state to disk.
func (m *Manager) Save() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(m.statePath), 0700); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}

	// Update last modified time
	m.state.LastModified = time.Now()

	// Marshal to YAML
	data, err := yaml.Marshal(m.state)
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file with atomic rename
	tmpPath := m.statePath + ".tmp"
	if err := os.WriteFile(tmpPath, data, 0600); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	if err := os.Rename(tmpPath, m.statePath); err != nil {
		os.Remove(tmpPath) // Clean up on error
		return fmt.Errorf("failed to save state file: %w", err)
	}

	return nil
}

// GetCurrentContext returns the current context.
func (m *Manager) GetCurrentContext() *Context {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx, exists := m.state.Contexts[m.state.CurrentContext]
	if !exists {
		// Return default context
		return NewContext("default")
	}
	return ctx
}

// SetCurrentContext sets the current context by name.
func (m *Manager) SetCurrentContext(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.state.Contexts[name]; !exists {
		return fmt.Errorf("context %q does not exist", name)
	}

	m.state.CurrentContext = name
	return nil
}

// CreateContext creates a new named context.
func (m *Manager) CreateContext(name string, ctx *Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.state.Contexts[name]; exists {
		return fmt.Errorf("context %q already exists", name)
	}

	ctx.Name = name
	m.state.Contexts[name] = ctx
	return nil
}

// UpdateContext updates an existing context.
func (m *Manager) UpdateContext(name string, ctx *Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.state.Contexts[name]; !exists {
		return fmt.Errorf("context %q does not exist", name)
	}

	ctx.Name = name
	m.state.Contexts[name] = ctx
	return nil
}

// DeleteContext deletes a context.
func (m *Manager) DeleteContext(name string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if name == "default" {
		return fmt.Errorf("cannot delete default context")
	}

	if _, exists := m.state.Contexts[name]; !exists {
		return fmt.Errorf("context %q does not exist", name)
	}

	delete(m.state.Contexts, name)

	// If we deleted the current context, switch to default
	if m.state.CurrentContext == name {
		m.state.CurrentContext = "default"
	}

	return nil
}

// ListContexts returns all context names.
func (m *Manager) ListContexts() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()

	names := make([]string, 0, len(m.state.Contexts))
	for name := range m.state.Contexts {
		names = append(names, name)
	}
	return names
}

// GetContext returns a specific context by name.
func (m *Manager) GetContext(name string) (*Context, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	ctx, exists := m.state.Contexts[name]
	if !exists {
		return nil, fmt.Errorf("context %q does not exist", name)
	}
	return ctx, nil
}

// GetRecent returns the recent values manager.
func (m *Manager) GetRecent() *Recent {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state.Recent
}

// AddRecentValue adds a value to a recent list.
func (m *Manager) AddRecentValue(listName, value string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.state.Recent.Add(listName, value)
}

// GetRecentValues returns recent values for a list.
func (m *Manager) GetRecentValues(listName string) []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state.Recent.Get(listName)
}

// SetPreference sets a preference for a resource instance.
func (m *Manager) SetPreference(resourceType, resourceID string, pref *ResourcePreference) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state.Preferences[resourceType] == nil {
		m.state.Preferences[resourceType] = make(map[string]*ResourcePreference)
	}

	pref.LastUsed = time.Now()
	pref.UseCount++
	m.state.Preferences[resourceType][resourceID] = pref
}

// GetPreference gets a preference for a resource instance.
func (m *Manager) GetPreference(resourceType, resourceID string) (*ResourcePreference, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.state.Preferences[resourceType] == nil {
		return nil, false
	}

	pref, exists := m.state.Preferences[resourceType][resourceID]
	return pref, exists
}

// MarkResourceUsed marks a resource as used (updates last used time and count).
func (m *Manager) MarkResourceUsed(resourceType, resourceID string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.state.Preferences[resourceType] == nil {
		m.state.Preferences[resourceType] = make(map[string]*ResourcePreference)
	}

	pref, exists := m.state.Preferences[resourceType][resourceID]
	if !exists {
		pref = &ResourcePreference{
			Metadata: make(map[string]string),
		}
		m.state.Preferences[resourceType][resourceID] = pref
	}

	pref.LastUsed = time.Now()
	pref.UseCount++
}

// SetSessionCommand sets the last command in the session.
func (m *Manager) SetSessionCommand(command string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state.Session.LastCommand = command
	m.state.Session.LastCommandTime = time.Now()
}

// GetSession returns the current session.
func (m *Manager) GetSession() *Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.state.Session
}

// getStatePath returns the path to the state file.
func getStatePath(cliName string) (string, error) {
	// Use XDG state directory
	stateDir := filepath.Join(xdg.StateHome, cliName)
	return filepath.Join(stateDir, "state.yaml"), nil
}

// GetStatePath returns the path to the state file (for external use).
func (m *Manager) GetStatePath() string {
	return m.statePath
}

// Reset resets the state to defaults.
func (m *Manager) Reset() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.state = newDefaultState()
	return nil
}

// GetState returns a copy of the current state (for inspection/debugging).
func (m *Manager) GetState() *State {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Return a shallow copy
	stateCopy := *m.state
	return &stateCopy
}
