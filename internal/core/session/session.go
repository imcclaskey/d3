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
	return string(p)
}

// ParsePhase converts a string to a Phase
func ParsePhase(s string) (Phase, error) {
	switch s {
	case "":
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

// Storage handles session state persistence
type Storage struct {
	sessionFile string
}

// NewStorage creates a new session storage handler
func NewStorage(d3Dir string) *Storage {
	return &Storage{
		sessionFile: filepath.Join(d3Dir, "session.yaml"),
	}
}

// Load loads the current session state from disk
func (s *Storage) Load() (*SessionState, error) {
	// If session file doesn't exist, error with the targeted path
	if _, err := os.Stat(s.sessionFile); os.IsNotExist(err) {
		return nil, fmt.Errorf("session file does not exist: %s", s.sessionFile)
	}

	// Read session file
	data, err := os.ReadFile(s.sessionFile)
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

// Save saves the current session state to disk
func (s *Storage) Save(state *SessionState) error {
	// Update last modified
	state.LastModified = time.Now()

	// Create directory if it doesn't exist
	if err := os.MkdirAll(filepath.Dir(s.sessionFile), 0755); err != nil {
		return fmt.Errorf("failed to create session file directory: %w", err)
	}

	// Marshal state to YAML
	data, err := yaml.Marshal(state)
	if err != nil {
		return fmt.Errorf("failed to marshal session state: %w", err)
	}

	// Write session file
	if err := os.WriteFile(s.sessionFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file: %w", err)
	}

	return nil
}
