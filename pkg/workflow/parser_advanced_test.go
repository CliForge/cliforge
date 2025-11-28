package workflow

import (
	"testing"
)

func TestParser_CollectReferences_APICall(t *testing.T) {
	parser := NewParser(&Workflow{})

	step := &Step{
		ID:        "test",
		Type:      StepTypeAPICall,
		Condition: "{flags.enabled} == true",
		APICall: &APICallStep{
			Endpoint: "/api/{steps.step1.id}",
			Headers: map[string]string{
				"Authorization": "Bearer {steps.auth.token}",
			},
			Query: map[string]string{
				"filter": "{flags.filter}",
			},
			Body: map[string]interface{}{
				"user_id": "{steps.user.id}",
				"nested": map[string]interface{}{
					"field": "{steps.data.field}",
				},
			},
		},
		Output: map[string]string{
			"result": "{response.data}",
		},
	}

	refs := parser.collectReferences(step)

	if len(refs) < 5 {
		t.Errorf("expected at least 5 references, got %d: %v", len(refs), refs)
	}

	// Check that condition is collected
	found := false
	for _, ref := range refs {
		if ref == "{flags.enabled} == true" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected condition to be in references")
	}
}

func TestParser_CollectReferences_Plugin(t *testing.T) {
	parser := NewParser(&Workflow{})

	step := &Step{
		ID:   "test",
		Type: StepTypePlugin,
		Plugin: &PluginStep{
			Plugin:  "{flags.plugin_name}",
			Command: "{flags.command}",
			Input: map[string]interface{}{
				"param": "{steps.prev.output}",
			},
		},
	}

	refs := parser.collectReferences(step)

	if len(refs) < 2 {
		t.Errorf("expected at least 2 references, got %d", len(refs))
	}

	// Check plugin and command are collected
	foundPlugin := false
	foundCommand := false
	for _, ref := range refs {
		if ref == "{flags.plugin_name}" {
			foundPlugin = true
		}
		if ref == "{flags.command}" {
			foundCommand = true
		}
	}
	if !foundPlugin || !foundCommand {
		t.Error("expected plugin and command to be in references")
	}
}

func TestParser_CollectReferences_Conditional(t *testing.T) {
	parser := NewParser(&Workflow{})

	step := &Step{
		ID:   "test",
		Type: StepTypeConditional,
		Conditional: &ConditionalStep{
			Condition: "{steps.check.result} == 'success'",
		},
	}

	refs := parser.collectReferences(step)

	found := false
	for _, ref := range refs {
		if ref == "{steps.check.result} == 'success'" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected conditional condition to be in references")
	}
}

func TestParser_CollectReferences_Loop(t *testing.T) {
	parser := NewParser(&Workflow{})

	step := &Step{
		ID:   "test",
		Type: StepTypeLoop,
		Loop: &LoopStep{
			Collection: "{steps.fetch.items}",
		},
	}

	refs := parser.collectReferences(step)

	found := false
	for _, ref := range refs {
		if ref == "{steps.fetch.items}" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected loop collection to be in references")
	}
}

func TestParser_CollectReferences_Wait(t *testing.T) {
	parser := NewParser(&Workflow{})

	step := &Step{
		ID:   "test",
		Type: StepTypeWait,
		Wait: &WaitStep{
			Condition: "{response.status} == 'ready'",
			Polling: &PollingConfig{
				Endpoint: "/api/status/{steps.create.id}",
			},
		},
	}

	refs := parser.collectReferences(step)

	foundCondition := false
	foundEndpoint := false
	for _, ref := range refs {
		if ref == "{response.status} == 'ready'" {
			foundCondition = true
		}
		if ref == "/api/status/{steps.create.id}" {
			foundEndpoint = true
		}
	}
	if !foundCondition || !foundEndpoint {
		t.Error("expected wait condition and polling endpoint to be in references")
	}
}

func TestParser_CollectMapValues_NestedStructures(t *testing.T) {
	parser := NewParser(&Workflow{})

	testMap := map[string]interface{}{
		"string_field": "{steps.step1.value}",
		"nested_map": map[string]interface{}{
			"inner_field": "{steps.step2.value}",
			"deep_nested": map[string]interface{}{
				"deep_field": "{steps.step3.value}",
			},
		},
		"array_field": []interface{}{
			"{steps.step4.value}",
			map[string]interface{}{
				"in_array": "{steps.step5.value}",
			},
		},
		"number_field": 42,
		"bool_field":   true,
	}

	values := parser.collectMapValues(testMap)

	// Should collect all string values including nested ones
	expectedValues := []string{
		"{steps.step1.value}",
		"{steps.step2.value}",
		"{steps.step3.value}",
		"{steps.step4.value}",
		"{steps.step5.value}",
	}

	for _, expected := range expectedValues {
		found := false
		for _, value := range values {
			if value == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find %s in collected values: %v", expected, values)
		}
	}
}

