package main

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"
)

// AuthServer handles OAuth2 authentication.
type AuthServer struct {
	mu            sync.RWMutex
	accessTokens  map[string]*TokenData
	refreshTokens map[string]*TokenData
	authCodes     map[string]*AuthCodeData
	tokenLifetime time.Duration
}

// TokenData stores token metadata.
type TokenData struct {
	Token     string
	ExpiresAt time.Time
	Scope     string
}

// AuthCodeData stores authorization code data.
type AuthCodeData struct {
	Code      string
	ExpiresAt time.Time
	Used      bool
}

// NewAuthServer creates a new AuthServer instance.
func NewAuthServer() *AuthServer {
	return &AuthServer{
		accessTokens:  make(map[string]*TokenData),
		refreshTokens: make(map[string]*TokenData),
		authCodes:     make(map[string]*AuthCodeData),
		tokenLifetime: 1 * time.Hour,
	}
}

// HandleToken handles the OAuth2 token endpoint.
func (a *AuthServer) HandleToken(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		a.sendError(w, "invalid_request", "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		a.sendError(w, "invalid_request", "Failed to parse form data", http.StatusBadRequest)
		return
	}

	grantType := r.FormValue("grant_type")

	switch grantType {
	case "authorization_code":
		a.handleAuthorizationCode(w, r)
	case "refresh_token":
		a.handleRefreshToken(w, r)
	default:
		a.sendError(w, "unsupported_grant_type", fmt.Sprintf("Grant type '%s' not supported", grantType), http.StatusBadRequest)
	}
}

// handleAuthorizationCode handles the authorization_code grant type.
func (a *AuthServer) handleAuthorizationCode(w http.ResponseWriter, r *http.Request) {
	code := r.FormValue("code")
	if code == "" {
		a.sendError(w, "invalid_request", "Missing code parameter", http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// For mock purposes, accept any code or validate if it exists
	authCode, exists := a.authCodes[code]
	if exists {
		// Validate code hasn't been used and hasn't expired
		if authCode.Used {
			a.sendError(w, "invalid_grant", "Authorization code already used", http.StatusBadRequest)
			return
		}
		if time.Now().After(authCode.ExpiresAt) {
			a.sendError(w, "invalid_grant", "Authorization code expired", http.StatusBadRequest)
			return
		}
		// Mark as used
		authCode.Used = true
	} else {
		// For mock server, generate a valid code on the fly for any code value
		// This allows easy testing without going through full OAuth flow
		authCode = &AuthCodeData{
			Code:      code,
			ExpiresAt: time.Now().Add(5 * time.Minute),
			Used:      true,
		}
		a.authCodes[code] = authCode
	}

	// Generate tokens
	accessToken := a.generateToken()
	refreshToken := a.generateToken()

	tokenData := &TokenData{
		Token:     accessToken,
		ExpiresAt: time.Now().Add(a.tokenLifetime),
		Scope:     "openid",
	}

	refreshData := &TokenData{
		Token:     refreshToken,
		ExpiresAt: time.Now().Add(24 * time.Hour), // Refresh tokens last longer
		Scope:     "openid",
	}

	a.accessTokens[accessToken] = tokenData
	a.refreshTokens[refreshToken] = refreshData

	// Send response
	response := TokenResponse{
		AccessToken:  accessToken,
		TokenType:    "Bearer",
		ExpiresIn:    int(a.tokenLifetime.Seconds()),
		RefreshToken: refreshToken,
		Scope:        "openid",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// handleRefreshToken handles the refresh_token grant type.
func (a *AuthServer) handleRefreshToken(w http.ResponseWriter, r *http.Request) {
	refreshToken := r.FormValue("refresh_token")
	if refreshToken == "" {
		a.sendError(w, "invalid_request", "Missing refresh_token parameter", http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()

	// Validate refresh token
	tokenData, exists := a.refreshTokens[refreshToken]
	if !exists {
		a.sendError(w, "invalid_grant", "Invalid refresh token", http.StatusBadRequest)
		return
	}

	if time.Now().After(tokenData.ExpiresAt) {
		delete(a.refreshTokens, refreshToken)
		a.sendError(w, "invalid_grant", "Refresh token expired", http.StatusBadRequest)
		return
	}

	// Generate new access token
	newAccessToken := a.generateToken()
	newTokenData := &TokenData{
		Token:     newAccessToken,
		ExpiresAt: time.Now().Add(a.tokenLifetime),
		Scope:     tokenData.Scope,
	}

	a.accessTokens[newAccessToken] = newTokenData

	// Send response (no new refresh token in this implementation)
	response := TokenResponse{
		AccessToken: newAccessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(a.tokenLifetime.Seconds()),
		Scope:       tokenData.Scope,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(response)
}

// ValidateToken validates a bearer token from the Authorization header.
func (a *AuthServer) ValidateToken(authHeader string) error {
	if authHeader == "" {
		return fmt.Errorf("missing authorization header")
	}

	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || parts[0] != "Bearer" {
		return fmt.Errorf("invalid authorization header format")
	}

	token := parts[1]

	a.mu.RLock()
	defer a.mu.RUnlock()

	tokenData, exists := a.accessTokens[token]
	if !exists {
		return fmt.Errorf("invalid token")
	}

	if time.Now().After(tokenData.ExpiresAt) {
		return fmt.Errorf("token expired")
	}

	return nil
}

// AuthMiddleware wraps handlers to require authentication.
func (a *AuthServer) AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		authHeader := r.Header.Get("Authorization")
		if err := a.ValidateToken(authHeader); err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnauthorized)
			_ = json.NewEncoder(w).Encode(ErrorResponse{
				Error:       "unauthorized",
				Description: err.Error(),
				Code:        http.StatusUnauthorized,
			})
			return
		}

		next(w, r)
	}
}

// generateToken generates a random token string.
func (a *AuthServer) generateToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.URLEncoding.EncodeToString(b)
}

// sendError sends an OAuth2 error response.
func (a *AuthServer) sendError(w http.ResponseWriter, errorCode, description string, statusCode int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"error":             errorCode,
		"error_description": description,
	})
}

// GenerateAuthCode creates a new authorization code (for testing purposes).
func (a *AuthServer) GenerateAuthCode() string {
	a.mu.Lock()
	defer a.mu.Unlock()

	code := a.generateToken()[:16] // Shorter code
	a.authCodes[code] = &AuthCodeData{
		Code:      code,
		ExpiresAt: time.Now().Add(5 * time.Minute),
		Used:      false,
	}

	return code
}
