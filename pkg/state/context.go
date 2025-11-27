package state

import (
	"fmt"
	"time"
)

// Context represents a named context with configuration values.
// Similar to kubectl contexts, this allows switching between different
// environments or configurations.
type Context struct {
	Name        string            `yaml:"name" json:"name"`
	Description string            `yaml:"description,omitempty" json:"description,omitempty"`
	Fields      map[string]string `yaml:"fields,omitempty" json:"fields,omitempty"`
	CreatedAt   time.Time         `yaml:"created_at" json:"created_at"`
	LastUsed    time.Time         `yaml:"last_used,omitempty" json:"last_used,omitempty"`
	UseCount    int               `yaml:"use_count,omitempty" json:"use_count,omitempty"`
}

// NewContext creates a new context with the given name.
func NewContext(name string) *Context {
	return &Context{
		Name:      name,
		Fields:    make(map[string]string),
		CreatedAt: time.Now(),
	}
}

// Set sets a field in the context.
func (c *Context) Set(key, value string) {
	if c.Fields == nil {
		c.Fields = make(map[string]string)
	}
	c.Fields[key] = value
}

// Get gets a field from the context.
func (c *Context) Get(key string) (string, bool) {
	if c.Fields == nil {
		return "", false
	}
	value, exists := c.Fields[key]
	return value, exists
}

// Delete deletes a field from the context.
func (c *Context) Delete(key string) {
	if c.Fields != nil {
		delete(c.Fields, key)
	}
}

// Has checks if a field exists in the context.
func (c *Context) Has(key string) bool {
	if c.Fields == nil {
		return false
	}
	_, exists := c.Fields[key]
	return exists
}

// Clear clears all fields in the context.
func (c *Context) Clear() {
	c.Fields = make(map[string]string)
}

// Clone creates a deep copy of the context.
func (c *Context) Clone() *Context {
	clone := &Context{
		Name:        c.Name,
		Description: c.Description,
		Fields:      make(map[string]string),
		CreatedAt:   c.CreatedAt,
		LastUsed:    c.LastUsed,
		UseCount:    c.UseCount,
	}

	for k, v := range c.Fields {
		clone.Fields[k] = v
	}

	return clone
}

// Merge merges another context into this one.
// Fields from the other context will overwrite existing fields.
func (c *Context) Merge(other *Context) {
	if c.Fields == nil {
		c.Fields = make(map[string]string)
	}

	for k, v := range other.Fields {
		c.Fields[k] = v
	}
}

// MarkUsed updates the last used time and increments use count.
func (c *Context) MarkUsed() {
	c.LastUsed = time.Now()
	c.UseCount++
}

// Validate validates the context.
func (c *Context) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("context name cannot be empty")
	}
	return nil
}

// ContextBuilder provides a fluent API for building contexts.
type ContextBuilder struct {
	context *Context
}

// NewContextBuilder creates a new context builder.
func NewContextBuilder(name string) *ContextBuilder {
	return &ContextBuilder{
		context: NewContext(name),
	}
}

// WithDescription sets the context description.
func (b *ContextBuilder) WithDescription(desc string) *ContextBuilder {
	b.context.Description = desc
	return b
}

// WithField sets a field in the context.
func (b *ContextBuilder) WithField(key, value string) *ContextBuilder {
	b.context.Set(key, value)
	return b
}

// WithFields sets multiple fields in the context.
func (b *ContextBuilder) WithFields(fields map[string]string) *ContextBuilder {
	for k, v := range fields {
		b.context.Set(k, v)
	}
	return b
}

// Build returns the built context.
func (b *ContextBuilder) Build() *Context {
	return b.context
}

// ContextField represents a field definition for contexts.
type ContextField struct {
	Name        string   `yaml:"name" json:"name"`
	Description string   `yaml:"description,omitempty" json:"description,omitempty"`
	Type        string   `yaml:"type,omitempty" json:"type,omitempty"`       // string, int, bool
	EnvVar      string   `yaml:"env_var,omitempty" json:"env_var,omitempty"` // Environment variable to read from
	Default     string   `yaml:"default,omitempty" json:"default,omitempty"`
	Required    bool     `yaml:"required,omitempty" json:"required,omitempty"`
	ValidValues []string `yaml:"valid_values,omitempty" json:"valid_values,omitempty"` // For enum-like fields
}

