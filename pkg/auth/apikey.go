package auth

import (
	"context"
	"fmt"
	"os"
	"strings"
)

// APIKeyAuth implements API key authentication.
type APIKeyAuth struct {
	config *APIKeyConfig
	key    string
}

// NewAPIKeyAuth creates a new API key authenticator.
func NewAPIKeyAuth(config *APIKeyConfig) (*APIKeyAuth, error) {
	if config == nil {
		return nil, fmt.Errorf("API key config is required")
	}

	auth := &APIKeyAuth{
		config: config,
	}

	if err := auth.Validate(); err != nil {
		return nil, err
	}

	// Resolve the API key
	if err := auth.resolveKey(); err != nil {
		return nil, err
	}

	return auth, nil
}

// Type returns the authentication type.
func (a *APIKeyAuth) Type() AuthType {
	return AuthTypeAPIKey
}

// Authenticate performs API key authentication.
// For API keys, this simply validates and returns the key as a token.
func (a *APIKeyAuth) Authenticate(ctx context.Context) (*Token, error) {
	if a.key == "" {
		return nil, fmt.Errorf("API key not configured")
	}

	return &Token{
		AccessToken: a.key,
		TokenType:   "apikey",
	}, nil
}

// RefreshToken is not applicable for API key authentication.
func (a *APIKeyAuth) RefreshToken(ctx context.Context, token *Token) (*Token, error) {
	return nil, fmt.Errorf("API key authentication does not support token refresh")
}

// GetHeaders returns the HTTP headers for API key authentication.
func (a *APIKeyAuth) GetHeaders(token *Token) map[string]string {
	if token == nil || token.AccessToken == "" {
		return nil
	}

	headers := make(map[string]string)

	if a.config.Location == APIKeyLocationHeader {
		value := token.AccessToken
		if a.config.Prefix != "" {
			value = a.config.Prefix + value
		}
		headers[a.config.Name] = value
	}

	return headers
}

// GetQueryParams returns query parameters for API key authentication.
// This should be called when Location is APIKeyLocationQuery.
func (a *APIKeyAuth) GetQueryParams(token *Token) map[string]string {
	if token == nil || token.AccessToken == "" {
		return nil
	}

	if a.config.Location == APIKeyLocationQuery {
		return map[string]string{
			a.config.Name: token.AccessToken,
		}
	}

	return nil
}

// Validate checks if the API key configuration is valid.
func (a *APIKeyAuth) Validate() error {
	if a.config == nil {
		return fmt.Errorf("API key config is required")
	}

	if a.config.Name == "" {
		return fmt.Errorf("API key name is required")
	}

	if a.config.Location == "" {
		return fmt.Errorf("API key location is required")
	}

	if a.config.Location != APIKeyLocationHeader && a.config.Location != APIKeyLocationQuery {
		return fmt.Errorf("invalid API key location: %s", a.config.Location)
	}

	if a.config.Key == "" && a.config.EnvVar == "" {
		return fmt.Errorf("either API key or environment variable must be specified")
	}

	return nil
}

// resolveKey resolves the API key from configuration or environment.
func (a *APIKeyAuth) resolveKey() error {
	// Check environment variable first
	if a.config.EnvVar != "" {
		if value := os.Getenv(a.config.EnvVar); value != "" {
			a.key = value
			return nil
		}
	}

	// Fall back to configured key
	if a.config.Key != "" {
		// Check if the key value is an environment variable reference
		if strings.HasPrefix(a.config.Key, "$") {
			envName := strings.TrimPrefix(a.config.Key, "$")
			if value := os.Getenv(envName); value != "" {
				a.key = value
				return nil
			}
			return fmt.Errorf("environment variable %s not found", envName)
		}
		a.key = a.config.Key
		return nil
	}

	return fmt.Errorf("API key not found in configuration or environment")
}

// SetKey allows setting the API key dynamically.
func (a *APIKeyAuth) SetKey(key string) {
	a.key = key
}
