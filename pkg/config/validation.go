package config

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/CliForge/cliforge/pkg/cli"
)

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidationErrors is a collection of validation errors.
type ValidationErrors []ValidationError

func (e ValidationErrors) Error() string {
	if len(e) == 0 {
		return ""
	}
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, err.Error())
	}
	return fmt.Sprintf("validation failed:\n  - %s", strings.Join(msgs, "\n  - "))
}

// Validator handles configuration validation.
type Validator struct {
	errors ValidationErrors
}

// NewValidator creates a new validator.
func NewValidator() *Validator {
	return &Validator{
		errors: make(ValidationErrors, 0),
	}
}

// Validate validates a complete CLI configuration.
func (v *Validator) Validate(config *cli.CLIConfig) error {
	v.errors = make(ValidationErrors, 0)

	// Validate metadata
	v.validateMetadata(&config.Metadata)

	// Validate API
	v.validateAPI(&config.API)

	// Validate defaults
	if config.Defaults != nil {
		v.validateDefaults(config.Defaults)
	}

	// Validate behaviors
	if config.Behaviors != nil {
		v.validateBehaviors(config.Behaviors)
	}

	// Validate updates
	if config.Updates != nil {
		v.validateUpdates(config.Updates)
	}

	if len(v.errors) > 0 {
		return v.errors
	}

	return nil
}

// validateMetadata validates the metadata section.
func (v *Validator) validateMetadata(m *cli.Metadata) {
	// Name is required and must be valid
	if m.Name == "" {
		v.addError("metadata.name", "name is required")
	} else {
		matched, _ := regexp.MatchString(`^[a-z0-9-]+$`, m.Name)
		if !matched {
			v.addError("metadata.name", "name must contain only lowercase letters, numbers, and hyphens")
		}
		if len(m.Name) > 50 {
			v.addError("metadata.name", "name must be 50 characters or less")
		}
	}

	// Version is required and must be semver
	if m.Version == "" {
		v.addError("metadata.version", "version is required")
	} else {
		matched, _ := regexp.MatchString(`^\d+\.\d+\.\d+`, m.Version)
		if !matched {
			v.addError("metadata.version", "version must follow semantic versioning (e.g., 1.0.0)")
		}
	}

	// Description is required
	if m.Description == "" {
		v.addError("metadata.description", "description is required")
	} else {
		if len(m.Description) < 10 {
			v.addError("metadata.description", "description must be at least 10 characters")
		}
		if len(m.Description) > 200 {
			v.addError("metadata.description", "description must be 200 characters or less")
		}
	}

	// Validate URLs if present
	if m.Homepage != "" && !v.isValidURL(m.Homepage) {
		v.addError("metadata.homepage", "homepage must be a valid URL")
	}
	if m.BugsURL != "" && !v.isValidURL(m.BugsURL) {
		v.addError("metadata.bugs_url", "bugs_url must be a valid URL")
	}
	if m.DocsURL != "" && !v.isValidURL(m.DocsURL) {
		v.addError("metadata.docs_url", "docs_url must be a valid URL")
	}
}

// validateAPI validates the API section.
func (v *Validator) validateAPI(a *cli.API) {
	// OpenAPI URL is required
	if a.OpenAPIURL == "" {
		v.addError("api.openapi_url", "openapi_url is required")
	} else if !v.isValidURLOrFilePath(a.OpenAPIURL) {
		v.addError("api.openapi_url", "openapi_url must be a valid URL or file path")
	}

	// Base URL is required
	if a.BaseURL == "" {
		v.addError("api.base_url", "base_url is required")
	} else if !v.isValidURL(a.BaseURL) {
		v.addError("api.base_url", "base_url must be a valid URL")
	}

	// Validate telemetry URL if present
	if a.TelemetryURL != "" && !v.isValidURL(a.TelemetryURL) {
		v.addError("api.telemetry_url", "telemetry_url must be a valid URL")
	}

	// Validate environments
	defaultCount := 0
	for i, env := range a.Environments {
		if env.Name == "" {
			v.addError(fmt.Sprintf("api.environments[%d].name", i), "environment name is required")
		}
		if env.OpenAPIURL == "" {
			v.addError(fmt.Sprintf("api.environments[%d].openapi_url", i), "openapi_url is required")
		} else if !v.isValidURLOrFilePath(env.OpenAPIURL) {
			v.addError(fmt.Sprintf("api.environments[%d].openapi_url", i), "openapi_url must be a valid URL or file path")
		}
		if env.BaseURL == "" {
			v.addError(fmt.Sprintf("api.environments[%d].base_url", i), "base_url is required")
		} else if !v.isValidURL(env.BaseURL) {
			v.addError(fmt.Sprintf("api.environments[%d].base_url", i), "base_url must be a valid URL")
		}
		if env.Default {
			defaultCount++
		}
	}

	if len(a.Environments) > 0 && defaultCount == 0 {
		v.addError("api.environments", "at least one environment must be marked as default")
	}
	if defaultCount > 1 {
		v.addError("api.environments", "only one environment can be marked as default")
	}
}

