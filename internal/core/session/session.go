// Package session provides state management for d3 sessions
package session

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"gopkg.in/yaml.v3"
)

// Phase represents a development phase
type Phase string

const (
	// None represents no active phase
	None Phase = ""
	// Define represents the define phase (exploration and brainstorming)
	Define Phase = "define"
	// Design represents the design phase (requirements and specifications)
	Design Phase = "design"
	// Deliver represents the deliver phase (coding and development)
	Deliver Phase = "deliver"
)

// Valid checks if a phase is valid
func (p Phase) Valid() bool {
	switch p {
	case None, Define, Design, Deliver:
		return true
	default:
		return false
	}
}

// Next returns the next phase in the standard progression
func (p Phase) Next() Phase {
	switch p {
	case None:
		return Define
	case Define:
		return Design
	case Design:
		return Deliver
	default:
		return p // Deliver has no next phase
	}
}

// String returns the string representation of the phase
func (p Phase) String() string {
	if p == "" {
		return "none"
	}
	return string(p)
}

// ParsePhase converts a string to a Phase
func ParsePhase(s string) (Phase, error) {
	switch s {
	case "", "none":
		return None, nil
	case "define":
		return Define, nil
	case "design":
		return Design, nil
	case "deliver":
		return Deliver, nil
	default:
		return None, fmt.Errorf("invalid phase: %s", s)
	}
}

// SessionState contains the simplified d3 session state
type SessionState struct {
	// Current context
	CurrentFeature string `yaml:"current_feature,omitempty"`
	CurrentPhase   Phase  `yaml:"current_phase,omitempty"`

	// Meta information
	LastModified time.Time `yaml:"last_modified,omitempty"`
	Version      string    `yaml:"version,omitempty"`
}

// Manager handles session state operations
type Manager struct {
	workspaceRoot string
	sessionFile   string
}

// NewManager creates a new session manager
func NewManager(workspaceRoot string) *Manager {
	d3Dir := filepath.Join(workspaceRoot, ".d3")
	return &Manager{
		workspaceRoot: workspaceRoot,
		sessionFile:   filepath.Join(d3Dir, "session.yaml"),
	}
}

// Load loads the current session state
func (m *Manager) Load() (*SessionState, error) {
	// If session file doesn't exist, create a new empty session
	if _, err := os.Stat(m.sessionFile); os.IsNotExist(err) {
		return &SessionState{
			Version:      "1.0",
			LastModified: time.Now(),
		}, nil
	}

	// Read session file
	data, err := os.ReadFile(m.sessionFile)
	if err != nil {
		return nil, fmt.Errorf("failed to read session file: %w", err)
	}

	// Parse session file
	var state SessionState
	if err := yaml.Unmarshal(data, &state); err != nil {
		return nil, fmt.Errorf("failed to parse session file: %w", err)
	}

	return &state, nil
}

// Save saves the current session state
func (m *Manager) Save(state *SessionState) error {
	// Update last modified
	state.LastModified = time.Now()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(m.sessionFile), 0755); err != nil {
		return fmt.Errorf("failed to create session file directory: %w", err)
	}

	// Marshal state to YAML
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal session state: %w", err)
	}

	// Write session file
	if err := os.WriteFile(m.sessionFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}

// SetPhase sets the current phase
func (m *Manager) SetPhase(phase Phase) error {
	// Load current state
	state, err := m.Load()
	if err != nil {
		return fmt.Errorf("failed to load session state: %w", err)
	}

	// Check if we have an active feature
	if state.CurrentFeature == "" {
		return fmt.Errorf("no active feature, use feature commands first")
	}

	// Validate phase
	if !phase.Valid() {
		return fmt.Errorf("invalid phase: %s", phase)
	}

	// Update state with new phase
	state.CurrentPhase = phase

	// Save updated state
	if err := m.Save(state); err != nil {
		return fmt.Errorf("failed to save session state: %w", err)
	}

	return nil
}

// GetCurrentPhase returns the current phase
func (m *Manager) GetCurrentPhase() (Phase, error) {
	state, err := m.Load()
	if err != nil {
		return None, fmt.Errorf("failed to load session state: %w", err)
	}
	return state.CurrentPhase, nil
}

// GetCurrentFeature returns the current feature
func (m *Manager) GetCurrentFeature() (string, error) {
	state, err := m.Load()
	if err != nil {
		return "", fmt.Errorf("failed to load session state: %w", err)
	}
	return state.CurrentFeature, nil
}

// GetContext returns the current session context (feature and phase)
func (m *Manager) GetContext() (string, Phase, error) {
	state, err := m.Load()
	if err != nil {
		return "", None, fmt.Errorf("failed to load session state: %w", err)
	}
	return state.CurrentFeature, state.CurrentPhase, nil
}
