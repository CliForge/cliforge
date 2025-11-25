package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/CliForge/cliforge/pkg/auth/types"
)

// MemoryStorage implements in-memory token storage.
// This storage is ephemeral and tokens are lost when the process exits.
type MemoryStorage struct {
	mu    sync.RWMutex
	token *types.Token
}

// NewMemoryStorage creates a new in-memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

// SaveToken saves a token to memory.
func (m *MemoryStorage) SaveToken(ctx context.Context, token *types.Token) error {
	if token == nil {
		return fmt.Errorf("token is nil")
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Create a copy to avoid external modifications
	tokenCopy := *token
	if token.Scopes != nil {
		tokenCopy.Scopes = make([]string, len(token.Scopes))
		copy(tokenCopy.Scopes, token.Scopes)
	}
	if token.Extra != nil {
		tokenCopy.Extra = make(map[string]interface{})
		for k, v := range token.Extra {
			tokenCopy.Extra[k] = v
		}
	}

	m.token = &tokenCopy
	return nil
}

// LoadToken loads a token from memory.
func (m *MemoryStorage) LoadToken(ctx context.Context) (*types.Token, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	if m.token == nil {
		return nil, fmt.Errorf("token not found in memory")
	}

	// Return a copy to avoid external modifications
	tokenCopy := *m.token
	if m.token.Scopes != nil {
		tokenCopy.Scopes = make([]string, len(m.token.Scopes))
		copy(tokenCopy.Scopes, m.token.Scopes)
	}
	if m.token.Extra != nil {
		tokenCopy.Extra = make(map[string]interface{})
		for k, v := range m.token.Extra {
			tokenCopy.Extra[k] = v
		}
	}

	return &tokenCopy, nil
}

// DeleteToken deletes the token from memory.
func (m *MemoryStorage) DeleteToken(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.token = nil
	return nil
}

// Clear is an alias for DeleteToken.
func (m *MemoryStorage) Clear() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.token = nil
}
