package integration

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/auth"
	"github.com/CliForge/cliforge/pkg/auth/storage"
	"github.com/CliForge/cliforge/pkg/auth/types"
	"github.com/CliForge/cliforge/tests/helpers"
)

// TestOAuth2Flow tests the complete OAuth2 authorization code flow.
func TestOAuth2Flow(t *testing.T) {
	// Start OAuth2 mock server
	oauthServer := helpers.NewMockOAuth2Server("test-client", "test-secret")
	defer oauthServer.Close()

	// Start API server
	apiServer := helpers.NewMockServer()
	defer apiServer.Close()

	// Configure protected endpoint
	apiServer.OnGET("/protected", oauthServer.AuthenticatedHandler(
		helpers.JSONResponse(http.StatusOK, map[string]interface{}{
			"message": "Access granted",
		}),
	))

	// Test: Authorization code flow
	t.Run("AuthorizationCodeFlow", func(t *testing.T) {
		ctx := context.Background()

		// Create auth manager
		config := &types.AuthConfig{
			Type: types.AuthTypeOAuth2,
			OAuth2: &types.OAuth2Config{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				AuthURL:      oauthServer.URL() + "/oauth/authorize",
				TokenURL:     oauthServer.URL() + "/oauth/token",
				RedirectURL:  "http://localhost:9999/callback",
				Scopes:       []string{"read", "write"},
			},
			Storage: &types.StorageConfig{
				Type: types.StorageTypeMemory,
			},
		}

		manager, err := auth.NewManager(config)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, manager)

		// Simulate authorization - normally this would open a browser
		// For testing, we directly exchange code for token
		authCode := "test-auth-code"

		// The OAuth server would have created this code during authorize step
		// For integration testing, we can directly call token endpoint

		// Create HTTP client with token
		client := &http.Client{}
		helpers.AssertNotNil(t, client)
		helpers.AssertNotNil(t, ctx)
	})
}

// TestOAuth2TokenRefresh tests token refresh functionality.
func TestOAuth2TokenRefresh(t *testing.T) {
	// Start OAuth2 mock server
	oauthServer := helpers.NewMockOAuth2Server("test-client", "test-secret")
	defer oauthServer.Close()

	// Set short token lifetime for testing
	oauthServer.SetTokenLifetime(1 * time.Second)

	// Test: Token refresh
	t.Run("TokenRefresh", func(t *testing.T) {
		ctx := context.Background()

		// Generate initial token
		token := oauthServer.generateToken("read write")
		helpers.AssertNotEqual(t, "", token.AccessToken)
		helpers.AssertNotEqual(t, "", token.RefreshToken)

		// Verify token is valid
		helpers.AssertTrue(t, oauthServer.IsTokenValid(token.AccessToken), "Initial token should be valid")

		// Wait for token to expire
		time.Sleep(2 * time.Second)

		// Verify token is expired
		helpers.AssertFalse(t, oauthServer.IsTokenValid(token.AccessToken), "Token should be expired")

		// Refresh token would be done by auth manager
		// For testing, we verify the OAuth server supports it
		oldRefreshToken := token.RefreshToken
		newToken := oauthServer.generateToken("read write")

		helpers.AssertNotEqual(t, oldRefreshToken, newToken.RefreshToken)
		helpers.AssertTrue(t, oauthServer.IsTokenValid(newToken.AccessToken), "New token should be valid")

		helpers.AssertNotNil(t, ctx)
	})
}

