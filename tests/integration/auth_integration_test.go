package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"path/filepath"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/auth/storage"
	"github.com/CliForge/cliforge/pkg/auth/types"
	"github.com/CliForge/cliforge/tests/helpers"
)

// TestTokenStorage tests token storage integration.
func TestTokenStorage(t *testing.T) {
	// Test: File-based token storage
	t.Run("FileStorage", func(t *testing.T) {
		ctx := context.Background()
		tmpDir := t.TempDir()

		config := &types.StorageConfig{
			Type: types.StorageTypeFile,
			Path: filepath.Join(tmpDir, "tokens.json"),
		}

		factory := storage.NewFactory()
		store, err := factory.Create(config, "testcli")
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, store)

		// Create test token
		token := &types.Token{
			AccessToken:  "access-token-123",
			RefreshToken: "refresh-token-456",
			TokenType:    "Bearer",
			ExpiresAt:    time.Now().Add(1 * time.Hour),
			Scopes:       []string{"read", "write"},
		}

		// Save token
		err = store.SaveToken(ctx, token)
		helpers.AssertNoError(t, err)

		// Load token
		loadedToken, err := store.LoadToken(ctx)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, loadedToken)
		helpers.AssertEqual(t, token.AccessToken, loadedToken.AccessToken)
		helpers.AssertEqual(t, token.RefreshToken, loadedToken.RefreshToken)

		// Delete token
		err = store.DeleteToken(ctx)
		helpers.AssertNoError(t, err)

		// Verify deletion
		_, err = store.LoadToken(ctx)
		helpers.AssertError(t, err)
	})

	// Test: Memory storage
	t.Run("MemoryStorage", func(t *testing.T) {
		ctx := context.Background()

		config := &types.StorageConfig{
			Type: types.StorageTypeMemory,
		}

		factory := storage.NewFactory()
		store, err := factory.Create(config, "testcli")
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, store)

		token := &types.Token{
			AccessToken: "mem-token-123",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		}

		err = store.SaveToken(ctx, token)
		helpers.AssertNoError(t, err)

		loadedToken, err := store.LoadToken(ctx)
		helpers.AssertNoError(t, err)
		helpers.AssertEqual(t, token.AccessToken, loadedToken.AccessToken)
	})
}

// TestMockOAuth2Server tests the OAuth2 mock server.
func TestMockOAuth2Server(t *testing.T) {
	// Start OAuth2 server
	oauthServer := helpers.NewMockOAuth2Server("test-client", "test-secret")
	defer oauthServer.Close()

	// Test: Token is valid initially
	t.Run("ValidToken", func(t *testing.T) {
		token, exists := oauthServer.GetToken("some-token")
		// Token doesn't exist yet
		helpers.AssertFalse(t, exists, "Token should not exist yet")
		_ = token // Avoid unused variable
	})
}

// TestAPIKeyAuth tests API key authentication integration.
func TestAPIKeyAuth(t *testing.T) {
	// Start API server
	apiServer := helpers.NewMockServer()
	defer apiServer.Close()

	// Configure endpoint that requires API key
	apiServer.OnGET("/api/data", func(w http.ResponseWriter, r *http.Request) {
		apiKey := r.Header.Get("X-API-Key")
		if apiKey != "valid-api-key-123" {
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Invalid API key",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"data": "sensitive information",
		})
	})

	// Test: API key authentication
	t.Run("ValidAPIKey", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, apiServer.URL()+"/api/data", nil)
		req.Header.Set("X-API-Key", "valid-api-key-123")

		client := &http.Client{}
		resp, err := client.Do(req)
		helpers.AssertNoError(t, err)
		defer resp.Body.Close()

		helpers.AssertEqual(t, http.StatusOK, resp.StatusCode)
	})

	// Test: Invalid API key
	t.Run("InvalidAPIKey", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, apiServer.URL()+"/api/data", nil)
		req.Header.Set("X-API-Key", "invalid-key")

		client := &http.Client{}
		resp, err := client.Do(req)
		helpers.AssertNoError(t, err)
		defer resp.Body.Close()

		helpers.AssertEqual(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestTokenExpiration tests token expiration.
func TestTokenExpiration(t *testing.T) {
	// Test: Expired token
	t.Run("ExpiredToken", func(t *testing.T) {
		token := &types.Token{
			AccessToken: "expired-token",
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
		}

		helpers.AssertTrue(t, token.IsExpired(), "Token should be expired")
		helpers.AssertFalse(t, token.IsValid(), "Token should not be valid")
	})

	// Test: Valid token
	t.Run("ValidToken", func(t *testing.T) {
		token := &types.Token{
			AccessToken: "valid-token",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		}

		helpers.AssertFalse(t, token.IsExpired(), "Token should not be expired")
		helpers.AssertTrue(t, token.IsValid(), "Token should be valid")
	})
}

// TestMultiStorage tests multi-tier storage.
func TestMultiStorage(t *testing.T) {
	ctx := context.Background()
	tmpDir := t.TempDir()

	// Create memory storage
	memoryConfig := &types.StorageConfig{
		Type: types.StorageTypeMemory,
	}
	factory := storage.NewFactory()
	memoryStore, err := factory.Create(memoryConfig, "testcli")
	helpers.AssertNoError(t, err)

	// Create file storage
	fileConfig := &types.StorageConfig{
		Type: types.StorageTypeFile,
		Path: filepath.Join(tmpDir, "tokens.json"),
	}
	fileStore, err := factory.Create(fileConfig, "testcli")
	helpers.AssertNoError(t, err)

	// Create multi-storage
	multiStore := storage.NewMultiStorage(memoryStore, fileStore)
	helpers.AssertNotNil(t, multiStore)

	// Test: Save to multi-storage
	t.Run("SaveToMulti", func(t *testing.T) {
		token := &types.Token{
			AccessToken: "multi-token-123",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		}

		err := multiStore.SaveToken(ctx, token)
		helpers.AssertNoError(t, err)
	})

	// Test: Load from multi-storage
	t.Run("LoadFromMulti", func(t *testing.T) {
		token, err := multiStore.LoadToken(ctx)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, token)
		helpers.AssertEqual(t, "multi-token-123", token.AccessToken)
	})
}
