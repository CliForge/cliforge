package auth

import (
	"context"
	"errors"
	"testing"

	"github.com/CliForge/cliforge/pkg/auth/types"
)

// mockTokenStorage is a mock implementation of TokenStorage for testing
type mockTokenStorage struct {
	token *types.Token
	err   error
}

func (m *mockTokenStorage) SaveToken(ctx context.Context, token *types.Token) error {
	m.token = token
	return nil
}

func (m *mockTokenStorage) LoadToken(ctx context.Context) (*types.Token, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.token, nil
}

func (m *mockTokenStorage) DeleteToken(ctx context.Context) error {
	m.token = nil
	return nil
}

func TestTokenResolver_FlagTakesPrecedence(t *testing.T) {
	// Set environment variables that should be ignored
	t.Setenv("ROSA_TOKEN", "env-rosa-token")
	t.Setenv("OCM_TOKEN", "env-ocm-token")

	storage := &mockTokenStorage{
		token: &types.Token{AccessToken: "storage-token"},
	}

	resolver := NewTokenResolver(
		WithFlagToken("flag-token"),
		WithStorage(storage),
		WithPromptFunc(func() (string, error) {
			return "prompt-token", nil
		}),
	)

	token, source, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "flag-token" {
		t.Errorf("Resolve() token = %v, want flag-token", token)
	}

	if source != TokenSourceFlag {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourceFlag)
	}
}

func TestTokenResolver_EnvROSA_Token(t *testing.T) {
	t.Setenv("ROSA_TOKEN", "rosa-env-token")
	t.Setenv("OCM_TOKEN", "ocm-env-token")

	storage := &mockTokenStorage{
		token: &types.Token{AccessToken: "storage-token"},
	}

	resolver := NewTokenResolver(
		WithStorage(storage),
		WithPromptFunc(func() (string, error) {
			return "prompt-token", nil
		}),
	)

	token, source, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "rosa-env-token" {
		t.Errorf("Resolve() token = %v, want rosa-env-token", token)
	}

	if source != TokenSourceEnvRosa {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourceEnvRosa)
	}
}

func TestTokenResolver_EnvOCM_Token(t *testing.T) {
	// Only set OCM_TOKEN, not ROSA_TOKEN
	t.Setenv("OCM_TOKEN", "ocm-env-token")

	storage := &mockTokenStorage{
		token: &types.Token{AccessToken: "storage-token"},
	}

	resolver := NewTokenResolver(
		WithStorage(storage),
		WithPromptFunc(func() (string, error) {
			return "prompt-token", nil
		}),
	)

	token, source, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "ocm-env-token" {
		t.Errorf("Resolve() token = %v, want ocm-env-token", token)
	}

	if source != TokenSourceEnvOCM {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourceEnvOCM)
	}
}

func TestTokenResolver_ROSABeforeOCM(t *testing.T) {
	// Both environment variables set - ROSA should take precedence
	t.Setenv("ROSA_TOKEN", "rosa-token")
	t.Setenv("OCM_TOKEN", "ocm-token")

	resolver := NewTokenResolver()

	token, source, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "rosa-token" {
		t.Errorf("Resolve() token = %v, want rosa-token", token)
	}

	if source != TokenSourceEnvRosa {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourceEnvRosa)
	}
}

func TestTokenResolver_FileStorage(t *testing.T) {
	// No flag, no env vars - should use storage
	storage := &mockTokenStorage{
		token: &types.Token{AccessToken: "storage-token"},
	}

	resolver := NewTokenResolver(
		WithStorage(storage),
		WithPromptFunc(func() (string, error) {
			return "prompt-token", nil
		}),
	)

	token, source, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "storage-token" {
		t.Errorf("Resolve() token = %v, want storage-token", token)
	}

	if source != TokenSourceFile {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourceFile)
	}
}

func TestTokenResolver_PromptFallback(t *testing.T) {
	// No other sources available - should use prompt
	resolver := NewTokenResolver(
		WithPromptFunc(func() (string, error) {
			return "prompt-token", nil
		}),
	)

	token, source, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "prompt-token" {
		t.Errorf("Resolve() token = %v, want prompt-token", token)
	}

	if source != TokenSourcePrompt {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourcePrompt)
	}
}

func TestTokenResolver_NoTokenFound(t *testing.T) {
	// No token sources configured
	resolver := NewTokenResolver()

	token, source, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "" {
		t.Errorf("Resolve() token = %v, want empty string", token)
	}

	if source != TokenSourceNone {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourceNone)
	}
}

func TestTokenResolver_EmptyFlagChecksNext(t *testing.T) {
	// Empty flag should not count - should continue to env vars
	t.Setenv("ROSA_TOKEN", "rosa-token")

	resolver := NewTokenResolver(
		WithFlagToken(""), // Empty flag
	)

	token, source, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "rosa-token" {
		t.Errorf("Resolve() token = %v, want rosa-token", token)
	}

	if source != TokenSourceEnvRosa {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourceEnvRosa)
	}
}

