package state

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// DefaultsProvider provides smart defaults based on usage, context, and configuration.
type DefaultsProvider struct {
	stateMgr        *Manager
	configDefaults  map[string]string
	builtinDefaults map[string]string
}

// NewDefaultsProvider creates a new defaults provider.
func NewDefaultsProvider(stateMgr *Manager) *DefaultsProvider {
	return &DefaultsProvider{
		stateMgr:        stateMgr,
		configDefaults:  make(map[string]string),
		builtinDefaults: make(map[string]string),
	}
}

// SetConfigDefaults sets configuration-based defaults.
func (d *DefaultsProvider) SetConfigDefaults(defaults map[string]string) {
	d.configDefaults = defaults
}

// SetBuiltinDefaults sets built-in defaults.
func (d *DefaultsProvider) SetBuiltinDefaults(defaults map[string]string) {
	d.builtinDefaults = defaults
}

// Get retrieves a default value with the following priority:
// 1. Recent (from usage)
// 2. Context (from current context)
// 3. Config (from configuration)
// 4. Built-in (from code)
func (d *DefaultsProvider) Get(key string) (string, bool) {
	// 1. Check recent values
	if recentValue := d.getFromRecent(key); recentValue != "" {
		return recentValue, true
	}

	// 2. Check current context
	if contextValue := d.getFromContext(key); contextValue != "" {
		return contextValue, true
	}

	// 3. Check config defaults
	if configValue, exists := d.configDefaults[key]; exists && configValue != "" {
		return configValue, true
	}

	// 4. Check built-in defaults
	if builtinValue, exists := d.builtinDefaults[key]; exists && builtinValue != "" {
		return builtinValue, true
	}

	return "", false
}

// GetWithPriority retrieves a default value and returns the priority level.
func (d *DefaultsProvider) GetWithPriority(key string) (string, DefaultPriority, bool) {
	// 1. Recent
	if recentValue := d.getFromRecent(key); recentValue != "" {
		return recentValue, PriorityRecent, true
	}

	// 2. Context
	if contextValue := d.getFromContext(key); contextValue != "" {
		return contextValue, PriorityContext, true
	}

	// 3. Config
	if configValue, exists := d.configDefaults[key]; exists && configValue != "" {
		return configValue, PriorityConfig, true
	}

	// 4. Built-in
	if builtinValue, exists := d.builtinDefaults[key]; exists && builtinValue != "" {
		return builtinValue, PriorityBuiltin, true
	}

	return "", PriorityNone, false
}

// GetAll returns all defaults merged by priority.
func (d *DefaultsProvider) GetAll() map[string]string {
	merged := make(map[string]string)

	// Start with built-in (lowest priority)
	for k, v := range d.builtinDefaults {
		merged[k] = v
	}

	// Override with config
	for k, v := range d.configDefaults {
		if v != "" {
			merged[k] = v
		}
	}

	// Override with context
	ctx := d.stateMgr.GetCurrentContext()
	if ctx != nil {
		for k, v := range ctx.Fields {
			if v != "" {
				merged[k] = v
			}
		}
	}

	// Override with recent (highest priority)
	// Note: This is tricky because recent values are per-list
	// We'll add them if they match known keys
	for key := range merged {
		if recentValue := d.getFromRecent(key); recentValue != "" {
			merged[key] = recentValue
		}
	}

	return merged
}

// GetWithFallback gets a value with a custom fallback.
func (d *DefaultsProvider) GetWithFallback(key, fallback string) string {
	if value, exists := d.Get(key); exists {
		return value
	}
	return fallback
}

// GetInt gets a default value as an integer.
func (d *DefaultsProvider) GetInt(key string, fallback int) int {
	value, exists := d.Get(key)
	if !exists {
		return fallback
	}

	intValue, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}

	return intValue
}

// GetBool gets a default value as a boolean.
func (d *DefaultsProvider) GetBool(key string, fallback bool) bool {
	value, exists := d.Get(key)
	if !exists {
		return fallback
	}

	boolValue, err := strconv.ParseBool(value)
	if err != nil {
		return fallback
	}

	return boolValue
}

// GetString gets a default value as a string.
func (d *DefaultsProvider) GetString(key, fallback string) string {
	return d.GetWithFallback(key, fallback)
}

