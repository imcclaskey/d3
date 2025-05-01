package command

import (
	"context"
	"fmt"
	"path/filepath"
	
	"github.com/imcclaskey/i3/internal/creator"
	"github.com/imcclaskey/i3/internal/rulegen"
	"github.com/imcclaskey/i3/internal/validation"
)

// Move transitions to a different phase or feature
type Move struct {
	Phase       string // The phase to move to
	Feature     string // The feature to move to (optional)
	ExitSession bool   // Whether to exit the current session before moving
}

// NewMove creates a move command
func NewMove(phase, feature string, exitSession bool) Move {
	return Move{
		Phase:       phase,
		Feature:     feature,
		ExitSession: exitSession,
	}
}

// Run implements the Command interface
func (m Move) Run(ctx context.Context, cfg Config) (Result, error) {
	// Validate the phase name
	if err := validation.Phase(m.Phase); err != nil {
		// Return zero Result on error
		return Result{}, err
	}
	
	// Create i3 core structure
	if err := creator.EnsureDirectories(cfg.I3Dir); err != nil {
		// Return zero Result on error
		return Result{}, err
	}
	
	if err := creator.EnsureBasicFiles(cfg.I3Dir); err != nil {
		// Return zero Result on error
		return Result{}, err
	}
	
	// Special case: setup phase doesn't require a feature
	if m.Phase != "setup" && m.Feature == "" {
		// Try to get the feature from the current session
		currentFeature, err := cfg.Session.Feature()
		if err != nil {
			// Return zero Result on error
			return Result{}, fmt.Errorf("failed to get current feature: %w", err)
		}
		
		if currentFeature == "" {
			// Return zero Result on error
			return Result{}, fmt.Errorf("no feature specified (suggestion: specify a feature name for non-setup phases)")
		}
		
		m.Feature = currentFeature
	}
	
	// For non-setup phases, create feature files
	if m.Phase != "setup" {
		if err := creator.EnsurePhaseFiles(cfg.FeaturesDir, m.Feature, m.Phase); err != nil {
			// Return zero Result on error
			return Result{}, err
		}
	}
	
	// Exit current session if requested
	if m.ExitSession {
		exitCmd := NewExit()
		// Exit command now returns Result, ignore it here or log it if needed
		if _, err := exitCmd.Run(ctx, cfg); err != nil {
			// Return zero Result on error
			return Result{}, err // Propagate error
		}
	}
	
	// Update session
	if err := cfg.Session.Start(m.Feature); err != nil {
		// Return zero Result on error
		return Result{}, fmt.Errorf("failed to start session: %w", err)
	}
	
	if err := cfg.Session.SetPhase(m.Phase); err != nil {
		// Return zero Result on error
		return Result{}, fmt.Errorf("failed to set phase: %w", err)
	}
	
	// Generate cursor rules for the current phase/feature
	if err := generateRules(cfg, m.Phase, m.Feature); err != nil {
		// Return zero Result on error
		return Result{}, fmt.Errorf("failed to generate cursor rules: %w", err)
	}
	
	// Generate success message
	message := fmt.Sprintf("Moved to %s phase", m.Phase)
	if m.Phase != "setup" {
		message += fmt.Sprintf(" of feature '%s'", m.Feature)
	}
	
	// Collect any content warnings
	warnings := validation.ContentWarnings(cfg.I3Dir)
	// Do not append warnings to message anymore
	/*
		if len(warnings) > 0 {
			message = fmt.Sprintf("%s\n\nWarnings:\n%s", message, strings.Join(warnings, "\n"))
		}
	*/
	
	// Return Result struct with message and warnings separated
	return NewResult(message, nil, warnings), nil
}

// generateRules generates cursor rules for the given phase and feature
func generateRules(cfg Config, phase, feature string) error {
	templatesDir := filepath.Join(cfg.I3Dir, "templates")
	outputDir := filepath.Join(cfg.WorkspaceRoot, ".cursor", "rules")
	
	ruleGen := rulegen.NewGenerator(templatesDir, outputDir)
	return ruleGen.CreateRuleFiles(phase, feature)
} 