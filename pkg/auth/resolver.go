package auth

import (
	"context"
	"fmt"
	"os"
)

// TokenSource represents where a token was found
type TokenSource string

const (
	TokenSourceFlag    TokenSource = "flag"
	TokenSourceEnvRosa TokenSource = "env:ROSA_TOKEN"
	TokenSourceEnvOCM  TokenSource = "env:OCM_TOKEN"
	TokenSourceFile    TokenSource = "file"
	TokenSourcePrompt  TokenSource = "prompt"
	TokenSourceNone    TokenSource = "none"
)

// TokenResolver finds tokens from multiple sources following ROSA's precedence
type TokenResolver struct {
	flagToken  string
	envVarRosa string // default: "ROSA_TOKEN"
	envVarOCM  string // default: "OCM_TOKEN"
	storage    TokenStorage
	promptFunc func() (string, error)
}

// TokenResolverOption configures the resolver
type TokenResolverOption func(*TokenResolver)

// NewTokenResolver creates a resolver with the specified options
func NewTokenResolver(opts ...TokenResolverOption) *TokenResolver {
	r := &TokenResolver{
		envVarRosa: "ROSA_TOKEN",
		envVarOCM:  "OCM_TOKEN",
	}

	for _, opt := range opts {
		opt(r)
	}

	return r
}

// WithFlagToken sets token from --token flag
func WithFlagToken(token string) TokenResolverOption {
	return func(r *TokenResolver) {
		r.flagToken = token
	}
}

// WithEnvVars sets custom env var names
func WithEnvVars(rosa, ocm string) TokenResolverOption {
	return func(r *TokenResolver) {
		if rosa != "" {
			r.envVarRosa = rosa
		}
		if ocm != "" {
			r.envVarOCM = ocm
		}
	}
}

// WithStorage sets the token storage for file lookups
func WithStorage(storage TokenStorage) TokenResolverOption {
	return func(r *TokenResolver) {
		r.storage = storage
	}
}

// WithPromptFunc sets the interactive prompt function
func WithPromptFunc(fn func() (string, error)) TokenResolverOption {
	return func(r *TokenResolver) {
		r.promptFunc = fn
	}
}

// Resolve finds a token using ROSA's precedence order
// Order: flag → ROSA_TOKEN → OCM_TOKEN → file → prompt
// Returns (token, source, error)
func (r *TokenResolver) Resolve(ctx context.Context) (string, TokenSource, error) {
	// 1. Check flag token first (highest priority)
	if r.flagToken != "" {
		return r.flagToken, TokenSourceFlag, nil
	}

	// 2. Check ROSA_TOKEN environment variable
	if token := os.Getenv(r.envVarRosa); token != "" {
		return token, TokenSourceEnvRosa, nil
	}

	// 3. Check OCM_TOKEN environment variable
	if token := os.Getenv(r.envVarOCM); token != "" {
		return token, TokenSourceEnvOCM, nil
	}

	// 4. Try to load from file storage
	if r.storage != nil {
		token, err := r.storage.LoadToken(ctx)
		if err == nil && token != nil && token.AccessToken != "" {
			return token.AccessToken, TokenSourceFile, nil
		}
		// Continue to next source on storage error
	}

	// 5. Try interactive prompt as last resort
	if r.promptFunc != nil {
		token, err := r.promptFunc()
		if err != nil {
			return "", TokenSourceNone, fmt.Errorf("prompt failed: %w", err)
		}
		if token != "" {
			return token, TokenSourcePrompt, nil
		}
	}

	// 6. No token found from any source
	return "", TokenSourceNone, nil
}
