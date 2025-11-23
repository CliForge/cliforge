package auth

import (
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
)

// BasicAuth implements HTTP Basic authentication.
type BasicAuth struct {
	config   *BasicConfig
	username string
	password string
}

// NewBasicAuth creates a new Basic authenticator.
func NewBasicAuth(config *BasicConfig) (*BasicAuth, error) {
	if config == nil {
		return nil, fmt.Errorf("Basic auth config is required")
	}

	auth := &BasicAuth{
		config: config,
	}

	if err := auth.Validate(); err != nil {
		return nil, err
	}

	// Resolve credentials
	if err := auth.resolveCredentials(); err != nil {
		return nil, err
	}

	return auth, nil
}

// Type returns the authentication type.
func (b *BasicAuth) Type() AuthType {
	return AuthTypeBasic
}

// Authenticate performs Basic authentication.
// For Basic auth, this creates a token containing the credentials.
func (b *BasicAuth) Authenticate(ctx context.Context) (*Token, error) {
	if b.username == "" || b.password == "" {
		return nil, fmt.Errorf("username and password are required")
	}

	// Encode credentials
	credentials := b.username + ":" + b.password
	encoded := base64.StdEncoding.EncodeToString([]byte(credentials))

	return &Token{
		AccessToken: encoded,
		TokenType:   "Basic",
		Extra: map[string]interface{}{
			"username": b.username,
		},
	}, nil
}

// RefreshToken is not applicable for Basic authentication.
func (b *BasicAuth) RefreshToken(ctx context.Context, token *Token) (*Token, error) {
	return nil, fmt.Errorf("Basic authentication does not support token refresh")
}

// GetHeaders returns the HTTP headers for Basic authentication.
func (b *BasicAuth) GetHeaders(token *Token) map[string]string {
	if token == nil || token.AccessToken == "" {
		return nil
	}

	return map[string]string{
		"Authorization": "Basic " + token.AccessToken,
	}
}

// Validate checks if the Basic auth configuration is valid.
func (b *BasicAuth) Validate() error {
	if b.config == nil {
		return fmt.Errorf("Basic auth config is required")
	}

	if b.config.Username == "" && b.config.EnvUsername == "" {
		return fmt.Errorf("username or env_username is required")
	}

	if b.config.Password == "" && b.config.EnvPassword == "" {
		return fmt.Errorf("password or env_password is required")
	}

	return nil
}

// resolveCredentials resolves credentials from configuration or environment.
func (b *BasicAuth) resolveCredentials() error {
	// Resolve username
	if err := b.resolveUsername(); err != nil {
		return err
	}

	// Resolve password
	if err := b.resolvePassword(); err != nil {
		return err
	}

	return nil
}

// resolveUsername resolves the username from configuration or environment.
func (b *BasicAuth) resolveUsername() error {
	// Check environment variable first
	if b.config.EnvUsername != "" {
		if value := os.Getenv(b.config.EnvUsername); value != "" {
			b.username = value
			return nil
		}
	}

	// Fall back to configured username
	if b.config.Username != "" {
		// Check if the username is an environment variable reference
		if strings.HasPrefix(b.config.Username, "$") {
			envName := strings.TrimPrefix(b.config.Username, "$")
			if value := os.Getenv(envName); value != "" {
				b.username = value
				return nil
			}
			return fmt.Errorf("environment variable %s not found", envName)
		}
		b.username = b.config.Username
		return nil
	}

	return fmt.Errorf("username not found in configuration or environment")
}

// resolvePassword resolves the password from configuration or environment.
func (b *BasicAuth) resolvePassword() error {
	// Check environment variable first
	if b.config.EnvPassword != "" {
		if value := os.Getenv(b.config.EnvPassword); value != "" {
			b.password = value
			return nil
		}
	}

	// Fall back to configured password
	if b.config.Password != "" {
		// Check if the password is an environment variable reference
		if strings.HasPrefix(b.config.Password, "$") {
			envName := strings.TrimPrefix(b.config.Password, "$")
			if value := os.Getenv(envName); value != "" {
				b.password = value
				return nil
			}
			return fmt.Errorf("environment variable %s not found", envName)
		}
		b.password = b.config.Password
		return nil
	}

	return fmt.Errorf("password not found in configuration or environment")
}

// SetCredentials allows setting credentials dynamically.
func (b *BasicAuth) SetCredentials(username, password string) {
	b.username = username
	b.password = password
}

// GetUsername returns the configured username.
func (b *BasicAuth) GetUsername() string {
	return b.username
}