// validateDefaults validates the defaults section.
func (v *Validator) validateDefaults(d *cli.Defaults) {
	// Validate HTTP defaults
	if d.HTTP != nil {
		if d.HTTP.Timeout != "" {
			if !v.isValidDuration(d.HTTP.Timeout) {
				v.addError("defaults.http.timeout", "timeout must be a valid duration (e.g., 30s, 1m)")
			}
		}
	}

	// Validate pagination defaults
	if d.Pagination != nil {
		if d.Pagination.Limit < 0 {
			v.addError("defaults.pagination.limit", "limit must be non-negative")
		}
	}

	// Validate output defaults
	if d.Output != nil {
		validFormats := []string{"json", "yaml", "table", "csv"}
		if d.Output.Format != "" && !contains(validFormats, d.Output.Format) {
			v.addError("defaults.output.format", "format must be one of: json, yaml, table, csv")
		}

		validColors := []string{"auto", "always", "never"}
		if d.Output.Color != "" && !contains(validColors, d.Output.Color) {
			v.addError("defaults.output.color", "color must be one of: auto, always, never")
		}
	}

	// Validate deprecation defaults
	if d.Deprecations != nil {
		validSeverities := []string{"info", "warning", "urgent", "critical", "removed"}
		if d.Deprecations.MinSeverity != "" && !contains(validSeverities, d.Deprecations.MinSeverity) {
			v.addError("defaults.deprecations.min_severity", "min_severity must be one of: info, warning, urgent, critical, removed")
		}
	}

	// Validate retry defaults
	if d.Retry != nil {
		if d.Retry.MaxAttempts < 0 {
			v.addError("defaults.retry.max_attempts", "max_attempts must be non-negative")
		}
		if d.Retry.MaxAttempts > 10 {
			v.addError("defaults.retry.max_attempts", "max_attempts should not exceed 10")
		}
	}
}

// validateBehaviors validates the behaviors section.
func (v *Validator) validateBehaviors(b *cli.Behaviors) {
	// Validate auth
	if b.Auth != nil {
		validAuthTypes := []string{"none", "api_key", "oauth2", "basic"}
		if b.Auth.Type != "" && !contains(validAuthTypes, b.Auth.Type) {
			v.addError("behaviors.auth.type", "type must be one of: none, api_key, oauth2, basic")
		}

		// Validate auth type-specific fields
		switch b.Auth.Type {
		case "api_key":
			if b.Auth.APIKey == nil {
				v.addError("behaviors.auth.api_key", "api_key configuration is required when type is api_key")
			} else {
				if b.Auth.APIKey.Header == "" {
					v.addError("behaviors.auth.api_key.header", "header is required")
				}
				if b.Auth.APIKey.EnvVar == "" {
					v.addError("behaviors.auth.api_key.env_var", "env_var is required")
				}
			}
		case "oauth2":
			if b.Auth.OAuth2 == nil {
				v.addError("behaviors.auth.oauth2", "oauth2 configuration is required when type is oauth2")
			} else {
				if b.Auth.OAuth2.ClientID == "" {
					v.addError("behaviors.auth.oauth2.client_id", "client_id is required")
				}
				if b.Auth.OAuth2.AuthURL == "" {
					v.addError("behaviors.auth.oauth2.auth_url", "auth_url is required")
				} else if !v.isValidURL(b.Auth.OAuth2.AuthURL) {
					v.addError("behaviors.auth.oauth2.auth_url", "auth_url must be a valid URL")
				}
				if b.Auth.OAuth2.TokenURL == "" {
					v.addError("behaviors.auth.oauth2.token_url", "token_url is required")
				} else if !v.isValidURL(b.Auth.OAuth2.TokenURL) {
					v.addError("behaviors.auth.oauth2.token_url", "token_url must be a valid URL")
				}
			}
		case "basic":
			if b.Auth.Basic == nil {
				v.addError("behaviors.auth.basic", "basic configuration is required when type is basic")
			} else {
				if b.Auth.Basic.UsernameEnv == "" {
					v.addError("behaviors.auth.basic.username_env", "username_env is required")
				}
				if b.Auth.Basic.PasswordEnv == "" {
					v.addError("behaviors.auth.basic.password_env", "password_env is required")
				}
			}
		}
	}

	// Validate caching
	if b.Caching != nil {
		if b.Caching.SpecTTL != "" && !v.isValidDuration(b.Caching.SpecTTL) {
			v.addError("behaviors.caching.spec_ttl", "spec_ttl must be a valid duration")
		}
		if b.Caching.ResponseTTL != "" && !v.isValidDuration(b.Caching.ResponseTTL) {
			v.addError("behaviors.caching.response_ttl", "response_ttl must be a valid duration")
		}
	}

	// Validate retry
	if b.Retry != nil {
		if b.Retry.InitialDelay != "" && !v.isValidDuration(b.Retry.InitialDelay) {
			v.addError("behaviors.retry.initial_delay", "initial_delay must be a valid duration")
		}
		if b.Retry.MaxDelay != "" && !v.isValidDuration(b.Retry.MaxDelay) {
			v.addError("behaviors.retry.max_delay", "max_delay must be a valid duration")
		}
		if b.Retry.BackoffMultiplier < 1.0 {
			v.addError("behaviors.retry.backoff_multiplier", "backoff_multiplier must be >= 1.0")
		}
	}

	// Validate pagination
	if b.Pagination != nil {
		if b.Pagination.MaxLimit < 1 {
			v.addError("behaviors.pagination.max_limit", "max_limit must be at least 1")
		}
		if b.Pagination.Delay != "" && !v.isValidDuration(b.Pagination.Delay) {
			v.addError("behaviors.pagination.delay", "delay must be a valid duration")
		}
	}

	// Validate secrets
	if b.Secrets != nil && b.Secrets.Masking != nil {
		validStyles := []string{"partial", "full", "hash"}
		if b.Secrets.Masking.Style != "" && !contains(validStyles, b.Secrets.Masking.Style) {
			v.addError("behaviors.secrets.masking.style", "style must be one of: partial, full, hash")
		}
		if b.Secrets.Masking.PartialShowChars < 0 {
			v.addError("behaviors.secrets.masking.partial_show_chars", "partial_show_chars must be non-negative")
		}
	}
}

