// Package command implements commands for the i3 CLI
package command

import (
	"context"
	"fmt"
	"path/filepath"
	
	"github.com/imcclaskey/i3/internal/errors"
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
func (r Refresh) Run(ctx context.Context, cfg Config) (string, error) {
	// Get current feature if not specified
	feature := r.Feature
	if feature == "" {
		var err error
		feature, err = cfg.Session.Feature()
		if err != nil {
			return "", errors.Wrap(err, "failed to get current feature")
		}
	}
	
	// Get current phase
	phase, err := cfg.Session.Phase()
	if err != nil {
		return "", errors.Wrap(err, "failed to get current phase")
	}
	
	// Check if we have an active session
	if phase == "" {
		return "No active session to refresh.", nil
	}
	
	// Regenerate cursor rules
	templatesDir := filepath.Join(cfg.I3Dir, "templates")
	outputDir := filepath.Join(cfg.WorkspaceRoot, ".cursor", "rules")
	
	ruleGen := rulegen.NewGenerator(templatesDir, outputDir)
	if err := ruleGen.CreateRuleFiles(phase, feature); err != nil {
		return "", errors.Wrap(err, "failed to regenerate cursor rules")
	}
	
	return fmt.Sprintf("Refreshed rules for %s phase of feature '%s'.", phase, feature), nil
} 