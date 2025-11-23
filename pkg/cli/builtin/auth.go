package builtin

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/CliForge/cliforge/pkg/auth"
	"github.com/spf13/cobra"
)

// AuthOptions configures the auth command behavior.
type AuthOptions struct {
	AuthManager *auth.Manager
	Output      io.Writer
}

// NewAuthCommand creates a new auth command group.
func NewAuthCommand(opts *AuthOptions) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Manage authentication",
		Long: `Manage authentication credentials.

Available subcommands:
  login   - Log in and store credentials
  logout  - Log out and remove credentials
  status  - Show authentication status
  refresh - Refresh authentication tokens`,
	}

	// Add subcommands
	cmd.AddCommand(newAuthLoginCommand(opts))
	cmd.AddCommand(newAuthLogoutCommand(opts))
	cmd.AddCommand(newAuthStatusCommand(opts))
	cmd.AddCommand(newAuthRefreshCommand(opts))

	return cmd
}

// newAuthLoginCommand creates the auth login subcommand.
func newAuthLoginCommand(opts *AuthOptions) *cobra.Command {
	var authType string

	cmd := &cobra.Command{
		Use:   "login",
		Short: "Log in and store credentials",
		Long: `Log in to the API and store credentials securely.

The login method depends on the API authentication configuration:
- api-key: Prompt for API key
- oauth2: Start OAuth2 flow
- basic: Prompt for username and password`,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthLogin(opts, authType)
		},
	}

	cmd.Flags().StringVar(&authType, "type", "", "Authentication type (api-key, oauth2, basic)")

	return cmd
}

// newAuthLogoutCommand creates the auth logout subcommand.
func newAuthLogoutCommand(opts *AuthOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "logout",
		Short: "Log out and remove credentials",
		Long:  "Remove stored authentication credentials.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthLogout(opts)
		},
	}
}

// newAuthStatusCommand creates the auth status subcommand.
func newAuthStatusCommand(opts *AuthOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "status",
		Short: "Show authentication status",
		Long:  "Display current authentication status and token information.",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthStatus(opts)
		},
	}
}

// newAuthRefreshCommand creates the auth refresh subcommand.
func newAuthRefreshCommand(opts *AuthOptions) *cobra.Command {
	return &cobra.Command{
		Use:   "refresh",
		Short: "Refresh authentication tokens",
		Long:  "Refresh authentication tokens (for OAuth2 and similar flows).",
		RunE: func(cmd *cobra.Command, args []string) error {
			return runAuthRefresh(opts)
		},
	}
}

// runAuthLogin performs the login flow.
func runAuthLogin(opts *AuthOptions, authType string) error {
	ctx := context.Background()

	// Determine auth type
	if authType == "" {
		// Use default authenticator
		authType = "default"
	}

	// Get authenticator
	authenticator, err := opts.AuthManager.GetAuthenticator(authType)
	if err != nil {
		return fmt.Errorf("unknown authentication type: %s", authType)
	}

	// Perform authentication
	fmt.Fprintln(opts.Output, "Starting authentication flow...")

	token, err := authenticator.Authenticate(ctx)
	if err != nil {
		return fmt.Errorf("authentication failed: %w", err)
	}

	// Store token
	storage, err := opts.AuthManager.GetStorage(authType)
	if err != nil {
		return fmt.Errorf("no storage configured for authentication type: %s", authType)
	}

	if err := storage.SaveToken(ctx, token); err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	fmt.Fprintln(opts.Output, "✓ Authentication successful")
	fmt.Fprintln(opts.Output, "Credentials stored securely")

	return nil
}

// runAuthLogout performs the logout flow.
func runAuthLogout(opts *AuthOptions) error {
	ctx := context.Background()

	// Get all registered storage backends
	storages := opts.AuthManager.ListStorages()

	removed := false
	for _, name := range storages {
		storage, err := opts.AuthManager.GetStorage(name)
		if err != nil {
			continue
		}

		// Check if credentials exist
		_, err = storage.LoadToken(ctx)
		if err == nil {
			// Remove credentials
			if err := storage.DeleteToken(ctx); err != nil {
				fmt.Fprintf(opts.Output, "Warning: failed to remove %s credentials: %v\n", name, err)
			} else {
				removed = true
			}
		}
	}

	if removed {
		fmt.Fprintln(opts.Output, "✓ Logged out")
		fmt.Fprintln(opts.Output, "Credentials removed")
	} else {
		fmt.Fprintln(opts.Output, "No credentials found")
	}

	return nil
}

// runAuthStatus displays authentication status.
func runAuthStatus(opts *AuthOptions) error {
	ctx := context.Background()

	fmt.Fprintln(opts.Output, "Authentication Status:")
	fmt.Fprintln(opts.Output)

	// Check all registered storage backends
	storages := opts.AuthManager.ListStorages()
	authenticated := false

	for _, name := range storages {
		storage := opts.AuthManager.GetStorage(name)
		if storage == nil {
			continue
		}

		// Check if credentials exist
		token, err := storage.Retrieve(ctx)
		if err == nil && token != nil {
			authenticated = true
			fmt.Fprintf(opts.Output, "  %s: ✓ Authenticated\n", name)

			// Show token details if available
			if !token.ExpiresAt.IsZero() {
				remaining := time.Until(token.ExpiresAt)
				if remaining > 0 {
					fmt.Fprintf(opts.Output, "    Expires: %s (in %s)\n",
						token.ExpiresAt.Format(time.RFC3339),
						formatDuration(remaining))
				} else {
					fmt.Fprintf(opts.Output, "    Status: ⚠️  Expired\n")
				}
			}
		} else {
			fmt.Fprintf(opts.Output, "  %s: ✗ Not authenticated\n", name)
		}
	}

	fmt.Fprintln(opts.Output)

	if !authenticated {
		fmt.Fprintln(opts.Output, "Run 'auth login' to authenticate")
	}

	return nil
}

// runAuthRefresh refreshes authentication tokens.
func runAuthRefresh(opts *AuthOptions) error {
	ctx := context.Background()

	fmt.Fprintln(opts.Output, "Refreshing authentication tokens...")

	// Get all registered authenticators
	authenticators := opts.AuthManager.ListAuthenticators()

	refreshed := false
	for _, name := range authenticators {
		authenticator := opts.AuthManager.GetAuthenticator(name)
		if authenticator == nil {
			continue
		}

		// Check if authenticator supports refresh
		storage := opts.AuthManager.GetStorage(name)
		if storage == nil {
			continue
		}

		// Get current token
		token, err := storage.Retrieve(ctx)
		if err != nil || token == nil {
			continue
		}

		// Try to refresh
		newToken, err := authenticator.Refresh(ctx, token)
		if err != nil {
			fmt.Fprintf(opts.Output, "Warning: failed to refresh %s: %v\n", name, err)
			continue
		}

		// Store new token
		if err := storage.Store(ctx, newToken); err != nil {
			fmt.Fprintf(opts.Output, "Warning: failed to store refreshed %s token: %v\n", name, err)
			continue
		}

		fmt.Fprintf(opts.Output, "✓ Refreshed %s token\n", name)
		refreshed = true
	}

	if !refreshed {
		return fmt.Errorf("no tokens were refreshed")
	}

	return nil
}
