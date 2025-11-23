package auth

import (
	"context"
	"fmt"

	"github.com/CliForge/cliforge/pkg/auth/storage"
)

// Manager coordinates authentication providers and token storage.
type Manager struct {
	authenticators map[string]Authenticator
	storages       map[string]TokenStorage
	defaultAuth    string
	cliName        string
}

// NewManager creates a new authentication manager.
func NewManager(cliName string) *Manager {
	return &Manager{
		authenticators: make(map[string]Authenticator),
		storages:       make(map[string]TokenStorage),
		cliName:        cliName,
	}
}

// RegisterAuthenticator registers an authenticator with a name.
func (m *Manager) RegisterAuthenticator(name string, auth Authenticator) error {
	if auth == nil {
		return fmt.Errorf("authenticator is nil")
	}

	if err := auth.Validate(); err != nil {
		return fmt.Errorf("invalid authenticator: %w", err)
	}

	m.authenticators[name] = auth

	// Set as default if it's the first one
	if m.defaultAuth == "" {
		m.defaultAuth = name
	}

	return nil
}

// RegisterStorage registers a token storage with a name.
func (m *Manager) RegisterStorage(name string, stor TokenStorage) {
	m.storages[name] = stor
}

// SetDefault sets the default authenticator.
func (m *Manager) SetDefault(name string) error {
	if _, exists := m.authenticators[name]; !exists {
		return fmt.Errorf("authenticator %s not found", name)
	}
	m.defaultAuth = name
	return nil
}

// GetAuthenticator returns an authenticator by name.
func (m *Manager) GetAuthenticator(name string) (Authenticator, error) {
	if name == "" {
		name = m.defaultAuth
	}

	auth, exists := m.authenticators[name]
	if !exists {
		return nil, fmt.Errorf("authenticator %s not found", name)
	}

	return auth, nil
}

// GetStorage returns a storage by name.
func (m *Manager) GetStorage(name string) (TokenStorage, error) {
	stor, exists := m.storages[name]
	if !exists {
		return nil, fmt.Errorf("storage %s not found", name)
	}
	return stor, nil
}

// CreateFromConfig creates authenticators and storages from configuration.
func (m *Manager) CreateFromConfig(configs map[string]*Config) error {
	for name, config := range configs {
		// Create authenticator
		auth, err := m.createAuthenticator(config)
		if err != nil {
			return fmt.Errorf("failed to create authenticator %s: %w", name, err)
		}

		if err := m.RegisterAuthenticator(name, auth); err != nil {
			return err
		}

		// Create storage if configured
		if config.Storage != nil {
			stor, err := m.createStorage(config.Storage)
			if err != nil {
				return fmt.Errorf("failed to create storage for %s: %w", name, err)
			}
			m.RegisterStorage(name, stor)
		}
	}

	return nil
}

// createAuthenticator creates an authenticator based on configuration.
func (m *Manager) createAuthenticator(config *Config) (Authenticator, error) {
	switch config.Type {
	case AuthTypeAPIKey:
		if config.APIKey == nil {
			return nil, fmt.Errorf("apikey config is required for API key auth")
		}
		return NewAPIKeyAuth(config.APIKey)

	case AuthTypeOAuth2:
		if config.OAuth2 == nil {
			return nil, fmt.Errorf("oauth2 config is required for OAuth2 auth")
		}
		return NewOAuth2Auth(config.OAuth2)

	case AuthTypeBasic:
		if config.Basic == nil {
			return nil, fmt.Errorf("basic config is required for Basic auth")
		}
		return NewBasicAuth(config.Basic)

	case AuthTypeNone:
		return &NoneAuth{}, nil

	default:
		return nil, fmt.Errorf("unsupported auth type: %s", config.Type)
	}
}

// createStorage creates a storage based on configuration.
func (m *Manager) createStorage(config *StorageConfig) (TokenStorage, error) {
	factory := storage.NewFactory()
	return factory.Create(config, m.cliName)
}

