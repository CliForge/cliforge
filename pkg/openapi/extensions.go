package openapi

import (
	"fmt"
	"time"

	"github.com/getkin/kin-openapi/openapi3"
)

// SpecExtensions contains all CLI extensions from the OpenAPI spec.
type SpecExtensions struct {
	// Global configuration
	Config *CLIConfig
	// Auth configuration
	AuthConfig *AuthConfig
	// Changelog entries
	Changelog []ChangelogEntry
	// Deprecations
	Deprecations []*DeprecationInfo
}

// CLIConfig represents the x-cli-config global extension.
type CLIConfig struct {
	Name        string                 `json:"name"`
	Version     string                 `json:"version"`
	Description string                 `json:"description"`
	Branding    *BrandingConfig        `json:"branding"`
	Auth        *AuthSettings          `json:"auth"`
	Output      *OutputSettings        `json:"output"`
	Features    *FeatureSettings       `json:"features"`
	Cache       *CacheSettings         `json:"cache"`
	Raw         map[string]interface{} `json:"-"` // Raw extension data
}

// BrandingConfig contains branding settings.
type BrandingConfig struct {
	ASCIIArt string                 `json:"ascii-art"`
	Colors   *ColorScheme           `json:"colors"`
	Raw      map[string]interface{} `json:"-"`
}

// ColorScheme defines CLI color scheme.
type ColorScheme struct {
	Primary   string `json:"primary"`
	Secondary string `json:"secondary"`
}

// AuthSettings contains authentication settings.
type AuthSettings struct {
	Type        string `json:"type"`    // oauth2, apikey, basic
	Storage     string `json:"storage"` // file, keyring, memory
	AutoRefresh bool   `json:"auto-refresh"`
}

// OutputSettings contains output format settings.
type OutputSettings struct {
	DefaultFormat    string   `json:"default-format"`
	SupportedFormats []string `json:"supported-formats"`
}

// FeatureSettings contains feature flags.
type FeatureSettings struct {
	InteractiveMode bool `json:"interactive-mode"`
	AutoComplete    bool `json:"auto-complete"`
	SelfUpdate      bool `json:"self-update"`
	Telemetry       bool `json:"telemetry"`
}

// CacheSettings contains cache configuration.
type CacheSettings struct {
	Enabled  bool   `json:"enabled"`
	TTL      int    `json:"ttl"` // seconds
	Location string `json:"location"`
}

// AuthConfig represents the x-auth-config security extension.
type AuthConfig struct {
	Flows        *OAuth2Flows   `json:"flows"`
	TokenStorage []TokenStorage `json:"token-storage"`
}

// OAuth2Flows contains OAuth2 flow configurations.
type OAuth2Flows struct {
	AuthorizationCode *OAuth2Flow `json:"authorizationCode"`
	ClientCredentials *OAuth2Flow `json:"clientCredentials"`
	Implicit          *OAuth2Flow `json:"implicit"`
	Password          *OAuth2Flow `json:"password"`
}

// OAuth2Flow represents a single OAuth2 flow.
type OAuth2Flow struct {
	AuthorizationURL string            `json:"authorizationUrl"`
	TokenURL         string            `json:"tokenUrl"`
	RefreshURL       string            `json:"refreshUrl"`
	Scopes           map[string]string `json:"scopes"`
}

// TokenStorage represents token storage configuration.
type TokenStorage struct {
	Type            string            `json:"type"` // file, keyring
	Path            string            `json:"path"`
	Service         string            `json:"service"`
	KeyringBackends map[string]string `json:"keyring-backends"`
}

// CLIFlag represents a single x-cli-flags entry.
type CLIFlag struct {
	Name        string      `json:"name"`
	Source      string      `json:"source"`
	Flag        string      `json:"flag"`
	Aliases     []string    `json:"aliases"`
	Required    bool        `json:"required"`
	Type        string      `json:"type"`
	Enum        []string    `json:"enum"`
	Default     interface{} `json:"default"`
	Description string      `json:"description"`
}

// CLIInteractive represents the x-cli-interactive extension.
type CLIInteractive struct {
	Enabled bool                 `json:"enabled"`
	Prompts []*InteractivePrompt `json:"prompts"`
}

