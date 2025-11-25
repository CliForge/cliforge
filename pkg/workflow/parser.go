package workflow

import (
	"fmt"
	"regexp"
)

// Parser parses workflow definitions and builds execution graphs.
type Parser struct {
	workflow *Workflow
	dag      *DAG
}

// NewParser creates a new workflow parser.
func NewParser(workflow *Workflow) *Parser {
	return &Parser{
		workflow: workflow,
		dag: &DAG{
			Nodes: make(map[string]*DAGNode),
			Edges: make(map[string][]string),
		},
	}
}

// Parse parses the workflow and builds the DAG.
func (p *Parser) Parse() (*DAG, error) {
	// First pass: create nodes
	if err := p.createNodes(); err != nil {
		return nil, err
	}

	// Second pass: build explicit dependencies
	if err := p.buildExplicitDependencies(); err != nil {
		return nil, err
	}

	// Third pass: detect implicit dependencies from output references
	if err := p.buildImplicitDependencies(); err != nil {
		return nil, err
	}

	// Validate: detect cycles
	if err := p.detectCycles(); err != nil {
		return nil, err
	}

	// Calculate node levels for topological ordering
	if err := p.calculateLevels(); err != nil {
		return nil, err
	}

	return p.dag, nil
}

// createNodes creates DAG nodes for all steps.
func (p *Parser) createNodes() error {
	for _, step := range p.workflow.Steps {
		if err := p.createNodeRecursive(step); err != nil {
			return err
		}
	}
	return nil
}

