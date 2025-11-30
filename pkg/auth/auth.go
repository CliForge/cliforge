// Package auth provides authentication mechanisms for CLI applications.
//
// The auth package supports multiple authentication strategies including
// API keys, OAuth2 (multiple flows), and HTTP Basic authentication. It handles
// token acquisition, refresh, storage, and automatic injection into HTTP requests.
//
// # Supported Authentication Types
//
//   - API Key: Header or query parameter based authentication
//   - OAuth2: Authorization code, client credentials, password, device code, direct token flows
//   - Basic: HTTP Basic authentication with username/password
//   - None: No authentication (for public APIs)
//
// # OAuth2 Flows
//
//   - Authorization Code with PKCE: Browser-based flow with security enhancement
//   - Client Credentials: Service-to-service authentication
//   - Password (Resource Owner): Direct username/password exchange
//   - Device Code: Headless/remote device authentication
//   - Token Injection: Direct token with automatic type detection
//
// # Token Resolution
//
// The TokenResolver provides ROSA-compatible token lookup with automatic fallback:
//
//	resolver := auth.NewTokenResolver(
//	    auth.WithFlagToken(flagValue),           // 1. --token flag
//	    auth.WithEnvVars("ROSA_TOKEN", "OCM_TOKEN"), // 2-3. Environment variables
//	    auth.WithStorage(storage),               // 4. Config file
//	    auth.WithPromptFunc(promptFunc),         // 5. Interactive prompt
//	)
//	token, source, err := resolver.Resolve(ctx)
//
// # JWT Token Detection
//
// Automatic detection of JWT token types for proper routing:
//
//	tokenType, err := auth.DetectTokenType(tokenString)
//	// Returns: TokenTypeAccess, TokenTypeRefresh, TokenTypeOffline, TokenTypeBearer
//
//	username, err := auth.ExtractUsername(tokenString)
//	// Extracts preferred_username or username claim from JWT
//
//	isEncrypted := auth.IsEncryptedToken(tokenString)
//	// Detects JWE (5-part encrypted) vs JWT (3-part) tokens
//
// # Storage Options
//
//   - File: Store tokens in XDG-compliant filesystem locations
//   - Keyring: Store tokens in system keyring (macOS Keychain, GNOME Keyring, Windows Credential Manager)
//   - Memory: Store tokens in-memory only (not persisted)
//
// # Example: API Key Authentication
//
//	config := &auth.APIKeyConfig{
//	    Key: "sk-...",
//	    Location: auth.APIKeyLocationHeader,
//	    Name: "Authorization",
//	    Prefix: "Bearer ",
//	}
//	authenticator, _ := auth.NewAPIKeyAuth(config)
//	client := auth.NewAuthenticatedClient(http.DefaultClient, authenticator, storage)
//
// # Example: OAuth2 with Browser Flow
//
//	oauth2Config := &auth.OAuth2Config{
//	    ClientID: "my-cli",
//	    AuthURL: "https://sso.example.com/auth",
//	    TokenURL: "https://sso.example.com/token",
//	    Flow: auth.OAuth2FlowAuthorizationCode,
//	    PKCE: true,
//	    AutoOpenBrowser: true,
//	    RedirectPort: 9998,
//	}
//	authenticator, _ := auth.NewOAuth2Auth(oauth2Config)
//	token, _ := authenticator.Authenticate(ctx)
//
// # Example: Token Injection with Auto-Detection
//
//	oauth2Config := &auth.OAuth2Config{
//	    ClientID: "my-cli",
//	    TokenURL: "https://sso.example.com/token",
//	    Token: "eyJhbGciOiJSUzI1NiIs...",  // Offline or access token
//	    Flow: auth.OAuth2FlowToken,
//	}
//	authenticator, _ := auth.NewOAuth2Auth(oauth2Config)
//	token, _ := authenticator.Authenticate(ctx)
//	// Automatically detects token type and routes appropriately
//
// The package automatically handles token refresh when tokens expire and
// stores credentials securely based on the configured storage backend.
package auth

import (
	"context"
	"net/http"

	"github.com/CliForge/cliforge/pkg/auth/types"
)

// AuthType represents the type of authentication mechanism.
type AuthType string

const (
	// AuthTypeAPIKey represents API key authentication.
	AuthTypeAPIKey AuthType = "apikey"
	// AuthTypeOAuth2 represents OAuth2 authentication.
	AuthTypeOAuth2 AuthType = "oauth2"
	// AuthTypeBasic represents Basic authentication.
	AuthTypeBasic AuthType = "basic"
	// AuthTypeNone represents no authentication.
	AuthTypeNone AuthType = "none"
)

