package secrets

import "github.com/CliForge/cliforge/pkg/cli"

// DefaultFieldPatterns returns the default field name patterns for secret detection.
// These patterns match common field names that typically contain sensitive data.
func DefaultFieldPatterns() []string {
	return []string{
		"*password*",
		"*passwd*",
		"*secret*",
		"*token*",
		"*key", // api_key, access_key, etc.
		"*apikey*",
		"*api_key*",
		"*access_key*",
		"*secret_key*",
		"*private_key*",
		"*credential*",
		"*credentials*",
		"auth",
		"authorization",
		"*bearer*",
		"session",
		"*session_id*",
		"*sessionid*",
		"*ssn*",       // Social security number
		"*cc_number*", // Credit card
		"*card_number*",
		"*cvv*",
		"*cvc*",
		"*pin*",
		"*pincode*",
		"*oauth*",
		"*jwt*",
		"*refresh_token*",
		"*access_token*",
		"*id_token*",
	}
}

// DefaultValuePatterns returns the default value patterns for secret detection.
// These are regex patterns that match common secret formats.
func DefaultValuePatterns() []cli.ValuePattern {
	return []cli.ValuePattern{
		{
			Name:    "AWS Access Key ID",
			Pattern: `AKIA[0-9A-Z]{16}`,
			Enabled: true,
		},
		{
			Name:    "AWS Secret Access Key",
			Pattern: `(?i)aws(.{0,20})?['\"][0-9a-zA-Z/+]{40}['\"]`,
			Enabled: true,
		},
		{
			Name:    "AWS Session Token",
			Pattern: `(?i)aws.{0,20}session.{0,20}token['\"]?\s*[:=]\s*['\"]?[0-9a-zA-Z/+=]{100,}`,
			Enabled: true,
		},
		{
			Name:    "Generic API Key (sk_ prefix)",
			Pattern: `[sS][kK]_[a-zA-Z0-9_]{10,}`,
			Enabled: true,
		},
		{
			Name:    "Generic API Key (pk_ prefix)",
			Pattern: `[pP][kK]_[a-zA-Z0-9_]{10,}`,
			Enabled: true,
		},
		{
			Name:    "Bearer Token",
			Pattern: `Bearer\s+[A-Za-z0-9\-._~+/]+=*`,
			Enabled: true,
		},
		{
			Name:    "Basic Auth",
			Pattern: `Basic\s+[A-Za-z0-9+/]+=*`,
			Enabled: true,
		},
		{
			Name:    "JWT Token",
			Pattern: `eyJ[A-Za-z0-9-_=]+\.eyJ[A-Za-z0-9-_=]+\.[A-Za-z0-9-_.+/=]*`,
			Enabled: true,
		},
		{
			Name:    "GitHub Personal Access Token",
			Pattern: `ghp_[0-9a-zA-Z]{36}`,
			Enabled: true,
		},
		{
			Name:    "GitHub OAuth Token",
			Pattern: `gho_[0-9a-zA-Z]{36}`,
			Enabled: true,
		},
		{
			Name:    "GitHub App Token",
			Pattern: `(ghu|ghs)_[0-9a-zA-Z]{36}`,
			Enabled: true,
		},
		{
			Name:    "GitLab Personal Access Token",
			Pattern: `glpat-[0-9a-zA-Z\-_]{20}`,
			Enabled: true,
		},
		{
			Name:    "Slack Token",
			Pattern: `xox[baprs]-[0-9a-zA-Z]{10,48}`,
			Enabled: true,
		},
		{
			Name:    "Slack Webhook",
			Pattern: `https://hooks\.slack\.com/services/T[a-zA-Z0-9_]{8}/B[a-zA-Z0-9_]{8}/[a-zA-Z0-9_]{24}`,
			Enabled: true,
		},
		{
			Name:    "Stripe API Key",
			Pattern: `(sk|pk)_(test|live)_[0-9a-zA-Z]{24,}`,
			Enabled: true,
		},
		{
			Name:    "Google API Key",
			Pattern: `AIza[0-9A-Za-z\-_]{35}`,
			Enabled: true,
		},
		{
			Name:    "Google OAuth Token",
			Pattern: `ya29\.[0-9A-Za-z\-_]+`,
			Enabled: true,
		},
		{
			Name:    "Azure Client Secret",
			Pattern: `[0-9a-zA-Z\-_~\.]{34,}`,
			Enabled: false, // Too many false positives, disabled by default
		},
		{
			Name:    "RSA Private Key",
			Pattern: `-----BEGIN RSA PRIVATE KEY-----`,
			Enabled: true,
		},
		{
			Name:    "SSH Private Key",
			Pattern: `-----BEGIN OPENSSH PRIVATE KEY-----`,
			Enabled: true,
		},
		{
			Name:    "PGP Private Key",
			Pattern: `-----BEGIN PGP PRIVATE KEY BLOCK-----`,
			Enabled: true,
		},
		{
			Name:    "Generic Private Key",
			Pattern: `-----BEGIN PRIVATE KEY-----`,
			Enabled: true,
		},
		{
			Name:    "Heroku API Key",
			Pattern: `\b[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}\b`,
			Enabled: false, // UUID pattern, too many false positives
		},
		{
			Name:    "MailChimp API Key",
			Pattern: `[0-9a-f]{32}-us[0-9]{1,2}`,
			Enabled: true,
		},
		{
			Name:    "Mailgun API Key",
			Pattern: `key-[0-9a-zA-Z]{32}`,
			Enabled: true,
		},
		{
			Name:    "NPM Token",
			Pattern: `\bnpm_[a-zA-Z0-9]{36}\b`,
			Enabled: true,
		},
		{
			Name:    "PyPI Token",
			Pattern: `\bpypi-[A-Za-z0-9\-_]{32,}\b`,
			Enabled: true,
		},
		{
			Name:    "Twilio Account SID",
			Pattern: `\bAC[0-9a-fA-F]{32}\b`,
			Enabled: true,
		},
		{
			Name:    "Twilio Auth Token",
			Pattern: `\b[0-9a-fA-F]{32}\b`,
			Enabled: false, // Too generic, disabled by default
		},
		{
			Name:    "SendGrid API Key",
			Pattern: `\bSG\.[a-zA-Z0-9\-_]{22}\.[a-zA-Z0-9\-_]{43}\b`,
			Enabled: true,
		},
		{
			Name:    "DigitalOcean Token",
			Pattern: `\bdop_v1_[0-9a-fA-F]{64}\b`,
			Enabled: true,
		},
		{
			Name:    "Cloudflare API Token",
			Pattern: `\b[A-Za-z0-9_-]{40}\b`,
			Enabled: false, // Too generic, disabled by default
		},
		{
			Name:    "Password in URL",
			Pattern: `[a-zA-Z]{3,10}://[^/\\s:@]{3,20}:[^/\\s:@]{3,20}@.{1,100}`,
			Enabled: true,
		},
	}
}