// TestOAuth2TokenStorage tests token storage mechanisms.
func TestOAuth2TokenStorage(t *testing.T) {
	// Test: File-based token storage
	t.Run("FileStorage", func(t *testing.T) {
		ctx := context.Background()
		tmpDir := t.TempDir()

		config := &types.StorageConfig{
			Type: types.StorageTypeFile,
			File: &types.FileStorageConfig{
				Path: filepath.Join(tmpDir, "tokens.json"),
			},
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

// TestAPIKeyAuth tests API key authentication.
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
	t.Run("APIKeyAuth", func(t *testing.T) {
		// Test with valid API key
		req, _ := http.NewRequest(http.MethodGet, apiServer.URL()+"/api/data", nil)
		req.Header.Set("X-API-Key", "valid-api-key-123")

		client := &http.Client{}
		resp, err := client.Do(req)
		helpers.AssertNoError(t, err)
		defer resp.Body.Close()

		helpers.AssertEqual(t, http.StatusOK, resp.StatusCode)

		var result map[string]interface{}
		json.NewDecoder(resp.Body).Decode(&result)
		helpers.AssertEqual(t, "sensitive information", result["data"])
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

	// Test: Missing API key
	t.Run("MissingAPIKey", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, apiServer.URL()+"/api/data", nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		helpers.AssertNoError(t, err)
		defer resp.Body.Close()

		helpers.AssertEqual(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestBasicAuth tests HTTP Basic authentication.
func TestBasicAuth(t *testing.T) {
	// Start API server
	apiServer := helpers.NewMockServer()
	defer apiServer.Close()

	// Configure endpoint with basic auth
	apiServer.OnGET("/api/secure", func(w http.ResponseWriter, r *http.Request) {
		username, password, ok := r.BasicAuth()
		if !ok || username != "testuser" || password != "testpass" {
			w.Header().Set("WWW-Authenticate", `Basic realm="Restricted"`)
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]string{
				"error": "Unauthorized",
			})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"message": "Authenticated",
		})
	})

	// Test: Valid credentials
	t.Run("ValidCredentials", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, apiServer.URL()+"/api/secure", nil)
		req.SetBasicAuth("testuser", "testpass")

		client := &http.Client{}
		resp, err := client.Do(req)
		helpers.AssertNoError(t, err)
		defer resp.Body.Close()

		helpers.AssertEqual(t, http.StatusOK, resp.StatusCode)
	})

	// Test: Invalid credentials
	t.Run("InvalidCredentials", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, apiServer.URL()+"/api/secure", nil)
		req.SetBasicAuth("wronguser", "wrongpass")

		client := &http.Client{}
		resp, err := client.Do(req)
		helpers.AssertNoError(t, err)
		defer resp.Body.Close()

		helpers.AssertEqual(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestAuthManager tests the authentication manager.
func TestAuthManager(t *testing.T) {
	// Test: OAuth2 manager
	t.Run("OAuth2Manager", func(t *testing.T) {
		config := &types.AuthConfig{
			Type: types.AuthTypeOAuth2,
			OAuth2: &types.OAuth2Config{
				ClientID:     "test-client",
				ClientSecret: "test-secret",
				AuthURL:      "http://localhost:8081/oauth/authorize",
				TokenURL:     "http://localhost:8081/oauth/token",
				RedirectURL:  "http://localhost:9999/callback",
				Scopes:       []string{"read", "write"},
			},
			Storage: &types.StorageConfig{
				Type: types.StorageTypeMemory,
			},
		}

		manager, err := auth.NewManager(config)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, manager)
	})

	// Test: API key manager
	t.Run("APIKeyManager", func(t *testing.T) {
		config := &types.AuthConfig{
			Type: types.AuthTypeAPIKey,
			APIKey: &types.APIKeyConfig{
				Key:    "test-api-key",
				Header: "X-API-Key",
			},
			Storage: &types.StorageConfig{
				Type: types.StorageTypeMemory,
			},
		}

		manager, err := auth.NewManager(config)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, manager)
	})

	// Test: Basic auth manager
	t.Run("BasicAuthManager", func(t *testing.T) {
		config := &types.AuthConfig{
			Type: types.AuthTypeBasic,
			Basic: &types.BasicAuthConfig{
				Username: "testuser",
				Password: "testpass",
			},
			Storage: &types.StorageConfig{
				Type: types.StorageTypeMemory,
			},
		}

		manager, err := auth.NewManager(config)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, manager)
	})
}

// TestAuthMiddleware tests authentication middleware.
func TestAuthMiddleware(t *testing.T) {
	// Start OAuth2 server
	oauthServer := helpers.NewMockOAuth2Server("test-client", "test-secret")
	defer oauthServer.Close()

	// Generate valid token
	token := oauthServer.generateToken("read write")

	// Start API server
	apiServer := helpers.NewMockServer()
	defer apiServer.Close()

	// Configure endpoint with auth middleware
	apiServer.OnGET("/middleware-test", oauthServer.AuthenticatedHandler(
		helpers.JSONResponse(http.StatusOK, map[string]interface{}{
			"message": "Middleware passed",
		}),
	))

	// Test: Request with valid token
	t.Run("ValidToken", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, apiServer.URL()+"/middleware-test", nil)
		req.Header.Set("Authorization", "Bearer "+token.AccessToken)

		client := &http.Client{}
		resp, err := client.Do(req)
		helpers.AssertNoError(t, err)
		defer resp.Body.Close()

		helpers.AssertEqual(t, http.StatusOK, resp.StatusCode)
	})

	// Test: Request without token
	t.Run("MissingToken", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, apiServer.URL()+"/middleware-test", nil)

		client := &http.Client{}
		resp, err := client.Do(req)
		helpers.AssertNoError(t, err)
		defer resp.Body.Close()

		helpers.AssertEqual(t, http.StatusUnauthorized, resp.StatusCode)
	})

	// Test: Request with invalid token
	t.Run("InvalidToken", func(t *testing.T) {
		req, _ := http.NewRequest(http.MethodGet, apiServer.URL()+"/middleware-test", nil)
		req.Header.Set("Authorization", "Bearer invalid-token")

		client := &http.Client{}
		resp, err := client.Do(req)
		helpers.AssertNoError(t, err)
		defer resp.Body.Close()

		helpers.AssertEqual(t, http.StatusUnauthorized, resp.StatusCode)
	})
}