// getFromRecent retrieves the most recent value for a key.
func (d *DefaultsProvider) getFromRecent(key string) string {
	recent := d.stateMgr.GetRecent()
	if recent == nil {
		return ""
	}

	// Try to get from a list named after the key
	values := recent.Get(key)
	if len(values) > 0 {
		return values[0] // Most recent
	}

	// Also try plural form (e.g., "cluster" -> "clusters")
	values = recent.Get(key + "s")
	if len(values) > 0 {
		return values[0]
	}

	return ""
}

// getFromContext retrieves a value from the current context.
func (d *DefaultsProvider) getFromContext(key string) string {
	ctx := d.stateMgr.GetCurrentContext()
	if ctx == nil {
		return ""
	}

	// Try exact key
	if value, exists := ctx.Get(key); exists {
		return value
	}

	// Try normalized key (e.g., "cluster-id" -> "cluster_id")
	normalizedKey := normalizeKey(key)
	if value, exists := ctx.Get(normalizedKey); exists {
		return value
	}

	return ""
}

// normalizeKey normalizes a key for lookup (handles - vs _).
func normalizeKey(key string) string {
	// Convert hyphens to underscores
	return strings.ReplaceAll(key, "-", "_")
}

// DefaultPriority represents the priority level of a default value.
type DefaultPriority int

const (
	// PriorityNone means no default was found.
	PriorityNone DefaultPriority = iota

	// PriorityBuiltin means the value came from built-in defaults.
	PriorityBuiltin

	// PriorityConfig means the value came from configuration.
	PriorityConfig

	// PriorityContext means the value came from the current context.
	PriorityContext

	// PriorityRecent means the value came from recent usage.
	PriorityRecent
)

// String returns a string representation of the priority.
func (p DefaultPriority) String() string {
	switch p {
	case PriorityNone:
		return "none"
	case PriorityBuiltin:
		return "builtin"
	case PriorityConfig:
		return "config"
	case PriorityContext:
		return "context"
	case PriorityRecent:
		return "recent"
	default:
		return fmt.Sprintf("unknown(%d)", p)
	}
}

// DefaultsResolver resolves defaults with environment variable support.
type DefaultsResolver struct {
	provider *DefaultsProvider
	envVars  map[string]string // Maps keys to env var names
}

// NewDefaultsResolver creates a new defaults resolver.
func NewDefaultsResolver(provider *DefaultsProvider) *DefaultsResolver {
	return &DefaultsResolver{
		provider: provider,
		envVars:  make(map[string]string),
	}
}

// SetEnvVar maps a key to an environment variable name.
func (r *DefaultsResolver) SetEnvVar(key, envVarName string) {
	r.envVars[key] = envVarName
}

// Resolve resolves a default value with environment variable override.
// Priority: ENV > Recent > Context > Config > Built-in
func (r *DefaultsResolver) Resolve(key string) (string, bool) {
	// Check environment variable first
	if envVarName, exists := r.envVars[key]; exists {
		if envValue := os.Getenv(envVarName); envValue != "" {
			return envValue, true
		}
	}

	// Fall back to provider
	return r.provider.Get(key)
}

// ResolveWithPriority resolves with priority information.
func (r *DefaultsResolver) ResolveWithPriority(key string) (string, DefaultPriority, bool) {
	// Check environment variable first
	if envVarName, exists := r.envVars[key]; exists {
		if envValue := os.Getenv(envVarName); envValue != "" {
			return envValue, PriorityEnv, true
		}
	}

	// Fall back to provider
	return r.provider.GetWithPriority(key)
}

// PriorityEnv is the highest priority for environment variables.
const PriorityEnv DefaultPriority = 100

// DefaultsBuilder provides a fluent API for building defaults.
type DefaultsBuilder struct {
	defaults map[string]string
}

// NewDefaultsBuilder creates a new defaults builder.
func NewDefaultsBuilder() *DefaultsBuilder {
	return &DefaultsBuilder{
		defaults: make(map[string]string),
	}
}

// Set sets a default value.
func (b *DefaultsBuilder) Set(key, value string) *DefaultsBuilder {
	b.defaults[key] = value
	return b
}

