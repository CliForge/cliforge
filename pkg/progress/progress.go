package progress

import (
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/pterm/pterm"
)

// Spinner implements a spinner progress indicator.
type Spinner struct {
	spinner *pterm.SpinnerPrinter
	config  *Config
	active  bool
	mu      sync.Mutex
}

// NewSpinner creates a new spinner progress indicator.
func NewSpinner(config *Config) *Spinner {
	if config == nil {
		config = DefaultConfig()
	}

	return &Spinner{
		config: config,
	}
}

// Start starts the spinner with a message.
func (s *Spinner) Start(message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.config.Enabled {
		return nil
	}

	if s.active {
		return fmt.Errorf("spinner already active")
	}

	var err error
	s.spinner, err = pterm.DefaultSpinner.Start(message)
	if err != nil {
		return fmt.Errorf("failed to start spinner: %w", err)
	}

	s.active = true
	return nil
}

// Update updates the spinner message.
func (s *Spinner) Update(message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.config.Enabled || !s.active || s.spinner == nil {
		return nil
	}

	s.spinner.UpdateText(message)
	return nil
}

// UpdateWithData updates the spinner with structured data.
func (s *Spinner) UpdateWithData(data *ProgressData) error {
	return s.Update(data.Message)
}

// Success marks the spinner as successful.
func (s *Spinner) Success(message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.config.Enabled || !s.active || s.spinner == nil {
		return nil
	}

	s.spinner.Success(message)
	s.active = false
	return nil
}

// Failure marks the spinner as failed.
func (s *Spinner) Failure(message string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.config.Enabled || !s.active || s.spinner == nil {
		return nil
	}

	s.spinner.Fail(message)
	s.active = false
	return nil
}

// Stop stops the spinner.
func (s *Spinner) Stop() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if !s.active || s.spinner == nil {
		return nil
	}

	s.spinner.Stop()
	s.active = false
	return nil
}

// IsActive returns true if the spinner is active.
func (s *Spinner) IsActive() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.active
}

// ProgressBar implements a progress bar indicator.
type ProgressBar struct {
	bar    *pterm.ProgressbarPrinter
	config *Config
	active bool
	total  int
	mu     sync.Mutex
}

// NewProgressBar creates a new progress bar.
func NewProgressBar(config *Config, total int) *ProgressBar {
	if config == nil {
		config = DefaultConfig()
	}

	return &ProgressBar{
		config: config,
		total:  total,
	}
}

// Start starts the progress bar.
func (p *ProgressBar) Start(message string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.config.Enabled {
		return nil
	}

	if p.active {
		return fmt.Errorf("progress bar already active")
	}

	bar := pterm.DefaultProgressbar.
		WithTotal(p.total).
		WithTitle(message)

	if p.config.Writer != nil && p.config.Writer != os.Stdout {
		bar = bar.WithWriter(p.config.Writer)
	}

	var err error
	p.bar, err = bar.Start()
	if err != nil {
		return fmt.Errorf("failed to start progress bar: %w", err)
	}

	p.active = true
	return nil
}

// Update updates the progress bar.
func (p *ProgressBar) Update(message string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.config.Enabled || !p.active || p.bar == nil {
		return nil
	}

	p.bar.UpdateTitle(message)
	return nil
}

// UpdateWithData updates the progress bar with structured data.
func (p *ProgressBar) UpdateWithData(data *ProgressData) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.config.Enabled || !p.active || p.bar == nil {
		return nil
	}

	if data.Message != "" {
		p.bar.UpdateTitle(data.Message)
	}

	if data.Current > 0 {
		p.bar.Current = data.Current
	}

	return nil
}

// Increment increments the progress bar by 1.
func (p *ProgressBar) Increment() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.config.Enabled || !p.active || p.bar == nil {
		return nil
	}

	p.bar.Increment()
	return nil
}

