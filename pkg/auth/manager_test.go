package auth

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/auth/storage"
)

func TestNewManager(t *testing.T) {
	manager := NewManager("test-cli")
	if manager == nil {
		t.Fatal("NewManager() returned nil")
	}
}

func TestManager_RegisterAuthenticator(t *testing.T) {
	manager := NewManager("test-cli")

	config := &APIKeyConfig{
		Key:      "test-key",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}

	auth, err := NewAPIKeyAuth(config)
	if err != nil {
		t.Fatalf("NewAPIKeyAuth() failed: %v", err)
	}

	err = manager.RegisterAuthenticator("api", auth)
	if err != nil {
		t.Fatalf("RegisterAuthenticator() failed: %v", err)
	}

	// Try to get it back
	retrieved, err := manager.GetAuthenticator("api")
	if err != nil {
		t.Fatalf("GetAuthenticator() failed: %v", err)
	}

	if retrieved == nil {
		t.Error("GetAuthenticator() returned nil")
	}
}

func TestManager_RegisterNilAuthenticator(t *testing.T) {
	manager := NewManager("test-cli")

	err := manager.RegisterAuthenticator("api", nil)
	if err == nil {
		t.Error("RegisterAuthenticator() should return error for nil authenticator")
	}
}

func TestManager_GetNonexistentAuthenticator(t *testing.T) {
	manager := NewManager("test-cli")

	_, err := manager.GetAuthenticator("nonexistent")
	if err == nil {
		t.Error("GetAuthenticator() should return error for nonexistent authenticator")
	}
}

func TestManager_SetDefault(t *testing.T) {
	manager := NewManager("test-cli")

	// Register two authenticators
	apiConfig := &APIKeyConfig{
		Key:      "test-key",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}
	apiAuth, _ := NewAPIKeyAuth(apiConfig)
	manager.RegisterAuthenticator("api", apiAuth)

	basicConfig := &BasicConfig{
		Username: "user",
		Password: "pass",
	}
	basicAuth, _ := NewBasicAuth(basicConfig)
	manager.RegisterAuthenticator("basic", basicAuth)

	// Set default to basic
	err := manager.SetDefault("basic")
	if err != nil {
		t.Fatalf("SetDefault() failed: %v", err)
	}

	// Get without name should return default
	auth, err := manager.GetAuthenticator("")
	if err != nil {
		t.Fatalf("GetAuthenticator() failed: %v", err)
	}

	if auth.Type() != AuthTypeBasic {
		t.Errorf("Default authenticator type = %v, want %v", auth.Type(), AuthTypeBasic)
	}
}

func TestManager_RegisterStorage(t *testing.T) {
	manager := NewManager("test-cli")

	stor := storage.NewMemoryStorage()
	manager.RegisterStorage("memory", stor)

	retrieved, err := manager.GetStorage("memory")
	if err != nil {
		t.Fatalf("GetStorage() failed: %v", err)
	}

	if retrieved == nil {
		t.Error("GetStorage() returned nil")
	}
}

func TestManager_CreateFromConfig(t *testing.T) {
	manager := NewManager("test-cli")

	configs := map[string]*Config{
		"api": {
			Type: AuthTypeAPIKey,
			APIKey: &APIKeyConfig{
				Key:      "test-key",
				Name:     "X-API-Key",
				Location: APIKeyLocationHeader,
			},
			Storage: &StorageConfig{
				Type: StorageTypeMemory,
			},
		},
		"basic": {
			Type: AuthTypeBasic,
			Basic: &BasicConfig{
				Username: "user",
				Password: "pass",
			},
		},
	}

	err := manager.CreateFromConfig(configs)
	if err != nil {
		t.Fatalf("CreateFromConfig() failed: %v", err)
	}

	// Verify authenticators were created
	apiAuth, err := manager.GetAuthenticator("api")
	if err != nil {
		t.Errorf("Failed to get api authenticator: %v", err)
	}
	if apiAuth.Type() != AuthTypeAPIKey {
		t.Errorf("api auth type = %v, want %v", apiAuth.Type(), AuthTypeAPIKey)
	}

	basicAuth, err := manager.GetAuthenticator("basic")
	if err != nil {
		t.Errorf("Failed to get basic authenticator: %v", err)
	}
	if basicAuth.Type() != AuthTypeBasic {
		t.Errorf("basic auth type = %v, want %v", basicAuth.Type(), AuthTypeBasic)
	}

	// Verify storage was created for api
	_, err = manager.GetStorage("api")
	if err != nil {
		t.Errorf("Failed to get api storage: %v", err)
	}
}