func TestParser_CollectMapValues_ArrayOfStrings(t *testing.T) {
	parser := NewParser(&Workflow{})

	testMap := map[string]interface{}{
		"tags": []interface{}{
			"{steps.step1.tag1}",
			"{steps.step1.tag2}",
			"{steps.step1.tag3}",
		},
	}

	values := parser.collectMapValues(testMap)

	if len(values) < 3 {
		t.Errorf("expected at least 3 values from array, got %d", len(values))
	}
}

func TestParser_CollectMapValues_ArrayOfMaps(t *testing.T) {
	parser := NewParser(&Workflow{})

	testMap := map[string]interface{}{
		"items": []interface{}{
			map[string]interface{}{
				"id":   "{steps.step1.id1}",
				"name": "{steps.step1.name1}",
			},
			map[string]interface{}{
				"id":   "{steps.step1.id2}",
				"name": "{steps.step1.name2}",
			},
		},
	}

	values := parser.collectMapValues(testMap)

	expectedCount := 4
	if len(values) < expectedCount {
		t.Errorf("expected at least %d values from array of maps, got %d", expectedCount, len(values))
	}
}

func TestParser_CollectMapValues_EmptyMap(t *testing.T) {
	parser := NewParser(&Workflow{})

	testMap := map[string]interface{}{}

	values := parser.collectMapValues(testMap)

	if len(values) != 0 {
		t.Errorf("expected 0 values from empty map, got %d", len(values))
	}
}

func TestParser_CollectMapValues_MixedTypes(t *testing.T) {
	parser := NewParser(&Workflow{})

	testMap := map[string]interface{}{
		"string":  "{steps.step1.value}",
		"number":  123,
		"boolean": true,
		"null":    nil,
		"array": []interface{}{
			"plain string",
			42,
			"{steps.step2.value}",
		},
	}

	values := parser.collectMapValues(testMap)

	// Should only collect string values
	expectedStrings := []string{
		"{steps.step1.value}",
		"plain string",
		"{steps.step2.value}",
	}

	for _, expected := range expectedStrings {
		found := false
		for _, value := range values {
			if value == expected {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("expected to find string %s in values", expected)
		}
	}
}

func TestParser_BuildImplicitDependencies_ShortFormat(t *testing.T) {
	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "step1",
				Type: StepTypeAPICall,
				Output: map[string]string{
					"id": "response.id",
				},
			},
			{
				ID:   "step2",
				Type: StepTypeAPICall,
				APICall: &APICallStep{
					// Short format: {step_id.output}
					Endpoint: "/api/resource/{step1.id}",
				},
			},
		},
	}

	parser := NewParser(workflow)
	_, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that step2 depends on step1
	node2 := parser.dag.Nodes["step2"]
	if node2 == nil {
		t.Fatal("step2 node not found")
	}

	foundDep := false
	for _, dep := range node2.Dependencies {
		if dep == "step1" {
			foundDep = true
			break
		}
	}

	if !foundDep {
		t.Error("expected step2 to depend on step1 (short format implicit dependency)")
	}
}

func TestParser_BuildImplicitDependencies_LongFormat(t *testing.T) {
	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "step1",
				Type: StepTypeAPICall,
			},
			{
				ID:   "step2",
				Type: StepTypeAPICall,
				APICall: &APICallStep{
					// Long format: {steps.step_id.output}
					Endpoint: "/api/resource/{steps.step1.id}",
				},
			},
		},
	}

	parser := NewParser(workflow)
	_, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	node2 := parser.dag.Nodes["step2"]
	foundDep := false
	for _, dep := range node2.Dependencies {
		if dep == "step1" {
			foundDep = true
			break
		}
	}

	if !foundDep {
		t.Error("expected step2 to depend on step1 (long format implicit dependency)")
	}
}

func TestParser_BuildImplicitDependencies_NonExistentStep(t *testing.T) {
	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "step1",
				Type: StepTypeAPICall,
				APICall: &APICallStep{
					Endpoint: "/api/{steps.nonexistent.id}",
				},
			},
		},
	}

	parser := NewParser(workflow)
	_, err := parser.Parse()
	if err == nil {
		t.Error("expected error for reference to non-existent step")
	}
	if !contains(err.Error(), "non-existent step") {
		t.Errorf("expected 'non-existent step' error, got: %v", err)
	}
}

func TestParser_BuildImplicitDependencies_SelfReference(t *testing.T) {
	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "step1",
				Type: StepTypeAPICall,
				APICall: &APICallStep{
					// Self-reference should be ignored
					Endpoint: "/api/{step1.id}",
				},
			},
		},
	}

	parser := NewParser(workflow)
	_, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	node1 := parser.dag.Nodes["step1"]
	if len(node1.Dependencies) > 0 {
		t.Error("expected no dependencies for self-reference")
	}
}

