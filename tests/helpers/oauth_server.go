package helpers

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"time"
)

// MockOAuth2Server provides a mock OAuth2 server for testing.
type MockOAuth2Server struct {
	server        *httptest.Server
	clientID      string
	clientSecret  string
	tokens        map[string]*OAuth2Token
	authCodes     map[string]*AuthCode
	mu            sync.RWMutex
	autoApprove   bool
	tokenLifetime time.Duration
}

// OAuth2Token represents an OAuth2 token.
type OAuth2Token struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	TokenType    string    `json:"token_type"`
	ExpiresIn    int       `json:"expires_in"`
	Scope        string    `json:"scope,omitempty"`
	IssuedAt     time.Time `json:"-"`
}

// AuthCode represents an authorization code.
type AuthCode struct {
	Code        string
	ClientID    string
	RedirectURI string
	Scope       string
	ExpiresAt   time.Time
	Used        bool
}

// NewMockOAuth2Server creates a new mock OAuth2 server.
func NewMockOAuth2Server(clientID, clientSecret string) *MockOAuth2Server {
	ms := &MockOAuth2Server{
		clientID:      clientID,
		clientSecret:  clientSecret,
		tokens:        make(map[string]*OAuth2Token),
		authCodes:     make(map[string]*AuthCode),
		autoApprove:   true,
		tokenLifetime: 1 * time.Hour,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/oauth/authorize", ms.handleAuthorize)
	mux.HandleFunc("/oauth/token", ms.handleToken)
	mux.HandleFunc("/oauth/revoke", ms.handleRevoke)
	mux.HandleFunc("/oauth/introspect", ms.handleIntrospect)

	ms.server = httptest.NewServer(mux)
	return ms
}

// URL returns the OAuth2 server URL.
func (ms *MockOAuth2Server) URL() string {
	return ms.server.URL
}

// Close shuts down the OAuth2 server.
func (ms *MockOAuth2Server) Close() {
	ms.server.Close()
}

// SetAutoApprove sets whether to auto-approve authorization requests.
func (ms *MockOAuth2Server) SetAutoApprove(auto bool) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.autoApprove = auto
}

// SetTokenLifetime sets the token lifetime.
func (ms *MockOAuth2Server) SetTokenLifetime(lifetime time.Duration) {
	ms.mu.Lock()
	defer ms.mu.Unlock()
	ms.tokenLifetime = lifetime
}

// handleAuthorize handles the authorization endpoint.
func (ms *MockOAuth2Server) handleAuthorize(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	clientID := r.FormValue("client_id")
	redirectURI := r.FormValue("redirect_uri")
	scope := r.FormValue("scope")
	state := r.FormValue("state")
	responseType := r.FormValue("response_type")

	// Validate client ID
	if clientID != ms.clientID {
		http.Error(w, "Invalid client_id", http.StatusBadRequest)
		return
	}

	// Validate response type
	if responseType != "code" {
		http.Error(w, "Unsupported response_type", http.StatusBadRequest)
		return
	}

	// Generate authorization code
	code := ms.generateAuthCode(clientID, redirectURI, scope)

	// Build redirect URL
	redirectURL, err := url.Parse(redirectURI)
	if err != nil {
		http.Error(w, "Invalid redirect_uri", http.StatusBadRequest)
		return
	}

	q := redirectURL.Query()
	q.Set("code", code)
	if state != "" {
		q.Set("state", state)
	}
	redirectURL.RawQuery = q.Encode()

	// Redirect back to client
	http.Redirect(w, r, redirectURL.String(), http.StatusFound)
}

// handleToken handles the token endpoint.
func (ms *MockOAuth2Server) handleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	grantType := r.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		ms.handleAuthCodeGrant(w, r)
	case "refresh_token":
		ms.handleRefreshTokenGrant(w, r)
	case "client_credentials":
		ms.handleClientCredentialsGrant(w, r)
	default:
		ms.sendError(w, "unsupported_grant_type", "Unsupported grant type", http.StatusBadRequest)
	}
}

// handleAuthCodeGrant handles authorization code grant.
func (ms *MockOAuth2Server) handleAuthCodeGrant(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	redirectURI := r.FormValue("redirect_uri")

	// Validate client credentials
	if clientID != ms.clientID || clientSecret != ms.clientSecret {
		ms.sendError(w, "invalid_client", "Invalid client credentials", http.StatusUnauthorized)
		return
	}

	// Validate authorization code
	ms.mu.Lock()
	authCode, exists := ms.authCodes[code]
	if !exists || authCode.Used || time.Now().After(authCode.ExpiresAt) {
		ms.mu.Unlock()
		ms.sendError(w, "invalid_grant", "Invalid authorization code", http.StatusBadRequest)
		return
	}

	// Validate redirect URI
	if authCode.RedirectURI != redirectURI {
		ms.mu.Unlock()
		ms.sendError(w, "invalid_grant", "Redirect URI mismatch", http.StatusBadRequest)
		return
	}

	// Mark code as used
	authCode.Used = true
	ms.mu.Unlock()

	// Generate tokens
	token := ms.generateToken(authCode.Scope)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(token)
}

