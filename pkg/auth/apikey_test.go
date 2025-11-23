package auth

import (
	"context"
	"os"
	"testing"
)

func TestNewAPIKeyAuth(t *testing.T) {
	tests := []struct {
		name    string
		config  *APIKeyConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "missing name",
			config: &APIKeyConfig{
				Key:      "test-key",
				Location: APIKeyLocationHeader,
			},
			wantErr: true,
		},
		{
			name: "missing location",
			config: &APIKeyConfig{
				Key:  "test-key",
				Name: "X-API-Key",
			},
			wantErr: true,
		},
		{
			name: "invalid location",
			config: &APIKeyConfig{
				Key:      "test-key",
				Name:     "X-API-Key",
				Location: "invalid",
			},
			wantErr: true,
		},
		{
			name: "missing key and env var",
			config: &APIKeyConfig{
				Name:     "X-API-Key",
				Location: APIKeyLocationHeader,
			},
			wantErr: true,
		},
		{
			name: "valid header config",
			config: &APIKeyConfig{
				Key:      "test-key-123",
				Name:     "X-API-Key",
				Location: APIKeyLocationHeader,
			},
			wantErr: false,
		},
		{
			name: "valid query config",
			config: &APIKeyConfig{
				Key:      "test-key-456",
				Name:     "api_key",
				Location: APIKeyLocationQuery,
			},
			wantErr: false,
		},
		{
			name: "with prefix",
			config: &APIKeyConfig{
				Key:      "test-key-789",
				Name:     "Authorization",
				Location: APIKeyLocationHeader,
				Prefix:   "Bearer ",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewAPIKeyAuth(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewAPIKeyAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && auth == nil {
				t.Error("NewAPIKeyAuth() returned nil auth without error")
			}
		})
	}
}

func TestAPIKeyAuth_Authenticate(t *testing.T) {
	config := &APIKeyConfig{
		Key:      "test-key-secret",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}

	auth, err := NewAPIKeyAuth(config)
	if err != nil {
		t.Fatalf("NewAPIKeyAuth() failed: %v", err)
	}

	ctx := context.Background()
	token, err := auth.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	if token.AccessToken != "test-key-secret" {
		t.Errorf("token.AccessToken = %v, want %v", token.AccessToken, "test-key-secret")
	}

	if token.TokenType != "apikey" {
		t.Errorf("token.TokenType = %v, want %v", token.TokenType, "apikey")
	}
}

func TestAPIKeyAuth_GetHeaders(t *testing.T) {
	tests := []struct {
		name       string
		config     *APIKeyConfig
		token      *Token
		wantHeader string
		wantValue  string
	}{
		{
			name: "header without prefix",
			config: &APIKeyConfig{
				Key:      "test-key",
				Name:     "X-API-Key",
				Location: APIKeyLocationHeader,
			},
			token: &Token{
				AccessToken: "test-key",
			},
			wantHeader: "X-API-Key",
			wantValue:  "test-key",
		},
		{
			name: "header with Bearer prefix",
			config: &APIKeyConfig{
				Key:      "test-key",
				Name:     "Authorization",
				Location: APIKeyLocationHeader,
				Prefix:   "Bearer ",
			},
			token: &Token{
				AccessToken: "test-key",
			},
			wantHeader: "Authorization",
			wantValue:  "Bearer test-key",
		},
		{
			name: "header with Token prefix",
			config: &APIKeyConfig{
				Key:      "test-key",
				Name:     "Authorization",
				Location: APIKeyLocationHeader,
				Prefix:   "Token ",
			},
			token: &Token{
				AccessToken: "test-key",
			},
			wantHeader: "Authorization",
			wantValue:  "Token test-key",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewAPIKeyAuth(tt.config)
			if err != nil {
				t.Fatalf("NewAPIKeyAuth() failed: %v", err)
			}

			headers := auth.GetHeaders(tt.token)
			if headers == nil {
				t.Fatal("GetHeaders() returned nil")
			}

			value, ok := headers[tt.wantHeader]
			if !ok {
				t.Errorf("header %s not found", tt.wantHeader)
			}

			if value != tt.wantValue {
				t.Errorf("header value = %v, want %v", value, tt.wantValue)
			}
		})
	}
}

func TestAPIKeyAuth_GetQueryParams(t *testing.T) {
	config := &APIKeyConfig{
		Key:      "test-query-key",
		Name:     "api_key",
		Location: APIKeyLocationQuery,
	}

	auth, err := NewAPIKeyAuth(config)
	if err != nil {
		t.Fatalf("NewAPIKeyAuth() failed: %v", err)
	}

	token := &Token{
		AccessToken: "test-query-key",
	}

	params := auth.GetQueryParams(token)
	if params == nil {
		t.Fatal("GetQueryParams() returned nil")
	}

	value, ok := params["api_key"]
	if !ok {
		t.Error("query param api_key not found")
	}

	if value != "test-query-key" {
		t.Errorf("query param value = %v, want %v", value, "test-query-key")
	}
}

func TestAPIKeyAuth_EnvVar(t *testing.T) {
	// Set environment variable
	os.Setenv("TEST_API_KEY", "env-secret-key")
	defer os.Unsetenv("TEST_API_KEY")

	config := &APIKeyConfig{
		EnvVar:   "TEST_API_KEY",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}

	auth, err := NewAPIKeyAuth(config)
	if err != nil {
		t.Fatalf("NewAPIKeyAuth() failed: %v", err)
	}

	ctx := context.Background()
	token, err := auth.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	if token.AccessToken != "env-secret-key" {
		t.Errorf("token.AccessToken = %v, want %v", token.AccessToken, "env-secret-key")
	}
}

func TestAPIKeyAuth_RefreshToken(t *testing.T) {
	config := &APIKeyConfig{
		Key:      "test-key",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}

	auth, err := NewAPIKeyAuth(config)
	if err != nil {
		t.Fatalf("NewAPIKeyAuth() failed: %v", err)
	}

	ctx := context.Background()
	token := &Token{AccessToken: "test-key"}

	_, err = auth.RefreshToken(ctx, token)
	if err == nil {
		t.Error("RefreshToken() should return error for API key auth")
	}
}

func TestAPIKeyAuth_Type(t *testing.T) {
	config := &APIKeyConfig{
		Key:      "test-key",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}

	auth, err := NewAPIKeyAuth(config)
	if err != nil {
		t.Fatalf("NewAPIKeyAuth() failed: %v", err)
	}

	if auth.Type() != AuthTypeAPIKey {
		t.Errorf("Type() = %v, want %v", auth.Type(), AuthTypeAPIKey)
	}
}