// createNodeRecursive creates nodes recursively for nested steps.
func (p *Parser) createNodeRecursive(step *Step) error {
	if step.ID == "" {
		return fmt.Errorf("step ID is required")
	}

	if _, exists := p.dag.Nodes[step.ID]; exists {
		return fmt.Errorf("duplicate step ID: %s", step.ID)
	}

	node := &DAGNode{
		Step:         step,
		Dependencies: make([]string, 0),
		Dependents:   make([]string, 0),
	}

	p.dag.Nodes[step.ID] = node

	// Process nested steps
	switch step.Type {
	case StepTypeConditional:
		if step.Conditional != nil {
			for _, thenStep := range step.Conditional.Then {
				if err := p.createNodeRecursive(thenStep); err != nil {
					return err
				}
			}
			for _, elseStep := range step.Conditional.Else {
				if err := p.createNodeRecursive(elseStep); err != nil {
					return err
				}
			}
		}
	case StepTypeLoop:
		if step.Loop != nil {
			for _, loopStep := range step.Loop.Steps {
				if err := p.createNodeRecursive(loopStep); err != nil {
					return err
				}
			}
		}
	case StepTypeParallel:
		if step.Parallel != nil {
			for _, parallelStep := range step.Parallel.Steps {
				if err := p.createNodeRecursive(parallelStep); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// buildExplicitDependencies builds dependencies from depends-on declarations.
func (p *Parser) buildExplicitDependencies() error {
	for stepID, node := range p.dag.Nodes {
		for _, depID := range node.Step.DependsOn {
			if _, exists := p.dag.Nodes[depID]; !exists {
				return fmt.Errorf("step %s depends on non-existent step %s", stepID, depID)
			}

			// Add dependency
			node.Dependencies = append(node.Dependencies, depID)
			p.dag.Nodes[depID].Dependents = append(p.dag.Nodes[depID].Dependents, stepID)

			// Add edge
			if p.dag.Edges[depID] == nil {
				p.dag.Edges[depID] = make([]string, 0)
			}
			p.dag.Edges[depID] = append(p.dag.Edges[depID], stepID)
		}
	}
	return nil
}

// buildImplicitDependencies detects dependencies from output references.
func (p *Parser) buildImplicitDependencies() error {
	// Regular expression to find step output references: {step_id.output_name} or {steps.step_id.output_name}
	refPattern := regexp.MustCompile(`\{(?:steps\.)?([a-zA-Z0-9_-]+)\.[a-zA-Z0-9_.-]+\}`)

	for stepID, node := range p.dag.Nodes {
		// Collect all string fields that might contain references
		refs := p.collectReferences(node.Step)

		// Extract referenced step IDs
		referencedSteps := make(map[string]bool)
		for _, ref := range refs {
			matches := refPattern.FindAllStringSubmatch(ref, -1)
			for _, match := range matches {
				if len(match) > 1 {
					refStepID := match[1]
					if refStepID != stepID { // Don't self-reference
						referencedSteps[refStepID] = true
					}
				}
			}
		}

		// Add implicit dependencies
		for refStepID := range referencedSteps {
			if _, exists := p.dag.Nodes[refStepID]; !exists {
				return fmt.Errorf("step %s references non-existent step %s", stepID, refStepID)
			}

			// Check if dependency already exists
			alreadyDepends := false
			for _, dep := range node.Dependencies {
				if dep == refStepID {
					alreadyDepends = true
					break
				}
			}

			if !alreadyDepends {
				node.Dependencies = append(node.Dependencies, refStepID)
				p.dag.Nodes[refStepID].Dependents = append(p.dag.Nodes[refStepID].Dependents, stepID)

				if p.dag.Edges[refStepID] == nil {
					p.dag.Edges[refStepID] = make([]string, 0)
				}
				p.dag.Edges[refStepID] = append(p.dag.Edges[refStepID], stepID)
			}
		}
	}

	return nil
}

// collectReferences collects all string values from a step that might contain references.
func (p *Parser) collectReferences(step *Step) []string {
	refs := make([]string, 0)

	// Common fields
	if step.Condition != "" {
		refs = append(refs, step.Condition)
	}

	// Output mappings
	for _, expr := range step.Output {
		refs = append(refs, expr)
	}

	// Type-specific fields
	switch step.Type {
	case StepTypeAPICall:
		if step.APICall != nil {
			refs = append(refs, step.APICall.Endpoint)
			for _, v := range step.APICall.Headers {
				refs = append(refs, v)
			}
			for _, v := range step.APICall.Query {
				refs = append(refs, v)
			}
			refs = append(refs, p.collectMapValues(step.APICall.Body)...)
		}
	case StepTypePlugin:
		if step.Plugin != nil {
			refs = append(refs, step.Plugin.Plugin)
			refs = append(refs, step.Plugin.Command)
			refs = append(refs, p.collectMapValues(step.Plugin.Input)...)
		}
	case StepTypeConditional:
		if step.Conditional != nil {
			refs = append(refs, step.Conditional.Condition)
		}
	case StepTypeLoop:
		if step.Loop != nil {
			refs = append(refs, step.Loop.Collection)
		}
	case StepTypeWait:
		if step.Wait != nil {
			if step.Wait.Condition != "" {
				refs = append(refs, step.Wait.Condition)
			}
			if step.Wait.Polling != nil {
				refs = append(refs, step.Wait.Polling.Endpoint)
			}
		}
	}

	return refs
}

// collectMapValues recursively collects string values from a map.
func (p *Parser) collectMapValues(m map[string]interface{}) []string {
	values := make([]string, 0)
	for _, v := range m {
		switch val := v.(type) {
		case string:
			values = append(values, val)
		case map[string]interface{}:
			values = append(values, p.collectMapValues(val)...)
		case []interface{}:
			for _, item := range val {
				if strVal, ok := item.(string); ok {
					values = append(values, strVal)
				} else if mapVal, ok := item.(map[string]interface{}); ok {
					values = append(values, p.collectMapValues(mapVal)...)
				}
			}
		}
	}
	return values
}

// detectCycles detects circular dependencies using DFS.
func (p *Parser) detectCycles() error {
	visited := make(map[string]bool)
	recStack := make(map[string]bool)

	for stepID := range p.dag.Nodes {
		if !visited[stepID] {
			if p.hasCycle(stepID, visited, recStack) {
				return fmt.Errorf("circular dependency detected involving step: %s", stepID)
			}
		}
	}

	return nil
}

// hasCycle performs DFS to detect cycles.
func (p *Parser) hasCycle(stepID string, visited, recStack map[string]bool) bool {
	visited[stepID] = true
	recStack[stepID] = true

	for _, dependent := range p.dag.Edges[stepID] {
		if !visited[dependent] {
			if p.hasCycle(dependent, visited, recStack) {
				return true
			}
		} else if recStack[dependent] {
			return true
		}
	}

	recStack[stepID] = false
	return false
}

// calculateLevels calculates the depth level of each node for topological ordering.
func (p *Parser) calculateLevels() error {
	// Initialize all levels to -1 (unset)
	for stepID, node := range p.dag.Nodes {
		node.Level = -1
		p.dag.Nodes[stepID] = node
	}

	// Calculate levels using topological approach
	// Nodes with no dependencies are level 0
	// Each node's level is 1 + max level of its dependencies
	visited := make(map[string]bool)

	var calculateLevel func(string) (int, error)
	calculateLevel = func(stepID string) (int, error) {
		if visited[stepID] {
			// Already calculated or in progress (cycle detection)
			return p.dag.Nodes[stepID].Level, nil
		}

		visited[stepID] = true
		node := p.dag.Nodes[stepID]

		// If no dependencies, level is 0
		if len(node.Dependencies) == 0 {
			node.Level = 0
			p.dag.Nodes[stepID] = node
			return 0, nil
		}

		// Calculate level based on dependencies
		maxDepLevel := -1
		for _, depID := range node.Dependencies {
			depLevel, err := calculateLevel(depID)
			if err != nil {
				return -1, err
			}
			if depLevel > maxDepLevel {
				maxDepLevel = depLevel
			}
		}

		node.Level = maxDepLevel + 1
		p.dag.Nodes[stepID] = node
		return node.Level, nil
	}

	// Calculate levels for all nodes
	for stepID := range p.dag.Nodes {
		if _, err := calculateLevel(stepID); err != nil {
			return err
		}
	}

	return nil
}

// GetExecutionOrder returns steps in the order they should be executed.
// Steps at the same level can be executed in parallel.
func (p *Parser) GetExecutionOrder() [][]*Step {
	if p.dag == nil {
		return nil
	}

	// Group steps by level
	levelMap := make(map[int][]*Step)
	maxLevel := 0

	for _, node := range p.dag.Nodes {
		level := node.Level
		if level > maxLevel {
			maxLevel = level
		}
		levelMap[level] = append(levelMap[level], node.Step)
	}

	// Convert to ordered slice
	order := make([][]*Step, maxLevel+1)
	for level := 0; level <= maxLevel; level++ {
		order[level] = levelMap[level]
	}

	return order
}
