package builtin

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/CliForge/cliforge/pkg/auth"
)

// Mock authenticator for testing
type mockAuthenticator struct {
	token *auth.Token
	err   error
}

func (m *mockAuthenticator) Type() auth.AuthType {
	return auth.AuthTypeAPIKey
}

func (m *mockAuthenticator) Authenticate(ctx context.Context) (*auth.Token, error) {
	if m.err != nil {
		return nil, m.err
	}
	if m.token == nil {
		return &auth.Token{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		}, nil
	}
	return m.token, nil
}

func (m *mockAuthenticator) RefreshToken(ctx context.Context, token *auth.Token) (*auth.Token, error) {
	if m.err != nil {
		return nil, m.err
	}
	return &auth.Token{
		AccessToken: "refreshed-token",
		TokenType:   "Bearer",
		ExpiresAt:   time.Now().Add(2 * time.Hour),
	}, nil
}

func (m *mockAuthenticator) GetHeaders(token *auth.Token) map[string]string {
	if token == nil {
		return nil
	}
	return map[string]string{
		"Authorization": token.TokenType + " " + token.AccessToken,
	}
}

func (m *mockAuthenticator) Validate() error {
	return nil
}

// Mock storage for testing
type mockStorage struct {
	token *auth.Token
	err   error
}

func (m *mockStorage) SaveToken(ctx context.Context, token *auth.Token) error {
	if m.err != nil {
		return m.err
	}
	m.token = token
	return nil
}

func (m *mockStorage) LoadToken(ctx context.Context) (*auth.Token, error) {
	if m.err != nil {
		return nil, m.err
	}
	return m.token, nil
}

func (m *mockStorage) DeleteToken(ctx context.Context) error {
	if m.err != nil {
		return m.err
	}
	m.token = nil
	return nil
}

func TestNewAuthCommand(t *testing.T) {
	output := &bytes.Buffer{}
	mgr := auth.NewManager("testcli")
	opts := &AuthOptions{
		AuthManager: mgr,
		Output:      output,
	}

	cmd := NewAuthCommand(opts)

	if cmd == nil {
		t.Fatal("expected command, got nil")
	}

	if cmd.Use != "auth" {
		t.Errorf("expected Use 'auth', got %q", cmd.Use)
	}

	// Check subcommands exist
	subcommands := []string{"login", "logout", "status", "refresh"}
	for _, subcmd := range subcommands {
		found := false
		for _, c := range cmd.Commands() {
			if c.Name() == subcmd {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected subcommand %q not found", subcmd)
		}
	}
}

func TestRunAuthLogin_Success(t *testing.T) {
	output := &bytes.Buffer{}
	mgr := auth.NewManager("testcli")

	// Register mock authenticator and storage
	mockAuth := &mockAuthenticator{}
	mockStore := &mockStorage{}

	if err := mgr.RegisterAuthenticator("default", mockAuth); err != nil {
		t.Fatalf("failed to register authenticator: %v", err)
	}
	mgr.RegisterStorage("default", mockStore)

	opts := &AuthOptions{
		AuthManager: mgr,
		Output:      output,
	}

	err := runAuthLogin(opts, "")
	if err != nil {
		t.Fatalf("runAuthLogin failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Authentication successful") {
		t.Errorf("expected success message in output, got: %s", result)
	}

	// Verify token was stored
	if mockStore.token == nil {
		t.Error("expected token to be stored")
	}
}

func TestRunAuthLogin_AuthenticationFailed(t *testing.T) {
	output := &bytes.Buffer{}
	mgr := auth.NewManager("testcli")

	// Register mock authenticator that returns error
	mockAuth := &mockAuthenticator{err: errors.New("auth failed")}
	mockStore := &mockStorage{}

	_ = mgr.RegisterAuthenticator("default", mockAuth)
	mgr.RegisterStorage("default", mockStore)

	opts := &AuthOptions{
		AuthManager: mgr,
		Output:      output,
	}

	err := runAuthLogin(opts, "")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "authentication failed") {
		t.Errorf("expected 'authentication failed' error, got: %v", err)
	}
}

func TestRunAuthLogin_UnknownAuthType(t *testing.T) {
	output := &bytes.Buffer{}
	mgr := auth.NewManager("testcli")

	opts := &AuthOptions{
		AuthManager: mgr,
		Output:      output,
	}

	err := runAuthLogin(opts, "unknown-type")
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "unknown authentication type") {
		t.Errorf("expected 'unknown authentication type' error, got: %v", err)
	}
}

