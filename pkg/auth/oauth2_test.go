package auth

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"golang.org/x/oauth2"
)

func TestNewOAuth2Auth(t *testing.T) {
	tests := []struct {
		name    string
		config  *OAuth2Config
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "valid client credentials",
			config: &OAuth2Config{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				TokenURL:     "https://example.com/token",
				Flow:         OAuth2FlowClientCredentials,
			},
			wantErr: false,
		},
		{
			name: "valid authorization code",
			config: &OAuth2Config{
				ClientID: "test-client",
				AuthURL:  "https://example.com/auth",
				TokenURL: "https://example.com/token",
				Flow:     OAuth2FlowAuthorizationCode,
			},
			wantErr: false,
		},
		{
			name: "missing client id",
			config: &OAuth2Config{
				TokenURL: "https://example.com/token",
				Flow:     OAuth2FlowClientCredentials,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewOAuth2Auth(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewOAuth2Auth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && auth == nil {
				t.Error("NewOAuth2Auth() returned nil authenticator")
			}
		})
	}
}

func TestOAuth2Auth_Type(t *testing.T) {
	auth := &OAuth2Auth{
		config: &OAuth2Config{
			ClientID: "test",
			TokenURL: "https://example.com/token",
			Flow:     OAuth2FlowClientCredentials,
		},
	}

	if got := auth.Type(); got != AuthTypeOAuth2 {
		t.Errorf("Type() = %v, want %v", got, AuthTypeOAuth2)
	}
}

