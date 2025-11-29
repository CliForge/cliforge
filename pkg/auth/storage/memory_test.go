package storage

import (
	"context"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/auth/types"
)

func TestMemoryStorage_SaveAndLoad(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()

	token := &types.Token{
		AccessToken:  "test-access-token",
		RefreshToken: "test-refresh-token",
		TokenType:    "Bearer",
		ExpiresAt:    time.Now().Add(1 * time.Hour),
		Scopes:       []string{"read", "write"},
		Extra: map[string]interface{}{
			"custom": "value",
		},
	}

	// Save token
	err := storage.SaveToken(ctx, token)
	if err != nil {
		t.Fatalf("SaveToken() failed: %v", err)
	}

	// Load token
	loaded, err := storage.LoadToken(ctx)
	if err != nil {
		t.Fatalf("LoadToken() failed: %v", err)
	}

	// Verify token fields
	if loaded.AccessToken != token.AccessToken {
		t.Errorf("AccessToken = %v, want %v", loaded.AccessToken, token.AccessToken)
	}
	if loaded.RefreshToken != token.RefreshToken {
		t.Errorf("RefreshToken = %v, want %v", loaded.RefreshToken, token.RefreshToken)
	}
	if loaded.TokenType != token.TokenType {
		t.Errorf("TokenType = %v, want %v", loaded.TokenType, token.TokenType)
	}
	if len(loaded.Scopes) != len(token.Scopes) {
		t.Errorf("Scopes length = %v, want %v", len(loaded.Scopes), len(token.Scopes))
	}
	if loaded.Extra["custom"] != "value" {
		t.Errorf("Extra[custom] = %v, want %v", loaded.Extra["custom"], "value")
	}
}

func TestMemoryStorage_LoadNotFound(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()

	_, err := storage.LoadToken(ctx)
	if err == nil {
		t.Error("LoadToken() should return error when no token is stored")
	}
}

func TestMemoryStorage_Delete(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()

	token := &types.Token{
		AccessToken: "test-token",
	}

	// Save token
	err := storage.SaveToken(ctx, token)
	if err != nil {
		t.Fatalf("SaveToken() failed: %v", err)
	}

	// Delete token
	err = storage.DeleteToken(ctx)
	if err != nil {
		t.Fatalf("DeleteToken() failed: %v", err)
	}

	// Try to load - should fail
	_, err = storage.LoadToken(ctx)
	if err == nil {
		t.Error("LoadToken() should return error after deletion")
	}
}

func TestMemoryStorage_SaveNil(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()

	err := storage.SaveToken(ctx, nil)
	if err == nil {
		t.Error("SaveToken() should return error for nil token")
	}
}

func TestMemoryStorage_Clear(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()

	token := &types.Token{
		AccessToken: "test-token",
	}

	// Save token
	_ = storage.SaveToken(ctx, token)

	// Clear storage
	storage.Clear()

	// Try to load - should fail
	_, err := storage.LoadToken(ctx)
	if err == nil {
		t.Error("LoadToken() should return error after Clear()")
	}
}

func TestMemoryStorage_Concurrent(_ *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()

	// Run concurrent operations
	done := make(chan bool)
	for i := 0; i < 10; i++ {
		go func(id int) {
			token := &types.Token{
				AccessToken: "test-token",
				Extra: map[string]interface{}{
					"id": id,
				},
			}
			_ = storage.SaveToken(ctx, token)
			_, _ = storage.LoadToken(ctx)
			_ = storage.DeleteToken(ctx)
			done <- true
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestMemoryStorage_IsolatedCopies(t *testing.T) {
	storage := NewMemoryStorage()
	ctx := context.Background()

	token := &types.Token{
		AccessToken: "original",
		Scopes:      []string{"read"},
		Extra: map[string]interface{}{
			"key": "value",
		},
	}

	// Save token
	_ = storage.SaveToken(ctx, token)

	// Modify original token
	token.AccessToken = "modified"
	token.Scopes[0] = "write"
	token.Extra["key"] = "changed"

	// Load token - should have original values
	loaded, err := storage.LoadToken(ctx)
	if err != nil {
		t.Fatalf("LoadToken() failed: %v", err)
	}

	if loaded.AccessToken != "original" {
		t.Errorf("AccessToken = %v, want %v", loaded.AccessToken, "original")
	}
	if loaded.Scopes[0] != "read" {
		t.Errorf("Scopes[0] = %v, want %v", loaded.Scopes[0], "read")
	}
	if loaded.Extra["key"] != "value" {
		t.Errorf("Extra[key] = %v, want %v", loaded.Extra["key"], "value")
	}
}
