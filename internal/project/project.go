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

// --- Internal Helper Functions for Active State ---

func (p *Project) getActiveFeatureName() (string, error) {
	name, err := p.features.GetActiveFeature()
	if err != nil {
		return "", fmt.Errorf("failed to get active feature name: %w", err)
	}
	return name, nil
}

func (p *Project) getActiveFeaturePhase(ctx context.Context) (phase.Phase, error) {
	activeFeatureName, err := p.getActiveFeatureName()
	if err != nil {
		return phase.None, err // Error already wrapped by getActiveFeatureName
	}
	if activeFeatureName == "" {
		return phase.None, nil // No active feature, so no active phase
	}
	return p.features.GetFeaturePhase(ctx, activeFeatureName)
}

// getFeaturePhase is a direct pass-through, but useful for consistency if ever needed.
// For now, direct calls to p.features.GetFeaturePhase might be clearer where featureName is explicit.
// func (p *Project) getFeaturePhase(ctx context.Context, featureName string) (phase.Phase, error) {
// 	return p.features.GetFeaturePhase(ctx, featureName)
// }

// --- Modified Project Methods ---

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
			// ClearActiveFeature is called below for all fresh starts
		}
		actionMessage = "Project cleaned and re-initialized successfully."
	} else if refresh {
		if !originalIsCurrentlyInitialized {
			actionMessage = "Project initialized successfully (refresh on non-existent project)."
		} else {
			actionMessage = "Project refreshed successfully."
		}
	} else { // Standard init (neither clean nor refresh)
		if originalIsCurrentlyInitialized {
			return NewResult("Project already initialized. Use --refresh to update or --clean to reset."), nil
		}
	}

	directories := []string{p.state.D3Dir, p.state.FeaturesDir}
	for _, dir := range directories {
		if err := p.fs.MkdirAll(dir, 0755); err != nil {
			return nil, fmt.Errorf("failed to create directory %s: %w", dir, err)
		}
	}

	err := p.fileOp.EnsureMCPJSON(p.fs, p.state.ProjectRoot)
	if err != nil {
		return nil, fmt.Errorf("failed to ensure mcp.json: %w", err)
	}

	if err = p.fileOp.EnsureD3GitignoreEntries(p.fs, p.state.D3Dir, p.state.CursorRulesDir, p.state.ProjectRoot); err != nil {
		return nil, fmt.Errorf("failed to update d3-specific .gitignore files: %w", err)
	}

	// Initialize/Refresh rules with no active feature/phase context
	if err := p.rules.RefreshRules("", string(phase.None)); err != nil {
		return nil, fmt.Errorf("failed to initialize/refresh rules: %w", err)
	}

	if performedClean || !originalIsCurrentlyInitialized {
		if err := p.features.ClearActiveFeature(); err != nil {
			return nil, fmt.Errorf("failed to initialize/clear active feature: %w", err)
		}
	}

	// p.state.CurrentFeature = "" // Removed
	// p.state.CurrentPhase = phase.None // Removed

	// OnStateChanged hook is typically called by methods that change feature/phase context.
	// Init effectively sets it to none, so a call here might be relevant if the hook is used broadly.
	if p.state.OnStateChanged != nil {
		p.state.OnStateChanged()
	}

	return NewResultWithRulesChanged(actionMessage), nil
}

// CreateFeature creates a new feature and sets it as the current feature
func (p *Project) CreateFeature(ctx context.Context, featureName string) (*Result, error) {
	if err := p.RequiresInitialized(); err != nil {
		return nil, err
	}

	featureInfo, err := p.features.CreateFeature(ctx, featureName) // This now creates .phase with default Define
	if err != nil {
		return nil, fmt.Errorf("failed to create feature using service: %w", err)
	}

	if err := p.features.SetActiveFeature(featureName); err != nil {
		// Attempt to clean up the created feature directory if setting active fails
		_, _ = p.features.DeleteFeature(ctx, featureName) // DeleteFeature handles its own errors, ignore both results here
		return nil, fmt.Errorf("failed to set active feature %s: %w", featureName, err)
	}

	if err := p.phases.EnsurePhaseFiles(featureInfo.Path); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to ensure phase files for %s: %v\n", featureName, err)
	}

	// p.state.CurrentFeature = featureName // Removed
	// p.state.CurrentPhase = phase.Define // Removed

	// Refresh rules with the new feature and its initial phase (Define)
	if err := p.rules.RefreshRules(featureName, string(phase.Define)); err != nil {
		return nil, fmt.Errorf("failed to refresh rules for new feature %s: %w", featureName, err)
	}

	if p.state.OnStateChanged != nil {
		p.state.OnStateChanged()
	}

	return NewResultWithRulesChanged(fmt.Sprintf("Feature '%s' created and set to define phase.", featureName)), nil
}

