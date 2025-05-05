// Package project provides project management functionality
package project

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imcclaskey/d3/internal/core/feature"
	"github.com/imcclaskey/d3/internal/core/rules"
	"github.com/imcclaskey/d3/internal/core/session"
)

// Common error definitions
var (
	ErrNotInitialized  = errors.New("project not initialized, please run 'd3 init' first")
	ErrNoActiveFeature = errors.New("no active feature")
)

// Result represents a simplified response from project operations
// that can be used by both CLI and MCP interfaces
type Result struct {
	// Message is a simple human-readable message (1-2 sentences max)
	Message string

	// RulesChanged indicates whether rule files were updated during the operation
	RulesChanged bool
}

// NewResult creates a new result with the given message
func NewResult(message string) *Result {
	return &Result{
		Message:      message,
		RulesChanged: false,
	}
}

// NewResultWithRulesChanged creates a new result with rules changed set to true
func NewResultWithRulesChanged(message string) *Result {
	return &Result{
		Message:      message,
		RulesChanged: true,
	}
}

// FormatCLI formats the result for CLI output
func (r *Result) FormatCLI() string {
	if r.RulesChanged {
		return fmt.Sprintf("%s Cursor rules have been updated.", r.Message)
	}

	return r.Message
}

// FormatMCP formats the result for MCP tool output
func (r *Result) FormatMCP() string {
	if r.RulesChanged {
		return fmt.Sprintf("%s Cursor rules have changed. Stop your current behavior and await further instruction.", r.Message)
	}

	return r.Message
}

// State manages the shared state of the project
type State struct {
	// Path configuration
	ProjectRoot    string
	D3Dir          string
	FeaturesDir    string
	CursorRulesDir string

	// Active context
	CurrentFeature string
	CurrentPhase   session.Phase

	// Hooks
	OnStateChanged func()
}

// Project coordinates all d3 services
type Project struct {
	state         *State
	features      *feature.Service
	session       *session.Storage
	rules         *rules.Service
	isInitialized bool // Tracks whether the project has been initialized
}

// New creates a new project instance from project root
func New(projectRoot string) *Project {
	state := newState(projectRoot)

	// Create services
	sessionStorage := session.NewStorage(state.D3Dir)
	rulesService := rules.NewService(state.ProjectRoot, state.CursorRulesDir)
	featuresService := feature.NewService(state.ProjectRoot, state.FeaturesDir, state.D3Dir)

	// Create project
	return &Project{
		state:         state,
		session:       sessionStorage,
		rules:         rulesService,
		features:      featuresService,
		isInitialized: false, // Will be set to true after checking or initializing
	}
}

// newState creates a new project state from project root
func newState(projectRoot string) *State {
	d3Dir := filepath.Join(projectRoot, ".d3")
	featuresDir := filepath.Join(d3Dir, "features")
	cursorRulesDir := filepath.Join(projectRoot, ".cursor", "rules")

	return &State{
		ProjectRoot:    projectRoot,
		D3Dir:          d3Dir,
		FeaturesDir:    featuresDir,
		CursorRulesDir: cursorRulesDir,
	}
}

// IsInitialized checks if the project is initialized
func (p *Project) IsInitialized() bool {
	// Check if .d3 directory exists
	_, err := os.Stat(p.state.D3Dir)
	return err == nil
}

// RequiresInitialized ensures the project is initialized before performing operations
func (p *Project) RequiresInitialized() error {
	if !p.IsInitialized() {
		return ErrNotInitialized
	}
	return nil
}

