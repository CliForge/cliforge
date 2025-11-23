package storage

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/CliForge/cliforge/pkg/auth/types"
	"github.com/adrg/xdg"
)

// FileStorage implements file-based token storage.
type FileStorage struct {
	path string
}

// NewFileStorage creates a new file-based storage.
func NewFileStorage(config *types.StorageConfig, cliName string) (*FileStorage, error) {
	path := config.Path
	if path == "" {
		// Use XDG-compliant default path
		configDir := filepath.Join(xdg.ConfigHome, cliName)
		path = filepath.Join(configDir, "auth.json")
	}

	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create auth directory: %w", err)
	}

	return &FileStorage{
		path: path,
	}, nil
}

// SaveToken saves a token to a file.
func (f *FileStorage) SaveToken(ctx context.Context, token *types.Token) error {
	if token == nil {
		return fmt.Errorf("token is nil")
	}

	// Marshal token to JSON
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Write to file with restricted permissions
	if err := os.WriteFile(f.path, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}

	return nil
}

// LoadToken loads a token from a file.
func (f *FileStorage) LoadToken(ctx context.Context) (*types.Token, error) {
	// Check if file exists
	if _, err := os.Stat(f.path); os.IsNotExist(err) {
		return nil, fmt.Errorf("token file not found")
	}

	// Read file
	data, err := os.ReadFile(f.path)
	if err != nil {
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	// Unmarshal token
	var token types.Token
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}

// DeleteToken deletes the token file.
func (f *FileStorage) DeleteToken(ctx context.Context) error {
	if err := os.Remove(f.path); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}
	return nil
}

// GetPath returns the path to the token file.
func (f *FileStorage) GetPath() string {
	return f.path
}
