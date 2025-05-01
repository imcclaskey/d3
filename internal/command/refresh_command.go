// Package command implements commands for the i3 CLI
package command

import (
	"context"
	"fmt"

	i3context "github.com/imcclaskey/i3/internal/context"
	"github.com/imcclaskey/i3/internal/rules"
	"github.com/imcclaskey/i3/internal/workspace"
	"github.com/imcclaskey/i3/internal/validation"
)

// Refresh ensures necessary i3 files and directories exist.
type Refresh struct{}

// NewRefresh creates a new Refresh command
func NewRefresh() Refresh {
	return Refresh{}
}

// Run implements the Command interface
func (r Refresh) Run(ctx context.Context, cfg Config) (Result, error) {
	// 1. Validate initialization
	if err := validation.Init(cfg.I3Dir); err != nil {
		return Result{}, err
	}

	// 2. Ensure base directories exist (idempotent)
	if err := workspace.EnsureDirectories(cfg.I3Dir); err != nil {
		return Result{}, fmt.Errorf("failed to ensure base directories: %w", err)
	}

	// 3. Ensure basic project files exist (idempotent)
	if err := workspace.EnsureBasicFiles(cfg.I3Dir); err != nil {
		return Result{}, fmt.Errorf("failed to ensure basic files: %w", err)
	}

	// 4. Ensure files for the *current* context (if any) exist
	currentCtx, err := i3context.LoadContext(cfg.I3Dir)
	if err != nil {
		return Result{}, fmt.Errorf("loading context for refresh: %w", err)
	}

	if currentCtx.Feature != "" && currentCtx.Phase != "" {
		if err := workspace.EnsurePhaseFiles(cfg.FeaturesDir, currentCtx.Feature, currentCtx.Phase); err != nil {
			return Result{}, fmt.Errorf("failed to ensure files for phase %s in feature %s: %w",
				currentCtx.Phase, currentCtx.Feature, err)
		}
		
		// 5. Ensure rule files are generated
		ruleGenerator := rules.NewRuleFileGenerator("", cfg.CursorRulesDir)
		if err := ruleGenerator.EnsureRuleFiles(currentCtx.Phase, currentCtx.Feature); err != nil {
			return Result{}, fmt.Errorf("ensuring rule files: %w", err)
		}
	}

	return NewResult("i3 workspace refreshed.", nil, nil), nil
} 