func TestManager_Authenticate(t *testing.T) {
	manager := NewManager("test-cli")

	config := &APIKeyConfig{
		Key:      "test-key",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}
	auth, _ := NewAPIKeyAuth(config)
	manager.RegisterAuthenticator("api", auth)

	ctx := context.Background()
	token, err := manager.Authenticate(ctx, "api")
	if err != nil {
		t.Fatalf("Authenticate() failed: %v", err)
	}

	if token.AccessToken != "test-key" {
		t.Errorf("token.AccessToken = %v, want %v", token.AccessToken, "test-key")
	}
}

func TestManager_GetToken(t *testing.T) {
	manager := NewManager("test-cli")

	config := &APIKeyConfig{
		Key:      "test-key",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}
	auth, _ := NewAPIKeyAuth(config)
	manager.RegisterAuthenticator("api", auth)

	// Register memory storage
	stor := storage.NewMemoryStorage()
	manager.RegisterStorage("api", stor)

	ctx := context.Background()

	// First call should authenticate
	token1, err := manager.GetToken(ctx, "api")
	if err != nil {
		t.Fatalf("GetToken() failed: %v", err)
	}

	// Second call should use cached token
	token2, err := manager.GetToken(ctx, "api")
	if err != nil {
		t.Fatalf("GetToken() second call failed: %v", err)
	}

	if token1.AccessToken != token2.AccessToken {
		t.Error("GetToken() returned different tokens")
	}
}

func TestManager_GetToken_WithRefresh(t *testing.T) {
	manager := NewManager("test-cli")

	// Use mock authenticator that supports refresh
	mockAuth := &mockAuthenticator{
		refreshFunc: func(ctx context.Context, token *Token) (*Token, error) {
			return &Token{
				AccessToken:  "refreshed-token",
				RefreshToken: token.RefreshToken,
				ExpiresAt:    time.Now().Add(time.Hour),
			}, nil
		},
	}
	manager.RegisterAuthenticator("mock", mockAuth)

	stor := storage.NewMemoryStorage()
	manager.RegisterStorage("mock", stor)

	ctx := context.Background()

	// Store an expired token with refresh token
	expiredToken := &Token{
		AccessToken:  "expired-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Now().Add(-time.Hour),
	}
	stor.SaveToken(ctx, expiredToken)

	// GetToken should refresh the expired token
	token, err := manager.GetToken(ctx, "mock")
	if err != nil {
		t.Fatalf("GetToken() failed: %v", err)
	}

	if token.AccessToken != "refreshed-token" {
		t.Errorf("GetToken() AccessToken = %v, want refreshed-token", token.AccessToken)
	}
}

func TestManager_GetToken_RefreshFails(t *testing.T) {
	manager := NewManager("test-cli")

	// Use mock authenticator where refresh fails
	mockAuth := &mockAuthenticator{
		refreshFunc: func(ctx context.Context, token *Token) (*Token, error) {
			return nil, fmt.Errorf("refresh failed")
		},
		authFunc: func(ctx context.Context) (*Token, error) {
			return &Token{
				AccessToken: "new-auth-token",
				ExpiresAt:   time.Now().Add(time.Hour),
			}, nil
		},
	}
	manager.RegisterAuthenticator("mock", mockAuth)

	stor := storage.NewMemoryStorage()
	manager.RegisterStorage("mock", stor)

	ctx := context.Background()

	// Store an expired token with refresh token
	expiredToken := &Token{
		AccessToken:  "expired-token",
		RefreshToken: "refresh-token",
		ExpiresAt:    time.Now().Add(-time.Hour),
	}
	stor.SaveToken(ctx, expiredToken)

	// GetToken should fall back to authentication when refresh fails
	token, err := manager.GetToken(ctx, "mock")
	if err != nil {
		t.Fatalf("GetToken() failed: %v", err)
	}

	if token.AccessToken != "new-auth-token" {
		t.Errorf("GetToken() AccessToken = %v, want new-auth-token", token.AccessToken)
	}
}

