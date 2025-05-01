// Package command implements commands for the i3 CLI
package command

import (
	"context"
	"fmt"
	"path/filepath"
	
	"github.com/imcclaskey/i3/internal/rulegen"
)

// Refresh regenerates cursor rules for the current session
type Refresh struct {
	Feature string // Optional feature to refresh (overrides current feature)
}

// NewRefresh creates a refresh command
func NewRefresh(feature string) Refresh {
	return Refresh{
		Feature: feature,
	}
}

// Run implements the Command interface
func (r Refresh) Run(ctx context.Context, cfg Config) (Result, error) {
	// Get current feature if not specified
	feature := r.Feature
	if feature == "" {
		var err error
		feature, err = cfg.Session.Feature()
		if err != nil {
			return Result{}, fmt.Errorf("failed to get current feature: %w", err)
		}
	}
	
	// Get current phase
	phase, err := cfg.Session.Phase()
	if err != nil {
		return Result{}, fmt.Errorf("failed to get current phase: %w", err)
	}
	
	// Check if we have an active session
	if phase == "" {
		return NewResult("No active session to refresh.", nil, nil), nil
	}
	
	// Regenerate cursor rules
	templatesDir := filepath.Join(cfg.I3Dir, "templates")
	outputDir := filepath.Join(cfg.WorkspaceRoot, ".cursor", "rules")
	
	ruleGen := rulegen.NewGenerator(templatesDir, outputDir)
	if err := ruleGen.CreateRuleFiles(phase, feature); err != nil {
		return Result{}, fmt.Errorf("failed to regenerate cursor rules: %w", err)
	}
	
	message := fmt.Sprintf("Refreshed rules for %s phase of feature '%s'.", phase, feature)
	return NewResult(message, nil, nil), nil
} 