// InteractivePrompt represents a single interactive prompt.
type InteractivePrompt struct {
	Parameter         string        `json:"parameter"`
	Type              string        `json:"type"` // text, select, confirm, number, password
	Message           string        `json:"message"`
	Default           interface{}   `json:"default"`
	Validation        string        `json:"validation"`
	ValidationMessage string        `json:"validation-message"`
	Source            *PromptSource `json:"source"`
}

// PromptSource defines where to fetch select options from.
type PromptSource struct {
	Endpoint     string `json:"endpoint"`
	ValueField   string `json:"value-field"`
	DisplayField string `json:"display-field"`
	Filter       string `json:"filter"`
}

// CLIPreflight represents x-cli-preflight checks.
type CLIPreflight struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Endpoint    string `json:"endpoint"`
	Method      string `json:"method"`
	Required    bool   `json:"required"`
	SkipFlag    string `json:"skip-flag"`
}

// CLIConfirmation represents the x-cli-confirmation extension.
type CLIConfirmation struct {
	Enabled bool   `json:"enabled"`
	Message string `json:"message"`
	Flag    string `json:"flag"`
}

// CLIAsync represents the x-cli-async extension.
type CLIAsync struct {
	Enabled        bool           `json:"enabled"`
	StatusField    string         `json:"status-field"`
	StatusEndpoint string         `json:"status-endpoint"`
	TerminalStates []string       `json:"terminal-states"`
	Polling        *PollingConfig `json:"polling"`
}

// PollingConfig defines polling behavior for async operations.
type PollingConfig struct {
	Interval int            `json:"interval"` // seconds
	Timeout  int            `json:"timeout"`  // seconds
	Backoff  *BackoffConfig `json:"backoff"`
}

// BackoffConfig defines exponential backoff settings.
type BackoffConfig struct {
	Enabled     bool    `json:"enabled"`
	Multiplier  float64 `json:"multiplier"`
	MaxInterval int     `json:"max-interval"` // seconds
}

// CLIOutput represents the x-cli-output extension.
type CLIOutput struct {
	Format         string       `json:"format"`
	SuccessMessage string       `json:"success-message"`
	ErrorMessage   string       `json:"error-message"`
	WatchStatus    bool         `json:"watch-status"`
	Table          *TableConfig `json:"table"`
}

// TableConfig defines table output configuration.
type TableConfig struct {
	Columns []*TableColumn `json:"columns"`
}

// TableColumn defines a single table column.
type TableColumn struct {
	Field     string `json:"field"`
	Header    string `json:"header"`
	Transform string `json:"transform"`
	Width     int    `json:"width"`
}

// CLIWorkflow represents the x-cli-workflow extension.
type CLIWorkflow struct {
	Steps  []*WorkflowStep `json:"steps"`
	Output *WorkflowOutput `json:"output"`
}

// WorkflowStep represents a single workflow step.
type WorkflowStep struct {
	ID        string           `json:"id"`
	Request   *WorkflowRequest `json:"request"`
	Condition string           `json:"condition"`
	ForEach   string           `json:"foreach"`
	As        string           `json:"as"`
}

// WorkflowRequest defines an HTTP request in a workflow.
type WorkflowRequest struct {
	Method  string                 `json:"method"`
	URL     string                 `json:"url"`
	Headers map[string]string      `json:"headers"`
	Body    map[string]interface{} `json:"body"`
	Query   map[string]string      `json:"query"`
}

// WorkflowOutput defines workflow output transformation.
type WorkflowOutput struct {
	Format    string `json:"format"`
	Transform string `json:"transform"`
}

// DeprecationInfo represents the x-cli-deprecation extension.
type DeprecationInfo struct {
	Path        string    `json:"path"`
	Deprecated  bool      `json:"deprecated"`
	Sunset      time.Time `json:"sunset"`
	Replacement string    `json:"replacement"`
	Migration   string    `json:"migration"`
	Level       string    `json:"level"` // info, warning, error
}

// CLISecret represents the x-cli-secret extension.
type CLISecret struct {
	Parameter string `json:"parameter"`
	Masked    bool   `json:"masked"`
	Storage   string `json:"storage"` // env, keyring, prompt
	EnvVar    string `json:"env-var"`
}

// CLIPlugin represents the x-cli-plugin extension.
type CLIPlugin struct {
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Command     string   `json:"command"`
	Executable  string   `json:"executable"`
	Args        []string `json:"args"`
}