func TestRunAuthLogout_Success(t *testing.T) {
	output := &bytes.Buffer{}
	mgr := auth.NewManager("testcli")

	// Register mock storage with a token
	mockStore := &mockStorage{
		token: &auth.Token{AccessToken: "test-token"},
	}

	mgr.RegisterStorage("default", mockStore)

	opts := &AuthOptions{
		AuthManager: mgr,
		Output:      output,
	}

	err := runAuthLogout(opts)
	if err != nil {
		t.Fatalf("runAuthLogout failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Logged out") {
		t.Errorf("expected 'Logged out' in output, got: %s", result)
	}

	// Verify token was deleted
	if mockStore.token != nil {
		t.Error("expected token to be deleted")
	}
}

func TestRunAuthLogout_NoCredentials(t *testing.T) {
	output := &bytes.Buffer{}
	mgr := auth.NewManager("testcli")

	// Register mock storage without a token
	mockStore := &mockStorage{err: errors.New("no token")}
	mgr.RegisterStorage("default", mockStore)

	opts := &AuthOptions{
		AuthManager: mgr,
		Output:      output,
	}

	err := runAuthLogout(opts)
	if err != nil {
		t.Fatalf("runAuthLogout failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "No credentials found") {
		t.Errorf("expected 'No credentials found' in output, got: %s", result)
	}
}

func TestRunAuthStatus_Authenticated(t *testing.T) {
	output := &bytes.Buffer{}
	mgr := auth.NewManager("testcli")

	// Register mock storage with a valid token
	mockStore := &mockStorage{
		token: &auth.Token{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(1 * time.Hour),
		},
	}

	mgr.RegisterStorage("default", mockStore)

	opts := &AuthOptions{
		AuthManager: mgr,
		Output:      output,
	}

	err := runAuthStatus(opts)
	if err != nil {
		t.Fatalf("runAuthStatus failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Authenticated") {
		t.Errorf("expected 'Authenticated' in output, got: %s", result)
	}
}

func TestRunAuthStatus_NotAuthenticated(t *testing.T) {
	output := &bytes.Buffer{}
	mgr := auth.NewManager("testcli")

	// Register mock storage without a token
	mockStore := &mockStorage{err: errors.New("no token")}
	mgr.RegisterStorage("default", mockStore)

	opts := &AuthOptions{
		AuthManager: mgr,
		Output:      output,
	}

	err := runAuthStatus(opts)
	if err != nil {
		t.Fatalf("runAuthStatus failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Not authenticated") {
		t.Errorf("expected 'Not authenticated' in output, got: %s", result)
	}

	if !strings.Contains(result, "auth login") {
		t.Errorf("expected 'auth login' suggestion in output, got: %s", result)
	}
}

func TestRunAuthStatus_ExpiredToken(t *testing.T) {
	output := &bytes.Buffer{}
	mgr := auth.NewManager("testcli")

	// Register mock storage with an expired token
	mockStore := &mockStorage{
		token: &auth.Token{
			AccessToken: "test-token",
			TokenType:   "Bearer",
			ExpiresAt:   time.Now().Add(-1 * time.Hour), // Expired
		},
	}

	mgr.RegisterStorage("default", mockStore)

	opts := &AuthOptions{
		AuthManager: mgr,
		Output:      output,
	}

	err := runAuthStatus(opts)
	if err != nil {
		t.Fatalf("runAuthStatus failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Expired") {
		t.Errorf("expected 'Expired' in output, got: %s", result)
	}
}

func TestRunAuthRefresh_Success(t *testing.T) {
	output := &bytes.Buffer{}
	mgr := auth.NewManager("testcli")

	// Register mock authenticator and storage
	mockAuth := &mockAuthenticator{}
	mockStore := &mockStorage{
		token: &auth.Token{
			AccessToken:  "old-token",
			RefreshToken: "refresh-token",
		},
	}

	_ = mgr.RegisterAuthenticator("default", mockAuth)
	mgr.RegisterStorage("default", mockStore)

	opts := &AuthOptions{
		AuthManager: mgr,
		Output:      output,
	}

	err := runAuthRefresh(opts)
	if err != nil {
		t.Fatalf("runAuthRefresh failed: %v", err)
	}

	result := output.String()
	if !strings.Contains(result, "Refreshed") {
		t.Errorf("expected 'Refreshed' in output, got: %s", result)
	}

	// Verify new token was stored
	if mockStore.token.AccessToken != "refreshed-token" {
		t.Errorf("expected token to be refreshed, got: %s", mockStore.token.AccessToken)
	}
}

func TestRunAuthRefresh_NoTokensToRefresh(t *testing.T) {
	output := &bytes.Buffer{}
	mgr := auth.NewManager("testcli")

	// Register mock storage without a token
	mockAuth := &mockAuthenticator{}
	mockStore := &mockStorage{err: errors.New("no token")}

	_ = mgr.RegisterAuthenticator("default", mockAuth)
	mgr.RegisterStorage("default", mockStore)

	opts := &AuthOptions{
		AuthManager: mgr,
		Output:      output,
	}

	err := runAuthRefresh(opts)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	if !strings.Contains(err.Error(), "no tokens were refreshed") {
		t.Errorf("expected 'no tokens were refreshed' error, got: %v", err)
	}
}

func TestRunAuthRefresh_RefreshFailed(t *testing.T) {
	output := &bytes.Buffer{}
	mgr := auth.NewManager("testcli")

	// Register mock authenticator that fails to refresh
	mockAuth := &mockAuthenticator{err: errors.New("refresh failed")}
	mockStore := &mockStorage{
		token: &auth.Token{
			AccessToken:  "old-token",
			RefreshToken: "refresh-token",
		},
	}

	_ = mgr.RegisterAuthenticator("default", mockAuth)
	mgr.RegisterStorage("default", mockStore)

	opts := &AuthOptions{
		AuthManager: mgr,
		Output:      output,
	}

	err := runAuthRefresh(opts)
	if err == nil {
		t.Fatal("expected error, got nil")
	}

	result := output.String()
	if !strings.Contains(result, "failed to refresh") {
		t.Errorf("expected 'failed to refresh' warning in output, got: %s", result)
	}
}
