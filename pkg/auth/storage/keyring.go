package storage

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/CliForge/cliforge/pkg/auth/types"
	"github.com/zalando/go-keyring"
)

// KeyringStorage implements OS keyring-based token storage.
type KeyringStorage struct {
	service string
	user    string
}

// NewKeyringStorage creates a new keyring-based storage.
func NewKeyringStorage(config *types.StorageConfig) (*KeyringStorage, error) {
	service := config.KeyringService
	if service == "" {
		return nil, fmt.Errorf("keyring_service is required for keyring storage")
	}

	user := config.KeyringUser
	if user == "" {
		user = "default"
	}

	return &KeyringStorage{
		service: service,
		user:    user,
	}, nil
}

// SaveToken saves a token to the OS keyring.
func (k *KeyringStorage) SaveToken(ctx context.Context, token *types.Token) error {
	if token == nil {
		return fmt.Errorf("token is nil")
	}

	// Marshal token to JSON
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	// Store in keyring
	if err := keyring.Set(k.service, k.user, string(data)); err != nil {
		return fmt.Errorf("failed to store token in keyring: %w", err)
	}

	return nil
}

// LoadToken loads a token from the OS keyring.
func (k *KeyringStorage) LoadToken(ctx context.Context) (*types.Token, error) {
	// Retrieve from keyring
	data, err := keyring.Get(k.service, k.user)
	if err != nil {
		if err == keyring.ErrNotFound {
			return nil, fmt.Errorf("token not found in keyring")
		}
		return nil, fmt.Errorf("failed to retrieve token from keyring: %w", err)
	}

	// Unmarshal token
	var token types.Token
	if err := json.Unmarshal([]byte(data), &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}

	return &token, nil
}

// DeleteToken deletes the token from the OS keyring.
func (k *KeyringStorage) DeleteToken(ctx context.Context) error {
	if err := keyring.Delete(k.service, k.user); err != nil {
		if err == keyring.ErrNotFound {
			return nil
		}
		return fmt.Errorf("failed to delete token from keyring: %w", err)
	}
	return nil
}

// GetService returns the keyring service name.
func (k *KeyringStorage) GetService() string {
	return k.service
}

// GetUser returns the keyring user name.
func (k *KeyringStorage) GetUser() string {
	return k.user
}