// CLIFileInput represents the x-cli-file-input extension.
type CLIFileInput struct {
	Parameter   string   `json:"parameter"`
	Accepts     []string `json:"accepts"` // file extensions
	MaxSize     int64    `json:"max-size"`
	Description string   `json:"description"`
}

// CLIWatch represents the x-cli-watch extension.
type CLIWatch struct {
	Enabled        bool             `json:"enabled"`
	Type           string           `json:"type"` // sse, websocket, polling
	Endpoint       string           `json:"endpoint"`
	Events         []string         `json:"events"`
	ExitConditions []*ExitCondition `json:"exit-on"`
	Reconnect      *ReconnectConfig `json:"reconnect"`
}

// ExitCondition defines when to exit watch mode.
type ExitCondition struct {
	Event     string `json:"event"`
	Condition string `json:"condition"`
	Message   string `json:"message"`
}

// ReconnectConfig defines reconnection behavior.
type ReconnectConfig struct {
	Enabled      bool `json:"enabled"`
	MaxAttempts  int  `json:"max-attempts"`
	IntervalSecs int  `json:"interval"`
}

// CLIProgress represents the x-cli-progress extension.
type CLIProgress struct {
	Enabled              *bool  `json:"enabled"`
	Type                 string `json:"type"` // spinner, bar, steps
	ShowStepDescriptions *bool  `json:"show-step-descriptions"`
	ShowTimestamps       *bool  `json:"show-timestamps"`
	Color                string `json:"color"` // auto, always, never
}

// ChangelogEntry represents a single x-cli-changelog entry.
type ChangelogEntry struct {
	Date       string    `json:"date"`
	Version    string    `json:"version"`
	Changes    []*Change `json:"changes"`
	Added      []string  `json:"added,omitempty"`
	Changed    []string  `json:"changed,omitempty"`
	Deprecated []string  `json:"deprecated,omitempty"`
	Removed    []string  `json:"removed,omitempty"`
	Breaking   []string  `json:"breaking,omitempty"`
	IsCurrent  bool      `json:"is_current,omitempty"`
}

// Change represents a single change in the changelog.
type Change struct {
	Type        string `json:"type"`     // added, removed, modified, deprecated, security
	Severity    string `json:"severity"` // breaking, dangerous, safe
	Description string `json:"description"`
	Path        string `json:"path"`
	Migration   string `json:"migration"`
	Sunset      string `json:"sunset"`
}

// ParseSpecExtensions extracts all CLI extensions from the OpenAPI spec.
func ParseSpecExtensions(spec *openapi3.T) (*SpecExtensions, error) {
	extensions := &SpecExtensions{}

	// Parse x-cli-config
	if configData, ok := spec.Extensions["x-cli-config"]; ok {
		config, err := parseCLIConfig(configData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse x-cli-config: %w", err)
		}
		extensions.Config = config
	}

	// Parse x-cli-changelog
	if changelogData, ok := spec.Info.Extensions["x-cli-changelog"]; ok {
		changelog, err := parseChangelog(changelogData)
		if err != nil {
			return nil, fmt.Errorf("failed to parse x-cli-changelog: %w", err)
		}
		extensions.Changelog = changelog
	}

	// Parse x-auth-config from security schemes
	if spec.Components != nil && spec.Components.SecuritySchemes != nil {
		for _, schemeRef := range spec.Components.SecuritySchemes {
			if schemeRef.Value != nil {
				if authConfigData, ok := schemeRef.Value.Extensions["x-auth-config"]; ok {
					authConfig, err := parseAuthConfig(authConfigData)
					if err != nil {
						return nil, fmt.Errorf("failed to parse x-auth-config: %w", err)
					}
					extensions.AuthConfig = authConfig
					break // Use first auth config found
				}
			}
		}
	}

	return extensions, nil
}