// handleRefreshTokenGrant handles refresh token grant.
func (ms *MockOAuth2Server) handleRefreshTokenGrant(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.FormValue("refresh_token")
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")

	// Validate client credentials
	if clientID != ms.clientID || clientSecret != ms.clientSecret {
		ms.sendError(w, "invalid_client", "Invalid client credentials", http.StatusUnauthorized)
		return
	}

	// Validate refresh token
	ms.mu.RLock()
	oldToken, exists := ms.tokens[refreshToken]
	ms.mu.RUnlock()

	if !exists {
		ms.sendError(w, "invalid_grant", "Invalid refresh token", http.StatusBadRequest)
		return
	}

	// Generate new tokens
	token := ms.generateToken(oldToken.Scope)

	// Send response
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(token)
}

// handleClientCredentialsGrant handles client credentials grant.
func (ms *MockOAuth2Server) handleClientCredentialsGrant(w http.ResponseWriter, r *http.Request) {
	clientID := r.FormValue("client_id")
	clientSecret := r.FormValue("client_secret")
	scope := r.FormValue("scope")

	// Validate client credentials
	if clientID != ms.clientID || clientSecret != ms.clientSecret {
		ms.sendError(w, "invalid_client", "Invalid client credentials", http.StatusUnauthorized)
		return
	}

	// Generate token (no refresh token for client credentials)
	token := ms.generateToken(scope)
	token.RefreshToken = ""

	// Send response
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(token)
}

// handleRevoke handles token revocation.
func (ms *MockOAuth2Server) handleRevoke(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.FormValue("token")

	ms.mu.Lock()
	delete(ms.tokens, token)
	ms.mu.Unlock()

	w.WriteHeader(http.StatusOK)
}

// handleIntrospect handles token introspection.
func (ms *MockOAuth2Server) handleIntrospect(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	token := r.FormValue("token")

	ms.mu.RLock()
	tokenData, exists := ms.tokens[token]
	ms.mu.RUnlock()

	response := map[string]interface{}{
		"active": false,
	}

	if exists && time.Since(tokenData.IssuedAt) < ms.tokenLifetime {
		response["active"] = true
		response["scope"] = tokenData.Scope
		response["token_type"] = tokenData.TokenType
		response["exp"] = tokenData.IssuedAt.Add(ms.tokenLifetime).Unix()
	}

	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(response)
}

// generateAuthCode generates a new authorization code.
func (ms *MockOAuth2Server) generateAuthCode(clientID, redirectURI, scope string) string {
	code := generateRandomString(32)

	ms.mu.Lock()
	ms.authCodes[code] = &AuthCode{
		Code:        code,
		ClientID:    clientID,
		RedirectURI: redirectURI,
		Scope:       scope,
		ExpiresAt:   time.Now().Add(5 * time.Minute),
		Used:        false,
	}
	ms.mu.Unlock()

	return code
}

// generateToken generates a new OAuth2 token.
func (ms *MockOAuth2Server) generateToken(scope string) *OAuth2Token {
	accessToken := generateRandomString(64)
	refreshToken := generateRandomString(64)

	token := &OAuth2Token{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(ms.tokenLifetime.Seconds()),
		Scope:        scope,
		IssuedAt:     time.Now(),
	}

	ms.mu.Lock()
	ms.tokens[accessToken] = token
	ms.tokens[refreshToken] = token
	ms.mu.Unlock()

	return token
}

// sendError sends an OAuth2 error response.
func (ms *MockOAuth2Server) sendError(w http.ResponseWriter, errorCode, description string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":             errorCode,
		"error_description": description,
	})
}

// generateRandomString generates a random string of the specified length.
func generateRandomString(length int) string {
	b := make([]byte, length)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)[:length]
}

// GetToken returns a stored token by access or refresh token.
func (ms *MockOAuth2Server) GetToken(token string) (*OAuth2Token, bool) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	t, exists := ms.tokens[token]
	return t, exists
}

// IsTokenValid checks if a token is valid.
func (ms *MockOAuth2Server) IsTokenValid(token string) bool {
	ms.mu.RLock()
	defer ms.mu.RUnlock()
	tokenData, exists := ms.tokens[token]
	if !exists {
		return false
	}
	return time.Since(tokenData.IssuedAt) < ms.tokenLifetime
}

// AuthenticatedHandler wraps a handler to require valid authentication.
func (ms *MockOAuth2Server) AuthenticatedHandler(handler http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := r.Header.Get("Authorization")
		if auth == "" {
			http.Error(w, "Missing authorization header", http.StatusUnauthorized)
			return
		}

		parts := strings.SplitN(auth, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
			return
		}

		token := parts[1]
		if !ms.IsTokenValid(token) {
			http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
			return
		}

		handler(w, r)
	}
}
