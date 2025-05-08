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
	ClearGeneratedRules() error
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
	state    *State
	features FeatureServicer
	session  StorageService
	rules    RulesServicer
	phases   PhaseServicer
	fs       ports.FileSystem
}

// New creates a new project instance from project root, now with dependency injection
// It no longer performs I/O.
func New(projectRoot string, fs ports.FileSystem, sessionSvc StorageService, featureSvc FeatureServicer, rulesSvc RulesServicer, phasesSvc PhaseServicer) *Project {
	state := newState(projectRoot)

	proj := &Project{
		state:    state,
		session:  sessionSvc,
		rules:    rulesSvc,
		phases:   phasesSvc,
		features: featureSvc,
		fs:       fs,
	}
	return proj
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

// checkInitialized checks if the project seems initialized (internal helper)
// Called directly by IsInitialized now.
func (p *Project) checkInitialized() bool {
	_, err := p.fs.Stat(p.state.D3Dir)
	return err == nil
}

// IsInitialized checks if the project is initialized (public method)
// It performs an fs.Stat check every time.
func (p *Project) IsInitialized() bool {
	return p.checkInitialized()
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
	// Check current initialized status via direct filesystem check
	isCurrentlyInitialized := p.IsInitialized()

	if isCurrentlyInitialized && !clean {
		return NewResult("Project already initialized."), nil
	}

	if clean && isCurrentlyInitialized {
		if err := p.fs.RemoveAll(p.state.D3Dir); err != nil {
			return nil, fmt.Errorf("failed to clean existing d3 directory: %w", err)
		}
		if err := p.session.ClearActiveFeature(); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to clear transient session during clean init: %v\n", err)
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

	// Initialize transient session (empty active feature)
	if err := p.session.ClearActiveFeature(); err != nil {
		return nil, fmt.Errorf("failed to initialize transient session: %w", err)
	}

	// Add .gitignore entries
	if err := p.ensureGitignoreEntries(); err != nil {
		// Log warning, but don't fail init
		fmt.Fprintf(os.Stderr, "warning: failed to update .gitignore: %v\n", err)
	}

	// Initialize rules
	if err := p.rules.RefreshRules("", ""); err != nil {
		return nil, fmt.Errorf("failed to initialize rules: %w", err)
	}

	// Mark project as initialized - No longer needed as IsInitialized checks disk
	// newInitState := true
	// p._isInitialized = &newInitState // Removed
	p.state.CurrentFeature = ""
	p.state.CurrentPhase = session.None

	return NewResultWithRulesChanged("Project initialized successfully."), nil
}

// ensureGitignoreEntries creates .gitignore files in specific d3 directories.
func (p *Project) ensureGitignoreEntries() error {
	type gitignoreTarget struct {
		path    string // Relative to project root
		content string
	}

	targets := []gitignoreTarget{
		{
			path:    filepath.Join(".d3", ".gitignore"),
			content: ".session\n", // Content for .d3/.gitignore
		},
		{
			path:    filepath.Join(".cursor", "rules", "d3", ".gitignore"),
			content: "*.gen.rules\n", // Content for .cursor/rules/d3/.gitignore
		},
	}

	for _, target := range targets {
		fullPath := filepath.Join(p.state.ProjectRoot, target.path)
		dir := filepath.Dir(fullPath)

		// Ensure the target directory exists
		if err := p.fs.MkdirAll(dir, 0755); err != nil {
			return fmt.Errorf("failed to create directory %s for .gitignore: %w", dir, err)
		}

		// Write the .gitignore file (overwrites if exists)
		if err := p.fs.WriteFile(fullPath, []byte(target.content), 0644); err != nil {
			return fmt.Errorf("failed to write %s: %w", fullPath, err)
		}
	}

	return nil
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

	// Update the transient session file
	if err := p.session.SaveActiveFeature(featureName); err != nil {
		return nil, fmt.Errorf("failed to save active feature to session: %w", err)
	}

	// Ensure standard phase files exist for the feature (existing logic)
	if err := p.phases.EnsurePhaseFiles(featureInfo.Path); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to ensure phase files for %s: %v\n", featureName, err)
	}

	// Update in-memory project state
	p.state.CurrentFeature = featureName
	p.state.CurrentPhase = session.Define

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

	// Reload CurrentFeature from transient store in case it changed externally?
	// Or rely on in-memory state which should be accurate if only d3 commands modify it.
	// Let's rely on in-memory for now.
	if p.state.CurrentFeature == "" {
		// Load from transient store as fallback?
		activeFeature, loadErr := p.session.LoadActiveFeature()
		if loadErr != nil || activeFeature == "" {
			return nil, ErrNoActiveFeature
		}
		p.state.CurrentFeature = activeFeature // Update memory
		// Need to load phase too if we just loaded feature
		phase, phaseErr := p.features.GetFeaturePhase(ctx, activeFeature)
		if phaseErr != nil {
			return nil, fmt.Errorf("failed to load phase for active feature %s: %w", activeFeature, phaseErr)
		}
		p.state.CurrentPhase = phase
	}

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

	// Get the feature's phase from state.yaml (Correct - uses state.yaml)
	phase, err := p.features.GetFeaturePhase(ctx, featureName)
	if err != nil {
		return nil, fmt.Errorf("cannot enter feature '%s': %w", featureName, err)
	}

	// Update the transient session file
	if err := p.session.SaveActiveFeature(featureName); err != nil {
		return nil, fmt.Errorf("failed to save active feature to session: %w", err)
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

	// Determine feature being exited from memory (or transient store?)
	// Let's rely on memory first, as commands should keep it sync'd
	exitedFeatureName := p.state.CurrentFeature
	if exitedFeatureName == "" {
		// Maybe try loading from transient store just in case memory is stale?
		loadedFeature, loadErr := p.session.LoadActiveFeature()
		if loadErr != nil {
			// Log error loading, but proceed as if no feature active
			fmt.Fprintf(os.Stderr, "warning: error loading active feature during exit: %v\n", loadErr)
		} else {
			exitedFeatureName = loadedFeature
		}
	}

	if exitedFeatureName == "" {
		// If still no feature after checking store, truly no active feature.
		// Ensure rules are cleared anyway (best effort)
		if err := p.rules.ClearGeneratedRules(); err != nil { // Call new Clear method
			fmt.Fprintf(os.Stderr, "warning: failed to clear rules during exit (no active feature): %v\n", err)
		}
		return NewResult("No active feature to exit."), nil
	}

	// Clear the transient session file
	if err := p.session.ClearActiveFeature(); err != nil {
		return nil, fmt.Errorf("failed to clear active feature session: %w", err)
	}

	// Clear in-memory project state
	p.state.CurrentFeature = ""
	p.state.CurrentPhase = session.None

	// Update/clear rules to reflect no active feature.
	if err := p.rules.ClearGeneratedRules(); err != nil { // Call new Clear method
		// Log error but proceed with exit.
		fmt.Fprintf(os.Stderr, "warning: failed to clear rules during exit: %v\n", err)
	}

	// Call state changed hook if available.
	if p.state.OnStateChanged != nil {
		p.state.OnStateChanged()
	}

	return NewResultWithRulesChanged(fmt.Sprintf("Exited feature '%s'. No active feature. Cursor rules cleared.", exitedFeatureName)), nil
}
