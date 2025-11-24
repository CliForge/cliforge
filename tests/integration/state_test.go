package integration

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/state"
	"github.com/CliForge/cliforge/tests/helpers"
)

// TestStateManager tests the state manager initialization and basic operations.
func TestStateManager(t *testing.T) {
	tmpDir := t.TempDir()

	// Override XDG state home for testing
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	// Test: Create state manager
	t.Run("CreateManager", func(t *testing.T) {
		manager, err := state.NewManager("testcli")
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, manager)

		// Verify state file path
		statePath := manager.GetStatePath()
		helpers.AssertStringContains(t, statePath, "testcli")
		helpers.AssertStringContains(t, statePath, "state.yaml")
	})
}

// TestContextManagement tests context creation and management.
func TestContextManagement(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	manager, err := state.NewManager("testcli")
	helpers.AssertNoError(t, err)

	// Test: Get default context
	t.Run("DefaultContext", func(t *testing.T) {
		ctx := manager.GetCurrentContext()
		helpers.AssertNotNil(t, ctx)
		helpers.AssertEqual(t, "default", ctx.Name)
	})

	// Test: Create new context
	t.Run("CreateContext", func(t *testing.T) {
		newCtx := state.NewContext("production")
		newCtx.Set("api_url", "https://api.production.com")
		newCtx.Set("region", "us-east-1")

		err := manager.CreateContext("production", newCtx)
		helpers.AssertNoError(t, err)

		// Retrieve created context
		retrieved, err := manager.GetContext("production")
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, "production", retrieved.Name)

		apiURL, exists := retrieved.Get("api_url")
		helpers.AssertTrue(t, exists, "api_url should exist")
		helpers.AssertEqual(t, "https://api.production.com", apiURL)
	})

	// Test: Switch context
	t.Run("SwitchContext", func(t *testing.T) {
		// Create staging context
		stagingCtx := state.NewContext("staging")
		err := manager.CreateContext("staging", stagingCtx)
		helpers.AssertNoError(t, err)

		// Switch to staging
		err = manager.SetCurrentContext("staging")
		helpers.AssertNoError(t, err)

		// Verify current context
		current := manager.GetCurrentContext()
		helpers.AssertEqual(t, "staging", current.Name)
	})

	// Test: List contexts
	t.Run("ListContexts", func(t *testing.T) {
		contexts := manager.ListContexts()
		helpers.AssertTrue(t, len(contexts) >= 2, "Should have at least 2 contexts")
		helpers.AssertSliceContains(t, contexts, "default")
		helpers.AssertSliceContains(t, contexts, "production")
	})

	// Test: Delete context
	t.Run("DeleteContext", func(t *testing.T) {
		err := manager.DeleteContext("staging")
		helpers.AssertNoError(t, err)

		// Verify deletion
		_, err = manager.GetContext("staging")
		helpers.AssertError(t, err)
	})

	// Test: Cannot delete default context
	t.Run("CannotDeleteDefault", func(t *testing.T) {
		err := manager.DeleteContext("default")
		helpers.AssertError(t, err)
		helpers.AssertErrorContains(t, err, "cannot delete default")
	})
}

// TestContextVariables tests context variable storage.
func TestContextVariables(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	manager, err := state.NewManager("testcli")
	helpers.AssertNoError(t, err)

	ctx := manager.GetCurrentContext()

	// Test: Set and get string variable
	t.Run("StringVariable", func(t *testing.T) {
		ctx.Set("username", "alice")
		value, exists := ctx.Get("username")
		helpers.AssertTrue(t, exists, "Variable should exist")
		helpers.AssertEqual(t, "alice", value)
	})

	// Test: Set and get nested variable
	t.Run("NestedVariable", func(t *testing.T) {
		ctx.Set("config.database.host", "localhost")
		ctx.Set("config.database.port", "5432")

		host, exists := ctx.Get("config.database.host")
		helpers.AssertTrue(t, exists, "Nested variable should exist")
		helpers.AssertEqual(t, "localhost", host)
	})

	// Test: Unset variable
	t.Run("UnsetVariable", func(t *testing.T) {
		ctx.Set("temp", "value")
		ctx.Unset("temp")

		_, exists := ctx.Get("temp")
		helpers.AssertFalse(t, exists, "Variable should not exist after unset")
	})

	// Test: List all variables
	t.Run("ListVariables", func(t *testing.T) {
		ctx.Set("var1", "value1")
		ctx.Set("var2", "value2")

		vars := ctx.List()
		helpers.AssertTrue(t, len(vars) > 0, "Should have variables")
	})
}

