package workflow

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/expr-lang/expr"
)

// ExprEvaluator evaluates expressions in the context of workflow execution.
type ExprEvaluator struct {
	context *ExecutionContext
}

// NewExprEvaluator creates a new expression evaluator.
func NewExprEvaluator(ctx *ExecutionContext) *ExprEvaluator {
	return &ExprEvaluator{
		context: ctx,
	}
}

// EvaluateCondition evaluates a boolean condition expression.
func (e *ExprEvaluator) EvaluateCondition(condition string) (bool, error) {
	if condition == "" {
		return true, nil
	}

	// Prepare environment for expr
	env := e.buildEnvironment()

	// Compile and run the expression
	program, err := expr.Compile(condition, expr.Env(env), expr.AsBool())
	if err != nil {
		return false, fmt.Errorf("failed to compile condition: %w", err)
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return false, fmt.Errorf("failed to evaluate condition: %w", err)
	}

	result, ok := output.(bool)
	if !ok {
		return false, fmt.Errorf("condition did not evaluate to boolean: %v", output)
	}

	return result, nil
}

// EvaluateExpression evaluates an expression and returns the result.
func (e *ExprEvaluator) EvaluateExpression(expression string) (interface{}, error) {
	if expression == "" {
		return nil, nil
	}

	env := e.buildEnvironment()

	program, err := expr.Compile(expression, expr.Env(env))
	if err != nil {
		return nil, fmt.Errorf("failed to compile expression: %w", err)
	}

	output, err := expr.Run(program, env)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %w", err)
	}

	return output, nil
}

// InterpolateString interpolates template strings like "{flags.name}" with actual values.
func (e *ExprEvaluator) InterpolateString(template string) (string, error) {
	if template == "" {
		return "", nil
	}

	// Pattern matches {expression}
	pattern := regexp.MustCompile(`\{([^}]+)\}`)

	result := template
	matches := pattern.FindAllStringSubmatch(template, -1)

	for _, match := range matches {
		if len(match) < 2 {
			continue
		}

		placeholder := match[0] // Full match including braces
		expression := match[1]  // Expression inside braces

		// Evaluate the expression
		value, err := e.EvaluateExpression(expression)
		if err != nil {
			return "", fmt.Errorf("failed to interpolate %s: %w", placeholder, err)
		}

		// Convert value to string
		strValue := fmt.Sprintf("%v", value)

		// Replace in result
		result = strings.ReplaceAll(result, placeholder, strValue)
	}

	return result, nil
}

// InterpolateMap interpolates all string values in a map.
func (e *ExprEvaluator) InterpolateMap(m map[string]interface{}) (map[string]interface{}, error) {
	result := make(map[string]interface{})

	for key, value := range m {
		interpolatedKey, err := e.InterpolateString(key)
		if err != nil {
			return nil, fmt.Errorf("failed to interpolate map key %s: %w", key, err)
		}

		switch v := value.(type) {
		case string:
			interpolatedValue, err := e.InterpolateString(v)
			if err != nil {
				return nil, fmt.Errorf("failed to interpolate map value for key %s: %w", key, err)
			}
			result[interpolatedKey] = interpolatedValue

		case map[string]interface{}:
			interpolatedValue, err := e.InterpolateMap(v)
			if err != nil {
				return nil, err
			}
			result[interpolatedKey] = interpolatedValue

		case []interface{}:
			interpolatedValue, err := e.interpolateSlice(v)
			if err != nil {
				return nil, err
			}
			result[interpolatedKey] = interpolatedValue

		default:
			result[interpolatedKey] = value
		}
	}

	return result, nil
}

// interpolateSlice interpolates all values in a slice.
func (e *ExprEvaluator) interpolateSlice(s []interface{}) ([]interface{}, error) {
	result := make([]interface{}, len(s))

	for i, value := range s {
		switch v := value.(type) {
		case string:
			interpolatedValue, err := e.InterpolateString(v)
			if err != nil {
				return nil, err
			}
			result[i] = interpolatedValue

		case map[string]interface{}:
			interpolatedValue, err := e.InterpolateMap(v)
			if err != nil {
				return nil, err
			}
			result[i] = interpolatedValue

		case []interface{}:
			interpolatedValue, err := e.interpolateSlice(v)
			if err != nil {
				return nil, err
			}
			result[i] = interpolatedValue

		default:
			result[i] = value
		}
	}

	return result, nil
}

// InterpolateStringMap interpolates all values in a string map.
func (e *ExprEvaluator) InterpolateStringMap(m map[string]string) (map[string]string, error) {
	result := make(map[string]string)

	for key, value := range m {
		interpolatedKey, err := e.InterpolateString(key)
		if err != nil {
			return nil, fmt.Errorf("failed to interpolate map key %s: %w", key, err)
		}

		interpolatedValue, err := e.InterpolateString(value)
		if err != nil {
			return nil, fmt.Errorf("failed to interpolate map value for key %s: %w", key, err)
		}

		result[interpolatedKey] = interpolatedValue
	}

	return result, nil
}

// buildEnvironment creates the environment for expression evaluation.
func (e *ExprEvaluator) buildEnvironment() map[string]interface{} {
	env := make(map[string]interface{})

	if e.context == nil {
		return env
	}

	// Add flags
	env["flags"] = e.context.Flags

	// Add step outputs
	stepOutputs := make(map[string]interface{})
	for stepID, result := range e.context.StepResults {
		stepData := make(map[string]interface{})
		stepData["success"] = result.Success
		stepData["error"] = result.Error != nil

		// Add output values
		for key, value := range result.Output {
			stepData[key] = value
		}

		stepOutputs[stepID] = stepData
	}
	env["steps"] = stepOutputs

	// Add context variables
	for key, value := range e.context.Variables {
		env[key] = value
	}

	// Add helper functions
	env["len"] = func(v interface{}) int {
		switch val := v.(type) {
		case string:
			return len(val)
		case []interface{}:
			return len(val)
		case map[string]interface{}:
			return len(val)
		default:
			return 0
		}
	}

	env["has"] = func(m map[string]interface{}, key string) bool {
		_, exists := m[key]
		return exists
	}

	env["startsWith"] = func(s, prefix string) bool {
		return strings.HasPrefix(s, prefix)
	}

	env["endsWith"] = func(s, suffix string) bool {
		return strings.HasSuffix(s, suffix)
	}

	env["contains"] = func(s, substr string) bool {
		return strings.Contains(s, substr)
	}

	return env
}