// SetInt sets an integer default value.
func (b *DefaultsBuilder) SetInt(key string, value int) *DefaultsBuilder {
	b.defaults[key] = strconv.Itoa(value)
	return b
}

// SetBool sets a boolean default value.
func (b *DefaultsBuilder) SetBool(key string, value bool) *DefaultsBuilder {
	b.defaults[key] = strconv.FormatBool(value)
	return b
}

// SetMultiple sets multiple defaults at once.
func (b *DefaultsBuilder) SetMultiple(defaults map[string]string) *DefaultsBuilder {
	for k, v := range defaults {
		b.defaults[k] = v
	}
	return b
}

// Build returns the built defaults map.
func (b *DefaultsBuilder) Build() map[string]string {
	return b.defaults
}

// MergeDefaults merges multiple default maps with priority.
// Later maps override earlier ones.
func MergeDefaults(maps ...map[string]string) map[string]string {
	result := make(map[string]string)

	for _, m := range maps {
		for k, v := range m {
			if v != "" {
				result[k] = v
			}
		}
	}

	return result
}

// FilterDefaults filters defaults by key prefix.
func FilterDefaults(defaults map[string]string, prefix string) map[string]string {
	filtered := make(map[string]string)

	for k, v := range defaults {
		if strings.HasPrefix(k, prefix) {
			filtered[k] = v
		}
	}

	return filtered
}

// TransformDefaults transforms default keys using a function.
func TransformDefaults(defaults map[string]string, fn func(string) string) map[string]string {
	transformed := make(map[string]string)

	for k, v := range defaults {
		newKey := fn(k)
		transformed[newKey] = v
	}

	return transformed
}

// ValidateDefaults validates defaults against a set of allowed keys.
func ValidateDefaults(defaults map[string]string, allowedKeys []string) error {
	allowed := make(map[string]bool)
	for _, key := range allowedKeys {
		allowed[key] = true
	}

	for key := range defaults {
		if !allowed[key] {
			return fmt.Errorf("invalid default key: %s", key)
		}
	}

	return nil
}

// DefaultsSnapshot captures a snapshot of all defaults for debugging.
type DefaultsSnapshot struct {
	Timestamp       string            `json:"timestamp"`
	Context         string            `json:"context"`
	BuiltinDefaults map[string]string `json:"builtin_defaults"`
	ConfigDefaults  map[string]string `json:"config_defaults"`
	ContextDefaults map[string]string `json:"context_defaults"`
	RecentDefaults  map[string]string `json:"recent_defaults"`
	MergedDefaults  map[string]string `json:"merged_defaults"`
	Priorities      map[string]string `json:"priorities"`
}

// Snapshot creates a snapshot of all defaults.
func (d *DefaultsProvider) Snapshot() *DefaultsSnapshot {
	snapshot := &DefaultsSnapshot{
		Timestamp:       fmt.Sprintf("%v", os.Getenv("timestamp")),
		BuiltinDefaults: make(map[string]string),
		ConfigDefaults:  make(map[string]string),
		ContextDefaults: make(map[string]string),
		RecentDefaults:  make(map[string]string),
		MergedDefaults:  make(map[string]string),
		Priorities:      make(map[string]string),
	}

	// Copy built-in
	for k, v := range d.builtinDefaults {
		snapshot.BuiltinDefaults[k] = v
	}

	// Copy config
	for k, v := range d.configDefaults {
		snapshot.ConfigDefaults[k] = v
	}

	// Get context defaults
	ctx := d.stateMgr.GetCurrentContext()
	if ctx != nil {
		snapshot.Context = ctx.Name
		for k, v := range ctx.Fields {
			snapshot.ContextDefaults[k] = v
		}
	}

	// Get recent defaults (sample known keys)
	allKeys := make(map[string]bool)
	for k := range d.builtinDefaults {
		allKeys[k] = true
	}
	for k := range d.configDefaults {
		allKeys[k] = true
	}

	for key := range allKeys {
		if recentValue := d.getFromRecent(key); recentValue != "" {
			snapshot.RecentDefaults[key] = recentValue
		}
	}

	// Get merged and priorities
	for key := range allKeys {
		value, priority, exists := d.GetWithPriority(key)
		if exists {
			snapshot.MergedDefaults[key] = value
			snapshot.Priorities[key] = priority.String()
		}
	}

	return snapshot
}
