package workflow

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

// StateManager manages workflow execution state persistence.
type StateManager struct {
	stateDir string
}

// NewStateManager creates a new state manager.
func NewStateManager() *StateManager {
	// Use XDG-compliant state directory
	stateDir := filepath.Join(xdg.StateHome, "cliforge", "workflows")
	return &StateManager{
		stateDir: stateDir,
	}
}

// NewStateManagerWithDir creates a new state manager with a custom directory.
func NewStateManagerWithDir(dir string) *StateManager {
	return &StateManager{
		stateDir: dir,
	}
}

// ensureStateDir ensures the state directory exists.
func (sm *StateManager) ensureStateDir() error {
	if err := os.MkdirAll(sm.stateDir, 0755); err != nil {
		return fmt.Errorf("failed to create state directory: %w", err)
	}
	return nil
}

// SaveState saves workflow execution state to disk.
func (sm *StateManager) SaveState(state *ExecutionState) error {
	if err := sm.ensureStateDir(); err != nil {
		return err
	}

	// Marshal state to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal state: %w", err)
	}

	// Write to file
	filename := filepath.Join(sm.stateDir, fmt.Sprintf("%s.json", state.WorkflowID))
	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("failed to write state file: %w", err)
	}

	return nil
}

// LoadState loads workflow execution state from disk.
func (sm *StateManager) LoadState(workflowID string) (*ExecutionState, error) {
	filename := filepath.Join(sm.stateDir, fmt.Sprintf("%s.json", workflowID))

	// Read file
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read state file: %w", err)
	}

	// Unmarshal state
	var state ExecutionState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to unmarshal state: %w", err)
	}

	return &state, nil
}

// DeleteState deletes a saved workflow state.
func (sm *StateManager) DeleteState(workflowID string) error {
	filename := filepath.Join(sm.stateDir, fmt.Sprintf("%s.json", workflowID))

	if err := os.Remove(filename); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete state file: %w", err)
	}

	return nil
}

// ListStates lists all saved workflow states.
func (sm *StateManager) ListStates() ([]*ExecutionState, error) {
	if err := sm.ensureStateDir(); err != nil {
		return nil, err
	}

	// Read directory
	entries, err := os.ReadDir(sm.stateDir)
	if err != nil {
		return nil, fmt.Errorf("failed to read state directory: %w", err)
	}

	states := make([]*ExecutionState, 0)

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		// Only process .json files
		if filepath.Ext(entry.Name()) != ".json" {
			continue
		}

		// Extract workflow ID from filename
		workflowID := entry.Name()[:len(entry.Name())-5] // Remove .json extension

		// Load state
		state, err := sm.LoadState(workflowID)
		if err != nil {
			// Skip files that can't be loaded
			continue
		}

		states = append(states, state)
	}

	return states, nil
}

// CleanupOldStates removes states older than the specified duration.
func (sm *StateManager) CleanupOldStates(maxAge int) error {
	states, err := sm.ListStates()
	if err != nil {
		return err
	}

	// Current time
	now := os.Getenv("NOW") // For testing
	var currentTime int64
	if now != "" {
		fmt.Sscanf(now, "%d", &currentTime)
	} else {
		currentTime = currentTime
	}

	for _, state := range states {
		age := currentTime - state.StartTime.Unix()
		if age > int64(maxAge) {
			if err := sm.DeleteState(state.WorkflowID); err != nil {
				// Log error but continue
				fmt.Printf("Warning: failed to delete old state %s: %v\n", state.WorkflowID, err)
			}
		}
	}

	return nil
}

// GetStateFilePath returns the path to a state file.
func (sm *StateManager) GetStateFilePath(workflowID string) string {
	return filepath.Join(sm.stateDir, fmt.Sprintf("%s.json", workflowID))
}

// StateExists checks if a state file exists.
func (sm *StateManager) StateExists(workflowID string) bool {
	filename := sm.GetStateFilePath(workflowID)
	_, err := os.Stat(filename)
	return err == nil
}
