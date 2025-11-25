package config

import (
	"fmt"

	"github.com/CliForge/cliforge/pkg/cli"
)

// mergeConfigs merges embedded and user configurations according to override rules.
// Returns the merged config and a map of debug overrides (if any).
func (l *Loader) mergeConfigs(embedded *cli.CLIConfig, user *cli.UserConfig) (*cli.CLIConfig, map[string]any, error) {
	// Start with a copy of embedded config
	merged := copyConfig(embedded)
	debugOverrides := make(map[string]any)

	// Apply debug overrides if in debug mode
	if embedded.Metadata.Debug && user.DebugOverride != nil {
		merged, debugOverrides = l.applyDebugOverrides(merged, user.DebugOverride)
	}

	// Apply user preferences (always applied)
	if user.Preferences != nil {
		merged = l.applyUserPreferences(merged, user.Preferences)
	}

	return merged, debugOverrides, nil
}

// applyDebugOverrides applies debug overrides to the configuration.
// This allows overriding ANY embedded config when metadata.debug is true.
func (l *Loader) applyDebugOverrides(config *cli.CLIConfig, override *cli.CLIConfig) (*cli.CLIConfig, map[string]any) {
	overrides := make(map[string]any)

	// Override API settings
	if override.API.BaseURL != "" && override.API.BaseURL != config.API.BaseURL {
		config.API.BaseURL = override.API.BaseURL
		overrides["api.base_url"] = override.API.BaseURL
	}
	if override.API.OpenAPIURL != "" && override.API.OpenAPIURL != config.API.OpenAPIURL {
		config.API.OpenAPIURL = override.API.OpenAPIURL
		overrides["api.openapi_url"] = override.API.OpenAPIURL
	}

	// Override metadata
	if override.Metadata.Name != "" && override.Metadata.Name != config.Metadata.Name {
		config.Metadata.Name = override.Metadata.Name
		overrides["metadata.name"] = override.Metadata.Name
	}

	// Override branding
	if override.Branding != nil {
		if override.Branding.Colors != nil && override.Branding.Colors.Primary != "" {
			if config.Branding == nil {
				config.Branding = &cli.Branding{}
			}
			if config.Branding.Colors == nil {
				config.Branding.Colors = &cli.Colors{}
			}
			config.Branding.Colors.Primary = override.Branding.Colors.Primary
			overrides["branding.colors.primary"] = override.Branding.Colors.Primary
		}
	}

	// Override behaviors (auth)
	if override.Behaviors != nil && override.Behaviors.Auth != nil {
		if override.Behaviors.Auth.Type != "" && override.Behaviors.Auth.Type != config.Behaviors.Auth.Type {
			if config.Behaviors == nil {
				config.Behaviors = &cli.Behaviors{}
			}
			if config.Behaviors.Auth == nil {
				config.Behaviors.Auth = &cli.AuthBehavior{}
			}
			config.Behaviors.Auth.Type = override.Behaviors.Auth.Type
			overrides["behaviors.auth.type"] = override.Behaviors.Auth.Type
		}
	}

	return config, overrides
}

// applyUserPreferences applies user preferences to override defaults.
func (l *Loader) applyUserPreferences(config *cli.CLIConfig, prefs *cli.UserPreferences) *cli.CLIConfig {
	// Initialize defaults if not present
	if config.Defaults == nil {
		config.Defaults = &cli.Defaults{}
	}

	// Apply HTTP preferences
	if prefs.HTTP != nil {
		if config.Defaults.HTTP == nil {
			config.Defaults.HTTP = &cli.DefaultsHTTP{}
		}
		if prefs.HTTP.Timeout != "" {
			config.Defaults.HTTP.Timeout = prefs.HTTP.Timeout
		}
	}

	// Apply caching preferences
	if prefs.Caching != nil {
		if config.Defaults.Caching == nil {
			config.Defaults.Caching = &cli.DefaultsCaching{}
		}
		config.Defaults.Caching.Enabled = prefs.Caching.Enabled
	}

	// Apply pagination preferences
	if prefs.Pagination != nil {
		if config.Defaults.Pagination == nil {
			config.Defaults.Pagination = &cli.DefaultsPagination{}
		}
		if prefs.Pagination.Limit > 0 {
			// Respect max_limit from behaviors
			maxLimit := 100 // default
			if config.Behaviors != nil && config.Behaviors.Pagination != nil && config.Behaviors.Pagination.MaxLimit > 0 {
				maxLimit = config.Behaviors.Pagination.MaxLimit
			}
			if prefs.Pagination.Limit <= maxLimit {
				config.Defaults.Pagination.Limit = prefs.Pagination.Limit
			}
		}
	}

	// Apply output preferences
	if prefs.Output != nil {
		if config.Defaults.Output == nil {
			config.Defaults.Output = &cli.DefaultsOutput{}
		}
		if prefs.Output.Format != "" {
			config.Defaults.Output.Format = prefs.Output.Format
		}
		if prefs.Output.Color != "" {
			config.Defaults.Output.Color = prefs.Output.Color
		}
		config.Defaults.Output.PrettyPrint = prefs.Output.PrettyPrint
		config.Defaults.Output.Paging = prefs.Output.Paging
	}

	// Apply deprecation preferences
	if prefs.Deprecations != nil {
		if config.Defaults.Deprecations == nil {
			config.Defaults.Deprecations = &cli.DefaultsDeprecations{}
		}
		config.Defaults.Deprecations.AlwaysShow = prefs.Deprecations.AlwaysShow
		if prefs.Deprecations.MinSeverity != "" {
			config.Defaults.Deprecations.MinSeverity = prefs.Deprecations.MinSeverity
		}
	}

	// Apply retry preferences
	if prefs.Retry != nil {
		if config.Defaults.Retry == nil {
			config.Defaults.Retry = &cli.DefaultsRetry{}
		}
		if prefs.Retry.MaxAttempts > 0 {
			config.Defaults.Retry.MaxAttempts = prefs.Retry.MaxAttempts
		}
	}

	return config
}