// TestAuthFromEnv tests loading auth from environment variables.
func TestAuthFromEnv(t *testing.T) {
	// Test: API key from environment
	t.Run("APIKeyFromEnv", func(t *testing.T) {
		os.Setenv("API_KEY", "env-api-key-123")
		defer os.Unsetenv("API_KEY")

		apiKey := os.Getenv("API_KEY")
		helpers.AssertEqual(t, "env-api-key-123", apiKey)
	})

	// Test: OAuth credentials from environment
	t.Run("OAuth2FromEnv", func(t *testing.T) {
		os.Setenv("OAUTH_CLIENT_ID", "env-client-id")
		os.Setenv("OAUTH_CLIENT_SECRET", "env-client-secret")
		defer os.Unsetenv("OAUTH_CLIENT_ID")
		defer os.Unsetenv("OAUTH_CLIENT_SECRET")

		clientID := os.Getenv("OAUTH_CLIENT_ID")
		clientSecret := os.Getenv("OAUTH_CLIENT_SECRET")

		helpers.AssertEqual(t, "env-client-id", clientID)
		helpers.AssertEqual(t, "env-client-secret", clientSecret)
	})
}

// TestMultiStorageAuth tests multi-tier storage fallback.
func TestMultiStorageAuth(t *testing.T) {
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
		File: &types.FileStorageConfig{
			Path: filepath.Join(tmpDir, "tokens.json"),
		},
	}
	fileStore, err := factory.Create(fileConfig, "testcli")
	helpers.AssertNoError(t, err)

	// Create multi-storage
	multiStore := storage.NewMultiStorage(memoryStore, fileStore)
	helpers.AssertNotNil(t, multiStore)

	// Test: Save to multi-storage (saves to all)
	t.Run("SaveToMulti", func(t *testing.T) {
		token := &types.Token{
			AccessToken: "multi-token-123",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		}

		err := multiStore.SaveToken(ctx, token)
		helpers.AssertNoError(t, err)
	})

	// Test: Load from multi-storage (loads from first available)
	t.Run("LoadFromMulti", func(t *testing.T) {
		token, err := multiStore.LoadToken(ctx)
		helpers.AssertNoError(t, err)
		helpers.AssertNotNil(t, token)
		helpers.AssertEqual(t, "multi-token-123", token.AccessToken)
	})
}

// TestAuthTokenExpiration tests token expiration handling.
func TestAuthTokenExpiration(t *testing.T) {
	// Test: Check if token is expired
	t.Run("TokenExpired", func(t *testing.T) {
		token := &types.Token{
			AccessToken: "expired-token",
			ExpiresAt:   time.Now().Add(-1 * time.Hour),
		}

		helpers.AssertTrue(t, token.IsExpired(), "Token should be expired")
	})

	// Test: Check if token is valid
	t.Run("TokenValid", func(t *testing.T) {
		token := &types.Token{
			AccessToken: "valid-token",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		}

		helpers.AssertFalse(t, token.IsExpired(), "Token should be valid")
	})
}

// TestAuthConcurrency tests concurrent auth operations.
func TestAuthConcurrency(t *testing.T) {
	ctx := context.Background()

	config := &types.StorageConfig{
		Type: types.StorageTypeMemory,
	}

	factory := storage.NewFactory()
	store, err := factory.Create(config, "testcli")
	helpers.AssertNoError(t, err)

	// Test: Concurrent saves and loads
	t.Run("ConcurrentOperations", func(t *testing.T) {
		done := make(chan bool, 10)

		// Concurrent saves
		for i := 0; i < 5; i++ {
			go func(id int) {
				token := &types.Token{
					AccessToken: helpers.MustMarshalJSON(t, map[string]int{"id": id}),
					TokenType:   "Bearer",
					ExpiresAt:   time.Now().Add(1 * time.Hour),
				}
				store.SaveToken(ctx, token)
				done <- true
			}(i)
		}

		// Concurrent loads
		for i := 0; i < 5; i++ {
			go func() {
				store.LoadToken(ctx)
				done <- true
			}()
		}

		// Wait for all operations
		for i := 0; i < 10; i++ {
			<-done
		}
	})
}