// Authenticate performs authentication using the specified authenticator.
func (m *Manager) Authenticate(ctx context.Context, authName string) (*Token, error) {
	auth, err := m.GetAuthenticator(authName)
	if err != nil {
		return nil, err
	}

	token, err := auth.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	// Save token if storage is available
	if stor, err := m.GetStorage(authName); err == nil {
		_ = stor.SaveToken(ctx, token)
	}

	return token, nil
}

// GetToken retrieves a valid token, refreshing if necessary.
func (m *Manager) GetToken(ctx context.Context, authName string) (*Token, error) {
	if authName == "" {
		authName = m.defaultAuth
	}

	auth, err := m.GetAuthenticator(authName)
	if err != nil {
		return nil, err
	}

	// Try to load from storage
	if stor, err := m.GetStorage(authName); err == nil {
		token, err := stor.LoadToken(ctx)
		if err == nil && token != nil {
			// Check if token is still valid
			if token.IsValid() {
				return token, nil
			}

			// Try to refresh if expired
			if token.RefreshToken != "" {
				refreshed, err := auth.RefreshToken(ctx, token)
				if err == nil {
					_ = stor.SaveToken(ctx, refreshed)
					return refreshed, nil
				}
			}
		}
	}

	// Perform new authentication
	return m.Authenticate(ctx, authName)
}

// RefreshToken refreshes a token using the specified authenticator.
func (m *Manager) RefreshToken(ctx context.Context, authName string, token *Token) (*Token, error) {
	auth, err := m.GetAuthenticator(authName)
	if err != nil {
		return nil, err
	}

	refreshed, err := auth.RefreshToken(ctx, token)
	if err != nil {
		return nil, err
	}

	// Save refreshed token if storage is available
	if stor, err := m.GetStorage(authName); err == nil {
		_ = stor.SaveToken(ctx, refreshed)
	}

	return refreshed, nil
}

// Logout removes stored tokens for the specified authenticator.
func (m *Manager) Logout(ctx context.Context, authName string) error {
	if authName == "" {
		authName = m.defaultAuth
	}

	if stor, err := m.GetStorage(authName); err == nil {
		return stor.DeleteToken(ctx)
	}

	return nil
}

// LogoutAll removes all stored tokens.
func (m *Manager) LogoutAll(ctx context.Context) error {
	var lastErr error

	for _, stor := range m.storages {
		if err := stor.DeleteToken(ctx); err != nil {
			lastErr = err
		}
	}

	return lastErr
}

// GetAuthenticatedClient creates an authenticated HTTP client.
func (m *Manager) GetAuthenticatedClient(authName string, storName string) (*AuthenticatedClient, error) {
	if authName == "" {
		authName = m.defaultAuth
	}
	if storName == "" {
		storName = authName
	}

	auth, err := m.GetAuthenticator(authName)
	if err != nil {
		return nil, err
	}

	var stor TokenStorage
	if s, err := m.GetStorage(storName); err == nil {
		stor = s
	}

	return NewAuthenticatedClient(nil, auth, stor), nil
}

// NoneAuth is a no-op authenticator for when no authentication is required.
type NoneAuth struct{}

// Type returns the authentication type.
func (n *NoneAuth) Type() AuthType {
	return AuthTypeNone
}

// Authenticate does nothing for no-auth.
func (n *NoneAuth) Authenticate(ctx context.Context) (*Token, error) {
	return &Token{}, nil
}

// RefreshToken does nothing for no-auth.
func (n *NoneAuth) RefreshToken(ctx context.Context, token *Token) (*Token, error) {
	return nil, fmt.Errorf("no authentication configured")
}

// GetHeaders returns no headers for no-auth.
func (n *NoneAuth) GetHeaders(token *Token) map[string]string {
	return nil
}

// Validate always succeeds for no-auth.
func (n *NoneAuth) Validate() error {
	return nil
}

// ListAuthenticators returns the names of all registered authenticators.
func (m *Manager) ListAuthenticators() []string {
	names := make([]string, 0, len(m.authenticators))
	for name := range m.authenticators {
		names = append(names, name)
	}
	return names
}

// ListStorages returns the names of all registered storages.
func (m *Manager) ListStorages() []string {
	names := make([]string, 0, len(m.storages))
	for name := range m.storages {
		names = append(names, name)
	}
	return names
}