// Token is an alias for types.Token for backward compatibility.
type Token = types.Token

// Authenticator is the main interface for authentication mechanisms.
type Authenticator interface {
	// Type returns the authentication type.
	Type() AuthType

	// Authenticate performs the authentication flow and returns a token.
	Authenticate(ctx context.Context) (*Token, error)

	// RefreshToken refreshes an expired token if supported.
	RefreshToken(ctx context.Context, token *Token) (*Token, error)

	// GetHeaders returns HTTP headers to be added to requests.
	GetHeaders(token *Token) map[string]string

	// Validate checks if the authenticator configuration is valid.
	Validate() error
}

// Config represents authentication configuration.
type Config struct {
	// Type is the authentication type.
	Type AuthType `yaml:"type" json:"type"`

	// APIKey configuration (for AuthTypeAPIKey)
	APIKey *APIKeyConfig `yaml:"apikey,omitempty" json:"apikey,omitempty"`

	// OAuth2 configuration (for AuthTypeOAuth2)
	OAuth2 *OAuth2Config `yaml:"oauth2,omitempty" json:"oauth2,omitempty"`

	// Basic configuration (for AuthTypeBasic)
	Basic *BasicConfig `yaml:"basic,omitempty" json:"basic,omitempty"`

	// Storage configuration for token persistence.
	Storage *StorageConfig `yaml:"storage,omitempty" json:"storage,omitempty"`
}

// APIKeyConfig represents API key authentication configuration.
type APIKeyConfig struct {
	// Key is the API key value or environment variable name.
	Key string `yaml:"key" json:"key"`
	// Location specifies where to send the API key (header, query).
	Location APIKeyLocation `yaml:"location" json:"location"`
	// Name is the header name or query parameter name.
	Name string `yaml:"name" json:"name"`
	// Prefix is an optional prefix (e.g., "Bearer ", "Token ").
	Prefix string `yaml:"prefix,omitempty" json:"prefix,omitempty"`
	// EnvVar is the environment variable to read the key from.
	EnvVar string `yaml:"env_var,omitempty" json:"env_var,omitempty"`
}

// APIKeyLocation specifies where the API key should be placed.
type APIKeyLocation string

const (
	// APIKeyLocationHeader places the API key in an HTTP header.
	APIKeyLocationHeader APIKeyLocation = "header"
	// APIKeyLocationQuery places the API key in a query parameter.
	APIKeyLocationQuery APIKeyLocation = "query"
)

// OAuth2Config represents OAuth2 authentication configuration.
type OAuth2Config struct {
	// ClientID is the OAuth2 client identifier.
	ClientID string `yaml:"client_id" json:"client_id"`
	// ClientSecret is the OAuth2 client secret.
	ClientSecret string `yaml:"client_secret,omitempty" json:"client_secret,omitempty"`
	// TokenURL is the OAuth2 token endpoint.
	TokenURL string `yaml:"token_url" json:"token_url"`
	// AuthURL is the OAuth2 authorization endpoint.
	AuthURL string `yaml:"auth_url,omitempty" json:"auth_url,omitempty"`
	// RedirectURL is the callback URL for authorization code flow.
	RedirectURL string `yaml:"redirect_url,omitempty" json:"redirect_url,omitempty"`
	// Scopes are the requested OAuth2 scopes.
	Scopes []string `yaml:"scopes,omitempty" json:"scopes,omitempty"`
	// Flow is the OAuth2 flow type.
	Flow OAuth2Flow `yaml:"flow" json:"flow"`
	// PKCE enables Proof Key for Code Exchange (recommended for public clients).
	PKCE bool `yaml:"pkce,omitempty" json:"pkce,omitempty"`
	// Username for password flow.
	Username string `yaml:"username,omitempty" json:"username,omitempty"`
	// Password for password flow.
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
	// DeviceCodeURL is the device authorization endpoint.
	DeviceCodeURL string `yaml:"device_code_url,omitempty" json:"device_code_url,omitempty"`
	// EndpointParams are additional parameters for token requests.
	EndpointParams map[string]string `yaml:"endpoint_params,omitempty" json:"endpoint_params,omitempty"`
	// Token is a pre-provided access or refresh/offline token.
	// When set, the token flow is used and the token type is auto-detected via JWT parsing.
	// Supports both access tokens (Bearer, "") and refresh/offline tokens ("Refresh", "Offline").
	// This enables ROSA-compatible token injection from console.redhat.com/openshift/token.
	Token string `yaml:"token,omitempty" json:"token,omitempty"`
	// AutoOpenBrowser enables automatic browser opening for authorization code flows.
	// When true, the system browser is launched automatically. When false, only the URL is displayed.
	// This is useful for headless environments or when the browser should not be opened automatically.
	AutoOpenBrowser bool `yaml:"auto_open_browser,omitempty" json:"auto_open_browser,omitempty"`
	// RedirectPort specifies the local port for OAuth callback server (default: 9998).
	// Port 9998 is used for ROSA compatibility. The callback URL becomes http://localhost:{port}/callback.
	// For authorization code flow, a local HTTP server listens on this port to receive the auth code.
	RedirectPort int `yaml:"redirect_port,omitempty" json:"redirect_port,omitempty"`
}

