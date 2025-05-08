// Package project provides project management functionality
package project

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imcclaskey/d3/internal/core/feature"
	"github.com/imcclaskey/d3/internal/core/ports"
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

//go:generate mockgen -package=project -destination=interfaces_mock.go github.com/imcclaskey/d3/internal/project StorageService,FeatureServicer,RulesServicer,PhaseServicer
//go:generate mockgen -package=project -destination=project_service_mock.go github.com/imcclaskey/d3/internal/project ProjectService

// StorageService defines the interface for session storage operations.
type StorageService interface {
	LoadActiveFeature() (string, error)
	SaveActiveFeature(featureName string) error
	ClearActiveFeature() error
}

// FeatureServicer defines the interface for feature management operations.
type FeatureServicer interface {
	CreateFeature(ctx context.Context, featureName string) (*feature.FeatureInfo, error)
	GetFeaturePhase(ctx context.Context, featureName string) (session.Phase, error)
	SetFeaturePhase(ctx context.Context, featureName string, phase session.Phase) error
}

// RulesServicer defines the interface for rule management operations.
type RulesServicer interface {
	RefreshRules(feature string, phaseStr string) error
}

// PhaseServicer defines the interface for phase management operations.
type PhaseServicer interface {
	EnsurePhaseFiles(featureDir string) error
}

// ProjectService defines the interface for project operations used by CLI and MCP.
// This allows for mocking the entire project service in tests for commands/tools.
type ProjectService interface {
	Init(clean bool) (*Result, error)
	CreateFeature(ctx context.Context, featureName string) (*Result, error)
	ChangePhase(ctx context.Context, targetPhase session.Phase) (*Result, error)
	EnterFeature(ctx context.Context, featureName string) (*Result, error)
	ExitFeature(ctx context.Context) (*Result, error)
	IsInitialized() bool
	RequiresInitialized() error
	// Add other project methods here as they are consumed by CLI/MCP
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
	features      FeatureServicer
	session       StorageService
	rules         RulesServicer
	phases        PhaseServicer
	fs            ports.FileSystem
	isInitialized bool // Tracks whether the project has been initialized
}