// parseCLIConfig parses the x-cli-config extension.
func parseCLIConfig(data interface{}) (*CLIConfig, error) {
	configMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("x-cli-config must be an object")
	}

	config := &CLIConfig{
		Raw: configMap,
	}

	if name, ok := configMap["name"].(string); ok {
		config.Name = name
	}
	if version, ok := configMap["version"].(string); ok {
		config.Version = version
	}
	if desc, ok := configMap["description"].(string); ok {
		config.Description = desc
	}

	// Parse branding
	if brandingData, ok := configMap["branding"].(map[string]interface{}); ok {
		config.Branding = &BrandingConfig{
			Raw: brandingData,
		}
		if art, ok := brandingData["ascii-art"].(string); ok {
			config.Branding.ASCIIArt = art
		}
		if colorsData, ok := brandingData["colors"].(map[string]interface{}); ok {
			config.Branding.Colors = &ColorScheme{}
			if primary, ok := colorsData["primary"].(string); ok {
				config.Branding.Colors.Primary = primary
			}
			if secondary, ok := colorsData["secondary"].(string); ok {
				config.Branding.Colors.Secondary = secondary
			}
		}
	}

	// Parse auth
	if authData, ok := configMap["auth"].(map[string]interface{}); ok {
		config.Auth = &AuthSettings{}
		if authType, ok := authData["type"].(string); ok {
			config.Auth.Type = authType
		}
		if storage, ok := authData["storage"].(string); ok {
			config.Auth.Storage = storage
		}
		if autoRefresh, ok := authData["auto-refresh"].(bool); ok {
			config.Auth.AutoRefresh = autoRefresh
		}
	}

	// Parse output
	if outputData, ok := configMap["output"].(map[string]interface{}); ok {
		config.Output = &OutputSettings{}
		if defaultFmt, ok := outputData["default-format"].(string); ok {
			config.Output.DefaultFormat = defaultFmt
		}
		if formats, ok := outputData["supported-formats"].([]interface{}); ok {
			for _, f := range formats {
				if str, ok := f.(string); ok {
					config.Output.SupportedFormats = append(config.Output.SupportedFormats, str)
				}
			}
		}
	}

	// Parse features
	if featuresData, ok := configMap["features"].(map[string]interface{}); ok {
		config.Features = &FeatureSettings{}
		if interactive, ok := featuresData["interactive-mode"].(bool); ok {
			config.Features.InteractiveMode = interactive
		}
		if autoComplete, ok := featuresData["auto-complete"].(bool); ok {
			config.Features.AutoComplete = autoComplete
		}
		if selfUpdate, ok := featuresData["self-update"].(bool); ok {
			config.Features.SelfUpdate = selfUpdate
		}
		if telemetry, ok := featuresData["telemetry"].(bool); ok {
			config.Features.Telemetry = telemetry
		}
	}

	// Parse cache
	if cacheData, ok := configMap["cache"].(map[string]interface{}); ok {
		config.Cache = &CacheSettings{}
		if enabled, ok := cacheData["enabled"].(bool); ok {
			config.Cache.Enabled = enabled
		}
		if ttl, ok := cacheData["ttl"].(float64); ok {
			config.Cache.TTL = int(ttl)
		}
		if location, ok := cacheData["location"].(string); ok {
			config.Cache.Location = location
		}
	}

	return config, nil
}

// parseAuthConfig parses the x-auth-config extension.
func parseAuthConfig(data interface{}) (*AuthConfig, error) {
	configMap, ok := data.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("x-auth-config must be an object")
	}

	config := &AuthConfig{}

	// Parse flows
	if flowsData, ok := configMap["flows"].(map[string]interface{}); ok {
		config.Flows = &OAuth2Flows{}

		if authCode, ok := flowsData["authorizationCode"].(map[string]interface{}); ok {
			config.Flows.AuthorizationCode = parseOAuth2Flow(authCode)
		}
		if clientCreds, ok := flowsData["clientCredentials"].(map[string]interface{}); ok {
			config.Flows.ClientCredentials = parseOAuth2Flow(clientCreds)
		}
		if implicit, ok := flowsData["implicit"].(map[string]interface{}); ok {
			config.Flows.Implicit = parseOAuth2Flow(implicit)
		}
		if password, ok := flowsData["password"].(map[string]interface{}); ok {
			config.Flows.Password = parseOAuth2Flow(password)
		}
	}

	// Parse token storage
	if storageList, ok := configMap["token-storage"].([]interface{}); ok {
		for _, storageData := range storageList {
			if storageMap, ok := storageData.(map[string]interface{}); ok {
				storage := TokenStorage{}
				if storageType, ok := storageMap["type"].(string); ok {
					storage.Type = storageType
				}
				if path, ok := storageMap["path"].(string); ok {
					storage.Path = path
				}
				if service, ok := storageMap["service"].(string); ok {
					storage.Service = service
				}
				if backends, ok := storageMap["keyring-backends"].(map[string]interface{}); ok {
					storage.KeyringBackends = make(map[string]string)
					for k, v := range backends {
						if str, ok := v.(string); ok {
							storage.KeyringBackends[k] = str
						}
					}
				}
				config.TokenStorage = append(config.TokenStorage, storage)
			}
		}
	}

	return config, nil
}

