// Package session provides state management for d3 sessions
package session

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/imcclaskey/d3/internal/core/ports"
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

// NOTE: SessionState struct is removed as it's no longer used for .d3/.session
// If other global, non-feature-specific state needs to be persisted in a committable way,
// a different mechanism or a separate configuration file might be needed.

// Storage handles transient active feature state persistence
type Storage struct {
	sessionFile string
	fs          ports.FileSystem
}

// NewStorage creates a new transient session storage handler
func NewStorage(d3Dir string, fs ports.FileSystem) *Storage {
	return &Storage{
		sessionFile: filepath.Join(d3Dir, ".session"), // Point to .session file
		fs:          fs,
	}
}

// LoadActiveFeature reads the active feature name from the session file.
// Returns an empty string and nil error if the file is empty or does not exist.
func (s *Storage) LoadActiveFeature() (string, error) {
	data, err := s.fs.ReadFile(s.sessionFile)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil // File not existing means no active feature
		}
		return "", fmt.Errorf("failed to read session file %s: %w", s.sessionFile, err)
	}
	// Return the content, trimming whitespace
	return strings.TrimSpace(string(data)), nil
}

// SaveActiveFeature saves the active feature name to the session file.
func (s *Storage) SaveActiveFeature(featureName string) error {
	// Ensure the base directory exists
	if err := s.fs.MkdirAll(filepath.Dir(s.sessionFile), 0755); err != nil {
		return fmt.Errorf("failed to create session file directory: %w", err)
	}

	// Write the feature name as plain text
	data := []byte(featureName)
	if err := s.fs.WriteFile(s.sessionFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write session file %s: %w", s.sessionFile, err)
	}
	return nil
}

// ClearActiveFeature removes the session file, effectively clearing the active feature.
func (s *Storage) ClearActiveFeature() error {
	err := s.fs.Remove(s.sessionFile)
	// Ignore "not exist" error, as it means the state is already cleared
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to remove session file %s: %w", s.sessionFile, err)
	}
	return nil
}