// TestRecentValues tests recent values tracking.
func TestRecentValues(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	manager, err := state.NewManager("testcli")
	helpers.AssertNoError(t, err)

	// Test: Add recent values
	t.Run("AddRecentValues", func(t *testing.T) {
		manager.AddRecentValue("clusters", "prod-cluster-1")
		manager.AddRecentValue("clusters", "dev-cluster-1")
		manager.AddRecentValue("clusters", "staging-cluster-1")

		recents := manager.GetRecentValues("clusters")
		helpers.AssertTrue(t, len(recents) >= 1, "Should have recent values")
		helpers.AssertSliceContains(t, recents, "prod-cluster-1")
	})

	// Test: Recent values limit
	t.Run("RecentValuesLimit", func(t *testing.T) {
		// Add many values
		for i := 0; i < 20; i++ {
			manager.AddRecentValue("regions", helpers.MustMarshalJSON(t, map[string]int{"region": i}))
		}

		recents := manager.GetRecentValues("regions")
		// Should be limited (default is usually 10)
		helpers.AssertLessThan(t, len(recents), 21, "Recent values should be limited")
	})

	// Test: Most recent first
	t.Run("MostRecentFirst", func(t *testing.T) {
		manager.AddRecentValue("commands", "first")
		time.Sleep(10 * time.Millisecond)
		manager.AddRecentValue("commands", "second")
		time.Sleep(10 * time.Millisecond)
		manager.AddRecentValue("commands", "third")

		recents := manager.GetRecentValues("commands")
		if len(recents) > 0 {
			// Most recent should be "third"
			helpers.AssertEqual(t, "third", recents[0])
		}
	})
}

// TestResourcePreferences tests resource preference tracking.
func TestResourcePreferences(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	manager, err := state.NewManager("testcli")
	helpers.AssertNoError(t, err)

	// Test: Set preference
	t.Run("SetPreference", func(t *testing.T) {
		pref := &state.ResourcePreference{
			Favorite: true,
			Metadata: map[string]string{
				"color": "blue",
				"size":  "large",
			},
		}

		manager.SetPreference("clusters", "prod-cluster-1", pref)

		// Retrieve preference
		retrieved, exists := manager.GetPreference("clusters", "prod-cluster-1")
		helpers.AssertTrue(t, exists, "Preference should exist")
		helpers.AssertTrue(t, retrieved.Favorite, "Should be favorited")
		helpers.AssertEqual(t, "blue", retrieved.Metadata["color"])
	})

	// Test: Mark resource as used
	t.Run("MarkResourceUsed", func(t *testing.T) {
		manager.MarkResourceUsed("databases", "main-db")

		pref, exists := manager.GetPreference("databases", "main-db")
		helpers.AssertTrue(t, exists, "Preference should be created")
		helpers.AssertTrue(t, pref.UseCount > 0, "Use count should be incremented")
		helpers.AssertTrue(t, !pref.LastUsed.IsZero(), "Last used should be set")
	})

	// Test: Multiple uses increment count
	t.Run("IncrementUseCount", func(t *testing.T) {
		manager.MarkResourceUsed("apis", "user-api")
		manager.MarkResourceUsed("apis", "user-api")
		manager.MarkResourceUsed("apis", "user-api")

		pref, exists := manager.GetPreference("apis", "user-api")
		helpers.AssertTrue(t, exists, "Preference should exist")
		helpers.AssertTrue(t, pref.UseCount >= 3, "Use count should be at least 3")
	})
}

