package workflow

import (
	"testing"
)

func TestParser_Parse(t *testing.T) {
	tests := []struct {
		name        string
		workflow    *Workflow
		wantErr     bool
		errContains string
	}{
		{
			name: "simple workflow",
			workflow: &Workflow{
				Steps: []*Step{
					{
						ID:   "step1",
						Type: StepTypeAPICall,
						APICall: &APICallStep{
							Endpoint: "/api/test",
							Method:   "GET",
						},
					},
					{
						ID:        "step2",
						Type:      StepTypeAPICall,
						DependsOn: []string{"step1"},
						APICall: &APICallStep{
							Endpoint: "/api/test2",
							Method:   "GET",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "duplicate step ID",
			workflow: &Workflow{
				Steps: []*Step{
					{
						ID:   "step1",
						Type: StepTypeAPICall,
					},
					{
						ID:   "step1",
						Type: StepTypeAPICall,
					},
				},
			},
			wantErr:     true,
			errContains: "duplicate step ID",
		},
		{
			name: "circular dependency",
			workflow: &Workflow{
				Steps: []*Step{
					{
						ID:        "step1",
						Type:      StepTypeAPICall,
						DependsOn: []string{"step2"},
					},
					{
						ID:        "step2",
						Type:      StepTypeAPICall,
						DependsOn: []string{"step1"},
					},
				},
			},
			wantErr:     true,
			errContains: "circular dependency",
		},
		{
			name: "non-existent dependency",
			workflow: &Workflow{
				Steps: []*Step{
					{
						ID:        "step1",
						Type:      StepTypeAPICall,
						DependsOn: []string{"nonexistent"},
					},
				},
			},
			wantErr:     true,
			errContains: "non-existent step",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parser := NewParser(tt.workflow)
			dag, err := parser.Parse()

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error containing %q, got nil", tt.errContains)
				} else if tt.errContains != "" && !contains(err.Error(), tt.errContains) {
					t.Errorf("expected error containing %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if dag == nil {
					t.Error("expected DAG, got nil")
				}
			}
		})
	}
}

func TestParser_ImplicitDependencies(t *testing.T) {
	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "step1",
				Type: StepTypeAPICall,
				APICall: &APICallStep{
					Endpoint: "/api/test",
					Method:   "GET",
				},
				Output: map[string]string{
					"value": "response.id",
				},
			},
			{
				ID:   "step2",
				Type: StepTypeAPICall,
				APICall: &APICallStep{
					Endpoint: "/api/test/{steps.step1.value}",
					Method:   "GET",
				},
			},
		},
	}

	parser := NewParser(workflow)
	dag, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that step2 depends on step1
	node2 := dag.Nodes["step2"]
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
		t.Error("expected step2 to depend on step1 (implicit dependency)")
	}
}

func TestParser_GetExecutionOrder(t *testing.T) {
	workflow := &Workflow{
		Steps: []*Step{
			{
				ID:   "a",
				Type: StepTypeAPICall,
			},
			{
				ID:        "b",
				Type:      StepTypeAPICall,
				DependsOn: []string{"a"},
			},
			{
				ID:        "c",
				Type:      StepTypeAPICall,
				DependsOn: []string{"a"},
			},
			{
				ID:        "d",
				Type:      StepTypeAPICall,
				DependsOn: []string{"b", "c"},
			},
		},
	}

	parser := NewParser(workflow)
	_, err := parser.Parse()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	order := parser.GetExecutionOrder()

	// Expected execution order:
	// Level 0: a
	// Level 1: b, c (can run in parallel)
	// Level 2: d

	if len(order) != 3 {
		t.Errorf("expected 3 levels, got %d", len(order))
	}

	// Check level 0
	if len(order[0]) != 1 || order[0][0].ID != "a" {
		t.Errorf("expected level 0 to contain only 'a'")
	}

	// Check level 1
	if len(order[1]) != 2 {
		t.Errorf("expected level 1 to contain 2 steps, got %d", len(order[1]))
	}

	// Check level 2
	if len(order[2]) != 1 || order[2][0].ID != "d" {
		t.Errorf("expected level 2 to contain only 'd'")
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) &&
		(s[:len(substr)] == substr || s[len(s)-len(substr):] == substr ||
			len(s) > len(substr)+1 && containsMiddle(s, substr)))
}

func containsMiddle(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
