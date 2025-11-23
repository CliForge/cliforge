package workflow

import (
	"testing"
)

func TestExprEvaluator_EvaluateCondition(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{
		"cluster_name": "test-cluster",
		"multi_az":     true,
		"region":       "us-east-1",
	})

	// Add step result
	ctx.SetStepResult("step1", &StepResult{
		StepID:  "step1",
		Success: true,
		Output: map[string]interface{}{
			"status": "ready",
			"count":  5,
		},
	})

	evaluator := NewExprEvaluator(ctx)

	tests := []struct {
		name      string
		condition string
		want      bool
		wantErr   bool
	}{
		{
			name:      "simple boolean",
			condition: "flags.multi_az == true",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "string comparison",
			condition: "flags.region == 'us-east-1'",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "step output access",
			condition: "steps.step1.success == true",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "complex condition",
			condition: "flags.multi_az && steps.step1.status == 'ready'",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "numeric comparison",
			condition: "steps.step1.count > 3",
			want:      true,
			wantErr:   false,
		},
		{
			name:      "empty condition",
			condition: "",
			want:      true,
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluator.EvaluateCondition(tt.condition)
			if (err != nil) != tt.wantErr {
				t.Errorf("EvaluateCondition() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("EvaluateCondition() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExprEvaluator_InterpolateString(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{
		"cluster_name": "test-cluster",
		"region":       "us-east-1",
	})

	ctx.SetStepResult("create-cluster", &StepResult{
		StepID:  "create-cluster",
		Success: true,
		Output: map[string]interface{}{
			"cluster_id": "abc123",
		},
	})

	evaluator := NewExprEvaluator(ctx)

	tests := []struct {
		name     string
		template string
		want     string
		wantErr  bool
	}{
		{
			name:     "simple interpolation",
			template: "cluster-{flags.cluster_name}",
			want:     "cluster-test-cluster",
			wantErr:  false,
		},
		{
			name:     "multiple interpolations",
			template: "{flags.cluster_name} in {flags.region}",
			want:     "test-cluster in us-east-1",
			wantErr:  false,
		},
		{
			name:     "step output interpolation",
			template: "/api/clusters/{steps['create-cluster'].cluster_id}",
			want:     "/api/clusters/abc123",
			wantErr:  false,
		},
		{
			name:     "no interpolation",
			template: "static-string",
			want:     "static-string",
			wantErr:  false,
		},
		{
			name:     "empty string",
			template: "",
			want:     "",
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := evaluator.InterpolateString(tt.template)
			if (err != nil) != tt.wantErr {
				t.Errorf("InterpolateString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("InterpolateString() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestExprEvaluator_InterpolateMap(t *testing.T) {
	ctx := NewExecutionContext(map[string]interface{}{
		"cluster_name": "test-cluster",
	})

	evaluator := NewExprEvaluator(ctx)

	input := map[string]interface{}{
		"name":   "{flags.cluster_name}",
		"region": "us-east-1",
		"nested": map[string]interface{}{
			"value": "{flags.cluster_name}-nested",
		},
	}

	result, err := evaluator.InterpolateMap(input)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result["name"] != "test-cluster" {
		t.Errorf("expected name to be 'test-cluster', got %v", result["name"])
	}

	if result["region"] != "us-east-1" {
		t.Errorf("expected region to be 'us-east-1', got %v", result["region"])
	}

	nested, ok := result["nested"].(map[string]interface{})
	if !ok {
		t.Fatal("expected nested to be a map")
	}

	if nested["value"] != "test-cluster-nested" {
		t.Errorf("expected nested value to be 'test-cluster-nested', got %v", nested["value"])
	}
}