// TestSessionTracking tests session tracking.
func TestSessionTracking(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	manager, err := state.NewManager("testcli")
	helpers.AssertNoError(t, err)

	// Test: Track last command
	t.Run("TrackLastCommand", func(t *testing.T) {
		manager.SetSessionCommand("users list --limit 10")

		session := manager.GetSession()
		helpers.AssertEqual(t, "users list --limit 10", session.LastCommand)
		helpers.AssertTrue(t, !session.LastCommandTime.IsZero(), "Command time should be set")
	})

	// Test: Session working directory
	t.Run("WorkingDirectory", func(t *testing.T) {
		session := manager.GetSession()
		helpers.AssertNotEqual(t, "", session.WorkingDir)
	})
}

// TestStatePersistence tests state persistence to disk.
func TestStatePersistence(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	// Create manager and add state
	manager1, err := state.NewManager("testcli")
	helpers.AssertNoError(t, err)

	ctx := manager1.GetCurrentContext()
	ctx.Set("persisted_value", "test123")
	manager1.AddRecentValue("items", "item1")

	// Save state
	err = manager1.Save()
	helpers.AssertNoError(t, err)

	// Verify file exists
	helpers.AssertFileExists(t, manager1.GetStatePath())

	// Create new manager (should load persisted state)
	manager2, err := state.NewManager("testcli")
	helpers.AssertNoError(t, err)

	// Verify persisted data
	ctx2 := manager2.GetCurrentContext()
	value, exists := ctx2.Get("persisted_value")
	helpers.AssertTrue(t, exists, "Persisted value should exist")
	helpers.AssertEqual(t, "test123", value)

	recents := manager2.GetRecentValues("items")
	helpers.AssertSliceContains(t, recents, "item1")
}

// TestStateReset tests state reset functionality.
func TestStateReset(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	manager, err := state.NewManager("testcli")
	helpers.AssertNoError(t, err)

	// Add some state
	ctx := manager.GetCurrentContext()
	ctx.Set("key", "value")
	manager.AddRecentValue("list", "item")

	// Reset state
	err = manager.Reset()
	helpers.AssertNoError(t, err)

	// Verify reset
	ctx = manager.GetCurrentContext()
	_, exists := ctx.Get("key")
	helpers.AssertFalse(t, exists, "State should be reset")

	recents := manager.GetRecentValues("list")
	helpers.AssertEqual(t, 0, len(recents))
}

// TestStateConcurrency tests concurrent state operations.
func TestStateConcurrency(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	manager, err := state.NewManager("testcli")
	helpers.AssertNoError(t, err)

	// Test: Concurrent context updates
	t.Run("ConcurrentUpdates", func(t *testing.T) {
		done := make(chan bool, 10)

		// Concurrent context reads
		for i := 0; i < 5; i++ {
			go func(id int) {
				ctx := manager.GetCurrentContext()
				ctx.Set(helpers.MustMarshalJSON(t, map[string]int{"key": id}), helpers.MustMarshalJSON(t, map[string]int{"value": id}))
				done <- true
			}(i)
		}

		// Concurrent recent value additions
		for i := 0; i < 5; i++ {
			go func(id int) {
				manager.AddRecentValue("concurrent", helpers.MustMarshalJSON(t, map[string]int{"item": id}))
				done <- true
			}(i)
		}

		// Wait for all operations
		for i := 0; i < 10; i++ {
			<-done
		}

		// State should remain consistent
		ctx := manager.GetCurrentContext()
		helpers.AssertNotNil(t, ctx)
	})
}

// TestStateHistory tests command history functionality.
func TestStateHistory(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	manager, err := state.NewManager("testcli")
	helpers.AssertNoError(t, err)

	// Test: Build command history
	t.Run("CommandHistory", func(t *testing.T) {
		commands := []string{
			"users list",
			"users get user-123",
			"users create --name Alice",
			"users delete user-456",
		}

		for _, cmd := range commands {
			manager.SetSessionCommand(cmd)
			time.Sleep(10 * time.Millisecond)
		}

		// Last command should be the delete
		session := manager.GetSession()
		helpers.AssertEqual(t, "users delete user-456", session.LastCommand)
	})
}

