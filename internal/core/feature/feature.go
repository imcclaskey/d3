// Package feature implements core feature operations
package feature

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imcclaskey/d3/internal/core/session"
)

// Service provides feature management operations
type Service struct {
	workspaceRoot string
	featuresDir   string
	d3Dir         string
	sessionMgr    *session.Manager
}

// NewService creates a new feature service
func NewService(workspaceRoot, featuresDir string) *Service {
	return &Service{
		workspaceRoot: workspaceRoot,
		featuresDir:   featuresDir,
		d3Dir:         filepath.Join(workspaceRoot, ".d3"),
		sessionMgr:    session.NewManager(workspaceRoot),
	}
}

// ContextProvider defines the interface for accessing the workspace context
type ContextProvider interface {
	GetContext() (*Context, error)
	UpdateContext(feature, phase string) error
}

// Context represents the current d3 context
type Context struct {
	Feature string
	Phase   string
}

// CreateResult contains the result of a create operation
type CreateResult struct {
	FeatureName string
	FeaturePath string
	Message     string
}

// Create creates a new feature and sets it as the current context
func (s *Service) Create(ctx context.Context, featureName string) (*CreateResult, error) {
	featurePath := filepath.Join(s.featuresDir, featureName)

	// Check if feature already exists
	if _, err := os.Stat(featurePath); err == nil {
		return nil, fmt.Errorf("feature %s already exists", featureName)
	}

	// Create feature directory
	if err := os.MkdirAll(featurePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create feature directory: %w", err)
	}

	// Load current session state
	state, err := s.sessionMgr.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load session state: %w", err)
	}

	// Update the session state
	state.CurrentFeature = featureName
	state.CurrentPhase = session.Define

	// Save the updated state
	if err := s.sessionMgr.Save(state); err != nil {
		return nil, fmt.Errorf("failed to update session state: %w", err)
	}

	message := fmt.Sprintf("Created feature %s and set as current context", featureName)
	return &CreateResult{
		FeatureName: featureName,
		FeaturePath: featurePath,
		Message:     message,
	}, nil
}

// EnterResult contains the result of an enter operation
type EnterResult struct {
	FeatureName string
	Message     string
}

// Enter sets a feature as the current context
func (s *Service) Enter(ctx context.Context, featureName string) (*EnterResult, error) {
	featurePath := filepath.Join(s.featuresDir, featureName)

	// Check if feature exists
	if _, err := os.Stat(featurePath); err != nil {
		return nil, fmt.Errorf("feature %s does not exist", featureName)
	}

	// Load current session state
	state, err := s.sessionMgr.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load session state: %w", err)
	}

	// Update the session state
	state.CurrentFeature = featureName
	state.CurrentPhase = session.Define

	// Save the updated state
	if err := s.sessionMgr.Save(state); err != nil {
		return nil, fmt.Errorf("failed to update session state: %w", err)
	}

	message := fmt.Sprintf("Entered feature %s", featureName)
	return &EnterResult{
		FeatureName: featureName,
		Message:     message,
	}, nil
}

// LeaveResult contains the result of a leave operation
type LeaveResult struct {
	Message string
}

// Leave clears the current feature context
func (s *Service) Leave(ctx context.Context) (*LeaveResult, error) {
	// Load current session state
	state, err := s.sessionMgr.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load session state: %w", err)
	}

	// Update the session state
	state.CurrentFeature = ""
	state.CurrentPhase = session.None

	// Save the updated state
	if err := s.sessionMgr.Save(state); err != nil {
		return nil, fmt.Errorf("failed to update session state: %w", err)
	}

	return &LeaveResult{
		Message: "Left current feature context",
	}, nil
}
