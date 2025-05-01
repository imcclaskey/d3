package command

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	
	"github.com/imcclaskey/i3/internal/creator"
	"github.com/imcclaskey/i3/internal/errors"
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
func (m Move) Run(ctx context.Context, cfg Config) (string, error) {
	// Validate the phase name
	if err := validation.Phase(m.Phase); err != nil {
		return "", err
	}
	
	// Create i3 core structure
	if err := creator.EnsureDirectories(cfg.I3Dir); err != nil {
		return "", err
	}
	
	if err := creator.EnsureBasicFiles(cfg.I3Dir); err != nil {
		return "", err
	}
	
	// Special case: setup phase doesn't require a feature
	if m.Phase != "setup" && m.Feature == "" {
		// Try to get the feature from the current session
		currentFeature, err := cfg.Session.Feature()
		if err != nil {
			return "", errors.Wrap(err, "failed to get current feature")
		}
		
		if currentFeature == "" {
			return "", errors.WithSuggestion(
				errors.New("no feature specified"),
				"specify a feature name for non-setup phases",
			)
		}
		
		m.Feature = currentFeature
	}
	
	// For non-setup phases, create feature files
	if m.Phase != "setup" {
		if err := creator.EnsurePhaseFiles(cfg.FeaturesDir, m.Feature, m.Phase); err != nil {
			return "", err
		}
	}
	
	// Exit current session if requested
	if m.ExitSession {
		exitCmd := NewExit()
		if _, err := exitCmd.Run(ctx, cfg); err != nil {
			return "", err
		}
	}
	
	// Update session
	if err := cfg.Session.Start(m.Feature); err != nil {
		return "", errors.Wrap(err, "failed to start session")
	}
	
	if err := cfg.Session.SetPhase(m.Phase); err != nil {
		return "", errors.Wrap(err, "failed to set phase")
	}
	
	// Generate cursor rules for the current phase/feature
	if err := generateRules(cfg, m.Phase, m.Feature); err != nil {
		return "", errors.Wrap(err, "failed to generate cursor rules")
	}
	
	// Generate success message
	message := fmt.Sprintf("Moved to %s phase", m.Phase)
	if m.Phase != "setup" {
		message += fmt.Sprintf(" of feature '%s'", m.Feature)
	}
	
	// Collect any content warnings
	warnings := validation.ContentWarnings(cfg.I3Dir)
	if len(warnings) > 0 {
		message = fmt.Sprintf("%s\n\nWarnings:\n%s", message, strings.Join(warnings, "\n"))
	}
	
	return message, nil
}

// generateRules generates cursor rules for the given phase and feature
func generateRules(cfg Config, phase, feature string) error {
	templatesDir := filepath.Join(cfg.I3Dir, "templates")
	outputDir := filepath.Join(cfg.WorkspaceRoot, ".cursor", "rules")
	
	ruleGen := rulegen.NewGenerator(templatesDir, outputDir)
	return ruleGen.CreateRuleFiles(phase, feature)
} 