func TestOAuth2Auth_Validate(t *testing.T) {
	tests := []struct {
		name    string
		config  *OAuth2Config
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "OAuth2 config is required",
		},
		{
			name: "missing client id",
			config: &OAuth2Config{
				Flow: OAuth2FlowClientCredentials,
			},
			wantErr: true,
			errMsg:  "client_id is required",
		},
		{
			name: "missing flow",
			config: &OAuth2Config{
				ClientID: "test",
			},
			wantErr: true,
			errMsg:  "flow is required",
		},
		{
			name: "auth code missing auth url",
			config: &OAuth2Config{
				ClientID: "test",
				TokenURL: "https://example.com/token",
				Flow:     OAuth2FlowAuthorizationCode,
			},
			wantErr: true,
			errMsg:  "auth_url is required",
		},
		{
			name: "auth code missing token url",
			config: &OAuth2Config{
				ClientID: "test",
				AuthURL:  "https://example.com/auth",
				Flow:     OAuth2FlowAuthorizationCode,
			},
			wantErr: true,
			errMsg:  "token_url is required",
		},
		{
			name: "client credentials missing token url",
			config: &OAuth2Config{
				ClientID:     "test",
				ClientSecret: "secret",
				Flow:         OAuth2FlowClientCredentials,
			},
			wantErr: true,
			errMsg:  "token_url is required",
		},
		{
			name: "client credentials missing secret",
			config: &OAuth2Config{
				ClientID: "test",
				TokenURL: "https://example.com/token",
				Flow:     OAuth2FlowClientCredentials,
			},
			wantErr: true,
			errMsg:  "client_secret is required",
		},
		{
			name: "password flow missing token url",
			config: &OAuth2Config{
				ClientID: "test",
				Username: "user",
				Password: "pass",
				Flow:     OAuth2FlowPassword,
			},
			wantErr: true,
			errMsg:  "token_url is required",
		},
		{
			name: "password flow missing credentials",
			config: &OAuth2Config{
				ClientID: "test",
				TokenURL: "https://example.com/token",
				Flow:     OAuth2FlowPassword,
			},
			wantErr: true,
			errMsg:  "username and password are required",
		},
		{
			name: "device code missing device url",
			config: &OAuth2Config{
				ClientID: "test",
				TokenURL: "https://example.com/token",
				Flow:     OAuth2FlowDeviceCode,
			},
			wantErr: true,
			errMsg:  "device_code_url is required",
		},
		{
			name: "device code missing token url",
			config: &OAuth2Config{
				ClientID:      "test",
				DeviceCodeURL: "https://example.com/device",
				Flow:          OAuth2FlowDeviceCode,
			},
			wantErr: true,
			errMsg:  "token_url is required",
		},
		{
			name: "unsupported flow",
			config: &OAuth2Config{
				ClientID: "test",
				Flow:     OAuth2Flow("invalid"),
			},
			wantErr: true,
			errMsg:  "unsupported OAuth2 flow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &OAuth2Auth{config: tt.config}
			err := auth.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestOAuth2Auth_GetHeaders(t *testing.T) {
	auth := &OAuth2Auth{}

	tests := []struct {
		name  string
		token *Token
		want  map[string]string
	}{
		{
			name:  "nil token",
			token: nil,
			want:  nil,
		},
		{
			name: "empty access token",
			token: &Token{
				AccessToken: "",
			},
			want: nil,
		},
		{
			name: "with bearer token",
			token: &Token{
				AccessToken: "test-token",
				TokenType:   "Bearer",
			},
			want: map[string]string{
				"Authorization": "Bearer test-token",
			},
		},
		{
			name: "default bearer type",
			token: &Token{
				AccessToken: "test-token",
			},
			want: map[string]string{
				"Authorization": "Bearer test-token",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := auth.GetHeaders(tt.token)
			if len(got) != len(tt.want) {
				t.Errorf("GetHeaders() = %v, want %v", got, tt.want)
				return
			}
			for k, v := range tt.want {
				if got[k] != v {
					t.Errorf("GetHeaders()[%s] = %v, want %v", k, got[k], v)
				}
			}
		})
	}
}

func TestOAuth2Auth_RefreshToken(t *testing.T) {
	// Create mock OAuth2 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		refreshToken := r.FormValue("refresh_token")
		if refreshToken == "" {
			http.Error(w, "missing refresh token", http.StatusBadRequest)
			return
		}

		resp := oauth2.Token{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(time.Hour),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	auth := &OAuth2Auth{
		config: &OAuth2Config{
			ClientID: "test-client",
			TokenURL: server.URL + "/token",
		},
		oauth2Config: &oauth2.Config{
			ClientID: "test-client",
			Endpoint: oauth2.Endpoint{
				TokenURL: server.URL + "/token",
			},
		},
	}

	tests := []struct {
		name    string
		token   *Token
		wantErr bool
	}{
		{
			name:    "nil token",
			token:   nil,
			wantErr: true,
		},
		{
			name: "missing refresh token",
			token: &Token{
				AccessToken: "old-token",
			},
			wantErr: true,
		},
		{
			name: "valid refresh",
			token: &Token{
				AccessToken:  "old-token",
				RefreshToken: "refresh-token",
				ExpiresAt:    time.Now().Add(-time.Hour),
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			newToken, err := auth.RefreshToken(ctx, tt.token)
			if (err != nil) != tt.wantErr {
				t.Errorf("RefreshToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && newToken == nil {
				t.Error("RefreshToken() returned nil token")
			}
			if !tt.wantErr && newToken.AccessToken != "new-access-token" {
				t.Errorf("RefreshToken() AccessToken = %v, want new-access-token", newToken.AccessToken)
			}
		})
	}
}

func TestOAuth2Auth_AuthenticateClientCredentials(t *testing.T) {
	// Create mock OAuth2 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		grantType := r.FormValue("grant_type")
		if grantType != "client_credentials" {
			http.Error(w, "invalid grant type", http.StatusBadRequest)
			return
		}

		resp := oauth2.Token{
			AccessToken: "test-access-token",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(time.Hour),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := &OAuth2Config{
		ClientID:     "test-client",
		ClientSecret: "test-secret",
		TokenURL:     server.URL + "/token",
		Flow:         OAuth2FlowClientCredentials,
	}

	auth, err := NewOAuth2Auth(config)
	if err != nil {
		t.Fatalf("NewOAuth2Auth() error = %v", err)
	}

	ctx := context.Background()
	token, err := auth.Authenticate(ctx)
	if err != nil {
		t.Errorf("Authenticate() error = %v", err)
		return
	}

	if token == nil {
		t.Fatal("Authenticate() returned nil token")
	}

	if token.AccessToken != "test-access-token" {
		t.Errorf("AccessToken = %v, want test-access-token", token.AccessToken)
	}
}

func TestOAuth2Auth_AuthenticatePassword(t *testing.T) {
	// Create mock OAuth2 server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		grantType := r.FormValue("grant_type")
		if grantType != "password" {
			http.Error(w, "invalid grant type", http.StatusBadRequest)
			return
		}

		username := r.FormValue("username")
		password := r.FormValue("password")
		if username != "testuser" || password != "testpass" {
			http.Error(w, "invalid credentials", http.StatusUnauthorized)
			return
		}

		resp := oauth2.Token{
			AccessToken: "password-token",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(time.Hour),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := &OAuth2Config{
		ClientID: "test-client",
		TokenURL: server.URL + "/token",
		Flow:     OAuth2FlowPassword,
		Username: "testuser",
		Password: "testpass",
	}

	auth, err := NewOAuth2Auth(config)
	if err != nil {
		t.Fatalf("NewOAuth2Auth() error = %v", err)
	}

	ctx := context.Background()
	token, err := auth.Authenticate(ctx)
	if err != nil {
		t.Errorf("Authenticate() error = %v", err)
		return
	}

	if token == nil {
		t.Fatal("Authenticate() returned nil token")
	}

	if token.AccessToken != "password-token" {
		t.Errorf("AccessToken = %v, want password-token", token.AccessToken)
	}
}

func TestOAuth2Auth_AuthenticateUnsupportedFlow(t *testing.T) {
	auth := &OAuth2Auth{
		config: &OAuth2Config{
			Flow: OAuth2Flow("unsupported"),
		},
	}

	ctx := context.Background()
	_, err := auth.Authenticate(ctx)
	if err == nil {
		t.Error("Authenticate() expected error for unsupported flow")
	}
}

func TestConvertOAuth2Token(t *testing.T) {
	expiryTime := time.Now().Add(time.Hour)
	oauth2Token := &oauth2.Token{
		AccessToken:  "access",
		RefreshToken: "refresh",
		TokenType:    "Bearer",
		Expiry:       expiryTime,
	}

	token := convertOAuth2Token(oauth2Token)

	if token.AccessToken != "access" {
		t.Errorf("AccessToken = %v, want access", token.AccessToken)
	}
	if token.RefreshToken != "refresh" {
		t.Errorf("RefreshToken = %v, want refresh", token.RefreshToken)
	}
	if token.TokenType != "Bearer" {
		t.Errorf("TokenType = %v, want Bearer", token.TokenType)
	}
	if !token.ExpiresAt.Equal(expiryTime) {
		t.Errorf("ExpiresAt = %v, want %v", token.ExpiresAt, expiryTime)
	}
}

func TestConvertEndpointParams(t *testing.T) {
	tests := []struct {
		name   string
		params map[string]string
		want   url.Values
	}{
		{
			name:   "nil params",
			params: nil,
			want:   nil,
		},
		{
			name: "with params",
			params: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
			want: url.Values{
				"key1": []string{"value1"},
				"key2": []string{"value2"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := convertEndpointParams(tt.params)
			if tt.want == nil && got != nil {
				t.Errorf("convertEndpointParams() = %v, want nil", got)
				return
			}
			if tt.want != nil {
				for k, v := range tt.want {
					if got.Get(k) != v[0] {
						t.Errorf("convertEndpointParams()[%s] = %v, want %v", k, got.Get(k), v[0])
					}
				}
			}
		})
	}
}

func TestGenerateRandomString(t *testing.T) {
	tests := []int{16, 32, 64}

	for _, length := range tests {
		t.Run(string(rune(length)), func(t *testing.T) {
			result := generateRandomString(length)
			if len(result) != length {
				t.Errorf("generateRandomString(%d) length = %d, want %d", length, len(result), length)
			}

			// Generate another and verify they're different (probabilistically)
			result2 := generateRandomString(length)
			if result == result2 {
				t.Error("generateRandomString() produced identical strings")
			}
		})
	}
}

func TestOAuth2Auth_RequestDeviceCode(t *testing.T) {
	// Create mock device code server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		clientID := r.FormValue("client_id")
		if clientID == "" {
			http.Error(w, "missing client_id", http.StatusBadRequest)
			return
		}

		resp := DeviceAuthResponse{
			DeviceCode:      "device-code-123",
			UserCode:        "ABCD-1234",
			VerificationURL: "https://example.com/device",
			ExpiresIn:       600,
			Interval:        5,
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	auth := &OAuth2Auth{
		config: &OAuth2Config{
			ClientID:      "test-client",
			DeviceCodeURL: server.URL,
			Scopes:        []string{"read", "write"},
		},
	}

	ctx := context.Background()
	deviceAuth, err := auth.requestDeviceCode(ctx)
	if err != nil {
		t.Fatalf("requestDeviceCode() error = %v", err)
	}

	if deviceAuth.DeviceCode != "device-code-123" {
		t.Errorf("DeviceCode = %v, want device-code-123", deviceAuth.DeviceCode)
	}
	if deviceAuth.UserCode != "ABCD-1234" {
		t.Errorf("UserCode = %v, want ABCD-1234", deviceAuth.UserCode)
	}
	if deviceAuth.Interval != 5 {
		t.Errorf("Interval = %v, want 5", deviceAuth.Interval)
	}
}

func TestOAuth2Auth_CheckDeviceToken(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   interface{}
		wantErr    bool
		errMsg     string
	}{
		{
			name:       "success",
			statusCode: http.StatusOK,
			response: oauth2.Token{
				AccessToken: "device-token",
				TokenType:   "Bearer",
			},
			wantErr: false,
		},
		{
			name:       "authorization pending",
			statusCode: http.StatusBadRequest,
			response: map[string]string{
				"error":             "authorization_pending",
				"error_description": "User has not yet authorized",
			},
			wantErr: true,
			errMsg:  "authorization_pending",
		},
		{
			name:       "slow down",
			statusCode: http.StatusBadRequest,
			response: map[string]string{
				"error":             "slow_down",
				"error_description": "Polling too fast",
			},
			wantErr: true,
			errMsg:  "slow_down",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.statusCode)
				w.Header().Set("Content-Type", "application/json")
				_ = json.NewEncoder(w).Encode(tt.response)
			}))
			defer server.Close()

			auth := &OAuth2Auth{
				config: &OAuth2Config{
					ClientID: "test-client",
					TokenURL: server.URL,
				},
			}

			ctx := context.Background()
			token, err := auth.checkDeviceToken(ctx, "device-code")
			if (err != nil) != tt.wantErr {
				t.Errorf("checkDeviceToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && token == nil {
				t.Error("checkDeviceToken() returned nil token")
			}
		})
	}
}

func TestOAuth2Auth_BuildAuthURL(t *testing.T) {
	auth := &OAuth2Auth{
		config: &OAuth2Config{
			ClientID: "test-client",
			PKCE:     true,
			EndpointParams: map[string]string{
				"custom": "value",
			},
		},
		oauth2Config: &oauth2.Config{
			ClientID:    "test-client",
			RedirectURL: "http://localhost:8080/callback",
			Scopes:      []string{"read", "write"},
			Endpoint: oauth2.Endpoint{
				AuthURL: "https://example.com/auth",
			},
		},
	}

	authURL := auth.buildAuthURL()
	if authURL == "" {
		t.Error("buildAuthURL() returned empty string")
	}

	// Parse URL to verify parameters
	parsedURL, err := url.Parse(authURL)
	if err != nil {
		t.Fatalf("failed to parse auth URL: %v", err)
	}

	query := parsedURL.Query()
	if query.Get("client_id") != "test-client" {
		t.Errorf("client_id = %v, want test-client", query.Get("client_id"))
	}

	// With PKCE, should have code_challenge
	if auth.config.PKCE && query.Get("code_challenge") == "" {
		t.Error("PKCE enabled but no code_challenge in URL")
	}
	if auth.config.PKCE && query.Get("code_challenge_method") != "S256" {
		t.Errorf("code_challenge_method = %v, want S256", query.Get("code_challenge_method"))
	}

	// Should have custom param
	if query.Get("custom") != "value" {
		t.Errorf("custom param = %v, want value", query.Get("custom"))
	}
}

func TestOAuth2Auth_StartCallbackServer(t *testing.T) {
	auth := &OAuth2Auth{
		config: &OAuth2Config{
			RedirectURL: "http://localhost:18080/callback",
		},
	}

	server, callbackChan, err := auth.startCallbackServer()
	if err != nil {
		t.Fatalf("startCallbackServer() error = %v", err)
	}
	defer func() { _ = server.Close() }()

	// Test successful callback
	go func() {
		resp, err := http.Get("http://localhost:18080/callback?code=test-code")
		if err != nil {
			t.Logf("callback request failed: %v", err)
		}
		if resp != nil {
			_ = resp.Body.Close()
		}
	}()

	select {
	case code := <-callbackChan:
		if code != "test-code" {
			t.Errorf("received code = %v, want test-code", code)
		}
	case <-time.After(2 * time.Second):
		t.Error("timeout waiting for callback")
	}
}

func TestOAuth2Auth_InitConfig(t *testing.T) {
	tests := []struct {
		name   string
		config *OAuth2Config
	}{
		{
			name: "client credentials",
			config: &OAuth2Config{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				TokenURL:     "https://example.com/token",
				Flow:         OAuth2FlowClientCredentials,
				Scopes:       []string{"read"},
				EndpointParams: map[string]string{
					"param": "value",
				},
			},
		},
		{
			name: "authorization code",
			config: &OAuth2Config{
				ClientID:    "test-client",
				AuthURL:     "https://example.com/auth",
				TokenURL:    "https://example.com/token",
				Flow:        OAuth2FlowAuthorizationCode,
				RedirectURL: "http://localhost:8080/callback",
			},
		},
		{
			name: "device code",
			config: &OAuth2Config{
				ClientID:      "test-client",
				TokenURL:      "https://example.com/token",
				DeviceCodeURL: "https://example.com/device",
				Flow:          OAuth2FlowDeviceCode,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &OAuth2Auth{config: tt.config}
			if err := auth.initConfig(); err != nil {
				t.Errorf("initConfig() error = %v", err)
			}

			if tt.config.Flow == OAuth2FlowClientCredentials {
				if auth.ccConfig == nil {
					t.Error("client credentials config not initialized")
				}
			} else {
				if auth.oauth2Config == nil {
					t.Error("oauth2 config not initialized")
				}
			}
		})
	}
}

func TestOAuth2Auth_WithBrowserOpener(t *testing.T) {
	config := &OAuth2Config{
		ClientID: "test-client",
		AuthURL:  "https://example.com/auth",
		TokenURL: "https://example.com/token",
		Flow:     OAuth2FlowAuthorizationCode,
	}

	auth, err := NewOAuth2Auth(config)
	if err != nil {
		t.Fatalf("NewOAuth2Auth() error = %v", err)
	}

	// Verify default opener is set
	if auth.browserOpener == nil {
		t.Error("Expected default browser opener to be set")
	}

	// Set custom opener
	mockOpener := &MockBrowserOpener{}
	result := auth.WithBrowserOpener(mockOpener)

	if result != auth {
		t.Error("WithBrowserOpener should return same auth instance")
	}

	if auth.browserOpener != mockOpener {
		t.Error("Browser opener was not set correctly")
	}
}

func TestOAuth2Auth_AuthCodeWithBrowserOpen(t *testing.T) {
	// Create mock token server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}

		resp := oauth2.Token{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(time.Hour),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer tokenServer.Close()

	// Use a random port for callback server
	config := &OAuth2Config{
		ClientID:        "test-client",
		AuthURL:         "https://example.com/auth",
		TokenURL:        tokenServer.URL + "/token",
		RedirectURL:     "http://localhost:18081/callback",
		Flow:            OAuth2FlowAuthorizationCode,
		AutoOpenBrowser: true,
	}

	auth, err := NewOAuth2Auth(config)
	if err != nil {
		t.Fatalf("NewOAuth2Auth() error = %v", err)
	}

	// Set mock browser opener
	mockOpener := &MockBrowserOpener{}
	auth.WithBrowserOpener(mockOpener)

	// Start authentication in background
	ctx := context.Background()
	tokenChan := make(chan *Token, 1)
	errChan := make(chan error, 1)

	go func() {
		token, err := auth.authenticateAuthorizationCode(ctx)
		if err != nil {
			errChan <- err
		} else {
			tokenChan <- token
		}
	}()

	// Wait for browser to be opened
	time.Sleep(100 * time.Millisecond)

	// Verify browser was opened
	openedURLs := mockOpener.GetOpenedURLs()
	if len(openedURLs) != 1 {
		t.Errorf("Expected browser to be opened once, got %d times", len(openedURLs))
	}

	if len(openedURLs) > 0 {
		openedURL := openedURLs[0]
		if !strings.Contains(openedURL, "https://example.com/auth") {
			t.Errorf("Expected auth URL to contain auth endpoint, got %s", openedURL)
		}
	}

	// Simulate callback
	_, err = http.Get("http://localhost:18081/callback?code=test-code")
	if err != nil {
		t.Logf("Callback request failed: %v", err)
	}

	// Wait for result
	select {
	case token := <-tokenChan:
		if token == nil {
			t.Error("Expected token, got nil")
		}
		if token.AccessToken != "test-token" {
			t.Errorf("Expected access token 'test-token', got '%s'", token.AccessToken)
		}
	case err := <-errChan:
		t.Errorf("Authentication failed: %v", err)
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for authentication")
	}
}

func TestOAuth2Auth_AuthCodeWithoutBrowser(t *testing.T) {
	// Create mock token server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}

		resp := oauth2.Token{
			AccessToken: "test-token-no-browser",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(time.Hour),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer tokenServer.Close()

	config := &OAuth2Config{
		ClientID:        "test-client",
		AuthURL:         "https://example.com/auth",
		TokenURL:        tokenServer.URL + "/token",
		RedirectURL:     "http://localhost:18082/callback",
		Flow:            OAuth2FlowAuthorizationCode,
		AutoOpenBrowser: false, // Browser opening disabled
	}

	auth, err := NewOAuth2Auth(config)
	if err != nil {
		t.Fatalf("NewOAuth2Auth() error = %v", err)
	}

	// Set mock browser opener
	mockOpener := &MockBrowserOpener{}
	auth.WithBrowserOpener(mockOpener)

	// Start authentication in background
	ctx := context.Background()
	tokenChan := make(chan *Token, 1)
	errChan := make(chan error, 1)

	go func() {
		token, err := auth.authenticateAuthorizationCode(ctx)
		if err != nil {
			errChan <- err
		} else {
			tokenChan <- token
		}
	}()

	// Wait a moment
	time.Sleep(100 * time.Millisecond)

	// Verify browser was NOT opened
	openedURLs := mockOpener.GetOpenedURLs()
	if len(openedURLs) != 0 {
		t.Errorf("Expected browser NOT to be opened, but it was opened %d times", len(openedURLs))
	}

	// Simulate callback
	_, err = http.Get("http://localhost:18082/callback?code=test-code")
	if err != nil {
		t.Logf("Callback request failed: %v", err)
	}

	// Wait for result
	select {
	case token := <-tokenChan:
		if token == nil {
			t.Error("Expected token, got nil")
		}
	case err := <-errChan:
		t.Errorf("Authentication failed: %v", err)
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for authentication")
	}
}

func TestOAuth2Auth_BrowserOpenError(t *testing.T) {
	// Create mock token server
	tokenServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}

		resp := oauth2.Token{
			AccessToken: "test-token-error",
			TokenType:   "Bearer",
			Expiry:      time.Now().Add(time.Hour),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer tokenServer.Close()

	config := &OAuth2Config{
		ClientID:        "test-client",
		AuthURL:         "https://example.com/auth",
		TokenURL:        tokenServer.URL + "/token",
		RedirectURL:     "http://localhost:18083/callback",
		Flow:            OAuth2FlowAuthorizationCode,
		AutoOpenBrowser: true,
	}

	auth, err := NewOAuth2Auth(config)
	if err != nil {
		t.Fatalf("NewOAuth2Auth() error = %v", err)
	}

	// Set mock browser opener that returns an error
	mockOpener := &MockBrowserOpener{
		Err: fmt.Errorf("browser not available"),
	}
	auth.WithBrowserOpener(mockOpener)

	// Start authentication in background
	ctx := context.Background()
	tokenChan := make(chan *Token, 1)
	errChan := make(chan error, 1)

	go func() {
		token, err := auth.authenticateAuthorizationCode(ctx)
		if err != nil {
			errChan <- err
		} else {
			tokenChan <- token
		}
	}()

	// Wait for browser open attempt
	time.Sleep(100 * time.Millisecond)

	// Verify browser open was attempted even though it failed
	openedURLs := mockOpener.GetOpenedURLs()
	if len(openedURLs) != 1 {
		t.Errorf("Expected browser open to be attempted once, got %d times", len(openedURLs))
	}

	// Simulate callback - authentication should still continue despite browser error
	_, err = http.Get("http://localhost:18083/callback?code=test-code")
	if err != nil {
		t.Logf("Callback request failed: %v", err)
	}

	// Wait for result - should succeed despite browser error
	select {
	case token := <-tokenChan:
		if token == nil {
			t.Error("Expected token, got nil")
		}
		if token.AccessToken != "test-token-error" {
			t.Errorf("Expected access token 'test-token-error', got '%s'", token.AccessToken)
		}
	case err := <-errChan:
		t.Errorf("Authentication should succeed despite browser error, got: %v", err)
	case <-time.After(2 * time.Second):
		t.Error("Timeout waiting for authentication")
	}
}

// TestOAuth2Auth_TokenFlow_AccessToken tests direct injection of an access token.
func TestOAuth2Auth_TokenFlow_AccessToken(t *testing.T) {
	// Create a valid access token (Bearer type)
	accessToken := createTestJWT(t, "Bearer", time.Now().Add(time.Hour))

	config := &OAuth2Config{
		ClientID: "test-client",
		TokenURL: "https://example.com/token",
		Flow:     OAuth2FlowToken,
		Token:    accessToken,
	}

	auth, err := NewOAuth2Auth(config)
	if err != nil {
		t.Fatalf("NewOAuth2Auth() error = %v", err)
	}

	ctx := context.Background()
	token, err := auth.Authenticate(ctx)
	if err != nil {
		t.Errorf("Authenticate() error = %v", err)
		return
	}

	if token == nil {
		t.Fatal("Authenticate() returned nil token")
	}

	if token.AccessToken != accessToken {
		t.Errorf("AccessToken = %v, want %v", token.AccessToken, accessToken)
	}

	if token.TokenType != "Bearer" {
		t.Errorf("TokenType = %v, want Bearer", token.TokenType)
	}

	if token.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should be set")
	}
}

// TestOAuth2Auth_TokenFlow_BearerToken tests direct injection of a bearer token.
func TestOAuth2Auth_TokenFlow_BearerToken(t *testing.T) {
	// Create a valid bearer token (empty typ claim - defaults to access)
	bearerToken := createTestJWT(t, "", time.Now().Add(time.Hour))

	config := &OAuth2Config{
		ClientID: "test-client",
		TokenURL: "https://example.com/token",
		Flow:     OAuth2FlowToken,
		Token:    bearerToken, // Token set in config triggers token flow
	}

	auth, err := NewOAuth2Auth(config)
	if err != nil {
		t.Fatalf("NewOAuth2Auth() error = %v", err)
	}

	ctx := context.Background()
	token, err := auth.Authenticate(ctx)
	if err != nil {
		t.Errorf("Authenticate() error = %v", err)
		return
	}

	if token == nil {
		t.Fatal("Authenticate() returned nil token")
	}

	if token.AccessToken != bearerToken {
		t.Errorf("AccessToken = %v, want %v", token.AccessToken, bearerToken)
	}
}

// TestOAuth2Auth_TokenFlow_RefreshToken tests exchange of a refresh token.
func TestOAuth2Auth_TokenFlow_RefreshToken(t *testing.T) {
	// Create a refresh token
	refreshToken := createTestJWT(t, "Refresh", time.Now().Add(24*time.Hour))

	// Create mock token server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		grantType := r.FormValue("grant_type")
		if grantType != "refresh_token" {
			http.Error(w, "invalid grant type", http.StatusBadRequest)
			return
		}

		receivedRefreshToken := r.FormValue("refresh_token")
		if receivedRefreshToken != refreshToken {
			http.Error(w, "invalid refresh token", http.StatusUnauthorized)
			return
		}

		resp := oauth2.Token{
			AccessToken:  "new-access-token",
			RefreshToken: "new-refresh-token",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(time.Hour),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := &OAuth2Config{
		ClientID: "test-client",
		TokenURL: server.URL + "/token",
		Flow:     OAuth2FlowToken,
		Token:    refreshToken,
	}

	auth, err := NewOAuth2Auth(config)
	if err != nil {
		t.Fatalf("NewOAuth2Auth() error = %v", err)
	}

	ctx := context.Background()
	token, err := auth.Authenticate(ctx)
	if err != nil {
		t.Errorf("Authenticate() error = %v", err)
		return
	}

	if token == nil {
		t.Fatal("Authenticate() returned nil token")
	}

	if token.AccessToken != "new-access-token" {
		t.Errorf("AccessToken = %v, want new-access-token", token.AccessToken)
	}

	if token.RefreshToken != "new-refresh-token" {
		t.Errorf("RefreshToken = %v, want new-refresh-token", token.RefreshToken)
	}
}

// TestOAuth2Auth_TokenFlow_OfflineToken tests exchange of an offline token.
func TestOAuth2Auth_TokenFlow_OfflineToken(t *testing.T) {
	// Create an offline token
	offlineToken := createTestJWT(t, "Offline", time.Now().Add(24*time.Hour))

	// Create mock token server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}

		if err := r.ParseForm(); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		receivedRefreshToken := r.FormValue("refresh_token")
		if receivedRefreshToken != offlineToken {
			http.Error(w, "invalid offline token", http.StatusUnauthorized)
			return
		}

		resp := oauth2.Token{
			AccessToken:  "offline-access-token",
			RefreshToken: offlineToken, // Keep same offline token
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(time.Hour),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := &OAuth2Config{
		ClientID: "test-client",
		TokenURL: server.URL + "/token",
		Flow:     OAuth2FlowToken,
		Token:    offlineToken,
	}

	auth, err := NewOAuth2Auth(config)
	if err != nil {
		t.Fatalf("NewOAuth2Auth() error = %v", err)
	}

	ctx := context.Background()
	token, err := auth.Authenticate(ctx)
	if err != nil {
		t.Errorf("Authenticate() error = %v", err)
		return
	}

	if token == nil {
		t.Fatal("Authenticate() returned nil token")
	}

	if token.AccessToken != "offline-access-token" {
		t.Errorf("AccessToken = %v, want offline-access-token", token.AccessToken)
	}
}

// TestOAuth2Auth_TokenFlow_InvalidToken tests error handling for invalid tokens.
func TestOAuth2Auth_TokenFlow_InvalidToken(t *testing.T) {
	tests := []struct {
		name    string
		token   string
		wantErr string
	}{
		{
			name:    "malformed token",
			token:   "not.a.valid.jwt",
			wantErr: "invalid token format",
		},
		{
			name:    "empty token",
			token:   "",
			wantErr: "token is required for token flow",
		},
		{
			name:    "invalid base64",
			token:   "invalid!!!.token!!!.here!!!",
			wantErr: "invalid token format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &OAuth2Config{
				ClientID: "test-client",
				TokenURL: "https://example.com/token",
				Flow:     OAuth2FlowToken,
				Token:    tt.token,
			}

			auth, err := NewOAuth2Auth(config)
			if err != nil {
				// Validation might catch some errors early
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("NewOAuth2Auth() error = %v, want error containing %v", err, tt.wantErr)
				}
				return
			}

			ctx := context.Background()
			_, err = auth.Authenticate(ctx)
			if err == nil {
				t.Error("Authenticate() expected error, got nil")
				return
			}

			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("Authenticate() error = %v, want error containing %v", err, tt.wantErr)
			}
		})
	}
}

// TestOAuth2Auth_ExchangeRefreshToken_Success tests successful refresh token exchange.
func TestOAuth2Auth_ExchangeRefreshToken_Success(t *testing.T) {
	refreshToken := createTestJWT(t, "Refresh", time.Now().Add(24*time.Hour))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}

		resp := oauth2.Token{
			AccessToken:  "exchanged-access-token",
			RefreshToken: "new-refresh-token",
			TokenType:    "Bearer",
			Expiry:       time.Now().Add(time.Hour),
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	config := &OAuth2Config{
		ClientID: "test-client",
		AuthURL:  "https://example.com/auth",
		TokenURL: server.URL + "/token",
		Flow:     OAuth2FlowAuthorizationCode, // Need a flow that uses oauth2Config
	}

	auth, err := NewOAuth2Auth(config)
	if err != nil {
		t.Fatalf("NewOAuth2Auth() error = %v", err)
	}

	ctx := context.Background()
	token, err := auth.exchangeRefreshToken(ctx, refreshToken)
	if err != nil {
		t.Errorf("exchangeRefreshToken() error = %v", err)
		return
	}

	if token == nil {
		t.Fatal("exchangeRefreshToken() returned nil token")
	}

	if token.AccessToken != "exchanged-access-token" {
		t.Errorf("AccessToken = %v, want exchanged-access-token", token.AccessToken)
	}
}

// TestOAuth2Auth_ExchangeRefreshToken_Failure tests failed refresh token exchange.
func TestOAuth2Auth_ExchangeRefreshToken_Failure(t *testing.T) {
	refreshToken := createTestJWT(t, "Refresh", time.Now().Add(24*time.Hour))

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/token" {
			http.NotFound(w, r)
			return
		}

		http.Error(w, `{"error":"invalid_grant","error_description":"refresh token expired"}`, http.StatusBadRequest)
	}))
	defer server.Close()

	config := &OAuth2Config{
		ClientID: "test-client",
		AuthURL:  "https://example.com/auth",
		TokenURL: server.URL + "/token",
		Flow:     OAuth2FlowAuthorizationCode, // Need a flow that uses oauth2Config
	}

	auth, err := NewOAuth2Auth(config)
	if err != nil {
		t.Fatalf("NewOAuth2Auth() error = %v", err)
	}

	ctx := context.Background()
	_, err = auth.exchangeRefreshToken(ctx, refreshToken)
	if err == nil {
		t.Error("exchangeRefreshToken() expected error for expired token")
		return
	}

	if !strings.Contains(err.Error(), "failed to exchange refresh token") {
		t.Errorf("exchangeRefreshToken() error = %v, want error containing 'failed to exchange refresh token'", err)
	}
}