// parseOAuth2Flow parses a single OAuth2 flow.
func parseOAuth2Flow(data map[string]interface{}) *OAuth2Flow {
	flow := &OAuth2Flow{}

	if authURL, ok := data["authorizationUrl"].(string); ok {
		flow.AuthorizationURL = authURL
	}
	if tokenURL, ok := data["tokenUrl"].(string); ok {
		flow.TokenURL = tokenURL
	}
	if refreshURL, ok := data["refreshUrl"].(string); ok {
		flow.RefreshURL = refreshURL
	}
	if scopes, ok := data["scopes"].(map[string]interface{}); ok {
		flow.Scopes = make(map[string]string)
		for k, v := range scopes {
			if str, ok := v.(string); ok {
				flow.Scopes[k] = str
			}
		}
	}

	return flow
}

// parseChangelog parses the x-cli-changelog extension.
func parseChangelog(data interface{}) ([]ChangelogEntry, error) {
	changelogList, ok := data.([]interface{})
	if !ok {
		return nil, fmt.Errorf("x-cli-changelog must be an array")
	}

	var entries []ChangelogEntry
	for _, entryData := range changelogList {
		entryMap, ok := entryData.(map[string]interface{})
		if !ok {
			continue
		}

		entry := ChangelogEntry{}
		if date, ok := entryMap["date"].(string); ok {
			entry.Date = date
		}
		if version, ok := entryMap["version"].(string); ok {
			entry.Version = version
		}

		if changes, ok := entryMap["changes"].([]interface{}); ok {
			for _, changeData := range changes {
				if changeMap, ok := changeData.(map[string]interface{}); ok {
					change := &Change{}
					if changeType, ok := changeMap["type"].(string); ok {
						change.Type = changeType
					}
					if severity, ok := changeMap["severity"].(string); ok {
						change.Severity = severity
					}
					if desc, ok := changeMap["description"].(string); ok {
						change.Description = desc
					}
					if path, ok := changeMap["path"].(string); ok {
						change.Path = path
					}
					if migration, ok := changeMap["migration"].(string); ok {
						change.Migration = migration
					}
					if sunset, ok := changeMap["sunset"].(string); ok {
						change.Sunset = sunset
					}
					entry.Changes = append(entry.Changes, change)
				}
			}
		}

		entries = append(entries, entry)
	}

	return entries, nil
}

// parseCLIFlags parses the x-cli-flags extension.
func parseCLIFlags(data []interface{}) ([]*CLIFlag, error) {
	var flags []*CLIFlag

	for _, flagData := range data {
		flagMap, ok := flagData.(map[string]interface{})
		if !ok {
			continue
		}

		flag := &CLIFlag{}
		if name, ok := flagMap["name"].(string); ok {
			flag.Name = name
		}
		if source, ok := flagMap["source"].(string); ok {
			flag.Source = source
		}
		if flagName, ok := flagMap["flag"].(string); ok {
			flag.Flag = flagName
		}
		if required, ok := flagMap["required"].(bool); ok {
			flag.Required = required
		}
		if flagType, ok := flagMap["type"].(string); ok {
			flag.Type = flagType
		}
		if desc, ok := flagMap["description"].(string); ok {
			flag.Description = desc
		}
		flag.Default = flagMap["default"]

		if aliases, ok := flagMap["aliases"].([]interface{}); ok {
			for _, alias := range aliases {
				if str, ok := alias.(string); ok {
					flag.Aliases = append(flag.Aliases, str)
				}
			}
		}

		if enum, ok := flagMap["enum"].([]interface{}); ok {
			for _, val := range enum {
				if str, ok := val.(string); ok {
					flag.Enum = append(flag.Enum, str)
				}
			}
		}

		flags = append(flags, flag)
	}

	return flags, nil
}

