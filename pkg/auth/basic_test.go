package auth

import (
	"context"
	"encoding/base64"
	"os"
	"strings"
	"testing"
)

func TestNewBasicAuth(t *testing.T) {
	tests := []struct {
		name    string
		config  *BasicConfig
		wantErr bool
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
		},
		{
			name: "missing username",
			config: &BasicConfig{
				Password: "password123",
			},
			wantErr: true,
		},
		{
			name: "missing password",
			config: &BasicConfig{
				Username: "user123",
			},
			wantErr: true,
		},
		{
			name: "valid config",
			config: &BasicConfig{
				Username: "user123",
				Password: "password123",
			},
			wantErr: false,
		},
		{
			name: "valid with env vars",
			config: &BasicConfig{
				EnvUsername: "USER",
				EnvPassword: "PASS",
			},
			wantErr: false,
		},
	}

	// Set up environment for env var test
	os.Setenv("USER", "envuser")
	os.Setenv("PASS", "envpass")
	defer func() {
		os.Unsetenv("USER")
		os.Unsetenv("PASS")
	}()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			auth, err := NewBasicAuth(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewBasicAuth() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && auth == nil {
				t.Error("NewBasicAuth() returned nil auth without error")
			}
		})
	}
}

func TestBasicAuth_Authenticate(t *testing.T) {
	config := &BasicConfig{
		Username: "testuser",
		Password: "testpass",
	}

	auth, err := NewBasicAuth(config)
	if err != nil {
		t.Fatalf("NewBasicAuth() failed: %v", err)
	}

	ctx := context.Background()
	token, err := auth.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	// Verify token type
	if token.TokenType != "Basic" {
		t.Errorf("token.TokenType = %v, want %v", token.TokenType, "Basic")
	}

	// Decode and verify credentials
	decoded, err := base64.StdEncoding.DecodeString(token.AccessToken)
	if err != nil {
		t.Fatalf("failed to decode token: %v", err)
	}

	credentials := string(decoded)
	expected := "testuser:testpass"
	if credentials != expected {
		t.Errorf("decoded credentials = %v, want %v", credentials, expected)
	}

	// Verify extra fields
	if username, ok := token.Extra["username"].(string); !ok || username != "testuser" {
		t.Errorf("token.Extra[username] = %v, want %v", username, "testuser")
	}
}

func TestBasicAuth_GetHeaders(t *testing.T) {
	config := &BasicConfig{
		Username: "user",
		Password: "pass",
	}

	auth, err := NewBasicAuth(config)
	if err != nil {
		t.Fatalf("NewBasicAuth() failed: %v", err)
	}

	ctx := context.Background()
	token, err := auth.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	headers := auth.GetHeaders(token)
	if headers == nil {
		t.Fatal("GetHeaders() returned nil")
	}

	authHeader, ok := headers["Authorization"]
	if !ok {
		t.Fatal("Authorization header not found")
	}

	if !strings.HasPrefix(authHeader, "Basic ") {
		t.Errorf("Authorization header = %v, want prefix 'Basic '", authHeader)
	}

	// Verify the encoded credentials
	encoded := strings.TrimPrefix(authHeader, "Basic ")
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		t.Fatalf("failed to decode Authorization header: %v", err)
	}

	if string(decoded) != "user:pass" {
		t.Errorf("decoded credentials = %v, want %v", string(decoded), "user:pass")
	}
}

func TestBasicAuth_EnvVars(t *testing.T) {
	// Set environment variables
	os.Setenv("TEST_USERNAME", "envuser")
	os.Setenv("TEST_PASSWORD", "envpass")
	defer func() {
		os.Unsetenv("TEST_USERNAME")
		os.Unsetenv("TEST_PASSWORD")
	}()

	config := &BasicConfig{
		EnvUsername: "TEST_USERNAME",
		EnvPassword: "TEST_PASSWORD",
	}

	auth, err := NewBasicAuth(config)
	if err != nil {
		t.Fatalf("NewBasicAuth() failed: %v", err)
	}

	if auth.GetUsername() != "envuser" {
		t.Errorf("username = %v, want %v", auth.GetUsername(), "envuser")
	}

	ctx := context.Background()
	token, err := auth.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	decoded, _ := base64.StdEncoding.DecodeString(token.AccessToken)
	if string(decoded) != "envuser:envpass" {
		t.Errorf("credentials = %v, want %v", string(decoded), "envuser:envpass")
	}
}

func TestBasicAuth_SetCredentials(t *testing.T) {
	config := &BasicConfig{
		Username: "original",
		Password: "original",
	}

	auth, err := NewBasicAuth(config)
	if err != nil {
		t.Fatalf("NewBasicAuth() failed: %v", err)
	}

	// Set new credentials
	auth.SetCredentials("newuser", "newpass")

	ctx := context.Background()
	token, err := auth.Authenticate(ctx)
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	decoded, _ := base64.StdEncoding.DecodeString(token.AccessToken)
	if string(decoded) != "newuser:newpass" {
		t.Errorf("credentials = %v, want %v", string(decoded), "newuser:newpass")
	}
}

func TestBasicAuth_RefreshToken(t *testing.T) {
	config := &BasicConfig{
		Username: "user",
		Password: "pass",
	}

	auth, err := NewBasicAuth(config)
	if err != nil {
		t.Fatalf("NewBasicAuth() failed: %v", err)
	}

	ctx := context.Background()
	token := &Token{AccessToken: "test"}

	_, err = auth.RefreshToken(ctx, token)
	if err == nil {
		t.Error("RefreshToken() should return error for Basic auth")
	}
}

func TestBasicAuth_Type(t *testing.T) {
	config := &BasicConfig{
		Username: "user",
		Password: "pass",
	}

	auth, err := NewBasicAuth(config)
	if err != nil {
		t.Fatalf("NewBasicAuth() failed: %v", err)
	}

	if auth.Type() != AuthTypeBasic {
		t.Errorf("Type() = %v, want %v", auth.Type(), AuthTypeBasic)
	}
}
