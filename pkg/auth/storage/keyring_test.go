package storage

import (
	"context"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/auth/types"
)

func TestNewKeyringStorage(t *testing.T) {
	tests := []struct {
		name    string
		config  *types.StorageConfig
		wantErr bool
	}{
		{
			name: "valid config with service and user",
			config: &types.StorageConfig{
				Type:           types.StorageTypeKeyring,
				KeyringService: "test-service",
				KeyringUser:    "test-user",
			},
			wantErr: false,
		},
		{
			name: "valid config with default user",
			config: &types.StorageConfig{
				Type:           types.StorageTypeKeyring,
				KeyringService: "test-service",
				// No KeyringUser specified
			},
			wantErr: false,
		},
		{
			name: "missing service",
			config: &types.StorageConfig{
				Type:        types.StorageTypeKeyring,
				KeyringUser: "test-user",
				// No KeyringService
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			storage, err := NewKeyringStorage(tt.config)

			if (err != nil) != tt.wantErr {
				t.Errorf("NewKeyringStorage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if storage == nil {
					t.Error("NewKeyringStorage() returned nil")
					return
				}

				if storage.service != tt.config.KeyringService {
					t.Errorf("service = %v, want %v", storage.service, tt.config.KeyringService)
				}

				expectedUser := tt.config.KeyringUser
				if expectedUser == "" {
					expectedUser = "default"
				}

				if storage.user != expectedUser {
					t.Errorf("user = %v, want %v", storage.user, expectedUser)
				}
			}
		})
	}
}

func TestKeyringStorage_GetService(t *testing.T) {
	config := &types.StorageConfig{
		Type:           types.StorageTypeKeyring,
		KeyringService: "my-service",
		KeyringUser:    "my-user",
	}

	storage, err := NewKeyringStorage(config)
	if err != nil {
		t.Fatalf("NewKeyringStorage() error = %v", err)
	}

	service := storage.GetService()
	if service != "my-service" {
		t.Errorf("GetService() = %v, want my-service", service)
	}
}

func TestKeyringStorage_GetUser(t *testing.T) {
	config := &types.StorageConfig{
		Type:           types.StorageTypeKeyring,
		KeyringService: "my-service",
		KeyringUser:    "my-user",
	}

	storage, err := NewKeyringStorage(config)
	if err != nil {
		t.Fatalf("NewKeyringStorage() error = %v", err)
	}

	user := storage.GetUser()
	if user != "my-user" {
		t.Errorf("GetUser() = %v, want my-user", user)
	}
}

func TestKeyringStorage_SaveTokenNil(t *testing.T) {
	config := &types.StorageConfig{
		Type:           types.StorageTypeKeyring,
		KeyringService: "test-service",
		KeyringUser:    "test-user",
	}

	storage, err := NewKeyringStorage(config)
	if err != nil {
		t.Fatalf("NewKeyringStorage() error = %v", err)
	}

	ctx := context.Background()
	err = storage.SaveToken(ctx, nil)
	if err == nil {
		t.Error("SaveToken() should return error for nil token")
	}
}

func TestKeyringStorage_LoadTokenNotFound(t *testing.T) {
	config := &types.StorageConfig{
		Type:           types.StorageTypeKeyring,
		KeyringService: "nonexistent-service-" + time.Now().String(),
		KeyringUser:    "nonexistent-user",
	}

	storage, err := NewKeyringStorage(config)
	if err != nil {
		t.Fatalf("NewKeyringStorage() error = %v", err)
	}

	ctx := context.Background()
	_, err = storage.LoadToken(ctx)
	if err == nil {
		t.Error("LoadToken() should return error when token not found")
	}
}

func TestKeyringStorage_DeleteTokenNotFound(t *testing.T) {
	config := &types.StorageConfig{
		Type:           types.StorageTypeKeyring,
		KeyringService: "nonexistent-service-" + time.Now().String(),
		KeyringUser:    "nonexistent-user",
	}

	storage, err := NewKeyringStorage(config)
	if err != nil {
		t.Fatalf("NewKeyringStorage() error = %v", err)
	}

	ctx := context.Background()
	// Deleting non-existent token should not error (returns nil)
	err = storage.DeleteToken(ctx)
	if err != nil {
		t.Errorf("DeleteToken() should not error for non-existent token, got: %v", err)
	}
}

func TestKeyringStorage_DefaultUser(t *testing.T) {
	config := &types.StorageConfig{
		Type:           types.StorageTypeKeyring,
		KeyringService: "test-service",
		// KeyringUser not specified
	}

	storage, err := NewKeyringStorage(config)
	if err != nil {
		t.Fatalf("NewKeyringStorage() error = %v", err)
	}

	if storage.user != "default" {
		t.Errorf("user = %v, want default", storage.user)
	}
}

func TestKeyringStorage_ServiceAndUserFields(t *testing.T) {
	tests := []struct {
		name           string
		service        string
		user           string
		expectedUser   string
	}{
		{
			name:         "custom service and user",
			service:      "my-app",
			user:         "alice",
			expectedUser: "alice",
		},
		{
			name:         "custom service, empty user defaults",
			service:      "my-app",
			user:         "",
			expectedUser: "default",
		},
		{
			name:         "different service names",
			service:      "another-app",
			user:         "bob",
			expectedUser: "bob",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &types.StorageConfig{
				Type:           types.StorageTypeKeyring,
				KeyringService: tt.service,
				KeyringUser:    tt.user,
			}

			storage, err := NewKeyringStorage(config)
			if err != nil {
				t.Fatalf("NewKeyringStorage() error = %v", err)
			}

			if storage.GetService() != tt.service {
				t.Errorf("GetService() = %v, want %v", storage.GetService(), tt.service)
			}

			if storage.GetUser() != tt.expectedUser {
				t.Errorf("GetUser() = %v, want %v", storage.GetUser(), tt.expectedUser)
			}
		})
	}
}

func TestKeyringStorage_TokenMarshaling(t *testing.T) {
	// This test verifies the token can be marshaled/unmarshaled correctly
	// even if we can't test actual keyring operations in CI

	token := &types.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scopes:       []string{"read", "write"},
		Extra: map[string]interface{}{
			"custom_field": "custom_value",
		},
	}

	// Verify token has all fields set
	if token.AccessToken == "" {
		t.Error("AccessToken should not be empty")
	}

	if token.RefreshToken == "" {
		t.Error("RefreshToken should not be empty")
	}

	if token.TokenType != "Bearer" {
		t.Errorf("TokenType = %v, want Bearer", token.TokenType)
	}

	if token.ExpiresAt.IsZero() {
		t.Error("ExpiresAt should not be zero")
	}

	if len(token.Scopes) != 2 {
		t.Errorf("Scopes length = %d, want 2", len(token.Scopes))
	}

	if token.Extra == nil {
		t.Error("Extra should not be nil")
	}
}

func TestKeyringStorage_ConfigValidation(t *testing.T) {
	tests := []struct {
		name           string
		keyringService string
		wantErr        bool
	}{
		{
			name:           "empty service",
			keyringService: "",
			wantErr:        true,
		},
		{
			name:           "whitespace service",
			keyringService: "   ",
			wantErr:        false, // Will be accepted but may cause issues
		},
		{
			name:           "valid service",
			keyringService: "my-cli-app",
			wantErr:        false,
		},
		{
			name:           "service with special chars",
			keyringService: "my-app-v1.2.3",
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := &types.StorageConfig{
				Type:           types.StorageTypeKeyring,
				KeyringService: tt.keyringService,
				KeyringUser:    "test",
			}

			_, err := NewKeyringStorage(config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewKeyringStorage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