// parseCLIInteractive parses the x-cli-interactive extension.
func parseCLIInteractive(data map[string]interface{}) (*CLIInteractive, error) {
	interactive := &CLIInteractive{}

	if enabled, ok := data["enabled"].(bool); ok {
		interactive.Enabled = enabled
	}

	if prompts, ok := data["prompts"].([]interface{}); ok {
		for _, promptData := range prompts {
			promptMap, ok := promptData.(map[string]interface{})
			if !ok {
				continue
			}

			prompt := &InteractivePrompt{}
			if param, ok := promptMap["parameter"].(string); ok {
				prompt.Parameter = param
			}
			if promptType, ok := promptMap["type"].(string); ok {
				prompt.Type = promptType
			}
			if message, ok := promptMap["message"].(string); ok {
				prompt.Message = message
			}
			if validation, ok := promptMap["validation"].(string); ok {
				prompt.Validation = validation
			}
			if validationMsg, ok := promptMap["validation-message"].(string); ok {
				prompt.ValidationMessage = validationMsg
			}
			prompt.Default = promptMap["default"]

			if source, ok := promptMap["source"].(map[string]interface{}); ok {
				prompt.Source = &PromptSource{}
				if endpoint, ok := source["endpoint"].(string); ok {
					prompt.Source.Endpoint = endpoint
				}
				if valueField, ok := source["value-field"].(string); ok {
					prompt.Source.ValueField = valueField
				}
				if displayField, ok := source["display-field"].(string); ok {
					prompt.Source.DisplayField = displayField
				}
				if filter, ok := source["filter"].(string); ok {
					prompt.Source.Filter = filter
				}
			}

			interactive.Prompts = append(interactive.Prompts, prompt)
		}
	}

	return interactive, nil
}

// parseCLIPreflight parses the x-cli-preflight extension.
func parseCLIPreflight(data []interface{}) ([]*CLIPreflight, error) {
	var checks []*CLIPreflight

	for _, checkData := range data {
		checkMap, ok := checkData.(map[string]interface{})
		if !ok {
			continue
		}

		check := &CLIPreflight{}
		if name, ok := checkMap["name"].(string); ok {
			check.Name = name
		}
		if desc, ok := checkMap["description"].(string); ok {
			check.Description = desc
		}
		if endpoint, ok := checkMap["endpoint"].(string); ok {
			check.Endpoint = endpoint
		}
		if method, ok := checkMap["method"].(string); ok {
			check.Method = method
		}
		if required, ok := checkMap["required"].(bool); ok {
			check.Required = required
		}
		if skipFlag, ok := checkMap["skip-flag"].(string); ok {
			check.SkipFlag = skipFlag
		}

		checks = append(checks, check)
	}

	return checks, nil
}

// parseCLIConfirmation parses the x-cli-confirmation extension.
func parseCLIConfirmation(data map[string]interface{}) (*CLIConfirmation, error) {
	confirmation := &CLIConfirmation{}

	if enabled, ok := data["enabled"].(bool); ok {
		confirmation.Enabled = enabled
	}
	if message, ok := data["message"].(string); ok {
		confirmation.Message = message
	}
	if flag, ok := data["flag"].(string); ok {
		confirmation.Flag = flag
	}

	return confirmation, nil
}

// parseCLIAsync parses the x-cli-async extension.
func parseCLIAsync(data map[string]interface{}) (*CLIAsync, error) {
	async := &CLIAsync{}

	if enabled, ok := data["enabled"].(bool); ok {
		async.Enabled = enabled
	}
	if statusField, ok := data["status-field"].(string); ok {
		async.StatusField = statusField
	}
	if statusEndpoint, ok := data["status-endpoint"].(string); ok {
		async.StatusEndpoint = statusEndpoint
	}

	if terminalStates, ok := data["terminal-states"].([]interface{}); ok {
		for _, state := range terminalStates {
			if str, ok := state.(string); ok {
				async.TerminalStates = append(async.TerminalStates, str)
			}
		}
	}

	if polling, ok := data["polling"].(map[string]interface{}); ok {
		async.Polling = &PollingConfig{}
		if interval, ok := polling["interval"].(float64); ok {
			async.Polling.Interval = int(interval)
		}
		if timeout, ok := polling["timeout"].(float64); ok {
			async.Polling.Timeout = int(timeout)
		}

		if backoff, ok := polling["backoff"].(map[string]interface{}); ok {
			async.Polling.Backoff = &BackoffConfig{}
			if enabled, ok := backoff["enabled"].(bool); ok {
				async.Polling.Backoff.Enabled = enabled
			}
			if multiplier, ok := backoff["multiplier"].(float64); ok {
				async.Polling.Backoff.Multiplier = multiplier
			}
			if maxInterval, ok := backoff["max-interval"].(float64); ok {
				async.Polling.Backoff.MaxInterval = int(maxInterval)
			}
		}
	}

	return async, nil
}