func TestManager_Logout(t *testing.T) {
	manager := NewManager("test-cli")

	config := &APIKeyConfig{
		Key:      "test-key",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}
	auth, _ := NewAPIKeyAuth(config)
	manager.RegisterAuthenticator("api", auth)

	stor := storage.NewMemoryStorage()
	manager.RegisterStorage("api", stor)

	ctx := context.Background()

	// Authenticate
	token, _ := manager.Authenticate(ctx, "api")
	stor.SaveToken(ctx, token)

	// Logout
	err := manager.Logout(ctx, "api")
	if err != nil {
		t.Fatalf("Logout() failed: %v", err)
	}

	// Token should be deleted
	_, err = stor.LoadToken(ctx)
	if err == nil {
		t.Error("Token should be deleted after logout")
	}
}

func TestManager_GetAuthenticatedClient(t *testing.T) {
	manager := NewManager("test-cli")

	config := &APIKeyConfig{
		Key:      "test-key",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}
	auth, _ := NewAPIKeyAuth(config)
	manager.RegisterAuthenticator("api", auth)

	client, err := manager.GetAuthenticatedClient("api", "")
	if err != nil {
		t.Fatalf("GetAuthenticatedClient() failed: %v", err)
	}

	if client == nil {
		t.Error("GetAuthenticatedClient() returned nil")
	}
}

func TestNoneAuth(t *testing.T) {
	auth := &NoneAuth{}

	if auth.Type() != AuthTypeNone {
		t.Errorf("Type() = %v, want %v", auth.Type(), AuthTypeNone)
	}

	if err := auth.Validate(); err != nil {
		t.Errorf("Validate() failed: %v", err)
	}

	ctx := context.Background()
	token, err := auth.Authenticate(ctx)
	if err != nil {
		t.Errorf("Authenticate() failed: %v", err)
	}
	if token == nil {
		t.Error("Authenticate() returned nil token")
	}

	headers := auth.GetHeaders(token)
	if headers != nil {
		t.Error("GetHeaders() should return nil for NoneAuth")
	}

	_, err = auth.RefreshToken(ctx, token)
	if err == nil {
		t.Error("RefreshToken() should return error for NoneAuth")
	}
}

func TestManager_ListAuthenticators(t *testing.T) {
	manager := NewManager("test-cli")

	apiConfig := &APIKeyConfig{
		Key:      "test-key",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}
	apiAuth, _ := NewAPIKeyAuth(apiConfig)
	manager.RegisterAuthenticator("api", apiAuth)

	basicConfig := &BasicConfig{
		Username: "user",
		Password: "pass",
	}
	basicAuth, _ := NewBasicAuth(basicConfig)
	manager.RegisterAuthenticator("basic", basicAuth)

	names := manager.ListAuthenticators()
	if len(names) != 2 {
		t.Errorf("ListAuthenticators() returned %d names, want 2", len(names))
	}

	found := make(map[string]bool)
	for _, name := range names {
		found[name] = true
	}

	if !found["api"] || !found["basic"] {
		t.Errorf("ListAuthenticators() = %v, want [api, basic]", names)
	}
}

func TestManager_ListStorages(t *testing.T) {
	manager := NewManager("test-cli")

	stor1 := storage.NewMemoryStorage()
	manager.RegisterStorage("memory1", stor1)

	stor2 := storage.NewMemoryStorage()
	manager.RegisterStorage("memory2", stor2)

	names := manager.ListStorages()
	if len(names) != 2 {
		t.Errorf("ListStorages() returned %d names, want 2", len(names))
	}

	found := make(map[string]bool)
	for _, name := range names {
		found[name] = true
	}

	if !found["memory1"] || !found["memory2"] {
		t.Errorf("ListStorages() = %v, want [memory1, memory2]", names)
	}
}