// OAuth2Flow represents the OAuth2 flow type.
type OAuth2Flow string

const (
	// OAuth2FlowAuthorizationCode is the authorization code flow.
	OAuth2FlowAuthorizationCode OAuth2Flow = "authorization_code"
	// OAuth2FlowClientCredentials is the client credentials flow.
	OAuth2FlowClientCredentials OAuth2Flow = "client_credentials"
	// OAuth2FlowPassword is the resource owner password credentials flow.
	OAuth2FlowPassword OAuth2Flow = "password"
	// OAuth2FlowDeviceCode is the device authorization flow.
	OAuth2FlowDeviceCode OAuth2Flow = "device_code"
	// OAuth2FlowToken is the direct token injection flow.
	OAuth2FlowToken OAuth2Flow = "token"
)

// BasicConfig represents Basic authentication configuration.
type BasicConfig struct {
	// Username is the username for basic auth.
	Username string `yaml:"username" json:"username"`
	// Password is the password for basic auth.
	Password string `yaml:"password,omitempty" json:"password,omitempty"`
	// EnvUsername is the environment variable to read username from.
	EnvUsername string `yaml:"env_username,omitempty" json:"env_username,omitempty"`
	// EnvPassword is the environment variable to read password from.
	EnvPassword string `yaml:"env_password,omitempty" json:"env_password,omitempty"`
}

// StorageConfig is an alias for types.StorageConfig for backward compatibility.
type StorageConfig = types.StorageConfig

// StorageType is an alias for types.StorageType for backward compatibility.
type StorageType = types.StorageType

// Storage type constants
const (
	StorageTypeFile    = types.StorageTypeFile
	StorageTypeKeyring = types.StorageTypeKeyring
	StorageTypeMemory  = types.StorageTypeMemory
)

// HTTPClient is an interface for making HTTP requests with authentication.
type HTTPClient interface {
	// Do executes an HTTP request with authentication headers.
	Do(req *http.Request) (*http.Response, error)
}

// AuthenticatedClient wraps an HTTP client with authentication.
type AuthenticatedClient struct {
	client        *http.Client
	authenticator Authenticator
	storage       TokenStorage
	token         *Token
}

// NewAuthenticatedClient creates a new authenticated HTTP client.
func NewAuthenticatedClient(client *http.Client, auth Authenticator, storage TokenStorage) *AuthenticatedClient {
	if client == nil {
		client = http.DefaultClient
	}
	return &AuthenticatedClient{
		client:        client,
		authenticator: auth,
		storage:       storage,
	}
}

// Do executes an HTTP request with authentication.
func (c *AuthenticatedClient) Do(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	// Get or refresh token
	token, err := c.getValidToken(ctx)
	if err != nil {
		return nil, err
	}

	// Add authentication headers
	headers := c.authenticator.GetHeaders(token)
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	return c.client.Do(req)
}

// getValidToken retrieves a valid token, refreshing if necessary.
func (c *AuthenticatedClient) getValidToken(ctx context.Context) (*Token, error) {
	// Try to use cached token
	if c.token != nil && c.token.IsValid() {
		return c.token, nil
	}

	// Try to load from storage
	if c.storage != nil {
		stored, err := c.storage.LoadToken(ctx)
		if err == nil && stored != nil && stored.IsValid() {
			c.token = stored
			return stored, nil
		}

		// Try to refresh expired token
		if stored != nil && stored.RefreshToken != "" {
			refreshed, err := c.authenticator.RefreshToken(ctx, stored)
			if err == nil && refreshed != nil {
				c.token = refreshed
				if err := c.storage.SaveToken(ctx, refreshed); err == nil {
					return refreshed, nil
				}
			}
		}
	}

	// Perform new authentication
	token, err := c.authenticator.Authenticate(ctx)
	if err != nil {
		return nil, err
	}

	// Cache and store the token
	c.token = token
	if c.storage != nil {
		_ = c.storage.SaveToken(ctx, token)
	}

	return token, nil
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