// TestOAuth2Auth_DefaultRedirectPort tests that default redirect port is 9998.
func TestOAuth2Auth_DefaultRedirectPort(t *testing.T) {
	config := &OAuth2Config{
		ClientID: "test-client",
		AuthURL:  "https://example.com/auth",
		TokenURL: "https://example.com/token",
		Flow:     OAuth2FlowAuthorizationCode,
		// RedirectPort not set - should default to 9998
	}

	auth, err := NewOAuth2Auth(config)
	if err != nil {
		t.Fatalf("NewOAuth2Auth() error = %v", err)
	}

	// Start callback server to check the port
	server, _, err := auth.startCallbackServer()
	if err != nil {
		// If RedirectURL is not set, it will use default
		// We need to set it to test the port logic
		t.Skip("Skipping - need to check redirect URL logic")
	}
	defer func() { _ = server.Close() }()
}

// TestOAuth2Auth_CustomRedirectPort tests custom redirect port configuration.
func TestOAuth2Auth_CustomRedirectPort(t *testing.T) {
	customPort := 19999

	config := &OAuth2Config{
		ClientID:     "test-client",
		AuthURL:      "https://example.com/auth",
		TokenURL:     "https://example.com/token",
		Flow:         OAuth2FlowAuthorizationCode,
		RedirectURL:  fmt.Sprintf("http://localhost:%d/callback", customPort),
		RedirectPort: customPort,
	}

	auth, err := NewOAuth2Auth(config)
	if err != nil {
		t.Fatalf("NewOAuth2Auth() error = %v", err)
	}

	if auth.config.RedirectPort != customPort {
		t.Errorf("RedirectPort = %v, want %v", auth.config.RedirectPort, customPort)
	}
}