// Success marks the progress bar as complete.
func (p *ProgressBar) Success(message string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.config.Enabled || !p.active || p.bar == nil {
		return nil
	}

	// Set to 100% complete
	p.bar.Current = p.total
	_, _ = p.bar.Stop()

	if message != "" {
		pterm.Success.Println(message)
	}

	p.active = false
	return nil
}

// Failure marks the progress bar as failed.
func (p *ProgressBar) Failure(message string) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.config.Enabled || !p.active || p.bar == nil {
		return nil
	}

	_, _ = p.bar.Stop()

	if message != "" {
		pterm.Error.Println(message)
	}

	p.active = false
	return nil
}

// Stop stops the progress bar.
func (p *ProgressBar) Stop() error {
	p.mu.Lock()
	defer p.mu.Unlock()

	if !p.active || p.bar == nil {
		return nil
	}

	_, _ = p.bar.Stop()
	p.active = false
	return nil
}

// IsActive returns true if the progress bar is active.
func (p *ProgressBar) IsActive() bool {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.active
}

// MultiStep implements a multi-step tree display.
type MultiStep struct {
	config   *Config
	steps    map[string]*StepInfo
	order    []string
	active   bool
	writer   io.Writer
	mu       sync.Mutex
	rootArea *pterm.AreaPrinter
}

// NewMultiStep creates a new multi-step progress display.
func NewMultiStep(config *Config) *MultiStep {
	if config == nil {
		config = DefaultConfig()
	}

	writer := config.Writer
	if writer == nil {
		writer = os.Stdout
	}

	return &MultiStep{
		config: config,
		steps:  make(map[string]*StepInfo),
		order:  make([]string, 0),
		writer: writer,
	}
}

// Start starts the multi-step display.
func (m *MultiStep) Start(message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.config.Enabled {
		return nil
	}

	if m.active {
		return fmt.Errorf("multi-step display already active")
	}

	if message != "" {
		pterm.Println(message)
	}

	area, err := pterm.DefaultArea.Start()
	if err != nil {
		return fmt.Errorf("failed to start area: %w", err)
	}
	m.rootArea = area
	m.active = true
	return nil
}

// AddStep adds a new step to the display.
func (m *MultiStep) AddStep(step *StepInfo) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.config.Enabled {
		return nil
	}

	m.steps[step.ID] = step
	m.order = append(m.order, step.ID)

	if m.active {
		m.render()
	}

	return nil
}

// UpdateStep updates a step's status.
func (m *MultiStep) UpdateStep(stepID string, status StepStatus, message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.config.Enabled {
		return nil
	}

	step, exists := m.steps[stepID]
	if !exists {
		return fmt.Errorf("step %s not found", stepID)
	}

	step.Status = status
	if message != "" {
		step.Description = message
	}

	if status == StepStatusRunning && step.StartTime.IsZero() {
		step.StartTime = time.Now()
	}

	if status == StepStatusCompleted || status == StepStatusFailed {
		step.EndTime = time.Now()
	}

	if m.active {
		m.render()
	}

	return nil
}

// Update updates the display with a message.
func (m *MultiStep) Update(message string) error {
	// Multi-step doesn't use generic messages
	return nil
}

// UpdateWithData updates with structured data.
func (m *MultiStep) UpdateWithData(data *ProgressData) error {
	// Extract step information from metadata if present
	if data.Metadata != nil {
		if stepID, ok := data.Metadata["step_id"].(string); ok {
			if statusStr, ok := data.Metadata["status"].(string); ok {
				status := StepStatus(statusStr)
				return m.UpdateStep(stepID, status, data.Message)
			}
		}
	}
	return nil
}

// Success marks all steps as complete.
func (m *MultiStep) Success(message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.config.Enabled || !m.active {
		return nil
	}

	if m.active && m.rootArea != nil {
		m.rootArea.Stop()
	}

	if message != "" {
		pterm.Success.Println(message)
	}

	m.active = false
	return nil
}