// DefaultHeaders returns the default HTTP headers that should be masked.
func DefaultHeaders() []string {
	return []string{
		"Authorization",
		"X-API-Key",
		"X-Api-Key",
		"X-Auth-Token",
		"X-Access-Token",
		"X-Session-Token",
		"Cookie",
		"Set-Cookie",
		"Proxy-Authorization",
		"X-CSRF-Token",
		"X-XSRF-Token",
		"Api-Key",
		"Apikey",
		"Auth-Token",
		"Access-Token",
		"Session-Token",
		"Private-Token",
		"X-GitHub-Token",
		"X-GitLab-Token",
		"X-Stripe-Client-Secret",
		"X-Twilio-Signature",
	}
}

// DefaultSecretsBehavior returns a default secrets configuration with sensible defaults.
func DefaultSecretsBehavior() *cli.SecretsBehavior {
	return &cli.SecretsBehavior{
		Enabled:       true,
		FieldPatterns: DefaultFieldPatterns(),
		ValuePatterns: DefaultValuePatterns(),
		Headers:       DefaultHeaders(),
		Masking: &cli.SecretsMasking{
			Style:            "partial",
			PartialShowChars: 6,
			Replacement:      "***",
		},
		MaskIn: &cli.SecretsMaskIn{
			Stdout:      true,
			Stderr:      true,
			Logs:        true,
			Audit:       false, // Don't mask in audit (will use hash instead)
			DebugOutput: true,
		},
		ExplicitFields: []string{},
	}
}

