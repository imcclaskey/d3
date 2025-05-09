// Package project provides project management functionality
package project

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/imcclaskey/d3/internal/core/phase"
	"github.com/imcclaskey/d3/internal/core/ports"
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

//go:generate mockgen -package=project -destination=project_service_mock.go . ProjectService

// ProjectService defines the interface for project operations used by CLI and MCP.
// This allows for mocking the entire project service in tests for commands/tools.
type ProjectService interface {
	Init(clean bool, refresh bool) (*Result, error)
	CreateFeature(ctx context.Context, featureName string) (*Result, error)
	ChangePhase(ctx context.Context, targetPhase phase.Phase) (*Result, error)
	EnterFeature(ctx context.Context, featureName string) (*Result, error)
	ExitFeature(ctx context.Context) (*Result, error)
	DeleteFeature(ctx context.Context, featureName string) (*Result, error)
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
	CurrentPhase   phase.Phase

	// Hooks
	OnStateChanged func()
}

// Project coordinates all d3 services
type Project struct {
	state    *State
	features FeatureServicer
	rules    RulesServicer
	phases   PhaseServicer
	fs       ports.FileSystem
	fileOp   FileOperator
}

// New creates a new project instance from project root, now with dependency injection
// It no longer performs I/O.
func New(projectRoot string, fs ports.FileSystem, featureSvc FeatureServicer, rulesSvc RulesServicer, phasesSvc PhaseServicer, fileOp FileOperator) *Project {
	state := newState(projectRoot)

	proj := &Project{
		state:    state,
		rules:    rulesSvc,
		phases:   phasesSvc,
		features: featureSvc,
		fs:       fs,
		fileOp:   fileOp,
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

// Init initializes or refreshes the project
func (p *Project) Init(clean bool, refresh bool) (*Result, error) {
	originalIsCurrentlyInitialized := p.IsInitialized()
	actionMessage := "Project initialized successfully." // Default message
	performedClean := false

	if clean {
		performedClean = true
		if originalIsCurrentlyInitialized {
			if err := p.fs.RemoveAll(p.state.D3Dir); err != nil {
				return nil, fmt.Errorf("failed to clean existing .d3 directory: %w", err)
			}
			if err := p.features.ClearActiveFeature(); err != nil {
				fmt.Fprintf(os.Stderr, "warning: failed to clear transient session during clean init: %v\n", err)
			}
		}
		actionMessage = "Project cleaned and re-initialized successfully."
	} else if refresh {
		if !originalIsCurrentlyInitialized {
			// Refreshing a non-existent project is like a standard init
			actionMessage = "Project initialized successfully (refresh on non-existent project)."
		} else {
			actionMessage = "Project refreshed successfully."
		}
		// For refresh, we proceed to ensure all components even if already initialized.
	} else { // Standard init (neither clean nor refresh)
		if originalIsCurrentlyInitialized {
			return NewResult("Project already initialized. Use --refresh to update or --clean to reset."), nil
		}
		// New project initialization, actionMessage is already "Project initialized successfully."
	}

	// Create .d3/ and .d3/features/ directories (idempotent)
	directories := []string{p.state.D3Dir, p.state.FeaturesDir}
	for _, dir := range directories {
		if err := p.fs.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	// Always attempt to preserve other entries in mcp.json.
	// EnsureMCPJSON will create a new file with only the d3 entry if mcp.json doesn't exist.
	err := p.fileOp.EnsureMCPJSON(p.fs, p.state.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure mcp.json: %w", err)
	}

	// Ensure d3-specific .gitignore files (e.g., .d3/.gitignore)
	// Call the exported helper function from projectfiles package
	if err = p.fileOp.EnsureD3GitignoreEntries(p.fs, p.state.D3Dir, p.state.CursorRulesDir, p.state.ProjectRoot); err != nil {
		// fmt.Fprintf(os.Stderr, "warning: failed to update d3-specific .gitignore files: %v\n", err)
		return nil, fmt.Errorf("failed to update d3-specific .gitignore files: %w", err)
	}

	// Initialize/Refresh rules (.cursor/rules/d3/)
	if err := p.rules.RefreshRules("", ""); err != nil {
		return nil, fmt.Errorf("failed to initialize/refresh rules: %w", err)
	}

	// Initialize transient session (empty active feature) only if it's a fresh start
	// A fresh start means it was cleaned, or it was not originally initialized.
	// Do not clear active feature during a refresh of an already initialized project.
	if performedClean || !originalIsCurrentlyInitialized {
		if err := p.features.ClearActiveFeature(); err != nil {
			// This might be an error if the session store itself is problematic, but init should still mostly succeed.
			// fmt.Fprintf(os.Stderr, "warning: failed to initialize/clear transient session: %v\n", err)
			return nil, fmt.Errorf("failed to initialize/clear transient session: %w", err)
		}
	}

	p.state.CurrentFeature = ""
	p.state.CurrentPhase = phase.None

	return NewResultWithRulesChanged(actionMessage), nil
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

	// Update the active feature file via feature service
	if err := p.features.SetActiveFeature(featureName); err != nil {
		return nil, fmt.Errorf("failed to save active feature: %w", err)
	}

	// Ensure standard phase files exist for the feature (existing logic)
	if err := p.phases.EnsurePhaseFiles(featureInfo.Path); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to ensure phase files for %s: %v\n", featureName, err)
	}

	// Update in-memory project state
	p.state.CurrentFeature = featureName
	p.state.CurrentPhase = phase.Define

	// Update the rules with the new context
	// The phase for rules refresh should come from the newly set in-memory state.
	if err := p.rules.RefreshRules(p.state.CurrentFeature, string(p.state.CurrentPhase)); err != nil {
		return nil, fmt.Errorf("failed to refresh rules: %w", err)
	}

	// Call state changed hook if available
	if p.state.OnStateChanged != nil {
		p.state.OnStateChanged()
	}

	return NewResultWithRulesChanged(fmt.Sprintf("Feature '%s' created and set to define phase.", featureName)), nil
}

// ChangePhase changes the current phase of the active feature
func (p *Project) ChangePhase(ctx context.Context, targetPhase phase.Phase) (*Result, error) {
	// Check if project is initialized
	if err := p.RequiresInitialized(); err != nil {
		return nil, err
	}

	// Reload CurrentFeature from transient store in case it changed externally?
	// Or rely on in-memory state which should be accurate if only d3 commands modify it.
	// Let's rely on in-memory for now.
	if p.state.CurrentFeature == "" {
		// Load from active feature file as fallback?
		activeFeature, loadErr := p.features.GetActiveFeature()
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
	if err := p.rules.RefreshRules(currentFeatureName, string(targetPhase)); err != nil {
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
	phaseDir := filepath.Join(p.state.FeaturesDir, currentFeatureName, string(targetPhase))
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

	// Update the active feature file via feature service
	if err := p.features.SetActiveFeature(featureName); err != nil {
		return nil, fmt.Errorf("failed to save active feature: %w", err)
	}

	// Update in-memory project state
	p.state.CurrentFeature = featureName
	p.state.CurrentPhase = phase

	// Update rules for the new context
	if err := p.rules.RefreshRules(p.state.CurrentFeature, string(p.state.CurrentPhase)); err != nil {
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
		// Maybe try loading from active feature file just in case memory is stale?
		loadedFeature, loadErr := p.features.GetActiveFeature()
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

	// Clear the active feature file via feature service
	if err := p.features.ClearActiveFeature(); err != nil {
		return nil, fmt.Errorf("failed to clear active feature: %w", err)
	}

	// Clear in-memory project state
	p.state.CurrentFeature = ""
	p.state.CurrentPhase = phase.None

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

// DeleteFeature removes a feature and its associated data.
// If the deleted feature is the active one, it also clears the active feature context.
func (p *Project) DeleteFeature(ctx context.Context, featureName string) (*Result, error) {
	if err := p.RequiresInitialized(); err != nil {
		return nil, err
	}

	// Attempt to delete the feature using the feature service
	// feature.Service.DeleteFeature now handles clearing of .active_feature file internally
	// and returns a boolean indicating if the active context was indeed cleared.
	activeContextCleared, err := p.features.DeleteFeature(ctx, featureName)
	if err != nil {
		return nil, fmt.Errorf("failed to delete feature '%s' using service: %w", featureName, err)
	}

	message := fmt.Sprintf("Feature '%s' deleted successfully.", featureName)
	rulesNeedUpdate := false // Renamed from rulesNeedRefresh for clarity

	if activeContextCleared {
		p.state.CurrentFeature = ""
		p.state.CurrentPhase = phase.None

		// Rules need to be cleared/refreshed to reflect no active feature
		if err := p.rules.ClearGeneratedRules(); err != nil {
			message += fmt.Sprintf(" Warning: failed to clear rules after deleting active feature: %v", err)
		}
		rulesNeedUpdate = true // Indicate rules were touched (cleared in this case)
		message += " Active feature context has been cleared."
	}

	if p.state.OnStateChanged != nil {
		p.state.OnStateChanged()
	}

	if rulesNeedUpdate {
		return NewResultWithRulesChanged(message), nil
	}
	return NewResult(message), nil
}