// parseCLIOutput parses the x-cli-output extension.
func parseCLIOutput(data map[string]interface{}) (*CLIOutput, error) {
	output := &CLIOutput{}

	if format, ok := data["format"].(string); ok {
		output.Format = format
	}
	if successMsg, ok := data["success-message"].(string); ok {
		output.SuccessMessage = successMsg
	}
	if errorMsg, ok := data["error-message"].(string); ok {
		output.ErrorMessage = errorMsg
	}
	if watchStatus, ok := data["watch-status"].(bool); ok {
		output.WatchStatus = watchStatus
	}

	if table, ok := data["table"].(map[string]interface{}); ok {
		output.Table = &TableConfig{}
		if columns, ok := table["columns"].([]interface{}); ok {
			for _, colData := range columns {
				if colMap, ok := colData.(map[string]interface{}); ok {
					col := &TableColumn{}
					if field, ok := colMap["field"].(string); ok {
						col.Field = field
					}
					if header, ok := colMap["header"].(string); ok {
						col.Header = header
					}
					if transform, ok := colMap["transform"].(string); ok {
						col.Transform = transform
					}
					if width, ok := colMap["width"].(float64); ok {
						col.Width = int(width)
					}
					output.Table.Columns = append(output.Table.Columns, col)
				}
			}
		}
	}

	return output, nil
}

// parseCLIWorkflow parses the x-cli-workflow extension.
func parseCLIWorkflow(data map[string]interface{}) (*CLIWorkflow, error) {
	workflow := &CLIWorkflow{}

	if steps, ok := data["steps"].([]interface{}); ok {
		for _, stepData := range steps {
			stepMap, ok := stepData.(map[string]interface{})
			if !ok {
				continue
			}

			step := &WorkflowStep{}
			if id, ok := stepMap["id"].(string); ok {
				step.ID = id
			}
			if condition, ok := stepMap["condition"].(string); ok {
				step.Condition = condition
			}
			if forEach, ok := stepMap["foreach"].(string); ok {
				step.ForEach = forEach
			}
			if as, ok := stepMap["as"].(string); ok {
				step.As = as
			}

			if request, ok := stepMap["request"].(map[string]interface{}); ok {
				step.Request = &WorkflowRequest{}
				if method, ok := request["method"].(string); ok {
					step.Request.Method = method
				}
				if url, ok := request["url"].(string); ok {
					step.Request.URL = url
				}
				if headers, ok := request["headers"].(map[string]interface{}); ok {
					step.Request.Headers = make(map[string]string)
					for k, v := range headers {
						if str, ok := v.(string); ok {
							step.Request.Headers[k] = str
						}
					}
				}
				if body, ok := request["body"].(map[string]interface{}); ok {
					step.Request.Body = body
				}
				if query, ok := request["query"].(map[string]interface{}); ok {
					step.Request.Query = make(map[string]string)
					for k, v := range query {
						if str, ok := v.(string); ok {
							step.Request.Query[k] = str
						}
					}
				}
			}

			workflow.Steps = append(workflow.Steps, step)
		}
	}

	if output, ok := data["output"].(map[string]interface{}); ok {
		workflow.Output = &WorkflowOutput{}
		if format, ok := output["format"].(string); ok {
			workflow.Output.Format = format
		}
		if transform, ok := output["transform"].(string); ok {
			workflow.Output.Transform = transform
		}
	}

	return workflow, nil
}

// Deprecation represents a deprecated API endpoint or parameter.
type Deprecation struct {
	OperationID string    `json:"operation_id"`
	Method      string    `json:"method"`
	Path        string    `json:"path"`
	Parameter   string    `json:"parameter,omitempty"`
	Sunset      time.Time `json:"sunset,omitempty"`
	Replacement string    `json:"replacement,omitempty"`
	Message     string    `json:"message,omitempty"`
	DocsURL     string    `json:"docs_url,omitempty"`
	Severity    string    `json:"severity"` // warning, critical
}
