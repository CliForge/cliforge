// Package types defines common types used across the auth package.
package types

import (
	"time"
)

// Token represents an authentication token with metadata.
type Token struct {
	// AccessToken is the actual token value.
	AccessToken string `json:"access_token"`
	// RefreshToken is used to obtain a new access token.
	RefreshToken string `json:"refresh_token,omitempty"`
	// TokenType is the type of token (e.g., "Bearer").
	TokenType string `json:"token_type,omitempty"`
	// ExpiresAt is when the token expires.
	ExpiresAt time.Time `json:"expires_at,omitempty"`
	// Scopes are the granted scopes for this token.
	Scopes []string `json:"scopes,omitempty"`
	// Extra holds additional token metadata.
	Extra map[string]interface{} `json:"extra,omitempty"`
}

// IsExpired returns true if the token has expired.
func (t *Token) IsExpired() bool {
	if t.ExpiresAt.IsZero() {
		return false
	}
	// Add 30 second buffer to account for clock skew
	return time.Now().Add(30 * time.Second).After(t.ExpiresAt)
}

// IsValid returns true if the token is valid and not expired.
func (t *Token) IsValid() bool {
	return t.AccessToken != "" && !t.IsExpired()
}

// StorageConfig represents token storage configuration.
type StorageConfig struct {
	// Type is the storage backend type.
	Type StorageType `yaml:"type" json:"type"`
	// Path is the file path for file-based storage.
	Path string `yaml:"path,omitempty" json:"path,omitempty"`
	// KeyringService is the service name for keyring storage.
	KeyringService string `yaml:"keyring_service,omitempty" json:"keyring_service,omitempty"`
	// KeyringUser is the user name for keyring storage.
	KeyringUser string `yaml:"keyring_user,omitempty" json:"keyring_user,omitempty"`
}

// StorageType represents the type of token storage.
type StorageType string

const (
	// StorageTypeFile uses file-based storage.
	StorageTypeFile StorageType = "file"
	// StorageTypeKeyring uses OS keyring storage.
	StorageTypeKeyring StorageType = "keyring"
	// StorageTypeMemory uses in-memory storage.
	StorageTypeMemory StorageType = "memory"
)