func TestParser_BuildImplicitDependencies_DuplicateExplicit(t *testing.T) {
	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "step1",
				Type: StepTypeAPICall,
			},
			{
				ID:        "step2",
				Type:      StepTypeAPICall,
				DependsOn: []string{"step1"}, // Explicit dependency
				APICall: &APICallStep{
					Endpoint: "/api/{steps.step1.id}", // Also implicit
				},
			},
		},
	}

	parser := NewParser(workflow)
	_, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	node2 := parser.dag.Nodes["step2"]

	// Should have only one dependency (no duplicates)
	depCount := 0
	for _, dep := range node2.Dependencies {
		if dep == "step1" {
			depCount++
		}
	}

	if depCount != 1 {
		t.Errorf("expected exactly 1 dependency on step1, got %d", depCount)
	}
}

func TestParser_CreateNodeRecursive_NestedConditional(t *testing.T) {
	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "conditional",
				Type: StepTypeConditional,
				Conditional: &ConditionalStep{
					Condition: "true",
					Then: []*Step{
						{
							ID:   "then-step",
							Type: StepTypeNoop,
						},
					},
					Else: []*Step{
						{
							ID:   "else-step",
							Type: StepTypeNoop,
						},
					},
				},
			},
		},
	}

	parser := NewParser(workflow)
	_, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check all nodes were created
	if parser.dag.Nodes["conditional"] == nil {
		t.Error("expected conditional node to be created")
	}
	if parser.dag.Nodes["then-step"] == nil {
		t.Error("expected then-step node to be created")
	}
	if parser.dag.Nodes["else-step"] == nil {
		t.Error("expected else-step node to be created")
	}
}

func TestParser_CreateNodeRecursive_NestedLoop(t *testing.T) {
	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "loop",
				Type: StepTypeLoop,
				Loop: &LoopStep{
					Iterator:   "item",
					Collection: "items",
					Steps: []*Step{
						{
							ID:   "loop-step",
							Type: StepTypeNoop,
						},
					},
				},
			},
		},
	}

	parser := NewParser(workflow)
	_, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parser.dag.Nodes["loop"] == nil {
		t.Error("expected loop node to be created")
	}
	if parser.dag.Nodes["loop-step"] == nil {
		t.Error("expected loop-step node to be created")
	}
}

func TestParser_CreateNodeRecursive_NestedParallel(t *testing.T) {
	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "parallel",
				Type: StepTypeParallel,
				Parallel: &ParallelStep{
					Steps: []*Step{
						{
							ID:   "parallel-1",
							Type: StepTypeNoop,
						},
						{
							ID:   "parallel-2",
							Type: StepTypeNoop,
						},
					},
				},
			},
		},
	}

	parser := NewParser(workflow)
	_, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if parser.dag.Nodes["parallel"] == nil {
		t.Error("expected parallel node to be created")
	}
	if parser.dag.Nodes["parallel-1"] == nil {
		t.Error("expected parallel-1 node to be created")
	}
	if parser.dag.Nodes["parallel-2"] == nil {
		t.Error("expected parallel-2 node to be created")
	}
}

func TestParser_CreateNodeRecursive_MissingStepID(t *testing.T) {
	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "", // Missing ID
				Type: StepTypeNoop,
			},
		},
	}

	parser := NewParser(workflow)
	_, err := parser.Parse()
	if err == nil {
		t.Error("expected error for missing step ID")
	}
	if !contains(err.Error(), "step ID is required") {
		t.Errorf("expected 'step ID is required' error, got: %v", err)
	}
}

func TestParser_CalculateLevels_ComplexDependencies(t *testing.T) {
	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "a",
				Type: StepTypeNoop,
			},
			{
				ID:        "b",
				Type:      StepTypeNoop,
				DependsOn: []string{"a"},
			},
			{
				ID:        "c",
				Type:      StepTypeNoop,
				DependsOn: []string{"a"},
			},
			{
				ID:        "d",
				Type:      StepTypeNoop,
				DependsOn: []string{"b"},
			},
			{
				ID:        "e",
				Type:      StepTypeNoop,
				DependsOn: []string{"b", "c"},
			},
			{
				ID:        "f",
				Type:      StepTypeNoop,
				DependsOn: []string{"d", "e"},
			},
		},
	}

	parser := NewParser(workflow)
	_, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify levels
	expectedLevels := map[string]int{
		"a": 0,
		"b": 1,
		"c": 1,
		"d": 2,
		"e": 2,
		"f": 3,
	}

	for stepID, expectedLevel := range expectedLevels {
		node := parser.dag.Nodes[stepID]
		if node.Level != expectedLevel {
			t.Errorf("step %s: expected level %d, got %d", stepID, expectedLevel, node.Level)
		}
	}
}

func TestParser_GetExecutionOrder_EmptyDAG(t *testing.T) {
	parser := &Parser{
		workflow: &Workflow{},
		dag:      nil,
	}

	order := parser.GetExecutionOrder()
	if order != nil {
		t.Error("expected nil order for nil DAG")
	}
}

func TestParser_CollectReferences_NoReferences(t *testing.T) {
	parser := NewParser(&Workflow{})

	step := &Step{
		ID:   "simple",
		Type: StepTypeNoop,
	}

	refs := parser.collectReferences(step)
	if len(refs) != 0 {
		t.Errorf("expected 0 references for simple step, got %d", len(refs))
	}
}