func TestManager_LogoutAll(t *testing.T) {
	manager := NewManager("test-cli")

	// Create multiple authenticators with storage
	apiConfig := &APIKeyConfig{
		Key:      "test-key",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}
	apiAuth, _ := NewAPIKeyAuth(apiConfig)
	manager.RegisterAuthenticator("api", apiAuth)

	basicConfig := &BasicConfig{
		Username: "user",
		Password: "pass",
	}
	basicAuth, _ := NewBasicAuth(basicConfig)
	manager.RegisterAuthenticator("basic", basicAuth)

	stor1 := storage.NewMemoryStorage()
	manager.RegisterStorage("api", stor1)

	stor2 := storage.NewMemoryStorage()
	manager.RegisterStorage("basic", stor2)

	ctx := context.Background()

	// Authenticate both
	token1, _ := manager.Authenticate(ctx, "api")
	stor1.SaveToken(ctx, token1)

	token2, _ := manager.Authenticate(ctx, "basic")
	stor2.SaveToken(ctx, token2)

	// Logout all
	err := manager.LogoutAll(ctx)
	if err != nil {
		t.Fatalf("LogoutAll() failed: %v", err)
	}

	// Verify all tokens are deleted
	_, err = stor1.LoadToken(ctx)
	if err == nil {
		t.Error("Token 1 should be deleted after LogoutAll")
	}

	_, err = stor2.LoadToken(ctx)
	if err == nil {
		t.Error("Token 2 should be deleted after LogoutAll")
	}
}

func TestManager_RefreshToken(t *testing.T) {
	manager := NewManager("test-cli")

	// Use a mock authenticator that supports refresh
	mockAuth := &mockAuthenticator{}
	manager.RegisterAuthenticator("mock", mockAuth)

	stor := storage.NewMemoryStorage()
	manager.RegisterStorage("mock", stor)

	ctx := context.Background()
	token := &Token{
		AccessToken:  "old-token",
		RefreshToken: "refresh-token",
	}

	refreshed, err := manager.RefreshToken(ctx, "mock", token)
	if err != nil {
		t.Fatalf("RefreshToken() failed: %v", err)
	}

	if refreshed == nil {
		t.Fatal("RefreshToken() returned nil")
	}

	if refreshed.AccessToken != "refreshed-token" {
		t.Errorf("RefreshToken() AccessToken = %v, want refreshed-token", refreshed.AccessToken)
	}
}

func TestManager_SetDefaultNonexistent(t *testing.T) {
	manager := NewManager("test-cli")

	err := manager.SetDefault("nonexistent")
	if err == nil {
		t.Error("SetDefault() should return error for nonexistent authenticator")
	}
}

func TestManager_GetStorageNonexistent(t *testing.T) {
	manager := NewManager("test-cli")

	_, err := manager.GetStorage("nonexistent")
	if err == nil {
		t.Error("GetStorage() should return error for nonexistent storage")
	}
}

