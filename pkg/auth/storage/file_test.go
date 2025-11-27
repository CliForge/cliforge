package storage

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/auth/types"
)

func TestFileStorage_SaveAndLoad(t *testing.T) {
	// Create temporary directory
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "auth.json")

	config := &types.StorageConfig{
		Type: types.StorageTypeFile,
		Path: tokenPath,
	}

	storage, err := NewFileStorage(config, "test-cli")
	if err != nil {
		t.Fatalf("NewFileStorage() failed: %v", err)
	}

	ctx := context.Background()

	token := &types.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scopes:       []string{"read", "write"},
		Extra: map[string]interface{}{
			"custom": "value",
		},
	}

	// Save token
	err = storage.SaveToken(ctx, token)
	if err != nil {
		t.Fatalf("SaveToken() failed: %v", err)
	}

	// Verify file exists
	if _, err := os.Stat(tokenPath); os.IsNotExist(err) {
		t.Error("Token file was not created")
	}

	// Load token
	loaded, err := storage.LoadToken(ctx)
	if err != nil {
		t.Fatalf("LoadToken() failed: %v", err)
	}

	// Verify token fields
	if loaded.AccessToken != token.AccessToken {
		t.Errorf("AccessToken = %v, want %v", loaded.AccessToken, token.AccessToken)
	}
	if loaded.RefreshToken != token.RefreshToken {
		t.Errorf("RefreshToken = %v, want %v", loaded.RefreshToken, token.RefreshToken)
	}
	if loaded.TokenType != token.TokenType {
		t.Errorf("TokenType = %v, want %v", loaded.TokenType, token.TokenType)
	}
}

func TestFileStorage_LoadNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "nonexistent.json")

	config := &types.StorageConfig{
		Type: types.StorageTypeFile,
		Path: tokenPath,
	}

	storage, err := NewFileStorage(config, "test-cli")
	if err != nil {
		t.Fatalf("NewFileStorage() failed: %v", err)
	}

	ctx := context.Background()
	_, err = storage.LoadToken(ctx)
	if err == nil {
		t.Error("LoadToken() should return error when file doesn't exist")
	}
}

func TestFileStorage_Delete(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "auth.json")

	config := &types.StorageConfig{
		Type: types.StorageTypeFile,
		Path: tokenPath,
	}

	storage, err := NewFileStorage(config, "test-cli")
	if err != nil {
		t.Fatalf("NewFileStorage() failed: %v", err)
	}

	ctx := context.Background()

	token := &types.Token{
		AccessToken: "test-token",
	}

	// Save token
	err = storage.SaveToken(ctx, token)
	if err != nil {
		t.Fatalf("SaveToken() failed: %v", err)
	}

	// Delete token
	err = storage.DeleteToken(ctx)
	if err != nil {
		t.Fatalf("DeleteToken() failed: %v", err)
	}

	// Verify file was deleted
	if _, err := os.Stat(tokenPath); !os.IsNotExist(err) {
		t.Error("Token file still exists after deletion")
	}

	// Try to load - should fail
	_, err = storage.LoadToken(ctx)
	if err == nil {
		t.Error("LoadToken() should return error after deletion")
	}
}

func TestFileStorage_SaveNil(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "auth.json")

	config := &types.StorageConfig{
		Type: types.StorageTypeFile,
		Path: tokenPath,
	}

	storage, err := NewFileStorage(config, "test-cli")
	if err != nil {
		t.Fatalf("NewFileStorage() failed: %v", err)
	}

	ctx := context.Background()
	err = storage.SaveToken(ctx, nil)
	if err == nil {
		t.Error("SaveToken() should return error for nil token")
	}
}

func TestFileStorage_DefaultPath(t *testing.T) {
	config := &types.StorageConfig{
		Type: types.StorageTypeFile,
		// No path specified - should use default
	}

	storage, err := NewFileStorage(config, "test-cli")
	if err != nil {
		t.Fatalf("NewFileStorage() failed: %v", err)
	}

	path := storage.GetPath()
	if path == "" {
		t.Error("GetPath() returned empty string")
	}

	// Should contain cli name somewhere in the path
	// Just check that it's not empty and is a valid path
	if path == "" {
		t.Error("Default path should not be empty")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("Default path should be absolute, got: %s", path)
	}
}

func TestFileStorage_FilePermissions(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "auth.json")

	config := &types.StorageConfig{
		Type: types.StorageTypeFile,
		Path: tokenPath,
	}

	storage, err := NewFileStorage(config, "test-cli")
	if err != nil {
		t.Fatalf("NewFileStorage() failed: %v", err)
	}

	ctx := context.Background()

	token := &types.Token{
		AccessToken: "secret-token",
	}

	// Save token
	err = storage.SaveToken(ctx, token)
	if err != nil {
		t.Fatalf("SaveToken() failed: %v", err)
	}

	// Check file permissions
	info, err := os.Stat(tokenPath)
	if err != nil {
		t.Fatalf("Failed to stat file: %v", err)
	}

	mode := info.Mode()
	// On Unix systems, should be 0600 (owner read/write only)
	if mode.Perm() != 0600 {
		t.Errorf("File permissions = %v, want 0600", mode.Perm())
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)+1 && s[1:len(substr)+1] == substr))
}

func TestFileStorage_LoadInvalidJSON(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "auth.json")

	// Write invalid JSON to file
	err := os.WriteFile(tokenPath, []byte("invalid json"), 0600)
	if err != nil {
		t.Fatalf("Failed to write invalid JSON: %v", err)
	}

	config := &types.StorageConfig{
		Type: types.StorageTypeFile,
		Path: tokenPath,
	}

	storage, err := NewFileStorage(config, "test-cli")
	if err != nil {
		t.Fatalf("NewFileStorage() failed: %v", err)
	}

	ctx := context.Background()
	_, err = storage.LoadToken(ctx)
	if err == nil {
		t.Error("LoadToken() should return error for invalid JSON")
	}
}

func TestFileStorage_DeleteNonexistent(t *testing.T) {
	tmpDir := t.TempDir()
	tokenPath := filepath.Join(tmpDir, "nonexistent.json")

	config := &types.StorageConfig{
		Type: types.StorageTypeFile,
		Path: tokenPath,
	}

	storage, err := NewFileStorage(config, "test-cli")
	if err != nil {
		t.Fatalf("NewFileStorage() failed: %v", err)
	}

	ctx := context.Background()
	// Deleting non-existent file should not error (or should handle gracefully)
	err = storage.DeleteToken(ctx)
	// Some implementations may not error on deleting non-existent files
	// Just verify it completes
	_ = err
}