// copyConfig creates a deep copy of a CLIConfig.
func copyConfig(src *cli.CLIConfig) *cli.CLIConfig {
	if src == nil {
		return nil
	}

	dst := &cli.CLIConfig{
		Metadata: src.Metadata,
		API:      src.API,
	}

	// Copy branding
	if src.Branding != nil {
		dst.Branding = &cli.Branding{}
		if src.Branding.Colors != nil {
			colors := *src.Branding.Colors
			dst.Branding.Colors = &colors
		}
		if src.Branding.Prompts != nil {
			prompts := *src.Branding.Prompts
			dst.Branding.Prompts = &prompts
		}
		if src.Branding.Theme != nil {
			theme := *src.Branding.Theme
			dst.Branding.Theme = &theme
		}
		dst.Branding.ASCIIArt = src.Branding.ASCIIArt
	}

	// Copy environments
	if src.API.Environments != nil {
		dst.API.Environments = make([]cli.Environment, len(src.API.Environments))
		copy(dst.API.Environments, src.API.Environments)
	}

	// Copy default headers
	if src.API.DefaultHeaders != nil {
		dst.API.DefaultHeaders = make(map[string]string)
		for k, v := range src.API.DefaultHeaders {
			dst.API.DefaultHeaders[k] = v
		}
	}

	// Copy defaults
	if src.Defaults != nil {
		dst.Defaults = &cli.Defaults{}
		if src.Defaults.HTTP != nil {
			http := *src.Defaults.HTTP
			dst.Defaults.HTTP = &http
		}
		if src.Defaults.Caching != nil {
			caching := *src.Defaults.Caching
			dst.Defaults.Caching = &caching
		}
		if src.Defaults.Pagination != nil {
			pagination := *src.Defaults.Pagination
			dst.Defaults.Pagination = &pagination
		}
		if src.Defaults.Output != nil {
			output := *src.Defaults.Output
			dst.Defaults.Output = &output
		}
		if src.Defaults.Deprecations != nil {
			deprecations := *src.Defaults.Deprecations
			dst.Defaults.Deprecations = &deprecations
		}
		if src.Defaults.Retry != nil {
			retry := *src.Defaults.Retry
			dst.Defaults.Retry = &retry
		}
	}

	// Copy updates
	if src.Updates != nil {
		updates := *src.Updates
		dst.Updates = &updates
	}

	// Copy behaviors (deep copy needed)
	if src.Behaviors != nil {
		dst.Behaviors = &cli.Behaviors{}
		if src.Behaviors.Auth != nil {
			auth := *src.Behaviors.Auth
			dst.Behaviors.Auth = &auth
			if src.Behaviors.Auth.APIKey != nil {
				apiKey := *src.Behaviors.Auth.APIKey
				dst.Behaviors.Auth.APIKey = &apiKey
			}
			if src.Behaviors.Auth.OAuth2 != nil {
				oauth2 := *src.Behaviors.Auth.OAuth2
				if src.Behaviors.Auth.OAuth2.Scopes != nil {
					oauth2.Scopes = make([]string, len(src.Behaviors.Auth.OAuth2.Scopes))
					copy(oauth2.Scopes, src.Behaviors.Auth.OAuth2.Scopes)
				}
				dst.Behaviors.Auth.OAuth2 = &oauth2
			}
			if src.Behaviors.Auth.Basic != nil {
				basic := *src.Behaviors.Auth.Basic
				dst.Behaviors.Auth.Basic = &basic
			}
		}
		if src.Behaviors.Caching != nil {
			caching := *src.Behaviors.Caching
			dst.Behaviors.Caching = &caching
		}
		if src.Behaviors.Retry != nil {
			retry := *src.Behaviors.Retry
			if src.Behaviors.Retry.RetryOnStatus != nil {
				retry.RetryOnStatus = make([]int, len(src.Behaviors.Retry.RetryOnStatus))
				copy(retry.RetryOnStatus, src.Behaviors.Retry.RetryOnStatus)
			}
			dst.Behaviors.Retry = &retry
		}
		if src.Behaviors.Pagination != nil {
			pagination := *src.Behaviors.Pagination
			dst.Behaviors.Pagination = &pagination
		}
		if src.Behaviors.Secrets != nil {
			secrets := *src.Behaviors.Secrets
			if src.Behaviors.Secrets.Masking != nil {
				masking := *src.Behaviors.Secrets.Masking
				secrets.Masking = &masking
			}
			if src.Behaviors.Secrets.FieldPatterns != nil {
				secrets.FieldPatterns = make([]string, len(src.Behaviors.Secrets.FieldPatterns))
				copy(secrets.FieldPatterns, src.Behaviors.Secrets.FieldPatterns)
			}
			if src.Behaviors.Secrets.Headers != nil {
				secrets.Headers = make([]string, len(src.Behaviors.Secrets.Headers))
				copy(secrets.Headers, src.Behaviors.Secrets.Headers)
			}
			if src.Behaviors.Secrets.MaskIn != nil {
				maskIn := *src.Behaviors.Secrets.MaskIn
				secrets.MaskIn = &maskIn
			}
			dst.Behaviors.Secrets = &secrets
		}
	}

	// Copy features
	if src.Features != nil {
		features := *src.Features
		dst.Features = &features
	}

	return dst
}