// ContextManager provides high-level context management operations.
type ContextManager struct {
	stateMgr *Manager
}

// NewContextManager creates a new context manager.
func NewContextManager(stateMgr *Manager) *ContextManager {
	return &ContextManager{
		stateMgr: stateMgr,
	}
}

// SwitchTo switches to a named context.
func (cm *ContextManager) SwitchTo(name string) error {
	// Mark the context as used
	ctx, err := cm.stateMgr.GetContext(name)
	if err != nil {
		return err
	}

	ctx.MarkUsed()
	if err := cm.stateMgr.UpdateContext(name, ctx); err != nil {
		return err
	}

	// Set as current
	return cm.stateMgr.SetCurrentContext(name)
}

// Create creates a new context.
func (cm *ContextManager) Create(name, description string, fields map[string]string) error {
	ctx := NewContext(name)
	ctx.Description = description
	ctx.Fields = fields

	return cm.stateMgr.CreateContext(name, ctx)
}

// Update updates an existing context.
func (cm *ContextManager) Update(name string, fields map[string]string) error {
	ctx, err := cm.stateMgr.GetContext(name)
	if err != nil {
		return err
	}

	for k, v := range fields {
		ctx.Set(k, v)
	}

	return cm.stateMgr.UpdateContext(name, ctx)
}

// Delete deletes a context.
func (cm *ContextManager) Delete(name string) error {
	return cm.stateMgr.DeleteContext(name)
}

// List returns all contexts.
func (cm *ContextManager) List() (map[string]*Context, error) {
	names := cm.stateMgr.ListContexts()
	contexts := make(map[string]*Context, len(names))

	for _, name := range names {
		ctx, err := cm.stateMgr.GetContext(name)
		if err != nil {
			return nil, err
		}
		contexts[name] = ctx
	}

	return contexts, nil
}

// Current returns the current context.
func (cm *ContextManager) Current() *Context {
	return cm.stateMgr.GetCurrentContext()
}

// CurrentName returns the name of the current context.
func (cm *ContextManager) CurrentName() string {
	return cm.stateMgr.state.CurrentContext
}

// Rename renames a context.
func (cm *ContextManager) Rename(oldName, newName string) error {
	ctx, err := cm.stateMgr.GetContext(oldName)
	if err != nil {
		return err
	}

	// Clone to new name
	clone := ctx.Clone()
	clone.Name = newName

	// Create new context
	if err := cm.stateMgr.CreateContext(newName, clone); err != nil {
		return err
	}

	// Update current context if needed
	if cm.CurrentName() == oldName {
		if err := cm.stateMgr.SetCurrentContext(newName); err != nil {
			return err
		}
	}

	// Delete old context
	return cm.stateMgr.DeleteContext(oldName)
}

// Export exports a context as a map for serialization.
func (cm *ContextManager) Export(name string) (map[string]interface{}, error) {
	ctx, err := cm.stateMgr.GetContext(name)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"name":        ctx.Name,
		"description": ctx.Description,
		"fields":      ctx.Fields,
		"created_at":  ctx.CreatedAt,
		"last_used":   ctx.LastUsed,
		"use_count":   ctx.UseCount,
	}, nil
}

// Import imports a context from a map.
func (cm *ContextManager) Import(data map[string]interface{}) error {
	name, ok := data["name"].(string)
	if !ok || name == "" {
		return fmt.Errorf("invalid context name")
	}

	ctx := NewContext(name)

	if desc, ok := data["description"].(string); ok {
		ctx.Description = desc
	}

	if fields, ok := data["fields"].(map[string]string); ok {
		ctx.Fields = fields
	} else if fields, ok := data["fields"].(map[string]interface{}); ok {
		// Handle case where fields are map[string]interface{}
		ctx.Fields = make(map[string]string)
		for k, v := range fields {
			if strVal, ok := v.(string); ok {
				ctx.Fields[k] = strVal
			}
		}
	}

	return cm.stateMgr.CreateContext(name, ctx)
}