func TestManager_CreateFromConfig_Errors(t *testing.T) {
	tests := []struct {
		name    string
		configs map[string]*Config
		wantErr bool
	}{
		{
			name: "missing APIKey config",
			configs: map[string]*Config{
				"api": {
					Type: AuthTypeAPIKey,
				},
			},
			wantErr: true,
		},
		{
			name: "missing OAuth2 config",
			configs: map[string]*Config{
				"oauth": {
					Type: AuthTypeOAuth2,
				},
			},
			wantErr: true,
		},
		{
			name: "missing Basic config",
			configs: map[string]*Config{
				"basic": {
					Type: AuthTypeBasic,
				},
			},
			wantErr: true,
		},
		{
			name: "unsupported auth type",
			configs: map[string]*Config{
				"unknown": {
					Type: AuthType("unknown"),
				},
			},
			wantErr: true,
		},
		{
			name: "none auth type",
			configs: map[string]*Config{
				"none": {
					Type: AuthTypeNone,
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			manager := NewManager("test-cli")
			err := manager.CreateFromConfig(tt.configs)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateFromConfig() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestManager_LogoutWithoutStorage(t *testing.T) {
	manager := NewManager("test-cli")

	config := &APIKeyConfig{
		Key:      "test-key",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}
	auth, _ := NewAPIKeyAuth(config)
	manager.RegisterAuthenticator("api", auth)

	ctx := context.Background()

	// Logout without storage should succeed
	err := manager.Logout(ctx, "api")
	if err != nil {
		t.Errorf("Logout() without storage failed: %v", err)
	}
}

func TestManager_LogoutDefaultAuth(t *testing.T) {
	manager := NewManager("test-cli")

	config := &APIKeyConfig{
		Key:      "test-key",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}
	auth, _ := NewAPIKeyAuth(config)
	manager.RegisterAuthenticator("api", auth)

	stor := storage.NewMemoryStorage()
	manager.RegisterStorage("api", stor)

	ctx := context.Background()

	// Authenticate
	token, _ := manager.Authenticate(ctx, "api")
	stor.SaveToken(ctx, token)

	// Logout with empty name should use default
	err := manager.Logout(ctx, "")
	if err != nil {
		t.Fatalf("Logout() with empty name failed: %v", err)
	}

	// Token should be deleted
	_, err = stor.LoadToken(ctx)
	if err == nil {
		t.Error("Token should be deleted after logout")
	}
}

func TestManager_GetTokenDefault(t *testing.T) {
	manager := NewManager("test-cli")

	config := &APIKeyConfig{
		Key:      "test-key",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}
	auth, _ := NewAPIKeyAuth(config)
	manager.RegisterAuthenticator("api", auth)

	ctx := context.Background()

	// GetToken with empty name should use default
	token, err := manager.GetToken(ctx, "")
	if err != nil {
		t.Fatalf("GetToken() with empty name failed: %v", err)
	}

	if token == nil {
		t.Fatal("GetToken() returned nil")
	}
}

func TestManager_AuthenticateNonexistent(t *testing.T) {
	manager := NewManager("test-cli")

	ctx := context.Background()
	_, err := manager.Authenticate(ctx, "nonexistent")
	if err == nil {
		t.Error("Authenticate() should return error for nonexistent authenticator")
	}
}

func TestManager_RefreshTokenNonexistent(t *testing.T) {
	manager := NewManager("test-cli")

	ctx := context.Background()
	token := &Token{AccessToken: "test"}
	_, err := manager.RefreshToken(ctx, "nonexistent", token)
	if err == nil {
		t.Error("RefreshToken() should return error for nonexistent authenticator")
	}
}

func TestManager_GetAuthenticatedClientDefaults(t *testing.T) {
	manager := NewManager("test-cli")

	config := &APIKeyConfig{
		Key:      "test-key",
		Name:     "X-API-Key",
		Location: APIKeyLocationHeader,
	}
	auth, _ := NewAPIKeyAuth(config)
	manager.RegisterAuthenticator("api", auth)

	// Test with empty auth name (should use default)
	client, err := manager.GetAuthenticatedClient("", "")
	if err != nil {
		t.Fatalf("GetAuthenticatedClient() failed: %v", err)
	}

	if client == nil {
		t.Error("GetAuthenticatedClient() returned nil")
	}
}

func TestManager_GetAuthenticatedClientNonexistent(t *testing.T) {
	manager := NewManager("test-cli")

	_, err := manager.GetAuthenticatedClient("nonexistent", "")
	if err == nil {
		t.Error("GetAuthenticatedClient() should return error for nonexistent authenticator")
	}
}

func TestManager_RegisterInvalidAuthenticator(t *testing.T) {
	manager := NewManager("test-cli")

	// Create invalid authenticator (missing required fields)
	config := &APIKeyConfig{
		Location: APIKeyLocationHeader,
		// Missing required Name field
	}

	// NewAPIKeyAuth should fail for invalid config
	auth, err := NewAPIKeyAuth(config)
	if err == nil {
		// If NewAPIKeyAuth doesn't fail, registering should fail
		err = manager.RegisterAuthenticator("invalid", auth)
		if err == nil {
			t.Error("RegisterAuthenticator() should return error for invalid authenticator")
		}
	}
	// Test is successful if either NewAPIKeyAuth or RegisterAuthenticator fails
}