// MergeDefaults applies built-in defaults to a config for any missing values.
func MergeDefaults(config *cli.CLIConfig) error {
	if config == nil {
		return fmt.Errorf("config is nil")
	}

	// Apply default values for defaults section
	if config.Defaults == nil {
		config.Defaults = &cli.Defaults{}
	}

	if config.Defaults.HTTP == nil {
		config.Defaults.HTTP = &cli.DefaultsHTTP{
			Timeout: "30s",
		}
	} else if config.Defaults.HTTP.Timeout == "" {
		config.Defaults.HTTP.Timeout = "30s"
	}

	if config.Defaults.Caching == nil {
		config.Defaults.Caching = &cli.DefaultsCaching{
			Enabled: true,
		}
	}

	if config.Defaults.Pagination == nil {
		config.Defaults.Pagination = &cli.DefaultsPagination{
			Limit: 20,
		}
	} else if config.Defaults.Pagination.Limit == 0 {
		config.Defaults.Pagination.Limit = 20
	}

	if config.Defaults.Output == nil {
		config.Defaults.Output = &cli.DefaultsOutput{
			Format:      "json",
			PrettyPrint: true,
			Color:       "auto",
			Paging:      true,
		}
	} else {
		if config.Defaults.Output.Format == "" {
			config.Defaults.Output.Format = "json"
		}
		if config.Defaults.Output.Color == "" {
			config.Defaults.Output.Color = "auto"
		}
	}

	if config.Defaults.Deprecations == nil {
		config.Defaults.Deprecations = &cli.DefaultsDeprecations{
			AlwaysShow:  false,
			MinSeverity: "info",
		}
	} else if config.Defaults.Deprecations.MinSeverity == "" {
		config.Defaults.Deprecations.MinSeverity = "info"
	}

	if config.Defaults.Retry == nil {
		config.Defaults.Retry = &cli.DefaultsRetry{
			MaxAttempts: 3,
		}
	} else if config.Defaults.Retry.MaxAttempts == 0 {
		config.Defaults.Retry.MaxAttempts = 3
	}

	// Apply default behaviors if not set
	if config.Behaviors != nil {
		if config.Behaviors.Caching != nil {
			if config.Behaviors.Caching.SpecTTL == "" {
				config.Behaviors.Caching.SpecTTL = "5m"
			}
			if config.Behaviors.Caching.ResponseTTL == "" {
				config.Behaviors.Caching.ResponseTTL = "1m"
			}
			if config.Behaviors.Caching.MaxSize == "" {
				config.Behaviors.Caching.MaxSize = "100MB"
			}
		}

		if config.Behaviors.Retry != nil {
			if config.Behaviors.Retry.InitialDelay == "" {
				config.Behaviors.Retry.InitialDelay = "1s"
			}
			if config.Behaviors.Retry.MaxDelay == "" {
				config.Behaviors.Retry.MaxDelay = "30s"
			}
			if config.Behaviors.Retry.BackoffMultiplier == 0 {
				config.Behaviors.Retry.BackoffMultiplier = 2.0
			}
			if config.Behaviors.Retry.RetryOnStatus == nil {
				config.Behaviors.Retry.RetryOnStatus = []int{429, 500, 502, 503, 504}
			}
		}

		if config.Behaviors.Pagination != nil {
			if config.Behaviors.Pagination.Delay == "" {
				config.Behaviors.Pagination.Delay = "100ms"
			}
			if config.Behaviors.Pagination.MaxLimit == 0 {
				config.Behaviors.Pagination.MaxLimit = 100
			}
		}
	}

	return nil
}