// validateUpdates validates the updates section.
func (v *Validator) validateUpdates(u *cli.Updates) {
	if u.Enabled {
		if u.UpdateURL == "" {
			v.addError("updates.update_url", "update_url is required when updates are enabled")
		} else if !v.isValidURL(u.UpdateURL) {
			v.addError("updates.update_url", "update_url must be a valid URL")
		}
	}

	if u.CheckInterval != "" && !v.isValidDuration(u.CheckInterval) {
		v.addError("updates.check_interval", "check_interval must be a valid duration")
	}
}

// ValidateUserPreferences validates user preferences for overridable settings.
func (v *Validator) ValidateUserPreferences(prefs *cli.UserPreferences) error {
	v.errors = make(ValidationErrors, 0)

	if prefs.HTTP != nil {
		if prefs.HTTP.Timeout != "" && !v.isValidDuration(prefs.HTTP.Timeout) {
			v.addError("preferences.http.timeout", "timeout must be a valid duration")
		}
		if prefs.HTTP.Proxy != "" && !v.isValidURL(prefs.HTTP.Proxy) {
			v.addError("preferences.http.proxy", "proxy must be a valid URL")
		}
		if prefs.HTTP.HTTPSProxy != "" && !v.isValidURL(prefs.HTTP.HTTPSProxy) {
			v.addError("preferences.http.https_proxy", "https_proxy must be a valid URL")
		}
	}

	if prefs.Pagination != nil {
		if prefs.Pagination.Limit < 0 {
			v.addError("preferences.pagination.limit", "limit must be non-negative")
		}
	}

	if prefs.Output != nil {
		validFormats := []string{"json", "yaml", "table", "csv"}
		if prefs.Output.Format != "" && !contains(validFormats, prefs.Output.Format) {
			v.addError("preferences.output.format", "format must be one of: json, yaml, table, csv")
		}

		validColors := []string{"auto", "always", "never"}
		if prefs.Output.Color != "" && !contains(validColors, prefs.Output.Color) {
			v.addError("preferences.output.color", "color must be one of: auto, always, never")
		}
	}

	if prefs.Deprecations != nil {
		validSeverities := []string{"info", "warning", "urgent", "critical", "removed"}
		if prefs.Deprecations.MinSeverity != "" && !contains(validSeverities, prefs.Deprecations.MinSeverity) {
			v.addError("preferences.deprecations.min_severity", "min_severity must be one of: info, warning, urgent, critical, removed")
		}
	}

	if prefs.Retry != nil {
		if prefs.Retry.MaxAttempts < 0 {
			v.addError("preferences.retry.max_attempts", "max_attempts must be non-negative")
		}
		if prefs.Retry.MaxAttempts > 10 {
			v.addError("preferences.retry.max_attempts", "max_attempts should not exceed 10")
		}
	}

	if len(v.errors) > 0 {
		return v.errors
	}

	return nil
}

// Helper methods

func (v *Validator) addError(field, message string) {
	v.errors = append(v.errors, ValidationError{
		Field:   field,
		Message: message,
	})
}

func (v *Validator) isValidURL(urlStr string) bool {
	if urlStr == "" {
		return false
	}
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	return u.Scheme != "" && u.Host != ""
}

func (v *Validator) isValidURLOrFilePath(urlStr string) bool {
	if urlStr == "" {
		return false
	}
	// Check if it's a URL
	if v.isValidURL(urlStr) {
		return true
	}
	// Check if it's a file path (starts with ./ or /)
	return strings.HasPrefix(urlStr, "./") || strings.HasPrefix(urlStr, "/") || strings.HasPrefix(urlStr, "file://")
}

func (v *Validator) isValidDuration(d string) bool {
	_, err := time.ParseDuration(d)
	return err == nil
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
