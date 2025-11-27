package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
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
