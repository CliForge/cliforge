package auth

import (
	"context"
	"testing"

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