func TestTokenResolver_ContextCancellation(t *testing.T) {
	// Context cancellation should be respected
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	storage := &mockTokenStorage{
		token: &types.Token{AccessToken: "storage-token"},
	}

	resolver := NewTokenResolver(
		WithStorage(storage),
	)

	// Even with cancelled context, resolve should work (it doesn't block)
	// This test verifies the resolver doesn't panic with cancelled context
	token, source, err := resolver.Resolve(ctx)
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "storage-token" {
		t.Errorf("Resolve() token = %v, want storage-token", token)
	}

	if source != TokenSourceFile {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourceFile)
	}
}

func TestTokenResolver_PromptError(t *testing.T) {
	// Prompt function returns error
	expectedErr := errors.New("user cancelled")

	resolver := NewTokenResolver(
		WithPromptFunc(func() (string, error) {
			return "", expectedErr
		}),
	)

	token, source, err := resolver.Resolve(context.Background())
	if err == nil {
		t.Fatal("Resolve() expected error, got nil")
	}

	if !errors.Is(err, expectedErr) {
		t.Errorf("Resolve() error = %v, want wrapped error containing %v", err, expectedErr)
	}

	if token != "" {
		t.Errorf("Resolve() token = %v, want empty string", token)
	}

	if source != TokenSourceNone {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourceNone)
	}
}

func TestTokenResolver_StorageError(t *testing.T) {
	// Storage error should continue to prompt
	storage := &mockTokenStorage{
		err: errors.New("storage read error"),
	}

	resolver := NewTokenResolver(
		WithStorage(storage),
		WithPromptFunc(func() (string, error) {
			return "prompt-token", nil
		}),
	)

	token, source, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "prompt-token" {
		t.Errorf("Resolve() token = %v, want prompt-token", token)
	}

	if source != TokenSourcePrompt {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourcePrompt)
	}
}

func TestTokenResolver_StorageEmptyToken(t *testing.T) {
	// Storage returns token with empty AccessToken - should continue to prompt
	storage := &mockTokenStorage{
		token: &types.Token{AccessToken: ""}, // Empty access token
	}

	resolver := NewTokenResolver(
		WithStorage(storage),
		WithPromptFunc(func() (string, error) {
			return "prompt-token", nil
		}),
	)

	token, source, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "prompt-token" {
		t.Errorf("Resolve() token = %v, want prompt-token", token)
	}

	if source != TokenSourcePrompt {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourcePrompt)
	}
}

func TestTokenResolver_CustomEnvVars(t *testing.T) {
	// Test custom environment variable names
	t.Setenv("CUSTOM_ROSA", "custom-rosa-token")
	t.Setenv("CUSTOM_OCM", "custom-ocm-token")

	resolver := NewTokenResolver(
		WithEnvVars("CUSTOM_ROSA", "CUSTOM_OCM"),
	)

	token, source, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "custom-rosa-token" {
		t.Errorf("Resolve() token = %v, want custom-rosa-token", token)
	}

	if source != TokenSourceEnvRosa {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourceEnvRosa)
	}
}

func TestTokenResolver_PromptReturnsEmpty(t *testing.T) {
	// Prompt returns empty string (user skipped) - should return none
	resolver := NewTokenResolver(
		WithPromptFunc(func() (string, error) {
			return "", nil // Empty but no error
		}),
	)

	token, source, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "" {
		t.Errorf("Resolve() token = %v, want empty string", token)
	}

	if source != TokenSourceNone {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourceNone)
	}
}

func TestTokenResolver_NoStorageConfigured(t *testing.T) {
	// Storage is nil - should skip to prompt
	resolver := NewTokenResolver(
		WithStorage(nil),
		WithPromptFunc(func() (string, error) {
			return "prompt-token", nil
		}),
	)

	token, source, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "prompt-token" {
		t.Errorf("Resolve() token = %v, want prompt-token", token)
	}

	if source != TokenSourcePrompt {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourcePrompt)
	}
}

func TestTokenResolver_StorageReturnsNil(t *testing.T) {
	// Storage returns nil token - should continue to prompt
	storage := &mockTokenStorage{
		token: nil,
	}

	resolver := NewTokenResolver(
		WithStorage(storage),
		WithPromptFunc(func() (string, error) {
			return "prompt-token", nil
		}),
	)

	token, source, err := resolver.Resolve(context.Background())
	if err != nil {
		t.Fatalf("Resolve() error = %v", err)
	}

	if token != "prompt-token" {
		t.Errorf("Resolve() token = %v, want prompt-token", token)
	}

	if source != TokenSourcePrompt {
		t.Errorf("Resolve() source = %v, want %v", source, TokenSourcePrompt)
	}
}