// ChangePhase changes the current phase of the active feature
func (p *Project) ChangePhase(ctx context.Context, targetPhase phase.Phase) (*Result, error) {
	if err := p.RequiresInitialized(); err != nil {
		return nil, err
	}

	currentFeatureName, err := p.getActiveFeatureName()
	if err != nil {
		return nil, err // Error already wrapped
	}
	if currentFeatureName == "" {
		return nil, ErrNoActiveFeature
	}

	currentPhase, err := p.features.GetFeaturePhase(ctx, currentFeatureName) // Use direct call, not getActiveFeaturePhase to avoid double read of active feature name
	if err != nil {
		return nil, fmt.Errorf("failed to get current phase for feature %s: %w", currentFeatureName, err)
	}

	if currentPhase == targetPhase {
		return NewResult(fmt.Sprintf("Already in the %s phase.", targetPhase)), nil
	}

	if err := p.features.SetFeaturePhase(ctx, currentFeatureName, targetPhase); err != nil {
		return nil, fmt.Errorf("failed to set feature phase for %s: %w", currentFeatureName, err)
	}

	// p.state.CurrentPhase = targetPhase // Removed

	if err := p.rules.RefreshRules(currentFeatureName, string(targetPhase)); err != nil {
		return nil, fmt.Errorf("failed to refresh rules after phase change: %w", err)
	}

	featureDirForPhaseFiles := filepath.Join(p.state.FeaturesDir, currentFeatureName)
	if err := p.phases.EnsurePhaseFiles(featureDirForPhaseFiles); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to ensure phase files for %s: %v\n", currentFeatureName, err)
	}

	if p.state.OnStateChanged != nil {
		p.state.OnStateChanged()
	}

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

	// Check if feature exists and get its phase first
	retrievedPhase, err := p.features.GetFeaturePhase(ctx, featureName) // This also handles if feature doesn't exist implicitly by FeatureExists check in GetFeaturePhase
	if err != nil {
		// GetFeaturePhase will return a specific error if feature does not exist due to its FeatureExists check.
		return nil, fmt.Errorf("cannot enter feature '%s': %w", featureName, err)
	}

	if err := p.features.SetActiveFeature(featureName); err != nil {
		return nil, fmt.Errorf("failed to set active feature %s: %w", featureName, err)
	}

	// p.state.CurrentFeature = featureName // Removed
	// p.state.CurrentPhase = retrievedPhase // Removed

	if err := p.rules.RefreshRules(featureName, string(retrievedPhase)); err != nil {
		return nil, fmt.Errorf("failed to refresh rules for feature '%s': %w", featureName, err)
	}

	if p.state.OnStateChanged != nil {
		p.state.OnStateChanged()
	}

	message := fmt.Sprintf("Entered feature '%s' in phase '%s'.", featureName, retrievedPhase)
	return NewResultWithRulesChanged(message), nil
}

// ExitFeature clears the active feature context.
func (p *Project) ExitFeature(ctx context.Context) (*Result, error) {
	if err := p.RequiresInitialized(); err != nil {
		return nil, err
	}

	exitedFeatureName, err := p.getActiveFeatureName()
	if err != nil {
		// Log the error from getActiveFeatureName, but proceed as if no feature was active if name is empty.
		fmt.Fprintf(os.Stderr, "warning: error determining active feature during exit: %v\n", err)
		// If getActiveFeatureName itself had an issue (not just file not found), exitedFeatureName might be empty.
	}

	if exitedFeatureName == "" { // No active feature to exit
		// Ensure rules are cleared anyway (best effort)
		if ruleErr := p.rules.ClearGeneratedRules(); ruleErr != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to clear rules during exit (no active feature): %v\n", ruleErr)
		}
		// OnStateChanged hook could be called here if exiting from "no feature" has meaning
		if p.state.OnStateChanged != nil {
			p.state.OnStateChanged()
		}
		return NewResultWithRulesChanged("No active feature to exit. Cursor rules cleared."), nil // Rules are cleared so RulesChanged is true.
	}

	if err := p.features.ClearActiveFeature(); err != nil {
		return nil, fmt.Errorf("failed to clear active feature: %w", err)
	}

	// p.state.CurrentFeature = "" // Removed
	// p.state.CurrentPhase = phase.None // Removed

	if err := p.rules.ClearGeneratedRules(); err != nil {
		fmt.Fprintf(os.Stderr, "warning: failed to clear rules during exit: %v\n", err)
	}

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

	// features.DeleteFeature now handles clearing of .feature file internally
	// and returns a boolean indicating if the active context was indeed cleared.
	activeContextCleared, err := p.features.DeleteFeature(ctx, featureName)
	if err != nil {
		return nil, fmt.Errorf("failed to delete feature '%s' using service: %w", featureName, err)
	}

	message := fmt.Sprintf("Feature '%s' deleted successfully.", featureName)
	rulesWereImpacted := false

	if activeContextCleared {
		// p.state.CurrentFeature = "" // Removed
		// p.state.CurrentPhase = phase.None // Removed

		if err := p.rules.ClearGeneratedRules(); err != nil {
			// Append warning to message but don't fail the whole operation
			message += fmt.Sprintf(" Warning: failed to clear rules after deleting active feature: %v", err)
		} else {
			rulesWereImpacted = true // Rules were cleared
		}
		message += " Active feature context has been cleared."
		// OnStateChanged would be relevant here as context changed significantly
		if p.state.OnStateChanged != nil {
			p.state.OnStateChanged()
		}
	}

	// If rules were impacted (cleared), return ResultWithRulesChanged
	if rulesWereImpacted {
		return NewResultWithRulesChanged(message), nil
	}
	return NewResult(message), nil
}