// New creates a new project instance from project root, now with dependency injection
func New(projectRoot string, fs ports.FileSystem, sessionSvc StorageService, featureSvc FeatureServicer, rulesSvc RulesServicer, phasesSvc PhaseServicer) *Project {
	state := newState(projectRoot)

	return &Project{
		state:         state,
		session:       sessionSvc,
		rules:         rulesSvc,
		phases:        phasesSvc,
		features:      featureSvc,
		fs:            fs,
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
	_, err := p.fs.Stat(p.state.D3Dir)
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
		if err := p.fs.RemoveAll(p.state.D3Dir); err != nil {
			return nil, fmt.Errorf("failed to clean existing d3 directory: %w", err)
		}
	}

	// Create directories
	directories := []string{
		p.state.D3Dir,
		p.state.FeaturesDir,
	}

	for _, dir := range directories {
		if err := p.fs.MkdirAll(dir, 0755); err != nil {
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

	// Create the feature using the feature service.
	// The feature.Service.CreateFeature is now responsible for creating the directory
	// AND the initial state.yaml file with a default phase (e.g., Define).
	featureInfo, err := p.features.CreateFeature(ctx, featureName)
	if err != nil {
		return nil, fmt.Errorf("failed to create feature using service: %w", err)
	}

	// Load current global session state (session.yml) to update CurrentFeature.
	sessionState, err := p.session.Load()
	if err != nil {
		// If session.yml doesn't exist, initialize a new one.
		if os.IsNotExist(errors.Unwrap(err)) {
			sessionState = &session.SessionState{Version: "1.0"}
		} else {
			return nil, fmt.Errorf("failed to load session: %w", err)
		}
	}

	// Update the current feature in session.yml
	sessionState.CurrentFeature = featureName

	// Save the session state (session.yml)
	if err := p.session.Save(sessionState); err != nil {
		return nil, fmt.Errorf("failed to save session: %w", err)
	}

	// Ensure standard phase files exist for the feature (existing logic)
	if err := p.phases.EnsurePhaseFiles(featureInfo.Path); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to ensure phase files for %s: %v\n", featureName, err)
	}

	// Update in-memory project state
	p.state.CurrentFeature = featureName
	p.state.CurrentPhase = session.Define // Set in-memory phase to Define, consistent with feature.Service creating state.yaml with Define

	// Update the rules with the new context
	// The phase for rules refresh should come from the newly set in-memory state.
	if err := p.rules.RefreshRules(p.state.CurrentFeature, p.state.CurrentPhase.String()); err != nil {
		return nil, fmt.Errorf("failed to refresh rules: %w", err)
	}

	// Call state changed hook if available
	if p.state.OnStateChanged != nil {
		p.state.OnStateChanged()
	}

	return NewResultWithRulesChanged(fmt.Sprintf("Feature '%s' created and set to define phase.", featureName)), nil
}

// ChangePhase changes the current phase of the active feature
func (p *Project) ChangePhase(ctx context.Context, targetPhase session.Phase) (*Result, error) {
	// Check if project is initialized
	if err := p.RequiresInitialized(); err != nil {
		return nil, err
	}

	// Ensure there is an active feature in memory
	if p.state.CurrentFeature == "" {
		return nil, ErrNoActiveFeature
	}

	// Get current state from in-memory p.state
	currentFeatureName := p.state.CurrentFeature
	currentInMemoryPhase := p.state.CurrentPhase

	// Check if we're already in the target phase
	if currentInMemoryPhase == targetPhase {
		return NewResult(fmt.Sprintf("Already in the %s phase.", targetPhase)), nil
	}

	// Persist the new phase to the feature's state.yaml file via FeatureServicer
	if err := p.features.SetFeaturePhase(ctx, currentFeatureName, targetPhase); err != nil {
		return nil, fmt.Errorf("failed to set feature phase for %s: %w", currentFeatureName, err)
	}

	// Update in-memory project state
	p.state.CurrentPhase = targetPhase

	// Load current global session state (session.yml) primarily to update LastModified
	sessionState, err := p.session.Load()
	if err != nil {
		// If session.yml doesn't exist, initialize a new one.
		if os.IsNotExist(errors.Unwrap(err)) {
			sessionState = &session.SessionState{Version: "1.0", CurrentFeature: currentFeatureName} // Ensure CurrentFeature is set
		} else {
			return nil, fmt.Errorf("failed to load session for ChangePhase: %w", err)
		}
	} else {
		// Ensure CurrentFeature in sessionState is consistent if it was loaded.
		sessionState.CurrentFeature = currentFeatureName
	}

	if err := p.session.Save(sessionState); err != nil {
		return nil, fmt.Errorf("failed to save session after phase change: %w", err)
	}

	// Update rules with the new context
	if err := p.rules.RefreshRules(currentFeatureName, targetPhase.String()); err != nil {
		return nil, fmt.Errorf("failed to refresh rules: %w", err)
	}

	// Ensure standard phase files exist for the new phase (existing logic)
	featureDirForPhaseFiles := filepath.Join(p.state.FeaturesDir, currentFeatureName)
	if err := p.phases.EnsurePhaseFiles(featureDirForPhaseFiles); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to ensure phase files for %s: %v\n", currentFeatureName, err)
	}

	// Call state changed hook if available
	if p.state.OnStateChanged != nil {
		p.state.OnStateChanged()
	}

	// Check for impact (existing files in target phase dir)
	hasImpact := false
	phaseDir := filepath.Join(p.state.FeaturesDir, currentFeatureName, targetPhase.String())
	if _, errStat := p.fs.Stat(phaseDir); errStat == nil {
		hasImpact = true
	} else if !os.IsNotExist(errStat) {
		return nil, fmt.Errorf("failed to check phase directory %s: %w", phaseDir, errStat)
	}

	message := fmt.Sprintf("Moved to %s phase.", targetPhase)
	if hasImpact {
		message += " Note: Existing files were detected for the target phase. Review required."
	}

	return NewResultWithRulesChanged(message), nil
}

