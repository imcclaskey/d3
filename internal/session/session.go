package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// Session represents the i3 session state
type Session struct {
	Active   bool                   `json:"active"`
	Feature  string                 `json:"feature"`
	Phase    string                 `json:"phase"`
	Metadata map[string]interface{} `json:"metadata"`
}

// Manager handles operations related to the i3 session
type Manager struct {
	path string
}

// NewManager creates a new session manager
func NewManager(i3Dir string) *Manager {
	// Create directory if it doesn't exist
	_ = os.MkdirAll(i3Dir, 0755)
	
	return &Manager{
		path: filepath.Join(i3Dir, "session.json"),
	}
}

// Get retrieves the current session data
func (m *Manager) Get() (Session, error) {
	data, err := os.ReadFile(m.path)
	
	// Return empty session if file doesn't exist or can't be read
	if err != nil {
		return Session{Metadata: make(map[string]interface{})}, nil
	}
	
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return Session{Metadata: make(map[string]interface{})}, nil
	}
	
	// Ensure metadata is initialized
	if s.Metadata == nil {
		s.Metadata = make(map[string]interface{})
	}
	
	return s, nil
}

// Save persists session data to disk
func (m *Manager) Save(s Session) error {
	// Ensure metadata is initialized
	if s.Metadata == nil {
		s.Metadata = make(map[string]interface{})
	}
	
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshaling session: %w", err)
	}
	
	if err := os.WriteFile(m.path, data, 0644); err != nil {
		return fmt.Errorf("writing session: %w", err)
	}
	
	return nil
}

// Start begins a new session with the given feature
func (m *Manager) Start(feature string) error {
	s, _ := m.Get()
	s.Active = true
	s.Feature = feature
	s.Phase = ""
	return m.Save(s)
}

// SetPhase updates the current phase
func (m *Manager) SetPhase(phase string) error {
	s, _ := m.Get()
	if !s.Active {
		return errors.New("no active session")
	}
	s.Phase = phase
	return m.Save(s)
}

// Stop ends the current session
func (m *Manager) Stop() error {
	s, _ := m.Get()
	s.Active = false
	return m.Save(s)
}

// UpdateMetadata adds or updates metadata entries
func (m *Manager) UpdateMetadata(entries map[string]interface{}) error {
	s, _ := m.Get()
	
	for k, v := range entries {
		s.Metadata[k] = v
	}
	
	return m.Save(s)
}

// Status returns the current session status
// If the session is active, it returns the feature and phase
func (m *Manager) Status() (active bool, feature, phase string, err error) {
	s, err := m.Get()
	if err != nil {
		return false, "", "", err
	}
	
	return s.Active, s.Feature, s.Phase, nil
}

// Feature returns the current feature
func (m *Manager) Feature() (string, error) {
	s, err := m.Get()
	if err != nil {
		return "", err
	}
	return s.Feature, nil
}

// GetCurrentFeature returns the current feature (alias for Feature)
func (m *Manager) GetCurrentFeature() (string, error) {
	return m.Feature()
}

// Phase returns the current phase
func (m *Manager) Phase() (string, error) {
	s, err := m.Get()
	if err != nil {
		return "", err
	}
	if !s.Active {
		return "", errors.New("no active session")
	}
	return s.Phase, nil
}