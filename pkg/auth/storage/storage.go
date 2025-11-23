// Package storage provides token storage implementations for authentication.
package storage

import (
	"context"
	"fmt"

	"github.com/CliForge/cliforge/pkg/auth/types"
)

// Factory creates token storage instances based on configuration.
type Factory struct{}

// NewFactory creates a new storage factory.
func NewFactory() *Factory {
	return &Factory{}
}

// TokenStorage is an interface for storing and retrieving tokens.
type TokenStorage interface {
	// SaveToken stores a token.
	SaveToken(ctx context.Context, token *types.Token) error
	// LoadToken retrieves a stored token.
	LoadToken(ctx context.Context) (*types.Token, error)
	// DeleteToken removes a stored token.
	DeleteToken(ctx context.Context) error
}

// Create creates a token storage instance based on the configuration.
func (f *Factory) Create(config *types.StorageConfig, cliName string) (TokenStorage, error) {
	if config == nil {
		return nil, fmt.Errorf("storage config is required")
	}

	switch config.Type {
	case types.StorageTypeFile:
		return NewFileStorage(config, cliName)
	case types.StorageTypeKeyring:
		return NewKeyringStorage(config)
	case types.StorageTypeMemory:
		return NewMemoryStorage(), nil
	default:
		return nil, fmt.Errorf("unsupported storage type: %s", config.Type)
	}
}

// MultiStorage implements a multi-tier token storage with fallback.
// It tries to save to all storages but reads from the first available.
type MultiStorage struct {
	storages []TokenStorage
}

// NewMultiStorage creates a new multi-tier storage.
func NewMultiStorage(storages ...TokenStorage) *MultiStorage {
	return &MultiStorage{
		storages: storages,
	}
}

// SaveToken saves the token to all available storages.
func (m *MultiStorage) SaveToken(ctx context.Context, token *types.Token) error {
	var lastErr error
	saved := false

	for _, storage := range m.storages {
		if err := storage.SaveToken(ctx, token); err != nil {
			lastErr = err
		} else {
			saved = true
		}
	}

	if !saved && lastErr != nil {
		return lastErr
	}

	return nil
}

// LoadToken loads the token from the first available storage.
func (m *MultiStorage) LoadToken(ctx context.Context) (*types.Token, error) {
	for _, storage := range m.storages {
		token, err := storage.LoadToken(ctx)
		if err == nil && token != nil {
			return token, nil
		}
	}

	return nil, fmt.Errorf("token not found in any storage")
}

// DeleteToken deletes the token from all storages.
func (m *MultiStorage) DeleteToken(ctx context.Context) error {
	var lastErr error

	for _, storage := range m.storages {
		if err := storage.DeleteToken(ctx); err != nil {
			lastErr = err
		}
	}

	return lastErr
}
