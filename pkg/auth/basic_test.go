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
	_ = os.Setenv("USER", "envuser")
	_ = os.Setenv("PASS", "envpass")
	defer func() {
		_ = os.Unsetenv("USER")
		_ = os.Unsetenv("PASS")
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
	_ = os.Setenv("TEST_USERNAME", "envuser")
	_ = os.Setenv("TEST_PASSWORD", "envpass")
	defer func() {
		_ = os.Unsetenv("TEST_USERNAME")
		_ = os.Unsetenv("TEST_PASSWORD")
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

	decoded, err := base64.StdEncoding.DecodeString(token.AccessToken)
	if err != nil {
		t.Fatalf("failed to decode token: %v", err)
	}
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

	decoded, err := base64.StdEncoding.DecodeString(token.AccessToken)
	if err != nil {
		t.Fatalf("failed to decode token: %v", err)
	}
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

func TestBasicAuth_GetHeaders_NilToken(t *testing.T) {
	config := &BasicConfig{
		Username: "user",
		Password: "pass",
	}

	auth, err := NewBasicAuth(config)
	if err != nil {
		t.Fatalf("NewBasicAuth() failed: %v", err)
	}

	headers := auth.GetHeaders(nil)
	if headers != nil {
		t.Error("GetHeaders(nil) should return nil")
	}
}

func TestBasicAuth_GetHeaders_EmptyToken(t *testing.T) {
	config := &BasicConfig{
		Username: "user",
		Password: "pass",
	}

	auth, err := NewBasicAuth(config)
	if err != nil {
		t.Fatalf("NewBasicAuth() failed: %v", err)
	}

	headers := auth.GetHeaders(&Token{AccessToken: ""})
	if headers != nil {
		t.Error("GetHeaders(empty token) should return nil")
	}
}

func TestBasicAuth_UsernameEnvVarReference(t *testing.T) {
	_ = os.Setenv("MY_USERNAME", "refuser")
	defer func() { _ = os.Unsetenv("MY_USERNAME") }()

	config := &BasicConfig{
		Username: "$MY_USERNAME",
		Password: "testpass",
	}

	auth, err := NewBasicAuth(config)
	if err != nil {
		t.Fatalf("NewBasicAuth() failed: %v", err)
	}

	if auth.GetUsername() != "refuser" {
		t.Errorf("username = %v, want refuser", auth.GetUsername())
	}
}

func TestBasicAuth_PasswordEnvVarReference(t *testing.T) {
	_ = os.Setenv("MY_PASSWORD", "refpass")
	defer func() { _ = os.Unsetenv("MY_PASSWORD") }()

	config := &BasicConfig{
		Username: "testuser",
		Password: "$MY_PASSWORD",
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

	decoded, err := base64.StdEncoding.DecodeString(token.AccessToken)
	if err != nil {
		t.Fatalf("failed to decode token: %v", err)
	}
	if string(decoded) != "testuser:refpass" {
		t.Errorf("credentials = %v, want testuser:refpass", string(decoded))
	}
}

func TestBasicAuth_UsernameEnvVarReferenceMissing(t *testing.T) {
	config := &BasicConfig{
		Username: "$NONEXISTENT_USER",
		Password: "testpass",
	}

	_, err := NewBasicAuth(config)
	if err == nil {
		t.Error("NewBasicAuth() should fail with missing username env var reference")
	}
}

func TestBasicAuth_PasswordEnvVarReferenceMissing(t *testing.T) {
	config := &BasicConfig{
		Username: "testuser",
		Password: "$NONEXISTENT_PASS",
	}

	_, err := NewBasicAuth(config)
	if err == nil {
		t.Error("NewBasicAuth() should fail with missing password env var reference")
	}
}

func TestBasicAuth_EnvUsernamePreferredOverUsername(t *testing.T) {
	_ = os.Setenv("PREFERRED_USER", "env-user")
	defer func() { _ = os.Unsetenv("PREFERRED_USER") }()

	config := &BasicConfig{
		Username:    "direct-user",
		EnvUsername: "PREFERRED_USER",
		Password:    "testpass",
	}

	auth, err := NewBasicAuth(config)
	if err != nil {
		t.Fatalf("NewBasicAuth() failed: %v", err)
	}

	if auth.GetUsername() != "env-user" {
		t.Errorf("username = %v, want env-user (EnvUsername should be preferred)", auth.GetUsername())
	}
}

func TestBasicAuth_EnvPasswordPreferredOverPassword(t *testing.T) {
	_ = os.Setenv("PREFERRED_PASS", "env-pass")
	defer func() { _ = os.Unsetenv("PREFERRED_PASS") }()

	config := &BasicConfig{
		Username:    "testuser",
		Password:    "direct-pass",
		EnvPassword: "PREFERRED_PASS",
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

	decoded, err := base64.StdEncoding.DecodeString(token.AccessToken)
	if err != nil {
		t.Fatalf("failed to decode token: %v", err)
	}
	if string(decoded) != "testuser:env-pass" {
		t.Errorf("credentials = %v, want testuser:env-pass (EnvPassword should be preferred)", string(decoded))
	}
}

func TestBasicAuth_Authenticate_EmptyCredentials(t *testing.T) {
	auth := &BasicAuth{
		config:   &BasicConfig{},
		username: "",
		password: "",
	}

	ctx := context.Background()
	_, err := auth.Authenticate(ctx)
	if err == nil {
		t.Error("Authenticate() should fail with empty credentials")
	}
}

func TestBasicAuth_Authenticate_EmptyUsername(t *testing.T) {
	auth := &BasicAuth{
		config:   &BasicConfig{},
		username: "",
		password: "pass",
	}

	ctx := context.Background()
	_, err := auth.Authenticate(ctx)
	if err == nil {
		t.Error("Authenticate() should fail with empty username")
	}
}

func TestBasicAuth_Authenticate_EmptyPassword(t *testing.T) {
	auth := &BasicAuth{
		config:   &BasicConfig{},
		username: "user",
		password: "",
	}

	ctx := context.Background()
	_, err := auth.Authenticate(ctx)
	if err == nil {
		t.Error("Authenticate() should fail with empty password")
	}
}