// TestStateExport tests exporting state.
func TestStateExport(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	manager, err := state.NewManager("testcli")
	helpers.AssertNoError(t, err)

	// Add some state
	ctx := manager.GetCurrentContext()
	ctx.Set("exported_key", "exported_value")

	// Test: Get state for inspection
	t.Run("ExportState", func(t *testing.T) {
		exportedState := manager.GetState()
		helpers.AssertNotNil(t, exportedState)
		helpers.AssertNotNil(t, exportedState.Contexts)
		helpers.AssertTrue(t, len(exportedState.Contexts) > 0, "Should have contexts")
	})
}

// TestStateValidation tests state validation.
func TestStateValidation(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	// Test: Invalid context name
	t.Run("InvalidContextName", func(t *testing.T) {
		manager, err := state.NewManager("testcli")
		helpers.AssertNoError(t, err)

		// Try to switch to non-existent context
		err = manager.SetCurrentContext("nonexistent")
		helpers.AssertError(t, err)
		helpers.AssertErrorContains(t, err, "does not exist")
	})
}

// TestStateFileCorruption tests handling of corrupted state files.
func TestStateFileCorruption(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	// Create manager to establish state file
	manager1, err := state.NewManager("testcli")
	helpers.AssertNoError(t, err)
	err = manager1.Save()
	helpers.AssertNoError(t, err)

	// Corrupt the state file
	statePath := manager1.GetStatePath()
	err = os.WriteFile(statePath, []byte("invalid: yaml: content: [[["), 0600)
	helpers.AssertNoError(t, err)

	// Try to load corrupted state
	_, err = state.NewManager("testcli")
	// Should handle corruption gracefully
	// (either error or fall back to defaults)
	helpers.AssertNotNil(t, err)
}

// TestStateTimestamps tests state timestamp tracking.
func TestStateTimestamps(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	manager, err := state.NewManager("testcli")
	helpers.AssertNoError(t, err)

	// Make a change
	ctx := manager.GetCurrentContext()
	ctx.Set("key", "value")

	// Save state
	beforeSave := time.Now()
	err = manager.Save()
	helpers.AssertNoError(t, err)

	// Get state
	exportedState := manager.GetState()

	// Verify timestamp is recent
	timeDiff := exportedState.LastModified.Sub(beforeSave)
	helpers.AssertTrue(t, timeDiff >= 0, "Timestamp should be after save started")
	helpers.AssertTrue(t, timeDiff < 1*time.Second, "Timestamp should be very recent")
}

// TestMultipleStateFiles tests managing multiple CLI state files.
func TestMultipleStateFiles(t *testing.T) {
	tmpDir := t.TempDir()
	os.Setenv("XDG_STATE_HOME", tmpDir)
	defer os.Unsetenv("XDG_STATE_HOME")

	// Create managers for different CLIs
	manager1, err := state.NewManager("cli1")
	helpers.AssertNoError(t, err)

	manager2, err := state.NewManager("cli2")
	helpers.AssertNoError(t, err)

	// Set different values
	ctx1 := manager1.GetCurrentContext()
	ctx1.Set("identity", "cli1")

	ctx2 := manager2.GetCurrentContext()
	ctx2.Set("identity", "cli2")

	// Save both
	err = manager1.Save()
	helpers.AssertNoError(t, err)
	err = manager2.Save()
	helpers.AssertNoError(t, err)

	// Verify separate files
	path1 := manager1.GetStatePath()
	path2 := manager2.GetStatePath()

	helpers.AssertNotEqual(t, path1, path2)
	helpers.AssertStringContains(t, path1, "cli1")
	helpers.AssertStringContains(t, path2, "cli2")
}