// EnterFeature sets the specified feature as the active one, resuming its last phase.
func (p *Project) EnterFeature(ctx context.Context, featureName string) (*Result, error) {
	if err := p.RequiresInitialized(); err != nil {
		return nil, err
	}

	// Get the feature's last active phase from its state.yaml (or default if new)
	phase, err := p.features.GetFeaturePhase(ctx, featureName)
	if err != nil {
		// Provide a more specific error if GetFeaturePhase indicates feature doesn't exist
		// For now, wrap the error generically.
		return nil, fmt.Errorf("cannot enter feature '%s': %w", featureName, err)
	}

	// Update global session.yml to set this as the CurrentFeature
	sessionState, err := p.session.Load()
	if err != nil {
		// If session.yml doesn't exist, initialize a new one.
		if os.IsNotExist(errors.Unwrap(err)) {
			sessionState = &session.SessionState{Version: "1.0"}
		} else {
			return nil, fmt.Errorf("failed to load session for EnterFeature: %w", err)
		}
	}
	sessionState.CurrentFeature = featureName
	if err := p.session.Save(sessionState); err != nil {
		return nil, fmt.Errorf("failed to save session for EnterFeature: %w", err)
	}

	// Update in-memory project state
	p.state.CurrentFeature = featureName
	p.state.CurrentPhase = phase

	// Update rules for the new context
	if err := p.rules.RefreshRules(p.state.CurrentFeature, p.state.CurrentPhase.String()); err != nil {
		// Entering a feature implies rules MUST be refreshed. Treat failure as critical.
		return nil, fmt.Errorf("failed to refresh rules for feature '%s': %w", featureName, err)
	}

	// Call state changed hook if available
	if p.state.OnStateChanged != nil {
		p.state.OnStateChanged()
	}

	// Construct the Result message
	message := fmt.Sprintf("Entered feature '%s' in phase '%s'.", featureName, phase)
	return NewResultWithRulesChanged(message), nil // Return *Result, indicate rules changed
}

// ExitFeature clears the active feature context.
func (p *Project) ExitFeature(ctx context.Context) (*Result, error) {
	if err := p.RequiresInitialized(); err != nil {
		return nil, err
	}

	// Load the current session state from session.yaml to reliably determine the active feature.
	sessionState, err := p.session.Load()
	if err != nil {
		// If session.yml doesn't exist, there's no persisted active feature.
		if os.IsNotExist(errors.Unwrap(err)) {
			// Ensure in-memory state is also clear, just in case.
			p.state.CurrentFeature = ""
			p.state.CurrentPhase = session.None
			// Attempt to refresh rules to a no-feature state. Errors are logged by RefreshRules itself if critical.
			_ = p.rules.RefreshRules("", "") // Best effort, ignore error for exit cleanliness.
			return NewResult("No active feature to exit."), nil
		}
		// For other load errors, we can't be sure of the state.
		return nil, fmt.Errorf("failed to load session to determine active feature: %w", err)
	}

	// If sessionState is somehow nil (though Load should return error if so) or CurrentFeature is empty.
	if sessionState == nil || sessionState.CurrentFeature == "" {
		// Ensure in-memory state is also clear.
		p.state.CurrentFeature = ""
		p.state.CurrentPhase = session.None
		_ = p.rules.RefreshRules("", "") // Best effort.
		return NewResult("No active feature to exit."), nil
	}

	// At this point, sessionState.CurrentFeature holds the feature to be exited.
	exitedFeatureName := sessionState.CurrentFeature

	// Update session.yaml: clear the CurrentFeature.
	sessionState.CurrentFeature = ""

	if err := p.session.Save(sessionState); err != nil {
		// If we can't save the cleared session, this is a problem for subsequent commands.
		return nil, fmt.Errorf("failed to save session after clearing feature: %w", err)
	}

	// Clear in-memory project state to match.
	p.state.CurrentFeature = ""
	p.state.CurrentPhase = session.None

	// Update/clear rules to reflect no active feature.
	if err := p.rules.RefreshRules("", ""); err != nil {
		// Log error but proceed with exit. Exiting should ideally always succeed in clearing context.
		fmt.Fprintf(os.Stderr, "warning: failed to refresh/clear rules during exit: %v\n", err)
	}

	// Call state changed hook if available.
	if p.state.OnStateChanged != nil {
		p.state.OnStateChanged()
	}

	return NewResultWithRulesChanged(fmt.Sprintf("Exited feature '%s'. No active feature. Cursor rules cleared.", exitedFeatureName)), nil
}
