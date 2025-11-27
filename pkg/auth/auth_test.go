package auth

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/auth/types"
)

func TestToken_IsExpired(t *testing.T) {
	tests := []struct {
		name        string
		token       *Token
		wantExpired bool
	}{
		{
			name: "valid token not expired",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(1 * time.Hour),
			},
			wantExpired: false,
		},
		{
			name: "expired token",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(-1 * time.Hour),
			},
			wantExpired: true,
		},
		{
			name: "token expiring in 15 seconds (within buffer)",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(15 * time.Second),
			},
			wantExpired: true,
		},
		{
			name: "token with zero expiry",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Time{},
			},
			wantExpired: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsExpired(); got != tt.wantExpired {
				t.Errorf("Token.IsExpired() = %v, want %v", got, tt.wantExpired)
			}
		})
	}
}

func TestToken_IsValid(t *testing.T) {
	tests := []struct {
		name  string
		token *Token
		want  bool
	}{
		{
			name: "valid token",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(1 * time.Hour),
			},
			want: true,
		},
		{
			name: "expired token",
			token: &Token{
				AccessToken: "test-token",
				ExpiresAt:   time.Now().Add(-1 * time.Hour),
			},
			want: false,
		},
		{
			name: "empty access token",
			token: &Token{
				AccessToken: "",
				ExpiresAt:   time.Now().Add(1 * time.Hour),
			},
			want: false,
		},
		{
			name: "valid token no expiry",
			token: &Token{
				AccessToken: "test-token",
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.token.IsValid(); got != tt.want {
				t.Errorf("Token.IsValid() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAuthType(t *testing.T) {
	types := []AuthType{
		AuthTypeAPIKey,
		AuthTypeOAuth2,
		AuthTypeBasic,
		AuthTypeNone,
	}

	expected := []string{"apikey", "oauth2", "basic", "none"}

	for i, authType := range types {
		if string(authType) != expected[i] {
			t.Errorf("AuthType[%d] = %v, want %v", i, authType, expected[i])
		}
	}
}

func TestAPIKeyLocation(t *testing.T) {
	locations := []APIKeyLocation{
		APIKeyLocationHeader,
		APIKeyLocationQuery,
	}

	expected := []string{"header", "query"}

	for i, loc := range locations {
		if string(loc) != expected[i] {
			t.Errorf("APIKeyLocation[%d] = %v, want %v", i, loc, expected[i])
		}
	}
}

func TestOAuth2Flow(t *testing.T) {
	flows := []OAuth2Flow{
		OAuth2FlowAuthorizationCode,
		OAuth2FlowClientCredentials,
		OAuth2FlowPassword,
		OAuth2FlowDeviceCode,
	}

	expected := []string{
		"authorization_code",
		"client_credentials",
		"password",
		"device_code",
	}

	for i, flow := range flows {
		if string(flow) != expected[i] {
			t.Errorf("OAuth2Flow[%d] = %v, want %v", i, flow, expected[i])
		}
	}
}

func TestStorageType(t *testing.T) {
	types := []StorageType{
		StorageTypeFile,
		StorageTypeKeyring,
		StorageTypeMemory,
	}

	expected := []string{"file", "keyring", "memory"}

	for i, storType := range types {
		if string(storType) != expected[i] {
			t.Errorf("StorageType[%d] = %v, want %v", i, storType, expected[i])
		}
	}
}

// Mock authenticator for testing
type mockAuthenticator struct {
	authFunc    func(ctx context.Context) (*Token, error)
	refreshFunc func(ctx context.Context, token *Token) (*Token, error)
	headersFunc func(token *Token) map[string]string
}

func (m *mockAuthenticator) Type() AuthType {
	return AuthTypeAPIKey
}

func (m *mockAuthenticator) Authenticate(ctx context.Context) (*Token, error) {
	if m.authFunc != nil {
		return m.authFunc(ctx)
	}
	return &Token{
		AccessToken: "test-token",
		ExpiresAt:   time.Now().Add(time.Hour),
	}, nil
}

func (m *mockAuthenticator) RefreshToken(ctx context.Context, token *Token) (*Token, error) {
	if m.refreshFunc != nil {
		return m.refreshFunc(ctx, token)
	}
	return &Token{
		AccessToken:  "refreshed-token",
		RefreshToken: token.RefreshToken,
		ExpiresAt:    time.Now().Add(time.Hour),
	}, nil
}

func (m *mockAuthenticator) GetHeaders(token *Token) map[string]string {
	if m.headersFunc != nil {
		return m.headersFunc(token)
	}
	return map[string]string{
		"Authorization": "Bearer " + token.AccessToken,
	}
}

func (m *mockAuthenticator) Validate() error {
	return nil
}

// Mock storage for testing
type mockStorage struct {
	saveFunc   func(ctx context.Context, token *types.Token) error
	loadFunc   func(ctx context.Context) (*types.Token, error)
	deleteFunc func(ctx context.Context) error
}

func (m *mockStorage) SaveToken(ctx context.Context, token *types.Token) error {
	if m.saveFunc != nil {
		return m.saveFunc(ctx, token)
	}
	return nil
}

func (m *mockStorage) LoadToken(ctx context.Context) (*types.Token, error) {
	if m.loadFunc != nil {
		return m.loadFunc(ctx)
	}
	return nil, fmt.Errorf("no token stored")
}

func (m *mockStorage) DeleteToken(ctx context.Context) error {
	if m.deleteFunc != nil {
		return m.deleteFunc(ctx)
	}
	return nil
}

func TestNewAuthenticatedClient(t *testing.T) {
	auth := &mockAuthenticator{}
	storage := &mockStorage{}

	tests := []struct {
		name   string
		client *http.Client
	}{
		{
			name:   "with custom client",
			client: &http.Client{Timeout: 5 * time.Second},
		},
		{
			name:   "with nil client",
			client: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewAuthenticatedClient(tt.client, auth, storage)
			if client == nil {
				t.Error("NewAuthenticatedClient() returned nil")
				return
			}
			if client.client == nil {
				t.Error("client.client is nil")
			}
		})
	}
}

func TestAuthenticatedClient_Do(t *testing.T) {
	tests := []struct {
		name        string
		auth        *mockAuthenticator
		storage     *mockStorage
		wantErr     bool
		checkHeader bool
	}{
		{
			name: "successful request with cached token",
			auth: &mockAuthenticator{
				authFunc: func(ctx context.Context) (*Token, error) {
					return &Token{
						AccessToken: "cached-token",
						ExpiresAt:   time.Now().Add(time.Hour),
					}, nil
				},
			},
			storage:     &mockStorage{},
			wantErr:     false,
			checkHeader: true,
		},
		{
			name: "request with stored token",
			auth: &mockAuthenticator{},
			storage: &mockStorage{
				loadFunc: func(ctx context.Context) (*types.Token, error) {
					return &types.Token{
						AccessToken: "stored-token",
						ExpiresAt:   time.Now().Add(time.Hour),
					}, nil
				},
			},
			wantErr:     false,
			checkHeader: true,
		},
		{
			name: "request with expired token refresh",
			auth: &mockAuthenticator{
				refreshFunc: func(ctx context.Context, token *Token) (*Token, error) {
					return &Token{
						AccessToken:  "refreshed-token",
						RefreshToken: "refresh",
						ExpiresAt:    time.Now().Add(time.Hour),
					}, nil
				},
			},
			storage: &mockStorage{
				loadFunc: func(ctx context.Context) (*types.Token, error) {
					return &types.Token{
						AccessToken:  "expired-token",
						RefreshToken: "refresh",
						ExpiresAt:    time.Now().Add(-time.Hour),
					}, nil
				},
			},
			wantErr:     false,
			checkHeader: true,
		},
		{
			name: "authentication error",
			auth: &mockAuthenticator{
				authFunc: func(ctx context.Context) (*Token, error) {
					return nil, fmt.Errorf("auth failed")
				},
			},
			storage: &mockStorage{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if tt.checkHeader && r.Header.Get("Authorization") == "" {
					t.Error("Authorization header not set")
				}
				w.WriteHeader(http.StatusOK)
				_, _ = w.Write([]byte("OK"))
			}))
			defer server.Close()

			client := NewAuthenticatedClient(nil, tt.auth, tt.storage)
			req, _ := http.NewRequest("GET", server.URL, nil)

			resp, err := client.Do(req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Do() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if resp != nil {
				_ = resp.Body.Close()
				if !tt.wantErr && resp.StatusCode != http.StatusOK {
					t.Errorf("Do() status = %v, want %v", resp.StatusCode, http.StatusOK)
				}
			}
		})
	}
}

func TestAuthenticatedClient_getValidToken(t *testing.T) {
	tests := []struct {
		name        string
		cachedToken *Token
		auth        *mockAuthenticator
		storage     *mockStorage
		wantErr     bool
		wantToken   string
	}{
		{
			name: "use cached valid token",
			cachedToken: &Token{
				AccessToken: "cached",
				ExpiresAt:   time.Now().Add(time.Hour),
			},
			auth:      &mockAuthenticator{},
			storage:   &mockStorage{},
			wantErr:   false,
			wantToken: "cached",
		},
		{
			name:        "load from storage",
			cachedToken: nil,
			auth:        &mockAuthenticator{},
			storage: &mockStorage{
				loadFunc: func(ctx context.Context) (*types.Token, error) {
					return &types.Token{
						AccessToken: "stored",
						ExpiresAt:   time.Now().Add(time.Hour),
					}, nil
				},
			},
			wantErr:   false,
			wantToken: "stored",
		},
		{
			name:        "refresh expired token",
			cachedToken: nil,
			auth: &mockAuthenticator{
				refreshFunc: func(ctx context.Context, token *Token) (*Token, error) {
					return &Token{
						AccessToken:  "refreshed",
						RefreshToken: "refresh",
						ExpiresAt:    time.Now().Add(time.Hour),
					}, nil
				},
			},
			storage: &mockStorage{
				loadFunc: func(ctx context.Context) (*types.Token, error) {
					return &types.Token{
						AccessToken:  "expired",
						RefreshToken: "refresh",
						ExpiresAt:    time.Now().Add(-time.Hour),
					}, nil
				},
			},
			wantErr:   false,
			wantToken: "refreshed",
		},
		{
			name:        "new authentication",
			cachedToken: nil,
			auth: &mockAuthenticator{
				authFunc: func(ctx context.Context) (*Token, error) {
					return &Token{
						AccessToken: "new-token",
						ExpiresAt:   time.Now().Add(time.Hour),
					}, nil
				},
			},
			storage:   &mockStorage{},
			wantErr:   false,
			wantToken: "new-token",
		},
		{
			name:        "authentication fails",
			cachedToken: nil,
			auth: &mockAuthenticator{
				authFunc: func(ctx context.Context) (*Token, error) {
					return nil, fmt.Errorf("auth error")
				},
			},
			storage: &mockStorage{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := &AuthenticatedClient{
				client:        http.DefaultClient,
				authenticator: tt.auth,
				storage:       tt.storage,
				token:         tt.cachedToken,
			}

			ctx := context.Background()
			token, err := client.getValidToken(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("getValidToken() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if token == nil {
					t.Fatal("getValidToken() returned nil token")
				}
				if token.AccessToken != tt.wantToken {
					t.Errorf("getValidToken() token = %v, want %v", token.AccessToken, tt.wantToken)
				}
			}
		})
	}
}
