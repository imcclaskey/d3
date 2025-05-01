package command

import (
	"context"
	"fmt"

	i3context "github.com/imcclaskey/i3/internal/context"
	"github.com/imcclaskey/i3/internal/rules"
	"github.com/imcclaskey/i3/internal/validation"
)

// Phase sets the current working phase within the active feature context.
type Phase struct {
	Phase string // Name of the phase to set
}

// NewPhase creates a new Phase command
func NewPhase(phase string) Phase {
	return Phase{Phase: phase}
}

// Run implements the Command interface
func (p Phase) Run(ctx context.Context, cfg Config) (Result, error) {
	// Validate phase name
	if err := validation.Phase(p.Phase); err != nil {
		return Result{}, err // Use validation error directly
	}

	// Load current context
	currentCtx, err := i3context.LoadContext(cfg.I3Dir)
	if err != nil {
		return Result{}, fmt.Errorf("loading context: %w", err)
	}

	// Check if a feature context is active
	if currentCtx.Feature == "" {
		return Result{}, fmt.Errorf("no active feature context. Use 'i3 enter <feature>' first")
	}

	// Update the phase
	updatedCtx := currentCtx
	updatedCtx.Phase = p.Phase

	// Save the updated context
	if err := i3context.SaveContext(cfg.I3Dir, updatedCtx); err != nil {
		return Result{}, fmt.Errorf("saving context: %w", err)
	}

	// Ensure rule files are generated
	ruleGenerator := rules.NewRuleFileGenerator("", cfg.CursorRulesDir)
	if err := ruleGenerator.EnsureRuleFiles(updatedCtx.Phase, updatedCtx.Feature); err != nil {
		return Result{}, fmt.Errorf("ensuring rule files: %w", err)
	}

	message := fmt.Sprintf("Current phase set to: %s (for feature: %s)", p.Phase, currentCtx.Feature)
	return NewResult(message, updatedCtx, nil), nil
} 