// Failure marks the display as failed.
func (m *MultiStep) Failure(message string) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.config.Enabled || !m.active {
		return nil
	}

	if m.active && m.rootArea != nil {
		m.rootArea.Stop()
	}

	if message != "" {
		pterm.Error.Println(message)
	}

	m.active = false
	return nil
}

// Stop stops the multi-step display.
func (m *MultiStep) Stop() error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.active {
		return nil
	}

	if m.rootArea != nil {
		m.rootArea.Stop()
	}

	m.active = false
	return nil
}

// IsActive returns true if the display is active.
func (m *MultiStep) IsActive() bool {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.active
}

// render renders the current state of all steps.
func (m *MultiStep) render() {
	if m.rootArea == nil {
		return
	}

	tree := m.buildTree()
	m.rootArea.Update(tree)
}

// buildTree builds a tree representation of the steps.
func (m *MultiStep) buildTree() string {
	var lines []string

	for _, stepID := range m.order {
		step := m.steps[stepID]
		line := m.formatStep(step, 0)
		lines = append(lines, line)

		// Render substeps
		for _, substep := range step.SubSteps {
			line := m.formatStep(substep, 1)
			lines = append(lines, line)
		}
	}

	result := ""
	for _, line := range lines {
		result += line + "\n"
	}
	return result
}

// formatStep formats a single step for display.
func (m *MultiStep) formatStep(step *StepInfo, indent int) string {
	prefix := ""
	for i := 0; i < indent; i++ {
		prefix += "  "
	}

	icon := m.getStatusIcon(step.Status)

	desc := step.Description
	if m.config.ShowTimestamps && !step.StartTime.IsZero() {
		duration := ""
		if !step.EndTime.IsZero() {
			duration = fmt.Sprintf(" (%s)", step.EndTime.Sub(step.StartTime).Round(time.Millisecond))
		}
		desc = fmt.Sprintf("%s%s", desc, duration)
	}

	return fmt.Sprintf("%s%s %s", prefix, icon, desc)
}

// getStatusIcon returns the icon for a step status.
func (m *MultiStep) getStatusIcon(status StepStatus) string {
	switch status {
	case StepStatusCompleted:
		return pterm.Green("✓")
	case StepStatusFailed:
		return pterm.Red("✗")
	case StepStatusRunning:
		return pterm.Yellow("⧗")
	case StepStatusSkipped:
		return pterm.Gray("○")
	case StepStatusPending:
		return pterm.Gray("☐")
	default:
		return "?"
	}
}

// NoopProgress is a progress indicator that does nothing.
type NoopProgress struct{}

// NewNoopProgress creates a new no-op progress indicator.
func NewNoopProgress() *NoopProgress {
	return &NoopProgress{}
}

// Start does nothing.
func (n *NoopProgress) Start(message string) error { return nil }

// Update does nothing.
func (n *NoopProgress) Update(message string) error { return nil }

// UpdateWithData does nothing.
func (n *NoopProgress) UpdateWithData(data *ProgressData) error { return nil }

// Success does nothing.
func (n *NoopProgress) Success(message string) error { return nil }

// Failure does nothing.
func (n *NoopProgress) Failure(message string) error { return nil }

// Stop does nothing.
func (n *NoopProgress) Stop() error { return nil }

// IsActive always returns false.
func (n *NoopProgress) IsActive() bool { return false }

// New creates a new progress indicator based on the config.
func New(config *Config, total int) Progress {
	if config == nil {
		config = DefaultConfig()
	}

	if !config.Enabled {
		return NewNoopProgress()
	}

	switch config.Type {
	case ProgressTypeSpinner:
		return NewSpinner(config)
	case ProgressTypeBar:
		return NewProgressBar(config, total)
	case ProgressTypeSteps:
		return NewMultiStep(config)
	case ProgressTypeNone:
		return NewNoopProgress()
	default:
		return NewSpinner(config)
	}
}