// MergeSecretsBehavior merges user-provided secrets config with defaults.
// User settings take precedence over defaults.
func MergeSecretsBehavior(userConfig, defaultConfig *cli.SecretsBehavior) *cli.SecretsBehavior {
	if userConfig == nil {
		return defaultConfig
	}
	if defaultConfig == nil {
		return userConfig
	}

	result := &cli.SecretsBehavior{
		Enabled: userConfig.Enabled,
	}

	// Merge masking config
	if userConfig.Masking != nil {
		result.Masking = userConfig.Masking
	} else {
		result.Masking = defaultConfig.Masking
	}

	// Merge field patterns (user patterns extend defaults)
	result.FieldPatterns = append([]string{}, defaultConfig.FieldPatterns...)
	result.FieldPatterns = append(result.FieldPatterns, userConfig.FieldPatterns...)

	// Merge value patterns (user patterns extend defaults)
	// Build a map to handle overrides by name
	patternMap := make(map[string]cli.ValuePattern)
	for _, p := range defaultConfig.ValuePatterns {
		patternMap[p.Name] = p
	}
	for _, p := range userConfig.ValuePatterns {
		patternMap[p.Name] = p // Override if same name
	}

	result.ValuePatterns = make([]cli.ValuePattern, 0, len(patternMap))
	for _, p := range patternMap {
		result.ValuePatterns = append(result.ValuePatterns, p)
	}

	// Merge headers (user headers extend defaults)
	headerMap := make(map[string]bool)
	for _, h := range defaultConfig.Headers {
		headerMap[h] = true
	}
	for _, h := range userConfig.Headers {
		headerMap[h] = true
	}

	result.Headers = make([]string, 0, len(headerMap))
	for h := range headerMap {
		result.Headers = append(result.Headers, h)
	}

	// Merge explicit fields (user only)
	result.ExplicitFields = append([]string{}, userConfig.ExplicitFields...)

	// Merge mask_in config
	if userConfig.MaskIn != nil {
		result.MaskIn = userConfig.MaskIn
	} else {
		result.MaskIn = defaultConfig.MaskIn
	}

	return result
}

// ValidateSecretsBehavior validates a secrets configuration.
func ValidateSecretsBehavior(config *cli.SecretsBehavior) error {
	if config == nil {
		return nil
	}

	// Validate masking style
	if config.Masking != nil {
		style := config.Masking.Style
		if style != "" && style != "partial" && style != "full" && style != "hash" {
			return &ValidationError{
				Field:   "masking.style",
				Message: "must be one of: partial, full, hash",
			}
		}

		// Validate partial_show_chars
		if style == "partial" && config.Masking.PartialShowChars < 0 {
			return &ValidationError{
				Field:   "masking.partial_show_chars",
				Message: "must be non-negative",
			}
		}
	}

	// Validate value patterns
	for i, vp := range config.ValuePatterns {
		if vp.Name == "" {
			return &ValidationError{
				Field:   "value_patterns[" + string(rune(i)) + "].name",
				Message: "pattern name is required",
			}
		}
		if vp.Pattern == "" {
			return &ValidationError{
				Field:   "value_patterns[" + string(rune(i)) + "].pattern",
				Message: "pattern regex is required",
			}
		}
	}

	return nil
}

// ValidationError represents a configuration validation error.
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Field + ": " + e.Message
}