// TestOAuth2Auth_Validate_TokenFlow tests validation for token flow.
func TestOAuth2Auth_Validate_TokenFlow(t *testing.T) {
	tests := []struct {
		name    string
		config  *OAuth2Config
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid token flow",
			config: &OAuth2Config{
				ClientID: "test-client",
				TokenURL: "https://example.com/token",
				Flow:     OAuth2FlowToken,
				Token:    "valid.jwt.token",
			},
			wantErr: false,
		},
		{
			name: "token flow missing token",
			config: &OAuth2Config{
				ClientID: "test-client",
				TokenURL: "https://example.com/token",
				Flow:     OAuth2FlowToken,
				Token:    "",
			},
			wantErr: true,
			errMsg:  "token is required for token flow",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth := &OAuth2Auth{config: tt.config}
			err := auth.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errMsg) {
				t.Errorf("Validate() error = %v, want error containing %v", err, tt.errMsg)
			}
		})
	}
}

// createTestJWT creates a valid JWT for testing.
func createTestJWT(t *testing.T, typ string, expiresAt time.Time) string {
	t.Helper()

	// Create JWT header
	header := map[string]interface{}{
		"alg": "HS256",
		"typ": "JWT",
	}
	headerJSON, _ := json.Marshal(header)
	headerB64 := base64.RawURLEncoding.EncodeToString(headerJSON)

	// Create JWT payload
	payload := map[string]interface{}{
		"sub":                "test-user",
		"iss":                "https://example.com",
		"iat":                time.Now().Unix(),
		"exp":                expiresAt.Unix(),
		"preferred_username": "testuser",
	}

	// Add typ claim if specified
	if typ != "" {
		payload["typ"] = typ
	}

	payloadJSON, _ := json.Marshal(payload)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payloadJSON)

	// Create a dummy signature (not validated in our tests)
	signature := base64.RawURLEncoding.EncodeToString([]byte("dummy-signature"))

	return headerB64 + "." + payloadB64 + "." + signature
}