// Init initializes the project
func (p *Project) Init(clean bool) (*Result, error) {
	// If already initialized and not clean, return success
	if p.IsInitialized() && !clean {
		return NewResult("Project already initialized."), nil
	}

	// If clean, remove the d3 directory
	if clean && p.IsInitialized() {
		if err := os.RemoveAll(p.state.D3Dir); err != nil {
			return nil, fmt.Errorf("failed to clean existing d3 directory: %w", err)
		}
	}

	// Create directories
	directories := []string{
		p.state.D3Dir,
		p.state.FeaturesDir,
	}

	for _, dir := range directories {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Initialize session
	state := &session.SessionState{
		Version: "1.0",
	}
	if err := p.session.Save(state); err != nil {
		return nil, fmt.Errorf("failed to initialize session: %w", err)
	}

	// Initialize rules
	if err := p.rules.RefreshRules("", ""); err != nil {
		return nil, fmt.Errorf("failed to initialize rules: %w", err)
	}

	// Mark project as initialized
	p.isInitialized = true

	return NewResultWithRulesChanged("Project initialized successfully."), nil
}

// CreateFeature creates a new feature and sets it as the current feature
func (p *Project) CreateFeature(ctx context.Context, featureName string) (*Result, error) {
	// Check if project is initialized
	if err := p.RequiresInitialized(); err != nil {
		return nil, err
	}

	// Create the feature
	_, err := p.features.CreateFeature(ctx, featureName)
	if err != nil {
		return nil, fmt.Errorf("failed to create feature: %w", err)
	}

	// Load current session state
	sessionState, err := p.session.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	// Update the current feature
	sessionState.CurrentFeature = featureName

	// Reset phase when changing feature
	sessionState.CurrentPhase = session.None

	// Save the session state
	if err := p.session.Save(sessionState); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Update the rules with the new context
	if err := p.rules.RefreshRules(featureName, sessionState.CurrentPhase.String()); err != nil {
		return nil, fmt.Errorf("failed to refresh rules: %w", err)
	}

	// Update state
	p.state.CurrentFeature = featureName
	p.state.CurrentPhase = sessionState.CurrentPhase

	// Call state changed hook if available
	if p.state.OnStateChanged != nil {
		p.state.OnStateChanged()
	}

	return NewResultWithRulesChanged(fmt.Sprintf("Feature '%s' created and set as the current context.", featureName)), nil
}

// ChangePhase changes the current phase of the active feature
func (p *Project) ChangePhase(ctx context.Context, targetPhase session.Phase) (*Result, error) {
	// Check if project is initialized
	if err := p.RequiresInitialized(); err != nil {
		return nil, err
	}

	// Load current session state
	sessionState, err := p.session.Load()
	if err != nil {
		return nil, fmt.Errorf("failed to load session: %w", err)
	}

	// Ensure there is an active feature
	if sessionState.CurrentFeature == "" {
		return nil, ErrNoActiveFeature
	}

	// Get current state
	currentFeature := sessionState.CurrentFeature
	currentPhase := sessionState.CurrentPhase

	// Check if we're already in the target phase
	if currentPhase == targetPhase {
		return NewResult(fmt.Sprintf("Already in the %s phase.", targetPhase)), nil
	}

	// Check for impact - if the feature already has files for the target phase
	hasImpact := false
	featureDir := filepath.Join(p.state.FeaturesDir, currentFeature)
	phaseDir := filepath.Join(featureDir, targetPhase.String())
	if _, err := os.Stat(phaseDir); err == nil {
		// Phase directory exists, probably has files
		hasImpact = true
	}

	// Update the session state with the new phase
	sessionState.CurrentPhase = targetPhase
	if err := p.session.Save(sessionState); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Update rules with the new context
	if err := p.rules.RefreshRules(currentFeature, targetPhase.String()); err != nil {
		return nil, fmt.Errorf("failed to refresh rules: %w", err)
	}

	// Update state
	p.state.CurrentPhase = targetPhase

	// Call state changed hook if available
	if p.state.OnStateChanged != nil {
		p.state.OnStateChanged()
	}

	// Create the result message
	message := fmt.Sprintf("Moved to %s phase.", targetPhase)

	// Add information about existing files if needed
	if hasImpact {
		message += " Note: Existing files were detected for the target phase. Review required."
	}

	return NewResultWithRulesChanged(message), nil
}
