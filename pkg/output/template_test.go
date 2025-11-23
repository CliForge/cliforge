package output

import (
	"strings"
	"testing"
)

func TestNewTemplateEngine(t *testing.T) {
	engine := NewTemplateEngine()
	if engine == nil {
		t.Error("Expected engine to be created")
	}
	if engine.programCache == nil {
		t.Error("Expected programCache to be initialized")
	}
}

func TestTemplateEngineRenderSimpleVariables(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "single variable",
			template: "Hello {name}",
			data:     map[string]interface{}{"name": "World"},
			expected: "Hello World",
		},
		{
			name:     "multiple variables",
			template: "{greeting} {name}, your score is {score}",
			data: map[string]interface{}{
				"greeting": "Hello",
				"name":     "Alice",
				"score":    100,
			},
			expected: "Hello Alice, your score is 100",
		},
		{
			name:     "nested field",
			template: "Email: {user.email}",
			data: map[string]interface{}{
				"user": map[string]interface{}{
					"email": "test@example.com",
				},
			},
			expected: "Email: test@example.com",
		},
		{
			name:     "no variables",
			template: "Static text",
			data:     map[string]interface{}{},
			expected: "Static text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, tt.data)
			if err != nil {
				t.Errorf("Render failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestTemplateEngineRenderExpressions(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "arithmetic expression",
			template: "Total: {{count * price}}",
			data: map[string]interface{}{
				"count": 5,
				"price": 10,
			},
			expected: "Total: 50",
		},
		{
			name:     "string concatenation",
			template: "Name: {{first + ' ' + last}}",
			data: map[string]interface{}{
				"first": "John",
				"last":  "Doe",
			},
			expected: "Name: John Doe",
		},
		{
			name:     "conditional expression",
			template: "Status: {{active ? 'Active' : 'Inactive'}}",
			data: map[string]interface{}{
				"active": true,
			},
			expected: "Status: Active",
		},
		{
			name:     "comparison",
			template: "Result: {{score >= 90}}",
			data: map[string]interface{}{
				"score": 95,
			},
			expected: "Result: true",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.Render(tt.template, tt.data)
			if err != nil {
				t.Errorf("Render failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestTemplateEngineRenderMixed(t *testing.T) {
	engine := NewTemplateEngine()

	template := "Cluster {name} is {{state == 'ready' ? 'ready' : 'not ready'}}"
	data := map[string]interface{}{
		"name":  "cluster-1",
		"state": "ready",
	}

	result, err := engine.Render(template, data)
	if err != nil {
		t.Errorf("Render failed: %v", err)
	}

	expected := "Cluster cluster-1 is ready"
	if result != expected {
		t.Errorf("Expected '%s', got '%s'", expected, result)
	}
}

func TestTemplateEngineRenderSuccess(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		expected string
	}{
		{
			name:     "with template",
			template: "Created {resource}",
			data:     map[string]interface{}{"resource": "cluster-1"},
			expected: "Created cluster-1",
		},
		{
			name:     "empty template",
			template: "",
			data:     map[string]interface{}{},
			expected: "Operation completed successfully",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.RenderSuccess(tt.template, tt.data)
			if err != nil {
				t.Errorf("RenderSuccess failed: %v", err)
			}
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestTemplateEngineRenderError(t *testing.T) {
	engine := NewTemplateEngine()

	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		check    func(t *testing.T, result string)
	}{
		{
			name:     "with template",
			template: "Failed to create {resource}: {error}",
			data: map[string]interface{}{
				"resource": "cluster-1",
				"error":    "quota exceeded",
			},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "Failed to create cluster-1") {
					t.Errorf("Expected error message, got '%s'", result)
				}
			},
		},
		{
			name:     "empty template with error data",
			template: "",
			data:     map[string]interface{}{"error": "something went wrong"},
			check: func(t *testing.T, result string) {
				if result != "something went wrong" {
					t.Errorf("Expected error message, got '%s'", result)
				}
			},
		},
		{
			name:     "empty template without error data",
			template: "",
			data:     map[string]interface{}{},
			check: func(t *testing.T, result string) {
				if result != "Operation failed" {
					t.Errorf("Expected default error message, got '%s'", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := engine.RenderError(tt.template, tt.data)
			if err != nil {
				t.Errorf("RenderError failed: %v", err)
			}
			tt.check(t, result)
		})
	}
}

func TestTemplateEngineClearCache(t *testing.T) {
	engine := NewTemplateEngine()

	// Render something to populate cache
	_, _ = engine.Render("{{1 + 1}}", nil)

	if len(engine.programCache) == 0 {
		t.Error("Expected cache to be populated")
	}

	engine.ClearCache()

	if len(engine.programCache) != 0 {
		t.Error("Expected cache to be cleared")
	}
}

func TestTemplateEnginePrecompile(t *testing.T) {
	engine := NewTemplateEngine()

	template := "Result: {{x + y}}"
	env := map[string]interface{}{
		"x": 0,
		"y": 0,
	}

	if err := engine.PrecompileTemplate(template, env); err != nil {
		t.Errorf("PrecompileTemplate failed: %v", err)
	}

	if len(engine.programCache) == 0 {
		t.Error("Expected cache to be populated")
	}

	// Now render with actual data
	result, err := engine.Render(template, map[string]interface{}{
		"x": 5,
		"y": 3,
	})
	if err != nil {
		t.Errorf("Render failed: %v", err)
	}
	if result != "Result: 8" {
		t.Errorf("Expected 'Result: 8', got '%s'", result)
	}
}

func TestMessageTemplate(t *testing.T) {
	template := NewMessageTemplate(
		"test",
		"Hello {name}",
		"Test template",
	)

	result, err := template.Render(map[string]interface{}{
		"name": "World",
	})
	if err != nil {
		t.Errorf("Render failed: %v", err)
	}
	if result != "Hello World" {
		t.Errorf("Expected 'Hello World', got '%s'", result)
	}
}

func TestTemplateLibrary(t *testing.T) {
	lib := NewTemplateLibrary()

	template := NewMessageTemplate(
		"greeting",
		"Hello {name}",
		"Greeting template",
	)
	lib.Add(template)

	// Test Get
	retrieved, err := lib.Get("greeting")
	if err != nil {
		t.Errorf("Get failed: %v", err)
	}
	if retrieved.Name != "greeting" {
		t.Error("Expected to retrieve 'greeting' template")
	}

	// Test Get non-existent
	_, err = lib.Get("nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent template")
	}

	// Test Render
	result, err := lib.Render("greeting", map[string]interface{}{
		"name": "Alice",
	})
	if err != nil {
		t.Errorf("Render failed: %v", err)
	}
	if result != "Hello Alice" {
		t.Errorf("Expected 'Hello Alice', got '%s'", result)
	}

	// Test List
	names := lib.List()
	if len(names) != 1 || names[0] != "greeting" {
		t.Error("Expected List to return ['greeting']")
	}
}

func TestDefaultTemplates(t *testing.T) {
	tests := []struct {
		name     string
		template string
		data     map[string]interface{}
		check    func(t *testing.T, result string)
	}{
		{
			name:     "created",
			template: "created",
			data: map[string]interface{}{
				"resource_type": "Cluster",
				"name":          "test-cluster",
			},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "Cluster") || !strings.Contains(result, "test-cluster") {
					t.Errorf("Unexpected result: %s", result)
				}
			},
		},
		{
			name:     "updated",
			template: "updated",
			data: map[string]interface{}{
				"resource_type": "User",
				"name":          "john",
			},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "User") || !strings.Contains(result, "john") {
					t.Errorf("Unexpected result: %s", result)
				}
			},
		},
		{
			name:     "deleted",
			template: "deleted",
			data: map[string]interface{}{
				"resource_type": "Instance",
				"name":          "i-123456",
			},
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "Instance") || !strings.Contains(result, "i-123456") {
					t.Errorf("Unexpected result: %s", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := DefaultTemplates.Render(tt.template, tt.data)
			if err != nil {
				t.Errorf("Render failed: %v", err)
			}
			tt.check(t, result)
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	tests := []struct {
		name     string
		fn       func(string, string) string
		resource string
		objName  string
		check    func(t *testing.T, result string)
	}{
		{
			name:     "RenderCreated",
			fn:       RenderCreated,
			resource: "Cluster",
			objName:  "test",
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "created") {
					t.Errorf("Expected 'created' in result: %s", result)
				}
			},
		},
		{
			name:     "RenderUpdated",
			fn:       RenderUpdated,
			resource: "User",
			objName:  "alice",
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "updated") {
					t.Errorf("Expected 'updated' in result: %s", result)
				}
			},
		},
		{
			name:     "RenderDeleted",
			fn:       RenderDeleted,
			resource: "Instance",
			objName:  "i-123",
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "deleted") {
					t.Errorf("Expected 'deleted' in result: %s", result)
				}
			},
		},
		{
			name:     "RenderNotFound",
			fn:       RenderNotFound,
			resource: "Pod",
			objName:  "my-pod",
			check: func(t *testing.T, result string) {
				if !strings.Contains(result, "not found") {
					t.Errorf("Expected 'not found' in result: %s", result)
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.fn(tt.resource, tt.objName)
			tt.check(t, result)
		})
	}
}
